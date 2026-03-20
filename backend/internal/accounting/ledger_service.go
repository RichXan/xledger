package accounting

import (
	"context"
	"errors"
)

const (
	LEDGER_DEFAULT_IMMUTABLE = "LEDGER_DEFAULT_IMMUTABLE"
	LEDGER_NOT_FOUND         = "LEDGER_NOT_FOUND"
	ACCOUNT_NOT_FOUND        = "ACCOUNT_NOT_FOUND"
	ACCOUNT_INVALID          = "ACCOUNT_INVALID"
)

type contractError struct {
	code string
	err  error
}

func (e *contractError) Error() string {
	if e.err == nil {
		return e.code
	}
	return e.code + ": " + e.err.Error()
}

func (e *contractError) Unwrap() error {
	return e.err
}

func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	var coded *contractError
	if errors.As(err, &coded) {
		return coded.code
	}
	return ""
}

type LedgerService struct {
	repo LedgerRepository
}

func NewLedgerService(repo LedgerRepository) *LedgerService {
	return &LedgerService{repo: repo}
}

func (s *LedgerService) DeleteLedger(_ context.Context, userID string, ledgerID string) error {
	ledger, found, err := s.repo.GetByIDForUser(userID, ledgerID)
	if err != nil {
		return err
	}
	if !found {
		return &contractError{code: LEDGER_NOT_FOUND}
	}
	if ledger.IsDefault {
		return &contractError{code: LEDGER_DEFAULT_IMMUTABLE}
	}

	deleted, err := s.repo.DeleteByIDForUser(userID, ledgerID)
	if err != nil {
		return err
	}
	if !deleted {
		return &contractError{code: LEDGER_NOT_FOUND}
	}
	return nil
}
