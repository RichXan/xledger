package reporting

import (
	"context"
	"testing"
	"time"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
)

func TestTrend_BasicWindowAggregation(t *testing.T) {
	ctx := context.Background()
	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	txnRepo := accounting.NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	txnService := accounting.NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)
	repo := NewRepository(accountRepo, txnRepo, categoryService)
	trend := NewTrendService(repo)

	ledger, err := ledgerRepo.Create("user-1", accounting.LedgerCreateInput{Name: "Main", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledger: %v", err)
	}
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: ledger.ID, Type: accounting.TransactionTypeIncome, Amount: 100, OccurredAt: base.Add(2 * time.Hour)}); err != nil {
		t.Fatalf("seed income: %v", err)
	}
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: ledger.ID, Type: accounting.TransactionTypeExpense, Amount: 30, OccurredAt: base.Add(26 * time.Hour)}); err != nil {
		t.Fatalf("seed expense: %v", err)
	}

	result, err := trend.GetTrend(ctx, "user-1", TrendQuery{From: base, To: base.Add(48 * time.Hour), Granularity: "day"})
	if err != nil {
		t.Fatalf("trend query: %v", err)
	}
	if len(result.Points) != 2 {
		t.Fatalf("expected 2 trend buckets, got %d", len(result.Points))
	}
	if result.Points[0].Income != 100 || result.Points[0].Expense != 0 {
		t.Fatalf("expected day1 income bucket, got %#v", result.Points[0])
	}
	if result.Points[1].Income != 0 || result.Points[1].Expense != 30 {
		t.Fatalf("expected day2 expense bucket, got %#v", result.Points[1])
	}
}
