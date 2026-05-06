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

func TestTransactionReviewQueue_SummarizesAndFiltersReasons(t *testing.T) {
	ctx := context.Background()
	service, ledger, _, account, _, categoryService, _ := newTransactionFilterFixture(t)
	category, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Food"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	seedReviewTransactions(t, ctx, service, ledger.ID, account.ID, category.ID)

	summary, err := service.GetReviewSummary(ctx, "user-1", TransactionQuery{})
	if err != nil {
		t.Fatalf("review summary: %v", err)
	}
	if summary.Review != 4 || summary.Uncategorized != 1 || summary.Duplicates != 1 || summary.Large != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}

	items, total, err := service.ListReviewItems(ctx, "user-1", TransactionReviewQuery{Reason: ReviewReasonUncategorized})
	if err != nil {
		t.Fatalf("review items: %v", err)
	}
	if total != 1 || len(items) != 1 || len(items[0].Reasons) != 1 || items[0].Reasons[0] != ReviewReasonUncategorized {
		t.Fatalf("unexpected uncategorized items total=%d items=%#v", total, items)
	}
}

func TestReviewEndpoints_ReturnSummaryAndReasonItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service, ledger, _, account, _, categoryService, _ := newTransactionFilterFixture(t)
	ctx := context.Background()
	category, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Food"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	seedReviewTransactions(t, ctx, service, ledger.ID, account.ID, category.ID)

	handler := NewHandler(nil, nil, service)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	r.GET("/transactions/review-summary", handler.ReviewSummary)
	r.GET("/transactions/review-items", handler.ReviewItems)

	summaryReq := httptest.NewRequest(http.MethodGet, "/transactions/review-summary", nil)
	summaryRec := httptest.NewRecorder()
	r.ServeHTTP(summaryRec, summaryReq)
	if summaryRec.Code != http.StatusOK {
		t.Fatalf("expected summary status 200, got %d body=%s", summaryRec.Code, summaryRec.Body.String())
	}
	var summaryEnvelope struct {
		Data TransactionReviewSummary `json:"data"`
	}
	if err := json.Unmarshal(summaryRec.Body.Bytes(), &summaryEnvelope); err != nil {
		t.Fatalf("decode summary: %v", err)
	}
	if summaryEnvelope.Data.Review != 4 || summaryEnvelope.Data.Uncategorized != 1 || summaryEnvelope.Data.Duplicates != 1 || summaryEnvelope.Data.Large != 1 {
		t.Fatalf("unexpected summary: %#v", summaryEnvelope.Data)
	}

	itemsReq := httptest.NewRequest(http.MethodGet, "/transactions/review-items?reason=duplicate", nil)
	itemsRec := httptest.NewRecorder()
	r.ServeHTTP(itemsRec, itemsReq)
	if itemsRec.Code != http.StatusOK {
		t.Fatalf("expected items status 200, got %d body=%s", itemsRec.Code, itemsRec.Body.String())
	}
	var itemsEnvelope struct {
		Data struct {
			Items []TransactionReviewItem `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(itemsRec.Body.Bytes(), &itemsEnvelope); err != nil {
		t.Fatalf("decode items: %v", err)
	}
	if len(itemsEnvelope.Data.Items) != 2 {
		t.Fatalf("expected 2 duplicate review items, got %d", len(itemsEnvelope.Data.Items))
	}
	for _, item := range itemsEnvelope.Data.Items {
		if len(item.Reasons) != 1 || item.Reasons[0] != ReviewReasonDuplicate {
			t.Fatalf("expected duplicate reason, got %#v", item.Reasons)
		}
	}
}

func seedReviewTransactions(t *testing.T, ctx context.Context, service *TransactionService, ledgerID string, accountID string, categoryID string) {
	t.Helper()
	base := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	seeds := []TransactionCreateInput{
		{LedgerID: ledgerID, AccountID: &accountID, Type: TransactionTypeExpense, Amount: 88, Memo: "needs category", OccurredAt: base},
		{LedgerID: ledgerID, AccountID: &accountID, CategoryID: &categoryID, Type: TransactionTypeExpense, Amount: 1250, Memo: "large flight", OccurredAt: base.Add(time.Hour)},
		{LedgerID: ledgerID, AccountID: &accountID, CategoryID: &categoryID, Type: TransactionTypeExpense, Amount: 25, Memo: "lunch", OccurredAt: base.Add(2 * time.Hour)},
		{LedgerID: ledgerID, AccountID: &accountID, CategoryID: &categoryID, Type: TransactionTypeExpense, Amount: 25, Memo: "lunch", OccurredAt: base.Add(3 * time.Hour)},
		{LedgerID: ledgerID, AccountID: &accountID, CategoryID: &categoryID, Type: TransactionTypeIncome, Amount: 5000, Memo: "salary", OccurredAt: base.Add(4 * time.Hour)},
	}
	for _, seed := range seeds {
		if _, err := service.CreateTransaction(ctx, "user-1", seed); err != nil {
			t.Fatalf("seed transaction: %v", err)
		}
	}
}
