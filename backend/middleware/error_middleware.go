package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/apperrors"
	"github.com/todo-app/backend/logger"
)

// ErrorHandler catches both panics and normal application errors added via c.Error(err).
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Recover from panics
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered",
					logger.F("error", err),
					logger.F("path", c.Request.URL.Path),
					logger.F("method", c.Request.Method),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Internal Server Error",
					"data":    nil,
				})
			}
		}()

		// 2. Process the request
		c.Next()

		// 3. Inspect registered errors after the handlers have run
		if len(c.Errors) > 0 {
			// Extract the last error added
			err := c.Errors.Last().Err

			var appErr *apperrors.AppError
			if errors.As(err, &appErr) {
				// We have a known application error with a status code
				if appErr.StatusCode >= 500 {
					log.Error("internal server error", logger.F("details", appErr.Error()), logger.F("path", c.Request.URL.Path))
				} else {
					log.Warn("client error", logger.F("details", appErr.Error()), logger.F("status", appErr.StatusCode))
				}
				
				c.JSON(appErr.StatusCode, gin.H{
					"success": false,
					"message": appErr.Message,
					"data":    nil,
				})
			} else {
				// Unknown/Generic error -> defaults to 500 Internal Server Error
				log.Error("unhandled generic error", logger.F("details", err.Error()), logger.F("path", c.Request.URL.Path))
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Internal Server Error",
					"data":    nil,
				})
			}
		}
	}
}
