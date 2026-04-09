package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestGoogleCallback_ValidatesStateNonce(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	if err := svc.SeedStateNonceForEmail(context.Background(), "state-a", "nonce-a", "oauth@example.com"); err != nil {
		t.Fatalf("seed oauth state nonce: %v", err)
	}

	_, err := svc.GoogleCallback(context.Background(), GoogleCallbackInput{State: "wrong-state", Nonce: "nonce-a", Email: "oauth@example.com"})
	if ErrorCode(err) != AUTH_OAUTH_FAILED {
		t.Fatalf("expected %s for invalid state, got %q", AUTH_OAUTH_FAILED, ErrorCode(err))
	}

	_, err = svc.GoogleCallback(context.Background(), GoogleCallbackInput{State: "state-a", Nonce: "nonce-a", Email: "oauth@example.com"})
	if err != nil {
		t.Fatalf("expected callback with valid state+nonce to succeed, got %v", err)
	}
}

func TestGoogleCallback_ReplayedNonceRejected(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 5, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	if err := svc.SeedStateNonceForEmail(context.Background(), "state-replay", "nonce-replay", "oauth@example.com"); err != nil {
		t.Fatalf("seed oauth state nonce: %v", err)
	}

	if _, err := svc.GoogleCallback(context.Background(), GoogleCallbackInput{State: "state-replay", Nonce: "nonce-replay", Email: "oauth@example.com"}); err != nil {
		t.Fatalf("first callback should succeed, got %v", err)
	}

	_, err := svc.GoogleCallback(context.Background(), GoogleCallbackInput{State: "state-replay", Nonce: "nonce-replay", Email: "oauth@example.com"})
	if ErrorCode(err) != AUTH_OAUTH_FAILED {
		t.Fatalf("expected replayed nonce to fail with %s, got %q", AUTH_OAUTH_FAILED, ErrorCode(err))
	}
}

func TestGoogleCallback_HTTPContract_IsGET_api_auth_google_callback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 3, 20, 12, 10, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	oauthSvc := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	if err := oauthSvc.SeedStateNonceForEmail(context.Background(), "state-http", "nonce-http", "oauth@example.com"); err != nil {
		t.Fatalf("seed oauth state nonce: %v", err)
	}

	h := NewHandlerWithServices(
		NewCodeService(repo, &oauthSessionSender{}, nil, func() time.Time { return now }, func() string { return "123456" }),
		oauthSvc,
		NewSessionService(repo, nil, func() time.Time { return now }),
	)
	r := gin.New()
	r.GET("/api/auth/google/callback", h.GoogleCallback)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?state=state-http&nonce=nonce-http&email=oauth@example.com", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected GET callback status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/google/callback", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected POST callback route mismatch status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGoogleCallback_InvalidProviderResponse_ReturnsAUTH_OAUTH_FAILED(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 15, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveOAuthStateNonce(context.Background(), "state-invalid", "nonce-invalid", 10*time.Minute); err != nil {
		t.Fatalf("seed oauth state nonce: %v", err)
	}

	svc := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	_, err := svc.GoogleCallback(context.Background(), GoogleCallbackInput{State: "state-invalid", Nonce: "nonce-invalid"})
	if ErrorCode(err) != AUTH_OAUTH_FAILED {
		t.Fatalf("expected %s for invalid provider response, got %q", AUTH_OAUTH_FAILED, ErrorCode(err))
	}
}

func TestGoogleCallback_DoesNotTrustQueryEmail(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 16, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	if err := svc.SeedStateNonceForEmail(context.Background(), "state-bound", "nonce-bound", "bound@example.com"); err != nil {
		t.Fatalf("seed oauth state nonce: %v", err)
	}

	tokens, err := svc.GoogleCallback(context.Background(), GoogleCallbackInput{State: "state-bound", Nonce: "nonce-bound", Email: "attacker@example.com"})
	if err != nil {
		t.Fatalf("expected callback to succeed for bound identity, got %v", err)
	}
	parsed, err := ParseSessionToken(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	if parsed.Email != "bound@example.com" {
		t.Fatalf("expected callback identity to come from bound state, got %q", parsed.Email)
	}
}

func TestGoogleCallback_RequiresBoundIdentity(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 17, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveOAuthStateNonce(context.Background(), "state-no-email", "nonce-no-email", 10*time.Minute); err != nil {
		t.Fatalf("seed oauth state nonce: %v", err)
	}
	svc := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })

	_, err := svc.GoogleCallback(context.Background(), GoogleCallbackInput{State: "state-no-email", Nonce: "nonce-no-email", Email: "query@example.com"})
	if ErrorCode(err) != AUTH_OAUTH_FAILED {
		t.Fatalf("expected %s when no server-bound identity exists, got %q", AUTH_OAUTH_FAILED, ErrorCode(err))
	}
}

