package accounting

import (
	"database/sql"
	"errors"
)

type PostgresAccountRepository struct {
	db *sql.DB
}

func NewPostgresAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) Create(userID string, input AccountCreateInput) (Account, error) {
	var account Account
	err := r.db.QueryRow(`
		INSERT INTO accounts (id, user_id, name, type, initial_balance, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		RETURNING id, user_id, name, type, initial_balance, archived_at
	`, userID, input.Name, input.Type, input.InitialBalance).Scan(
		&account.ID, &account.UserID, &account.Name, &account.Type, &account.InitialBalance, &account.ArchivedAt,
	)
	return account, err
}

func (r *PostgresAccountRepository) ListByUser(userID string) ([]Account, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name, type, initial_balance, archived_at
		FROM accounts
			WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]Account, 0)
	for rows.Next() {
		var account Account
		if err := rows.Scan(&account.ID, &account.UserID, &account.Name, &account.Type, &account.InitialBalance, &account.ArchivedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

func (r *PostgresAccountRepository) GetByIDForUser(userID string, accountID string) (Account, bool, error) {
	var account Account
	err := r.db.QueryRow(`
		SELECT id, user_id, name, type, initial_balance, archived_at
		FROM accounts
		WHERE id = $1 AND user_id = $2
	`, accountID, userID).Scan(
		&account.ID, &account.UserID, &account.Name, &account.Type, &account.InitialBalance, &account.ArchivedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, false, nil
	}
	if err != nil {
		return Account{}, false, err
	}
	return account, true, nil
}

func (r *PostgresAccountRepository) SaveByIDForUser(userID string, accountID string, account Account) (Account, bool, error) {
	var updated Account
	err := r.db.QueryRow(`
		UPDATE accounts
		SET name = $3, type = $4, archived_at = $5
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, type, initial_balance, archived_at
	`, accountID, userID, account.Name, account.Type, account.ArchivedAt).Scan(
		&updated.ID, &updated.UserID, &updated.Name, &updated.Type, &updated.InitialBalance, &updated.ArchivedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, false, nil
	}
	if err != nil {
		return Account{}, false, err
	}
	return updated, true, nil
}

func (r *PostgresAccountRepository) DeleteByIDForUser(userID string, accountID string) (bool, error) {
	result, err := r.db.Exec(`
		DELETE FROM accounts
		WHERE id = $1 AND user_id = $2
	`, accountID, userID)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
