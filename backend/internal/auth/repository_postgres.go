package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

var ErrAuthUserAlreadyExists = errors.New("auth user already exists")

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) SaveVerificationCode(ctx context.Context, email string, code string, ttl time.Duration) error {
	expiresAt := time.Now().UTC().Add(ttl)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO verification_codes (email, code_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			code_hash = EXCLUDED.code_hash,
			expires_at = EXCLUDED.expires_at,
			created_at = EXCLUDED.created_at,
			failed_attempts = 0
	`, email, code, expiresAt, time.Now().UTC())
	return err
}

func (r *PostgresRepository) GetVerificationCode(ctx context.Context, email string) (string, error) {
	var codeHash string
	var expiresAt time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT code_hash, expires_at FROM verification_codes
		WHERE email = $1 AND expires_at > NOW()
	`, email).Scan(&codeHash, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrCodeNotFound
	}
	if err != nil {
		return "", err
	}
	return codeHash, nil
}

func (r *PostgresRepository) VerifyAndConsumeCode(ctx context.Context, email string, codeDigest string, maxAttempts int) (VerifyConsumeResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return VerifyConsumeNone, err
	}
	defer tx.Rollback()

	var codeHash string
	var expiresAt time.Time
	var failedAttempts int
	err = tx.QueryRowContext(ctx, `
		SELECT code_hash, expires_at, failed_attempts FROM verification_codes
		WHERE email = $1 FOR UPDATE
	`, email).Scan(&codeHash, &expiresAt, &failedAttempts)
	if errors.Is(err, sql.ErrNoRows) {
		return VerifyConsumeNone, nil
	}
	if err != nil {
		return VerifyConsumeNone, err
	}

	if time.Now().UTC().After(expiresAt) {
		tx.ExecContext(ctx, `DELETE FROM verification_codes WHERE email = $1`, email)
		tx.Commit()
		return VerifyConsumeExpired, nil
	}

	if failedAttempts >= maxAttempts {
		tx.ExecContext(ctx, `DELETE FROM verification_codes WHERE email = $1`, email)
		tx.Commit()
		return VerifyConsumeMismatch, nil
	}

	if codeHash != codeDigest {
		tx.ExecContext(ctx, `
			UPDATE verification_codes SET failed_attempts = failed_attempts + 1
			WHERE email = $1
		`, email)
		tx.Commit()
		return VerifyConsumeMismatch, nil
	}

	tx.ExecContext(ctx, `DELETE FROM verification_codes WHERE email = $1`, email)
	return VerifyConsumeMatch, tx.Commit()
}

func (r *PostgresRepository) DeleteVerificationCode(ctx context.Context, email string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM verification_codes WHERE email = $1`, email)
	return err
}

func (r *PostgresRepository) RecordFailedVerificationAttempt(ctx context.Context, email string) (int, error) {
	var attempts int
	err := r.db.QueryRowContext(ctx, `
		UPDATE verification_codes
		SET failed_attempts = failed_attempts + 1
		WHERE email = $1
		RETURNING failed_attempts
	`, email).Scan(&attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return attempts, err
}

func (r *PostgresRepository) AcquireIPHourlySlot(ctx context.Context, ip string, at time.Time, ttl time.Duration, cap int) (bool, error) {
	windowStart := at.Truncate(ttl)
	windowEnd := windowStart.Add(ttl)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM ip_rate_limits
		WHERE ip_address = $1 AND window_start = $2
	`, ip, windowStart).Scan(&count)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	if count >= cap {
		return false, nil
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ip_rate_limits (ip_address, window_start, window_end, request_count)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (ip_address, window_start) DO UPDATE SET
			request_count = ip_rate_limits.request_count + 1
	`, ip, windowStart, windowEnd)
	if err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (r *PostgresRepository) AcquireSendLock(ctx context.Context, email string, at time.Time, ttl time.Duration) (bool, error) {
	expiresAt := at.Add(ttl)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var currentExpiresAt *time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT expires_at FROM send_locks WHERE email = $1 FOR UPDATE
	`, email).Scan(&currentExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO send_locks (email, expires_at, created_at)
			VALUES ($1, $2, $3)
		`, email, expiresAt, at)
		if err != nil {
			return false, err
		}
		return true, tx.Commit()
	}
	if err != nil {
		return false, err
	}

	if currentExpiresAt != nil && time.Now().UTC().Before(*currentExpiresAt) {
		return false, nil
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE send_locks SET expires_at = $2, created_at = $3
		WHERE email = $1
	`, email, expiresAt, at)
	if err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func (r *PostgresRepository) ReleaseSendLock(ctx context.Context, email string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM send_locks WHERE email = $1`, email)
	return err
}

func (r *PostgresRepository) CreateSession(ctx context.Context, email string) error {
	return r.UpsertUser(ctx, email, "")
}

func (r *PostgresRepository) UpsertUser(ctx context.Context, email string, displayName string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, email, display_name, created_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())
		ON CONFLICT (email) DO UPDATE SET
			display_name = CASE
				WHEN EXCLUDED.display_name = '' THEN users.display_name
				ELSE EXCLUDED.display_name
			END
	`, email, displayName)
	return err
}

func (r *PostgresRepository) GetUserDisplayName(ctx context.Context, email string) (string, bool, error) {
	var displayName string
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(display_name, '')
		FROM users
		WHERE email = $1
	`, email).Scan(&displayName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return displayName, true, nil
}

func (r *PostgresRepository) CreateUserWithPassword(ctx context.Context, email string, displayName string, passwordHash string) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, email, display_name, password_hash, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
		ON CONFLICT (email) DO NOTHING
	`, email, displayName, passwordHash)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAuthUserAlreadyExists
	}
	return nil
}

