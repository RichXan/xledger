package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"
)

func TestSendCode_UsesSMTPAndRedis(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	sender := &stubSender{}
	svc := NewCodeService(repo, sender, nil, func() time.Time { return now }, func() string { return "123456" })

	err := svc.SendCode(context.Background(), "user@example.com", "10.0.0.1")
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

	err := svc.SendCode(context.Background(), "digits@example.com", "10.0.0.2")
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

func TestVerifyCode_DeleteFailure_DoesNotCreateSession(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	baseRepo := NewInMemoryRepository(func() time.Time { return now })
	if err := baseRepo.SaveVerificationCode(context.Background(), "deletefail@example.com", "444444", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}
	repo := &deleteFailRepository{InMemoryRepository: baseRepo}
	svc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)

	_, err := svc.VerifyCode(context.Background(), "deletefail@example.com", "444444")
	if err == nil {
		t.Fatal("expected verify to fail when code deletion fails")
	}
	if repo.SessionCount() != 0 {
		t.Fatalf("expected no session side effect when verify fails, got %d", repo.SessionCount())
	}
}

func TestVerifyCode_RepeatedWrongAttempts_InvalidateCode(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveVerificationCode(context.Background(), "bruteforce@example.com", "555555", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}
	svc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)

	for i := 0; i < 5; i++ {
		_, err := svc.VerifyCode(context.Background(), "bruteforce@example.com", "000000")
		if ErrorCode(err) != AUTH_CODE_INVALID {
			t.Fatalf("attempt %d expected %s, got %q", i+1, AUTH_CODE_INVALID, ErrorCode(err))
		}
	}

	_, err := svc.VerifyCode(context.Background(), "bruteforce@example.com", "555555")
	if ErrorCode(err) != AUTH_CODE_INVALID {
		t.Fatalf("expected code to be invalidated after repeated failures, got %q", ErrorCode(err))
	}
	if repo.SessionCount() != 0 {
		t.Fatalf("expected no session created after brute-force invalidation, got %d", repo.SessionCount())
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

	err := svc.SendCode(context.Background(), "smtpfail@example.com", "10.0.0.3")
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

	_ = svc.SendCode(context.Background(), "smtpnosession@example.com", "10.0.0.4")

	if repo.SessionCount() != 0 {
		t.Fatalf("expected no session side effects on SMTP failure, got %d", repo.SessionCount())
	}
}

func TestSendCode_SMTPFailure_DoesNotBypassResendLimit(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	sender := &stubSender{err: errors.New("smtp down")}
	svc := NewCodeService(repo, sender, nil, func() time.Time { return now }, func() string { return "123456" })

	err := svc.SendCode(context.Background(), "ratelimit-after-fail@example.com", "10.0.0.5")
	if ErrorCode(err) != AUTH_CODE_SEND_FAILED {
		t.Fatalf("expected first attempt %s, got %q", AUTH_CODE_SEND_FAILED, ErrorCode(err))
	}

	now = now.Add(30 * time.Second)
	err = svc.SendCode(context.Background(), "ratelimit-after-fail@example.com", "10.0.0.5")
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

	if err := svc.SendCode(context.Background(), "fast@example.com", "10.0.0.6"); err != nil {
		t.Fatalf("first send should succeed, got: %v", err)
	}

	now = now.Add(30 * time.Second)
	err := svc.SendCode(context.Background(), "fast@example.com", "10.0.0.6")
	if ErrorCode(err) != AUTH_CODE_RATE_LIMIT {
		t.Fatalf("expected %s, got %q", AUTH_CODE_RATE_LIMIT, ErrorCode(err))
	}
}

func TestSendCode_IPHourlyCap_ReturnsRateLimitAcrossDifferentEmails(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	sender := &stubSender{}
	svc := NewCodeService(repo, sender, nil, func() time.Time { return now }, func() string { return "222222" })

	for i := 0; i < svc.ipHourlyCap; i++ {
		email := fmt.Sprintf("ipcap-%d@example.com", i)
		if err := svc.SendCode(context.Background(), email, "203.0.113.7"); err != nil {
			t.Fatalf("attempt %d should succeed, got: %v", i+1, err)
		}
	}

	err := svc.SendCode(context.Background(), "ipcap-over@example.com", "203.0.113.7")
	if ErrorCode(err) != AUTH_CODE_RATE_LIMIT {
		t.Fatalf("expected %s after hourly IP cap, got %q", AUTH_CODE_RATE_LIMIT, ErrorCode(err))
	}
	if sender.calls != svc.ipHourlyCap {
		t.Fatalf("expected sender calls to stop at cap=%d, got %d", svc.ipHourlyCap, sender.calls)
	}
}

func TestInMemoryRepository_OpportunisticCleanup_RemovesExpiredStaleEntries(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })

	if err := repo.SaveVerificationCode(context.Background(), "stale@example.com", "999999", time.Second); err != nil {
		t.Fatalf("save stale code: %v", err)
	}
	ok, err := repo.AcquireSendLock(context.Background(), "stale@example.com", now, time.Second)
	if err != nil || !ok {
		t.Fatalf("acquire stale lock: ok=%v err=%v", ok, err)
	}
	ipOK, ipErr := repo.AcquireIPHourlySlot(context.Background(), "198.51.100.20", now, time.Second, 5)
	if ipErr != nil || !ipOK {
		t.Fatalf("acquire stale ip lock: ok=%v err=%v", ipOK, ipErr)
	}

	now = now.Add(2 * time.Second)
	if err := repo.SaveVerificationCode(context.Background(), "fresh@example.com", "123123", 10*time.Minute); err != nil {
		t.Fatalf("save fresh code: %v", err)
	}

	codeCount, lockCount, failureCount, ipCount := repo.stateCounts()
	if codeCount != 1 || lockCount != 0 || failureCount != 1 || ipCount != 0 {
		t.Fatalf("unexpected state counts after cleanup: codes=%d locks=%d failures=%d ip=%d", codeCount, lockCount, failureCount, ipCount)
	}
}

type stubSender struct {
	err   error
	calls int
}

type deleteFailRepository struct {
	*InMemoryRepository
}

func (r *deleteFailRepository) DeleteVerificationCode(ctx context.Context, email string) error {
	return errors.New("delete failed")
}

func (s *stubSender) Send(to, subject, body string) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	return nil
}
