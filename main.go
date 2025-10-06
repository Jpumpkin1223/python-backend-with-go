package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
)

var (
	// In-memory user storage (use database in production)
	users      = make(map[int]User)
	nextUserID = 1

	// In-memory follow relationships storage
	follows        = make(map[string]Follow)        // key: "followerID:followingID"
	userFollowers  = make(map[int]map[int]bool)     // key: userID, value: set of follower IDs
	userFollowing  = make(map[int]map[int]bool)     // key: userID, value: set of following IDs
)

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

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Get port from environment variable (default: 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create new ServeMux (Go 1.22+ with enhanced routing)
	mux := http.NewServeMux()

	// Register routes with method-specific handlers
	mux.HandleFunc("GET /", handleRoot)
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /api/hello", handleAPIHello)
	mux.HandleFunc("POST /api/signup", handleSignup)

	// Follow/Unfollow routes
	mux.HandleFunc("POST /api/users/{userID}/follow", handleFollow)
	mux.HandleFunc("DELETE /api/users/{userID}/follow", handleUnfollow)
	mux.HandleFunc("GET /api/users/{userID}/followers", handleGetFollowers)
	mux.HandleFunc("GET /api/users/{userID}/following", handleGetFollowing)
	mux.HandleFunc("GET /api/users/{userID}/follow-status", handleGetFollowStatus)

	// Apply middleware chain
	handler := loggingMiddleware(recoveryMiddleware(corsMiddleware(securityHeadersMiddleware(mux))))

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited gracefully")
}

