package accounting

import (
	"context"
	"testing"
	"time"
)

func newTransferServiceFixture(t *testing.T) (*InMemoryTransactionRepository, *InMemoryLedgerRepository, *TransactionService, Ledger, Ledger, Account, Account) {
	t.Helper()

	ledgerRepo := NewInMemoryLedgerRepository()
	accountRepo := NewInMemoryAccountRepository()
	txnRepo := NewInMemoryTransactionRepository()
	service := NewTransactionService(txnRepo, ledgerRepo, accountRepo)

	fromLedger, err := ledgerRepo.Create("user-1", LedgerCreateInput{Name: "Primary", IsDefault: true})
	if err != nil {
		t.Fatalf("seed from ledger: %v", err)
	}
	toLedger, err := ledgerRepo.Create("user-1", LedgerCreateInput{Name: "Secondary"})
	if err != nil {
		t.Fatalf("seed to ledger: %v", err)
	}

	fromAccount, err := accountRepo.Create("user-1", AccountCreateInput{Name: "Wallet", Type: "cash", InitialBalance: 100})
	if err != nil {
		t.Fatalf("seed from account: %v", err)
	}
	toAccount, err := accountRepo.Create("user-1", AccountCreateInput{Name: "Bank", Type: "bank", InitialBalance: 200})
	if err != nil {
		t.Fatalf("seed to account: %v", err)
	}

	return txnRepo, ledgerRepo, service, fromLedger, toLedger, fromAccount, toAccount
}

func TestTransfer_CreateEditDeletePairAtomically(t *testing.T) {
	repo, _, service, fromLedger, _, fromAccount, toAccount := newTransferServiceFixture(t)

	created, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      fromLedger.ID,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		Amount:        25,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	pairID := ptrString(created.TransferPairID)
	if pairID == "" {
		t.Fatalf("expected transfer pair id to be set")
	}

	pair, listErr := repo.ListByTransferPairForUser("user-1", pairID)
	if listErr != nil {
		t.Fatalf("list pair after create: %v", listErr)
	}
	if len(pair) != 2 {
		t.Fatalf("expected 2 transfer sides after create, got %d", len(pair))
	}

	edited, editErr := service.EditTransaction(context.Background(), "user-1", created.ID, TransactionEditInput{Amount: 40})
	if editErr != nil {
		t.Fatalf("edit transfer: %v", editErr)
	}
	if edited.Amount != 40 {
		t.Fatalf("expected edited amount 40, got %v", edited.Amount)
	}

	pair, listErr = repo.ListByTransferPairForUser("user-1", pairID)
	if listErr != nil {
		t.Fatalf("list pair after edit: %v", listErr)
	}
	for _, side := range pair {
		if side.Amount != 40 {
			t.Fatalf("expected amount 40 across pair, got %v", side.Amount)
		}
	}

	if deleteErr := service.DeleteTransaction(context.Background(), "user-1", created.ID, nil); deleteErr != nil {
		t.Fatalf("delete transfer: %v", deleteErr)
	}

	pair, listErr = repo.ListByTransferPairForUser("user-1", pairID)
	if listErr != nil {
		t.Fatalf("list pair after delete: %v", listErr)
	}
	if len(pair) != 0 {
		t.Fatalf("expected pair to be fully deleted, remaining=%d", len(pair))
	}
}

func TestTransfer_ConflictReturnsTXN_CONFLICT(t *testing.T) {
	repo, _, service, fromLedger, _, fromAccount, toAccount := newTransferServiceFixture(t)

	created, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      fromLedger.ID,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		Amount:        25,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	pairID := ptrString(created.TransferPairID)
	lockErr := repo.WithTransferPairLock("user-1", pairID, func() error {
		_, editErr := service.EditTransaction(context.Background(), "user-1", created.ID, TransactionEditInput{Amount: 33})
		if ErrorCode(editErr) != TXN_CONFLICT {
			t.Fatalf("expected %s, got %q", TXN_CONFLICT, ErrorCode(editErr))
		}
		return nil
	})
	if lockErr != nil {
		t.Fatalf("acquire lock: %v", lockErr)
	}
}

