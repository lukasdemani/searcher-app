package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		expectedAPIKey := os.Getenv("API_KEY")
		if expectedAPIKey == "" {
			expectedAPIKey = "dev-api-key-2025"
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "API key is required",
				"message": "Please provide X-API-Key header",
			})
			c.Abort()
			return
		}

		if apiKey != expectedAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid API key",
				"message": "The provided API key is not valid",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func OptionalAPIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		expectedAPIKey := os.Getenv("API_KEY")
		if expectedAPIKey == "" {
			expectedAPIKey = "dev-api-key-2025"
		}

		if apiKey == "" {
			c.Header("X-Auth-Status", "unauthorized")
		} else if apiKey == expectedAPIKey {
			c.Header("X-Auth-Status", "authorized")
		} else {
			c.Header("X-Auth-Status", "invalid")
		}

		c.Next()
	}
}
