package accounting

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Account struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	InitialBalance float64    `json:"initial_balance"`
	ArchivedAt     *time.Time `json:"archived_at,omitempty"`
}

type AccountCreateInput struct {
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	InitialBalance float64 `json:"initial_balance"`
}

type AccountUpdateInput struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	HasName    bool
	HasType    bool
	HasArchive bool
	Archive    bool
}

type AccountRepository interface {
	Create(userID string, input AccountCreateInput) (Account, error)
	GetByIDForUser(userID string, accountID string) (Account, bool, error)
	UpdateByIDForUser(userID string, accountID string, input AccountUpdateInput) (Account, bool, error)
	DeleteByIDForUser(userID string, accountID string) (bool, error)
}

type InMemoryAccountRepository struct {
	mu       sync.Mutex
	accounts map[string]Account
}

var globalIDCounter int64

func nextID() string {
	value := atomic.AddInt64(&globalIDCounter, 1)
	return "id-" + strconv.FormatInt(value, 10)
}

func NewInMemoryAccountRepository() *InMemoryAccountRepository {
	return &InMemoryAccountRepository{accounts: map[string]Account{}}
}

func (r *InMemoryAccountRepository) Create(userID string, input AccountCreateInput) (Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	account := Account{
		ID:             nextID(),
		UserID:         userID,
		Name:           strings.TrimSpace(input.Name),
		Type:           strings.TrimSpace(input.Type),
		InitialBalance: input.InitialBalance,
	}
	r.accounts[account.ID] = account
	return account, nil
}

func (r *InMemoryAccountRepository) GetByIDForUser(userID string, accountID string) (Account, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	account, ok := r.accounts[accountID]
	if !ok || account.UserID != userID {
		return Account{}, false, nil
	}
	return account, true, nil
}

func (r *InMemoryAccountRepository) UpdateByIDForUser(userID string, accountID string, input AccountUpdateInput) (Account, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	account, ok := r.accounts[accountID]
	if !ok || account.UserID != userID {
		return Account{}, false, nil
	}

	if input.HasName {
		account.Name = strings.TrimSpace(input.Name)
	}
	if input.HasType {
		account.Type = strings.TrimSpace(input.Type)
	}
	if input.HasArchive {
		if input.Archive {
			now := time.Now().UTC()
			account.ArchivedAt = &now
		} else {
			account.ArchivedAt = nil
		}
	}

	r.accounts[accountID] = account
	return account, true, nil
}

func (r *InMemoryAccountRepository) DeleteByIDForUser(userID string, accountID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	account, ok := r.accounts[accountID]
	if !ok || account.UserID != userID {
		return false, nil
	}

	delete(r.accounts, accountID)
	return true, nil
}
