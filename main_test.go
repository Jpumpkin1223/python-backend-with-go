package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleSignup(t *testing.T) {
	// Reset users map before each test
	setupTest := func() {
		users = make(map[int]User)
		nextUserID = 1
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedUserID int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful signup",
			requestBody: SignupRequest{
				Name:     "홍길동",
				Email:    "hong@example.com",
				Password: "password123",
				Profile:  "안녕하세요",
			},
			expectedStatus: http.StatusCreated,
			expectedUserID: 1,
			checkResponse: func(t *testing.T, body []byte) {
				var resp SignupResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.UserID != 1 {
					t.Errorf("Expected user_id 1, got %d", resp.UserID)
				}
				if resp.Message != "회원가입이 완료되었습니다." {
					t.Errorf("Unexpected message: %s", resp.Message)
				}
			},
		},
		{
			name: "missing name",
			requestBody: SignupRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "name, email, and password are required" {
					t.Errorf("Unexpected error message: %s", resp.Message)
				}
			},
		},
		{
			name: "missing email",
			requestBody: SignupRequest{
				Name:     "홍길동",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "name, email, and password are required" {
					t.Errorf("Unexpected error message: %s", resp.Message)
				}
			},
		},
		{
			name: "missing password",
			requestBody: SignupRequest{
				Name:  "홍길동",
				Email: "hong@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "name, email, and password are required" {
					t.Errorf("Unexpected error message: %s", resp.Message)
				}
			},
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "invalid request body" {
					t.Errorf("Unexpected error message: %s", resp.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest()

			// Create request body
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handleSignup(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response body
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandleSignup_DuplicateEmail(t *testing.T) {
	// Reset users map
	users = make(map[int]User)
	nextUserID = 1

	// Create first user
	users[1] = User{
		ID:       1,
		Name:     "기존 사용자",
		Email:    "duplicate@example.com",
		Password: "password123",
		Profile:  "기존 프로필",
	}
	nextUserID = 2

	// Try to create user with same email
	requestBody := SignupRequest{
		Name:     "새 사용자",
		Email:    "duplicate@example.com",
		Password: "newpassword",
		Profile:  "새 프로필",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handleSignup(w, req)

	// Check status code
	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}

	// Check error message
	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Message != "email already exists" {
		t.Errorf("Expected 'email already exists', got '%s'", resp.Message)
	}

	// Verify user was not created
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
}

func TestHandleSignup_AutoIncrement(t *testing.T) {
	// Reset users map
	users = make(map[int]User)
	nextUserID = 1

	// Create three users
	testUsers := []SignupRequest{
		{Name: "User 1", Email: "user1@example.com", Password: "pass1"},
		{Name: "User 2", Email: "user2@example.com", Password: "pass2"},
		{Name: "User 3", Email: "user3@example.com", Password: "pass3"},
	}

	for i, user := range testUsers {
		body, _ := json.Marshal(user)
		req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handleSignup(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Request %d failed with status %d", i+1, w.Code)
		}

		var resp SignupResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		expectedID := i + 1
		if resp.UserID != expectedID {
			t.Errorf("Expected user_id %d, got %d", expectedID, resp.UserID)
		}
	}

	// Verify all users are stored
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}
