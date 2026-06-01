package http

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/auth"
	"xledger/backend/internal/budget"
	"xledger/backend/internal/classification"
	"xledger/backend/internal/portability"
	"xledger/backend/internal/reporting"
)

func TestNewRouter_InvalidTrustedProxies_ReturnsError(t *testing.T) {
	_, err := NewRouter([]string{"not-a-cidr-or-ip"})
	if err == nil {
		t.Fatal("expected NewRouter to return error for invalid trusted proxies")
	}
}

func TestNewRouterWithDependencies_UsesInjectedAuthHandler(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := auth.NewInMemoryRepository(func() time.Time { return now })
	sender := &countingSender{}
	handler := auth.NewHandler(auth.NewCodeService(repo, sender, nil, func() time.Time { return now }, func() string { return "123456" }))

	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: handler})
	if err != nil {
		t.Fatalf("expected NewRouterWithDependencies to succeed, got: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/send-code", strings.NewReader(`{"email":"inject@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if sender.calls != 1 {
		t.Fatalf("expected injected sender to be called once, got %d", sender.calls)
	}
}

func TestNewRouterWithDependencies_DoesNotMountSessionRoutesForCodeOnlyHandler(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := auth.NewInMemoryRepository(func() time.Time { return now })
	handler := auth.NewHandler(auth.NewCodeService(repo, &countingSender{}, nil, func() time.Time { return now }, func() string { return "123456" }))

	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: handler})
	if err != nil {
		t.Fatalf("expected NewRouterWithDependencies to succeed, got: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", strings.NewReader(`{"refresh_token":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected refresh route to be absent for code-only handler, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountingRoutes_RejectExpiredAccessToken(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := auth.NewInMemoryRepository(func() time.Time { return now })
	handler := auth.NewHandler(auth.NewCodeService(repo, &countingSender{}, nil, func() time.Time { return now }, func() string { return "123456" }))

	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: handler})
	if err != nil {
		t.Fatalf("expected NewRouterWithDependencies to succeed, got: %v", err)
	}

	pastNow := func() time.Time { return time.Now().UTC().Add(-2 * time.Hour) }
	issued, err := auth.NewSessionService(repo, nil, pastNow).IssueSession(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("issue expired session token: %v", err)
	}
	expired := issued.AccessToken
	req := httptest.NewRequest(http.MethodGet, "/api/ledgers", nil)
	req.Header.Set("Authorization", "Bearer "+expired)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rec.Code, rec.Body.String())
	}
}

func TestDefaultHandlers_ShareClassificationStateWithTransactions(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := auth.NewInMemoryRepository(func() time.Time { return now })
	authHandler := auth.NewHandler(auth.NewCodeService(repo, &countingSender{}, nil, func() time.Time { return now }, func() string { return "123456" }))
	patService := portability.NewPATService(nil)

	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: authHandler, PATService: patService})
	if err != nil {
		t.Fatalf("expected NewRouterWithDependencies to succeed, got: %v", err)
	}
	authz := issuePATToken(t, patService, "shared@example.com")

	ledgerID := responseFieldFromData(t, performJSON(t, r, http.MethodPost, "/api/ledgers", `{"name":"Main","is_default":true}`, authz, http.StatusCreated), "id")
	categoryID := responseFieldFromData(t, performJSON(t, r, http.MethodPost, "/api/categories", `{"name":"Salary"}`, authz, http.StatusCreated), "id")
	tagID := responseFieldFromData(t, performJSON(t, r, http.MethodPost, "/api/tags", `{"name":"monthly"}`, authz, http.StatusCreated), "id")

	txnBody := `{"ledger_id":"` + ledgerID + `","type":"income","amount":100,"category_id":"` + categoryID + `","tag_ids":["` + tagID + `"]}`
	txnResp := performJSON(t, r, http.MethodPost, "/api/transactions", txnBody, authz, http.StatusCreated)
	if got := responseFieldFromData(t, txnResp, "category_name"); got != "Salary" {
		t.Fatalf("expected transaction category snapshot Salary, got %q", got)
	}

	deleteResp := performJSON(t, r, http.MethodDelete, "/api/categories/"+categoryID, ``, authz, http.StatusOK)
	deleteData := responseDataMap(t, deleteResp)
	if archived, ok := deleteData["archived"].(bool); !ok || !archived {
		t.Fatalf("expected referenced category delete to archive, got %#v", deleteResp)
	}
}

