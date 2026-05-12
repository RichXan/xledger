package budget

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateBudget(ctx context.Context, budget *Budget) error {
	now := time.Now().UTC()
	budget.ID = generateID()
	budget.CreatedAt = now
	budget.UpdatedAt = now
	return r.db.QueryRowContext(ctx, `
		INSERT INTO budgets (id, user_id, category_id, amount, period, alert_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, category_id, amount, period, alert_at, created_at, updated_at
	`, budget.ID, budget.UserID, budget.CategoryID, budget.Amount, budget.Period, budget.AlertAt, budget.CreatedAt, budget.UpdatedAt).Scan(
		&budget.ID, &budget.UserID, &budget.CategoryID, &budget.Amount, &budget.Period, &budget.AlertAt, &budget.CreatedAt, &budget.UpdatedAt,
	)
}

func (r *PostgresRepository) GetBudget(ctx context.Context, id string) (*Budget, error) {
	var budget Budget
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, category_id, amount, period, alert_at, created_at, updated_at
		FROM budgets
		WHERE id = $1
	`, id).Scan(&budget.ID, &budget.UserID, &budget.CategoryID, &budget.Amount, &budget.Period, &budget.AlertAt, &budget.CreatedAt, &budget.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &budget, nil
}

func (r *PostgresRepository) ListBudgets(ctx context.Context, userID string) ([]Budget, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, category_id, amount, period, alert_at, created_at, updated_at
		FROM budgets
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var budget Budget
		if scanErr := rows.Scan(&budget.ID, &budget.UserID, &budget.CategoryID, &budget.Amount, &budget.Period, &budget.AlertAt, &budget.CreatedAt, &budget.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}
		budgets = append(budgets, budget)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return budgets, nil
}

func (r *PostgresRepository) UpdateBudget(ctx context.Context, budget *Budget) error {
	budget.UpdatedAt = time.Now().UTC()
	return r.db.QueryRowContext(ctx, `
		UPDATE budgets
		SET amount = $2, alert_at = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, user_id, category_id, amount, period, alert_at, created_at, updated_at
	`, budget.ID, budget.Amount, budget.AlertAt, budget.UpdatedAt).Scan(
		&budget.ID, &budget.UserID, &budget.CategoryID, &budget.Amount, &budget.Period, &budget.AlertAt, &budget.CreatedAt, &budget.UpdatedAt,
	)
}

func (r *PostgresRepository) DeleteBudget(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM budgets WHERE id = $1`, id)
	return err
}

func (r *PostgresRepository) CreateAlert(ctx context.Context, alert *BudgetAlert) error {
	alert.ID = generateID()
	if alert.TriggeredAt.IsZero() {
		alert.TriggeredAt = time.Now().UTC()
	}
	return r.db.QueryRowContext(ctx, `
		INSERT INTO budget_alerts (id, user_id, budget_id, triggered_at, alert_type, spent_amount, budget_amount, message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, alert.ID, alert.UserID, alert.BudgetID, alert.TriggeredAt, alert.AlertType, alert.SpentAmount, alert.BudgetAmount, alert.Message).Scan(&alert.ID)
}

func (r *PostgresRepository) ListAlerts(ctx context.Context, userID string, limit int) ([]BudgetAlert, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, budget_id, triggered_at, alert_type, spent_amount, budget_amount, message
		FROM budget_alerts
		WHERE user_id = $1
		ORDER BY triggered_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []BudgetAlert
	for rows.Next() {
		var alert BudgetAlert
		if scanErr := rows.Scan(&alert.ID, &alert.UserID, &alert.BudgetID, &alert.TriggeredAt, &alert.AlertType, &alert.SpentAmount, &alert.BudgetAmount, &alert.Message); scanErr != nil {
			return nil, scanErr
		}
		alerts = append(alerts, alert)
	}
	return alerts, rows.Err()
}

func (r *PostgresRepository) GetPreference(ctx context.Context, userID string) (*UserNotificationPref, error) {
	var pref UserNotificationPref
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, realtime_alert, daily_digest, weekly_digest, push_endpoint, push_key
		FROM user_notification_prefs
		WHERE user_id = $1
	`, userID).Scan(&pref.UserID, &pref.RealtimeAlert, &pref.DailyDigest, &pref.WeeklyDigest, &pref.PushEndpoint, &pref.PushKey)
	if errors.Is(err, sql.ErrNoRows) {
		return &UserNotificationPref{UserID: userID, RealtimeAlert: true}, nil
	}
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

func (r *PostgresRepository) UpdatePreference(ctx context.Context, pref *UserNotificationPref) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_notification_prefs (user_id, realtime_alert, daily_digest, weekly_digest, push_endpoint, push_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id)
		DO UPDATE SET realtime_alert = EXCLUDED.realtime_alert,
			daily_digest = EXCLUDED.daily_digest,
			weekly_digest = EXCLUDED.weekly_digest,
			push_endpoint = EXCLUDED.push_endpoint,
			push_key = EXCLUDED.push_key
	`, pref.UserID, pref.RealtimeAlert, pref.DailyDigest, pref.WeeklyDigest, pref.PushEndpoint, pref.PushKey)
	return err
}
