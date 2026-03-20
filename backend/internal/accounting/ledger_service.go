package accounting

import (
	"context"
	"errors"
	"strings"
)

const (
	LEDGER_INVALID           = "LEDGER_INVALID"
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

func (s *LedgerService) CreateLedger(_ context.Context, userID string, input LedgerCreateInput) (Ledger, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedName := strings.TrimSpace(input.Name)
	if normalizedUserID == "" || normalizedName == "" {
		return Ledger{}, &contractError{code: LEDGER_INVALID}
	}
	if input.IsDefault {
		ledgers, err := s.repo.ListByUser(normalizedUserID)
		if err != nil {
			return Ledger{}, err
		}
		for _, ledger := range ledgers {
			if ledger.IsDefault {
				return Ledger{}, &contractError{code: LEDGER_INVALID, err: errors.New("default ledger already exists")}
			}
		}
	}
	input.Name = normalizedName
	return s.repo.Create(normalizedUserID, input)
}

func (s *LedgerService) ListLedgers(_ context.Context, userID string) ([]Ledger, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, &contractError{code: LEDGER_INVALID}
	}
	return s.repo.ListByUser(normalizedUserID)
}

func (s *LedgerService) UpdateLedger(_ context.Context, userID string, ledgerID string, input LedgerCreateInput) (Ledger, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedLedgerID := strings.TrimSpace(ledgerID)
	normalizedName := strings.TrimSpace(input.Name)
	if normalizedUserID == "" || normalizedLedgerID == "" || normalizedName == "" {
		return Ledger{}, &contractError{code: LEDGER_INVALID}
	}

	ledger, found, err := s.repo.GetByIDForUser(normalizedUserID, normalizedLedgerID)
	if err != nil {
		return Ledger{}, err
	}
	if !found {
		return Ledger{}, &contractError{code: LEDGER_NOT_FOUND}
	}
	ledger.Name = normalizedName

	updated, saved, saveErr := s.repo.SaveByIDForUser(normalizedUserID, normalizedLedgerID, ledger)
	if saveErr != nil {
		return Ledger{}, saveErr
	}
	if !saved {
		return Ledger{}, &contractError{code: LEDGER_NOT_FOUND}
	}
	return updated, nil
}

func (s *LedgerService) DeleteLedger(_ context.Context, userID string, ledgerID string) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedLedgerID := strings.TrimSpace(ledgerID)
	if normalizedUserID == "" || normalizedLedgerID == "" {
		return &contractError{code: LEDGER_INVALID}
	}

	ledger, found, err := s.repo.GetByIDForUser(normalizedUserID, normalizedLedgerID)
	if err != nil {
		return err
	}
	if !found {
		return &contractError{code: LEDGER_NOT_FOUND}
	}
	if ledger.IsDefault {
		return &contractError{code: LEDGER_DEFAULT_IMMUTABLE}
	}

	deleted, err := s.repo.DeleteByIDForUser(normalizedUserID, normalizedLedgerID)
	if err != nil {
		return err
	}
	if !deleted {
		return &contractError{code: LEDGER_NOT_FOUND}
	}
	return nil
}
