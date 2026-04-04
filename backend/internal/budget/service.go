package budget

import (
    "context"
    "fmt"
    "time"

    "xledger/backend/internal/accounting"
)

type Service struct {
    repo             Repository
    txnService       *accounting.TransactionService
    alertService    *AlertService
}

func NewService(repo Repository, txnService *accounting.TransactionService) *Service {
    return &Service{
        repo:       repo,
        txnService: txnService,
    }
}

func (s *Service) CreateBudget(ctx context.Context, userID, categoryID string, amount float64, alertAt float64) (*Budget, error) {
    budget := &Budget{
        UserID:     userID,
        CategoryID: categoryID,
        Amount:     amount,
        Period:     "monthly",
        AlertAt:    alertAt,
    }
    if err := s.repo.CreateBudget(ctx, budget); err != nil {
        return nil, err
    }
    return budget, nil
}

func (s *Service) GetUserBudgets(ctx context.Context, userID string) ([]BudgetWithUsage, error) {
    budgets, err := s.repo.ListBudgets(ctx, userID)
    if err != nil {
        return nil, err
    }

    now := time.Now()
    startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
    endOfMonth := startOfMonth.AddDate(0, 1, 0)

    var result []BudgetWithUsage
    for _, b := range budgets {
        // Calculate spending for this category in current month
        spent, err := s.txnService.GetCategorySpentInPeriod(ctx, userID, b.CategoryID, startOfMonth, endOfMonth)
        if err != nil {
            spent = 0
        }

        remaining := b.Amount - spent
        if remaining < 0 {
            remaining = 0
        }

        percent := 0.0
        if b.Amount > 0 {
            percent = (spent / b.Amount) * 100
        }

        result = append(result, BudgetWithUsage{
            Budget:    b,
            Spent:     spent,
            Remaining: remaining,
            Percent:   percent,
        })
    }

    return result, nil
}

func (s *Service) UpdateBudget(ctx context.Context, id string, amount float64, alertAt float64) (*Budget, error) {
    budget, err := s.repo.GetBudget(ctx, id)
    if err != nil {
        return nil, err
    }
    budget.Amount = amount
    budget.AlertAt = alertAt
    if err := s.repo.UpdateBudget(ctx, budget); err != nil {
        return nil, err
    }
    return budget, nil
}

func (s *Service) DeleteBudget(ctx context.Context, id string) error {
    return s.repo.DeleteBudget(ctx, id)
}

func (s *Service) ListAlerts(ctx context.Context, userID string, limit int) ([]BudgetAlert, error) {
    if limit <= 0 {
        limit = 20
    }
    return s.repo.ListAlerts(ctx, userID, limit)
}

func (s *Service) GetPreference(ctx context.Context, userID string) (*UserNotificationPref, error) {
    return s.repo.GetPreference(ctx, userID)
}

func (s *Service) UpdatePreference(ctx context.Context, pref *UserNotificationPref) error {
    return s.repo.UpdatePreference(ctx, pref)
}

func (s *Service) CheckAndAlert(ctx context.Context, userID string, categoryID string, newAmount float64) {
    budgets, err := s.repo.ListBudgets(ctx, userID)
    if err != nil {
        return
    }

    now := time.Now()
    startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
    endOfMonth := startOfMonth.AddDate(0, 1, 0)

    for _, b := range budgets {
        if b.CategoryID != categoryID {
            continue
        }

        spent, err := s.txnService.GetCategorySpentInPeriod(ctx, userID, categoryID, startOfMonth, endOfMonth)
        if err != nil {
            continue
        }

        // Check if over threshold
        if b.AlertAt > 0 {
            percent := (spent / b.Amount) * 100
            if percent >= b.AlertAt {
                alert := &BudgetAlert{
                    UserID:       userID,
                    BudgetID:     b.ID,
                    TriggeredAt:  now,
                    AlertType:    "threshold",
                    SpentAmount:  spent,
                    BudgetAmount: b.Amount,
                    Message:      fmt.Sprintf("Budget alert: You've spent %.0f%% of your %s budget", percent, categoryID),
                }
                s.repo.CreateAlert(ctx, alert)
                if s.alertService != nil {
                    s.alertService.SendBudgetAlert(ctx, userID, alert)
                }
            }
        }

        // Check if over budget
        if spent > b.Amount {
            alert := &BudgetAlert{
                UserID:       userID,
                BudgetID:     b.ID,
                TriggeredAt:  now,
                AlertType:    "over_budget",
                SpentAmount:  spent,
                BudgetAmount: b.Amount,
                Message:      fmt.Sprintf("Over budget: You've exceeded your %s budget of %.2f", categoryID, b.Amount),
            }
            s.repo.CreateAlert(ctx, alert)
            if s.alertService != nil {
                s.alertService.SendBudgetAlert(ctx, userID, alert)
            }
        }
    }
}

func (s *Service) SetAlertService(alertService *AlertService) {
    s.alertService = alertService
}