func (r *PostgresRepository) GetUserCredential(ctx context.Context, email string) (UserCredentialRecord, bool, error) {
	var record UserCredentialRecord
	err := r.db.QueryRowContext(ctx, `
		SELECT email, COALESCE(display_name, ''), COALESCE(password_hash, '')
		FROM users
		WHERE email = $1
	`, email).Scan(&record.Email, &record.DisplayName, &record.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return UserCredentialRecord{}, false, nil
	}
	if err != nil {
		return UserCredentialRecord{}, false, err
	}
	return record, true, nil
}

func (r *PostgresRepository) UpdateUserPassword(ctx context.Context, email string, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE email = $1
	`, email, passwordHash)
	return err
}

func (r *PostgresRepository) UpdateUserDisplayName(ctx context.Context, email string, displayName string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET display_name = $2, updated_at = NOW()
		WHERE email = $1
	`, email, displayName)
	return err
}

func (r *PostgresRepository) SaveOAuthStateNonce(ctx context.Context, state string, nonce string, ttl time.Duration) error {
	return r.SaveOAuthStateNonceForEmail(ctx, state, nonce, "", ttl)
}

func (r *PostgresRepository) SaveOAuthStateNonceForEmail(ctx context.Context, state string, nonce string, email string, ttl time.Duration) error {
	expiresAt := time.Now().UTC().Add(ttl)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO oauth_states (state, nonce, email, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, state, nonce, email, expiresAt, time.Now().UTC())
	return err
}

func (r *PostgresRepository) ConsumeOAuthStateNonce(ctx context.Context, state string, nonce string) (bool, error) {
	_, ok, err := r.ConsumeOAuthStateNonceForEmail(ctx, state, nonce)
	return ok, err
}

func (r *PostgresRepository) CheckOAuthStateNonce(ctx context.Context, state string, nonce string) (bool, error) {
	var storedNonce string
	var expiresAt time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT nonce, expires_at
		FROM oauth_states
		WHERE state = $1
	`, state).Scan(&storedNonce, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if time.Now().UTC().After(expiresAt) {
		return false, nil
	}
	if storedNonce != nonce {
		return false, nil
	}
	return true, nil
}

func (r *PostgresRepository) ConsumeOAuthStateNonceForEmail(ctx context.Context, state string, nonce string) (string, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()

	var storedNonce, email string
	var expiresAt time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT nonce, email, expires_at FROM oauth_states
		WHERE state = $1 FOR UPDATE
	`, state).Scan(&storedNonce, &email, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	tx.ExecContext(ctx, `DELETE FROM oauth_states WHERE state = $1`, state)

	if time.Now().UTC().After(expiresAt) {
		tx.Commit()
		return "", false, nil
	}
	if storedNonce != nonce {
		tx.Commit()
		return "", false, nil
	}

	return email, true, tx.Commit()
}

func (r *PostgresRepository) StoreRefreshToken(ctx context.Context, tokenID string, email string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, consumed)
		SELECT $1, u.id, $2, $3, false
		FROM users u WHERE u.email = $4
	`, tokenID, tokenID, expiresAt, email)
	return err
}

func (r *PostgresRepository) ConsumeRefreshToken(ctx context.Context, tokenID string) (RefreshSession, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return RefreshSession{}, false, err
	}
	defer tx.Rollback()

	var email string
	var expiresAt time.Time
	var consumed bool
	err = tx.QueryRowContext(ctx, `
		SELECT u.email, rt.expires_at, rt.consumed
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.id = $1 FOR UPDATE
	`, tokenID).Scan(&email, &expiresAt, &consumed)
	if errors.Is(err, sql.ErrNoRows) {
		return RefreshSession{}, false, nil
	}
	if err != nil {
		return RefreshSession{}, false, err
	}

	if consumed || time.Now().UTC().After(expiresAt) {
		return RefreshSession{}, false, nil
	}

	tx.ExecContext(ctx, `UPDATE refresh_tokens SET consumed = true WHERE id = $1`, tokenID)
	return RefreshSession{Email: email, ExpiresAt: expiresAt}, true, tx.Commit()
}

func (r *PostgresRepository) BlacklistRefreshToken(ctx context.Context, tokenID string, at time.Time, expiresAt time.Time) error {
	if !at.Before(expiresAt) {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_token_blacklist (token_id, blacklisted_at, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token_id) DO NOTHING
	`, tokenID, at, expiresAt)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET consumed = true WHERE id = $1
	`, tokenID)
	return err
}

func (r *PostgresRepository) IsRefreshTokenBlacklisted(ctx context.Context, tokenID string) (bool, time.Duration, error) {
	var expiresAt time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT expires_at FROM refresh_token_blacklist
		WHERE token_id = $1 AND expires_at > NOW()
	`, tokenID).Scan(&expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return true, 0, nil
}

func (r *PostgresRepository) RecordAlertEvent(ctx context.Context, event string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO alert_events (event, created_at)
		VALUES ($1, NOW())
	`, event)
	return err
}

func (r *PostgresRepository) EnsureDefaultLedger(ctx context.Context, email string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM ledgers l
			JOIN users u ON u.id = l.user_id
			WHERE u.email = $1 AND l.is_default = true
		)
	`, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ledgers (id, user_id, name, is_default)
		SELECT gen_random_uuid(), u.id, 'Default Ledger', true
		FROM users u WHERE u.email = $1
	`, email)
	if err != nil {
		return false, err
	}

	return true, tx.Commit()
}
