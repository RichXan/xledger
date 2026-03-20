package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	AUTH_BAD_REQUEST     = "AUTH_BAD_REQUEST"
	AUTH_UNAUTHORIZED    = "AUTH_UNAUTHORIZED"
	AUTH_REFRESH_EXPIRED = "AUTH_REFRESH_EXPIRED"
	AUTH_REFRESH_REVOKED = "AUTH_REFRESH_REVOKED"
)

type SessionServiceOptions struct {
	BlacklistStrictMode bool
	BlacklistSLA        time.Duration
}

type SessionToken struct {
	Type      string
	Email     string
	ExpiresAt time.Time
	ID        string
}

type SessionService struct {
	repo                CodeRepository
	now                 func() time.Time
	refreshTTL          time.Duration
	accessTTL           time.Duration
	blacklistStrictMode bool
	blacklistSLA        time.Duration
}

var globalSessionTokenCounter int64

func NewSessionService(repo CodeRepository, opts *SessionServiceOptions, now func() time.Time) *SessionService {
	if now == nil {
		now = time.Now
	}
	service := &SessionService{
		repo:       repo,
		now:        now,
		refreshTTL: 7 * 24 * time.Hour,
		accessTTL:  15 * time.Minute,
	}
	if opts != nil {
		service.blacklistStrictMode = opts.BlacklistStrictMode
		service.blacklistSLA = opts.BlacklistSLA
	}
	if service.blacklistSLA <= 0 {
		service.blacklistSLA = 5 * time.Second
	}
	return service
}

func (s *SessionService) IssueSession(ctx context.Context, email string) (TokenPair, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if normalizedEmail == "" {
		return TokenPair{}, &authError{code: AUTH_BAD_REQUEST, err: errors.New("email is required")}
	}

	now := s.now()
	refreshToken := s.nextToken("refresh", normalizedEmail, now.Add(s.refreshTTL))
	parsedRefresh, err := ParseSessionToken(refreshToken)
	if err != nil {
		return TokenPair{}, fmt.Errorf("parse generated refresh token: %w", err)
	}
	if err := s.repo.StoreRefreshToken(ctx, parsedRefresh.ID, parsedRefresh.Email, parsedRefresh.ExpiresAt); err != nil {
		return TokenPair{}, fmt.Errorf("store refresh token: %w", err)
	}
	if err := s.repo.CreateSession(ctx, normalizedEmail); err != nil {
		return TokenPair{}, fmt.Errorf("create session: %w", err)
	}
	if _, err := s.repo.EnsureDefaultLedger(ctx, normalizedEmail); err != nil {
		return TokenPair{}, fmt.Errorf("bootstrap default ledger: %w", err)
	}

	return TokenPair{
		AccessToken:  s.nextToken("access", normalizedEmail, now.Add(s.accessTTL)),
		RefreshToken: refreshToken,
	}, nil
}

func (s *SessionService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	token, err := ParseSessionToken(refreshToken)
	if err != nil {
		return TokenPair{}, &authError{code: AUTH_BAD_REQUEST, err: err}
	}
	if token.Type != "refresh" {
		return TokenPair{}, &authError{code: AUTH_BAD_REQUEST, err: errors.New("refresh token required")}
	}
	if !s.now().Before(token.ExpiresAt) {
		return TokenPair{}, &authError{code: AUTH_REFRESH_EXPIRED, err: errors.New("refresh token expired")}
	}

	blacklisted, lag, err := s.repo.IsRefreshTokenBlacklisted(ctx, token.ID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("check refresh blacklist: %w", err)
	}
	if blacklisted {
		return TokenPair{}, &authError{code: AUTH_REFRESH_REVOKED, err: errors.New("refresh token revoked")}
	}
	if s.blacklistStrictMode && lag > s.blacklistSLA {
		_ = s.repo.RecordAlertEvent(ctx, "auth.refresh.blacklist_sla_exceeded")
		return TokenPair{}, &authError{code: AUTH_REFRESH_REVOKED, err: errors.New("blacklist lag exceeded")}
	}

	stored, ok, err := s.repo.ConsumeRefreshToken(ctx, token.ID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("consume refresh token: %w", err)
	}
	if !ok {
		return TokenPair{}, &authError{code: AUTH_REFRESH_REVOKED, err: errors.New("refresh token already used")}
	}
	if !s.now().Before(stored.ExpiresAt) {
		return TokenPair{}, &authError{code: AUTH_REFRESH_EXPIRED, err: errors.New("refresh token expired")}
	}

	now := s.now()
	newRefresh := s.nextToken("refresh", stored.Email, now.Add(s.refreshTTL))
	parsedNewRefresh, parseErr := ParseSessionToken(newRefresh)
	if parseErr != nil {
		return TokenPair{}, fmt.Errorf("parse rotated refresh token: %w", parseErr)
	}
	if storeErr := s.repo.StoreRefreshToken(ctx, parsedNewRefresh.ID, parsedNewRefresh.Email, parsedNewRefresh.ExpiresAt); storeErr != nil {
		return TokenPair{}, fmt.Errorf("store rotated refresh token: %w", storeErr)
	}

	return TokenPair{
		AccessToken:  s.nextToken("access", stored.Email, now.Add(s.accessTTL)),
		RefreshToken: newRefresh,
	}, nil
}

func (s *SessionService) Logout(ctx context.Context, refreshToken string) error {
	trimmed := strings.TrimSpace(refreshToken)
	if trimmed == "" {
		return &authError{code: AUTH_UNAUTHORIZED, err: errors.New("missing credentials")}
	}

	token, err := ParseSessionToken(trimmed)
	if err != nil {
		return &authError{code: AUTH_BAD_REQUEST, err: err}
	}
	if token.Type != "refresh" {
		return &authError{code: AUTH_BAD_REQUEST, err: errors.New("refresh token required")}
	}

	if blacklistErr := s.repo.BlacklistRefreshToken(ctx, token.ID, s.now(), token.ExpiresAt); blacklistErr != nil {
		return fmt.Errorf("blacklist refresh token: %w", blacklistErr)
	}
	return nil
}

func ParseSessionToken(raw string) (SessionToken, error) {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 4 {
		return SessionToken{}, errors.New("invalid token format")
	}
	expiresUnix, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return SessionToken{}, errors.New("invalid token expiry")
	}
	if strings.TrimSpace(parts[1]) == "" || strings.TrimSpace(parts[3]) == "" {
		return SessionToken{}, errors.New("invalid token payload")
	}

	return SessionToken{
		Type:      parts[0],
		Email:     parts[1],
		ExpiresAt: time.Unix(expiresUnix, 0).UTC(),
		ID:        parts[3],
	}, nil
}

func (s *SessionService) nextToken(tokenType string, email string, expiresAt time.Time) string {
	next := atomic.AddInt64(&globalSessionTokenCounter, 1)
	id := strconv.FormatInt(s.now().UnixNano(), 10) + "-" + strconv.FormatInt(next, 10)
	return tokenType + ":" + email + ":" + strconv.FormatInt(expiresAt.UTC().Unix(), 10) + ":" + id
}
