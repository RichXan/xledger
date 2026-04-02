package accounting

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) Create(userID string, input TransactionCreateInput) (Transaction, error) {
	var txn Transaction
	var accountID, categoryID, transferPairID, transferSide *string
	var fromAccountID, toAccountID *string

	err := r.db.QueryRow(`
		INSERT INTO transactions (
			id, user_id, ledger_id, account_id, category_id, category_name, memo,
			from_account_id, to_account_id, transfer_pair_id, transfer_side,
			type, amount, version, occurred_at, created_at
		) VALUES (
				gen_random_uuid(), $1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10,
				$11, $12, 1, $13, NOW()
		)
		RETURNING id, user_id, ledger_id, account_id, category_id, category_name, memo,
			from_account_id, to_account_id, transfer_pair_id, transfer_side,
			type, amount, version, occurred_at
	`, userID, input.LedgerID, input.AccountID, input.CategoryID, nil, strings.TrimSpace(input.Memo),
		input.FromAccountID, input.ToAccountID, nil, nil,
		input.Type, input.Amount, input.OccurredAt).Scan(
		&txn.ID, &txn.UserID, &txn.LedgerID, &accountID, &categoryID, &txn.CategoryName, &txn.Memo,
		&fromAccountID, &toAccountID, &transferPairID, &transferSide,
		&txn.Type, &txn.Amount, &txn.Version, &txn.OccurredAt,
	)
	if err != nil {
		return Transaction{}, err
	}

	txn.AccountID = accountID
	txn.CategoryID = categoryID
	txn.FromAccountID = fromAccountID
	txn.ToAccountID = toAccountID
	txn.TransferPairID = transferPairID
	if transferSide != nil {
		txn.TransferSide = *transferSide
	}
	return txn, nil
}

func (r *PostgresTransactionRepository) GetByIDForUser(userID string, txnID string) (Transaction, bool, error) {
	var txn Transaction
	var accountID, categoryID, transferPairID, transferSide *string
	var fromAccountID, toAccountID *string
	var memo sql.NullString

	err := r.db.QueryRow(`
		SELECT id, user_id, ledger_id, account_id, category_id, category_name, memo,
			from_account_id, to_account_id, transfer_pair_id, transfer_side,
			type, amount, version, occurred_at
		FROM transactions
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, txnID, userID).Scan(
		&txn.ID, &txn.UserID, &txn.LedgerID, &accountID, &categoryID, &txn.CategoryName, &memo,
		&fromAccountID, &toAccountID, &transferPairID, &transferSide,
		&txn.Type, &txn.Amount, &txn.Version, &txn.OccurredAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, false, nil
	}
	if err != nil {
		return Transaction{}, false, err
	}

	txn.AccountID = accountID
	txn.CategoryID = categoryID
	if memo.Valid {
		txn.Memo = memo.String
	}
	txn.FromAccountID = fromAccountID
	txn.ToAccountID = toAccountID
	txn.TransferPairID = transferPairID
	if transferSide != nil {
		txn.TransferSide = *transferSide
	}
	return txn, true, nil
}

func (r *PostgresTransactionRepository) SaveByIDForUser(userID string, txnID string, txn Transaction) (Transaction, bool, error) {
	var updated Transaction
	var accountID, categoryID, transferPairID, transferSide *string
	var fromAccountID, toAccountID *string

	err := r.db.QueryRow(`
		UPDATE transactions
		SET amount = $3, category_id = $4, category_name = $5, memo = $6, version = $7
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, ledger_id, account_id, category_id, category_name, memo,
			from_account_id, to_account_id, transfer_pair_id, transfer_side,
			type, amount, version, occurred_at
	`, txnID, userID, txn.Amount, txn.CategoryID, txn.CategoryName, strings.TrimSpace(txn.Memo), txn.Version).Scan(
		&updated.ID, &updated.UserID, &updated.LedgerID, &accountID, &categoryID, &updated.CategoryName, &updated.Memo,
		&fromAccountID, &toAccountID, &transferPairID, &transferSide,
		&updated.Type, &updated.Amount, &updated.Version, &updated.OccurredAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Transaction{}, false, nil
	}
	if err != nil {
		return Transaction{}, false, err
	}

	updated.AccountID = accountID
	updated.CategoryID = categoryID
	updated.FromAccountID = fromAccountID
	updated.ToAccountID = toAccountID
	updated.TransferPairID = transferPairID
	if transferSide != nil {
		updated.TransferSide = *transferSide
	}
	return updated, true, nil
}

func (r *PostgresTransactionRepository) DeleteByIDForUser(userID string, txnID string) (bool, error) {
	result, err := r.db.Exec(`
		UPDATE transactions SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, txnID, userID)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *PostgresTransactionRepository) CreateTransferPair(userID string, pairID string, fromInput TransactionCreateInput, toInput TransactionCreateInput) (Transaction, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return Transaction{}, err
	}
	defer tx.Rollback()

	var fromTxn Transaction
	var fromAccountID, toAccountID *string
	err = tx.QueryRow(`
		INSERT INTO transactions (
			id, user_id, ledger_id, from_account_id, to_account_id,
			transfer_pair_id, transfer_side, type, amount, version, occurred_at, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, 'from', 'transfer', $6, 1, $7, NOW()
		)
		RETURNING id, user_id, ledger_id, from_account_id, to_account_id,
			transfer_pair_id, transfer_side, type, amount, version, occurred_at
	`, userID, fromInput.LedgerID, fromInput.FromAccountID, fromInput.ToAccountID, pairID, fromInput.Amount, fromInput.OccurredAt).Scan(
		&fromTxn.ID, &fromTxn.UserID, &fromTxn.LedgerID, &fromAccountID, &toAccountID,
		&fromTxn.TransferPairID, &fromTxn.TransferSide, &fromTxn.Type, &fromTxn.Amount, &fromTxn.Version, &fromTxn.OccurredAt,
	)
	if err != nil {
		return Transaction{}, err
	}
	fromTxn.FromAccountID = fromAccountID
	fromTxn.ToAccountID = toAccountID

	_, err = tx.Exec(`
		INSERT INTO transactions (
			id, user_id, ledger_id, from_account_id, to_account_id,
			transfer_pair_id, transfer_side, type, amount, version, occurred_at, created_at
		) VALUES (
				gen_random_uuid(), $1, $2, $3, $4, $5, 'to', 'transfer', $6, 1, $7, NOW()
		)
	`, userID, toInput.LedgerID, toInput.FromAccountID, toInput.ToAccountID, pairID, toInput.Amount, toInput.OccurredAt)
	if err != nil {
		return Transaction{}, err
	}

	return fromTxn, tx.Commit()
}

