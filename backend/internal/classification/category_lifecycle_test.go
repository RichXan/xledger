package classification_test

import (
	"context"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"xledger/backend/internal/auth"
	bootstraphttp "xledger/backend/internal/bootstrap/http"
	"xledger/backend/internal/classification"
	"xledger/backend/internal/portability"
)

func TestDeleteReferencedCategory_Archives_ReturnsCAT_IN_USE_ARCHIVED(t *testing.T) {
	ctx := context.Background()
	repo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(repo)

	created, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Groceries"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	if _, err := categoryService.RecordCategoryUsage(ctx, "user-1", created.ID); err != nil {
		t.Fatalf("record category usage: %v", err)
	}

	result, err := categoryService.DeleteCategory(ctx, "user-1", created.ID)
	if classification.ErrorCode(err) != classification.CAT_IN_USE_ARCHIVED {
		t.Fatalf("expected %s, got %q", classification.CAT_IN_USE_ARCHIVED, classification.ErrorCode(err))
	}
	if !result.Archived {
		t.Fatalf("expected category to be archived when in use")
	}
}

func TestCreateCategory_InvalidParent_ReturnsCAT_INVALID_PARENT(t *testing.T) {
	ctx := context.Background()
	repo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(repo)

	_, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{
		Name:     "Snacks",
		ParentID: ptr("missing-parent"),
	})
	if classification.ErrorCode(err) != classification.CAT_INVALID_PARENT {
		t.Fatalf("expected %s, got %q", classification.CAT_INVALID_PARENT, classification.ErrorCode(err))
	}
}

func TestCreateCategory_DepthExceeded_ReturnsCAT_INVALID_PARENT(t *testing.T) {
	ctx := context.Background()
	repo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(repo)

	root, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Food"})
	if err != nil {
		t.Fatalf("create root: %v", err)
	}
	child, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Dining", ParentID: &root.ID})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	_, err = categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Dinner", ParentID: &child.ID})
	if classification.ErrorCode(err) != classification.CAT_INVALID_PARENT {
		t.Fatalf("expected %s, got %q", classification.CAT_INVALID_PARENT, classification.ErrorCode(err))
	}
}

func TestArchivedCategory_NotSelectableInTxnEdit(t *testing.T) {
	ctx := context.Background()
	repo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(repo)

	created, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Travel"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	archive := true
	if _, err := categoryService.UpdateCategory(ctx, "user-1", created.ID, classification.CategoryUpdateInput{Archive: &archive}); err != nil {
		t.Fatalf("archive category: %v", err)
	}

	err = categoryService.ValidateCategorySelectable(ctx, "user-1", created.ID)
	if classification.ErrorCode(err) != classification.CAT_ARCHIVED {
		t.Fatalf("expected %s for archived category select, got %q", classification.CAT_ARCHIVED, classification.ErrorCode(err))
	}
}

func TestStatsAndExport_KeepHistoricalCategoryNameAfterArchive(t *testing.T) {
	ctx := context.Background()
	repo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(repo)

	created, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Salary"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	historyName, err := categoryService.RecordCategoryUsage(ctx, "user-1", created.ID)
	if err != nil {
		t.Fatalf("record usage: %v", err)
	}
	if historyName != "Salary" {
		t.Fatalf("expected recorded historical name Salary, got %q", historyName)
	}

	if _, err := categoryService.DeleteCategory(ctx, "user-1", created.ID); classification.ErrorCode(err) != classification.CAT_IN_USE_ARCHIVED {
		t.Fatalf("expected delete-in-use to archive with %s, got %q", classification.CAT_IN_USE_ARCHIVED, classification.ErrorCode(err))
	}

	retainedName, found := categoryService.GetHistoricalCategoryName(ctx, "user-1", created.ID)
	if !found {
		t.Fatalf("expected historical category name to be retained")
	}
	if retainedName != "Salary" {
		t.Fatalf("expected historical category name Salary, got %q", retainedName)
	}
}

func TestCategoriesAndTags_AcceptsAccessAndPAT(t *testing.T) {
	now := func() time.Time { return time.Now().UTC() }
	authRepo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(authRepo, &noopSender{}, nil, now, func() string { return "123456" }))
	sessionService := auth.NewSessionService(authRepo, nil, now)
	pair, err := sessionService.IssueSession(context.Background(), "category-auth@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	classificationHandler := classification.NewHandler(
		classification.NewCategoryService(classification.NewInMemoryRepository()),
		classification.NewTagService(classification.NewInMemoryRepository()),
	)

	patService := portability.NewPATService(now)
	patToken, _, err := patService.CreatePAT(context.Background(), "category-auth@example.com", "test-pat", nil)
	if err != nil {
		t.Fatalf("create PAT: %v", err)
	}

	r, err := bootstraphttp.NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, bootstraphttp.Dependencies{
		AuthHandler:           authHandler,
		ClassificationHandler: classificationHandler,
		PATService:            patService,
	})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	requireAuthOK(t, r, "/api/categories", "Bearer "+pair.AccessToken)
	requireAuthOK(t, r, "/api/tags", "Bearer "+patToken)
}

func requireAuthOK(t *testing.T, r stdhttp.Handler, path string, authz string) {
	t.Helper()
	req := httptest.NewRequest(stdhttp.MethodGet, path, nil)
	req.Header.Set("Authorization", authz)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("expected status %d for %s, got %d body=%s", stdhttp.StatusOK, path, rec.Code, rec.Body.String())
	}
}

type noopSender struct{}

func (noopSender) Send(string, string, string) error {
	return nil
}

func ptr(value string) *string {
	return &value
}
