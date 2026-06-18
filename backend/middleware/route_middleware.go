package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NotFoundHandler responds with a consistent JSON payload for unknown routes.
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Route not found",
			"data":    nil,
		})
	}
}

// MethodNotAllowedHandler responds with a consistent JSON payload when the
// request method is not allowed for the matched route.
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"success": false,
			"message": "Method not allowed",
			"data":    nil,
		})
	}
}
