package services

import (
	"testing"

	"python-backend-with-go/models"
	"python-backend-with-go/repository"
)

func TestUserService_Signup(t *testing.T) {
	tests := []struct {
		name        string
		request     models.SignupRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful signup",
			request: models.SignupRequest{
				Name:     "홍길동",
				Email:    "hong@example.com",
				Password: "password123",
				Profile:  "안녕하세요",
			},
			expectError: false,
		},
		{
			name: "missing name",
			request: models.SignupRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "name, email, and password are required",
		},
		{
			name: "missing email",
			request: models.SignupRequest{
				Name:     "홍길동",
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "name, email, and password are required",
		},
		{
			name: "missing password",
			request: models.SignupRequest{
				Name:  "홍길동",
				Email: "hong@example.com",
			},
			expectError: true,
			errorMsg:    "name, email, and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			userRepo := repository.NewInMemoryUserRepository()
			userService := NewUserService(userRepo)

			// Execute
			resp, err := userService.Signup(tt.request)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.UserID != 1 {
					t.Errorf("Expected user_id 1, got %d", resp.UserID)
				}
			}
		})
	}
}

func TestUserService_Signup_DuplicateEmail(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	userService := NewUserService(userRepo)

	// Create first user
	req1 := models.SignupRequest{
		Name:     "User1",
		Email:    "duplicate@example.com",
		Password: "password123",
	}
	_, err := userService.Signup(req1)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// Try to create user with same email
	req2 := models.SignupRequest{
		Name:     "User2",
		Email:    "duplicate@example.com",
		Password: "password456",
	}
	_, err = userService.Signup(req2)

	if err == nil {
		t.Error("Expected error for duplicate email, got none")
	} else if err.Error() != "email already exists" {
		t.Errorf("Expected 'email already exists', got '%s'", err.Error())
	}
}
