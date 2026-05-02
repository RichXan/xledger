package portability

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Repository struct {
	mu           sync.Mutex
	now          func() time.Time
	jobs         map[string]importJob
	rows         map[string][]storedImportRow
	transactions map[string]bool
}

type importJob struct {
	UserID         string
	Path           string
	IdempotencyKey string
	CreatedAt      time.Time
	Response       ImportConfirmResponse
	ErrorCode      string
}

type storedImportRow struct {
	Date        string
	Amount      float64
	Description string
	Type        string
}

func NewRepository(now func() time.Time) *Repository {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Repository{
		now:          now,
		jobs:         map[string]importJob{},
		rows:         map[string][]storedImportRow{},
		transactions: map[string]bool{},
	}
}

func (r *Repository) SetNow(now func() time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if now != nil {
		r.now = now
	}
}

func (r *Repository) FindJob(userID string, path string, idempotencyKey string) (importJob, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := r.jobKey(userID, path, idempotencyKey)
	job, ok := r.jobs[key]
	if !ok {
		return importJob{}, false
	}
	if r.now().Sub(job.CreatedAt) >= 24*time.Hour {
		delete(r.jobs, key)
		return importJob{}, false
	}
	return job, true
}

func (r *Repository) SaveJob(job importJob) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[r.jobKey(job.UserID, job.Path, job.IdempotencyKey)] = job
}

func (r *Repository) HasTriple(userID string, row storedImportRow) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.transactions[r.transactionKey(userID, row)]
}

func (r *Repository) SaveRow(userID string, row storedImportRow) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[userID] = append(r.rows[userID], row)
}

func (r *Repository) StoredRowCount(userID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.rows[userID])
}

func (r *Repository) SaveImportedTransaction(userID string, row ImportRow) error {
	trimmedDate := strings.TrimSpace(row.Date)
	trimmedDesc := strings.TrimSpace(row.Description)
	if trimmedDate == "" || trimmedDesc == "" || row.Amount <= 0 || math.IsNaN(row.Amount) || math.IsInf(row.Amount, 0) {
		return errors.New("invalid import row")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := storedImportRow{Date: trimmedDate, Amount: row.Amount, Description: trimmedDesc, Type: normalizeImportTransactionType(row.Type)}
	r.transactions[r.transactionKey(userID, stored)] = true
	return nil
}

func (r *Repository) Now() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.now()
}

func (r *Repository) jobKey(userID string, path string, key string) string {
	return strings.TrimSpace(userID) + "|" + strings.TrimSpace(path) + "|" + strings.TrimSpace(key)
}

func (r *Repository) transactionKey(userID string, row storedImportRow) string {
	return strings.TrimSpace(userID) + "|" + normalizeImportTransactionType(row.Type) + "|" + strings.TrimSpace(row.Date) + "|" + strconv.FormatFloat(math.Abs(row.Amount), 'f', 2, 64) + "|" + strings.TrimSpace(strings.ToLower(row.Description))
}
