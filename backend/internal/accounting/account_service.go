package accounting

import (
	"context"
	"strings"
)

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(_ context.Context, userID string, input AccountCreateInput) (Account, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Type) == "" {
		return Account{}, &contractError{code: ACCOUNT_INVALID}
	}
	return s.repo.Create(userID, input)
}

func (s *AccountService) GetAccount(_ context.Context, userID string, accountID string) (Account, error) {
	account, found, err := s.repo.GetByIDForUser(userID, accountID)
	if err != nil {
		return Account{}, err
	}
	if !found {
		return Account{}, &contractError{code: ACCOUNT_NOT_FOUND}
	}
	return account, nil
}

func (s *AccountService) UpdateAccount(_ context.Context, userID string, accountID string, input AccountUpdateInput) (Account, error) {
	if input.HasName && strings.TrimSpace(input.Name) == "" {
		return Account{}, &contractError{code: ACCOUNT_INVALID}
	}
	if input.HasType && strings.TrimSpace(input.Type) == "" {
		return Account{}, &contractError{code: ACCOUNT_INVALID}
	}

	account, found, err := s.repo.UpdateByIDForUser(userID, accountID, input)
	if err != nil {
		return Account{}, err
	}
	if !found {
		return Account{}, &contractError{code: ACCOUNT_NOT_FOUND}
	}
	return account, nil
}

func (s *AccountService) DeleteAccount(_ context.Context, userID string, accountID string) error {
	deleted, err := s.repo.DeleteByIDForUser(userID, accountID)
	if err != nil {
		return err
	}
	if !deleted {
		return &contractError{code: ACCOUNT_NOT_FOUND}
	}
	return nil
}
