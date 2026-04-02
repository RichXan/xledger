package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/auth"
	"xledger/backend/internal/bootstrap/config"
	"xledger/backend/internal/bootstrap/infrastructure"
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
	UserIDResolver       userIDResolver
	PATService           *portability.PATService
}

func NewRouterWithDependencies(trustedProxies []string, deps Dependencies) (*gin.Engine, error) {
	r := gin.New()
	r.Use(gin.Logger())
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}
	r.Use(gin.Recovery())
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"code": "RESOURCE_NOT_FOUND", "message": "资源不存在", "data": nil})
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	handler := deps.AuthHandler
	if handler == nil {
		handler = newDefaultAuthHandlerFromEnv()
	}

	authGroup := r.Group("/api/auth")
	authGroup.Use(rejectPATOnAuthEndpoints())
	authGroup.POST("/send-code", handler.SendCode)
	authGroup.POST("/verify-code", handler.VerifyCode)
	if handler.HasSessionService() {
		authGroup.POST("/register", handler.RegisterWithPassword)
		authGroup.POST("/login", handler.LoginWithPassword)
		authGroup.POST("/change-password", handler.ChangePassword)
		authGroup.PATCH("/profile", handler.UpdateProfile)
	}
	if gin.Mode() != gin.ReleaseMode && handler.HasSessionService() {
		authGroup.POST("/dev-login", handler.DevLogin)
	}
	authGroup.GET("/me", handler.Me)
	if handler.HasOAuthService() {
		authGroup.GET("/google", handler.GoogleStart)
		authGroup.GET("/google/callback", handler.GoogleCallback)
		authGroup.POST("/google/exchange-code", handler.GoogleExchangeCode)
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
	// Use deps.PATService if provided, otherwise fall back to businessDeps.patService
	patService := deps.PATService
	if patService == nil && businessDeps != nil {
		patService = businessDeps.patService
	}
	if accountingHandler == nil {
		accountingHandler = newDefaultAccountingHandler(businessDeps)
	}
	if accountingHandler != nil {
		accountingGroup := r.Group("/api")
		accountingGroup.Use(accountingAuthMiddleware(deps.UserIDResolver, patService))
		accountingGroup.GET("/ledgers", accountingHandler.ListLedgers)
		accountingGroup.POST("/ledgers", accountingHandler.CreateLedger)
		accountingGroup.PATCH("/ledgers/:id", accountingHandler.UpdateLedger)
		accountingGroup.DELETE("/ledgers/:id", accountingHandler.DeleteLedger)

		accountingGroup.GET("/accounts", accountingHandler.ListAccounts)
		accountingGroup.POST("/accounts", accountingHandler.CreateAccount)
		accountingGroup.GET("/accounts/:id", accountingHandler.GetAccount)
		accountingGroup.PATCH("/accounts/:id", accountingHandler.UpdateAccount)
		accountingGroup.DELETE("/accounts/:id", accountingHandler.DeleteAccount)

		accountingGroup.POST("/transactions", accountingHandler.CreateTransaction)
		accountingGroup.GET("/transactions", accountingHandler.ListTransactions)
		accountingGroup.PATCH("/transactions/:id", accountingHandler.UpdateTransaction)
		accountingGroup.DELETE("/transactions/:id", accountingHandler.DeleteTransaction)

		quickAddHandler := accounting.NewQuickAddHandler(accountingHandler.GetTransactionService(), accountingHandler.GetLedgerService(), businessDeps.categoryService)
		accountingGroup.POST("/quick-add", quickAddHandler.QuickAdd)
		accountingGroup.GET("/quick-add/categories", quickAddHandler.ListCategories)
	}

	if classificationHandler == nil {
		classificationHandler = newDefaultClassificationHandler(businessDeps)
	}
	if classificationHandler != nil {
		classificationGroup := r.Group("/api")
		classificationGroup.Use(accountingAuthMiddleware(deps.UserIDResolver, patService))
		classificationGroup.GET("/categories", classificationHandler.ListCategories)
		classificationGroup.POST("/categories", classificationHandler.CreateCategory)
		classificationGroup.PATCH("/categories/:id", classificationHandler.UpdateCategory)
		classificationGroup.DELETE("/categories/:id", classificationHandler.DeleteCategory)

		classificationGroup.GET("/tags", classificationHandler.ListTags)
		classificationGroup.POST("/tags", classificationHandler.CreateTag)
		classificationGroup.PATCH("/tags/:id", classificationHandler.UpdateTag)
		classificationGroup.DELETE("/tags/:id", classificationHandler.DeleteTag)
	}

	if portabilityHandler == nil {
		portabilityHandler = newDefaultPortabilityHandler(businessDeps)
	}
	if portabilityHandler != nil {
		portabilityGroup := r.Group("/api")
		portabilityGroup.Use(accountingAuthMiddleware(deps.UserIDResolver, patService))
		portabilityGroup.POST("/import/csv", portabilityHandler.ImportPreview)
		portabilityGroup.POST("/import/csv/confirm", portabilityHandler.ImportConfirm)
		portabilityGroup.GET("/export", portabilityHandler.Export)
		patGroup := r.Group("/api/personal-access-tokens")
		patGroup.Use(accountingAuthMiddleware(deps.UserIDResolver, patService), accessOnlyMiddleware())
		patGroup.GET("", portabilityHandler.ListPATs)
		patGroup.POST("", portabilityHandler.CreatePAT)
		patGroup.DELETE("/:id", portabilityHandler.RevokePAT)

		shortcutGroup := r.Group("/api/shortcuts")
		shortcutGroup.Use(accountingAuthMiddleware(deps.UserIDResolver, patService))
		shortcutGroup.POST("/generate", portabilityHandler.GenerateShortcut)
		shortcutGroup.POST("/quick-add", portabilityHandler.QuickAdd)
		shortcutGroup.GET("/categories", portabilityHandler.ListCategories)
	}

	if reportingHandler == nil {
		reportingHandler = newDefaultReportingHandler(businessDeps)
	}
	if reportingHandler != nil {
		reportingGroup := r.Group("/api")
		reportingGroup.Use(accountingAuthMiddleware(deps.UserIDResolver, patService))
		reportingGroup.GET("/stats/overview", reportingHandler.Overview)
		reportingGroup.GET("/stats/trend", reportingHandler.Trend)
		reportingGroup.GET("/stats/category", reportingHandler.Category)
	}

	return r, nil
}

