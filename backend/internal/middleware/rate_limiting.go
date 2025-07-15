package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	clients map[string]*rate.Limiter
	mu      sync.RWMutex
	rate    rate.Limit
	burst   int
	cleanup time.Duration
}

func NewRateLimiter(r rate.Limit, b int, cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*rate.Limiter),
		rate:    r,
		burst:   b,
		cleanup: cleanup,
	}

	go rl.cleanupClients()

	return rl
}

func (rl *RateLimiter) getLimiter(clientID string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.clients[clientID]
	rl.mu.RUnlock()

	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.mu.Lock()
		rl.clients[clientID] = limiter
		rl.mu.Unlock()
	}

	return limiter
}

func (rl *RateLimiter) cleanupClients() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		<-ticker.C
		rl.mu.Lock()
		for clientID, limiter := range rl.clients {
			if limiter.TokensAt(time.Now()) == float64(rl.burst) {
				delete(rl.clients, clientID)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.ClientIP()

		limiter := rl.getLimiter(clientID)

		if !limiter.Allow() {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", float64(rl.rate)))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests from this IP address",
				"retry_after": "1s",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", float64(rl.rate)))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%.0f", limiter.TokensAt(time.Now())))

		c.Next()
	}
}

func APIRateLimitMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(10, 20, 5*time.Minute)
	return limiter.RateLimitMiddleware()
}

func StrictRateLimitMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(5, 10, 5*time.Minute)
	return limiter.RateLimitMiddleware()
}

func BulkOperationRateLimitMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(2, 5, 10*time.Minute)
	return limiter.RateLimitMiddleware()
}