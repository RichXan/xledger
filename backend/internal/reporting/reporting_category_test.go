package reporting

import (
	"context"
	"testing"
	"time"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
)

func TestCategoryStats_UsesHistoricalCategoryNames(t *testing.T) {
	ctx := context.Background()
	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	txnRepo := accounting.NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	txnService := accounting.NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)
	repo := NewRepository(accountRepo, txnRepo, categoryService)
	categoryStats := NewCategoryService(repo)

	ledger, err := ledgerRepo.Create("user-1", accounting.LedgerCreateInput{Name: "Main", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledger: %v", err)
	}
	food, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Food"})
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: ledger.ID, Type: accounting.TransactionTypeExpense, Amount: 45, OccurredAt: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), CategoryID: &food.ID}); err != nil {
		t.Fatalf("seed categorized expense: %v", err)
	}
	if _, err := categoryService.DeleteCategory(ctx, "user-1", food.ID); classification.ErrorCode(err) != classification.CAT_IN_USE_ARCHIVED {
		t.Fatalf("archive referenced category: %q", classification.ErrorCode(err))
	}

	result, err := categoryStats.GetCategoryStats(ctx, "user-1", CategoryQuery{})
	if err != nil {
		t.Fatalf("category stats query: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 category bucket, got %d", len(result.Items))
	}
	if result.Items[0].CategoryName != "Food" || result.Items[0].Amount != 45 {
		t.Fatalf("expected historical category name Food with amount 45, got %#v", result.Items[0])
	}
}
