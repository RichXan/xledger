package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/auth"
)

func NewRouter(trustedProxies []string) (*gin.Engine, error) {
	return NewRouterWithDependencies(trustedProxies, Dependencies{})
}

type Dependencies struct {
	AuthHandler *auth.Handler
}

func NewRouterWithDependencies(trustedProxies []string, deps Dependencies) (*gin.Engine, error) {
	r := gin.New()
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	handler := deps.AuthHandler
	if handler == nil {
		handler = newDefaultAuthHandlerFromEnv()
	}

	authGroup := r.Group("/api/auth")
	authGroup.POST("/send-code", handler.SendCode)
	authGroup.POST("/verify-code", handler.VerifyCode)
	authGroup.GET("/google/callback", handler.GoogleCallback)
	authGroup.POST("/refresh", handler.Refresh)
	authGroup.POST("/logout", handler.Logout)

	return r, nil
}