func (r *PostgresTransactionRepository) GetTransferPairByTxnID(userID string, txnID string) ([]Transaction, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.user_id, t.ledger_id, t.account_id, t.category_id, t.category_name, t.memo,
			t.from_account_id, t.to_account_id, t.transfer_pair_id, t.transfer_side,
			t.type, t.amount, t.version, t.occurred_at
		FROM transactions t
		WHERE t.transfer_pair_id = (
			SELECT transfer_pair_id FROM transactions
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		) AND t.user_id = $2 AND t.deleted_at IS NULL
	`, txnID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pair []Transaction
	for rows.Next() {
		var txn Transaction
		var accountID, categoryID, transferPairID, transferSide *string
		var fromAccountID, toAccountID *string
		var memo sql.NullString
		if err := rows.Scan(
			&txn.ID, &txn.UserID, &txn.LedgerID, &accountID, &categoryID, &txn.CategoryName, &memo,
			&fromAccountID, &toAccountID, &transferPairID, &transferSide,
			&txn.Type, &txn.Amount, &txn.Version, &txn.OccurredAt,
		); err != nil {
			return nil, err
		}
		txn.AccountID = accountID
		txn.CategoryID = categoryID
		if memo.Valid {
			txn.Memo = memo.String
		}
		txn.FromAccountID = fromAccountID
		txn.ToAccountID = toAccountID
		txn.TransferPairID = transferPairID
		if transferSide != nil {
			txn.TransferSide = *transferSide
		}
		pair = append(pair, txn)
	}
	return pair, rows.Err()
}

func (r *PostgresTransactionRepository) UpdateTransferPairAmount(userID string, pairID string, amount float64, expectedVersion *int) (Transaction, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return Transaction{}, err
	}
	defer tx.Rollback()

	// Acquire advisory lock BEFORE reading version to prevent race condition
	_, err = tx.Exec(`
		SELECT pg_advisory_xact_lock(hashtext($1 || '|' || $2))
	`, userID, pairID)
	if err != nil {
		return Transaction{}, err
	}

	var currentVersion int
	err = tx.QueryRow(`
		SELECT version FROM transactions
			WHERE transfer_pair_id = $1 AND user_id = $2 AND transfer_side = 'from' AND deleted_at IS NULL
	`, pairID, userID).Scan(&currentVersion)
	if err != nil {
		return Transaction{}, err
	}

	if expectedVersion != nil && currentVersion != *expectedVersion {
		return Transaction{}, errTransferVersionConflict
	}

	var updated Transaction
	var accountID, categoryID, transferPairID, transferSide *string
	var fromAccountID, toAccountID *string
	var memo sql.NullString
	err = tx.QueryRow(`
		WITH updated AS (
			UPDATE transactions
			SET amount = $3, version = version + 1
			WHERE transfer_pair_id = $1 AND user_id = $2 AND deleted_at IS NULL
			RETURNING id, user_id, ledger_id, account_id, category_id, category_name, memo,
				from_account_id, to_account_id, transfer_pair_id, transfer_side,
				type, amount, version, occurred_at
		)
		SELECT id, user_id, ledger_id, account_id, category_id, category_name, memo,
			from_account_id, to_account_id, transfer_pair_id, transfer_side,
			type, amount, version, occurred_at
		FROM updated
		WHERE transfer_side = 'from'
		LIMIT 1
	`, pairID, userID, amount).Scan(
		&updated.ID, &updated.UserID, &updated.LedgerID, &accountID, &categoryID, &updated.CategoryName, &memo,
		&fromAccountID, &toAccountID, &transferPairID, &transferSide,
		&updated.Type, &updated.Amount, &updated.Version, &updated.OccurredAt,
	)
	if err != nil {
		return Transaction{}, err
	}

	updated.AccountID = accountID
	updated.CategoryID = categoryID
	if memo.Valid {
		updated.Memo = memo.String
	}
	updated.FromAccountID = fromAccountID
	updated.ToAccountID = toAccountID
	updated.TransferPairID = transferPairID
	if transferSide != nil {
		updated.TransferSide = *transferSide
	}
	return updated, tx.Commit()
}

func (r *PostgresTransactionRepository) DeleteTransferPairByTxnID(userID string, txnID string, expectedVersion *int) ([]string, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var pairID string
	err = tx.QueryRow(`
		SELECT transfer_pair_id FROM transactions
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, txnID, userID).Scan(&pairID)
	if err != nil {
		return nil, err
	}

	if expectedVersion != nil {
		var currentVersion int
		err = tx.QueryRow(`
			SELECT version FROM transactions
				WHERE transfer_pair_id = $1 AND user_id = $2 AND deleted_at IS NULL
				LIMIT 1
		`, pairID, userID).Scan(&currentVersion)
		if err != nil {
			return nil, err
		}
		if currentVersion != *expectedVersion {
			return nil, errTransferVersionConflict
		}
	}

	rows, err := tx.Query(`
		UPDATE transactions SET deleted_at = NOW()
		WHERE transfer_pair_id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING ledger_id
	`, pairID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ledgers []string
	for rows.Next() {
		var ledgerID string
		if err := rows.Scan(&ledgerID); err != nil {
			return nil, err
		}
		ledgers = append(ledgers, ledgerID)
	}
	return ledgers, tx.Commit()
}

