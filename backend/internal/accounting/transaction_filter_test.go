package accounting

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/classification"
)

func newTransactionFilterFixture(t *testing.T) (*TransactionService, Ledger, Ledger, Account, Account, *classification.CategoryService, *classification.TagService) {
	t.Helper()

	ledgerRepo := NewInMemoryLedgerRepository()
	accountRepo := NewInMemoryAccountRepository()
	txnRepo := NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	service := NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)

	ledgerA, err := ledgerRepo.Create("user-1", LedgerCreateInput{Name: "Main", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledgerA: %v", err)
	}
	ledgerB, err := ledgerRepo.Create("user-1", LedgerCreateInput{Name: "Side"})
	if err != nil {
		t.Fatalf("seed ledgerB: %v", err)
	}
	accountA, err := accountRepo.Create("user-1", AccountCreateInput{Name: "Wallet", Type: "cash", InitialBalance: 100})
	if err != nil {
		t.Fatalf("seed accountA: %v", err)
	}
	accountB, err := accountRepo.Create("user-1", AccountCreateInput{Name: "Bank", Type: "bank", InitialBalance: 100})
	if err != nil {
		t.Fatalf("seed accountB: %v", err)
	}

	return service, ledgerA, ledgerB, accountA, accountB, categoryService, tagService
}

func TestListTransactions_MultiFilterStableOrder(t *testing.T) {
	ctx := context.Background()
	service, ledgerA, ledgerB, accountA, accountB, categoryService, tagService := newTransactionFilterFixture(t)

	categoryA, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Groceries"})
	if err != nil {
		t.Fatalf("create categoryA: %v", err)
	}
	categoryB, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Travel"})
	if err != nil {
		t.Fatalf("create categoryB: %v", err)
	}
	tagA, err := tagService.CreateTag(ctx, "user-1", classification.TagCreateInput{Name: "weekly"})
	if err != nil {
		t.Fatalf("create tagA: %v", err)
	}
	tagB, err := tagService.CreateTag(ctx, "user-1", classification.TagCreateInput{Name: "ignore"})
	if err != nil {
		t.Fatalf("create tagB: %v", err)
	}

	baseTime := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	first, err := service.CreateTransaction(ctx, "user-1", TransactionCreateInput{
		LedgerID:   ledgerA.ID,
		AccountID:  &accountA.ID,
		CategoryID: &categoryA.ID,
		TagIDs:     []string{tagA.ID},
		Type:       TransactionTypeExpense,
		Amount:     11,
		OccurredAt: baseTime,
	})
	if err != nil {
		t.Fatalf("create first matching txn: %v", err)
	}
	second, err := service.CreateTransaction(ctx, "user-1", TransactionCreateInput{
		LedgerID:   ledgerA.ID,
		AccountID:  &accountA.ID,
		CategoryID: &categoryA.ID,
		TagIDs:     []string{tagA.ID},
		Type:       TransactionTypeExpense,
		Amount:     22,
		OccurredAt: baseTime,
	})
	if err != nil {
		t.Fatalf("create second matching txn: %v", err)
	}

	seedFilterNoise(t, ctx, service, ledgerB.ID, accountA.ID, categoryA.ID, tagA.ID, baseTime)
	seedFilterNoise(t, ctx, service, ledgerA.ID, accountB.ID, categoryA.ID, tagA.ID, baseTime)
	seedFilterNoise(t, ctx, service, ledgerA.ID, accountA.ID, categoryB.ID, tagA.ID, baseTime)
	seedFilterNoise(t, ctx, service, ledgerA.ID, accountA.ID, categoryA.ID, tagB.ID, baseTime)
	seedFilterNoise(t, ctx, service, ledgerA.ID, accountA.ID, categoryA.ID, tagA.ID, baseTime.Add(-48*time.Hour))

	items, err := service.ListTransactions(ctx, "user-1", TransactionQuery{
		LedgerID:     ledgerA.ID,
		AccountID:    accountA.ID,
		CategoryID:   categoryA.ID,
		TagID:        tagA.ID,
		OccurredFrom: baseTime.Add(-time.Hour),
		OccurredTo:   baseTime.Add(time.Hour),
		Page:         1,
		PageSize:     50,
	})
	if err != nil {
		t.Fatalf("list filtered transactions: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 matching items, got %d", len(items))
	}
	if items[0].ID != second.ID || items[1].ID != first.ID {
		t.Fatalf("expected stable order by occurred_at desc then id desc, got ids %q then %q", items[0].ID, items[1].ID)
	}
}

func TestListTransactions_InvalidRange_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil, nil, &TransactionService{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	r.GET("/transactions", handler.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/transactions?date_from=2026-03-21T00:00:00Z&date_to=2026-03-20T00:00:00Z", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assertTransactionFilterBadRequest(t, rec)
}

func TestListTransactions_InvalidPage_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil, nil, &TransactionService{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	r.GET("/transactions", handler.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/transactions?page=0&page_size=50", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assertTransactionFilterBadRequest(t, rec)
}

func TestListTransactions_InvalidPageSize_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil, nil, &TransactionService{})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	r.GET("/transactions", handler.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/transactions?page=1&page_size=0", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assertTransactionFilterBadRequest(t, rec)
}

func seedFilterNoise(t *testing.T, ctx context.Context, service *TransactionService, ledgerID string, accountID string, categoryID string, tagID string, occurredAt time.Time) {
	t.Helper()
	if _, err := service.CreateTransaction(ctx, "user-1", TransactionCreateInput{
		LedgerID:   ledgerID,
		AccountID:  &accountID,
		CategoryID: &categoryID,
		TagIDs:     []string{tagID},
		Type:       TransactionTypeExpense,
		Amount:     5,
		OccurredAt: occurredAt,
	}); err != nil {
		t.Fatalf("seed noise transaction: %v", err)
	}
}

func assertTransactionFilterBadRequest(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body["error_code"] != TXN_VALIDATION_FAILED {
		t.Fatalf("expected error_code=%s, got %q", TXN_VALIDATION_FAILED, body["error_code"])
	}
}
