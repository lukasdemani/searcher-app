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
	"searcher-app/internal/repository"
	"searcher-app/internal/services"
	"searcher-app/internal/worker"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	dbConfig := &database.DatabaseConfig{
		Host:           getEnv("DB_HOST", "127.0.0.1"),
		Port:           getEnvInt("DB_PORT", 3306),
		Username:       getEnv("DB_USER", "analyzer_user"),
		Password:       getEnv("DB_PASSWORD", "analyzer_pass"),
		Database:       getEnv("DB_NAME", "website_analyzer"),
		ConnectTimeout: 30 * time.Second,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		Charset:        "utf8mb4",
		ParseTime:      true,
		Location:       "Local",
	}

	db, err := database.NewDatabase(dbConfig, logger)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Enhanced database connection established successfully")

	urlRepo := repository.NewMySQLURLRepository(db.DB)

	workerPool := worker.NewWorkerPool(10, 100, logger)
	workerPool.Start(ctx)
	defer workerPool.Stop()

	log.Println("Worker pool initialized with 10 workers")

	crawlerConfig := &services.CrawlerConfig{
		MaxConcurrentCrawls: 10,
		RequestTimeout:      30 * time.Second,
		UserAgent:           "WebsiteAnalyzer/1.0",
		MaxRedirects:        5,
		MaxResponseSize:     10 * 1024 * 1024,
		RetryAttempts:       3,
		RetryDelay:          1 * time.Second,
	}

	crawlerService := services.NewCrawlerService(urlRepo, workerPool, crawlerConfig, logger)

	wsHandler := handlers.NewWebSocketHandler()
	go wsHandler.Run()

	urlHandler := handlers.NewURLHandler(crawlerService, wsHandler)

	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"message":   "Website Analyzer API is running",
			"timestamp": time.Now().UTC(),
		})
	})

	r.GET("/ws", wsHandler.HandleWebSocket)

	api := r.Group("/api")
	{
		api.GET("/urls", urlHandler.GetURLs)
		api.GET("/urls/:id", urlHandler.GetURL)
		api.GET("/urls/:id/broken-links", urlHandler.GetBrokenLinks)
		api.POST("/urls", urlHandler.CreateURL)
		api.PUT("/urls/:id/analyze", urlHandler.AnalyzeURL)
		api.DELETE("/urls/:id", urlHandler.DeleteURL)
		api.POST("/urls/bulk-analyze", urlHandler.BulkAnalyze)
		api.POST("/urls/bulk-delete", urlHandler.BulkDelete)
	}

	port := getEnv("PORT", "8080")

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
	log.Printf("Database connected to %s:%d", dbConfig.Host, dbConfig.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}

	log.Println("Server stopped")
}
