package accounting

import (
	"sync"
)

type Ledger struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

type LedgerCreateInput struct {
	Name      string
	IsDefault bool
}

type LedgerRepository interface {
	Create(userID string, input LedgerCreateInput) (Ledger, error)
	GetByIDForUser(userID string, ledgerID string) (Ledger, bool, error)
	DeleteByIDForUser(userID string, ledgerID string) (bool, error)
}

type InMemoryLedgerRepository struct {
	mu      sync.Mutex
	ledgers map[string]Ledger
}

func NewInMemoryLedgerRepository() *InMemoryLedgerRepository {
	return &InMemoryLedgerRepository{ledgers: map[string]Ledger{}}
}

func (r *InMemoryLedgerRepository) Create(userID string, input LedgerCreateInput) (Ledger, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ledger := Ledger{
		ID:        nextID(),
		UserID:    userID,
		Name:      input.Name,
		IsDefault: input.IsDefault,
	}
	r.ledgers[ledger.ID] = ledger
	return ledger, nil
}

func (r *InMemoryLedgerRepository) GetByIDForUser(userID string, ledgerID string) (Ledger, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ledger, ok := r.ledgers[ledgerID]
	if !ok || ledger.UserID != userID {
		return Ledger{}, false, nil
	}
	return ledger, true, nil
}

func (r *InMemoryLedgerRepository) DeleteByIDForUser(userID string, ledgerID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ledger, ok := r.ledgers[ledgerID]
	if !ok || ledger.UserID != userID {
		return false, nil
	}
	delete(r.ledgers, ledgerID)
	return true, nil
}
