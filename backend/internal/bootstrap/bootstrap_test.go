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
	router, err := bootstraphttp.NewRouter([]string{"127.0.0.1", "::1"})
	if err != nil {
		t.Fatalf("expected router creation to succeed, got error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestConfig_AllowsMissingSMTPEnv(t *testing.T) {
	t.Setenv("SMTP_HOST", "")
	t.Setenv("AUTH_CODE_PEPPER", "test-pepper")
	t.Setenv("AUTH_TOKEN_SECRET", "test-token-secret")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config load to succeed with empty SMTP host, got %v", err)
	}
	if cfg.SMTPHost != "" {
		t.Fatalf("expected empty SMTP host, got %q", cfg.SMTPHost)
	}
}

func TestConfig_RequiresAuthCodePepperEnv(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("AUTH_CODE_PEPPER", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when AUTH_CODE_PEPPER is missing")
	}

	if !strings.Contains(err.Error(), "AUTH_CODE_PEPPER") {
		t.Fatalf("expected error to mention AUTH_CODE_PEPPER, got %q", err.Error())
	}
}

func TestConfig_DefaultsAPIAddr(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("AUTH_CODE_PEPPER", "test-pepper")
	t.Setenv("AUTH_TOKEN_SECRET", "test-token-secret")
	t.Setenv("API_ADDR", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}

	if cfg.APIAddr != ":8080" {
		t.Fatalf("expected default APIAddr %q, got %q", ":8080", cfg.APIAddr)
	}
}

func TestConfig_DefaultsTrustedProxies(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("AUTH_CODE_PEPPER", "test-pepper")
	t.Setenv("AUTH_TOKEN_SECRET", "test-token-secret")
	t.Setenv("TRUSTED_PROXIES", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}

	if len(cfg.TrustedProxies) != 2 {
		t.Fatalf("expected 2 default trusted proxies, got %d", len(cfg.TrustedProxies))
	}

	if cfg.TrustedProxies[0] != "127.0.0.1" || cfg.TrustedProxies[1] != "::1" {
		t.Fatalf("unexpected default trusted proxies: %#v", cfg.TrustedProxies)
	}
}

func TestConfig_ParsesTrustedProxiesFromEnv(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("AUTH_CODE_PEPPER", "test-pepper")
	t.Setenv("AUTH_TOKEN_SECRET", "test-token-secret")
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8, 192.168.0.0/16")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}

	if len(cfg.TrustedProxies) != 2 {
		t.Fatalf("expected 2 trusted proxies from env, got %d", len(cfg.TrustedProxies))
	}

	if cfg.TrustedProxies[0] != "10.0.0.0/8" || cfg.TrustedProxies[1] != "192.168.0.0/16" {
		t.Fatalf("unexpected trusted proxies from env: %#v", cfg.TrustedProxies)
	}
}
