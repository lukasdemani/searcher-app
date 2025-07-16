package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"searcher-app/internal/database"
	"searcher-app/internal/handlers"
	"searcher-app/internal/middleware"
	"searcher-app/internal/repository"
	"searcher-app/internal/services"
	"searcher-app/internal/worker"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt gets environment variable as int with fallback
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func main() {
	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create database configuration from environment variables with fallbacks
	dbConfig := &database.DatabaseConfig{
		Host:     getEnv("DB_HOST", "127.0.0.1"),
		Port:     getEnvInt("DB_PORT", 3306),
		Username: getEnv("DB_USER", "analyzer_user"),
		Password: getEnv("DB_PASSWORD", "analyzer_pass"),
		Database: getEnv("DB_NAME", "website_analyzer"),
	}

	// Enhanced database connection with pooling
	db, err := database.NewDatabase(dbConfig, logger)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Enhanced database connection established successfully")

	// Initialize repositories
	urlRepo := repository.NewMySQLURLRepository(db.DB)

	// Initialize worker pool for controlled concurrency
	workerPool := worker.NewWorkerPool(10, 100, logger) // 10 workers, 100 job capacity, logger
	workerPool.Start(ctx)
	defer workerPool.Stop()

	log.Println("Worker pool initialized with 10 workers")

	// Create crawler configuration
	crawlerConfig := &services.CrawlerConfig{
		MaxConcurrentCrawls: 10,
		RequestTimeout:      30 * time.Second,
		UserAgent:           "WebsiteAnalyzer/1.0",
		MaxRedirects:        5,
		MaxResponseSize:     10 * 1024 * 1024, // 10MB
		RetryAttempts:       3,
		RetryDelay:          1 * time.Second,
	}

	// Initialize services
	crawlerService := services.NewCrawlerService(urlRepo, workerPool, crawlerConfig, logger)

	// Initialize WebSocket handler
	wsHandler := handlers.NewWebSocketHandler()
	go wsHandler.Run()

	// Initialize handlers
	urlHandler := handlers.NewURLHandler(crawlerService, wsHandler)

	// Initialize Gin router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}))

	// Global rate limiting middleware (less restrictive)
	r.Use(middleware.APIRateLimitMiddleware())

	// Global input validation middleware
	r.Use(middleware.ValidationMiddleware())

	// // Enhanced health check endpoint with database status
	// r.GET("/health", func(c *gin.Context) {
	// 	dbStatus := "healthy"
	// 	if err := db.HealthCheck(ctx); err != nil {
	// 		dbStatus = "unhealthy"
	// 	}

	// 	c.JSON(http.StatusOK, gin.H{
	// 		"status":    "healthy",
	// 		"message":   "Website Analyzer API is running",
	// 		"database":  dbStatus,
	// 		"workers":   workerPool.GetStats(),
	// 		"timestamp": time.Now().UTC(),
	// 	})
	// })

	// WebSocket endpoint (no rate limiting for WebSocket)
	r.GET("/ws", wsHandler.HandleWebSocket)

	// API routes with standard rate limiting
	api := r.Group("/api")
	{
		// Public routes with standard rate limiting
		api.GET("/urls", urlHandler.GetURLs)
		api.GET("/urls/:id", urlHandler.GetURL)
		api.GET("/urls/:id/broken-links", urlHandler.GetBrokenLinks)

		// Routes that modify data with stricter rate limiting
		apiStrict := api.Group("/")
		apiStrict.Use(middleware.StrictRateLimitMiddleware())
		{
			apiStrict.POST("/urls", urlHandler.CreateURL)
			apiStrict.PUT("/urls/:id/analyze", urlHandler.AnalyzeURL)
			apiStrict.DELETE("/urls/:id", urlHandler.DeleteURL)
		}

		// Bulk operations with most restrictive rate limiting
		apiBulk := api.Group("/")
		apiBulk.Use(middleware.BulkOperationRateLimitMiddleware())
		{
			apiBulk.POST("/urls/bulk-analyze", urlHandler.BulkAnalyze)
			apiBulk.DELETE("/urls/bulk-delete", urlHandler.BulkDelete)
		}
	}

	// Start server
	port := getEnv("PORT", "8080")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		<-sigterm

		log.Println("Shutting down server...")
		cancel() // Cancel all contexts

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", port)
	log.Printf("Database connected to %s:%d", dbConfig.Host, dbConfig.Port)
	log.Printf("Rate limiting enabled: API (10 req/s), Strict (5 req/s), Bulk (2 req/s)")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}

	log.Println("Server stopped")
}