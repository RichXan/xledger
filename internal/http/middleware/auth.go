package middleware

import (
	"net/http"
	"strings"

	"github.com/RichXan/xcommon/xoauth"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthHeaderKey   = "Authorization"
	AuthUserIdKey   = "user_id"
	AuthUsernameKey = "username"
)

// Auth 认证中间件
func Auth(claim xoauth.Claim) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取token
		authHeader := c.GetHeader(AuthHeaderKey)
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
		claims, err := claim.ParseAccessToken(parts[1])
		if err != nil {
			if err == jwt.ErrTokenExpired {
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
		c.Set(AuthUserIdKey, claims.UserID)

		c.Next()
	}
}

// GetCurrentUser 从上下文中获取当前用户信息
func GetCurrentUser(c *gin.Context) (string, string, bool) {
	userID, exists := c.Get(AuthUserIdKey)
	if !exists {
		return "", "", false
	}

	username, exists := c.Get(AuthUsernameKey)
	if !exists {
		return "", "", false
	}

	return userID.(string), username.(string), true
}
