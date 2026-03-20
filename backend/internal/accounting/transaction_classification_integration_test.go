package accounting

import (
	"context"
	"testing"
	"time"

	"xledger/backend/internal/classification"
)

func newTransactionClassificationFixture(t *testing.T) (*InMemoryTransactionRepository, *TransactionService, *classification.CategoryService, *classification.TagService, Ledger) {
	t.Helper()

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

	return txnRepo, service, categoryService, tagService, ledger
}

func TestCreateTxn_WithCategoryAndTag_PersistsSnapshotAndFilterVisibility(t *testing.T) {
	ctx := context.Background()
	_, service, categoryService, tagService, ledger := newTransactionClassificationFixture(t)

	category, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Salary"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	tag, err := tagService.CreateTag(ctx, "user-1", classification.TagCreateInput{Name: "monthly"})
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}

	created, err := service.CreateTransaction(ctx, "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeIncome,
		Amount:     100,
		OccurredAt: time.Now().UTC(),
		CategoryID: &category.ID,
		TagIDs:     []string{tag.ID},
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if ptrString(created.CategoryID) != category.ID {
		t.Fatalf("expected category id %q, got %q", category.ID, ptrString(created.CategoryID))
	}
	if created.CategoryName != "Salary" {
		t.Fatalf("expected category snapshot Salary, got %q", created.CategoryName)
	}

	if _, err := categoryService.DeleteCategory(ctx, "user-1", category.ID); classification.ErrorCode(err) != classification.CAT_IN_USE_ARCHIVED {
		t.Fatalf("expected %s after referenced delete, got %q", classification.CAT_IN_USE_ARCHIVED, classification.ErrorCode(err))
	}

	items, err := service.ListTransactions(ctx, "user-1", TransactionQuery{TagID: tag.ID})
	if err != nil {
		t.Fatalf("list by tag: %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("expected tag filter to return %q, got %#v", created.ID, items)
	}

	if err := service.DeleteTransaction(ctx, "user-1", created.ID, nil); err != nil {
		t.Fatalf("delete transaction: %v", err)
	}
	items, err = service.ListTransactions(ctx, "user-1", TransactionQuery{TagID: tag.ID})
	if err != nil {
		t.Fatalf("list by tag after delete: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected delete to remove tag association, got %#v", items)
	}
}

func TestEditTxn_ReplacesCategorySnapshotAndTagFilter(t *testing.T) {
	ctx := context.Background()
	_, service, categoryService, tagService, ledger := newTransactionClassificationFixture(t)

	oldCategory, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Food"})
	if err != nil {
		t.Fatalf("create old category: %v", err)
	}
	newCategory, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Transport"})
	if err != nil {
		t.Fatalf("create new category: %v", err)
	}
	oldTag, err := tagService.CreateTag(ctx, "user-1", classification.TagCreateInput{Name: "old"})
	if err != nil {
		t.Fatalf("create old tag: %v", err)
	}
	newTag, err := tagService.CreateTag(ctx, "user-1", classification.TagCreateInput{Name: "new"})
	if err != nil {
		t.Fatalf("create new tag: %v", err)
	}

	created, err := service.CreateTransaction(ctx, "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeExpense,
		Amount:     10,
		OccurredAt: time.Now().UTC(),
		CategoryID: &oldCategory.ID,
		TagIDs:     []string{oldTag.ID},
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	edited, err := service.EditTransaction(ctx, "user-1", created.ID, TransactionEditInput{
		Amount:      25,
		HasCategory: true,
		CategoryID:  &newCategory.ID,
		HasTagIDs:   true,
		TagIDs:      []string{newTag.ID},
	})
	if err != nil {
		t.Fatalf("edit transaction: %v", err)
	}
	if ptrString(edited.CategoryID) != newCategory.ID {
		t.Fatalf("expected new category id %q, got %q", newCategory.ID, ptrString(edited.CategoryID))
	}
	if edited.CategoryName != "Transport" {
		t.Fatalf("expected updated category snapshot Transport, got %q", edited.CategoryName)
	}

	oldItems, err := service.ListTransactions(ctx, "user-1", TransactionQuery{TagID: oldTag.ID})
	if err != nil {
		t.Fatalf("list by old tag: %v", err)
	}
	if len(oldItems) != 0 {
		t.Fatalf("expected old tag filter to be empty, got %#v", oldItems)
	}

	newItems, err := service.ListTransactions(ctx, "user-1", TransactionQuery{TagID: newTag.ID})
	if err != nil {
		t.Fatalf("list by new tag: %v", err)
	}
	if len(newItems) != 1 || newItems[0].ID != created.ID {
		t.Fatalf("expected new tag filter to return %q, got %#v", created.ID, newItems)
	}
}

func TestTxn_ArchivedCategoryRejectedOnCreateAndEdit(t *testing.T) {
	ctx := context.Background()
	_, service, categoryService, _, ledger := newTransactionClassificationFixture(t)

	archivedCategory, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Travel"})
	if err != nil {
		t.Fatalf("create archived category seed: %v", err)
	}
	archive := true
	if _, err := categoryService.UpdateCategory(ctx, "user-1", archivedCategory.ID, classification.CategoryUpdateInput{Archive: &archive}); err != nil {
		t.Fatalf("archive category: %v", err)
	}

	_, err = service.CreateTransaction(ctx, "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeExpense,
		Amount:     12,
		OccurredAt: time.Now().UTC(),
		CategoryID: &archivedCategory.ID,
	})
	if classification.ErrorCode(err) != classification.CAT_ARCHIVED {
		t.Fatalf("expected %s on create, got %q", classification.CAT_ARCHIVED, classification.ErrorCode(err))
	}

	created, err := service.CreateTransaction(ctx, "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeExpense,
		Amount:     20,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create uncategorized transaction: %v", err)
	}

	_, err = service.EditTransaction(ctx, "user-1", created.ID, TransactionEditInput{
		Amount:      21,
		HasCategory: true,
		CategoryID:  &archivedCategory.ID,
	})
	if classification.ErrorCode(err) != classification.CAT_ARCHIVED {
		t.Fatalf("expected %s on edit, got %q", classification.CAT_ARCHIVED, classification.ErrorCode(err))
	}
}
