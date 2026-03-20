package accounting

import (
	"sort"
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
	ListByUser(userID string) ([]Ledger, error)
	GetByIDForUser(userID string, ledgerID string) (Ledger, bool, error)
	SaveByIDForUser(userID string, ledgerID string, ledger Ledger) (Ledger, bool, error)
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

func (r *InMemoryLedgerRepository) ListByUser(userID string) ([]Ledger, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ledgers := make([]Ledger, 0)
	for _, ledger := range r.ledgers {
		if ledger.UserID == userID {
			ledgers = append(ledgers, ledger)
		}
	}
	sort.Slice(ledgers, func(i, j int) bool {
		return ledgers[i].ID < ledgers[j].ID
	})
	return ledgers, nil
}

func (r *InMemoryLedgerRepository) SaveByIDForUser(userID string, ledgerID string, ledger Ledger) (Ledger, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.ledgers[ledgerID]
	if !ok || current.UserID != userID {
		return Ledger{}, false, nil
	}
	ledger.ID = current.ID
	ledger.UserID = current.UserID
	r.ledgers[ledgerID] = ledger
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
