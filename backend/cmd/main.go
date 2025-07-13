package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"website-analyzer/internal/config"
	"website-analyzer/internal/database"
	"website-analyzer/internal/handlers"
	"website-analyzer/internal/services"
)

func main() {
	cfg := config.Load()

	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	crawlerService := services.NewCrawlerService(db)

	urlHandler := handlers.NewURLHandler(crawlerService)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	api := router.Group("/api")
	{
		urls := api.Group("/urls")
		{
			urls.GET("", urlHandler.GetURLs)
			urls.POST("", urlHandler.CreateURL)
			urls.DELETE("/urls/:id", urlHandler.DeleteURL)
		}

		apiBulk.POST("/urls/bulk-analyze", urlHandler.BulkAnalyze)
		apiBulk.DELETE("/urls/bulk-delete", urlHandler.BulkDelete)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
} 