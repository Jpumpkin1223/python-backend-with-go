package models

import "time"

// User represents a user in the system
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Profile  string `json:"profile"`
}

// Follow represents a follow relationship
type Follow struct {
	FollowerID  int       `json:"follower_id"`
	FollowingID int       `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// SignupRequest represents the signup request body
type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Profile  string `json:"profile"`
}

// SignupResponse represents the signup response
type SignupResponse struct {
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
}

// FollowRequest represents the follow request body
type FollowRequest struct {
	FollowerID int `json:"follower_id"`
}

// FollowResponse represents the follow response
type FollowResponse struct {
	Message     string `json:"message"`
	FollowerID  int    `json:"follower_id"`
	FollowingID int    `json:"following_id"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// UserInfo represents basic user information for follow lists
type UserInfo struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Profile string `json:"profile"`
}

// FollowListResponse represents followers/following list response
type FollowListResponse struct {
	Users []UserInfo `json:"users"`
	Count int        `json:"count"`
}

// FollowStatusResponse represents follow status response
type FollowStatusResponse struct {
	IsFollowing bool `json:"is_following"`
	FollowerID  int  `json:"follower_id"`
	FollowingID int  `json:"following_id"`
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}
