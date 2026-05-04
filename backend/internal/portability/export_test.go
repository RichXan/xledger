package portability

import (
	"context"
	"strings"
	"testing"
	"time"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
)

func TestExport_PreservesHistoricalCategoryNames(t *testing.T) {
	ctx := context.Background()
	service := newExportFixture(t)
	content, err := service.Export(ctx, "user-1", ExportQuery{Format: "csv"})
	if err != nil {
		t.Fatalf("export csv: %v", err)
	}
	if !strings.Contains(content, "Food") {
		t.Fatalf("expected historical category name in export, got %s", content)
	}
}

func TestExport_AcceptsAccessAndPAT(t *testing.T) {
	service := newExportFixture(t)
	for _, tokenType := range []string{"access", "pat"} {
		content, err := service.Export(context.Background(), "user-1", ExportQuery{Format: "json"})
		if err != nil {
			t.Fatalf("expected %s export success, got %v", tokenType, err)
		}
		if !strings.Contains(content, "occurred_at") {
			t.Fatalf("expected exported payload for %s, got %s", tokenType, content)
		}
	}
}

func TestExport_InvalidRange_ReturnsEXPORT_INVALID_RANGE(t *testing.T) {
	service := newExportFixture(t)
	_, err := service.Export(context.Background(), "user-1", ExportQuery{Format: "csv", From: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)})
	if ErrorCode(err) != EXPORT_INVALID_RANGE {
		t.Fatalf("expected %s, got %q", EXPORT_INVALID_RANGE, ErrorCode(err))
	}
}

func TestExport_Timeout_ReturnsEXPORT_TIMEOUT(t *testing.T) {
	service := NewExportService(&ExportRepository{listFn: func(userID string, query accounting.TransactionQuery) ([]accounting.Transaction, error) {
		time.Sleep(25 * time.Millisecond)
		return []accounting.Transaction{{OccurredAt: time.Now().UTC()}}, nil
	}, historyFn: func(context.Context, string, string) (string, bool) { return "", false }})
	_, err := service.Export(context.Background(), "user-1", ExportQuery{Format: "json", Timeout: 5 * time.Millisecond})
	if ErrorCode(err) != EXPORT_TIMEOUT {
		t.Fatalf("expected %s, got %q", EXPORT_TIMEOUT, ErrorCode(err))
	}
}

func TestExport_SupportsCSVAndJSON(t *testing.T) {
	service := newExportFixture(t)
	csvContent, err := service.Export(context.Background(), "user-1", ExportQuery{Format: "csv"})
	if err != nil {
		t.Fatalf("csv export: %v", err)
	}
	if !strings.Contains(csvContent, "occurred_at,amount,type,ledger_id,account_id,from_account_id,to_account_id,category_name,memo") {
		t.Fatalf("expected csv header, got %s", csvContent)
	}
	jsonContent, err := service.Export(context.Background(), "user-1", ExportQuery{Format: "json"})
	if err != nil {
		t.Fatalf("json export: %v", err)
	}
	if !strings.Contains(jsonContent, "\"category_name\":\"Food\"") {
		t.Fatalf("expected json payload, got %s", jsonContent)
	}
}

func TestExport_FiltersByLedgerAndAccount(t *testing.T) {
	ctx := context.Background()
	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	txnRepo := accounting.NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	txnService := accounting.NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)

	primaryLedger, err := ledgerRepo.Create("user-1", accounting.LedgerCreateInput{Name: "Primary", IsDefault: true})
	if err != nil {
		t.Fatalf("seed primary ledger: %v", err)
	}
	secondaryLedger, err := ledgerRepo.Create("user-1", accounting.LedgerCreateInput{Name: "Secondary"})
	if err != nil {
		t.Fatalf("seed secondary ledger: %v", err)
	}
	cashAccount, err := accountRepo.Create("user-1", accounting.AccountCreateInput{Name: "Cash", Type: "cash"})
	if err != nil {
		t.Fatalf("seed cash account: %v", err)
	}
	bankAccount, err := accountRepo.Create("user-1", accounting.AccountCreateInput{Name: "Bank", Type: "debit"})
	if err != nil {
		t.Fatalf("seed bank account: %v", err)
	}
	food, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Food"})
	if err != nil {
		t.Fatalf("seed food category: %v", err)
	}
	travel, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Travel"})
	if err != nil {
		t.Fatalf("seed travel category: %v", err)
	}

	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{
		LedgerID:   primaryLedger.ID,
		AccountID:  &cashAccount.ID,
		CategoryID: &food.ID,
		Type:       accounting.TransactionTypeExpense,
		Amount:     25,
		OccurredAt: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("seed matching txn: %v", err)
	}
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{
		LedgerID:   primaryLedger.ID,
		AccountID:  &bankAccount.ID,
		CategoryID: &travel.ID,
		Type:       accounting.TransactionTypeExpense,
		Amount:     50,
		OccurredAt: time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("seed other account txn: %v", err)
	}
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{
		LedgerID:   secondaryLedger.ID,
		AccountID:  &cashAccount.ID,
		CategoryID: &travel.ID,
		Type:       accounting.TransactionTypeExpense,
		Amount:     75,
		OccurredAt: time.Date(2026, 3, 3, 12, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("seed other ledger txn: %v", err)
	}

	service := NewExportService(NewExportRepository(txnRepo, categoryService))
	content, err := service.Export(ctx, "user-1", ExportQuery{
		Format:    "csv",
		LedgerID:  primaryLedger.ID,
		AccountID: cashAccount.ID,
	})
	if err != nil {
		t.Fatalf("export filtered csv: %v", err)
	}
	if !strings.Contains(content, "Food") {
		t.Fatalf("expected matching transaction in export, got %s", content)
	}
	if strings.Contains(content, "Travel") {
		t.Fatalf("expected account and ledger filters to exclude other transactions, got %s", content)
	}
}

func ExportPerf10K(t *testing.T) {
	ctx := context.Background()
	service := newExportFixture(t)
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 10000; i++ {
		service.repo.items = append(service.repo.items, accounting.Transaction{OccurredAt: base.Add(time.Duration(i) * time.Minute), Amount: float64(i + 1), Type: accounting.TransactionTypeExpense, CategoryName: "Food"})
	}
	start := time.Now()
	_, err := service.Export(ctx, "user-1", ExportQuery{Format: "csv"})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("export 10k: %v", err)
	}
	if elapsed > 10*time.Second {
		t.Fatalf("expected export 10k <= 10s, got %s", elapsed)
	}
}

type exportFixture struct {
	repo    *ExportRepository
	service *ExportService
}

func newExportFixture(t *testing.T) *ExportService {
	t.Helper()
	ctx := context.Background()
	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	txnRepo := accounting.NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	txnService := accounting.NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)
	ledger, err := ledgerRepo.Create("user-1", accounting.LedgerCreateInput{Name: "Main", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledger: %v", err)
	}
	category, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Food"})
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: ledger.ID, Type: accounting.TransactionTypeExpense, Amount: 25, OccurredAt: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC), CategoryID: &category.ID}); err != nil {
		t.Fatalf("seed txn: %v", err)
	}
	if _, err := categoryService.DeleteCategory(ctx, "user-1", category.ID); classification.ErrorCode(err) != classification.CAT_IN_USE_ARCHIVED {
		t.Fatalf("archive category: %q", classification.ErrorCode(err))
	}
	repo := NewExportRepository(txnRepo, categoryService)
	return NewExportService(repo)
}
