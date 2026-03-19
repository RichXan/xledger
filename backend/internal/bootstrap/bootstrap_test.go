package bootstrap_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"xledger/backend/internal/bootstrap/config"
	bootstraphttp "xledger/backend/internal/bootstrap/http"
)

func TestRouter_Healthz(t *testing.T) {
	router := bootstraphttp.NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestConfig_RequiresSMTPEnv(t *testing.T) {
	t.Setenv("SMTP_HOST", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when SMTP_HOST is missing")
	}

	if !strings.Contains(err.Error(), "SMTP_HOST") {
		t.Fatalf("expected error to mention SMTP_HOST, got %q", err.Error())
	}
}
