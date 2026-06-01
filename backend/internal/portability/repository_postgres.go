package portability

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) FindJob(userID string, path string, idempotencyKey string) (importJob, bool) {
	var job importJob
	var responseJSON string
	err := r.db.QueryRow(`
		SELECT user_id, path, idempotency_key, created_at, response_json, COALESCE(error_code, ''),
			COALESCE(status, ''), COALESCE(total_rows, 0), COALESCE(processed_rows, 0)
		FROM import_jobs
		WHERE user_id = $1 AND path = $2 AND idempotency_key = $3
		AND created_at > NOW() - INTERVAL '24 hours'
	`, userID, path, idempotencyKey).Scan(
		&job.UserID, &job.Path, &job.IdempotencyKey, &job.CreatedAt, &responseJSON, &job.ErrorCode,
		&job.Status, &job.TotalRows, &job.ProcessedRows,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return importJob{}, false
	}
	if err != nil {
		return importJob{}, false
	}
	if strings.TrimSpace(responseJSON) != "" {
		if err := json.Unmarshal([]byte(responseJSON), &job.Response); err != nil {
			return importJob{}, false
		}
	}
	return job, true
}

func (r *PostgresRepository) SaveJob(job importJob) {
	responseJSON, err := json.Marshal(job.Response)
	if err != nil {
		responseJSON = []byte("{}")
	}
	r.db.Exec(`
		INSERT INTO import_jobs (user_id, path, idempotency_key, created_at, response_json, error_code, status, total_rows, processed_rows)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, path, idempotency_key) DO UPDATE SET
			response_json = EXCLUDED.response_json,
			error_code = EXCLUDED.error_code,
			created_at = EXCLUDED.created_at,
			status = EXCLUDED.status,
			total_rows = EXCLUDED.total_rows,
			processed_rows = EXCLUDED.processed_rows
	`, job.UserID, job.Path, job.IdempotencyKey, job.CreatedAt, string(responseJSON), job.ErrorCode, job.Status, job.TotalRows, job.ProcessedRows)
}

func (r *PostgresRepository) HasTriple(userID string, row storedImportRow) bool {
	var exists bool
	occurredAt, err := parseImportOccurredAt(row.Date)
	if err != nil {
		return false
	}
	err = r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM transactions
			WHERE user_id = $1
				AND deleted_at IS NULL
				AND type = $2
				AND occurred_at = $3
				AND ROUND(amount::numeric, 2) = ROUND($4::numeric, 2)
				AND lower(trim(coalesce(memo, ''))) = lower(trim($5))
		)
	`, strings.TrimSpace(userID), normalizeImportTransactionType(row.Type), occurredAt.UTC(), math.Abs(row.Amount), strings.TrimSpace(row.Description)).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func (r *PostgresRepository) SaveRow(userID string, row storedImportRow) {
	tripleKey := r.importRowKey(userID, row)
	r.db.Exec(`
		INSERT INTO import_rows (user_id, date, amount, description, triple_key, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, userID, row.Date, row.Amount, row.Description, tripleKey)
	r.db.Exec(`
		INSERT INTO import_dedup (user_id, triple_key)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, tripleKey)
}

func (r *PostgresRepository) StoredRowCount(userID string) int {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM import_rows WHERE user_id = $1
	`, userID).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (r *PostgresRepository) importRowKey(userID string, row storedImportRow) string {
	return strings.TrimSpace(userID) + "|" + normalizeImportTransactionType(row.Type) + "|" + strings.TrimSpace(row.Date) + "|" + strconv.FormatFloat(math.Abs(row.Amount), 'f', 2, 64) + "|" + strings.TrimSpace(strings.ToLower(row.Description))
}

type PostgresImportJob struct {
	UserID         string
	Path           string
	IdempotencyKey string
	CreatedAt      time.Time
	ResponseJSON   string
	ErrorCode      string
}

func (r *PostgresRepository) Now() time.Time {
	return time.Now().UTC()
}

func (r *PostgresRepository) ResolveImportReferences(userID string, row ImportRow) ImportRow {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return row
	}

	row.LedgerID = r.resolveImportLedgerID(trimmedUserID, row.LedgerID, row.Ledger)
	row.AccountID = r.resolveImportAccountID(trimmedUserID, row.AccountID, row.Account)
	row.CategoryID = r.resolveImportCategoryID(trimmedUserID, row.CategoryID)
	return row
}

