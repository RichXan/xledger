package budget

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateBudget(ctx context.Context, budget *Budget) error {
	budget.ID = generateID()
	budget.CreatedAt = time.Now()
	budget.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(budget).Error
}

func (r *PostgresRepository) GetBudget(ctx context.Context, id string) (*Budget, error) {
	var budget Budget
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&budget).Error
	if err != nil {
		return nil, err
	}
	return &budget, nil
}

func (r *PostgresRepository) ListBudgets(ctx context.Context, userID string) ([]Budget, error) {
	var budgets []Budget
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&budgets).Error
	return budgets, err
}

func (r *PostgresRepository) UpdateBudget(ctx context.Context, budget *Budget) error {
	budget.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(budget).Error
}

func (r *PostgresRepository) DeleteBudget(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Budget{}, "id = ?", id).Error
}

func (r *PostgresRepository) CreateAlert(ctx context.Context, alert *BudgetAlert) error {
	alert.ID = generateID()
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *PostgresRepository) ListAlerts(ctx context.Context, userID string, limit int) ([]BudgetAlert, error) {
	var alerts []BudgetAlert
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("triggered_at DESC").
		Limit(limit).
		Find(&alerts).Error
	return alerts, err
}

func (r *PostgresRepository) GetPreference(ctx context.Context, userID string) (*UserNotificationPref, error) {
	var pref UserNotificationPref
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&pref).Error
	if err == gorm.ErrRecordNotFound {
		return &UserNotificationPref{UserID: userID}, nil
	}
	return &pref, err
}

func (r *PostgresRepository) UpdatePreference(ctx context.Context, pref *UserNotificationPref) error {
	return r.db.WithContext(ctx).Save(pref).Error
}
