package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	google   GoogleOAuthProvider
}

func NewOAuthService(repo CodeRepository, sessions *SessionService, now func() time.Time) *OAuthService {
	if now == nil {
		now = time.Now
	}
	return &OAuthService{
		repo:     repo,
		sessions: sessions,
		now:      now,
		ttl:      30 * time.Minute,
	}
}

func (s *OAuthService) SetGoogleProvider(provider GoogleOAuthProvider) {
	s.google = provider
}

func (s *OAuthService) GoogleLoginURL(ctx context.Context) (string, error) {
	if s.google == nil || !s.google.IsEnabled() {
		return "", &authError{code: AUTH_OAUTH_FAILED, err: errors.New("google oauth is not configured")}
	}
	state, err := randomHex(16)
	if err != nil {
		return "", err
	}
	if err := s.repo.SaveOAuthStateNonce(ctx, state, state, s.ttl); err != nil {
		return "", err
	}
	return s.google.BuildAuthURL(state), nil
}

func (s *OAuthService) GoogleCallbackByCode(ctx context.Context, state string, code string) (TokenPair, error) {
	if s.google == nil || !s.google.IsEnabled() {
		return TokenPair{}, &authError{code: AUTH_OAUTH_FAILED, err: errors.New("google oauth is not configured")}
	}

	state = strings.TrimSpace(state)
	code = strings.TrimSpace(code)
	if state == "" || code == "" {
		return TokenPair{}, &authError{code: AUTH_OAUTH_FAILED, err: errors.New("invalid oauth callback input")}
	}

	ok, err := s.repo.CheckOAuthStateNonce(ctx, state, state)
	if err != nil {
		return TokenPair{}, err
	}
	if !ok {
		return TokenPair{}, &authError{code: AUTH_OAUTH_FAILED, err: errors.New("invalid or expired oauth state")}
	}

	profileCtx, cancelProfile := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelProfile()
	profile, err := s.google.ResolveProfile(profileCtx, code)
	if err != nil {
		return TokenPair{}, &authError{code: AUTH_OAUTH_FAILED, err: err}
	}

	consumeCtx, cancelConsume := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelConsume()
	ok, err = s.repo.ConsumeOAuthStateNonce(consumeCtx, state, state)
	if err != nil {
		return TokenPair{}, err
	}
	if !ok {
		return TokenPair{}, &authError{code: AUTH_OAUTH_FAILED, err: errors.New("invalid or expired oauth state")}
	}

	email := strings.TrimSpace(strings.ToLower(profile.Email))
	if email == "" {
		return TokenPair{}, &authError{code: AUTH_OAUTH_FAILED, err: errors.New("google account email is empty")}
	}
	if s.sessions == nil {
		return TokenPair{}, errors.New("session service not configured")
	}
	issueCtx, cancelIssue := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelIssue()
	return s.sessions.IssueSessionWithProfile(issueCtx, email, profile.Name)
}

func (s *OAuthService) SeedStateNonce(ctx context.Context, state string, nonce string) error {
	return s.repo.SaveOAuthStateNonce(ctx, strings.TrimSpace(state), strings.TrimSpace(nonce), s.ttl)
}

func (s *OAuthService) SeedStateNonceForEmail(ctx context.Context, state string, nonce string, email string) error {
	return s.repo.SaveOAuthStateNonceForEmail(
		ctx,
		strings.TrimSpace(state),
		strings.TrimSpace(nonce),
		strings.TrimSpace(strings.ToLower(email)),
		s.ttl,
	)
}

func (s *OAuthService) GoogleCallback(ctx context.Context, input GoogleCallbackInput) (TokenPair, error) {
	state := strings.TrimSpace(input.State)
	nonce := strings.TrimSpace(input.Nonce)

	if state == "" || nonce == "" {
		return TokenPair{}, &authError{code: AUTH_OAUTH_FAILED, err: errors.New("invalid oauth callback input")}
	}

	email, valid, err := s.repo.ConsumeOAuthStateNonceForEmail(ctx, state, nonce)
	if err != nil {
		return TokenPair{}, err
	}
	if !valid || strings.TrimSpace(email) == "" {
		return TokenPair{}, &authError{code: AUTH_OAUTH_FAILED, err: errors.New("invalid state or nonce")}
	}

	if s.sessions == nil {
		return TokenPair{}, errors.New("session service not configured")
	}
	return s.sessions.IssueSession(ctx, strings.TrimSpace(strings.ToLower(email)))
}

func randomHex(nBytes int) (string, error) {
	if nBytes <= 0 {
		nBytes = 16
	}
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *OAuthService) StoreExchangeCode(code string, tokens TokenPair) error {
	return s.repo.StoreOAuthExchangeCode(context.Background(), code, tokens, s.now().Add(10*time.Minute))
}

func (s *OAuthService) ConsumeExchangeCode(code string) (TokenPair, bool) {
	tokens, ok, err := s.repo.ConsumeOAuthExchangeCode(context.Background(), code)
	if err != nil {
		return TokenPair{}, false
	}
	return tokens, ok
}
