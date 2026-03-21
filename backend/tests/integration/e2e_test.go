package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"xledger/backend/internal/auth"
	bootstraphttp "xledger/backend/internal/bootstrap/http"
)

func TestE2E_CorePath_AuthToExport_WithoutN8N(t *testing.T) {
	now := func() time.Time { return time.Now().UTC() }
	repo := auth.NewInMemoryRepository(now)
	authHandler := auth.NewHandler(auth.NewCodeService(repo, noopSender{}, nil, now, func() string { return "123456" }))
	r, err := bootstraphttp.NewRouterWithDependencies([]string{"127.0.0.1", "::1"}, bootstraphttp.Dependencies{AuthHandler: authHandler})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	performJSONStatus(t, r, http.MethodPost, "/api/auth/send-code", `{"email":"e2e@example.com"}`, "", http.StatusOK)
	verifyResp := performJSONStatus(t, r, http.MethodPost, "/api/auth/verify-code", `{"email":"e2e@example.com","code":"123456"}`, "", http.StatusOK)
	accessToken := responseStringFromData(t, verifyResp, "access_token")

	ledgerResp := performJSONStatus(t, r, http.MethodPost, "/api/ledgers", `{"name":"Main","is_default":true}`, accessToken, http.StatusCreated)
	ledgerID := responseStringFromData(t, ledgerResp, "id")
	performJSONStatus(t, r, http.MethodPost, "/api/transactions", `{"ledger_id":"`+ledgerID+`","type":"expense","amount":25}`, accessToken, http.StatusCreated)

	body, contentType := buildCSVMultipart(t, "date,amount,description\n2026-03-01,12.5,lunch\n")
	performMultipartStatus(t, r, "/api/import/csv", accessToken, body, contentType, http.StatusOK)
	performExportStatus(t, r, "/api/export?format=json", accessToken, http.StatusOK)
}

func performJSONStatus(t *testing.T, handler http.Handler, method string, path string, body string, accessToken string, wantStatus int) map[string]any {
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
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode json: %v body=%s", err, rec.Body.String())
		}
	}
	return payload
}

func performMultipartStatus(t *testing.T, handler http.Handler, path string, accessToken string, body *bytes.Buffer, contentType string, wantStatus int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("expected status %d for multipart %s, got %d body=%s", wantStatus, path, rec.Code, rec.Body.String())
	}
}

func performExportStatus(t *testing.T, handler http.Handler, path string, accessToken string, wantStatus int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("expected status %d for export %s, got %d body=%s", wantStatus, path, rec.Code, rec.Body.String())
	}
}

func buildCSVMultipart(t *testing.T, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "preview.csv")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write csv content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func responseString(t *testing.T, payload map[string]any, key string) string {
	t.Helper()
	value, ok := payload[key].(string)
	if !ok || value == "" {
		t.Fatalf("expected string field %s in %#v", key, payload)
	}
	return value
}

func responseStringFromData(t *testing.T, payload map[string]any, key string) string {
	t.Helper()
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data map in %#v", payload)
	}
	value, ok := data[key].(string)
	if !ok || value == "" {
		t.Fatalf("expected string field %s in data %#v", key, payload)
	}
	return value
}

type noopSender struct{}

func (noopSender) Send(string, string, string) error { return nil }

func init() {
	_ = time.Now
}
