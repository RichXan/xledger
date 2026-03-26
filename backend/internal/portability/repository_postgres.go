package portability

import (
	"database/sql"
	"encoding/json"
	"errors"
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
		SELECT user_id, path, idempotency_key, created_at, response_json, error_code
		FROM import_jobs
		WHERE user_id = $1 AND path = $2 AND idempotency_key = $3
		AND created_at > NOW() - INTERVAL '24 hours'
	`, userID, path, idempotencyKey).Scan(
		&job.UserID, &job.Path, &job.IdempotencyKey, &job.CreatedAt, &responseJSON, &job.ErrorCode,
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
		INSERT INTO import_jobs (user_id, path, idempotency_key, created_at, response_json, error_code)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, path, idempotency_key) DO UPDATE SET
			response_json = EXCLUDED.response_json,
			error_code = EXCLUDED.error_code,
			created_at = EXCLUDED.created_at
	`, job.UserID, job.Path, job.IdempotencyKey, job.CreatedAt, string(responseJSON), job.ErrorCode)
}

func (r *PostgresRepository) HasTriple(userID string, row storedImportRow) bool {
	var exists bool
	tripleKey := r.tripleKey(userID, row)
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM import_dedup
			WHERE user_id = $1 AND triple_key = $2
		)
	`, userID, tripleKey).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func (r *PostgresRepository) SaveRow(userID string, row storedImportRow) {
	tripleKey := r.tripleKey(userID, row)
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

func (r *PostgresRepository) tripleKey(userID string, row storedImportRow) string {
	return strings.TrimSpace(userID) + "|" + strings.TrimSpace(row.Date) + "|" + strconv.FormatFloat(row.Amount, 'f', 2, 64) + "|" + strings.TrimSpace(strings.ToLower(row.Description))
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
