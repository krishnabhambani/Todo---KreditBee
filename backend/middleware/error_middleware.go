package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/logger"
	"github.com/todo-app/backend/response"
)

// ErrorHandler catches panics. Application errors should be returned explicitly using response.HandleError.
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Recover from panics
		defer func() {
			if err := recover(); err != nil {
				log.Error(c.Request.Context(), "panic recovered",
					logger.F("error", err),
					logger.F("path", c.Request.URL.Path),
					logger.F("method", c.Request.Method),
				)
				response.Error(c, http.StatusInternalServerError, "Internal Server Error", nil)
			}
		}()

		// 2. Process the request
		c.Next()
	}
}
