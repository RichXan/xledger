package accounting

import (
	"sort"
	"strconv"
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
	ListByUser(userID string) ([]Account, error)
	GetByIDForUser(userID string, accountID string) (Account, bool, error)
	SaveByIDForUser(userID string, accountID string, account Account) (Account, bool, error)
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
		Name:           input.Name,
		Type:           input.Type,
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

func (r *InMemoryAccountRepository) ListByUser(userID string) ([]Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	accounts := make([]Account, 0)
	for _, account := range r.accounts {
		if account.UserID == userID {
			accounts = append(accounts, account)
		}
	}
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].ID < accounts[j].ID
	})
	return accounts, nil
}

func (r *InMemoryAccountRepository) SaveByIDForUser(userID string, accountID string, account Account) (Account, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.accounts[accountID]
	if !ok || current.UserID != userID {
		return Account{}, false, nil
	}
	account.ID = current.ID
	account.UserID = current.UserID
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
