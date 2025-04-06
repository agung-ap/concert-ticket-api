package middleware

import (
	"time"

	"concert-ticket-api/pkg/logger"

	"github.com/gin-gonic/gin"
)

// RequestLogger creates a Gin middleware for request logging
func RequestLogger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate request processing time
		latency := time.Since(start)

		// Get status code and client IP
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// Add query string if it exists
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request details
		log.Info("Request: %s | %s | %d | %s | %s",
			method, path, statusCode, clientIP, latency.String())
	}
}