func TestStatsEndpoints_AcceptsAccessAndPAT(t *testing.T) {
	now := func() time.Time { return time.Now().UTC() }
	authRepo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(authRepo, &countingSender{}, nil, now, func() string { return "123456" }))
	pair, err := auth.NewSessionService(authRepo, nil, now).IssueSession(context.Background(), "reporting-auth@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	patService := portability.NewPATService(nil)
	patToken := issuePATToken(t, patService, "reporting-auth@example.com")

	accountRepo := accounting.NewInMemoryAccountRepository()
	txnRepo := accounting.NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	reportingRepo := reporting.NewRepository(accountRepo, txnRepo, categoryService)
	reportingHandler := reporting.NewHandler(
		reporting.NewOverviewService(reportingRepo, nil),
		reporting.NewTrendService(reportingRepo, nil),
		reporting.NewCategoryService(reportingRepo),
		reporting.NewKeywordService(reportingRepo),
	)

	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{
		AuthHandler:      authHandler,
		ReportingHandler: reportingHandler,
		PATService:       patService,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	if len(performJSON(t, r, http.MethodGet, "/api/stats/overview", ``, pair.AccessToken, http.StatusOK)) == 0 {
		t.Fatalf("expected overview payload")
	}
	if len(performJSON(t, r, http.MethodGet, "/api/stats/trend?from=2026-03-01T00:00:00Z&to=2026-03-02T00:00:00Z", ``, patToken, http.StatusOK)) == 0 {
		t.Fatalf("expected trend payload")
	}
	if len(performJSON(t, r, http.MethodGet, "/api/stats/category", ``, pair.AccessToken, http.StatusOK)) == 0 {
		t.Fatalf("expected category payload")
	}
	if len(performJSON(t, r, http.MethodGet, "/api/stats/keywords", ``, pair.AccessToken, http.StatusOK)) == 0 {
		t.Fatalf("expected keyword payload")
	}
}

func TestBudgetEndpoints_AcceptsAccessAndPAT(t *testing.T) {
	now := func() time.Time { return time.Now().UTC() }
	authRepo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(authRepo, &countingSender{}, nil, now, func() string { return "123456" }))
	pair, err := auth.NewSessionService(authRepo, nil, now).IssueSession(context.Background(), "budget-auth@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	patService := portability.NewPATService(nil)
	patToken := issuePATToken(t, patService, "budget-auth@example.com")

	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	txnRepo := accounting.NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	txnService := accounting.NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)
	budgetHandler := budget.NewHandler(budget.NewService(budget.NewInMemoryRepository(), txnService))

	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{
		AuthHandler:   authHandler,
		BudgetHandler: budgetHandler,
		PATService:    patService,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	category, err := categoryService.CreateCategory(context.Background(), "budget-auth@example.com", classification.CategoryCreateInput{Name: "Food"})
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}

	performJSON(t, r, http.MethodGet, "/api/budgets", ``, pair.AccessToken, http.StatusOK)
	created := performJSON(t, r, http.MethodPost, "/api/budgets", `{"category_id":"`+category.ID+`","amount":800,"alert_at":80}`, pair.AccessToken, http.StatusCreated)
	if got := responseFieldFromData(t, created, "category_id"); got != category.ID {
		t.Fatalf("expected created budget category %q, got %q", category.ID, got)
	}
	performJSON(t, r, http.MethodGet, "/api/budgets", ``, patToken, http.StatusOK)
}

