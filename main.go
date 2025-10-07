package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"python-backend-with-go/db"
	"python-backend-with-go/handlers"
	"python-backend-with-go/repository"
	"python-backend-with-go/services"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, using environment variables")
	}

	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize database
	if err := db.InitDatabase(); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.CloseDatabase()

	// Get port from environment variable (default: 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize repositories (now using GORM)
	userRepo := repository.NewGormUserRepository(db.DB)
	followRepo := repository.NewGormFollowRepository(db.DB)
	postRepo := repository.NewGormPostRepository(db.DB)

	// Initialize services
	userService := services.NewUserService(userRepo)
	followService := services.NewFollowService(followRepo, userRepo)
	postService := services.NewPostService(postRepo, userRepo, followRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	followHandler := handlers.NewFollowHandler(followService)
	postHandler := handlers.NewPostHandler(postService)

	// Create new ServeMux (Go 1.22+ with enhanced routing)
	mux := http.NewServeMux()

	// Register routes with method-specific handlers
	mux.HandleFunc("GET /", handlers.HandleRoot)
	mux.HandleFunc("GET /health", handlers.HandleHealth)
	mux.HandleFunc("GET /api/hello", handlers.HandleAPIHello)
	mux.HandleFunc("POST /api/signup", userHandler.HandleSignup)

	// Follow/Unfollow routes
	mux.HandleFunc("POST /api/users/{userID}/follow", followHandler.HandleFollow)
	mux.HandleFunc("DELETE /api/users/{userID}/follow", followHandler.HandleUnfollow)
	mux.HandleFunc("GET /api/users/{userID}/followers", followHandler.HandleGetFollowers)
	mux.HandleFunc("GET /api/users/{userID}/following", followHandler.HandleGetFollowing)
	mux.HandleFunc("GET /api/users/{userID}/follow-status", followHandler.HandleGetFollowStatus)

	// Post routes
	mux.HandleFunc("POST /api/posts", postHandler.HandleCreatePost)
	mux.HandleFunc("PUT /api/posts/{postID}", postHandler.HandleUpdatePost)
	mux.HandleFunc("DELETE /api/posts/{postID}", postHandler.HandleDeletePost)
	mux.HandleFunc("GET /api/users/{userID}/posts", postHandler.HandleGetUserPosts)
	mux.HandleFunc("GET /api/users/{userID}/timeline", postHandler.HandleGetTimeline)

	// Apply middleware chain
	handler := handlers.LoggingMiddleware(
		handlers.RecoveryMiddleware(
			handlers.CORSMiddleware(
				handlers.SecurityHeadersMiddleware(mux),
			),
		),
	)

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
