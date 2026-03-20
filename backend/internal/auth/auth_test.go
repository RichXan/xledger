package auth

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"
)

func TestSendCode_UsesSMTPAndRedis(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	sender := &stubSender{}
	svc := NewCodeService(repo, sender, nil, func() time.Time { return now }, func() string { return "123456" })

	err := svc.SendCode(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("expected send code to succeed, got error: %v", err)
	}

	if sender.calls != 1 {
		t.Fatalf("expected SMTP sender to be called once, got %d", sender.calls)
	}

	if repo.StoredCode("user@example.com") != "123456" {
		t.Fatalf("expected code to be stored in repository")
	}
}

func TestSendCode_GeneratesSixDigitCode(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	sender := &stubSender{}
	svc := NewCodeService(repo, sender, nil, func() time.Time { return now }, nil)

	err := svc.SendCode(context.Background(), "digits@example.com")
	if err != nil {
		t.Fatalf("expected send code to succeed, got error: %v", err)
	}

	stored := repo.StoredCode("digits@example.com")
	if !regexp.MustCompile(`^\d{6}$`).MatchString(stored) {
		t.Fatalf("expected six digit numeric code, got %q", stored)
	}
}

func TestVerifyCode_Success_ReturnsAccessAndRefresh(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveVerificationCode(context.Background(), "ok@example.com", "222222", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}
	svc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)

	tokens, err := svc.VerifyCode(context.Background(), "ok@example.com", "222222")
	if err != nil {
		t.Fatalf("expected verify code to succeed, got error: %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected access and refresh tokens, got %#v", tokens)
	}
}

func TestVerifyCode_Invalid_ReturnsAUTH_CODE_INVALID(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveVerificationCode(context.Background(), "invalid@example.com", "111111", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}
	svc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)

	_, err := svc.VerifyCode(context.Background(), "invalid@example.com", "999999")
	if ErrorCode(err) != AUTH_CODE_INVALID {
		t.Fatalf("expected %s, got %q", AUTH_CODE_INVALID, ErrorCode(err))
	}
}

func TestVerifyCode_Expired_ReturnsAUTH_CODE_EXPIRED(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveVerificationCode(context.Background(), "expired@example.com", "333333", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}
	now = now.Add(11 * time.Minute)
	svc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)

	_, err := svc.VerifyCode(context.Background(), "expired@example.com", "333333")
	if ErrorCode(err) != AUTH_CODE_EXPIRED {
		t.Fatalf("expected %s, got %q", AUTH_CODE_EXPIRED, ErrorCode(err))
	}
}

func TestSendCode_SMTPFailure_ReturnsAUTH_CODE_SEND_FAILED(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewCodeService(
		repo,
		&stubSender{err: errors.New("smtp down")},
		nil,
		func() time.Time { return now },
		func() string { return "123456" },
	)

	err := svc.SendCode(context.Background(), "smtpfail@example.com")
	if ErrorCode(err) != AUTH_CODE_SEND_FAILED {
		t.Fatalf("expected %s, got %q", AUTH_CODE_SEND_FAILED, ErrorCode(err))
	}
}

func TestSendCode_SMTPFailure_CreatesNoSession(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewCodeService(
		repo,
		&stubSender{err: errors.New("smtp down")},
		nil,
		func() time.Time { return now },
		func() string { return "123456" },
	)

	_ = svc.SendCode(context.Background(), "smtpnosession@example.com")

	if repo.SessionCount() != 0 {
		t.Fatalf("expected no session side effects on SMTP failure, got %d", repo.SessionCount())
	}
}

func TestSendCode_SMTPFailure_DoesNotBypassResendLimit(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	sender := &stubSender{err: errors.New("smtp down")}
	svc := NewCodeService(repo, sender, nil, func() time.Time { return now }, func() string { return "123456" })

	err := svc.SendCode(context.Background(), "ratelimit-after-fail@example.com")
	if ErrorCode(err) != AUTH_CODE_SEND_FAILED {
		t.Fatalf("expected first attempt %s, got %q", AUTH_CODE_SEND_FAILED, ErrorCode(err))
	}

	now = now.Add(30 * time.Second)
	err = svc.SendCode(context.Background(), "ratelimit-after-fail@example.com")
	if ErrorCode(err) != AUTH_CODE_RATE_LIMIT {
		t.Fatalf("expected retry within 60s to return %s, got %q", AUTH_CODE_RATE_LIMIT, ErrorCode(err))
	}

	if sender.calls != 1 {
		t.Fatalf("expected SMTP sender not called on rate-limited retry, got %d calls", sender.calls)
	}
}

func TestSendCode_TooFrequent_ReturnsAUTH_CODE_RATE_LIMIT(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	sender := &stubSender{}
	svc := NewCodeService(repo, sender, nil, func() time.Time { return now }, func() string { return "111111" })

	if err := svc.SendCode(context.Background(), "fast@example.com"); err != nil {
		t.Fatalf("first send should succeed, got: %v", err)
	}

	now = now.Add(30 * time.Second)
	err := svc.SendCode(context.Background(), "fast@example.com")
	if ErrorCode(err) != AUTH_CODE_RATE_LIMIT {
		t.Fatalf("expected %s, got %q", AUTH_CODE_RATE_LIMIT, ErrorCode(err))
	}
}

type stubSender struct {
	err   error
	calls int
}

func (s *stubSender) Send(to, subject, body string) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	return nil
}
