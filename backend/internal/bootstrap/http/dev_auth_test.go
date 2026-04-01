package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDevLoginEndpoint_AvailableOnlyOutsideRelease(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("AUTH_CODE_PEPPER", "test-pepper")
	t.Setenv("AUTH_TOKEN_SECRET", "test-token-secret")
	t.Setenv("ENABLE_DEV_LOGIN", "1")

	prevMode := os.Getenv("GIN_MODE")
	defer os.Setenv("GIN_MODE", prevMode)

	if err := os.Setenv("GIN_MODE", "release"); err != nil {
		t.Fatalf("set release mode: %v", err)
	}
	r, err := NewRouter(nil)
	if err != nil {
		t.Fatalf("new router in release: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/auth/dev-login", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected dev login hidden in release, got %d body=%s", rec.Code, rec.Body.String())
	}

	if err := os.Setenv("GIN_MODE", "debug"); err != nil {
		t.Fatalf("set debug mode: %v", err)
	}
	r, err = NewRouter(nil)
	if err != nil {
		t.Fatalf("new router in debug: %v", err)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/auth/dev-login", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected dev login in debug, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Code string `json:"code"`
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			Email        string `json:"email"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Code != "OK" || payload.Data.AccessToken == "" || payload.Data.RefreshToken == "" || payload.Data.Email == "" {
		t.Fatalf("expected non-empty dev login payload, got %+v", payload)
	}
}
