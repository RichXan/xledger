package accounting

import (
	"context"
	"testing"
	"time"

	"xledger/backend/internal/classification"
)

func TestListTransactions_Perf1K(t *testing.T) {
	ctx := context.Background()
	ledgerRepo := NewInMemoryLedgerRepository()
	accountRepo := NewInMemoryAccountRepository()
	txnRepo := NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	service := NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)

	ledger, err := ledgerRepo.Create("user-1", LedgerCreateInput{Name: "Main", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledger: %v", err)
	}
	account, err := accountRepo.Create("user-1", AccountCreateInput{Name: "Wallet", Type: "cash", InitialBalance: 100})
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}
	category, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Groceries"})
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}
	tag, err := tagService.CreateTag(ctx, "user-1", classification.TagCreateInput{Name: "weekly"})
	if err != nil {
		t.Fatalf("seed tag: %v", err)
	}

	base := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 1000; i++ {
		occurredAt := base.Add(time.Duration(i) * time.Minute)
		if _, err := service.CreateTransaction(ctx, "user-1", TransactionCreateInput{
			LedgerID:   ledger.ID,
			AccountID:  &account.ID,
			CategoryID: &category.ID,
			TagIDs:     []string{tag.ID},
			Type:       TransactionTypeExpense,
			Amount:     float64(i + 1),
			OccurredAt: occurredAt,
		}); err != nil {
			t.Fatalf("seed transaction %d: %v", i, err)
		}
	}

	start := time.Now()
	items, err := service.ListTransactions(ctx, "user-1", TransactionQuery{
		LedgerID:     ledger.ID,
		AccountID:    account.ID,
		CategoryID:   category.ID,
		TagID:        tag.ID,
		OccurredFrom: base,
		OccurredTo:   base.Add(1000 * time.Minute),
		Page:         1,
		PageSize:     50,
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("list filtered transactions: %v", err)
	}
	if len(items) != 50 {
		t.Fatalf("expected first page to return 50 items, got %d", len(items))
	}
	if elapsed > time.Second {
		t.Fatalf("expected p95-style filter query under 1s for 1k records, got %s", elapsed)
	}
}
