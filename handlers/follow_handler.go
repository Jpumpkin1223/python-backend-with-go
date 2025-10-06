package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"python-backend-with-go/models"
	"python-backend-with-go/services"
)

// FollowHandler handles follow-related HTTP requests
type FollowHandler struct {
	followService *services.FollowService
}

// NewFollowHandler creates a new follow handler
func NewFollowHandler(followService *services.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

// HandleFollow handles follow requests
func (h *FollowHandler) HandleFollow(w http.ResponseWriter, r *http.Request) {
	var req models.FollowRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Get following ID from URL path
	followingIDStr := r.PathValue("userID")
	followingID := 0
	if _, err := fmt.Sscanf(followingIDStr, "%d", &followingID); err != nil {
		handleError(w, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.followService.Follow(req.FollowerID, followingID)
	if err != nil {
		switch err.Error() {
		case "follower_id and following user ID are required":
			handleError(w, err, http.StatusBadRequest)
		case "cannot follow yourself":
			handleError(w, err, http.StatusBadRequest)
		case "follower user not found", "following user not found":
			handleError(w, err, http.StatusNotFound)
		case "already following this user":
			handleError(w, err, http.StatusConflict)
		default:
			handleError(w, err, http.StatusInternalServerError)
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

	slog.Info("Follow created", "follower_id", req.FollowerID, "following_id", followingID)
}

// HandleUnfollow handles unfollow requests
func (h *FollowHandler) HandleUnfollow(w http.ResponseWriter, r *http.Request) {
	var req models.FollowRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Get following ID from URL path
	followingIDStr := r.PathValue("userID")
	followingID := 0
	if _, err := fmt.Sscanf(followingIDStr, "%d", &followingID); err != nil {
		handleError(w, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.followService.Unfollow(req.FollowerID, followingID)
	if err != nil {
		switch err.Error() {
		case "follower_id and following user ID are required":
			handleError(w, err, http.StatusBadRequest)
		case "follow relationship not found":
			handleError(w, err, http.StatusNotFound)
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

	slog.Info("Follow deleted", "follower_id", req.FollowerID, "following_id", followingID)
}

// HandleGetFollowers handles get followers requests
func (h *FollowHandler) HandleGetFollowers(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL path
	userIDStr := r.PathValue("userID")
	userID := 0
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		handleError(w, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.followService.GetFollowers(userID)
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

	slog.Info("Followers retrieved", "user_id", userID, "count", resp.Count)
}

// HandleGetFollowing handles get following requests
func (h *FollowHandler) HandleGetFollowing(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL path
	userIDStr := r.PathValue("userID")
	userID := 0
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		handleError(w, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.followService.GetFollowing(userID)
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

	slog.Info("Following retrieved", "user_id", userID, "count", resp.Count)
}

// HandleGetFollowStatus handles get follow status requests
func (h *FollowHandler) HandleGetFollowStatus(w http.ResponseWriter, r *http.Request) {
	// Get following ID from URL path
	followingIDStr := r.PathValue("userID")
	followingID := 0
	if _, err := fmt.Sscanf(followingIDStr, "%d", &followingID); err != nil {
		handleError(w, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Get follower ID from query parameter
	followerIDStr := r.URL.Query().Get("follower_id")
	if followerIDStr == "" {
		handleError(w, fmt.Errorf("follower_id query parameter is required"), http.StatusBadRequest)
		return
	}

	followerID := 0
	if _, err := fmt.Sscanf(followerIDStr, "%d", &followerID); err != nil {
		handleError(w, fmt.Errorf("invalid follower_id"), http.StatusBadRequest)
		return
	}

	// Call service
	resp := h.followService.GetFollowStatus(followerID, followingID)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Follow status checked", "follower_id", followerID, "following_id", followingID, "is_following", resp.IsFollowing)
}
