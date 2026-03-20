package accounting

import (
	"context"
	"testing"
	"time"
)

func TestCreateTxn_LedgerRequired(t *testing.T) {
	repo := NewInMemoryTransactionRepository()
	service := NewTransactionService(repo)

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
	repo := NewInMemoryTransactionRepository()
	service := NewTransactionService(repo)

	_, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   "ledger-1",
		Type:       TransactionTypeExpense,
		Amount:     11,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create expense with nil account should pass: %v", err)
	}

	_, err = service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   "ledger-1",
		Type:       TransactionTypeIncome,
		Amount:     21,
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create income with nil account should pass: %v", err)
	}
}

func TestCreateTransfer_RequiresBothAccounts(t *testing.T) {
	repo := NewInMemoryTransactionRepository()
	service := NewTransactionService(repo)

	from := "acc-from"
	_, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      "ledger-1",
		FromAccountID: &from,
		Amount:        10,
		OccurredAt:    time.Now().UTC(),
	})
	if ErrorCode(err) != TXN_VALIDATION_FAILED {
		t.Fatalf("expected %s when to-account missing, got %q", TXN_VALIDATION_FAILED, ErrorCode(err))
	}

	to := "acc-to"
	_, err = service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:    "ledger-1",
		ToAccountID: &to,
		Amount:      10,
		OccurredAt:  time.Now().UTC(),
	})
	if ErrorCode(err) != TXN_VALIDATION_FAILED {
		t.Fatalf("expected %s when from-account missing, got %q", TXN_VALIDATION_FAILED, ErrorCode(err))
	}
}

func TestEditTxn_RecalculatesBalances(t *testing.T) {
	repo := NewInMemoryTransactionRepository()
	service := NewTransactionService(repo)

	txn, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   "ledger-1",
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
	repo := NewInMemoryTransactionRepository()
	service := NewTransactionService(repo)

	txn, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   "ledger-1",
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
	repo := NewInMemoryTransactionRepository()
	service := NewTransactionService(repo)

	txn, err := service.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:   "ledger-1",
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

	err = service.DeleteTransaction(context.Background(), "user-1", txn.ID)
	if err != nil {
		t.Fatalf("delete txn: %v", err)
	}

	if repo.StatsInputRecalculationCount() != 2 {
		t.Fatalf("expected stats-input recalculation count 2, got %d", repo.StatsInputRecalculationCount())
	}
}

func TestEditTxn_NotFound_ReturnsTXN_NOT_FOUND(t *testing.T) {
	repo := NewInMemoryTransactionRepository()
	service := NewTransactionService(repo)

	_, err := service.EditTransaction(context.Background(), "user-1", "missing", TransactionEditInput{Amount: 20})
	if ErrorCode(err) != TXN_NOT_FOUND {
		t.Fatalf("expected %s, got %q", TXN_NOT_FOUND, ErrorCode(err))
	}
}