func TestVerifyCode_IssuedRefreshTokenCompatibleWithSessionFlow(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 18, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveVerificationCode(context.Background(), "verify-session@example.com", "111111", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}

	codeSvc := NewCodeService(repo, &oauthSessionSender{}, nil, func() time.Time { return now }, nil)
	pair, err := codeSvc.VerifyCode(context.Background(), "verify-session@example.com", "111111")
	if err != nil {
		t.Fatalf("verify code should succeed, got %v", err)
	}

	sessionSvc := NewSessionService(repo, nil, func() time.Time { return now })
	if _, err := sessionSvc.Refresh(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("expected verify-issued refresh token to work with refresh flow, got %v", err)
	}
}

func TestRefresh_RotationInvalidatesOldToken(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 20, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "rotate@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	rotated, err := svc.Refresh(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh should succeed, got %v", err)
	}
	if rotated.RefreshToken == pair.RefreshToken {
		t.Fatalf("expected rotation to issue new refresh token")
	}

	_, err = svc.Refresh(context.Background(), pair.RefreshToken)
	if ErrorCode(err) != AUTH_REFRESH_REVOKED {
		t.Fatalf("expected reused refresh token to fail with %s, got %q", AUTH_REFRESH_REVOKED, ErrorCode(err))
	}
}

func TestRefresh_ExpiryIsSevenDays(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 25, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "expiry@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	parsed, err := ParseSessionToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	if parsed.ExpiresAt.Sub(now) != 7*24*time.Hour {
		t.Fatalf("expected refresh expiry to be 7 days, got %s", parsed.ExpiresAt.Sub(now))
	}
}

func TestIssueSession_RefreshTokenIDIsUUID(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 27, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "uuid-refresh@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	parsed, err := ParseSessionToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	if _, err := uuid.Parse(parsed.ID); err != nil {
		t.Fatalf("expected refresh token id to be uuid, got %q", parsed.ID)
	}
}

func TestRefresh_Expired_ReturnsAUTH_REFRESH_EXPIRED(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 30, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "expired-refresh@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	now = now.Add(7*24*time.Hour + time.Second)
	_, err = svc.Refresh(context.Background(), pair.RefreshToken)
	if ErrorCode(err) != AUTH_REFRESH_EXPIRED {
		t.Fatalf("expected %s, got %q", AUTH_REFRESH_EXPIRED, ErrorCode(err))
	}
}

func TestRefresh_RejectsAccessTokenType(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 35, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "access-type@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	_, err = svc.Refresh(context.Background(), pair.AccessToken)
	if ErrorCode(err) != AUTH_BAD_REQUEST {
		t.Fatalf("expected %s for non-refresh token, got %q", AUTH_BAD_REQUEST, ErrorCode(err))
	}
}

func TestLogout_BlacklistsAccessTokenImmediately(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 41, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "logout-access@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	if err := svc.Logout(context.Background(), pair.RefreshToken, pair.AccessToken); err != nil {
		t.Fatalf("logout should succeed, got %v", err)
	}

	parsedAccess, err := ParseSessionToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	blacklisted, err := repo.IsAccessTokenBlacklisted(context.Background(), parsedAccess.ID)
	if err != nil {
		t.Fatalf("query access blacklist: %v", err)
	}
	if !blacklisted {
		t.Fatal("expected access token to be blacklisted on logout")
	}
}

