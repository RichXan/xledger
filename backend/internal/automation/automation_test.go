package automation

import (
	"context"
	"testing"
	"time"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
)

func TestQuickEntry_ForwardedPATInvalidSignal_ReturnsQE_PAT_INVALID_NoWrite(t *testing.T) {
	fixture := newAutomationFixture(t)
	result, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "lunch 25", PATInvalid: true})
	if ErrorCode(err) != QE_PAT_INVALID {
		t.Fatalf("expected %s, got %q", QE_PAT_INVALID, ErrorCode(err))
	}
	if result.WroteTransaction {
		t.Fatalf("expected no write on invalid PAT signal")
	}
	if fixture.txnRepoCount() != 0 {
		t.Fatalf("expected zero transactions, got %d", fixture.txnRepoCount())
	}
}

func TestQuickEntry_LLMUnavailable_ReturnsQE_LLM_UNAVAILABLE_NoWrite(t *testing.T) {
	fixture := newAutomationFixture(t)
	result, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "lunch 25", LLMUnavailable: true})
	if ErrorCode(err) != QE_LLM_UNAVAILABLE {
		t.Fatalf("expected %s, got %q", QE_LLM_UNAVAILABLE, ErrorCode(err))
	}
	if result.WroteTransaction || fixture.txnRepoCount() != 0 {
		t.Fatalf("expected no write on llm unavailable")
	}
}

func TestQuickEntry_ParseFailed_ReturnsQE_PARSE_FAILED_NoWrite(t *testing.T) {
	fixture := newAutomationFixture(t)
	result, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "???", ParseFailed: true})
	if ErrorCode(err) != QE_PARSE_FAILED {
		t.Fatalf("expected %s, got %q", QE_PARSE_FAILED, ErrorCode(err))
	}
	if result.WroteTransaction || fixture.txnRepoCount() != 0 {
		t.Fatalf("expected no write on parse failure")
	}
}

func TestQuickEntry_Timeout_ReturnsQE_TIMEOUT_NoWrite(t *testing.T) {
	fixture := newAutomationFixture(t)
	result, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "lunch 25", TimedOut: true})
	if ErrorCode(err) != QE_TIMEOUT {
		t.Fatalf("expected %s, got %q", QE_TIMEOUT, ErrorCode(err))
	}
	if result.WroteTransaction || fixture.txnRepoCount() != 0 {
		t.Fatalf("expected no write on timeout")
	}
}

func TestQuickEntry_Success_ReturnsStructuredConfirmation(t *testing.T) {
	fixture := newAutomationFixture(t)
	result, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "lunch 25", ParsedAmount: 25, ParsedType: accounting.TransactionTypeExpense})
	if err != nil {
		t.Fatalf("unexpected quick-entry error: %v", err)
	}
	if !result.WroteTransaction || result.Amount != 25 || result.Type != accounting.TransactionTypeExpense {
		t.Fatalf("expected structured confirmation, got %#v", result)
	}
	if fixture.txnRepoCount() != 1 {
		t.Fatalf("expected exactly one transaction, got %d", fixture.txnRepoCount())
	}
}

func TestQuickEntry_AmbiguousAccountHint_DowngradesToNullAccount_WithHint(t *testing.T) {
	fixture := newAutomationFixture(t)
	_, err := fixture.accountRepo.Create("user-1", accounting.AccountCreateInput{Name: "Wallet", Type: "cash", InitialBalance: 0})
	if err != nil {
		t.Fatalf("seed account1: %v", err)
	}
	_, err = fixture.accountRepo.Create("user-1", accounting.AccountCreateInput{Name: "Wallet 2", Type: "cash", InitialBalance: 0})
	if err != nil {
		t.Fatalf("seed account2: %v", err)
	}
	result, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "lunch 25 wallet", ParsedAmount: 25, ParsedType: accounting.TransactionTypeExpense, AccountHint: "wallet"})
	if err != nil {
		t.Fatalf("unexpected quick-entry error: %v", err)
	}
	if result.AccountID != nil {
		t.Fatalf("expected nil account id on ambiguous hint, got %#v", result.AccountID)
	}
	if result.AccountHintStatus == "" {
		t.Fatalf("expected account hint downgrade note, got %#v", result)
	}
}

