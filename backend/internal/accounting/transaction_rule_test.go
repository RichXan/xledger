package accounting

import (
	"context"
	"testing"
	"time"
)

func newTransactionServiceFixture(t *testing.T) (*InMemoryTransactionRepository, *TransactionService, Ledger, Account, Account) {
	t.Helper()
	ledgerRepo := NewInMemoryLedgerRepository()
	accountRepo := NewInMemoryAccountRepository()
	txnRepo := NewInMemoryTransactionRepository()
	service := NewTransactionService(txnRepo, ledgerRepo, accountRepo)

	ledger, err := ledgerRepo.Create("user-1", LedgerCreateInput{Name: "Main", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledger: %v", err)
	}
	fromAccount, err := accountRepo.Create("user-1", AccountCreateInput{Name: "Wallet", Type: "cash", InitialBalance: 100})
	if err != nil {
		t.Fatalf("seed from account: %v", err)
	}
	toAccount, err := accountRepo.Create("user-1", AccountCreateInput{Name: "Bank", Type: "bank", InitialBalance: 200})
	if err != nil {
		t.Fatalf("seed to account: %v", err)
	}

	return txnRepo, service, ledger, fromAccount, toAccount
}

func TestCreateTxn_LedgerRequired(t *testing.T) {
	_, service, _, _, _ := newTransactionServiceFixture(t)

	_, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		Type:       TransactionTypeExpense,
		Amount:     10,
		OccurredAt: time.Now().UTC(),
	})
	if ErrorCode(err) != TXN_VALIDATION_FAILED {
		t.Fatalf("expected %s, got %q", TXN_VALIDATION_FAILED, ErrorCode(err))
	}
}

func TestCreateTxn_AccountNullableForExpenseIncome(t *testing.T) {
	_, service, ledger, fromAccount, _ := newTransactionServiceFixture(t)

	_, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeExpense,
		Amount:     11,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create expense with nil account should pass: %v", err)
	}

	_, err = service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeIncome,
		Amount:     21,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create income with nil account should pass: %v", err)
	}

	_, err = service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeExpense,
		AccountID:  &fromAccount.ID,
		Amount:     31,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create expense with valid account should pass: %v", err)
	}
}

func TestCreateTxn_InvalidLedger_ReturnsValidationFailed(t *testing.T) {
	_, service, _, _, _ := newTransactionServiceFixture(t)

	_, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   "missing-ledger",
		Type:       TransactionTypeExpense,
		Amount:     20,
		OccurredAt: time.Now().UTC(),
	})
	if ErrorCode(err) != TXN_VALIDATION_FAILED {
		t.Fatalf("expected %s, got %q", TXN_VALIDATION_FAILED, ErrorCode(err))
	}
}

func TestCreateTxn_InvalidOptionalAccount_ReturnsValidationFailed(t *testing.T) {
	_, service, ledger, _, _ := newTransactionServiceFixture(t)
	account := "missing-account"

	_, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeExpense,
		AccountID:  &account,
		Amount:     20,
		OccurredAt: time.Now().UTC(),
	})
	if ErrorCode(err) != TXN_VALIDATION_FAILED {
		t.Fatalf("expected %s, got %q", TXN_VALIDATION_FAILED, ErrorCode(err))
	}
}

func TestCreateTransfer_RequiresBothAccounts(t *testing.T) {
	_, service, ledger, _, _ := newTransactionServiceFixture(t)

	from := "acc-from"
	_, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      ledger.ID,
		FromAccountID: &from,
		Amount:        10,
		OccurredAt:    time.Now().UTC(),
	})
	if ErrorCode(err) != TXN_VALIDATION_FAILED {
		t.Fatalf("expected %s when to-account missing, got %q", TXN_VALIDATION_FAILED, ErrorCode(err))
	}

	to := "acc-to"
	_, err = service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:    ledger.ID,
		ToAccountID: &to,
		Amount:      10,
		OccurredAt:  time.Now().UTC(),
	})
	if ErrorCode(err) != TXN_VALIDATION_FAILED {
		t.Fatalf("expected %s when from-account missing, got %q", TXN_VALIDATION_FAILED, ErrorCode(err))
	}
}

