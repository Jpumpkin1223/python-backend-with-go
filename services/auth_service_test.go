package services

import (
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"python-backend-with-go/models"
	"python-backend-with-go/repository"
)

func TestAuthService_Login(t *testing.T) {
	// Set JWT_SECRET for testing
	os.Setenv("JWT_SECRET", "test_secret_key_for_testing")
	defer os.Unsetenv("JWT_SECRET")

	// Setup repository and services
	userRepo := repository.NewInMemoryUserRepository()
	userService := NewUserService(userRepo)
	authService := NewAuthService(userRepo)

	// Create a test user
	signupReq := models.SignupRequest{
		Name:     "홍길동",
		Email:    "hong@test.com",
		Password: "password123",
		Profile:  "테스트 사용자",
	}
	_, err := userService.Signup(signupReq)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name        string
		request     models.LoginRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful login",
			request: models.LoginRequest{
				Email:    "hong@test.com",
				Password: "password123",
			},
			expectError: false,
		},
		{
			name: "missing email",
			request: models.LoginRequest{
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "email and password are required",
		},
		{
			name: "missing password",
			request: models.LoginRequest{
				Email: "hong@test.com",
			},
			expectError: true,
			errorMsg:    "email and password are required",
		},
		{
			name: "invalid email",
			request: models.LoginRequest{
				Email:    "nonexistent@test.com",
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "invalid email or password",
		},
		{
			name: "wrong password",
			request: models.LoginRequest{
				Email:    "hong@test.com",
				Password: "wrongpassword",
			},
			expectError: true,
			errorMsg:    "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := authService.Login(tt.request)

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
				if resp.AccessToken == "" {
					t.Errorf("Expected access token, got empty string")
				}
				if resp.UserID != 1 {
					t.Errorf("Expected user_id 1, got %d", resp.UserID)
				}
				if resp.Message != "로그인 성공" {
					t.Errorf("Expected message '로그인 성공', got '%s'", resp.Message)
				}
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	// Set JWT_SECRET for testing
	os.Setenv("JWT_SECRET", "test_secret_key_for_testing")
	defer os.Unsetenv("JWT_SECRET")

	// Setup repository and services
	userRepo := repository.NewInMemoryUserRepository()
	userService := NewUserService(userRepo)
	authService := NewAuthService(userRepo)

	// Create a test user and login to get a valid token
	signupReq := models.SignupRequest{
		Name:     "홍길동",
		Email:    "hong@test.com",
		Password: "password123",
	}
	_, err := userService.Signup(signupReq)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	loginReq := models.LoginRequest{
		Email:    "hong@test.com",
		Password: "password123",
	}
	loginResp, err := authService.Login(loginReq)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		expectError bool
		expectEmail string
		expectID    int
	}{
		{
			name:        "valid token",
			token:       loginResp.AccessToken,
			expectError: false,
			expectEmail: "hong@test.com",
			expectID:    1,
		},
		{
			name:        "invalid token",
			token:       "invalid.token.here",
			expectError: true,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "malformed token",
			token:       "not.a.jwt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authService.ValidateToken(tt.token)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if claims.Email != tt.expectEmail {
					t.Errorf("Expected email '%s', got '%s'", tt.expectEmail, claims.Email)
				}
				if claims.UserID != tt.expectID {
					t.Errorf("Expected user_id %d, got %d", tt.expectID, claims.UserID)
				}
			}
		})
	}
}

func TestAuthService_ValidateToken_WrongSecret(t *testing.T) {
	// Set JWT_SECRET for generating token
	os.Setenv("JWT_SECRET", "secret1")

	// Setup and create token
	userRepo := repository.NewInMemoryUserRepository()
	userService := NewUserService(userRepo)
	authService := NewAuthService(userRepo)

	signupReq := models.SignupRequest{
		Name:     "홍길동",
		Email:    "hong@test.com",
		Password: "password123",
	}
	_, err := userService.Signup(signupReq)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	loginReq := models.LoginRequest{
		Email:    "hong@test.com",
		Password: "password123",
	}
	loginResp, err := authService.Login(loginReq)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Change JWT_SECRET
	os.Setenv("JWT_SECRET", "secret2")
	defer os.Unsetenv("JWT_SECRET")

	// Try to validate with different secret
	_, err = authService.ValidateToken(loginResp.AccessToken)
	if err == nil {
		t.Errorf("Expected error when validating token with wrong secret, got none")
	}
}

func TestAuthService_Login_PasswordHashing(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key_for_testing")
	defer os.Unsetenv("JWT_SECRET")

	userRepo := repository.NewInMemoryUserRepository()
	userService := NewUserService(userRepo)
	authService := NewAuthService(userRepo)

	// Create user
	password := "mySecretPassword123"
	signupReq := models.SignupRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: password,
	}
	signupResp, err := userService.Signup(signupReq)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify password is hashed in repository
	user, err := userRepo.GetByID(signupResp.UserID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	// Password should be hashed (not equal to plain text)
	if user.HashedPassword == password {
		t.Errorf("Password should be hashed, but it's stored as plain text")
	}

	// Password should be valid bcrypt hash
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		t.Errorf("Password hash is not valid: %v", err)
	}

	// Login should work with correct password
	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}
	_, err = authService.Login(loginReq)
	if err != nil {
		t.Errorf("Login failed with correct password: %v", err)
	}
}

func TestAuthService_GenerateToken_NoSecret(t *testing.T) {
	// Unset JWT_SECRET
	os.Unsetenv("JWT_SECRET")

	userRepo := repository.NewInMemoryUserRepository()
	userService := NewUserService(userRepo)
	authService := NewAuthService(userRepo)

	// Create user
	signupReq := models.SignupRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := userService.Signup(signupReq)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Try to login without JWT_SECRET
	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err = authService.Login(loginReq)

	if err == nil {
		t.Errorf("Expected error when JWT_SECRET is not set, got none")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("Expected error message to mention JWT_SECRET, got: %s", err.Error())
	}
}
