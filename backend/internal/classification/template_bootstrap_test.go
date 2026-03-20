package classification_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"xledger/backend/internal/auth"
	"xledger/backend/internal/classification"
)

func TestFirstLogin_CopiesDefaultCategoryTemplateOnce(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	authRepo := auth.NewInMemoryRepository(func() time.Time { return now })
	templateRepo := classification.NewInMemoryTemplateRepository()
	templateService := classification.NewTemplateService(templateRepo)

	sessionService := auth.NewSessionService(authRepo, &auth.SessionServiceOptions{
		PostLoginBootstrap: templateService.EnsureUserDefaults,
	}, func() time.Time { return now })

	if _, err := sessionService.IssueSession(context.Background(), "first-login@example.com"); err != nil {
		t.Fatalf("first session issue should succeed: %v", err)
	}
	if _, err := sessionService.IssueSession(context.Background(), "first-login@example.com"); err != nil {
		t.Fatalf("second session issue should succeed: %v", err)
	}

	if templateRepo.CopyCount("first-login@example.com") != 1 {
		t.Fatalf("expected default template copied once, got %d", templateRepo.CopyCount("first-login@example.com"))
	}
}

func TestFirstLogin_TemplateBootstrapFailure_DoesNotBlockSessionIssue(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	authRepo := auth.NewInMemoryRepository(func() time.Time { return now })

	sessionService := auth.NewSessionService(authRepo, &auth.SessionServiceOptions{
		PostLoginBootstrap: func(context.Context, string) error {
			return errors.New("template bootstrap failed")
		},
	}, func() time.Time { return now })

	if _, err := sessionService.IssueSession(context.Background(), "bootstrap-failure@example.com"); err != nil {
		t.Fatalf("session issue should remain successful when template bootstrap fails: %v", err)
	}

	if authRepo.AlertEventCount("auth.session.user_defaults_bootstrap_failed") != 1 {
		t.Fatalf("expected bootstrap failure alert count 1, got %d", authRepo.AlertEventCount("auth.session.user_defaults_bootstrap_failed"))
	}
}
