package reporting

import (
	"context"
	"testing"
	"time"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
)

func TestOverview_TotalAssetsIndependentFromLedgerFilter(t *testing.T) {
	ctx := context.Background()
	deps := newReportingFixture(t)

	overview := NewOverviewService(NewRepository(deps.accountRepo, deps.txnRepo, deps.categoryService))
	all, err := overview.GetOverview(ctx, "user-1", OverviewQuery{})
	if err != nil {
		t.Fatalf("overview without ledger filter: %v", err)
	}
	filtered, err := overview.GetOverview(ctx, "user-1", OverviewQuery{LedgerID: deps.secondaryLedger.ID})
	if err != nil {
		t.Fatalf("overview with ledger filter: %v", err)
	}
	if all.TotalAssets != filtered.TotalAssets {
		t.Fatalf("expected total_assets invariant across ledger filter, got %v vs %v", all.TotalAssets, filtered.TotalAssets)
	}
	if all.TotalAssets != 300 {
		t.Fatalf("expected total_assets 300 from account balances, got %v", all.TotalAssets)
	}
}

func TestOverview_AccountNullIncludedInIncomeExpenseNotAssets(t *testing.T) {
	ctx := context.Background()
	deps := newReportingFixture(t)

	overview := NewOverviewService(NewRepository(deps.accountRepo, deps.txnRepo, deps.categoryService))
	result, err := overview.GetOverview(ctx, "user-1", OverviewQuery{})
	if err != nil {
		t.Fatalf("overview query: %v", err)
	}
	if result.Income != 50 {
		t.Fatalf("expected null-account income to be included, got %v", result.Income)
	}
	if result.Expense != 30 {
		t.Fatalf("expected null-account expense to be included, got %v", result.Expense)
	}
	if result.TotalAssets != 300 {
		t.Fatalf("expected total_assets to exclude null-account txns, got %v", result.TotalAssets)
	}
}

func TestOverview_TransferOffsetsExcludedFromIncomeExpense(t *testing.T) {
	ctx := context.Background()
	deps := newReportingFixture(t)

	overview := NewOverviewService(NewRepository(deps.accountRepo, deps.txnRepo, deps.categoryService))
	result, err := overview.GetOverview(ctx, "user-1", OverviewQuery{})
	if err != nil {
		t.Fatalf("overview query: %v", err)
	}
	if result.Income != 50 || result.Expense != 30 || result.Net != 20 {
		t.Fatalf("expected transfer pair excluded from income/expense/net, got income=%v expense=%v net=%v", result.Income, result.Expense, result.Net)
	}
}

type reportingFixture struct {
	accountRepo     *accounting.InMemoryAccountRepository
	txnRepo         *accounting.InMemoryTransactionRepository
	categoryService *classification.CategoryService
	primaryLedger   accounting.Ledger
	secondaryLedger accounting.Ledger
}

func newReportingFixture(t *testing.T) reportingFixture {
	t.Helper()
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
	wallet, err := accountRepo.Create("user-1", accounting.AccountCreateInput{Name: "Wallet", Type: "cash", InitialBalance: 100})
	if err != nil {
		t.Fatalf("seed wallet: %v", err)
	}
	bank, err := accountRepo.Create("user-1", accounting.AccountCreateInput{Name: "Bank", Type: "bank", InitialBalance: 200})
	if err != nil {
		t.Fatalf("seed bank: %v", err)
	}
	food, err := categoryService.CreateCategory(ctx, "user-1", classification.CategoryCreateInput{Name: "Food"})
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}

	base := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: primaryLedger.ID, Type: accounting.TransactionTypeIncome, Amount: 50, OccurredAt: base}); err != nil {
		t.Fatalf("seed null-account income: %v", err)
	}
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: primaryLedger.ID, Type: accounting.TransactionTypeExpense, Amount: 30, OccurredAt: base.Add(time.Hour), CategoryID: &food.ID}); err != nil {
		t.Fatalf("seed null-account expense: %v", err)
	}
	if _, err := txnService.CreateTransfer(ctx, "user-1", accounting.TransactionTransferInput{LedgerID: primaryLedger.ID, FromLedgerID: &primaryLedger.ID, ToLedgerID: &secondaryLedger.ID, FromAccountID: &wallet.ID, ToAccountID: &bank.ID, Amount: 40, OccurredAt: base.Add(2 * time.Hour)}); err != nil {
		t.Fatalf("seed transfer pair: %v", err)
	}
	if _, err := categoryService.DeleteCategory(ctx, "user-1", food.ID); classification.ErrorCode(err) != classification.CAT_IN_USE_ARCHIVED {
		t.Fatalf("archive referenced category: %q", classification.ErrorCode(err))
	}

	return reportingFixture{accountRepo: accountRepo, txnRepo: txnRepo, categoryService: categoryService, primaryLedger: primaryLedger, secondaryLedger: secondaryLedger}
}
