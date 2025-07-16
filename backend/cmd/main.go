package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"seacher-app/internal/database"
	"seacher-app/internal/handlers"
	"seacher-app/internal/middleware"
	"seacher-app/internal/repository"
	"seacher-app/internal/services"
	"seacher-app/internal/worker"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	dbConfig := &database.DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "analyzer_user",
		Password: "analyzer_pass",
		Database: "website_analyzer",
	}

	db, err := database.NewDatabase(dbConfig, logger)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Enhanced database connection established successfully")

	urlRepo := repository.NewMySQLURLRepository(db.DB)

	workerPool := worker.NewWorkerPool(10, 100, logger) // 10 workers, 100 job capacity, logger
	workerPool.Start(ctx)
	defer workerPool.Stop()

	log.Println("Worker pool initialized with 10 workers")

	crawlerConfig := &services.CrawlerConfig{
		MaxConcurrentCrawls: 10,
		RequestTimeout:      30 * time.Second,
		UserAgent:           "WebsiteAnalyzer/1.0",
		MaxRedirects:        5,
		MaxResponseSize:     10 * 1024 * 1024, // 10MB
		RetryAttempts:       3,
		RetryDelay:          1 * time.Second,
	}

	crawlerService := services.NewCrawlerService(urlRepo, workerPool, crawlerConfig, logger)

	wsHandler := handlers.NewWebSocketHandler()
	go wsHandler.Run()

	urlHandler := handlers.NewURLHandler(crawlerService, wsHandler)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	}))

	r.Use(middleware.APIRateLimitMiddleware())

	r.Use(middleware.ValidationMiddleware())

	r.GET("/health", func(c *gin.Context) {
		dbStatus := "healthy"
		if err := db.HealthCheck(ctx); err != nil {
			dbStatus = "unhealthy"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"message":   "Website Analyzer API is running",
			"database":  dbStatus,
			"workers":   workerPool.GetStats(),
			"timestamp": time.Now().UTC(),
		})
	})

	r.GET("/ws", wsHandler.HandleWebSocket)

	api := r.Group("/api")
	{
		api.GET("/urls", urlHandler.GetURLs)
		api.GET("/urls/:id", urlHandler.GetURL)
		api.GET("/urls/:id/broken-links", urlHandler.GetBrokenLinks)

		apiStrict := api.Group("/")
		apiStrict.Use(middleware.StrictRateLimitMiddleware())
		{
			apiStrict.POST("/urls", urlHandler.CreateURL)
			apiStrict.PUT("/urls/:id/analyze", urlHandler.AnalyzeURL)
			apiStrict.DELETE("/urls/:id", urlHandler.DeleteURL)
		}

		apiBulk := api.Group("/")
		apiBulk.Use(middleware.BulkOperationRateLimitMiddleware())
		{
			apiBulk.POST("/urls/bulk-analyze", urlHandler.BulkAnalyze)
			apiBulk.DELETE("/urls/bulk-delete", urlHandler.BulkDelete)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		<-sigterm

		log.Println("Shutting down server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", port)
	log.Printf("Rate limiting enabled: API (10 req/s), Strict (5 req/s), Bulk (2 req/s)")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}

	log.Println("Server stopped")
}