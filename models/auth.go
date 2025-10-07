package models

import "github.com/golang-jwt/jwt/v5"

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	UserID      int    `json:"user_id"`
}

// Claims represents JWT claims
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
