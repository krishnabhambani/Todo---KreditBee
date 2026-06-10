package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] %v", err)
				
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Internal Server Error",
					"data":    nil,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
