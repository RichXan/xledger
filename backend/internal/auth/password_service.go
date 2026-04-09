package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	AUTH_USER_EXISTS       = "AUTH_USER_EXISTS"
	AUTH_USER_NOT_FOUND    = "AUTH_USER_NOT_FOUND"
	AUTH_PASSWORD_INVALID  = "AUTH_PASSWORD_INVALID"
	AUTH_PASSWORD_WEAK     = "AUTH_PASSWORD_WEAK"
	AUTH_PROFILE_BAD_INPUT = "AUTH_PROFILE_BAD_INPUT"
	minPasswordLength      = 8
	bcryptCostDefault      = 12
)

type UserCredentialRecord struct {
	Email        string
	DisplayName  string
	PasswordHash string
}

type PasswordRepository interface {
	CreateUserWithPassword(ctx context.Context, email string, displayName string, passwordHash string) error
	GetUserCredential(ctx context.Context, email string) (UserCredentialRecord, bool, error)
	UpdateUserPassword(ctx context.Context, email string, passwordHash string) error
	UpdateUserDisplayName(ctx context.Context, email string, displayName string) error
}

type PasswordService struct {
	repo PasswordRepository
}

func NewPasswordService(repo PasswordRepository) *PasswordService {
	return &PasswordService{repo: repo}
}

func (s *PasswordService) Register(ctx context.Context, email string, password string, displayName string) (UserCredentialRecord, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	trimmedName := strings.TrimSpace(displayName)
	if normalizedEmail == "" {
		return UserCredentialRecord{}, &authError{code: AUTH_PROFILE_BAD_INPUT, err: errors.New("email is required")}
	}
	if len(strings.TrimSpace(password)) < minPasswordLength {
		return UserCredentialRecord{}, &authError{code: AUTH_PASSWORD_WEAK, err: errors.New("password too short")}
	}

	existing, ok, err := s.repo.GetUserCredential(ctx, normalizedEmail)
	if err != nil {
		return UserCredentialRecord{}, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCostDefault)
	if err != nil {
		return UserCredentialRecord{}, err
	}

	if ok {
		if strings.TrimSpace(existing.PasswordHash) != "" {
			return UserCredentialRecord{}, &authError{code: AUTH_USER_EXISTS, err: errors.New("user already exists")}
		}
		if err := s.repo.UpdateUserPassword(ctx, normalizedEmail, string(passwordHash)); err != nil {
			return UserCredentialRecord{}, err
		}
		if trimmedName != "" {
			if err := s.repo.UpdateUserDisplayName(ctx, normalizedEmail, trimmedName); err != nil {
				return UserCredentialRecord{}, err
			}
		}
		updated, _, lookupErr := s.repo.GetUserCredential(ctx, normalizedEmail)
		if lookupErr != nil {
			return UserCredentialRecord{}, lookupErr
		}
		return updated, nil
	}

	if err := s.repo.CreateUserWithPassword(ctx, normalizedEmail, trimmedName, string(passwordHash)); err != nil {
		return UserCredentialRecord{}, err
	}
	return UserCredentialRecord{
		Email:        normalizedEmail,
		DisplayName:  trimmedName,
		PasswordHash: string(passwordHash),
	}, nil
}

func (s *PasswordService) Login(ctx context.Context, email string, password string) (UserCredentialRecord, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	record, ok, err := s.repo.GetUserCredential(ctx, normalizedEmail)
	if err != nil {
		return UserCredentialRecord{}, err
	}
	if !ok {
		return UserCredentialRecord{}, &authError{code: AUTH_USER_NOT_FOUND, err: errors.New("user not found")}
	}
	if strings.TrimSpace(record.PasswordHash) == "" {
		return UserCredentialRecord{}, &authError{code: AUTH_PASSWORD_INVALID, err: errors.New("password login unavailable")}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(record.PasswordHash), []byte(password)); err != nil {
		return UserCredentialRecord{}, &authError{code: AUTH_PASSWORD_INVALID, err: err}
	}
	return record, nil
}

func (s *PasswordService) ChangePassword(ctx context.Context, email string, oldPassword string, newPassword string) error {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if len(strings.TrimSpace(newPassword)) < minPasswordLength {
		return &authError{code: AUTH_PASSWORD_WEAK, err: errors.New("password too short")}
	}

	record, ok, err := s.repo.GetUserCredential(ctx, normalizedEmail)
	if err != nil {
		return err
	}
	if !ok {
		return &authError{code: AUTH_USER_NOT_FOUND, err: errors.New("user not found")}
	}
	if strings.TrimSpace(record.PasswordHash) != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(record.PasswordHash), []byte(oldPassword)); err != nil {
			return &authError{code: AUTH_PASSWORD_INVALID, err: err}
		}
	}
	nextHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCostDefault)
	if err != nil {
		return err
	}
	return s.repo.UpdateUserPassword(ctx, normalizedEmail, string(nextHash))
}

func (s *PasswordService) UpdateDisplayName(ctx context.Context, email string, displayName string) error {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if normalizedEmail == "" {
		return &authError{code: AUTH_PROFILE_BAD_INPUT, err: errors.New("email is required")}
	}
	return s.repo.UpdateUserDisplayName(ctx, normalizedEmail, strings.TrimSpace(displayName))
}

func (s *PasswordService) Now() time.Time {
	return time.Now().UTC()
}
