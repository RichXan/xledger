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
	ledgerRepo, accountRepo, txnRepo, categoryService, txnService := newTrendFixture(t)
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

func TestTrend_UsesUserTimezoneOrUTC8Default(t *testing.T) {
	ctx := context.Background()
	ledgerRepo, accountRepo, txnRepo, categoryService, txnService := newTrendFixture(t)
	repo := NewRepository(accountRepo, txnRepo, categoryService)
	trend := NewTrendService(repo)

	ledger, err := ledgerRepo.Create("user-1", accounting.LedgerCreateInput{Name: "Main", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledger: %v", err)
	}
	instant := time.Date(2026, 3, 1, 16, 30, 0, 0, time.UTC)
	if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: ledger.ID, Type: accounting.TransactionTypeIncome, Amount: 88, OccurredAt: instant}); err != nil {
		t.Fatalf("seed timezone income: %v", err)
	}

	defaultTZ, err := trend.GetTrend(ctx, "user-1", TrendQuery{From: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC), Granularity: "day"})
	if err != nil {
		t.Fatalf("default timezone trend: %v", err)
	}
	if got := defaultTZ.Points[1].Income; got != 88 {
		t.Fatalf("expected UTC+8 default bucketing to place income in second day, got %#v", defaultTZ.Points)
	}

	laTZ, err := trend.GetTrend(ctx, "user-1", TrendQuery{From: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC), Granularity: "day", Timezone: "America/Los_Angeles"})
	if err != nil {
		t.Fatalf("explicit timezone trend: %v", err)
	}
	if got := laTZ.Points[0].Income; got != 88 {
		t.Fatalf("expected explicit timezone bucketing to place income in first day, got %#v", laTZ.Points)
	}
}

func TestTrend_EmptyWindowReturnsZeroSeries(t *testing.T) {
	repo := NewRepository(accounting.NewInMemoryAccountRepository(), accounting.NewInMemoryTransactionRepository(), classification.NewCategoryService(classification.NewInMemoryRepository()))
	trend := NewTrendService(repo)

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC)
	result, err := trend.GetTrend(context.Background(), "user-1", TrendQuery{From: from, To: to, Granularity: "day"})
	if err != nil {
		t.Fatalf("empty trend query: %v", err)
	}
	if len(result.Points) != 3 {
		t.Fatalf("expected 3 zero buckets, got %d", len(result.Points))
	}
	for _, point := range result.Points {
		if point.Income != 0 || point.Expense != 0 {
			t.Fatalf("expected zero bucket, got %#v", point)
		}
	}
}

func TestTrend_InvalidParams_ReturnsSTAT_QUERY_INVALID(t *testing.T) {
	repo := NewRepository(accounting.NewInMemoryAccountRepository(), accounting.NewInMemoryTransactionRepository(), classification.NewCategoryService(classification.NewInMemoryRepository()))
	trend := NewTrendService(repo)

	_, err := trend.GetTrend(context.Background(), "user-1", TrendQuery{From: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Granularity: "day"})
	if ErrorCode(err) != STAT_QUERY_INVALID {
		t.Fatalf("expected %s, got %q", STAT_QUERY_INVALID, ErrorCode(err))
	}
	_, err = trend.GetTrend(context.Background(), "user-1", TrendQuery{From: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Granularity: "monthish"})
	if ErrorCode(err) != STAT_QUERY_INVALID {
		t.Fatalf("expected %s for bad granularity, got %q", STAT_QUERY_INVALID, ErrorCode(err))
	}
	_, err = trend.GetTrend(context.Background(), "user-1", TrendQuery{From: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Granularity: "day", Timezone: "Mars/OlympusMons"})
	if ErrorCode(err) != STAT_QUERY_INVALID {
		t.Fatalf("expected %s for bad timezone, got %q", STAT_QUERY_INVALID, ErrorCode(err))
	}
}

