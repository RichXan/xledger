package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"xledger/backend/internal/auth"
	bootstraphttp "xledger/backend/internal/bootstrap/http"
)

func TestAPIContract_AuthSendCode_UsesUnifiedEnvelope(t *testing.T) {
	r := newContractRouter(t)
	resp := performJSONEnvelope(t, r, http.MethodPost, "/api/auth/send-code", `{"email":"contract@example.com"}`, "", http.StatusOK)
	assertEnvelopeShape(t, resp)
	assertCode(t, resp, "OK")
}

func TestAPIContract_AccountsUpdate_UsesPATCH_NotPUT(t *testing.T) {
	r := newContractRouter(t)
	accessToken := issueContractAccessToken(t, r)
	accountID := responseStringFromEnvelope(t, performJSONEnvelope(t, r, http.MethodPost, "/api/accounts", `{"name":"Cash","type":"cash","initial_balance":0}`, accessToken, http.StatusCreated), "id")

	performJSONRaw(t, r, http.MethodPut, "/api/accounts/"+accountID, `{"name":"Cash 2"}`, accessToken, http.StatusNotFound)
	resp := performJSONEnvelope(t, r, http.MethodPatch, "/api/accounts/"+accountID, `{"name":"Cash 2"}`, accessToken, http.StatusOK)
	assertEnvelopeShape(t, resp)
	assertCode(t, resp, "OK")
}

func TestAPIContract_PersonalAccessTokens_UsesResourcePath(t *testing.T) {
	r := newContractRouter(t)
	accessToken := issueContractAccessToken(t, r)
	resp := performJSONEnvelope(t, r, http.MethodGet, "/api/personal-access-tokens", ``, accessToken, http.StatusOK)
	assertEnvelopeShape(t, resp)
	assertCode(t, resp, "OK")
}

func TestAPIContract_AuthMe_ExistsAndUsesEnvelope(t *testing.T) {
	r := newContractRouter(t)
	accessToken := issueContractAccessToken(t, r)
	resp := performJSONEnvelope(t, r, http.MethodGet, "/api/auth/me", ``, accessToken, http.StatusOK)
	assertEnvelopeShape(t, resp)
	assertCode(t, resp, "OK")
}

func TestAPIContract_ListResponse_UsesDataItemsPagination(t *testing.T) {
	r := newContractRouter(t)
	accessToken := issueContractAccessToken(t, r)
	resp := performJSONEnvelope(t, r, http.MethodGet, "/api/accounts?page=1&page_size=20", ``, accessToken, http.StatusOK)
	assertEnvelopeShape(t, resp)
	data := responseMap(t, resp, "data")
	if _, ok := data["items"].([]any); !ok {
		t.Fatalf("expected data.items array in %#v", data)
	}
	if _, ok := data["pagination"].(map[string]any); !ok {
		t.Fatalf("expected data.pagination object in %#v", data)
	}
}

func newContractRouter(t *testing.T) http.Handler {
	t.Helper()
	now := func() time.Time { return time.Now().UTC() }
	repo := auth.NewInMemoryRepository(now)
	sessionService := auth.NewSessionService(repo, nil, now)
	authHandler := auth.NewHandler(auth.NewCodeService(repo, contractNoopSender{}, auth.NewSessionTokenIssuer(sessionService), now, func() string { return "123456" }))
	r, err := bootstraphttp.NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, bootstraphttp.Dependencies{AuthHandler: authHandler})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}
	return r
}

func issueContractAccessToken(t *testing.T, handler http.Handler) string {
	t.Helper()
	_ = performJSONRaw(t, handler, http.MethodPost, "/api/auth/send-code", `{"email":"contract@example.com"}`, "", http.StatusOK)
	verify := performJSONRaw(t, handler, http.MethodPost, "/api/auth/verify-code", `{"email":"contract@example.com","code":"123456"}`, "", http.StatusOK)
	return responseStringFromEnvelope(t, verify, "access_token")
}

func performJSONEnvelope(t *testing.T, handler http.Handler, method string, path string, body string, accessToken string, wantStatus int) map[string]any {
	t.Helper()
	payload := performJSONRaw(t, handler, method, path, body, accessToken, wantStatus)
	return payload
}

func performJSONRaw(t *testing.T, handler http.Handler, method string, path string, body string, accessToken string, wantStatus int) map[string]any {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("expected status %d for %s %s, got %d body=%s", wantStatus, method, path, rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode json: %v body=%s", err, rec.Body.String())
	}
	return payload
}

func assertEnvelopeShape(t *testing.T, payload map[string]any) {
	t.Helper()
	if _, ok := payload["code"].(string); !ok {
		t.Fatalf("expected code in envelope, got %#v", payload)
	}
	if _, ok := payload["message"].(string); !ok {
		t.Fatalf("expected message in envelope, got %#v", payload)
	}
	if _, ok := payload["data"]; !ok {
		t.Fatalf("expected data in envelope, got %#v", payload)
	}
}

func assertCode(t *testing.T, payload map[string]any, want string) {
	t.Helper()
	if got, _ := payload["code"].(string); got != want {
		t.Fatalf("expected code %q, got %#v", want, payload)
	}
}

func responseMap(t *testing.T, payload map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := payload[key].(map[string]any)
	if !ok {
		t.Fatalf("expected map field %q in %#v", key, payload)
	}
	return value
}

func responseStringFromEnvelope(t *testing.T, payload map[string]any, key string) string {
	t.Helper()
	data := responseMap(t, payload, "data")
	value, ok := data[key].(string)
	if !ok || value == "" {
		t.Fatalf("expected string field %q in envelope %#v", key, payload)
	}
	return value
}

func responseStringRaw(t *testing.T, payload map[string]any, key string) string {
	t.Helper()
	value, ok := payload[key].(string)
	if !ok || value == "" {
		t.Fatalf("expected string field %q in raw payload %#v", key, payload)
	}
	return value
}

type contractNoopSender struct{}

func (contractNoopSender) Send(string, string, string) error { return nil }
