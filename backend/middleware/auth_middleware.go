package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/todo-app/backend/response"
	"github.com/todo-app/backend/utils"
)

// AuthMiddleware validates Bearer JWT tokens on protected routes.
// The jwtSecret is injected at construction time — no global config access.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization header is required", nil)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Error(c, http.StatusUnauthorized, "Authorization header must be Bearer token", nil)
			return
		}

		claims, err := utils.ValidateJWT(parts[1], jwtSecret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Invalid or expired token", nil)
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
