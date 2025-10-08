package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
func (h *FollowHandler) HandleFollow(c *gin.Context) {
	var req models.FollowRequest

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		handleErrorGin(c, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Get following ID from URL path
	followingIDStr := c.Param("userID")
	followingID, err := strconv.Atoi(followingIDStr)
	if err != nil {
		handleErrorGin(c, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.followService.Follow(req.FollowerID, followingID)
	if err != nil {
		switch err.Error() {
		case "follower_id and following user ID are required":
			handleErrorGin(c, err, http.StatusBadRequest)
		case "cannot follow yourself":
			handleErrorGin(c, err, http.StatusBadRequest)
		case "follower user not found", "following user not found":
			handleErrorGin(c, err, http.StatusNotFound)
		case "already following this user":
			handleErrorGin(c, err, http.StatusConflict)
		default:
			handleErrorGin(c, err, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, resp)
	slog.Info("Follow created", "follower_id", req.FollowerID, "following_id", followingID)
}

// HandleUnfollow handles unfollow requests
func (h *FollowHandler) HandleUnfollow(c *gin.Context) {
	var req models.FollowRequest

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		handleErrorGin(c, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Get following ID from URL path
	followingIDStr := c.Param("userID")
	followingID, err := strconv.Atoi(followingIDStr)
	if err != nil {
		handleErrorGin(c, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.followService.Unfollow(req.FollowerID, followingID)
	if err != nil {
		switch err.Error() {
		case "follower_id and following user ID are required":
			handleErrorGin(c, err, http.StatusBadRequest)
		case "follow relationship not found":
			handleErrorGin(c, err, http.StatusNotFound)
		default:
			handleErrorGin(c, err, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, resp)
	slog.Info("Follow deleted", "follower_id", req.FollowerID, "following_id", followingID)
}

// HandleGetFollowers handles get followers requests
func (h *FollowHandler) HandleGetFollowers(c *gin.Context) {
	// Get user ID from URL path
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		handleErrorGin(c, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.followService.GetFollowers(userID)
	if err != nil {
		if err.Error() == "user not found" {
			handleErrorGin(c, err, http.StatusNotFound)
		} else {
			handleErrorGin(c, err, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, resp)
	slog.Info("Followers retrieved", "user_id", userID, "count", resp.Count)
}

// HandleGetFollowing handles get following requests
func (h *FollowHandler) HandleGetFollowing(c *gin.Context) {
	// Get user ID from URL path
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		handleErrorGin(c, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.followService.GetFollowing(userID)
	if err != nil {
		if err.Error() == "user not found" {
			handleErrorGin(c, err, http.StatusNotFound)
		} else {
			handleErrorGin(c, err, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, resp)
	slog.Info("Following retrieved", "user_id", userID, "count", resp.Count)
}

// HandleGetFollowStatus handles get follow status requests
func (h *FollowHandler) HandleGetFollowStatus(c *gin.Context) {
	// Get following ID from URL path
	followingIDStr := c.Param("userID")
	followingID, err := strconv.Atoi(followingIDStr)
	if err != nil {
		handleErrorGin(c, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Get follower ID from query parameter
	followerIDStr := c.Query("follower_id")
	if followerIDStr == "" {
		handleErrorGin(c, fmt.Errorf("follower_id query parameter is required"), http.StatusBadRequest)
		return
	}

	followerID, err := strconv.Atoi(followerIDStr)
	if err != nil {
		handleErrorGin(c, fmt.Errorf("invalid follower_id"), http.StatusBadRequest)
		return
	}

	// Call service
	resp := h.followService.GetFollowStatus(followerID, followingID)

	// Return success response
	c.JSON(http.StatusOK, resp)
	slog.Info("Follow status checked", "follower_id", followerID, "following_id", followingID, "is_following", resp.IsFollowing)
}
