package middleware

import (
	"net/http"
	"strings"

	"github.com/RichXan/xcommon/xauth"
	"github.com/gin-gonic/gin"
)

// Auth 认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// 检查Authorization格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		// 解析token
		claims, err := xauth.ParseAccessToken(parts[1])
		if err != nil {
			if err == xauth.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Token has expired",
					"code":  "token_expired",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// GetCurrentUser 从上下文中获取当前用户信息
func GetCurrentUser(c *gin.Context) (uint64, string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, "", false
	}

	username, exists := c.Get("username")
	if !exists {
		return 0, "", false
	}

	return userID.(uint64), username.(string), true
}
