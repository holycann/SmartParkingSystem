package utils

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger is a middleware function that logs request details
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log request details
		log.Printf("[%s] %s %s | %d | %v | %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			c.Writer.Status(),
			latency,
			c.Errors.String(),
		)
	}
}
