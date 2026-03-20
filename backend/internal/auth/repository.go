package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrCodeNotFound = errors.New("verification code not found")
	ErrCodeExpired  = errors.New("verification code expired")
)

type CodeRepository interface {
	SaveVerificationCode(ctx context.Context, email string, code string, ttl time.Duration) error
	GetVerificationCode(ctx context.Context, email string) (string, error)
	VerifyAndConsumeCode(ctx context.Context, email string, codeDigest string) (VerifyConsumeResult, error)
	DeleteVerificationCode(ctx context.Context, email string) error
	RecordFailedVerificationAttempt(ctx context.Context, email string) (int, error)
	AcquireIPHourlySlot(ctx context.Context, ip string, at time.Time, ttl time.Duration, cap int) (bool, error)
	AcquireSendLock(ctx context.Context, email string, at time.Time, ttl time.Duration) (bool, error)
	ReleaseSendLock(ctx context.Context, email string) error
	CreateSession(ctx context.Context, email string) error
}

type VerifyConsumeResult string

const (
	VerifyConsumeNone     VerifyConsumeResult = "none"
	VerifyConsumeMatch    VerifyConsumeResult = "match"
	VerifyConsumeMismatch VerifyConsumeResult = "mismatch"
	VerifyConsumeExpired  VerifyConsumeResult = "expired"
)

type inMemoryCode struct {
	value     string
	expiresAt time.Time
}

type inMemoryTimestamp struct {
	value     time.Time
	expiresAt time.Time
}

type inMemoryCounter struct {
	value     int
	expiresAt time.Time
}

type inMemoryIPCounter struct {
	value     int
	expiresAt time.Time
}

type InMemoryRepository struct {
	mu         sync.Mutex
	now        func() time.Time
	codes      map[string]inMemoryCode
	lastSent   map[string]inMemoryTimestamp
	verifyFail map[string]inMemoryCounter
	ipHourly   map[string]inMemoryIPCounter
	sessionCnt int
}

func NewInMemoryRepository(now func() time.Time) *InMemoryRepository {
	if now == nil {
		now = time.Now
	}

	return &InMemoryRepository{
		now:        now,
		codes:      make(map[string]inMemoryCode),
		lastSent:   make(map[string]inMemoryTimestamp),
		verifyFail: make(map[string]inMemoryCounter),
		ipHourly:   make(map[string]inMemoryIPCounter),
	}
}

func (r *InMemoryRepository) SaveVerificationCode(_ context.Context, email string, code string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	expiresAt := r.now().Add(ttl)
	r.codes[email] = inMemoryCode{value: hashVerificationCode(code), expiresAt: expiresAt}
	r.verifyFail[email] = inMemoryCounter{value: 0, expiresAt: expiresAt}
	return nil
}

func (r *InMemoryRepository) GetVerificationCode(_ context.Context, email string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, ok := r.codes[email]
	if !ok {
		return "", ErrCodeNotFound
	}

	if !r.now().Before(record.expiresAt) {
		delete(r.codes, email)
		delete(r.verifyFail, email)
		return "", ErrCodeExpired
	}

	return record.value, nil
}

func (r *InMemoryRepository) VerifyAndConsumeCode(_ context.Context, email string, codeDigest string) (VerifyConsumeResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, ok := r.codes[email]
	if !ok {
		return VerifyConsumeNone, nil
	}
	if !r.now().Before(record.expiresAt) {
		delete(r.codes, email)
		delete(r.verifyFail, email)
		return VerifyConsumeExpired, nil
	}
	if !secureCodeEqual(record.value, codeDigest) {
		return VerifyConsumeMismatch, nil
	}

	delete(r.codes, email)
	delete(r.verifyFail, email)
	return VerifyConsumeMatch, nil
}

func (r *InMemoryRepository) DeleteVerificationCode(_ context.Context, email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	delete(r.codes, email)
	delete(r.verifyFail, email)
	return nil
}

func (r *InMemoryRepository) RecordFailedVerificationAttempt(_ context.Context, email string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	counter, ok := r.verifyFail[email]
	if !ok {
		expiresAt := r.now().Add(10 * time.Minute)
		if code, codeOK := r.codes[email]; codeOK {
			expiresAt = code.expiresAt
		}
		counter = inMemoryCounter{value: 0, expiresAt: expiresAt}
	}

	counter.value++
	r.verifyFail[email] = counter

	return counter.value, nil
}

func (r *InMemoryRepository) AcquireSendLock(_ context.Context, email string, at time.Time, ttl time.Duration) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	record, ok := r.lastSent[email]
	if ok && r.now().Before(record.expiresAt) {
		return false, nil
	}

	r.lastSent[email] = inMemoryTimestamp{value: at, expiresAt: at.Add(ttl)}
	return true, nil
}

func (r *InMemoryRepository) AcquireIPHourlySlot(_ context.Context, ip string, at time.Time, ttl time.Duration, cap int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	key := ip
	if key == "" {
		key = "unknown"
	}

	record, ok := r.ipHourly[key]
	if !ok || !r.now().Before(record.expiresAt) {
		r.ipHourly[key] = inMemoryIPCounter{value: 1, expiresAt: at.Add(ttl)}
		return true, nil
	}

	if record.value >= cap {
		return false, nil
	}

	record.value++
	r.ipHourly[key] = record
	return true, nil
}

func (r *InMemoryRepository) ReleaseSendLock(_ context.Context, email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	delete(r.lastSent, email)
	return nil
}

func (r *InMemoryRepository) CreateSession(_ context.Context, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	r.sessionCnt++
	return nil
}

func (r *InMemoryRepository) StoredCode(email string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	record, ok := r.codes[email]
	if !ok {
		return ""
	}
	if !r.now().Before(record.expiresAt) {
		return ""
	}

	return record.value
}

func (r *InMemoryRepository) SessionCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	return r.sessionCnt
}

func (r *InMemoryRepository) stateCounts() (int, int, int, int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	return len(r.codes), len(r.lastSent), len(r.verifyFail), len(r.ipHourly)
}

func (r *InMemoryRepository) cleanupExpiredLocked(now time.Time) {
	for email, code := range r.codes {
		if !now.Before(code.expiresAt) {
			delete(r.codes, email)
			delete(r.verifyFail, email)
		}
	}
	for email, sent := range r.lastSent {
		if !now.Before(sent.expiresAt) {
			delete(r.lastSent, email)
		}
	}
	for email, fail := range r.verifyFail {
		if !now.Before(fail.expiresAt) {
			delete(r.verifyFail, email)
		}
	}
	for ip, slot := range r.ipHourly {
		if !now.Before(slot.expiresAt) {
			delete(r.ipHourly, ip)
		}
	}
}
