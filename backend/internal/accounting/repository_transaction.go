package accounting

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	TransactionTypeIncome   = "income"
	TransactionTypeExpense  = "expense"
	TransactionTypeTransfer = "transfer"
	transferSideFrom        = "from"
	transferSideTo          = "to"
)

var (
	errTransferConflict          = errors.New("transfer pair conflict")
	errTransferVersionConflict   = errors.New("transfer version conflict")
	errTransferBilateralMismatch = errors.New("transfer bilateral mismatch")
	errTransferNotFound          = errors.New("transfer not found")
	errTransferInjectedFailure   = errors.New("transfer injected failure")
)

type Transaction struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	LedgerID       string     `json:"ledger_id"`
	AccountID      *string    `json:"account_id,omitempty"`
	FromAccountID  *string    `json:"from_account_id,omitempty"`
	ToAccountID    *string    `json:"to_account_id,omitempty"`
	TransferPairID *string    `json:"transfer_pair_id,omitempty"`
	TransferSide   string     `json:"transfer_side,omitempty"`
	Version        int        `json:"version"`
	Type           string     `json:"type"`
	Amount         float64    `json:"amount"`
	OccurredAt     time.Time  `json:"occurred_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
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
	FromLedgerID  *string
	ToLedgerID    *string
	FromAccountID *string
	ToAccountID   *string
	Amount        float64
	OccurredAt    time.Time
}

type TransactionEditInput struct {
	Amount  float64
	Version *int
}

type TransactionQuery struct {
	LedgerID string
}

type TransactionRepository interface {
	Create(userID string, input TransactionCreateInput) (Transaction, error)
	GetByIDForUser(userID string, txnID string) (Transaction, bool, error)
	SaveByIDForUser(userID string, txnID string, txn Transaction) (Transaction, bool, error)
	DeleteByIDForUser(userID string, txnID string) (bool, error)
	CreateTransferPair(userID string, pairID string, fromInput TransactionCreateInput, toInput TransactionCreateInput) (Transaction, error)
	GetTransferPairByTxnID(userID string, txnID string) ([]Transaction, error)
	UpdateTransferPairAmount(userID string, pairID string, amount float64, expectedVersion *int) (Transaction, error)
	DeleteTransferPairByTxnID(userID string, txnID string, expectedVersion *int) ([]string, error)
	WithTransferPairLock(userID string, pairID string, fn func() error) error
	ListByUser(userID string, query TransactionQuery) ([]Transaction, error)
	ListByTransferPairForUser(userID string, pairID string) ([]Transaction, error)
	MarkBalancesRecalculated(userID string, ledgerID string) error
	MarkStatsInputRecalculated(userID string, ledgerID string) error
}

type InMemoryTransactionRepository struct {
	mu                    sync.Mutex
	transactions          map[string]Transaction
	pairLocks             map[string]bool
	failCreateAfterFrom   bool
	failUpdateAfterFirst  bool
	failDeleteAfterFirst  bool
	balanceRecalcCount    int
	balanceRecalcByLedger map[string]int
	statsInputRecalcCount int
	statsInputByLedger    map[string]int
}

func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		transactions:          map[string]Transaction{},
		pairLocks:             map[string]bool{},
		balanceRecalcByLedger: map[string]int{},
		statsInputByLedger:    map[string]int{},
	}
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
		Version:       1,
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

func (r *InMemoryTransactionRepository) CreateTransferPair(userID string, pairID string, fromInput TransactionCreateInput, toInput TransactionCreateInput) (Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pairID = strings.TrimSpace(pairID)
	if pairID == "" {
		pairID = nextID()
	}

	fromTxn := Transaction{
		ID:             nextID(),
		UserID:         userID,
		LedgerID:       fromInput.LedgerID,
		FromAccountID:  cloneStringPtr(fromInput.FromAccountID),
		ToAccountID:    cloneStringPtr(fromInput.ToAccountID),
		TransferPairID: &pairID,
		TransferSide:   transferSideFrom,
		Version:        1,
		Type:           TransactionTypeTransfer,
		Amount:         fromInput.Amount,
		OccurredAt:     fromInput.OccurredAt,
	}

	toTxn := Transaction{
		ID:             nextID(),
		UserID:         userID,
		LedgerID:       toInput.LedgerID,
		FromAccountID:  cloneStringPtr(toInput.FromAccountID),
		ToAccountID:    cloneStringPtr(toInput.ToAccountID),
		TransferPairID: &pairID,
		TransferSide:   transferSideTo,
		Version:        1,
		Type:           TransactionTypeTransfer,
		Amount:         toInput.Amount,
		OccurredAt:     toInput.OccurredAt,
	}

	r.transactions[fromTxn.ID] = fromTxn
	if r.failCreateAfterFrom {
		r.failCreateAfterFrom = false
		delete(r.transactions, fromTxn.ID)
		return Transaction{}, errTransferInjectedFailure
	}
	r.transactions[toTxn.ID] = toTxn

	return fromTxn, nil
}

func (r *InMemoryTransactionRepository) GetTransferPairByTxnID(userID string, txnID string) ([]Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	txn, ok := r.transactions[txnID]
	if !ok || txn.UserID != userID {
		return nil, errTransferNotFound
	}

	pairID := strings.TrimSpace(ptrString(txn.TransferPairID))
	if pairID == "" {
		return nil, errTransferBilateralMismatch
	}

	pair := r.collectTransferPairLocked(userID, pairID)
	if err := validateTransferPair(pair); err != nil {
		return nil, err
	}

	return pair, nil
}

func (r *InMemoryTransactionRepository) UpdateTransferPairAmount(userID string, pairID string, amount float64, expectedVersion *int) (Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pair := r.collectTransferPairLocked(userID, pairID)
	if err := validateTransferPair(pair); err != nil {
		return Transaction{}, err
	}

	fromTxn := pair[0]
	if pair[1].TransferSide == transferSideFrom {
		fromTxn = pair[1]
	}

	for _, side := range pair {
		if expectedVersion != nil && side.Version != *expectedVersion {
			return Transaction{}, errTransferVersionConflict
		}
		original := side
		side.Amount = amount
		side.Version++
		r.transactions[side.ID] = side
		if r.failUpdateAfterFirst {
			r.failUpdateAfterFirst = false
			r.transactions[side.ID] = original
			return Transaction{}, errTransferInjectedFailure
		}
		if side.TransferSide == transferSideFrom {
			fromTxn = side
		}
	}

	return fromTxn, nil
}

func (r *InMemoryTransactionRepository) DeleteTransferPairByTxnID(userID string, txnID string, expectedVersion *int) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	txn, ok := r.transactions[txnID]
	if !ok || txn.UserID != userID {
		return nil, errTransferNotFound
	}

	pairID := strings.TrimSpace(ptrString(txn.TransferPairID))
	if pairID == "" {
		return nil, errTransferBilateralMismatch
	}

	pair := r.collectTransferPairLocked(userID, pairID)
	if err := validateTransferPair(pair); err != nil {
		return nil, err
	}

	ledgers := make([]string, 0, 2)
	for _, side := range pair {
		if expectedVersion != nil && side.Version != *expectedVersion {
			return nil, errTransferVersionConflict
		}
		ledgers = appendUnique(ledgers, side.LedgerID)
		original := side
		delete(r.transactions, side.ID)
		if r.failDeleteAfterFirst {
			r.failDeleteAfterFirst = false
			r.transactions[side.ID] = original
			return nil, errTransferInjectedFailure
		}
	}

	return ledgers, nil
}

func (r *InMemoryTransactionRepository) WithTransferPairLock(userID string, pairID string, fn func() error) error {
	key := strings.TrimSpace(userID) + "|" + strings.TrimSpace(pairID)

	r.mu.Lock()
	if r.pairLocks[key] {
		r.mu.Unlock()
		return errTransferConflict
	}
	r.pairLocks[key] = true
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		delete(r.pairLocks, key)
		r.mu.Unlock()
	}()

	return fn()
}

func (r *InMemoryTransactionRepository) ListByUser(userID string, query TransactionQuery) ([]Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ledgerFilter := strings.TrimSpace(query.LedgerID)
	items := make([]Transaction, 0)
	seenTransferPairs := map[string]bool{}
	for _, txn := range r.transactions {
		if txn.UserID != userID {
			continue
		}
		if ledgerFilter != "" && txn.LedgerID != ledgerFilter {
			continue
		}
		if txn.Type == TransactionTypeTransfer {
			pairID := strings.TrimSpace(ptrString(txn.TransferPairID))
			if pairID != "" {
				if ledgerFilter == "" {
					if txn.TransferSide != transferSideFrom {
						continue
					}
				} else {
					if txn.TransferSide != transferSideFrom && r.transferPairHasSideInLedgerLocked(userID, pairID, transferSideFrom, ledgerFilter) {
						continue
					}
				}
				key := pairID + "|" + ledgerFilter
				if seenTransferPairs[key] {
					continue
				}
				seenTransferPairs[key] = true
			}
		}
		items = append(items, txn)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].OccurredAt.Before(items[j].OccurredAt)
	})
	return items, nil
}

func (r *InMemoryTransactionRepository) ListByTransferPairForUser(userID string, pairID string) ([]Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pair := r.collectTransferPairLocked(userID, pairID)
	sort.Slice(pair, func(i, j int) bool {
		return pair[i].TransferSide < pair[j].TransferSide
	})
	return pair, nil
}

func (r *InMemoryTransactionRepository) MarkBalancesRecalculated(_ string, ledgerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.balanceRecalcCount++
	r.balanceRecalcByLedger[ledgerID]++
	return nil
}

func (r *InMemoryTransactionRepository) MarkStatsInputRecalculated(_ string, ledgerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statsInputRecalcCount++
	r.statsInputByLedger[ledgerID]++
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

func (r *InMemoryTransactionRepository) BalanceRecalculationCountForLedger(ledgerID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.balanceRecalcByLedger[ledgerID]
}

func (r *InMemoryTransactionRepository) StatsInputRecalculationCountForLedger(ledgerID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.statsInputByLedger[ledgerID]
}

func (r *InMemoryTransactionRepository) collectTransferPairLocked(userID string, pairID string) []Transaction {
	trimmedPairID := strings.TrimSpace(pairID)
	pair := make([]Transaction, 0, 2)
	for _, txn := range r.transactions {
		if txn.UserID != userID {
			continue
		}
		if strings.TrimSpace(ptrString(txn.TransferPairID)) != trimmedPairID {
			continue
		}
		pair = append(pair, txn)
	}
	return pair
}

func (r *InMemoryTransactionRepository) transferPairHasSideInLedgerLocked(userID string, pairID string, side string, ledgerID string) bool {
	for _, txn := range r.transactions {
		if txn.UserID != userID {
			continue
		}
		if strings.TrimSpace(ptrString(txn.TransferPairID)) != strings.TrimSpace(pairID) {
			continue
		}
		if txn.TransferSide != side {
			continue
		}
		if strings.TrimSpace(txn.LedgerID) != strings.TrimSpace(ledgerID) {
			continue
		}
		return true
	}
	return false
}

func validateTransferPair(pair []Transaction) error {
	if len(pair) == 0 {
		return errTransferNotFound
	}
	if len(pair) != 2 {
		return errTransferBilateralMismatch
	}

	fromCount := 0
	toCount := 0
	for _, side := range pair {
		switch side.TransferSide {
		case transferSideFrom:
			fromCount++
		case transferSideTo:
			toCount++
		default:
			return errTransferBilateralMismatch
		}
	}

	if fromCount != 1 || toCount != 1 {
		return errTransferBilateralMismatch
	}

	return nil
}

func appendUnique(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func (r *InMemoryTransactionRepository) InjectTransferCreateFailureAfterFrom() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failCreateAfterFrom = true
}

func (r *InMemoryTransactionRepository) InjectTransferUpdateFailureAfterFirst() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failUpdateAfterFirst = true
}

func (r *InMemoryTransactionRepository) InjectTransferDeleteFailureAfterFirst() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failDeleteAfterFirst = true
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
