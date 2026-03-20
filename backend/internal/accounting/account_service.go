package accounting

import (
	"context"
	"strings"
	"time"
)

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(_ context.Context, userID string, input AccountCreateInput) (Account, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedName := strings.TrimSpace(input.Name)
	normalizedType := strings.TrimSpace(input.Type)
	if normalizedUserID == "" || normalizedName == "" || normalizedType == "" {
		return Account{}, &contractError{code: ACCOUNT_INVALID}
	}
	input.Name = normalizedName
	input.Type = normalizedType
	return s.repo.Create(normalizedUserID, input)
}

func (s *AccountService) ListAccounts(_ context.Context, userID string) ([]Account, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, &contractError{code: ACCOUNT_INVALID}
	}
	return s.repo.ListByUser(normalizedUserID)
}

func (s *AccountService) GetAccount(_ context.Context, userID string, accountID string) (Account, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedAccountID := strings.TrimSpace(accountID)
	if normalizedUserID == "" || normalizedAccountID == "" {
		return Account{}, &contractError{code: ACCOUNT_INVALID}
	}

	account, found, err := s.repo.GetByIDForUser(normalizedUserID, normalizedAccountID)
	if err != nil {
		return Account{}, err
	}
	if !found {
		return Account{}, &contractError{code: ACCOUNT_NOT_FOUND}
	}
	return account, nil
}

func (s *AccountService) UpdateAccount(_ context.Context, userID string, accountID string, input AccountUpdateInput) (Account, error) {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedAccountID := strings.TrimSpace(accountID)
	if normalizedUserID == "" || normalizedAccountID == "" {
		return Account{}, &contractError{code: ACCOUNT_INVALID}
	}

	if input.HasName && strings.TrimSpace(input.Name) == "" {
		return Account{}, &contractError{code: ACCOUNT_INVALID}
	}
	if input.HasType && strings.TrimSpace(input.Type) == "" {
		return Account{}, &contractError{code: ACCOUNT_INVALID}
	}
	if input.HasName {
		input.Name = strings.TrimSpace(input.Name)
	}
	if input.HasType {
		input.Type = strings.TrimSpace(input.Type)
	}

	account, found, err := s.repo.GetByIDForUser(normalizedUserID, normalizedAccountID)
	if err != nil {
		return Account{}, err
	}
	if !found {
		return Account{}, &contractError{code: ACCOUNT_NOT_FOUND}
	}

	if input.HasName {
		account.Name = input.Name
	}
	if input.HasType {
		account.Type = input.Type
	}
	if input.HasArchive {
		if input.Archive {
			now := time.Now().UTC()
			account.ArchivedAt = &now
		} else {
			account.ArchivedAt = nil
		}
	}

	updated, saved, saveErr := s.repo.SaveByIDForUser(normalizedUserID, normalizedAccountID, account)
	if saveErr != nil {
		return Account{}, saveErr
	}
	if !saved {
		return Account{}, &contractError{code: ACCOUNT_NOT_FOUND}
	}

	return updated, nil
}

func (s *AccountService) DeleteAccount(_ context.Context, userID string, accountID string) error {
	normalizedUserID := strings.TrimSpace(userID)
	normalizedAccountID := strings.TrimSpace(accountID)
	if normalizedUserID == "" || normalizedAccountID == "" {
		return &contractError{code: ACCOUNT_INVALID}
	}

	deleted, err := s.repo.DeleteByIDForUser(normalizedUserID, normalizedAccountID)
	if err != nil {
		return err
	}
	if !deleted {
		return &contractError{code: ACCOUNT_NOT_FOUND}
	}
	return nil
}
