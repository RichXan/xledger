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
	DeleteVerificationCode(ctx context.Context, email string) error
	GetLastCodeSentAt(ctx context.Context, email string) (time.Time, bool, error)
	SetLastCodeSentAt(ctx context.Context, email string, at time.Time, ttl time.Duration) error
	CreateSession(ctx context.Context, email string) error
}

type inMemoryCode struct {
	value     string
	expiresAt time.Time
}

type inMemoryTimestamp struct {
	value     time.Time
	expiresAt time.Time
}

type InMemoryRepository struct {
	mu         sync.Mutex
	now        func() time.Time
	codes      map[string]inMemoryCode
	lastSent   map[string]inMemoryTimestamp
	sessionCnt int
}

func NewInMemoryRepository(now func() time.Time) *InMemoryRepository {
	if now == nil {
		now = time.Now
	}

	return &InMemoryRepository{
		now:      now,
		codes:    make(map[string]inMemoryCode),
		lastSent: make(map[string]inMemoryTimestamp),
	}
}

func (r *InMemoryRepository) SaveVerificationCode(_ context.Context, email string, code string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.codes[email] = inMemoryCode{value: code, expiresAt: r.now().Add(ttl)}
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
		return "", ErrCodeExpired
	}

	return record.value, nil
}

func (r *InMemoryRepository) DeleteVerificationCode(_ context.Context, email string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.codes, email)
	return nil
}

func (r *InMemoryRepository) GetLastCodeSentAt(_ context.Context, email string) (time.Time, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	record, ok := r.lastSent[email]
	if !ok {
		return time.Time{}, false, nil
	}

	if !r.now().Before(record.expiresAt) {
		delete(r.lastSent, email)
		return time.Time{}, false, nil
	}

	return record.value, true, nil
}

func (r *InMemoryRepository) SetLastCodeSentAt(_ context.Context, email string, at time.Time, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastSent[email] = inMemoryTimestamp{value: at, expiresAt: at.Add(ttl)}
	return nil
}

func (r *InMemoryRepository) CreateSession(_ context.Context, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessionCnt++
	return nil
}

func (r *InMemoryRepository) StoredCode(email string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

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

	return r.sessionCnt
}
