package portability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	PAT_EXPIRED           = "PAT_EXPIRED"
	PAT_REVOKED           = "PAT_REVOKED"
	PAT_FORBIDDEN_ON_AUTH = "PAT_FORBIDDEN_ON_AUTH"
	defaultPATTTL         = 90 * 24 * time.Hour
	defaultPATPrefix      = "pat:"
	maxExpiryYear         = 10
)

type PATRecord struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	TokenHash string     `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type PATService struct {
	mu            sync.Mutex
	now           func() time.Time
	items         map[string]PATRecord
	tokenToID     map[string]string
	tokenToUser   map[string]string
	revokedTokens map[string]bool // immediate revocation blacklist, checked before lag
	strictMode    bool
	alertEvents   []string
	revocationLag time.Duration
}

func NewPATService(now func() time.Time) *PATService {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &PATService{now: now, items: map[string]PATRecord{}, tokenToID: map[string]string{}, tokenToUser: map[string]string{}, revokedTokens: map[string]bool{}}
}

func (s *PATService) SetNow(now func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if now != nil {
		s.now = now
	}
}

func (s *PATService) SetRevocationLag(lag time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revocationLag = lag
}

func (s *PATService) StrictModeEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.strictMode
}

func (s *PATService) AlertEvents() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.alertEvents...)
}

func (s *PATService) CreatePAT(_ context.Context, userID string, name string, expiresAt *time.Time) (string, PATRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	plain := defaultPATPrefix + strings.TrimSpace(userID) + ":" + nextPATID()

	var expiresAtPtr *time.Time
	if expiresAt != nil {
		t := expiresAt.UTC()
		maxExpiry := now.Add(maxExpiryYear * 365 * 24 * time.Hour)
		if !t.After(maxExpiry) {
			expiresAtPtr = &t
		}
		// If nil or exceeds max, leave as nil (no expiry)
	} else {
		// Default TTL: 90 days
		defaultExpiry := now.Add(defaultPATTTL)
		expiresAtPtr = &defaultExpiry
	}

	record := PATRecord{
		ID:        nextPATID(),
		Name:      strings.TrimSpace(name),
		TokenHash: hashPAT(plain),
		CreatedAt: now,
		ExpiresAt: expiresAtPtr,
	}
	s.items[record.ID] = record
	s.tokenToID[record.TokenHash] = record.ID
	s.tokenToUser[record.TokenHash] = strings.TrimSpace(userID)
	return plain, record, nil
}

func (s *PATService) RevokePAT(_ context.Context, userID string, patID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.items[patID]
	if !ok {
		return &contractError{code: PAT_REVOKED}
	}
	// Validate that the PAT belongs to the user requesting revocation
	if s.tokenToUser[record.TokenHash] != strings.TrimSpace(userID) {
		return &contractError{code: PAT_REVOKED}
	}
	now := s.now()
	record.RevokedAt = &now
	s.items[patID] = record
	// Add to immediate revocation blacklist
	s.revokedTokens[record.TokenHash] = true
	return nil
}

func (s *PATService) ValidatePAT(_ context.Context, token string, path string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.CanUsePATOnPath(token, path) {
		return "", &contractError{code: PAT_FORBIDDEN_ON_AUTH}
	}
	tokenHash := hashPAT(token)
	// Check immediate revocation blacklist first
	if s.revokedTokens[tokenHash] {
		// Still detect breach even on immediate rejection
		if id, ok := s.tokenToID[tokenHash]; ok {
			if record, ok := s.items[id]; ok && record.RevokedAt != nil {
				if s.revocationLag > 5*time.Second {
					s.strictMode = true
					s.alertEvents = append(s.alertEvents, "pat.revoke.blacklist_sla_exceeded")
				}
			}
		}
		return "", &contractError{code: PAT_REVOKED}
	}
	id, ok := s.tokenToID[tokenHash]
	if !ok {
		return "", &contractError{code: PAT_REVOKED}
	}
	record := s.items[id]
	now := s.now()
	if record.ExpiresAt != nil && now.After(*record.ExpiresAt) {
		return "", &contractError{code: PAT_EXPIRED}
	}
	if record.RevokedAt != nil {
		if s.revocationLag > 5*time.Second {
			s.strictMode = true
			s.alertEvents = append(s.alertEvents, "pat.revoke.blacklist_sla_exceeded")
		}
		return "", &contractError{code: PAT_REVOKED}
	}
	parts := strings.Split(strings.TrimPrefix(token, defaultPATPrefix), ":")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "", &contractError{code: PAT_REVOKED}
	}
	return strings.TrimSpace(parts[0]), nil
}

func (s *PATService) CanUsePATOnPath(token string, path string) bool {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(token)), defaultPATPrefix) {
		return true
	}
	path = strings.TrimSpace(strings.ToLower(path))
	if strings.HasPrefix(path, "/api/auth") {
		return false
	}
	if strings.HasPrefix(path, "/api/personal-access-tokens") {
		return false
	}
	return true
}

func (s *PATService) ListPATs(_ context.Context, userID string) []PATRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]PATRecord, 0)
	for _, record := range s.items {
		if s.tokenToUser[record.TokenHash] != strings.TrimSpace(userID) {
			continue
		}
		copy := record
		copy.TokenHash = ""
		items = append(items, copy)
	}
	return items
}

var patIDCounter uint64

func hashPAT(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func nextPATID() string {
	value := atomic.AddUint64(&patIDCounter, 1)
	return "pat-" + strconv.FormatUint(value, 10)
}
