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
func (h *PostHandler) HandleCreatePost(c *gin.Context) {
	var req models.CreatePostRequest

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Failed to decode create post request", "error", err)
		handleErrorGin(c, fmt.Errorf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.postService.CreatePost(req)
	if err != nil {
		switch err.Error() {
		case "user_id is required":
			handleErrorGin(c, err, http.StatusBadRequest)
		case "user not found":
			handleErrorGin(c, err, http.StatusNotFound)
		case "content is required":
			handleErrorGin(c, err, http.StatusBadRequest)
		case fmt.Sprintf("content must be %d characters or less", services.MaxPostContentLength):
			handleErrorGin(c, err, http.StatusBadRequest)
		default:
			if err.Error() == fmt.Sprintf("content must be %d characters or less", services.MaxPostContentLength) {
				handleErrorGin(c, err, http.StatusBadRequest)
			} else {
				handleErrorGin(c, err, http.StatusInternalServerError)
			}
		}
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, resp)
	slog.Info("Post created", "post_id", resp.PostID, "user_id", req.UserID)
}

// HandleUpdatePost handles update post requests
func (h *PostHandler) HandleUpdatePost(c *gin.Context) {
	var req models.UpdatePostRequest

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		handleErrorGin(c, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Get post ID from URL path
	postIDStr := c.Param("postID")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		handleErrorGin(c, fmt.Errorf("invalid post ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.postService.UpdatePost(postID, req)
	if err != nil {
		switch err.Error() {
		case "post_id and user_id are required":
			handleErrorGin(c, err, http.StatusBadRequest)
		case "content is required":
			handleErrorGin(c, err, http.StatusBadRequest)
		case "post not found":
			handleErrorGin(c, err, http.StatusNotFound)
		case "unauthorized to update this post":
			handleErrorGin(c, err, http.StatusForbidden)
		default:
			if err.Error() == fmt.Sprintf("content must be %d characters or less", services.MaxPostContentLength) {
				handleErrorGin(c, err, http.StatusBadRequest)
			} else {
				handleErrorGin(c, err, http.StatusInternalServerError)
			}
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, resp)
	slog.Info("Post updated", "post_id", postID, "user_id", req.UserID)
}

// HandleDeletePost handles delete post requests
func (h *PostHandler) HandleDeletePost(c *gin.Context) {
	var req models.DeletePostRequest

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		handleErrorGin(c, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Get post ID from URL path
	postIDStr := c.Param("postID")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		handleErrorGin(c, fmt.Errorf("invalid post ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.postService.DeletePost(postID, req.UserID)
	if err != nil {
		switch err.Error() {
		case "post_id and user_id are required":
			handleErrorGin(c, err, http.StatusBadRequest)
		case "post not found":
			handleErrorGin(c, err, http.StatusNotFound)
		case "unauthorized to delete this post":
			handleErrorGin(c, err, http.StatusForbidden)
		default:
			handleErrorGin(c, err, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, resp)
	slog.Info("Post deleted", "post_id", postID, "user_id", req.UserID)
}

// HandleGetUserPosts handles get user posts requests
func (h *PostHandler) HandleGetUserPosts(c *gin.Context) {
	// Get user ID from URL path
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		handleErrorGin(c, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.postService.GetUserPosts(userID)
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
	slog.Info("User posts retrieved", "user_id", userID, "count", resp.Count)
}

// HandleGetTimeline handles get timeline requests
func (h *PostHandler) HandleGetTimeline(c *gin.Context) {
	// Get user ID from URL path
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		handleErrorGin(c, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.postService.GetTimeline(userID)
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
	slog.Info("Timeline retrieved", "user_id", userID, "count", resp.Count)
}
