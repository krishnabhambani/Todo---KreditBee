package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/logger"
)

// LoggerMiddleware logs every HTTP request as a structured log entry.
// The logger is injected — no global state.
func LoggerMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Info("http request",
			logger.F("status", c.Writer.Status()),
			logger.F("latency", time.Since(start)),
			logger.F("ip", c.ClientIP()),
			logger.F("method", c.Request.Method),
			logger.F("path", path),
		)
	}
}