func TestTransfer_VersionConflictReturnsTXN_CONFLICT(t *testing.T) {
	_, _, service, fromLedger, _, fromAccount, toAccount := newTransferServiceFixture(t)

	created, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      fromLedger.ID,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		Amount:        25,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	staleVersion := created.Version
	_, err = service.EditTransaction(context.Background(), "user-1", created.ID, TransactionEditInput{Amount: 30, Version: &staleVersion})
	if err != nil {
		t.Fatalf("first edit with current version should pass: %v", err)
	}

	_, err = service.EditTransaction(context.Background(), "user-1", created.ID, TransactionEditInput{Amount: 35, Version: &staleVersion})
	if ErrorCode(err) != TXN_CONFLICT {
		t.Fatalf("expected %s, got %q", TXN_CONFLICT, ErrorCode(err))
	}
}

func TestTransfer_BilateralMismatchReturnsTXN_CONFLICT(t *testing.T) {
	repo, _, service, fromLedger, _, fromAccount, toAccount := newTransferServiceFixture(t)

	created, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      fromLedger.ID,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		Amount:        25,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	pair, listErr := repo.ListByTransferPairForUser("user-1", ptrString(created.TransferPairID))
	if listErr != nil {
		t.Fatalf("list pair: %v", listErr)
	}
	if len(pair) != 2 {
		t.Fatalf("expected pair with 2 sides, got %d", len(pair))
	}

	deleteID := pair[0].ID
	if deleteID == created.ID {
		deleteID = pair[1].ID
	}
	if _, deleteErr := repo.DeleteByIDForUser("user-1", deleteID); deleteErr != nil {
		t.Fatalf("delete one side: %v", deleteErr)
	}

	_, err = service.EditTransaction(context.Background(), "user-1", created.ID, TransactionEditInput{Amount: 30})
	if ErrorCode(err) != TXN_CONFLICT {
		t.Fatalf("expected %s, got %q", TXN_CONFLICT, ErrorCode(err))
	}
}

func TestTransfer_CrossLedgerAllowed(t *testing.T) {
	repo, _, service, fromLedger, toLedger, fromAccount, toAccount := newTransferServiceFixture(t)

	created, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      fromLedger.ID,
		FromLedgerID:  &fromLedger.ID,
		ToLedgerID:    &toLedger.ID,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		Amount:        65,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("cross-ledger transfer should be allowed: %v", err)
	}

	pair, listErr := repo.ListByTransferPairForUser("user-1", ptrString(created.TransferPairID))
	if listErr != nil {
		t.Fatalf("list pair: %v", listErr)
	}
	if len(pair) != 2 {
		t.Fatalf("expected 2 transfer sides, got %d", len(pair))
	}

	if pair[0].LedgerID == pair[1].LedgerID {
		t.Fatalf("expected different ledgers for cross-ledger transfer, got %q and %q", pair[0].LedgerID, pair[1].LedgerID)
	}
}

func TestTransfer_CrossLedger_KeepsLedgerScopedAggregationInputs(t *testing.T) {
	repo, _, service, fromLedger, toLedger, fromAccount, toAccount := newTransferServiceFixture(t)

	_, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      fromLedger.ID,
		FromLedgerID:  &fromLedger.ID,
		ToLedgerID:    &toLedger.ID,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		Amount:        65,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create cross-ledger transfer: %v", err)
	}

	if repo.BalanceRecalculationCountForLedger(fromLedger.ID) != 1 {
		t.Fatalf("expected balance recalculation for from-ledger once, got %d", repo.BalanceRecalculationCountForLedger(fromLedger.ID))
	}
	if repo.BalanceRecalculationCountForLedger(toLedger.ID) != 1 {
		t.Fatalf("expected balance recalculation for to-ledger once, got %d", repo.BalanceRecalculationCountForLedger(toLedger.ID))
	}
	if repo.StatsInputRecalculationCountForLedger(fromLedger.ID) != 1 {
		t.Fatalf("expected stats-input recalculation for from-ledger once, got %d", repo.StatsInputRecalculationCountForLedger(fromLedger.ID))
	}
	if repo.StatsInputRecalculationCountForLedger(toLedger.ID) != 1 {
		t.Fatalf("expected stats-input recalculation for to-ledger once, got %d", repo.StatsInputRecalculationCountForLedger(toLedger.ID))
	}
}

