package reporting

import (
	"context"
	"testing"
	"time"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
)

func TestKeywordStats_AggregatesExpenseMemoAndCategoryTerms(t *testing.T) {
	ctx := context.Background()
	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	txnRepo := accounting.NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	txnService := accounting.NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)
	keywords := NewKeywordService(NewRepository(accountRepo, txnRepo, categoryService))

	ledger, err := ledgerRepo.Create("user-1", accounting.LedgerCreateInput{Name: "Main", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledger: %v", err)
	}
	food, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "餐饮"})
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}
	base := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: ledger.ID, Type: accounting.TransactionTypeExpense, Amount: 18, OccurredAt: base, CategoryID: &food.ID, Memo: "麻辣烫"}); err != nil {
		t.Fatalf("seed expense 1: %v", err)
	}
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: ledger.ID, Type: accounting.TransactionTypeExpense, Amount: 12, OccurredAt: base.Add(time.Hour), CategoryID: &food.ID, Memo: "麻辣烫 咖啡"}); err != nil {
		t.Fatalf("seed expense 2: %v", err)
	}
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: ledger.ID, Type: accounting.TransactionTypeIncome, Amount: 100, OccurredAt: base.Add(2 * time.Hour), Memo: "salary"}); err != nil {
		t.Fatalf("seed income: %v", err)
	}

	result, err := keywords.GetKeywordStats(ctx, "user-1", KeywordQuery{From: base.Add(-time.Hour), To: base.Add(24 * time.Hour), Limit: 10})
	if err != nil {
		t.Fatalf("keyword stats query: %v", err)
	}

	got := map[string]KeywordStatItem{}
	for _, item := range result.Items {
		got[item.Text] = item
	}
	if got["麻辣烫"].Amount != 30 || got["麻辣烫"].Count != 2 {
		t.Fatalf("expected 麻辣烫 to aggregate amount 30/count 2, got %#v", got["麻辣烫"])
	}
	if got["餐饮"].Amount != 30 || got["餐饮"].Count != 2 {
		t.Fatalf("expected category term to aggregate amount 30/count 2, got %#v", got["餐饮"])
	}
	if _, ok := got["salary"]; ok {
		t.Fatalf("income memo should not be included in expense keyword stats")
	}
}

func TestKeywordStats_RejectsInvalidRange(t *testing.T) {
	keywords := NewKeywordService(NewRepository(accounting.NewInMemoryAccountRepository(), accounting.NewInMemoryTransactionRepository(), nil))
	_, err := keywords.GetKeywordStats(context.Background(), "user-1", KeywordQuery{
		From: time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
	})
	if ErrorCode(err) != STAT_QUERY_INVALID {
		t.Fatalf("expected %s, got %q", STAT_QUERY_INVALID, ErrorCode(err))
	}
}
