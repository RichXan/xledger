package accounting

import (
	"database/sql"
	"errors"
)

type PostgresLedgerRepository struct {
	db *sql.DB
}

func NewPostgresLedgerRepository(db *sql.DB) *PostgresLedgerRepository {
	return &PostgresLedgerRepository{db: db}
}

func (r *PostgresLedgerRepository) Create(userID string, input LedgerCreateInput) (Ledger, error) {
	var ledger Ledger
	err := r.db.QueryRow(`
		INSERT INTO ledgers (id, user_id, name, is_default, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		RETURNING id, user_id, name, is_default
	`, userID, input.Name, input.IsDefault).Scan(&ledger.ID, &ledger.UserID, &ledger.Name, &ledger.IsDefault)
	return ledger, err
}

func (r *PostgresLedgerRepository) ListByUser(userID string) ([]Ledger, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name, is_default
		FROM ledgers
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ledgers := make([]Ledger, 0)
	for rows.Next() {
		var ledger Ledger
		if err := rows.Scan(&ledger.ID, &ledger.UserID, &ledger.Name, &ledger.IsDefault); err != nil {
			return nil, err
		}
		ledgers = append(ledgers, ledger)
	}
	return ledgers, rows.Err()
}

func (r *PostgresLedgerRepository) GetByIDForUser(userID string, ledgerID string) (Ledger, bool, error) {
	var ledger Ledger
	err := r.db.QueryRow(`
		SELECT id, user_id, name, is_default
		FROM ledgers
		WHERE id = $1 AND user_id = $2
	`, ledgerID, userID).Scan(&ledger.ID, &ledger.UserID, &ledger.Name, &ledger.IsDefault)
	if errors.Is(err, sql.ErrNoRows) {
		return Ledger{}, false, nil
	}
	if err != nil {
		return Ledger{}, false, err
	}
	return ledger, true, nil
}

func (r *PostgresLedgerRepository) SaveByIDForUser(userID string, ledgerID string, ledger Ledger) (Ledger, bool, error) {
	var updated Ledger
	err := r.db.QueryRow(`
		UPDATE ledgers
		SET name = $3
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, is_default
	`, ledgerID, userID, ledger.Name).Scan(&updated.ID, &updated.UserID, &updated.Name, &updated.IsDefault)
	if errors.Is(err, sql.ErrNoRows) {
		return Ledger{}, false, nil
	}
	if err != nil {
		return Ledger{}, false, err
	}
	return updated, true, nil
}

func (r *PostgresLedgerRepository) DeleteByIDForUser(userID string, ledgerID string) (bool, error) {
	result, err := r.db.Exec(`
		DELETE FROM ledgers
			WHERE id = $1 AND user_id = $2 AND is_default = false
	`, ledgerID, userID)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