func TestRequireAccessToken_RejectsBlacklistedAccessToken(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 42, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "middleware-access@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	parsedAccess, err := ParseSessionToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if err := repo.BlacklistAccessToken(context.Background(), parsedAccess.ID, now, parsedAccess.ExpiresAt); err != nil {
		t.Fatalf("blacklist access token: %v", err)
	}

	if _, err := svc.ValidateAccessToken(context.Background(), pair.AccessToken); ErrorCode(err) != AUTH_UNAUTHORIZED {
		t.Fatalf("expected %s for blacklisted access token, got %q", AUTH_UNAUTHORIZED, ErrorCode(err))
	}
}

func TestLogout_BlacklistPropagationLTE5s(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 45, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "propagation@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	if err := svc.Logout(context.Background(), pair.RefreshToken, pair.AccessToken); err != nil {
		t.Fatalf("logout should succeed, got %v", err)
	}

	parsed, err := ParseSessionToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	_, lag, err := repo.IsRefreshTokenBlacklisted(context.Background(), parsed.ID)
	if err != nil {
		t.Fatalf("query blacklist state: %v", err)
	}
	if lag > 5*time.Second {
		t.Fatalf("expected blacklist propagation lag <=5s, got %s", lag)
	}
}

func TestLogout_BlacklistEntryExpiresAfterTokenLifetime(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 46, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "blacklist-expiry@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	if err := svc.Logout(context.Background(), pair.RefreshToken, pair.AccessToken); err != nil {
		t.Fatalf("logout should succeed, got %v", err)
	}

	if repo.BlacklistCount() != 1 {
		t.Fatalf("expected one blacklist entry after logout, got %d", repo.BlacklistCount())
	}

	now = now.Add(8 * 24 * time.Hour)
	if err := repo.SaveVerificationCode(context.Background(), "cleanup-trigger@example.com", "999999", 10*time.Minute); err != nil {
		t.Fatalf("trigger cleanup: %v", err)
	}

	if repo.BlacklistCount() != 0 {
		t.Fatalf("expected expired blacklist entry to be cleaned up, got %d", repo.BlacklistCount())
	}
}

func TestInMemoryRepository_BlacklistCleanupPreventsUnboundedGrowth(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 47, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	for i := 0; i < 20; i++ {
		email := "blacklist-growth-" + strconv.Itoa(i) + "@example.com"
		pair, err := svc.IssueSession(context.Background(), email)
		if err != nil {
			t.Fatalf("issue session %d: %v", i, err)
		}
		if err := svc.Logout(context.Background(), pair.RefreshToken, pair.AccessToken); err != nil {
			t.Fatalf("logout %d: %v", i, err)
		}
	}

	if repo.BlacklistCount() != 20 {
		t.Fatalf("expected 20 blacklist entries, got %d", repo.BlacklistCount())
	}

	now = now.Add(9 * 24 * time.Hour)
	if _, err := svc.IssueSession(context.Background(), "fresh-after-expiry@example.com"); err != nil {
		t.Fatalf("issue fresh session: %v", err)
	}

	if repo.BlacklistCount() != 0 {
		t.Fatalf("expected expired blacklist entries to be evicted, got %d", repo.BlacklistCount())
	}
}

func TestLogout_RejectsRefreshTokenType(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 50, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "logout-type@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	err = svc.Logout(context.Background(), pair.AccessToken, "")
	if ErrorCode(err) != AUTH_BAD_REQUEST {
		t.Fatalf("expected %s for non-refresh logout token, got %q", AUTH_BAD_REQUEST, ErrorCode(err))
	}
}

func TestLogout_Unauthorized_ReturnsAUTH_UNAUTHORIZED(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 55, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewSessionService(repo, nil, func() time.Time { return now })

	err := svc.Logout(context.Background(), "", "")
	if ErrorCode(err) != AUTH_UNAUTHORIZED {
		t.Fatalf("expected %s, got %q", AUTH_UNAUTHORIZED, ErrorCode(err))
	}
}

