package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func rejectPATOnAuthEndpoints() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[len("Bearer "):])
		}
		if strings.HasPrefix(strings.ToLower(token), "pat:") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "PAT_FORBIDDEN_ON_AUTH"})
			return
		}
		c.Next()
	}
}

func accessOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[len("Bearer "):])
		}
		if strings.HasPrefix(strings.ToLower(token), "pat:") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
			return
		}
		c.Next()
	}
}
