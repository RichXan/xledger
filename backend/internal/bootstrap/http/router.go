package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/auth"
)

func NewRouter(trustedProxies []string) (*gin.Engine, error) {
	return NewRouterWithDependencies(trustedProxies, Dependencies{})
}

type Dependencies struct {
	AuthHandler       *auth.Handler
	AccountingHandler *accounting.Handler
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
	if handler.HasOAuthService() {
		authGroup.GET("/google/callback", handler.GoogleCallback)
	}
	if handler.HasSessionService() {
		authGroup.POST("/refresh", handler.Refresh)
		authGroup.POST("/logout", handler.Logout)
	}

	accountingHandler := deps.AccountingHandler
	if accountingHandler == nil {
		accountingHandler = newDefaultAccountingHandler()
	}
	if accountingHandler != nil {
		accountingGroup := r.Group("/api")
		accountingGroup.Use(accountingAuthMiddleware())
		accountingGroup.GET("/ledgers", accountingHandler.ListLedgers)
		accountingGroup.POST("/ledgers", accountingHandler.CreateLedger)
		accountingGroup.PUT("/ledgers/:id", accountingHandler.UpdateLedger)
		accountingGroup.DELETE("/ledgers/:id", accountingHandler.DeleteLedger)

		accountingGroup.GET("/accounts", accountingHandler.ListAccounts)
		accountingGroup.POST("/accounts", accountingHandler.CreateAccount)
		accountingGroup.GET("/accounts/:id", accountingHandler.GetAccount)
		accountingGroup.PUT("/accounts/:id", accountingHandler.UpdateAccount)
		accountingGroup.DELETE("/accounts/:id", accountingHandler.DeleteAccount)

		accountingGroup.POST("/transactions", accountingHandler.CreateTransaction)
		accountingGroup.PUT("/transactions/:id", accountingHandler.UpdateTransaction)
		accountingGroup.DELETE("/transactions/:id", accountingHandler.DeleteTransaction)
	}

	return r, nil
}
