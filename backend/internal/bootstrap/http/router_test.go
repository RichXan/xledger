package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"xledger/backend/internal/auth"
)

func TestNewRouter_InvalidTrustedProxies_ReturnsError(t *testing.T) {
	_, err := NewRouter([]string{"not-a-cidr-or-ip"})
	if err == nil {
		t.Fatal("expected NewRouter to return error for invalid trusted proxies")
	}
}

func TestNewRouterWithDependencies_UsesInjectedAuthHandler(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := auth.NewInMemoryRepository(func() time.Time { return now })
	sender := &countingSender{}
	handler := auth.NewHandler(auth.NewCodeService(repo, sender, nil, func() time.Time { return now }, func() string { return "123456" }))

	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: handler})
	if err != nil {
		t.Fatalf("expected NewRouterWithDependencies to succeed, got: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/send-code", strings.NewReader(`{"email":"inject@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if sender.calls != 1 {
		t.Fatalf("expected injected sender to be called once, got %d", sender.calls)
	}
}

func TestNewRouterWithDependencies_DoesNotMountSessionRoutesForCodeOnlyHandler(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := auth.NewInMemoryRepository(func() time.Time { return now })
	handler := auth.NewHandler(auth.NewCodeService(repo, &countingSender{}, nil, func() time.Time { return now }, func() string { return "123456" }))

	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: handler})
	if err != nil {
		t.Fatalf("expected NewRouterWithDependencies to succeed, got: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", strings.NewReader(`{"refresh_token":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected refresh route to be absent for code-only handler, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAccountingRoutes_RejectExpiredAccessToken(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := auth.NewInMemoryRepository(func() time.Time { return now })
	handler := auth.NewHandler(auth.NewCodeService(repo, &countingSender{}, nil, func() time.Time { return now }, func() string { return "123456" }))

	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: handler})
	if err != nil {
		t.Fatalf("expected NewRouterWithDependencies to succeed, got: %v", err)
	}

	pastNow := func() time.Time { return time.Now().UTC().Add(-2 * time.Hour) }
	issued, err := auth.NewSessionService(repo, nil, pastNow).IssueSession(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("issue expired session token: %v", err)
	}
	expired := issued.AccessToken
	req := httptest.NewRequest(http.MethodGet, "/api/ledgers", nil)
	req.Header.Set("Authorization", "Bearer "+expired)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rec.Code, rec.Body.String())
	}
}

func TestDefaultHandlers_ShareClassificationStateWithTransactions(t *testing.T) {
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	repo := auth.NewInMemoryRepository(func() time.Time { return now })
	authHandler := auth.NewHandler(auth.NewCodeService(repo, &countingSender{}, nil, func() time.Time { return now }, func() string { return "123456" }))

	r, err := NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, Dependencies{AuthHandler: authHandler})
	if err != nil {
		t.Fatalf("expected NewRouterWithDependencies to succeed, got: %v", err)
	}
	authz := "pat:shared@example.com"

	ledgerID := responseField(t, performJSON(t, r, http.MethodPost, "/api/ledgers", `{"name":"Main","is_default":true}`, authz, http.StatusCreated), "id")
	categoryID := responseField(t, performJSON(t, r, http.MethodPost, "/api/categories", `{"name":"Salary"}`, authz, http.StatusCreated), "id")
	tagID := responseField(t, performJSON(t, r, http.MethodPost, "/api/tags", `{"name":"monthly"}`, authz, http.StatusCreated), "id")

	txnBody := `{"ledger_id":"` + ledgerID + `","type":"income","amount":100,"category_id":"` + categoryID + `","tag_ids":["` + tagID + `"]}`
	txnResp := performJSON(t, r, http.MethodPost, "/api/transactions", txnBody, authz, http.StatusCreated)
	if got := responseField(t, txnResp, "category_name"); got != "Salary" {
		t.Fatalf("expected transaction category snapshot Salary, got %q", got)
	}

	deleteResp := performJSON(t, r, http.MethodDelete, "/api/categories/"+categoryID, ``, authz, http.StatusOK)
	if archived, ok := deleteResp["archived"].(bool); !ok || !archived {
		t.Fatalf("expected referenced category delete to archive, got %#v", deleteResp)
	}
}

func performJSON(t *testing.T, handler http.Handler, method string, path string, body string, accessToken string, wantStatus int) map[string]any {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("expected status %d for %s %s, got %d body=%s", wantStatus, method, path, rec.Code, rec.Body.String())
	}
	if rec.Body.Len() == 0 {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response for %s %s: %v body=%s", method, path, err, rec.Body.String())
	}
	return payload
}

func responseField(t *testing.T, payload map[string]any, key string) string {
	t.Helper()
	value, ok := payload[key].(string)
	if !ok {
		t.Fatalf("expected string field %q in %#v", key, payload)
	}
	return value
}

type countingSender struct {
	calls int
}

func (s *countingSender) Send(to string, subject string, body string) error {
	s.calls++
	return nil
}
