package budget

import "context"

type Repository interface {
	CreateBudget(ctx context.Context, budget *Budget) error
	GetBudget(ctx context.Context, id string) (*Budget, error)
	ListBudgets(ctx context.Context, userID string) ([]Budget, error)
	UpdateBudget(ctx context.Context, budget *Budget) error
	DeleteBudget(ctx context.Context, id string) error

	CreateAlert(ctx context.Context, alert *BudgetAlert) error
	ListAlerts(ctx context.Context, userID string, limit int) ([]BudgetAlert, error)
	GetPreference(ctx context.Context, userID string) (*UserNotificationPref, error)
	UpdatePreference(ctx context.Context, pref *UserNotificationPref) error
}
