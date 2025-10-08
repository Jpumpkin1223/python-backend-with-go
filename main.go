package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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
	authService := services.NewAuthService(userRepo)
	followService := services.NewFollowService(followRepo, userRepo)
	postService := services.NewPostService(postRepo, userRepo, followRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(authService)
	followHandler := handlers.NewFollowHandler(followService)
	postHandler := handlers.NewPostHandler(postService)

	// Create Gin engine with middleware
	router := gin.New()

	// Apply global middleware
	router.Use(handlers.GinLoggingMiddleware())
	router.Use(handlers.GinRecoveryMiddleware())
	router.Use(handlers.GinCORSMiddleware())
	router.Use(handlers.GinSecurityHeadersMiddleware())
	// Note: Auth middleware is applied per route group for protected routes

	// Register routes
	// Public routes
	router.GET("/", handlers.HandleRoot)
	router.GET("/health", handlers.HandleHealth)
	router.GET("/api/hello", handlers.HandleAPIHello)
	router.POST("/api/signup", userHandler.HandleSignup)
	router.POST("/api/login", authHandler.HandleLogin)

	// Public read-only routes (팔로우/게시글 조회는 공개)
	router.GET("/api/users/:userID/followers", followHandler.HandleGetFollowers)
	router.GET("/api/users/:userID/following", followHandler.HandleGetFollowing)
	router.GET("/api/users/:userID/posts", postHandler.HandleGetUserPosts)

	// Protected routes (require authentication)
	protected := router.Group("/")
	protected.Use(handlers.AuthMiddlewareGin(authService))
	{
		// Follow/Unfollow routes (작성/수정/삭제는 인증 필요)
		protected.POST("/api/users/:userID/follow", followHandler.HandleFollow)
		protected.DELETE("/api/users/:userID/follow", followHandler.HandleUnfollow)
		protected.GET("/api/users/:userID/follow-status", followHandler.HandleGetFollowStatus)

		// Post routes (작성/수정/삭제는 인증 필요)
		protected.POST("/api/posts", postHandler.HandleCreatePost)
		protected.PUT("/api/posts/:postID", postHandler.HandleUpdatePost)
		protected.DELETE("/api/posts/:postID", postHandler.HandleDeletePost)

		// Timeline routes (인증 필요 - 개인화된 콘텐츠)
		protected.GET("/api/users/:userID/timeline", postHandler.HandleGetTimeline)
	}

	// Create HTTP server with timeouts (wrapping Gin)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
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
