package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/auth"
	"xledger/backend/internal/classification"
	"xledger/backend/internal/portability"
	"xledger/backend/internal/reporting"
)

func NewRouter(trustedProxies []string) (*gin.Engine, error) {
	return NewRouterWithDependencies(trustedProxies, Dependencies{})
}

type Dependencies struct {
	AuthHandler           *auth.Handler
	AccountingHandler     *accounting.Handler
	ClassificationHandler *classification.Handler
	PortabilityHandler    *portability.Handler
	ReportingHandler      *reporting.Handler
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
	classificationHandler := deps.ClassificationHandler
	portabilityHandler := deps.PortabilityHandler
	reportingHandler := deps.ReportingHandler
	var businessDeps *defaultBusinessDeps
	if accountingHandler == nil || classificationHandler == nil || portabilityHandler == nil || reportingHandler == nil {
		businessDeps = newDefaultBusinessDeps()
	}
	if accountingHandler == nil {
		accountingHandler = newDefaultAccountingHandler(businessDeps)
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
		accountingGroup.GET("/transactions", accountingHandler.ListTransactions)
		accountingGroup.PUT("/transactions/:id", accountingHandler.UpdateTransaction)
		accountingGroup.DELETE("/transactions/:id", accountingHandler.DeleteTransaction)
	}

	if classificationHandler == nil {
		classificationHandler = newDefaultClassificationHandler(businessDeps)
	}
	if classificationHandler != nil {
		classificationGroup := r.Group("/api")
		classificationGroup.Use(accountingAuthMiddleware())
		classificationGroup.GET("/categories", classificationHandler.ListCategories)
		classificationGroup.POST("/categories", classificationHandler.CreateCategory)
		classificationGroup.PUT("/categories/:id", classificationHandler.UpdateCategory)
		classificationGroup.DELETE("/categories/:id", classificationHandler.DeleteCategory)

		classificationGroup.GET("/tags", classificationHandler.ListTags)
		classificationGroup.POST("/tags", classificationHandler.CreateTag)
		classificationGroup.PUT("/tags/:id", classificationHandler.UpdateTag)
		classificationGroup.DELETE("/tags/:id", classificationHandler.DeleteTag)
	}

	if portabilityHandler == nil {
		portabilityHandler = newDefaultPortabilityHandler(businessDeps)
	}
	if portabilityHandler != nil {
		portabilityGroup := r.Group("/api")
		portabilityGroup.Use(accountingAuthMiddleware())
		portabilityGroup.POST("/import/csv", portabilityHandler.ImportPreview)
		portabilityGroup.POST("/import/csv/confirm", portabilityHandler.ImportConfirm)
		portabilityGroup.GET("/export", portabilityHandler.Export)
	}

	if reportingHandler == nil {
		reportingHandler = newDefaultReportingHandler(businessDeps)
	}
	if reportingHandler != nil {
		reportingGroup := r.Group("/api")
		reportingGroup.Use(accountingAuthMiddleware())
		reportingGroup.GET("/stats/overview", reportingHandler.Overview)
		reportingGroup.GET("/stats/trend", reportingHandler.Trend)
		reportingGroup.GET("/stats/category", reportingHandler.Category)
	}

	return r, nil
}