func TestImportPreviewEndpoint_AcceptsAccessAndPAT(t *testing.T) {
	now := func() time.Time { return time.Now().UTC() }
	authRepo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(authRepo, &countingSender{}, nil, now, func() string { return "123456" }))
	pair, err := auth.NewSessionService(authRepo, nil, now).IssueSession(context.Background(), "import-auth@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	patService := portability.NewPATService(nil)
	patToken := issuePATToken(t, patService, "import-auth@example.com")

	repo := portability.NewRepository(nil)
	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{
		AuthHandler:        authHandler,
		PortabilityHandler: portability.NewHandler(portability.NewImportPreviewService(), portability.NewImportConfirmService(repo), nil, patService),
		PATService:         patService,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	performMultipartCSV(t, r, "/api/import/csv", pair.AccessToken, http.StatusOK)
	performMultipartCSV(t, r, "/api/import/csv", patToken, http.StatusOK)
}

func TestExportEndpoint_AcceptsAccessAndPAT(t *testing.T) {
	now := func() time.Time { return time.Now().UTC() }
	authRepo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(authRepo, &countingSender{}, nil, now, func() string { return "123456" }))
	pair, err := auth.NewSessionService(authRepo, nil, now).IssueSession(context.Background(), "export-auth@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	patService := portability.NewPATService(nil)
	patToken := issuePATToken(t, patService, "export-auth@example.com")

	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	txnRepo := accounting.NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	txnService := accounting.NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)
	ledger, err := ledgerRepo.Create("bddde8db-cd9c-56a8-a4a5-fae9e6424fa0", accounting.LedgerCreateInput{Name: "Main", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledger: %v", err)
	}
	category, err := categoryService.CreateCategory(context.Background(), "bddde8db-cd9c-56a8-a4a5-fae9e6424fa0", classification.CategoryCreateInput{Name: "Food"})
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if _, err := txnService.CreateTransaction(context.Background(), "bddde8db-cd9c-56a8-a4a5-fae9e6424fa0", accounting.TransactionCreateInput{LedgerID: ledger.ID, Type: accounting.TransactionTypeExpense, Amount: 25, OccurredAt: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC), CategoryID: &category.ID}); err != nil {
		t.Fatalf("seed export txn: %v", err)
	}
	if _, err := categoryService.DeleteCategory(context.Background(), "bddde8db-cd9c-56a8-a4a5-fae9e6424fa0", category.ID); classification.ErrorCode(err) != classification.CAT_IN_USE_ARCHIVED {
		t.Fatalf("archive category: %q", classification.ErrorCode(err))
	}
	exportRepo := portability.NewExportRepository(txnRepo, categoryService)
	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{
		AuthHandler: authHandler,
		PortabilityHandler: portability.NewHandler(
			portability.NewImportPreviewService(),
			portability.NewImportConfirmService(portability.NewRepository(nil)),
			portability.NewExportService(exportRepo),
			patService,
		),
		PATService: patService,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	performExportRequest(t, r, "/api/export?format=csv", pair.AccessToken, http.StatusOK, "text/csv")
	performExportRequest(t, r, "/api/export?format=json", patToken, http.StatusOK, "application/json")
}

func TestPAT_CannotAccessAuthEndpoints(t *testing.T) {
	now := func() time.Time { return time.Now().UTC() }
	authRepo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(authRepo, &countingSender{}, nil, now, func() string { return "123456" }))
	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: authHandler})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	for _, path := range []string{"/api/auth/send-code", "/api/auth/verify-code"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"email":"x@example.com","code":"123456","refresh_token":"t"}`))
		req.Header.Set("Authorization", "Bearer pat:blocked@example.com")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected PAT auth rejection for %s, got %d body=%s", path, rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "PAT_FORBIDDEN_ON_AUTH") {
			t.Fatalf("expected PAT_FORBIDDEN_ON_AUTH for %s, got %s", path, rec.Body.String())
		}
	}
}

func TestAccountingAuthMiddleware_RejectsPATWhenPATServiceMissing(t *testing.T) {
	now := func() time.Time { return time.Now().UTC() }
	authRepo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(authRepo, &countingSender{}, nil, now, func() string { return "123456" }))
	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: authHandler})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/ledgers", nil)
	req.Header.Set("Authorization", "Bearer pat:missing@example.com")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for PAT without service, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountingAuthMiddleware_RejectsPATWhenValidationFails(t *testing.T) {
	now := func() time.Time { return time.Now().UTC() }
	authRepo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(authRepo, &countingSender{}, nil, now, func() string { return "123456" }))
	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: authHandler, PATService: portability.NewPATService(nil)})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/ledgers", nil)
	req.Header.Set("Authorization", "Bearer pat:invalid@example.com")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for invalid PAT, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountingAuthMiddleware_AcceptsPATUserIDWhenResolverExists(t *testing.T) {
	patService := portability.NewPATService(func() time.Time { return time.Now().UTC() })
	patToken, _, err := patService.CreatePAT(context.Background(), "user-1", "shortcut", nil)
	if err != nil {
		t.Fatalf("create pat: %v", err)
	}

	resolverCalled := false
	r := gin.New()
	r.Use(accountingAuthMiddleware(func(context.Context, string) (string, error) {
		resolverCalled = true
		return "", nil
	}, patService))
	r.POST("/api/shortcuts/quick-add", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("user_id"))
	})

	req := httptest.NewRequest(http.MethodPost, "/api/shortcuts/quick-add", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+patToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected PAT to authenticate as user id, got %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "user-1" {
		t.Fatalf("expected user-1, got %q", rec.Body.String())
	}
	if resolverCalled {
		t.Fatalf("expected PAT user id to skip email resolver")
	}
}

func TestPATEndpoints_AccessOnly_GETPOSTDELETE_On_api_settings_tokens(t *testing.T) {
	now := func() time.Time { return time.Now().UTC() }
	authRepo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(authRepo, &countingSender{}, nil, now, func() string { return "123456" }))
	pair, err := auth.NewSessionService(authRepo, nil, now).IssueSession(context.Background(), "pat-admin@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{
		AuthHandler: authHandler,
		PortabilityHandler: portability.NewHandler(
			portability.NewImportPreviewService(),
			portability.NewImportConfirmService(portability.NewRepository(nil)),
			nil,
			portability.NewPATService(nil),
		),
		PATService: portability.NewPATService(nil),
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	performJSON(t, r, http.MethodGet, "/api/personal-access-tokens", ``, pair.AccessToken, http.StatusOK)
	performJSON(t, r, http.MethodPost, "/api/personal-access-tokens", `{}`, pair.AccessToken, http.StatusOK)
	req := httptest.NewRequest(http.MethodDelete, "/api/personal-access-tokens/pat-1", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected delete PAT endpoint to be mounted, got %d body=%s", rec.Code, rec.Body.String())
	}
	performJSON(t, r, http.MethodGet, "/api/personal-access-tokens", ``, "pat:pat-admin@example.com", http.StatusUnauthorized)
}

func TestQuickAddPreviewAndConfirm_AcceptsPATBoundUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := newDefaultBusinessDeps()
	userID := "quick-add-user"
	ledger, err := deps.ledgerRepo.Create(userID, accounting.LedgerCreateInput{Name: "Daily", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledger: %v", err)
	}
	patToken := issuePATToken(t, deps.patService, userID)
	now := func() time.Time { return time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC) }
	authRepo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(authRepo, &countingSender{}, nil, now, func() string { return "123456" }))
	accountingHandler := newDefaultAccountingHandler(deps)
	portabilityHandler := newDefaultPortabilityHandler(deps)
	router, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{
		AuthHandler:        authHandler,
		AccountingHandler:  accountingHandler,
		PortabilityHandler: portabilityHandler,
		PATService:         deps.patService,
	})
	if err != nil {
		t.Fatalf("router: %v", err)
	}

	generateReq := httptest.NewRequest(http.MethodPost, "/api/shortcuts/generate", strings.NewReader(`{"default_ledger_id":"`+ledger.ID+`","mode":"ocr_confirm"}`))
	generateReq.Header.Set("Content-Type", "application/json")
	generateReq.Header.Set("Authorization", "Bearer "+patToken)
	generateRec := httptest.NewRecorder()
	router.ServeHTTP(generateRec, generateReq)
	if generateRec.Code != http.StatusOK {
		t.Fatalf("generate status=%d body=%s", generateRec.Code, generateRec.Body.String())
	}

	previewReq := httptest.NewRequest(http.MethodPost, "/api/quick-add/preview", strings.NewReader(`{"raw_text":"微信支付\n收款方：瑞幸咖啡\n支付金额 ￥35.00","source":"ios_shortcuts_ocr","idempotency_key":"qe-router"}`))
	previewReq.Header.Set("Content-Type", "application/json")
	previewReq.Header.Set("Authorization", "Bearer "+patToken)
	previewRec := httptest.NewRecorder()
	router.ServeHTTP(previewRec, previewReq)
	if previewRec.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", previewRec.Code, previewRec.Body.String())
	}

	confirmReq := httptest.NewRequest(http.MethodPost, "/api/quick-add/confirm", strings.NewReader(`{"idempotency_key":"qe-router","ledger_id":"`+ledger.ID+`","type":"expense","amount":35,"memo":"瑞幸咖啡 · 微信支付"}`))
	confirmReq.Header.Set("Content-Type", "application/json")
	confirmReq.Header.Set("Authorization", "Bearer "+patToken)
	confirmRec := httptest.NewRecorder()
	router.ServeHTTP(confirmRec, confirmReq)
	if confirmRec.Code != http.StatusCreated {
		t.Fatalf("confirm status=%d body=%s", confirmRec.Code, confirmRec.Body.String())
	}
}

func performJSON(t *testing.T, handler http.Handler, method string, path string, body string, accessToken string, wantStatus int) map[string]any {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("expected status %d for %s %s, got %d body=%s", wantStatus, method, path, rec.Code, rec.Body.String())
	}
	if rec.Body.Len() == 0 {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response for %s %s: %v body=%s", method, path, err, rec.Body.String())
	}
	return payload
}

func responseField(t *testing.T, payload map[string]any, key string) string {
	t.Helper()
	value, ok := payload[key].(string)
	if !ok {
		t.Fatalf("expected string field %q in %#v", key, payload)
	}
	return value
}

func responseFieldFromData(t *testing.T, payload map[string]any, key string) string {
	t.Helper()
	data := responseDataMap(t, payload)
	value, ok := data[key].(string)
	if !ok {
		t.Fatalf("expected string field %q in data %#v", key, payload)
	}
	return value
}

func responseDataMap(t *testing.T, payload map[string]any) map[string]any {
	t.Helper()
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data map in %#v", payload)
	}
	return data
}

func performMultipartCSV(t *testing.T, handler http.Handler, path string, accessToken string, wantStatus int) {
	t.Helper()
	body := &strings.Builder{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "preview.csv")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("date,amount\n2026-03-01,12.5\n")); err != nil {
		t.Fatalf("write csv payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body.String()))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("expected status %d for multipart %s, got %d body=%s", wantStatus, path, rec.Code, rec.Body.String())
	}
}

func performExportRequest(t *testing.T, handler http.Handler, path string, accessToken string, wantStatus int, wantContentType string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("expected status %d for export %s, got %d body=%s", wantStatus, path, rec.Code, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.HasPrefix(contentType, wantContentType) {
		t.Fatalf("expected content-type %s, got %s", wantContentType, contentType)
	}
}

func issuePATToken(t *testing.T, patService *portability.PATService, userID string) string {
	t.Helper()
	token, _, err := patService.CreatePAT(context.Background(), userID, "test-pat", nil)
	if err != nil {
		t.Fatalf("create PAT for %s: %v", userID, err)
	}
	return token
}

type countingSender struct {
	calls int
}

func (s *countingSender) Send(to string, subject string, body string) error {
	s.calls++
	return nil
}
