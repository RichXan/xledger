package accounting

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	TransactionTypeIncome   = "income"
	TransactionTypeExpense  = "expense"
	TransactionTypeTransfer = "transfer"
)

type Transaction struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	LedgerID      string     `json:"ledger_id"`
	AccountID     *string    `json:"account_id,omitempty"`
	FromAccountID *string    `json:"from_account_id,omitempty"`
	ToAccountID   *string    `json:"to_account_id,omitempty"`
	Type          string     `json:"type"`
	Amount        float64    `json:"amount"`
	OccurredAt    time.Time  `json:"occurred_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type TransactionCreateInput struct {
	LedgerID      string
	AccountID     *string
	FromAccountID *string
	ToAccountID   *string
	Type          string
	Amount        float64
	OccurredAt    time.Time
}

type TransactionTransferInput struct {
	LedgerID      string
	FromAccountID *string
	ToAccountID   *string
	Amount        float64
	OccurredAt    time.Time
}

type TransactionEditInput struct {
	Amount float64
}

type TransactionQuery struct {
	LedgerID string
}

type TransactionRepository interface {
	Create(userID string, input TransactionCreateInput) (Transaction, error)
	GetByIDForUser(userID string, txnID string) (Transaction, bool, error)
	SaveByIDForUser(userID string, txnID string, txn Transaction) (Transaction, bool, error)
	DeleteByIDForUser(userID string, txnID string) (bool, error)
	ListByUser(userID string, query TransactionQuery) ([]Transaction, error)
	MarkBalancesRecalculated(userID string, ledgerID string) error
	MarkStatsInputRecalculated(userID string, ledgerID string) error
}

type InMemoryTransactionRepository struct {
	mu                    sync.Mutex
	transactions          map[string]Transaction
	balanceRecalcCount    int
	statsInputRecalcCount int
}

func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{transactions: map[string]Transaction{}}
}

func (r *InMemoryTransactionRepository) Create(userID string, input TransactionCreateInput) (Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	txn := Transaction{
		ID:            nextID(),
		UserID:        userID,
		LedgerID:      input.LedgerID,
		AccountID:     cloneStringPtr(input.AccountID),
		FromAccountID: cloneStringPtr(input.FromAccountID),
		ToAccountID:   cloneStringPtr(input.ToAccountID),
		Type:          input.Type,
		Amount:        input.Amount,
		OccurredAt:    input.OccurredAt,
	}
	r.transactions[txn.ID] = txn
	return txn, nil
}

func (r *InMemoryTransactionRepository) GetByIDForUser(userID string, txnID string) (Transaction, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	txn, ok := r.transactions[txnID]
	if !ok || txn.UserID != userID {
		return Transaction{}, false, nil
	}
	return txn, true, nil
}

func (r *InMemoryTransactionRepository) SaveByIDForUser(userID string, txnID string, txn Transaction) (Transaction, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.transactions[txnID]
	if !ok || current.UserID != userID {
		return Transaction{}, false, nil
	}
	txn.ID = current.ID
	txn.UserID = current.UserID
	r.transactions[txnID] = txn
	return txn, true, nil
}

func (r *InMemoryTransactionRepository) DeleteByIDForUser(userID string, txnID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	txn, ok := r.transactions[txnID]
	if !ok || txn.UserID != userID {
		return false, nil
	}

	delete(r.transactions, txnID)
	return true, nil
}

func (r *InMemoryTransactionRepository) ListByUser(userID string, query TransactionQuery) ([]Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ledgerFilter := strings.TrimSpace(query.LedgerID)
	items := make([]Transaction, 0)
	for _, txn := range r.transactions {
		if txn.UserID != userID {
			continue
		}
		if ledgerFilter != "" && txn.LedgerID != ledgerFilter {
			continue
		}
		items = append(items, txn)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].OccurredAt.Before(items[j].OccurredAt)
	})
	return items, nil
}

func (r *InMemoryTransactionRepository) MarkBalancesRecalculated(_ string, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.balanceRecalcCount++
	return nil
}

func (r *InMemoryTransactionRepository) MarkStatsInputRecalculated(_ string, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statsInputRecalcCount++
	return nil
}

func (r *InMemoryTransactionRepository) BalanceRecalculationCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.balanceRecalcCount
}

func (r *InMemoryTransactionRepository) StatsInputRecalculationCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.statsInputRecalcCount
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