func NewRouterWithPostgreSQL(db *sql.DB, cfg config.Config, redisClient *redis.Client) *gin.Engine {
	authRepo := auth.NewPostgresRepository(db)
	var mailSender auth.SMTPSender = auth.NewSMTPMailSender(auth.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUser,
		Password: cfg.SMTPPass,
		From:     cfg.SMTPFrom,
	})
	if isPlaceholderSMTP(cfg.SMTPHost, cfg.SMTPPass) {
		mailSender = auth.NewDevMailSender()
	}
	templateService := classification.NewTemplateService(classification.NewPostgresTemplateRepository(db))
	sessionService := auth.NewSessionService(authRepo, &auth.SessionServiceOptions{
		PostLoginBootstrap: templateService.EnsureUserDefaults,
	}, time.Now)
	codeService := auth.NewCodeService(authRepo, mailSender, auth.NewSessionTokenIssuer(sessionService), time.Now, nil)
	oauthService := auth.NewOAuthService(authRepo, sessionService, time.Now)
	oauthService.SetGoogleProvider(auth.NewGoogleOAuthProvider(auth.GoogleOAuthConfig{
		ClientID:     cfg.GoogleAuthClientID,
		ClientSecret: cfg.GoogleAuthClientSecret,
		RedirectURL:  cfg.GoogleAuthRedirectURL,
	}))
	authHandler := auth.NewHandlerWithServices(codeService, oauthService, sessionService)
	authHandler.SetPasswordService(auth.NewPasswordService(authRepo))
	authHandler.SetGoogleFrontendReturnURL(cfg.GoogleAuthFrontendReturn)

	// Accounting domain wired via accounting_wiring.go
	acctDeps := newAccountingHandlerWithPostgreSQL(db)

	// Classification wired (shares CategoryService with accounting)
	classificationHandler := classification.NewHandler(acctDeps.CategoryService, nil)

	// Reporting wired via reporting package directly
	reportingRepo := reporting.NewRepository(nil, acctDeps.TxnRepo, acctDeps.CategoryService)
	var reportingCache reporting.Cache
	if redisClient != nil {
		reportingCache = infrastructure.NewRedisReportingCache(redisClient)
	}
	reportingHandler := reporting.NewHandler(
		reporting.NewOverviewService(reportingRepo, reportingCache),
		reporting.NewTrendService(reportingRepo, reportingCache),
		reporting.NewCategoryService(reportingRepo),
	)

	// Portability domain wired via portability_wiring.go
	portabilityHandler := newPortabilityHandlerWithPostgreSQL(db, acctDeps.TxnRepo, acctDeps.LedgerService, acctDeps.CategoryService)

	deps := Dependencies{
		AuthHandler:           authHandler,
		AccountingHandler:     acctDeps.AccountingHandler,
		ClassificationHandler: classificationHandler,
		PortabilityHandler:    portabilityHandler,
		ReportingHandler:      reportingHandler,
		UserIDResolver:       postgresUserIDResolver(db),
		PATService:           portabilityHandler.GetPATService(),
	}

	r, err := NewRouterWithDependencies(cfg.TrustedProxies, deps)
	if err != nil {
		panic(fmt.Sprintf("failed to create router: %v", err))
	}
	return r
}

func isPlaceholderSMTP(host string, password string) bool {
	host = strings.TrimSpace(strings.ToLower(host))
	password = strings.TrimSpace(strings.ToLower(password))
	if host == "" {
		return true
	}
	if strings.Contains(host, "example.com") || strings.Contains(host, "localhost") {
		return true
	}
	return password == "" || password == "replace-me"
}
