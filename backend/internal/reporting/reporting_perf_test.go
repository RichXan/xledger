package reporting

import (
	"context"
	"sort"
	"testing"
	"time"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
)

func TestTrend_Perf1K(t *testing.T) {
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
	for i := 0; i < 1000; i++ {
		typeName := accounting.TransactionTypeExpense
		if i%2 == 0 {
			typeName = accounting.TransactionTypeIncome
		}
		if _, err := txnService.CreateTransaction(ctx, "user-1", accounting.TransactionCreateInput{LedgerID: ledger.ID, Type: typeName, Amount: float64(i + 1), OccurredAt: base.Add(time.Duration(i) * time.Minute)}); err != nil {
			t.Fatalf("seed txn %d: %v", i, err)
		}
	}

	start := time.Now()
	durations := make([]time.Duration, 0, 20)
	for i := 0; i < 20; i++ {
		start = time.Now()
		result, err := trend.GetTrend(ctx, "user-1", TrendQuery{From: base, To: base.Add(8 * 24 * time.Hour), Granularity: "day"})
		durations = append(durations, time.Since(start))
		if err != nil {
			t.Fatalf("trend query run %d: %v", i, err)
		}
		if len(result.Points) == 0 {
			t.Fatalf("expected non-empty points on run %d", i)
		}
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p95 := durations[(len(durations)*95+99)/100-1]
	if p95 > time.Second {
		t.Fatalf("expected p95 1k trend query under 1s, got %s", p95)
	}
}
