package middleware

import (
	"incident-flow/backend/internal/logger"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware logs incoming requests and outgoing responses
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Log incoming request
		logger.Debug("incoming request",
			"method", c.Request.Method,
			"path", c.Request.RequestURI,
			"remote_addr", c.ClientIP(),
		)

		// Call next handler
		c.Next()

		// Log response
		duration := time.Since(startTime)
		logger.Info("request completed",
			"method", c.Request.Method,
			"path", c.Request.RequestURI,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
			"remote_addr", c.ClientIP(),
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("request error",
					"method", c.Request.Method,
					"path", c.Request.RequestURI,
					"error", err.Error(),
				)
			}
		}
	}
}