func TestTrend_TimeoutReturnsSTAT_TIMEOUT_NoPartialPayload(t *testing.T) {
	trend := NewTrendService(&Repository{txnRepo: slowTransactionRepo{delay: 25 * time.Millisecond}})
	_, err := trend.GetTrend(context.Background(), "user-1", TrendQuery{From: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Granularity: "day", Timeout: 5 * time.Millisecond})
	if ErrorCode(err) != STAT_TIMEOUT {
		t.Fatalf("expected %s, got %q", STAT_TIMEOUT, ErrorCode(err))
	}
}

func TestTrend_ReadOnlyDegradationRejectsHeavyWindow(t *testing.T) {
	repo := NewRepository(accounting.NewInMemoryAccountRepository(), accounting.NewInMemoryTransactionRepository(), classification.NewCategoryService(classification.NewInMemoryRepository()))
	trend := NewTrendService(repo)
	_, err := trend.GetTrend(context.Background(), "user-1", TrendQuery{From: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), Granularity: "day"})
	if ErrorCode(err) != STAT_QUERY_INVALID {
		t.Fatalf("expected %s for heavy window rejection, got %q", STAT_QUERY_INVALID, ErrorCode(err))
	}
}

func newTrendFixture(t *testing.T) (*accounting.InMemoryLedgerRepository, *accounting.InMemoryAccountRepository, *accounting.InMemoryTransactionRepository, *classification.CategoryService, *accounting.TransactionService) {
	t.Helper()
	ledgerRepo := accounting.NewInMemoryLedgerRepository()
	accountRepo := accounting.NewInMemoryAccountRepository()
	txnRepo := accounting.NewInMemoryTransactionRepository()
	classificationRepo := classification.NewInMemoryRepository()
	categoryService := classification.NewCategoryService(classificationRepo)
	tagService := classification.NewTagService(classificationRepo)
	txnService := accounting.NewTransactionService(txnRepo, ledgerRepo, accountRepo, categoryService, tagService)
	return ledgerRepo, accountRepo, txnRepo, categoryService, txnService
}

type slowTransactionRepo struct{ delay time.Duration }

func (r slowTransactionRepo) Create(string, accounting.TransactionCreateInput) (accounting.Transaction, error) {
	panic("not used")
}
func (r slowTransactionRepo) GetByIDForUser(string, string) (accounting.Transaction, bool, error) {
	panic("not used")
}
func (r slowTransactionRepo) SaveByIDForUser(string, string, accounting.Transaction) (accounting.Transaction, bool, error) {
	panic("not used")
}
func (r slowTransactionRepo) DeleteByIDForUser(string, string) (bool, error) { panic("not used") }
func (r slowTransactionRepo) CreateTransferPair(string, string, accounting.TransactionCreateInput, accounting.TransactionCreateInput) (accounting.Transaction, error) {
	panic("not used")
}
func (r slowTransactionRepo) GetTransferPairByTxnID(string, string) ([]accounting.Transaction, error) {
	panic("not used")
}
func (r slowTransactionRepo) UpdateTransferPairAmount(string, string, float64, *int) (accounting.Transaction, error) {
	panic("not used")
}
func (r slowTransactionRepo) DeleteTransferPairByTxnID(string, string, *int) ([]string, error) {
	panic("not used")
}
func (r slowTransactionRepo) WithTransferPairLock(string, string, func() error) error {
	panic("not used")
}
func (r slowTransactionRepo) ListByUser(string, accounting.TransactionQuery) ([]accounting.Transaction, error) {
	time.Sleep(r.delay)
	return []accounting.Transaction{{Type: accounting.TransactionTypeIncome, Amount: 1, OccurredAt: time.Now().UTC()}}, nil
}
func (r slowTransactionRepo) ListByTransferPairForUser(string, string) ([]accounting.Transaction, error) {
	panic("not used")
}
func (r slowTransactionRepo) MarkBalancesRecalculated(string, string) error   { panic("not used") }
func (r slowTransactionRepo) MarkStatsInputRecalculated(string, string) error { panic("not used") }
