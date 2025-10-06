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

func TestHandleFollow(t *testing.T) {
	setupFollowTest := func() {
		users = make(map[int]User)
		follows = make(map[string]Follow)
		userFollowers = make(map[int]map[int]bool)
		userFollowing = make(map[int]map[int]bool)
		nextUserID = 1

		// Create test users
		users[1] = User{ID: 1, Name: "User1", Email: "user1@test.com", Password: "pass1"}
		users[2] = User{ID: 2, Name: "User2", Email: "user2@test.com", Password: "pass2"}
		nextUserID = 3
	}

	tests := []struct {
		name           string
		userID         string
		requestBody    FollowRequest
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:   "successful follow",
			userID: "2",
			requestBody: FollowRequest{
				FollowerID: 1,
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var resp FollowResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.FollowerID != 1 || resp.FollowingID != 2 {
					t.Errorf("Expected follower_id=1, following_id=2, got follower_id=%d, following_id=%d",
						resp.FollowerID, resp.FollowingID)
				}
				if resp.Message != "팔로우 성공" {
					t.Errorf("Unexpected message: %s", resp.Message)
				}
			},
		},
		{
			name:   "cannot follow yourself",
			userID: "1",
			requestBody: FollowRequest{
				FollowerID: 1,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "cannot follow yourself" {
					t.Errorf("Expected 'cannot follow yourself', got '%s'", resp.Message)
				}
			},
		},
		{
			name:   "follower user not found",
			userID: "2",
			requestBody: FollowRequest{
				FollowerID: 999,
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "follower user not found" {
					t.Errorf("Expected 'follower user not found', got '%s'", resp.Message)
				}
			},
		},
		{
			name:   "following user not found",
			userID: "999",
			requestBody: FollowRequest{
				FollowerID: 1,
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "following user not found" {
					t.Errorf("Expected 'following user not found', got '%s'", resp.Message)
				}
			},
		},
		{
			name:   "invalid user ID",
			userID: "abc",
			requestBody: FollowRequest{
				FollowerID: 1,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "invalid user ID" {
					t.Errorf("Expected 'invalid user ID', got '%s'", resp.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupFollowTest()

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/users/"+tt.userID+"/follow", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("userID", tt.userID)

			w := httptest.NewRecorder()
			handleFollow(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandleFollow_DuplicateFollow(t *testing.T) {
	users = make(map[int]User)
	follows = make(map[string]Follow)
	userFollowers = make(map[int]map[int]bool)
	userFollowing = make(map[int]map[int]bool)

	users[1] = User{ID: 1, Name: "User1", Email: "user1@test.com"}
	users[2] = User{ID: 2, Name: "User2", Email: "user2@test.com"}

	// Create first follow
	follows["1:2"] = Follow{FollowerID: 1, FollowingID: 2}
	userFollowers[2] = map[int]bool{1: true}
	userFollowing[1] = map[int]bool{2: true}

	// Try to follow again
	requestBody := FollowRequest{FollowerID: 1}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/users/2/follow", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("userID", "2")

	w := httptest.NewRecorder()
	handleFollow(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "already following this user" {
		t.Errorf("Expected 'already following this user', got '%s'", resp.Message)
	}
}

func TestHandleUnfollow(t *testing.T) {
	setupUnfollowTest := func() {
		users = make(map[int]User)
		follows = make(map[string]Follow)
		userFollowers = make(map[int]map[int]bool)
		userFollowing = make(map[int]map[int]bool)

		users[1] = User{ID: 1, Name: "User1", Email: "user1@test.com"}
		users[2] = User{ID: 2, Name: "User2", Email: "user2@test.com"}

		// Create existing follow relationship
		follows["1:2"] = Follow{FollowerID: 1, FollowingID: 2}
		userFollowers[2] = map[int]bool{1: true}
		userFollowing[1] = map[int]bool{2: true}
	}

	tests := []struct {
		name           string
		userID         string
		requestBody    FollowRequest
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:   "successful unfollow",
			userID: "2",
			requestBody: FollowRequest{
				FollowerID: 1,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp FollowResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.FollowerID != 1 || resp.FollowingID != 2 {
					t.Errorf("Expected follower_id=1, following_id=2, got follower_id=%d, following_id=%d",
						resp.FollowerID, resp.FollowingID)
				}
				if resp.Message != "언팔로우 성공" {
					t.Errorf("Unexpected message: %s", resp.Message)
				}
			},
		},
		{
			name:   "invalid user ID",
			userID: "abc",
			requestBody: FollowRequest{
				FollowerID: 1,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "invalid user ID" {
					t.Errorf("Expected 'invalid user ID', got '%s'", resp.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupUnfollowTest()

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodDelete, "/api/users/"+tt.userID+"/follow", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("userID", tt.userID)

			w := httptest.NewRecorder()
			handleUnfollow(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandleUnfollow_NotFollowing(t *testing.T) {
	users = make(map[int]User)
	follows = make(map[string]Follow)
	userFollowers = make(map[int]map[int]bool)
	userFollowing = make(map[int]map[int]bool)

	users[1] = User{ID: 1, Name: "User1", Email: "user1@test.com"}
	users[2] = User{ID: 2, Name: "User2", Email: "user2@test.com"}

	// Try to unfollow when not following
	requestBody := FollowRequest{FollowerID: 1}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodDelete, "/api/users/2/follow", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("userID", "2")

	w := httptest.NewRecorder()
	handleUnfollow(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "follow relationship not found" {
		t.Errorf("Expected 'follow relationship not found', got '%s'", resp.Message)
	}
}

func TestHandleGetFollowers(t *testing.T) {
	setupFollowersTest := func() {
		users = make(map[int]User)
		follows = make(map[string]Follow)
		userFollowers = make(map[int]map[int]bool)
		userFollowing = make(map[int]map[int]bool)

		// Create test users
		users[1] = User{ID: 1, Name: "User1", Email: "user1@test.com", Profile: "Profile1"}
		users[2] = User{ID: 2, Name: "User2", Email: "user2@test.com", Profile: "Profile2"}
		users[3] = User{ID: 3, Name: "User3", Email: "user3@test.com", Profile: "Profile3"}

		// User 2 and 3 follow User 1
		follows["2:1"] = Follow{FollowerID: 2, FollowingID: 1}
		follows["3:1"] = Follow{FollowerID: 3, FollowingID: 1}
		userFollowers[1] = map[int]bool{2: true, 3: true}
		userFollowing[2] = map[int]bool{1: true}
		userFollowing[3] = map[int]bool{1: true}
	}

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "get followers successfully",
			userID:         "1",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp FollowListResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Count != 2 {
					t.Errorf("Expected count 2, got %d", resp.Count)
				}
				if len(resp.Users) != 2 {
					t.Fatalf("Expected 2 users, got %d", len(resp.Users))
				}
			},
		},
		{
			name:           "no followers",
			userID:         "2",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp FollowListResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Count != 0 {
					t.Errorf("Expected count 0, got %d", resp.Count)
				}
				if len(resp.Users) != 0 {
					t.Errorf("Expected 0 users, got %d", len(resp.Users))
				}
			},
		},
		{
			name:           "user not found",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "user not found" {
					t.Errorf("Expected 'user not found', got '%s'", resp.Message)
				}
			},
		},
		{
			name:           "invalid user ID",
			userID:         "abc",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "invalid user ID" {
					t.Errorf("Expected 'invalid user ID', got '%s'", resp.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupFollowersTest()

			req := httptest.NewRequest(http.MethodGet, "/api/users/"+tt.userID+"/followers", nil)
			req.SetPathValue("userID", tt.userID)

			w := httptest.NewRecorder()
			handleGetFollowers(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandleGetFollowing(t *testing.T) {
	setupFollowingTest := func() {
		users = make(map[int]User)
		follows = make(map[string]Follow)
		userFollowers = make(map[int]map[int]bool)
		userFollowing = make(map[int]map[int]bool)

		// Create test users
		users[1] = User{ID: 1, Name: "User1", Email: "user1@test.com", Profile: "Profile1"}
		users[2] = User{ID: 2, Name: "User2", Email: "user2@test.com", Profile: "Profile2"}
		users[3] = User{ID: 3, Name: "User3", Email: "user3@test.com", Profile: "Profile3"}

		// User 1 follows User 2 and 3
		follows["1:2"] = Follow{FollowerID: 1, FollowingID: 2}
		follows["1:3"] = Follow{FollowerID: 1, FollowingID: 3}
		userFollowing[1] = map[int]bool{2: true, 3: true}
		userFollowers[2] = map[int]bool{1: true}
		userFollowers[3] = map[int]bool{1: true}
	}

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "get following successfully",
			userID:         "1",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp FollowListResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Count != 2 {
					t.Errorf("Expected count 2, got %d", resp.Count)
				}
				if len(resp.Users) != 2 {
					t.Fatalf("Expected 2 users, got %d", len(resp.Users))
				}
			},
		},
		{
			name:           "no following",
			userID:         "2",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp FollowListResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Count != 0 {
					t.Errorf("Expected count 0, got %d", resp.Count)
				}
				if len(resp.Users) != 0 {
					t.Errorf("Expected 0 users, got %d", len(resp.Users))
				}
			},
		},
		{
			name:           "user not found",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "user not found" {
					t.Errorf("Expected 'user not found', got '%s'", resp.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupFollowingTest()

			req := httptest.NewRequest(http.MethodGet, "/api/users/"+tt.userID+"/following", nil)
			req.SetPathValue("userID", tt.userID)

			w := httptest.NewRecorder()
			handleGetFollowing(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandleGetFollowStatus(t *testing.T) {
	setupStatusTest := func() {
		users = make(map[int]User)
		follows = make(map[string]Follow)
		userFollowers = make(map[int]map[int]bool)
		userFollowing = make(map[int]map[int]bool)

		users[1] = User{ID: 1, Name: "User1", Email: "user1@test.com"}
		users[2] = User{ID: 2, Name: "User2", Email: "user2@test.com"}

		// User 1 follows User 2
		follows["1:2"] = Follow{FollowerID: 1, FollowingID: 2}
		userFollowers[2] = map[int]bool{1: true}
		userFollowing[1] = map[int]bool{2: true}
	}

	tests := []struct {
		name           string
		userID         string
		followerID     string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "is following",
			userID:         "2",
			followerID:     "1",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp FollowStatusResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if !resp.IsFollowing {
					t.Errorf("Expected IsFollowing=true, got false")
				}
				if resp.FollowerID != 1 || resp.FollowingID != 2 {
					t.Errorf("Expected follower_id=1, following_id=2, got follower_id=%d, following_id=%d",
						resp.FollowerID, resp.FollowingID)
				}
			},
		},
		{
			name:           "is not following",
			userID:         "1",
			followerID:     "2",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp FollowStatusResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.IsFollowing {
					t.Errorf("Expected IsFollowing=false, got true")
				}
				if resp.FollowerID != 2 || resp.FollowingID != 1 {
					t.Errorf("Expected follower_id=2, following_id=1, got follower_id=%d, following_id=%d",
						resp.FollowerID, resp.FollowingID)
				}
			},
		},
		{
			name:           "missing follower_id parameter",
			userID:         "2",
			followerID:     "",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "follower_id query parameter is required" {
					t.Errorf("Expected 'follower_id query parameter is required', got '%s'", resp.Message)
				}
			},
		},
		{
			name:           "invalid follower_id",
			userID:         "2",
			followerID:     "abc",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "invalid follower_id" {
					t.Errorf("Expected 'invalid follower_id', got '%s'", resp.Message)
				}
			},
		},
		{
			name:           "invalid user ID",
			userID:         "xyz",
			followerID:     "1",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var resp ErrorResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if resp.Message != "invalid user ID" {
					t.Errorf("Expected 'invalid user ID', got '%s'", resp.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupStatusTest()

			url := "/api/users/" + tt.userID + "/follow-status"
			if tt.followerID != "" {
				url += "?follower_id=" + tt.followerID
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.SetPathValue("userID", tt.userID)

			w := httptest.NewRecorder()
			handleGetFollowStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}
