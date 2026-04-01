package portability

import (
	"context"
	"testing"
	"time"
)

func TestPAT_CannotAccessAuthEndpoints(t *testing.T) {
	service := NewPATService(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	plain, _, err := service.CreatePAT(context.Background(), "user-1", "cli", nil)
	if err != nil {
		t.Fatalf("create pat: %v", err)
	}
	if service.CanUsePATOnPath(plain, "/api/auth/refresh") {
		t.Fatalf("expected PAT to be forbidden on auth endpoints")
	}
}

func TestPATEndpoints_AccessOnly_GETPOSTDELETE_On_api_settings_tokens(t *testing.T) {
	service := NewPATService(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	plain, _, err := service.CreatePAT(context.Background(), "user-1", "cli", nil)
	if err != nil {
		t.Fatalf("create pat: %v", err)
	}
	if service.CanUsePATOnPath(plain, "/api/personal-access-tokens") {
		t.Fatalf("expected PAT to be forbidden from PAT management endpoints")
	}
}

func TestPAT_DefaultTTL90Days_StoredAsHash(t *testing.T) {
	now := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	service := NewPATService(func() time.Time { return now })
	plain, created, err := service.CreatePAT(context.Background(), "user-1", "cli", nil)
	if err != nil {
		t.Fatalf("create pat: %v", err)
	}
	if created.TokenHash == plain {
		t.Fatalf("expected hash-only storage, got raw token persisted")
	}
	if created.ExpiresAt == nil {
		t.Fatalf("expected default ttl 90d, got nil")
	}
	if created.ExpiresAt.Sub(now) != 90*24*time.Hour {
		t.Fatalf("expected default ttl 90d, got %s", created.ExpiresAt.Sub(now))
	}
}

func TestPAT_Expired_ReturnsPAT_EXPIRED(t *testing.T) {
	now := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	service := NewPATService(func() time.Time { return now })
	plain, _, err := service.CreatePAT(context.Background(), "user-1", "cli", ptrTime(now.Add(1*time.Hour)))
	if err != nil {
		t.Fatalf("create pat: %v", err)
	}
	service.SetNow(func() time.Time { return now.Add(2 * time.Hour) })
	_, err = service.ValidatePAT(context.Background(), plain, "/api/transactions")
	if ErrorCode(err) != PAT_EXPIRED {
		t.Fatalf("expected %s, got %q", PAT_EXPIRED, ErrorCode(err))
	}
}

func TestPAT_Revoked_ReturnsPAT_REVOKED(t *testing.T) {
	service := NewPATService(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	plain, created, err := service.CreatePAT(context.Background(), "user-1", "cli", nil)
	if err != nil {
		t.Fatalf("create pat: %v", err)
	}
	if err := service.RevokePAT(context.Background(), "user-1", created.ID); err != nil {
		t.Fatalf("revoke pat: %v", err)
	}
	_, err = service.ValidatePAT(context.Background(), plain, "/api/transactions")
	if ErrorCode(err) != PAT_REVOKED {
		t.Fatalf("expected %s, got %q", PAT_REVOKED, ErrorCode(err))
	}
}

func TestPAT_RevokePropagationOver5s_TriggersStrictModeAlert(t *testing.T) {
	now := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	service := NewPATService(func() time.Time { return now })
	plain, created, err := service.CreatePAT(context.Background(), "user-1", "cli", nil)
	if err != nil {
		t.Fatalf("create pat: %v", err)
	}
	if err := service.RevokePAT(context.Background(), "user-1", created.ID); err != nil {
		t.Fatalf("revoke pat: %v", err)
	}
	service.SetRevocationLag(6 * time.Second)
	_, err = service.ValidatePAT(context.Background(), plain, "/api/transactions")
	if ErrorCode(err) != PAT_REVOKED {
		t.Fatalf("expected %s under strict mode, got %q", PAT_REVOKED, ErrorCode(err))
	}
	if !service.StrictModeEnabled() {
		t.Fatalf("expected strict mode enabled after revoke propagation breach")
	}
	if alerts := service.AlertEvents(); len(alerts) == 0 {
		t.Fatalf("expected alert event after revoke propagation breach")
	}
}

func ptrTime(value time.Time) *time.Time { return &value }
