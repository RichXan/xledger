package classification_test

import (
	"context"
	"testing"

	"xledger/backend/internal/classification"
)

func TestCreateTag_Duplicate_ReturnsTAG_DUPLICATED(t *testing.T) {
	ctx := context.Background()
	repo := classification.NewInMemoryRepository()
	tagService := classification.NewTagService(repo)

	if _, err := tagService.CreateTag(ctx, "user-1", classification.TagCreateInput{Name: "food"}); err != nil {
		t.Fatalf("create first tag: %v", err)
	}
	_, err := tagService.CreateTag(ctx, "user-1", classification.TagCreateInput{Name: " food "})
	if classification.ErrorCode(err) != classification.TAG_DUPLICATED {
		t.Fatalf("expected %s, got %q", classification.TAG_DUPLICATED, classification.ErrorCode(err))
	}
}

func TestUpdateTag_NotFound_ReturnsTAG_NOT_FOUND(t *testing.T) {
	ctx := context.Background()
	repo := classification.NewInMemoryRepository()
	tagService := classification.NewTagService(repo)

	_, err := tagService.UpdateTag(ctx, "user-1", "missing-tag", classification.TagUpdateInput{Name: "commute"})
	if classification.ErrorCode(err) != classification.TAG_NOT_FOUND {
		t.Fatalf("expected %s, got %q", classification.TAG_NOT_FOUND, classification.ErrorCode(err))
	}
}

func TestTag_AssociationWithTransactionVisibleInListFilter(t *testing.T) {
	ctx := context.Background()
	repo := classification.NewInMemoryRepository()
	tagService := classification.NewTagService(repo)

	tag, err := tagService.CreateTag(ctx, "user-1", classification.TagCreateInput{Name: "subscription"})
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}

	if err := tagService.AttachTagToTransaction(ctx, "user-1", tag.ID, "txn-1"); err != nil {
		t.Fatalf("attach txn-1 to tag: %v", err)
	}
	if err := tagService.AttachTagToTransaction(ctx, "user-1", tag.ID, "txn-2"); err != nil {
		t.Fatalf("attach txn-2 to tag: %v", err)
	}

	txnIDs, err := tagService.ListTransactionIDsByTag(ctx, "user-1", tag.ID)
	if err != nil {
		t.Fatalf("list txns by tag: %v", err)
	}
	if len(txnIDs) != 2 || txnIDs[0] != "txn-1" || txnIDs[1] != "txn-2" {
		t.Fatalf("expected tag filter to expose txn-1 and txn-2, got %#v", txnIDs)
	}
}
