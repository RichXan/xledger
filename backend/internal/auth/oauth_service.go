package auth

import (
	"context"
	"errors"
	"strings"
	"time"
)

const AUTH_OAUTH_FAILED = "AUTH_OAUTH_FAILED"

type GoogleCallbackInput struct {
	State string
	Nonce string
	Email string
}

type OAuthService struct {
	repo     CodeRepository
	sessions *SessionService
	now      func() time.Time
	ttl      time.Duration
}

func NewOAuthService(repo CodeRepository, sessions *SessionService, now func() time.Time) *OAuthService {
	if now == nil {
		now = time.Now
	}
	return &OAuthService{
		repo:     repo,
		sessions: sessions,
		now:      now,
		ttl:      10 * time.Minute,
	}
}

func (s *OAuthService) SeedStateNonce(ctx context.Context, state string, nonce string) error {
	return s.repo.SaveOAuthStateNonce(ctx, strings.TrimSpace(state), strings.TrimSpace(nonce), s.ttl)
}

func (s *OAuthService) GoogleCallback(ctx context.Context, input GoogleCallbackInput) (TokenPair, error) {
	state := strings.TrimSpace(input.State)
	nonce := strings.TrimSpace(input.Nonce)
	email := strings.TrimSpace(strings.ToLower(input.Email))

	if state == "" || nonce == "" || email == "" {
		return TokenPair{}, &authError{code: AUTH_OAUTH_FAILED, err: errors.New("invalid oauth callback input")}
	}

	valid, err := s.repo.ConsumeOAuthStateNonce(ctx, state, nonce)
	if err != nil {
		return TokenPair{}, err
	}
	if !valid {
		return TokenPair{}, &authError{code: AUTH_OAUTH_FAILED, err: errors.New("invalid state or nonce")}
	}

	if s.sessions == nil {
		return TokenPair{}, errors.New("session service not configured")
	}
	return s.sessions.IssueSession(ctx, email)
}
