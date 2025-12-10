package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger middleware logs request details
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process Request
		c.Next()

		// calculate Latency
		latency := time.Since(start)

		// Get Status Code
		statusCode := c.Writer.Status()

		// Build Log message
		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[%s] %s %s %d %v\n", c.Request.Method, path, c.ClientIP(), statusCode, latency)
	}
}
