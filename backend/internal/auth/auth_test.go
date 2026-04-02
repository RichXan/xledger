package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
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

	if repo.StoredCode("user@example.com") == "123456" {
		t.Fatalf("expected code not to be stored in plaintext")
	}
	if repo.StoredCode("user@example.com") != hashVerificationCode("123456") {
		t.Fatalf("expected hashed code to be stored in repository")
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

	codeMatch := regexp.MustCompile(`\b\d{6}\b`).FindString(sender.lastBody)
	if codeMatch == "" {
		t.Fatalf("expected email body to include six digit code, got %q", sender.lastBody)
	}
	stored := repo.StoredCode("digits@example.com")
	if stored != hashVerificationCode(codeMatch) {
		t.Fatalf("expected repository to store code hash, got %q", stored)
	}
}

func TestVerifyCode_HashedStorage_SupportsSuccessAndFailure(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveVerificationCode(context.Background(), "hashflow@example.com", "676767", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}
	if repo.StoredCode("hashflow@example.com") == "676767" {
		t.Fatal("expected stored code to be hashed")
	}

	badSvc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)
	if _, err := badSvc.VerifyCode(context.Background(), "hashflow@example.com", "000000"); ErrorCode(err) != AUTH_CODE_INVALID {
		t.Fatalf("expected wrong code to fail with %s, got %q", AUTH_CODE_INVALID, ErrorCode(err))
	}

	if err := repo.SaveVerificationCode(context.Background(), "hashflow@example.com", "676767", 10*time.Minute); err != nil {
		t.Fatalf("reseed code: %v", err)
	}
	goodSvc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)
	if _, err := goodSvc.VerifyCode(context.Background(), "hashflow@example.com", "676767"); err != nil {
		t.Fatalf("expected correct code to verify from hashed storage, got: %v", err)
	}
}

func TestHashVerificationCode_UsesConfiguredPepperDeterministically(t *testing.T) {
	t.Setenv("AUTH_CODE_PEPPER", "pepper-a")
	h1 := hashVerificationCode("123456")
	h2 := hashVerificationCode("123456")
	if h1 != h2 {
		t.Fatalf("expected deterministic digest for same pepper")
	}

	t.Setenv("AUTH_CODE_PEPPER", "pepper-b")
	h3 := hashVerificationCode("123456")
	if h1 == h3 {
		t.Fatalf("expected digest to change when pepper changes")
	}
}

func TestVerifyCode_WithConfiguredPepper_Succeeds(t *testing.T) {
	t.Setenv("AUTH_CODE_PEPPER", "pepper-verify")
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveVerificationCode(context.Background(), "pepper@example.com", "565656", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}

	svc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)
	if _, err := svc.VerifyCode(context.Background(), "pepper@example.com", "565656"); err != nil {
		t.Fatalf("expected verify to succeed with configured pepper, got: %v", err)
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

func TestVerifyCode_ConsumeFailure_DoesNotCreateSession(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	baseRepo := NewInMemoryRepository(func() time.Time { return now })
	if err := baseRepo.SaveVerificationCode(context.Background(), "deletefail@example.com", "444444", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}
	repo := &consumeFailRepository{InMemoryRepository: baseRepo}
	svc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)

	_, err := svc.VerifyCode(context.Background(), "deletefail@example.com", "444444")
	if err == nil {
		t.Fatal("expected verify to fail when consume operation fails")
	}
	if repo.SessionCount() != 0 {
		t.Fatalf("expected no session side effect when verify fails, got %d", repo.SessionCount())
	}
}

func TestInMemoryRepository_VerifyAndConsumeCode_AllowsSingleSuccess(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveVerificationCode(context.Background(), "single-use@example.com", "818181", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}

	result1, err := repo.VerifyAndConsumeCode(context.Background(), "single-use@example.com", hashVerificationCode("818181"), 5)
	if err != nil {
		t.Fatalf("first consume error: %v", err)
	}
	if result1 != VerifyConsumeMatch {
		t.Fatalf("expected first consume to match, got %v", result1)
	}

	result2, err := repo.VerifyAndConsumeCode(context.Background(), "single-use@example.com", hashVerificationCode("818181"), 5)
	if err != nil {
		t.Fatalf("second consume error: %v", err)
	}
	if result2 == VerifyConsumeMatch {
		t.Fatalf("expected second consume not to match, got %v", result2)
	}
}

