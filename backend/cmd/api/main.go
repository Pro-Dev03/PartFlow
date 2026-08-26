package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/partflow/smart-store/internal/auth"
	"github.com/partflow/smart-store/internal/inventory"
	"github.com/partflow/smart-store/internal/api"

	"github.com/partflow/smart-store/pkg/config"
	"github.com/partflow/smart-store/pkg/database"
	"github.com/partflow/smart-store/pkg/errors"
	"github.com/partflow/smart-store/pkg/health"
	"github.com/partflow/smart-store/pkg/logger"
	"github.com/partflow/smart-store/pkg/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize structured logger
	logConfig := logger.DefaultConfig()
	logConfig.Level = cfg.LogLevel
	logConfig.EnableConsole = true
	logConfig.EnableFile = true
	logConfig.EnableCaller = true
	logConfig.TimeFormat = time.RFC3339
	if err := logger.Initialize(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	logger.Info("Starting PartFlow API Server...")

	// Initialize database
	if err := database.Initialize(); err != nil {
		logger.Fatal("Failed to initialize database", err)
	}
	defer database.Close()

	// Set Gin mode
	gin.SetMode(cfg.ServerMode)

	// Create router
	router := gin.New()

	// Register middleware
	router.Use(middleware.CORS())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.ErrorLoggingMiddleware())
	router.Use(errors.ErrorHandlerMiddleware())
	router.Use(middleware.RequestID())
	router.Use(middleware.SecurityLoggingMiddleware())
	router.Use(middleware.PerformanceLoggingMiddleware())
	if cfg.RateLimitEnabled {
		router.Use(middleware.RateLimiter())
	}

	// Set JWT secret
	middleware.SetJWTSecret(cfg.JWTSecret)

	// Set disable auth flag for development
	middleware.SetDisableAuth(cfg.DisableAuth)

	// Initialize services
	db := database.GetDB()

	// Auth service
	authService, err := auth.NewService(db, cfg.JWTSecret, cfg.UseSupabaseAuth, cfg.SupabaseURL, cfg.SupabaseKey)
	if err != nil {
		logger.Fatal("Failed to initialize auth service", err)
	}

	// Inventory service initialization
	_ = inventory.NewRepository(db) // Initialize inventory repository

	// Health checker
	healthChecker := health.NewHealthChecker(db)
	
	// Register health check routes
	router.GET("/health", healthChecker.Check)
	router.GET("/ready", healthChecker.Readiness)
	router.GET("/alive", healthChecker.Liveness)

	// Log application startup
	logger.LogSystemEvent("application_start", "api", map[string]interface{}{
		"server_mode": cfg.ServerMode,
		"port": cfg.ServerPort,
		"rate_limit_enabled": cfg.RateLimitEnabled,
		"architecture": "Centralized routing inspired by Fynexa",
	})

	// Register API routes using centralized router (inspired by Fynexa architecture)
	// This provides better maintainability and clearer structure
	middleware.SetDatabase(db)
	api.SetupRoutes(router, db, authService)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting...", map[string]interface{}{"port": cfg.ServerPort})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", err)
	}

	logger.Info("Server exited")
}
