package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter creates a Gin middleware for rate limiting
func RateLimiter(rps int) gin.HandlerFunc {
	// Create a map to store limiters for each client IP
	limiters := &sync.Map{}

	// Define the limit rate (requests per second)
	limit := rate.Limit(rps)

	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Get or create limiter for this client
		limiterI, _ := limiters.LoadOrStore(clientIP, rate.NewLimiter(limit, rps))
		limiter := limiterI.(*rate.Limiter)

		// Check if the request can proceed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
