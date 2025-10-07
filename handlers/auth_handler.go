package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

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
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.authService.Login(req)
	if err != nil {
		switch err.Error() {
		case "email and password are required":
			handleError(w, err, http.StatusBadRequest)
		case "invalid email or password":
			handleError(w, err, http.StatusUnauthorized)
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

	slog.Info("User logged in successfully", "user_id", resp.UserID, "email", req.Email)
}