func (r *PostgresTransactionRepository) WithTransferPairLock(userID string, pairID string, fn func() error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		SELECT pg_advisory_xact_lock(hashtext($1 || '|' || $2))
	`, userID, pairID)
	if err != nil {
		return err
	}

	if err := fn(); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresTransactionRepository) ListByUser(userID string, query TransactionQuery) ([]Transaction, error) {
	sqlQuery := `
		SELECT id, user_id, ledger_id, account_id, category_id, category_name, memo,
			from_account_id, to_account_id, transfer_pair_id, transfer_side,
			type, amount, version, occurred_at
		FROM transactions
			WHERE user_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{userID}
	argIdx := 2

	if query.LedgerID != "" {
		sqlQuery += ` AND ledger_id = $` + itoa(argIdx)
		args = append(args, query.LedgerID)
		argIdx++
	}
	if query.AccountID != "" {
		sqlQuery += ` AND (account_id = $` + itoa(argIdx) + ` OR from_account_id = $` + itoa(argIdx) + ` OR to_account_id = $` + itoa(argIdx) + `)`
		args = append(args, query.AccountID)
		argIdx++
	}
	if query.CategoryID != "" {
		sqlQuery += ` AND category_id = $` + itoa(argIdx)
		args = append(args, query.CategoryID)
		argIdx++
	}
	if !query.OccurredFrom.IsZero() {
		sqlQuery += ` AND occurred_at >= $` + itoa(argIdx)
		args = append(args, query.OccurredFrom)
		argIdx++
	}
	if !query.OccurredTo.IsZero() {
		sqlQuery += ` AND occurred_at <= $` + itoa(argIdx)
		args = append(args, query.OccurredTo)
		argIdx++
	}
	if query.UseTransactionIDs && len(query.TransactionIDs) > 0 {
		sqlQuery += ` AND id = ANY($` + itoa(argIdx) + `)`
		args = append(args, query.TransactionIDs)
		argIdx++
	}

	sqlQuery += ` ORDER BY occurred_at DESC, id DESC`

	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		sqlQuery += ` LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
		args = append(args, query.PageSize, offset)
	}

	rows, err := r.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Transaction
	for rows.Next() {
		var txn Transaction
		var accountID, categoryID, transferPairID, transferSide *string
		var fromAccountID, toAccountID *string
		var memo sql.NullString
		if err := rows.Scan(
			&txn.ID, &txn.UserID, &txn.LedgerID, &accountID, &categoryID, &txn.CategoryName, &memo,
			&fromAccountID, &toAccountID, &transferPairID, &transferSide,
			&txn.Type, &txn.Amount, &txn.Version, &txn.OccurredAt,
		); err != nil {
			return nil, err
		}
		txn.AccountID = accountID
		txn.CategoryID = categoryID
		if memo.Valid {
			txn.Memo = memo.String
		}
		txn.FromAccountID = fromAccountID
		txn.ToAccountID = toAccountID
		txn.TransferPairID = transferPairID
		if transferSide != nil {
			txn.TransferSide = *transferSide
		}
		items = append(items, txn)
	}
	return items, rows.Err()
}

func (r *PostgresTransactionRepository) CountByUser(userID string, query TransactionQuery) (int, error) {
	sqlQuery := `
		SELECT COUNT(*)
		FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{userID}
	argIdx := 2

	if query.LedgerID != "" {
		sqlQuery += ` AND ledger_id = $` + itoa(argIdx)
		args = append(args, query.LedgerID)
		argIdx++
	}
	if query.AccountID != "" {
		sqlQuery += ` AND (account_id = $` + itoa(argIdx) + ` OR from_account_id = $` + itoa(argIdx) + ` OR to_account_id = $` + itoa(argIdx) + `)`
		args = append(args, query.AccountID)
		argIdx++
	}
	if query.CategoryID != "" {
		sqlQuery += ` AND category_id = $` + itoa(argIdx)
		args = append(args, query.CategoryID)
		argIdx++
	}
	if !query.OccurredFrom.IsZero() {
		sqlQuery += ` AND occurred_at >= $` + itoa(argIdx)
		args = append(args, query.OccurredFrom)
		argIdx++
	}
	if !query.OccurredTo.IsZero() {
		sqlQuery += ` AND occurred_at <= $` + itoa(argIdx)
		args = append(args, query.OccurredTo)
		argIdx++
	}
	if query.UseTransactionIDs && len(query.TransactionIDs) > 0 {
		sqlQuery += ` AND id = ANY($` + itoa(argIdx) + `)`
		args = append(args, query.TransactionIDs)
		argIdx++
	}

	var count int
	err := r.db.QueryRow(sqlQuery, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresTransactionRepository) ListByTransferPairForUser(userID string, pairID string) ([]Transaction, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, ledger_id, account_id, category_id, category_name, memo,
				from_account_id, to_account_id, transfer_pair_id, transfer_side,
				type, amount, version, occurred_at
			FROM transactions
			WHERE transfer_pair_id = $1 AND user_id = $2 AND deleted_at IS NULL
			ORDER BY transfer_side ASC
	`, pairID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pair []Transaction
	for rows.Next() {
		var txn Transaction
		var accountID, categoryID, transferPairID, transferSide *string
		var fromAccountID, toAccountID *string
		var memo sql.NullString
		if err := rows.Scan(
			&txn.ID, &txn.UserID, &txn.LedgerID, &accountID, &categoryID, &txn.CategoryName, &memo,
			&fromAccountID, &toAccountID, &transferPairID, &transferSide,
			&txn.Type, &txn.Amount, &txn.Version, &txn.OccurredAt,
		); err != nil {
			return nil, err
		}
		txn.AccountID = accountID
		txn.CategoryID = categoryID
		if memo.Valid {
			txn.Memo = memo.String
		}
		txn.FromAccountID = fromAccountID
		txn.ToAccountID = toAccountID
		txn.TransferPairID = transferPairID
		if transferSide != nil {
			txn.TransferSide = *transferSide
		}
		pair = append(pair, txn)
	}
	return pair, rows.Err()
}

func (r *PostgresTransactionRepository) MarkBalancesRecalculated(userID string, ledgerID string) error {
	_, err := r.db.Exec(`
		INSERT INTO balance_recalc_log (user_id, ledger_id, recalculated_at)
		VALUES ($1, $2, NOW())
	`, userID, ledgerID)
	return err
}

func (r *PostgresTransactionRepository) MarkStatsInputRecalculated(userID string, ledgerID string) error {
	_, err := r.db.Exec(`
		INSERT INTO stats_recalc_log (user_id, ledger_id, recalculated_at)
		VALUES ($1, $2, NOW())
	`, userID, ledgerID)
	return err
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

// GetOverviewStats uses SQL SUM/FILTER to aggregate income and expense without
// loading transaction rows into Go memory.
func (r *PostgresTransactionRepository) GetOverviewStats(userID string, query TransactionQuery) (float64, float64, error) {
	sqlQuery := `
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0),
			COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0)
		FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL AND type != 'transfer'
	`
	args := []interface{}{userID}
	argIdx := 2

	if query.LedgerID != "" {
		sqlQuery += ` AND ledger_id = $` + itoa(argIdx)
		args = append(args, query.LedgerID)
		argIdx++
	}
	if !query.OccurredFrom.IsZero() {
		sqlQuery += ` AND occurred_at >= $` + itoa(argIdx)
		args = append(args, query.OccurredFrom)
		argIdx++
	}
	if !query.OccurredTo.IsZero() {
		sqlQuery += ` AND occurred_at <= $` + itoa(argIdx)
		args = append(args, query.OccurredTo)
		_ = argIdx
	}

	var income, expense float64
	if err := r.db.QueryRow(sqlQuery, args...).Scan(&income, &expense); err != nil {
		return 0, 0, err
	}
	return income, expense, nil
}

// GetTrendStats uses date_trunc + GROUP BY to compute per-bucket income/expense
// aggregations on the database side.
func (r *PostgresTransactionRepository) GetTrendStats(userID string, query TransactionQuery, granularity string, loc *time.Location) ([]TrendRow, error) {
	if loc == nil {
		loc = time.FixedZone("UTC+8", 8*60*60)
	}
	tz := loc.String()
	// time.FixedZone returns names like "UTC+8" which Postgres doesn't understand;
	// map to the conventional IANA name used in practice.
	if tz == "" || strings.HasPrefix(tz, "UTC+") || strings.HasPrefix(tz, "UTC-") {
		tz = "Asia/Shanghai" // sensible default matching UTC+8
	}

	sqlQuery := `
		SELECT
			date_trunc($1, occurred_at AT TIME ZONE $2) AS bucket,
			COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0),
			COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0)
		FROM transactions
		WHERE user_id = $3 AND deleted_at IS NULL AND type != 'transfer'
	`
	args := []interface{}{granularity, tz, userID}
	argIdx := 4

	if query.LedgerID != "" {
		sqlQuery += ` AND ledger_id = $` + itoa(argIdx)
		args = append(args, query.LedgerID)
		argIdx++
	}
	if !query.OccurredFrom.IsZero() {
		sqlQuery += ` AND occurred_at >= $` + itoa(argIdx)
		args = append(args, query.OccurredFrom)
		argIdx++
	}
	if !query.OccurredTo.IsZero() {
		sqlQuery += ` AND occurred_at <= $` + itoa(argIdx)
		args = append(args, query.OccurredTo)
		_ = argIdx
	}

	sqlQuery += ` GROUP BY bucket ORDER BY bucket ASC`

	rows, err := r.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TrendRow
	for rows.Next() {
		var row TrendRow
		if err := rows.Scan(&row.BucketStart, &row.Income, &row.Expense); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
