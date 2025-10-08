package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"python-backend-with-go/models"
	"python-backend-with-go/services"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// HandleSignup handles user registration
func (h *UserHandler) HandleSignup(c *gin.Context) {
	var req models.SignupRequest

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		handleErrorGin(c, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.userService.Signup(req)
	if err != nil {
		switch err.Error() {
		case "name, email, and password are required":
			handleErrorGin(c, err, http.StatusBadRequest)
		case "email already exists":
			handleErrorGin(c, err, http.StatusConflict)
		default:
			handleErrorGin(c, err, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, resp)
	slog.Info("User registered successfully", "user_id", resp.UserID, "email", req.Email)
}

// HandleRoot handles the root endpoint
func HandleRoot(c *gin.Context) {
	c.String(http.StatusOK, "안녕하세요! Go 백엔드 서버입니다. 🚀\n요청 경로: %s\n요청 메서드: %s\n", c.Request.URL.Path, c.Request.Method)
}

// HandleHealth handles health check endpoint
func HandleHealth(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// HandleAPIHello handles API hello endpoint
func HandleAPIHello(c *gin.Context) {
	response := models.SuccessResponse{
		Message: "안녕하세요! API가 정상적으로 작동합니다.",
		Status:  "success",
	}

	c.JSON(http.StatusOK, response)
}

// HandleError sends an error response (net/http version - kept for reference)
func HandleError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Note: This function is kept for reference but not used in Gin handlers
	// Use handleErrorGin for Gin handlers instead
}
