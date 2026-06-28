package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/response"
)

// NotFoundHandler responds with a consistent JSON payload for unknown routes.
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusNotFound, "Route not found", nil)
	}
}

// MethodNotAllowedHandler responds with a consistent JSON payload when the
// request method is not allowed for the matched route.
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusMethodNotAllowed, "Method not allowed", nil)
	}
}
