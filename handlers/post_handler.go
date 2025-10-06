package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"python-backend-with-go/models"
	"python-backend-with-go/services"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	postService *services.PostService
}

// NewPostHandler creates a new post handler
func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// HandleCreatePost handles create post requests
func (h *PostHandler) HandleCreatePost(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePostRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.postService.CreatePost(req)
	if err != nil {
		switch err.Error() {
		case "user_id is required":
			handleError(w, err, http.StatusBadRequest)
		case "user not found":
			handleError(w, err, http.StatusNotFound)
		case "content is required":
			handleError(w, err, http.StatusBadRequest)
		case fmt.Sprintf("content must be %d characters or less", services.MaxPostContentLength):
			handleError(w, err, http.StatusBadRequest)
		default:
			if err.Error() == fmt.Sprintf("content must be %d characters or less", services.MaxPostContentLength) {
				handleError(w, err, http.StatusBadRequest)
			} else {
				handleError(w, err, http.StatusInternalServerError)
			}
		}
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Post created", "post_id", resp.PostID, "user_id", req.UserID)
}

// HandleUpdatePost handles update post requests
func (h *PostHandler) HandleUpdatePost(w http.ResponseWriter, r *http.Request) {
	var req models.UpdatePostRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Get post ID from URL path
	postIDStr := r.PathValue("postID")
	postID := 0
	if _, err := fmt.Sscanf(postIDStr, "%d", &postID); err != nil {
		handleError(w, fmt.Errorf("invalid post ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.postService.UpdatePost(postID, req)
	if err != nil {
		switch err.Error() {
		case "post_id and user_id are required":
			handleError(w, err, http.StatusBadRequest)
		case "content is required":
			handleError(w, err, http.StatusBadRequest)
		case "post not found":
			handleError(w, err, http.StatusNotFound)
		case "unauthorized to update this post":
			handleError(w, err, http.StatusForbidden)
		default:
			if err.Error() == fmt.Sprintf("content must be %d characters or less", services.MaxPostContentLength) {
				handleError(w, err, http.StatusBadRequest)
			} else {
				handleError(w, err, http.StatusInternalServerError)
			}
		}
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Post updated", "post_id", postID, "user_id", req.UserID)
}

// HandleDeletePost handles delete post requests
func (h *PostHandler) HandleDeletePost(w http.ResponseWriter, r *http.Request) {
	var req models.DeletePostRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Get post ID from URL path
	postIDStr := r.PathValue("postID")
	postID := 0
	if _, err := fmt.Sscanf(postIDStr, "%d", &postID); err != nil {
		handleError(w, fmt.Errorf("invalid post ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.postService.DeletePost(postID, req.UserID)
	if err != nil {
		switch err.Error() {
		case "post_id and user_id are required":
			handleError(w, err, http.StatusBadRequest)
		case "post not found":
			handleError(w, err, http.StatusNotFound)
		case "unauthorized to delete this post":
			handleError(w, err, http.StatusForbidden)
		default:
			handleError(w, err, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Post deleted", "post_id", postID, "user_id", req.UserID)
}

// HandleGetUserPosts handles get user posts requests
func (h *PostHandler) HandleGetUserPosts(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL path
	userIDStr := r.PathValue("userID")
	userID := 0
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		handleError(w, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.postService.GetUserPosts(userID)
	if err != nil {
		if err.Error() == "user not found" {
			handleError(w, err, http.StatusNotFound)
		} else {
			handleError(w, err, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("User posts retrieved", "user_id", userID, "count", resp.Count)
}

// HandleGetTimeline handles get timeline requests
func (h *PostHandler) HandleGetTimeline(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL path
	userIDStr := r.PathValue("userID")
	userID := 0
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		handleError(w, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.postService.GetTimeline(userID)
	if err != nil {
		if err.Error() == "user not found" {
			handleError(w, err, http.StatusNotFound)
		} else {
			handleError(w, err, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Timeline retrieved", "user_id", userID, "count", resp.Count)
}