func TestCreateTransfer_InvalidAccounts_ReturnsValidationFailed(t *testing.T) {
	_, service, ledger, _, _ := newTransactionServiceFixture(t)
	from := "missing-from"
	to := "missing-to"

	_, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      ledger.ID,
		FromAccountID: &from,
		ToAccountID:   &to,
		Amount:        10,
		OccurredAt:    time.Now().UTC(),
	})
	if ErrorCode(err) != TXN_VALIDATION_FAILED {
		t.Fatalf("expected %s for invalid transfer accounts, got %q", TXN_VALIDATION_FAILED, ErrorCode(err))
	}
}

func TestCreateTransfer_ValidAccounts_Succeeds(t *testing.T) {
	repo, service, ledger, fromAccount, toAccount := newTransactionServiceFixture(t)

	_, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      ledger.ID,
		FromAccountID: ptr(" " + fromAccount.ID + " "),
		ToAccountID:   ptr(" " + toAccount.ID + " "),
		Amount:        10,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create transfer with valid accounts should pass: %v", err)
	}

	txns, listErr := repo.ListByUser("user-1", TransactionQuery{LedgerID: ledger.ID})
	if listErr != nil {
		t.Fatalf("list txns: %v", listErr)
	}
	if len(txns) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txns))
	}
	if ptrString(txns[0].FromAccountID) != fromAccount.ID || ptrString(txns[0].ToAccountID) != toAccount.ID {
		t.Fatalf("expected trimmed account IDs, got from=%q to=%q", ptrString(txns[0].FromAccountID), ptrString(txns[0].ToAccountID))
	}
}

func TestEditTxn_RecalculatesBalances(t *testing.T) {
	repo, service, ledger, _, _ := newTransactionServiceFixture(t)

	txn, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeExpense,
		Amount:     10,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("seed txn: %v", err)
	}

	_, err = service.EditTransaction(context.Background(), "user-1", txn.ID, TransactionEditInput{
		Amount: 99,
	})
	if err != nil {
		t.Fatalf("edit txn: %v", err)
	}

	if repo.BalanceRecalculationCount() != 1 {
		t.Fatalf("expected balance recalculation count 1, got %d", repo.BalanceRecalculationCount())
	}
}

func TestDeleteTxn_RecalculatesBalances(t *testing.T) {
	repo, service, ledger, _, _ := newTransactionServiceFixture(t)

	txn, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeExpense,
		Amount:     10,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("seed txn: %v", err)
	}

	err = service.DeleteTransaction(context.Background(), "user-1", txn.ID)
	if err != nil {
		t.Fatalf("delete txn: %v", err)
	}

	if repo.BalanceRecalculationCount() != 1 {
		t.Fatalf("expected balance recalculation count 1, got %d", repo.BalanceRecalculationCount())
	}
}

func TestEditDeleteTxn_RecalculatesStatsInput(t *testing.T) {
	repo, service, ledger, _, _ := newTransactionServiceFixture(t)

	txn, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   ledger.ID,
		Type:       TransactionTypeExpense,
		Amount:     10,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("seed txn: %v", err)
	}

	_, err = service.EditTransaction(context.Background(), "user-1", txn.ID, TransactionEditInput{Amount: 40})
	if err != nil {
		t.Fatalf("edit txn: %v", err)
	}
	if repo.StatsInputRecalculationCount() != 1 {
		t.Fatalf("expected stats-input recalculation count 1 after edit, got %d", repo.StatsInputRecalculationCount())
	}

	err = service.DeleteTransaction(context.Background(), "user-1", txn.ID)
	if err != nil {
		t.Fatalf("delete txn: %v", err)
	}

	if repo.StatsInputRecalculationCount() != 2 {
		t.Fatalf("expected stats-input recalculation count 2, got %d", repo.StatsInputRecalculationCount())
	}
}

func TestEditTxn_NotFound_ReturnsTXN_NOT_FOUND(t *testing.T) {
	_, service, _, _, _ := newTransactionServiceFixture(t)

	_, err := service.EditTransaction(context.Background(), "user-1", "missing", TransactionEditInput{Amount: 20})
	if ErrorCode(err) != TXN_NOT_FOUND {
		t.Fatalf("expected %s, got %q", TXN_NOT_FOUND, ErrorCode(err))
	}
}

func ptr(value string) *string {
	return &value
}
