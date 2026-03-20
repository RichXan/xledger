package accounting

import (
	"context"
	"strings"
	"time"
)

const (
	TXN_VALIDATION_FAILED = "TXN_VALIDATION_FAILED"
	TXN_NOT_FOUND         = "TXN_NOT_FOUND"
)

type TransactionService struct {
	repo        TransactionRepository
	ledgerRepo  LedgerRepository
	accountRepo AccountRepository
}

func NewTransactionService(repo TransactionRepository, ledgerRepo LedgerRepository, accountRepo AccountRepository) *TransactionService {
	return &TransactionService{repo: repo, ledgerRepo: ledgerRepo, accountRepo: accountRepo}
}

func (s *TransactionService) CreateTransaction(_ context.Context, userID string, input TransactionCreateInput) (Transaction, error) {
	userID = strings.TrimSpace(userID)
	input.LedgerID = strings.TrimSpace(input.LedgerID)
	input.Type = strings.TrimSpace(input.Type)
	input.AccountID = normalizeOptionalID(input.AccountID)
	input.FromAccountID = normalizeOptionalID(input.FromAccountID)
	input.ToAccountID = normalizeOptionalID(input.ToAccountID)
	if userID == "" || input.LedgerID == "" {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}
	ledgerExists, ledgerErr := s.ledgerExists(userID, input.LedgerID)
	if ledgerErr != nil {
		return Transaction{}, ledgerErr
	}
	if !ledgerExists {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	switch input.Type {
	case TransactionTypeIncome, TransactionTypeExpense:
		accountExists, accountErr := s.accountOptionalExists(userID, input.AccountID)
		if accountErr != nil {
			return Transaction{}, accountErr
		}
		if !accountExists {
			return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
		}
	case TransactionTypeTransfer:
		if strings.TrimSpace(ptrString(input.FromAccountID)) == "" || strings.TrimSpace(ptrString(input.ToAccountID)) == "" {
			return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
		}
		fromExists, fromErr := s.accountRequiredExists(userID, input.FromAccountID)
		if fromErr != nil {
			return Transaction{}, fromErr
		}
		toExists, toErr := s.accountRequiredExists(userID, input.ToAccountID)
		if toErr != nil {
			return Transaction{}, toErr
		}
		if !fromExists || !toExists {
			return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
		}
	default:
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	if input.Amount <= 0 {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	}

	return s.repo.Create(userID, input)
}

func (s *TransactionService) CreateTransfer(ctx context.Context, userID string, input TransactionTransferInput) (Transaction, error) {
	if strings.TrimSpace(ptrString(input.FromAccountID)) == "" || strings.TrimSpace(ptrString(input.ToAccountID)) == "" {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	return s.CreateTransaction(ctx, userID, TransactionCreateInput{
		LedgerID:      input.LedgerID,
		Type:          TransactionTypeTransfer,
		FromAccountID: input.FromAccountID,
		ToAccountID:   input.ToAccountID,
		Amount:        input.Amount,
		OccurredAt:    input.OccurredAt,
	})
}

func (s *TransactionService) EditTransaction(_ context.Context, userID string, txnID string, input TransactionEditInput) (Transaction, error) {
	userID = strings.TrimSpace(userID)
	txnID = strings.TrimSpace(txnID)
	if userID == "" || txnID == "" || input.Amount <= 0 {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	txn, found, err := s.repo.GetByIDForUser(userID, txnID)
	if err != nil {
		return Transaction{}, err
	}
	if !found {
		return Transaction{}, &contractError{code: TXN_NOT_FOUND}
	}

	txn.Amount = input.Amount
	updated, saved, saveErr := s.repo.SaveByIDForUser(userID, txnID, txn)
	if saveErr != nil {
		return Transaction{}, saveErr
	}
	if !saved {
		return Transaction{}, &contractError{code: TXN_NOT_FOUND}
	}

	if err := s.recalculateForLedger(userID, txn.LedgerID); err != nil {
		return Transaction{}, err
	}

	return updated, nil
}

func (s *TransactionService) DeleteTransaction(_ context.Context, userID string, txnID string) error {
	userID = strings.TrimSpace(userID)
	txnID = strings.TrimSpace(txnID)
	if userID == "" || txnID == "" {
		return &contractError{code: TXN_VALIDATION_FAILED}
	}

	txn, found, err := s.repo.GetByIDForUser(userID, txnID)
	if err != nil {
		return err
	}
	if !found {
		return &contractError{code: TXN_NOT_FOUND}
	}

	deleted, deleteErr := s.repo.DeleteByIDForUser(userID, txnID)
	if deleteErr != nil {
		return deleteErr
	}
	if !deleted {
		return &contractError{code: TXN_NOT_FOUND}
	}

	return s.recalculateForLedger(userID, txn.LedgerID)
}

func (s *TransactionService) ListTransactions(_ context.Context, userID string, query TransactionQuery) ([]Transaction, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, &contractError{code: TXN_VALIDATION_FAILED}
	}
	return s.repo.ListByUser(userID, query)
}

func (s *TransactionService) recalculateForLedger(userID string, ledgerID string) error {
	if err := s.repo.MarkBalancesRecalculated(userID, ledgerID); err != nil {
		return err
	}
	if err := s.repo.MarkStatsInputRecalculated(userID, ledgerID); err != nil {
		return err
	}
	return nil
}

func ptrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *TransactionService) ledgerExists(userID string, ledgerID string) (bool, error) {
	if s.ledgerRepo == nil {
		return true, nil
	}
	_, found, err := s.ledgerRepo.GetByIDForUser(userID, ledgerID)
	if err != nil {
		return false, err
	}
	return found, nil
}

func (s *TransactionService) accountOptionalExists(userID string, accountID *string) (bool, error) {
	if strings.TrimSpace(ptrString(accountID)) == "" {
		return true, nil
	}
	return s.accountRequiredExists(userID, accountID)
}

func (s *TransactionService) accountRequiredExists(userID string, accountID *string) (bool, error) {
	if s.accountRepo == nil {
		return true, nil
	}
	account := strings.TrimSpace(ptrString(accountID))
	if account == "" {
		return false, nil
	}
	_, found, err := s.accountRepo.GetByIDForUser(userID, account)
	if err != nil {
		return false, err
	}
	return found, nil
}

func normalizeOptionalID(value *string) *string {
	trimmed := strings.TrimSpace(ptrString(value))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
