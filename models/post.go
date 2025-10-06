package models

import "time"

// Post represents a post in the system
type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreatePostRequest represents the create post request body
type CreatePostRequest struct {
	UserID  int    `json:"user_id"`
	Content string `json:"content"`
}

// CreatePostResponse represents the create post response
type CreatePostResponse struct {
	Message string `json:"message"`
	PostID  int    `json:"post_id"`
	Post    Post   `json:"post"`
}

// UpdatePostRequest represents the update post request body
type UpdatePostRequest struct {
	UserID  int    `json:"user_id"`
	Content string `json:"content"`
}

// UpdatePostResponse represents the update post response
type UpdatePostResponse struct {
	Message string `json:"message"`
	PostID  int    `json:"post_id"`
	Post    Post   `json:"post"`
}

// DeletePostRequest represents the delete post request body
type DeletePostRequest struct {
	UserID int `json:"user_id"`
}

// DeletePostResponse represents the delete post response
type DeletePostResponse struct {
	Message string `json:"message"`
	PostID  int    `json:"post_id"`
}

// PostWithUser represents a post with user information
type PostWithUser struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	UserName  string    `json:"user_name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TimelineResponse represents timeline response
type TimelineResponse struct {
	Posts []PostWithUser `json:"posts"`
	Count int            `json:"count"`
}

// UserPostsResponse represents user posts response
type UserPostsResponse struct {
	Posts []Post `json:"posts"`
	Count int    `json:"count"`
}
