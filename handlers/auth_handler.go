package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"python-backend-with-go/models"
	"python-backend-with-go/services"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// HandleLogin handles user login
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req models.LoginRequest

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		handleErrorGin(c, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.authService.Login(req)
	if err != nil {
		switch err.Error() {
		case "email and password are required":
			handleErrorGin(c, err, http.StatusBadRequest)
		case "invalid email or password":
			handleErrorGin(c, err, http.StatusUnauthorized)
		default:
			handleErrorGin(c, err, http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	c.JSON(http.StatusOK, resp)
	slog.Info("User logged in successfully", "user_id", resp.UserID, "email", req.Email)
}