// Handlers

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "안녕하세요! Go 백엔드 서버입니다. 🚀\n")
	fmt.Fprintf(w, "요청 경로: %s\n", r.URL.Path)
	fmt.Fprintf(w, "요청 메서드: %s\n", r.Method)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func handleAPIHello(w http.ResponseWriter, r *http.Request) {
	response := SuccessResponse{
		Message: "안녕하세요! API가 정상적으로 작동합니다.",
		Status:  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.Email == "" || req.Password == "" {
		handleError(w, fmt.Errorf("name, email, and password are required"), http.StatusBadRequest)
		return
	}

	// Check if email already exists
	for _, user := range users {
		if user.Email == req.Email {
			handleError(w, fmt.Errorf("email already exists"), http.StatusConflict)
			return
		}
	}

	// Create new user
	userID := nextUserID
	nextUserID++

	newUser := User{
		ID:       userID,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Profile:  req.Profile,
	}

	// Store user
	users[userID] = newUser

	// Return success response
	response := SignupResponse{
		Message: "회원가입이 완료되었습니다.",
		UserID:  userID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("User registered successfully", "user_id", userID, "email", req.Email)
}

func handleFollow(w http.ResponseWriter, r *http.Request) {
	var req FollowRequest

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

	followerID := req.FollowerID

	// Validate required fields
	if followerID == 0 || followingID == 0 {
		handleError(w, fmt.Errorf("follower_id and following user ID are required"), http.StatusBadRequest)
		return
	}

	// Check if trying to follow themselves
	if followerID == followingID {
		handleError(w, fmt.Errorf("cannot follow yourself"), http.StatusBadRequest)
		return
	}

	// Check if both users exist
	if _, exists := users[followerID]; !exists {
		handleError(w, fmt.Errorf("follower user not found"), http.StatusNotFound)
		return
	}
	if _, exists := users[followingID]; !exists {
		handleError(w, fmt.Errorf("following user not found"), http.StatusNotFound)
		return
	}

	// Check if already following
	followKey := fmt.Sprintf("%d:%d", followerID, followingID)
	if _, exists := follows[followKey]; exists {
		handleError(w, fmt.Errorf("already following this user"), http.StatusConflict)
		return
	}

	// Create follow relationship
	now := time.Now()
	follow := Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
		CreatedAt:   now,
	}

	// Store follow relationship
	follows[followKey] = follow

	// Update indexes
	if userFollowers[followingID] == nil {
		userFollowers[followingID] = make(map[int]bool)
	}
	userFollowers[followingID][followerID] = true

	if userFollowing[followerID] == nil {
		userFollowing[followerID] = make(map[int]bool)
	}
	userFollowing[followerID][followingID] = true

	// Return success response
	response := FollowResponse{
		Message:     "팔로우 성공",
		FollowerID:  followerID,
		FollowingID: followingID,
		CreatedAt:   now.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Follow created", "follower_id", followerID, "following_id", followingID)
}

func handleUnfollow(w http.ResponseWriter, r *http.Request) {
	var req FollowRequest

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

	followerID := req.FollowerID

	// Validate required fields
	if followerID == 0 || followingID == 0 {
		handleError(w, fmt.Errorf("follower_id and following user ID are required"), http.StatusBadRequest)
		return
	}

	// Check if follow relationship exists
	followKey := fmt.Sprintf("%d:%d", followerID, followingID)
	if _, exists := follows[followKey]; !exists {
		handleError(w, fmt.Errorf("follow relationship not found"), http.StatusNotFound)
		return
	}

	// Delete follow relationship
	delete(follows, followKey)

	// Update indexes
	if userFollowers[followingID] != nil {
		delete(userFollowers[followingID], followerID)
	}
	if userFollowing[followerID] != nil {
		delete(userFollowing[followerID], followingID)
	}

	// Return success response
	response := FollowResponse{
		Message:     "언팔로우 성공",
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Follow deleted", "follower_id", followerID, "following_id", followingID)
}

func handleGetFollowers(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL path
	userIDStr := r.PathValue("userID")
	userID := 0
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		handleError(w, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Check if user exists
	if _, exists := users[userID]; !exists {
		handleError(w, fmt.Errorf("user not found"), http.StatusNotFound)
		return
	}

	// Get followers
	followerIDs := userFollowers[userID]
	userInfos := make([]UserInfo, 0, len(followerIDs))

	for followerID := range followerIDs {
		if user, exists := users[followerID]; exists {
			userInfos = append(userInfos, UserInfo{
				ID:      user.ID,
				Name:    user.Name,
				Email:   user.Email,
				Profile: user.Profile,
			})
		}
	}

	response := FollowListResponse{
		Users: userInfos,
		Count: len(userInfos),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Followers retrieved", "user_id", userID, "count", len(userInfos))
}

func handleGetFollowing(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL path
	userIDStr := r.PathValue("userID")
	userID := 0
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		handleError(w, fmt.Errorf("invalid user ID"), http.StatusBadRequest)
		return
	}

	// Check if user exists
	if _, exists := users[userID]; !exists {
		handleError(w, fmt.Errorf("user not found"), http.StatusNotFound)
		return
	}

	// Get following
	followingIDs := userFollowing[userID]
	userInfos := make([]UserInfo, 0, len(followingIDs))

	for followingID := range followingIDs {
		if user, exists := users[followingID]; exists {
			userInfos = append(userInfos, UserInfo{
				ID:      user.ID,
				Name:    user.Name,
				Email:   user.Email,
				Profile: user.Profile,
			})
		}
	}

	response := FollowListResponse{
		Users: userInfos,
		Count: len(userInfos),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Following retrieved", "user_id", userID, "count", len(userInfos))
}

func handleGetFollowStatus(w http.ResponseWriter, r *http.Request) {
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

	// Check if follow relationship exists
	followKey := fmt.Sprintf("%d:%d", followerID, followingID)
	_, isFollowing := follows[followKey]

	response := FollowStatusResponse{
		IsFollowing: isFollowing,
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Follow status checked", "follower_id", followerID, "following_id", followingID, "is_following", isFollowing)
}

// Middleware

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate request ID
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		// Create response writer wrapper to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		slog.Info("Request completed",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := r.Context().Value("request_id").(string)
				slog.Error("Panic recovered",
					"request_id", requestID,
					"error", err,
					"path", r.URL.Path,
				)

				handleError(w, fmt.Errorf("internal server error"), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// Helper types and functions

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func handleError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: err.Error(),
	}

	json.NewEncoder(w).Encode(errorResponse)
}
