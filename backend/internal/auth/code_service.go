package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	AUTH_CODE_INVALID     = "AUTH_CODE_INVALID"
	AUTH_CODE_EXPIRED     = "AUTH_CODE_EXPIRED"
	AUTH_CODE_SEND_FAILED = "AUTH_CODE_SEND_FAILED"
	AUTH_CODE_RATE_LIMIT  = "AUTH_CODE_RATE_LIMIT"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenIssuer interface {
	Issue(email string) (TokenPair, error)
}

type staticTokenIssuer struct{}

func (staticTokenIssuer) Issue(email string) (TokenPair, error) {
	normalized := strings.TrimSpace(strings.ToLower(email))
	if normalized == "" {
		return TokenPair{}, errors.New("email is required")
	}

	return TokenPair{
		AccessToken:  "access-" + normalized,
		RefreshToken: "refresh-" + normalized,
	}, nil
}

type authError struct {
	code string
	err  error
}

func (e *authError) Error() string {
	if e.err == nil {
		return e.code
	}

	return e.code + ": " + e.err.Error()
}

func (e *authError) Unwrap() error {
	return e.err
}

func ErrorCode(err error) string {
	if err == nil {
		return ""
	}

	var e *authError
	if errors.As(err, &e) {
		return e.code
	}

	return ""
}

type CodeService struct {
	repo           CodeRepository
	sender         SMTPSender
	issuer         TokenIssuer
	now            func() time.Time
	codeGenerator  func() string
	codeTTL        time.Duration
	resendInterval time.Duration
	ipHourlyCap    int
	ipWindow       time.Duration
	maxVerifyTries int
}

func NewCodeService(repo CodeRepository, sender SMTPSender, issuer TokenIssuer, now func() time.Time, codeGenerator func() string) *CodeService {
	if now == nil {
		now = time.Now
	}
	if issuer == nil {
		issuer = staticTokenIssuer{}
	}
	if codeGenerator == nil {
		codeGenerator = generateCode
	}

	return &CodeService{
		repo:           repo,
		sender:         sender,
		issuer:         issuer,
		now:            now,
		codeGenerator:  codeGenerator,
		codeTTL:        10 * time.Minute,
		resendInterval: 60 * time.Second,
		ipHourlyCap:    10,
		ipWindow:       time.Hour,
		maxVerifyTries: 5,
	}
}

func (s *CodeService) SendCode(ctx context.Context, email string, clientIP string) error {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if normalizedEmail == "" {
		return &authError{code: AUTH_CODE_INVALID, err: errors.New("email is required")}
	}

	allowed, err := s.repo.AcquireSendLock(ctx, normalizedEmail, s.now(), s.resendInterval)
	if err != nil {
		return fmt.Errorf("check resend limit: %w", err)
	}
	if !allowed {
		return &authError{code: AUTH_CODE_RATE_LIMIT}
	}

	ipAllowed, ipErr := s.repo.AcquireIPHourlySlot(ctx, strings.TrimSpace(clientIP), s.now(), s.ipWindow, s.ipHourlyCap)
	if ipErr != nil {
		_ = s.repo.ReleaseSendLock(ctx, normalizedEmail)
		return fmt.Errorf("check ip hourly limit: %w", ipErr)
	}
	if !ipAllowed {
		_ = s.repo.ReleaseSendLock(ctx, normalizedEmail)
		return &authError{code: AUTH_CODE_RATE_LIMIT}
	}

	code := s.codeGenerator()
	if !isSixDigitNumeric(code) {
		code = generateCode()
	}

	if err := s.repo.SaveVerificationCode(ctx, normalizedEmail, code, s.codeTTL); err != nil {
		_ = s.repo.ReleaseSendLock(ctx, normalizedEmail)
		return fmt.Errorf("save verification code: %w", err)
	}

	if err := s.sender.Send(normalizedEmail, "Your verification code", "Your XLedger verification code is: "+code); err != nil {
		_ = s.repo.DeleteVerificationCode(ctx, normalizedEmail)
		return &authError{code: AUTH_CODE_SEND_FAILED, err: err}
	}

	return nil
}

func (s *CodeService) VerifyCode(ctx context.Context, email string, code string) (TokenPair, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	providedCode := strings.TrimSpace(code)
	providedCodeDigest := hashVerificationCode(providedCode)
	consumeResult, err := s.repo.VerifyAndConsumeCode(ctx, normalizedEmail, providedCodeDigest)
	if err != nil {
		return TokenPair{}, fmt.Errorf("verify and consume code: %w", err)
	}

	switch consumeResult {
	case VerifyConsumeExpired:
		return TokenPair{}, &authError{code: AUTH_CODE_EXPIRED, err: ErrCodeExpired}
	case VerifyConsumeMismatch:
		attempts, attemptErr := s.repo.RecordFailedVerificationAttempt(ctx, normalizedEmail)
		if attemptErr != nil {
			return TokenPair{}, fmt.Errorf("record verification attempt: %w", attemptErr)
		}
		if attempts >= s.maxVerifyTries {
			_ = s.repo.DeleteVerificationCode(ctx, normalizedEmail)
		}
		return TokenPair{}, &authError{code: AUTH_CODE_INVALID, err: errors.New("code mismatch")}
	case VerifyConsumeNone:
		return TokenPair{}, &authError{code: AUTH_CODE_INVALID, err: ErrCodeNotFound}
	case VerifyConsumeMatch:
	default:
		return TokenPair{}, fmt.Errorf("unsupported consume result: %s", consumeResult)
	}

	tokens, err := s.issuer.Issue(normalizedEmail)
	if err != nil {
		return TokenPair{}, fmt.Errorf("issue tokens: %w", err)
	}

	if err := s.repo.CreateSession(ctx, normalizedEmail); err != nil {
		return TokenPair{}, fmt.Errorf("create session: %w", err)
	}

	return tokens, nil
}

func hashVerificationCode(code string) string {
	mac := hmac.New(sha256.New, []byte(verificationCodePepper()))
	_, _ = mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

func secureCodeEqual(left string, right string) bool {
	if len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func verificationCodePepper() string {
	pepper := strings.TrimSpace(os.Getenv("AUTH_CODE_PEPPER"))
	if pepper == "" {
		return "dev-task2-otp-pepper"
	}

	return pepper
}

func generateCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "000000"
	}

	code := strconv.FormatInt(n.Int64(), 10)
	for len(code) < 6 {
		code = "0" + code
	}

	return code
}

func isSixDigitNumeric(value string) bool {
	if len(value) != 6 {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