func TestRefresh_BlacklistStrictMode_ReturnsAUTH_REFRESH_REVOKED(t *testing.T) {
	now := time.Date(2026, 3, 20, 13, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	repo.SetForcedBlacklistLag(6 * time.Second)
	svc := NewSessionService(repo, &SessionServiceOptions{BlacklistStrictMode: true}, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "strict@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	_, err = svc.Refresh(context.Background(), pair.RefreshToken)
	if ErrorCode(err) != AUTH_REFRESH_REVOKED {
		t.Fatalf("expected %s in strict mode when lag exceeds SLA, got %q", AUTH_REFRESH_REVOKED, ErrorCode(err))
	}
}

func TestRefresh_BlacklistStrictMode_EmitsAlertEvent(t *testing.T) {
	now := time.Date(2026, 3, 20, 13, 5, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	repo.SetForcedBlacklistLag(6 * time.Second)
	svc := NewSessionService(repo, &SessionServiceOptions{BlacklistStrictMode: true}, func() time.Time { return now })

	pair, err := svc.IssueSession(context.Background(), "strict-alert@example.com")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	_, _ = svc.Refresh(context.Background(), pair.RefreshToken)
	if repo.AlertEventCount("auth.refresh.blacklist_sla_exceeded") != 1 {
		t.Fatalf("expected strict-mode blacklist lag alert to be recorded")
	}
}

func TestFirstLogin_CreatesDefaultLedger(t *testing.T) {
	now := time.Date(2026, 3, 20, 13, 10, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	oauth := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	if err := oauth.SeedStateNonceForEmail(context.Background(), "state-ledger", "nonce-ledger", "first-login@example.com"); err != nil {
		t.Fatalf("seed oauth state nonce: %v", err)
	}

	if _, err := oauth.GoogleCallback(context.Background(), GoogleCallbackInput{State: "state-ledger", Nonce: "nonce-ledger", Email: "first-login@example.com"}); err != nil {
		t.Fatalf("oauth callback should succeed, got %v", err)
	}
	if repo.DefaultLedgerCount("first-login@example.com") != 1 {
		t.Fatalf("expected default ledger to be created once on first login")
	}

	if err := oauth.SeedStateNonceForEmail(context.Background(), "state-ledger-2", "nonce-ledger-2", "first-login@example.com"); err != nil {
		t.Fatalf("seed oauth state nonce: %v", err)
	}
	if _, err := oauth.GoogleCallback(context.Background(), GoogleCallbackInput{State: "state-ledger-2", Nonce: "nonce-ledger-2", Email: "first-login@example.com"}); err != nil {
		t.Fatalf("second oauth callback should succeed, got %v", err)
	}
	if repo.DefaultLedgerCount("first-login@example.com") != 1 {
		t.Fatalf("expected default ledger bootstrap to be idempotent")
	}
}

func TestOAuthFailure_DoesNotAffectSendCodeFlow(t *testing.T) {
	now := time.Date(2026, 3, 20, 13, 15, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	oauth := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	codeSvc := NewCodeService(repo, &oauthSessionSender{}, nil, func() time.Time { return now }, func() string { return "112233" })

	_, err := oauth.GoogleCallback(context.Background(), GoogleCallbackInput{State: "missing", Nonce: "missing"})
	if ErrorCode(err) != AUTH_OAUTH_FAILED {
		t.Fatalf("expected oauth callback failure, got %q", ErrorCode(err))
	}

	err = codeSvc.SendCode(context.Background(), "still-works@example.com", "203.0.113.10")
	if err != nil {
		t.Fatalf("expected send-code flow to remain healthy after oauth failure, got %v", err)
	}
}

type failingGoogleOAuthProvider struct {
	enabled bool
	err     error
}

func (p failingGoogleOAuthProvider) IsEnabled() bool {
	return p.enabled
}

func (p failingGoogleOAuthProvider) BuildAuthURL(state string) string {
	return "https://accounts.google.com/o/oauth2/v2/auth?state=" + state
}

func (p failingGoogleOAuthProvider) ResolveProfile(ctx context.Context, code string) (GoogleOAuthProfile, error) {
	return GoogleOAuthProfile{}, p.err
}

func TestGoogleOAuthErrorReason_MapsCommonFailures(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{name: "not configured", err: &authError{code: AUTH_OAUTH_FAILED, err: errString("google oauth is not configured")}, expected: GoogleOAuthReasonNotConfigured},
		{name: "invalid callback", err: &authError{code: AUTH_OAUTH_FAILED, err: errString("invalid oauth callback input")}, expected: GoogleOAuthReasonInvalidCallback},
		{name: "invalid state", err: &authError{code: AUTH_OAUTH_FAILED, err: errString("invalid or expired oauth state")}, expected: GoogleOAuthReasonInvalidState},
		{name: "invalid code", err: &authError{code: AUTH_OAUTH_FAILED, err: ErrGoogleOAuthCodeInvalidOrExpired}, expected: GoogleOAuthReasonCodeInvalidOrExpired},
		{name: "token exchange", err: &authError{code: AUTH_OAUTH_FAILED, err: errString("google token exchange failed: redirect_uri_mismatch")}, expected: GoogleOAuthReasonTokenExchangeFailed},
		{name: "profile fetch", err: &authError{code: AUTH_OAUTH_FAILED, err: errString("google userinfo failed: 401 Unauthorized")}, expected: GoogleOAuthReasonProfileFetchFailed},
		{name: "email missing", err: &authError{code: AUTH_OAUTH_FAILED, err: errString("google account email is empty")}, expected: GoogleOAuthReasonEmailMissing},
	}

	for _, tc := range tests {
		if got := GoogleOAuthErrorReason(tc.err); got != tc.expected {
			t.Fatalf("%s: expected reason %q, got %q", tc.name, tc.expected, got)
		}
	}
}

func TestGoogleCallbackByCode_InvalidGrant_MapsReason(t *testing.T) {
	now := time.Date(2026, 3, 20, 13, 10, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	svc.SetGoogleProvider(failingGoogleOAuthProvider{enabled: true, err: ErrGoogleOAuthCodeInvalidOrExpired})
	if err := repo.SaveOAuthStateNonce(context.Background(), "state-code", "state-code", 10*time.Minute); err != nil {
		t.Fatalf("seed oauth state nonce: %v", err)
	}

	_, err := svc.GoogleCallbackByCode(context.Background(), "state-code", "google-code")
	if ErrorCode(err) != AUTH_OAUTH_FAILED {
		t.Fatalf("expected %s, got %q", AUTH_OAUTH_FAILED, ErrorCode(err))
	}
	if got := GoogleOAuthErrorReason(err); got != GoogleOAuthReasonCodeInvalidOrExpired {
		t.Fatalf("expected reason %q, got %q", GoogleOAuthReasonCodeInvalidOrExpired, got)
	}
}

func TestGoogleCallbackByCode_RedirectsWithDetailedReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 3, 20, 13, 15, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	oauthSvc := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	oauthSvc.SetGoogleProvider(failingGoogleOAuthProvider{enabled: true, err: errString("google token exchange failed: redirect_uri_mismatch")})
	if err := repo.SaveOAuthStateNonce(context.Background(), "state-redirect", "state-redirect", 10*time.Minute); err != nil {
		t.Fatalf("seed oauth state nonce: %v", err)
	}

	h := NewHandlerWithServices(NewCodeService(repo, &oauthSessionSender{}, nil, func() time.Time { return now }, func() string { return "123456" }), oauthSvc, NewSessionService(repo, nil, func() time.Time { return now }))
	h.SetGoogleFrontendReturnURL("http://127.0.0.1:4173/auth/google/callback")
	r := gin.New()
	r.GET("/api/auth/google/callback", h.GoogleCallback)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?state=state-redirect&code=bad-code", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected %d, got %d body=%s", http.StatusTemporaryRedirect, rec.Code, rec.Body.String())
	}
	location := rec.Header().Get("Location")
	if !strings.Contains(location, "error_code=AUTH_OAUTH_FAILED") {
		t.Fatalf("expected AUTH_OAUTH_FAILED in redirect, got %s", location)
	}
	if !strings.Contains(location, "error_reason="+GoogleOAuthReasonTokenExchangeFailed) {
		t.Fatalf("expected detailed error reason in redirect, got %s", location)
	}
}

type oauthSessionSender struct{}

func (oauthSessionSender) Send(to string, subject string, body string) error {
	return nil
}

type errString string

func (e errString) Error() string { return string(e) }
