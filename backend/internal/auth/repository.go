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
	VerifyAndConsumeCode(ctx context.Context, email string, codeDigest string, maxAttempts int) (VerifyConsumeResult, error)
	DeleteVerificationCode(ctx context.Context, email string) error
	RecordFailedVerificationAttempt(ctx context.Context, email string) (int, error)
	AcquireIPHourlySlot(ctx context.Context, ip string, at time.Time, ttl time.Duration, cap int) (bool, error)
	AcquireSendLock(ctx context.Context, email string, at time.Time, ttl time.Duration) (bool, error)
	ReleaseSendLock(ctx context.Context, email string) error
	CreateSession(ctx context.Context, email string) error
	SaveOAuthStateNonce(ctx context.Context, state string, nonce string, ttl time.Duration) error
	SaveOAuthStateNonceForEmail(ctx context.Context, state string, nonce string, email string, ttl time.Duration) error
	ConsumeOAuthStateNonce(ctx context.Context, state string, nonce string) (bool, error)
	ConsumeOAuthStateNonceForEmail(ctx context.Context, state string, nonce string) (string, bool, error)
	StoreRefreshToken(ctx context.Context, tokenID string, email string, expiresAt time.Time) error
	ConsumeRefreshToken(ctx context.Context, tokenID string) (RefreshSession, bool, error)
	BlacklistRefreshToken(ctx context.Context, tokenID string, at time.Time) error
	IsRefreshTokenBlacklisted(ctx context.Context, tokenID string) (bool, time.Duration, error)
	RecordAlertEvent(ctx context.Context, event string) error
	EnsureDefaultLedger(ctx context.Context, email string) (bool, error)
}

type RefreshSession struct {
	Email     string
	ExpiresAt time.Time
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

type inMemoryOAuthState struct {
	nonce     string
	email     string
	expiresAt time.Time
}

type inMemoryRefreshSession struct {
	email     string
	expiresAt time.Time
	consumed  bool
}

type InMemoryRepository struct {
	mu         sync.Mutex
	now        func() time.Time
	codes      map[string]inMemoryCode
	lastSent   map[string]inMemoryTimestamp
	verifyFail map[string]inMemoryCounter
	ipHourly   map[string]inMemoryIPCounter
	oauthState map[string]inMemoryOAuthState
	refresh    map[string]inMemoryRefreshSession
	blacklist  map[string]time.Time
	alerts     map[string]int
	ledgers    map[string]int
	forcedLag  time.Duration
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
		oauthState: make(map[string]inMemoryOAuthState),
		refresh:    make(map[string]inMemoryRefreshSession),
		blacklist:  make(map[string]time.Time),
		alerts:     make(map[string]int),
		ledgers:    make(map[string]int),
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

func (r *InMemoryRepository) VerifyAndConsumeCode(_ context.Context, email string, codeDigest string, maxAttempts int) (VerifyConsumeResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if failRecord, ok := r.verifyFail[email]; ok {
		if r.now().Before(failRecord.expiresAt) && failRecord.value >= maxAttempts {
			return VerifyConsumeMismatch, nil
		}
	}

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

func (r *InMemoryRepository) SaveOAuthStateNonce(_ context.Context, state string, nonce string, ttl time.Duration) error {
	return r.SaveOAuthStateNonceForEmail(context.Background(), state, nonce, "", ttl)
}

func (r *InMemoryRepository) SaveOAuthStateNonceForEmail(_ context.Context, state string, nonce string, email string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	r.oauthState[state] = inMemoryOAuthState{nonce: nonce, email: email, expiresAt: r.now().Add(ttl)}
	return nil
}

func (r *InMemoryRepository) ConsumeOAuthStateNonce(_ context.Context, state string, nonce string) (bool, error) {
	_, ok, err := r.ConsumeOAuthStateNonceForEmail(context.Background(), state, nonce)
	return ok, err
}

func (r *InMemoryRepository) ConsumeOAuthStateNonceForEmail(_ context.Context, state string, nonce string) (string, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	record, ok := r.oauthState[state]
	if !ok {
		return "", false, nil
	}
	delete(r.oauthState, state)
	if !secureCodeEqual(record.nonce, nonce) {
		return "", false, nil
	}
	return record.email, true, nil
}

func (r *InMemoryRepository) StoreRefreshToken(_ context.Context, tokenID string, email string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	r.refresh[tokenID] = inMemoryRefreshSession{email: email, expiresAt: expiresAt}
	return nil
}

func (r *InMemoryRepository) ConsumeRefreshToken(_ context.Context, tokenID string) (RefreshSession, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	record, ok := r.refresh[tokenID]
	if !ok {
		return RefreshSession{}, false, nil
	}
	if record.consumed {
		return RefreshSession{}, false, nil
	}
	if !r.now().Before(record.expiresAt) {
		delete(r.refresh, tokenID)
		return RefreshSession{}, false, nil
	}
	record.consumed = true
	r.refresh[tokenID] = record
	return RefreshSession{Email: record.email, ExpiresAt: record.expiresAt}, true, nil
}

func (r *InMemoryRepository) BlacklistRefreshToken(_ context.Context, tokenID string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	r.blacklist[tokenID] = at
	if session, ok := r.refresh[tokenID]; ok {
		session.consumed = true
		r.refresh[tokenID] = session
	}
	return nil
}

func (r *InMemoryRepository) IsRefreshTokenBlacklisted(_ context.Context, tokenID string) (bool, time.Duration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	if _, ok := r.blacklist[tokenID]; ok {
		return true, 0, nil
	}
	if r.forcedLag > 0 {
		return false, r.forcedLag, nil
	}
	return false, 0, nil
}

func (r *InMemoryRepository) RecordAlertEvent(_ context.Context, event string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	r.alerts[event]++
	return nil
}

func (r *InMemoryRepository) EnsureDefaultLedger(_ context.Context, email string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	if r.ledgers[email] > 0 {
		return false, nil
	}
	r.ledgers[email] = 1
	return true, nil
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

func (r *InMemoryRepository) RefreshBlacklistTime(tokenID string) (time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	at, ok := r.blacklist[tokenID]
	return at, ok
}

func (r *InMemoryRepository) SetForcedBlacklistLag(lag time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.forcedLag = lag
}

func (r *InMemoryRepository) AlertEventCount(event string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	return r.alerts[event]
}

func (r *InMemoryRepository) DefaultLedgerCount(email string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cleanupExpiredLocked(r.now())

	return r.ledgers[email]
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
	for state, oauth := range r.oauthState {
		if !now.Before(oauth.expiresAt) {
			delete(r.oauthState, state)
		}
	}
	for tokenID, session := range r.refresh {
		if !now.Before(session.expiresAt) {
			delete(r.refresh, tokenID)
		}
	}
}