func (r *PostgresRepository) SaveImportedTransaction(userID string, row ImportRow) error {
	trimmedUserID := strings.TrimSpace(userID)
	trimmedDate := strings.TrimSpace(row.Date)
	trimmedDescription := strings.TrimSpace(row.Description)
	if trimmedUserID == "" || trimmedDate == "" || trimmedDescription == "" || row.Amount <= 0 || math.IsNaN(row.Amount) || math.IsInf(row.Amount, 0) {
		return errors.New("invalid import row")
	}

	row = r.ResolveImportReferences(trimmedUserID, row)
	ledgerID := strings.TrimSpace(row.LedgerID)
	if ledgerID == "" {
		return errors.New("missing import ledger")
	}
	accountID := sql.NullString{String: strings.TrimSpace(row.AccountID), Valid: strings.TrimSpace(row.AccountID) != ""}

	occurredAt, err := parseImportOccurredAt(trimmedDate)
	if err != nil {
		return err
	}

	txnType := normalizeImportTransactionType(row.Type)
	amount := math.Abs(row.Amount)
	categoryName := strings.TrimSpace(row.Category)
	categoryID := sql.NullString{String: strings.TrimSpace(row.CategoryID), Valid: strings.TrimSpace(row.CategoryID) != ""}

	_, err = r.db.Exec(`
		INSERT INTO transactions (
			id, user_id, ledger_id, account_id, category_id, type, amount, occurred_at, category_name, memo, created_at
		)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`, trimmedUserID, ledgerID, nullableString(accountID), nullableString(categoryID), txnType, amount, occurredAt.UTC(), categoryName, trimmedDescription)
	return err
}

func (r *PostgresRepository) resolveImportLedgerID(userID string, ledgerID string, ledgerName string) string {
	if ownedID := r.findOwnedLedgerID(userID, ledgerID); ownedID != "" {
		return ownedID
	}
	if strings.TrimSpace(ledgerName) != "" {
		var matchedID string
		_ = r.db.QueryRow(`
			SELECT id::text
			FROM ledgers
			WHERE user_id = $1 AND LOWER(name) = LOWER($2)
			ORDER BY is_default DESC, created_at ASC, id ASC
			LIMIT 1
		`, userID, strings.TrimSpace(ledgerName)).Scan(&matchedID)
		if matchedID != "" {
			return matchedID
		}
	}
	var defaultID string
	_ = r.db.QueryRow(`
		SELECT id::text
		FROM ledgers
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at ASC, id ASC
		LIMIT 1
	`, userID).Scan(&defaultID)
	return defaultID
}

func (r *PostgresRepository) resolveImportAccountID(userID string, accountID string, accountName string) string {
	if ownedID := r.findOwnedAccountID(userID, accountID); ownedID != "" {
		return ownedID
	}
	if strings.TrimSpace(accountName) != "" {
		var matchedID string
		_ = r.db.QueryRow(`
			SELECT id::text
			FROM accounts
			WHERE user_id = $1 AND archived_at IS NULL AND LOWER(name) = LOWER($2)
			ORDER BY created_at ASC, id ASC
			LIMIT 1
		`, userID, strings.TrimSpace(accountName)).Scan(&matchedID)
		if matchedID != "" {
			return matchedID
		}
	}
	var fallbackID string
	_ = r.db.QueryRow(`
		SELECT id::text
		FROM accounts
		WHERE user_id = $1 AND archived_at IS NULL
		ORDER BY created_at ASC, id ASC
		LIMIT 1
	`, userID).Scan(&fallbackID)
	return fallbackID
}

func (r *PostgresRepository) resolveImportCategoryID(userID string, categoryID string) string {
	trimmedCategoryID := strings.TrimSpace(categoryID)
	if trimmedCategoryID == "" {
		return ""
	}
	var ownedID string
	_ = r.db.QueryRow(`
		SELECT id::text
		FROM categories
		WHERE user_id = $1 AND id = $2 AND archived_at IS NULL
		LIMIT 1
	`, userID, trimmedCategoryID).Scan(&ownedID)
	return ownedID
}

func (r *PostgresRepository) findOwnedLedgerID(userID string, ledgerID string) string {
	trimmedLedgerID := strings.TrimSpace(ledgerID)
	if trimmedLedgerID == "" {
		return ""
	}
	var ownedID string
	_ = r.db.QueryRow(`
		SELECT id::text
		FROM ledgers
		WHERE user_id = $1 AND id = $2
		LIMIT 1
	`, userID, trimmedLedgerID).Scan(&ownedID)
	return ownedID
}

func (r *PostgresRepository) findOwnedAccountID(userID string, accountID string) string {
	trimmedAccountID := strings.TrimSpace(accountID)
	if trimmedAccountID == "" {
		return ""
	}
	var ownedID string
	_ = r.db.QueryRow(`
		SELECT id::text
		FROM accounts
		WHERE user_id = $1 AND id = $2 AND archived_at IS NULL
		LIMIT 1
	`, userID, trimmedAccountID).Scan(&ownedID)
	return ownedID
}

func nullableString(value sql.NullString) interface{} {
	if !value.Valid {
		return nil
	}
	return value.String
}

func parseImportOccurredAt(value string) (time.Time, error) {
	layouts := []string{
		"2006/01/02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
		"2006/01/02",
		"2006-01-02",
	}
	trimmed := strings.TrimSpace(value)
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, errors.New("invalid occurred_at")
}