func TestQuickEntry_Idempotency_SamePATSameNormalizedTextSameKey_Dedupes(t *testing.T) {
	fixture := newAutomationFixture(t)
	first, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "lunch 25", ParsedAmount: 25, ParsedType: accounting.TransactionTypeExpense})
	if err != nil {
		t.Fatalf("first quick-entry: %v", err)
	}
	second, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "lunch 25", ParsedAmount: 25, ParsedType: accounting.TransactionTypeExpense})
	if err != nil {
		t.Fatalf("second quick-entry: %v", err)
	}
	if second.TransactionID != first.TransactionID {
		t.Fatalf("expected deduped replay to return same transaction, got %#v vs %#v", first, second)
	}
	if fixture.txnRepoCount() != 1 {
		t.Fatalf("expected dedup to keep one transaction, got %d", fixture.txnRepoCount())
	}
}

func TestQuickEntry_Idempotency_SamePATDifferentNormalizedText_NoDedup(t *testing.T) {
	fixture := newAutomationFixture(t)
	_, _ = fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "lunch 25", ParsedAmount: 25, ParsedType: accounting.TransactionTypeExpense})
	_, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "dinner 30", ParsedAmount: 30, ParsedType: accounting.TransactionTypeExpense})
	if err != nil {
		t.Fatalf("unexpected second quick-entry error: %v", err)
	}
	if fixture.txnRepoCount() != 2 {
		t.Fatalf("expected no dedup on different normalized text, got %d", fixture.txnRepoCount())
	}
}

func TestQuickEntry_Idempotency_ReplayAfter24h_NotDeduped(t *testing.T) {
	now := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	fixture := newAutomationFixtureAt(t, now)
	_, _ = fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "lunch 25", ParsedAmount: 25, ParsedType: accounting.TransactionTypeExpense})
	fixture.adapter.SetNow(func() time.Time { return now.Add(25 * time.Hour) })
	_, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "lunch 25", ParsedAmount: 25, ParsedType: accounting.TransactionTypeExpense})
	if err != nil {
		t.Fatalf("unexpected replay-after-24h error: %v", err)
	}
	if fixture.txnRepoCount() != 2 {
		t.Fatalf("expected replay after 24h not deduped, got %d", fixture.txnRepoCount())
	}
}

func TestQuickEntry_GoValidationError_IsPassthrough(t *testing.T) {
	fixture := newAutomationFixture(t)
	_, err := fixture.adapter.Process(context.Background(), QuickEntryRequest{UserID: "user-1", PATID: "pat-1", IdempotencyKey: "qe-1", Text: "invalid", ParsedAmount: 0, ParsedType: accounting.TransactionTypeExpense})
	if accounting.ErrorCode(err) != accounting.TXN_VALIDATION_FAILED {
		t.Fatalf("expected passthrough %s, got %q", accounting.TXN_VALIDATION_FAILED, accounting.ErrorCode(err))
	}
}

func TestManualBookkeepingFlow_UnchangedWhenAutomationEnabled(t *testing.T) {
	fixture := newAutomationFixture(t)
	_, err := fixture.txnService.CreateTransaction(context.Background(), "user-1", accounting.TransactionCreateInput{LedgerID: fixture.ledger.ID, Type: accounting.TransactionTypeExpense, Amount: 15, OccurredAt: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("manual transaction should still work: %v", err)
	}
	if fixture.txnRepoCount() != 1 {
		t.Fatalf("expected manual bookkeeping unchanged, got %d transactions", fixture.txnRepoCount())
	}
}

type automationFixture struct {
	adapter     *QuickEntryAdapter
	txnService  *accounting.TransactionService
	txnRepo     *accounting.InMemoryTransactionRepository
	accountRepo *accounting.InMemoryAccountRepository
	ledger      accounting.Ledger
}

func newAutomationFixture(t *testing.T) *automationFixture {
	return newAutomationFixtureAt(t, time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC))
}

func newAutomationFixtureAt(t *testing.T, now time.Time) *automationFixture {
	t.Helper()
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
	adapter := NewQuickEntryAdapter(func() time.Time { return now }, txnService, ledger.ID)
	return &automationFixture{adapter: adapter, txnService: txnService, txnRepo: txnRepo, accountRepo: accountRepo, ledger: ledger}
}

func (f *automationFixture) txnRepoCount() int {
	items, err := f.txnRepo.ListByUser("user-1", accounting.TransactionQuery{})
	if err != nil {
		panic(err)
	}
	return len(items)
}