func TestTransfer_CrossLedger_ListByLedgerIncludesDestinationSide(t *testing.T) {
	repo, _, service, fromLedger, toLedger, fromAccount, toAccount := newTransferServiceFixture(t)

	created, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      fromLedger.ID,
		FromLedgerID:  &fromLedger.ID,
		ToLedgerID:    &toLedger.ID,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		Amount:        88,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create cross-ledger transfer: %v", err)
	}

	fromItems, fromErr := repo.ListByUser("user-1", TransactionQuery{LedgerID: fromLedger.ID})
	if fromErr != nil {
		t.Fatalf("list from-ledger: %v", fromErr)
	}
	if len(fromItems) != 1 {
		t.Fatalf("expected 1 item in from-ledger list, got %d", len(fromItems))
	}

	toItems, toErr := repo.ListByUser("user-1", TransactionQuery{LedgerID: toLedger.ID})
	if toErr != nil {
		t.Fatalf("list to-ledger: %v", toErr)
	}
	if len(toItems) != 1 {
		t.Fatalf("expected 1 item in to-ledger list, got %d", len(toItems))
	}
	if ptrString(toItems[0].TransferPairID) != ptrString(created.TransferPairID) {
		t.Fatalf("expected destination list item from same pair, got %q want %q", ptrString(toItems[0].TransferPairID), ptrString(created.TransferPairID))
	}
}

func TestTransfer_CreatePairFailure_RollsBackAllOrNothing(t *testing.T) {
	repo, _, service, fromLedger, _, fromAccount, toAccount := newTransferServiceFixture(t)
	repo.InjectTransferCreateFailureAfterFrom()

	_, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      fromLedger.ID,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		Amount:        25,
		OccurredAt:    time.Now().UTC(),
	})
	if err == nil {
		t.Fatalf("expected create transfer to fail")
	}

	items, listErr := repo.ListByUser("user-1", TransactionQuery{})
	if listErr != nil {
		t.Fatalf("list after failed create: %v", listErr)
	}
	if len(items) != 0 {
		t.Fatalf("expected zero transactions after failed create, got %d", len(items))
	}
}

func TestTransfer_EditPairFailure_RollsBackAllOrNothing(t *testing.T) {
	repo, _, service, fromLedger, _, fromAccount, toAccount := newTransferServiceFixture(t)

	created, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      fromLedger.ID,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		Amount:        41,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	repo.InjectTransferUpdateFailureAfterFirst()
	_, err = service.EditTransaction(context.Background(), "user-1", created.ID, TransactionEditInput{Amount: 99})
	if err == nil {
		t.Fatalf("expected edit transfer to fail")
	}

	pair, pairErr := repo.ListByTransferPairForUser("user-1", ptrString(created.TransferPairID))
	if pairErr != nil {
		t.Fatalf("list pair: %v", pairErr)
	}
	if len(pair) != 2 {
		t.Fatalf("expected intact pair, got %d sides", len(pair))
	}
	for _, side := range pair {
		if side.Amount != 41 {
			t.Fatalf("expected rollback to original amount 41, got %v", side.Amount)
		}
	}
}

func TestTransfer_DeletePairFailure_RollsBackAllOrNothing(t *testing.T) {
	repo, _, service, fromLedger, _, fromAccount, toAccount := newTransferServiceFixture(t)

	created, err := service.CreateTransfer(context.Background(), "user-1", TransactionTransferInput{
		LedgerID:      fromLedger.ID,
		FromAccountID: &fromAccount.ID,
		ToAccountID:   &toAccount.ID,
		Amount:        51,
		OccurredAt:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	repo.InjectTransferDeleteFailureAfterFirst()
	err = service.DeleteTransaction(context.Background(), "user-1", created.ID, nil)
	if err == nil {
		t.Fatalf("expected delete transfer to fail")
	}

	pair, pairErr := repo.ListByTransferPairForUser("user-1", ptrString(created.TransferPairID))
	if pairErr != nil {
		t.Fatalf("list pair: %v", pairErr)
	}
	if len(pair) != 2 {
		t.Fatalf("expected pair rollback to keep 2 sides, got %d", len(pair))
	}
}