func TestVerifyCode_ConcurrentRequests_AtMostOneSuccess(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	if err := repo.SaveVerificationCode(context.Background(), "concurrent@example.com", "919191", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}
	svc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)

	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.VerifyCode(context.Background(), "concurrent@example.com", "919191")
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	successCount := 0
	for err := range results {
		if err == nil {
			successCount++
		}
	}
	if successCount != 1 {
		t.Fatalf("expected exactly one successful verification, got %d", successCount)
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

func TestVerifyCode_LockoutDeleteFailure_DoesNotFailOpen(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	baseRepo := NewInMemoryRepository(func() time.Time { return now })
	if err := baseRepo.SaveVerificationCode(context.Background(), "lockout-delete-fail@example.com", "121212", 10*time.Minute); err != nil {
		t.Fatalf("seed code: %v", err)
	}
	repo := &deleteFailOnLockoutRepository{InMemoryRepository: baseRepo}
	svc := NewCodeService(repo, &stubSender{}, nil, func() time.Time { return now }, nil)

	for i := 0; i < 4; i++ {
		_, err := svc.VerifyCode(context.Background(), "lockout-delete-fail@example.com", "000000")
		if ErrorCode(err) != AUTH_CODE_INVALID {
			t.Fatalf("attempt %d expected %s, got %q", i+1, AUTH_CODE_INVALID, ErrorCode(err))
		}
	}

	_, err := svc.VerifyCode(context.Background(), "lockout-delete-fail@example.com", "000000")
	if err == nil {
		t.Fatal("expected lockout delete failure to return error")
	}
	if ErrorCode(err) != "" {
		t.Fatalf("expected internal error without auth code, got %q", ErrorCode(err))
	}

	_, err = svc.VerifyCode(context.Background(), "lockout-delete-fail@example.com", "121212")
	if err == nil {
		t.Fatal("expected code to stay unusable after lockout delete failure")
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

func TestSendCode_IPCapReject_DoesNotConsumeEmailCooldown(t *testing.T) {
	now := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	sender := &stubSender{}
	svc := NewCodeService(repo, sender, nil, func() time.Time { return now }, func() string { return "343434" })

	for i := 0; i < svc.ipHourlyCap; i++ {
		email := fmt.Sprintf("burn-check-%d@example.com", i)
		if err := svc.SendCode(context.Background(), email, "198.51.100.10"); err != nil {
			t.Fatalf("seed attempt %d should succeed, got: %v", i+1, err)
		}
	}

	err := svc.SendCode(context.Background(), "target@example.com", "198.51.100.10")
	if ErrorCode(err) != AUTH_CODE_RATE_LIMIT {
		t.Fatalf("expected ip cap rejection %s, got %q", AUTH_CODE_RATE_LIMIT, ErrorCode(err))
	}

	err = svc.SendCode(context.Background(), "target@example.com", "198.51.100.11")
	if err != nil {
		t.Fatalf("expected target email to send from different IP, got: %v", err)
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

func TestOAuthExchangeCode_StoreConsumeSingleUse(t *testing.T) {
	now := time.Date(2026, 3, 20, 13, 20, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	pair := TokenPair{AccessToken: "access-token", RefreshToken: "refresh-token"}

	svc.StoreExchangeCode("exchange-1", pair)
	consumed, ok := svc.ConsumeExchangeCode("exchange-1")
	if !ok {
		t.Fatal("expected first exchange code consume to succeed")
	}
	if consumed != pair {
		t.Fatalf("expected consumed token pair %#v, got %#v", pair, consumed)
	}

	_, ok = svc.ConsumeExchangeCode("exchange-1")
	if ok {
		t.Fatal("expected exchange code to be single use")
	}
}

func TestOAuthExchangeCode_Expires(t *testing.T) {
	now := time.Date(2026, 3, 20, 13, 25, 0, 0, time.UTC)
	repo := NewInMemoryRepository(func() time.Time { return now })
	svc := NewOAuthService(repo, NewSessionService(repo, nil, func() time.Time { return now }), func() time.Time { return now })
	svc.StoreExchangeCode("exchange-expired", TokenPair{AccessToken: "access-token", RefreshToken: "refresh-token"})

	now = now.Add(11 * time.Minute)
	_, ok := svc.ConsumeExchangeCode("exchange-expired")
	if ok {
		t.Fatal("expected expired exchange code to fail consumption")
	}
}

type stubSender struct {
	err      error
	calls    int
	lastBody string
}

type consumeFailRepository struct {
	*InMemoryRepository
}

func (r *consumeFailRepository) VerifyAndConsumeCode(ctx context.Context, email string, codeDigest string, maxAttempts int) (VerifyConsumeResult, error) {
	return VerifyConsumeNone, errors.New("consume failed")
}

type deleteFailOnLockoutRepository struct {
	*InMemoryRepository
}

func (r *deleteFailOnLockoutRepository) DeleteVerificationCode(ctx context.Context, email string) error {
	return errors.New("delete failed")
}

func (s *stubSender) Send(to, subject, body string) error {
	s.calls++
	s.lastBody = body
	if s.err != nil {
		return s.err
	}
	return nil
}
