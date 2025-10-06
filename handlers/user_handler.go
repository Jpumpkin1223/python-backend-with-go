package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"python-backend-with-go/models"
	"python-backend-with-go/services"
	"log/slog"
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
func (h *UserHandler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.userService.Signup(req)
	if err != nil {
		switch err.Error() {
		case "name, email, and password are required":
			handleError(w, err, http.StatusBadRequest)
		case "email already exists":
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

	slog.Info("User registered successfully", "user_id", resp.UserID, "email", req.Email)
}

// HandleRoot handles the root endpoint
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "안녕하세요! Go 백엔드 서버입니다. 🚀\n")
	fmt.Fprintf(w, "요청 경로: %s\n", r.URL.Path)
	fmt.Fprintf(w, "요청 메서드: %s\n", r.Method)
}

// HandleHealth handles health check endpoint
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// HandleAPIHello handles API hello endpoint
func HandleAPIHello(w http.ResponseWriter, r *http.Request) {
	response := models.SuccessResponse{
		Message: "안녕하세요! API가 정상적으로 작동합니다.",
		Status:  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
}

// handleError sends an error response
func handleError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := models.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: err.Error(),
	}

	json.NewEncoder(w).Encode(errorResponse)
}
