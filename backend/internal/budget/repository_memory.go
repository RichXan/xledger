package budget

import (
	"context"
	"sort"
	"sync"
	"time"
)

type InMemoryRepository struct {
	mu          sync.Mutex
	budgets     map[string]Budget
	alerts      map[string]BudgetAlert
	preferences map[string]UserNotificationPref
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		budgets:     map[string]Budget{},
		alerts:      map[string]BudgetAlert{},
		preferences: map[string]UserNotificationPref{},
	}
}

func (r *InMemoryRepository) CreateBudget(_ context.Context, budget *Budget) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	budget.ID = generateID()
	budget.CreatedAt = now
	budget.UpdatedAt = now
	r.budgets[budget.ID] = *budget
	return nil
}

func (r *InMemoryRepository) GetBudget(_ context.Context, id string) (*Budget, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	budget, ok := r.budgets[id]
	if !ok {
		return nil, nil
	}
	return &budget, nil
}

func (r *InMemoryRepository) ListBudgets(_ context.Context, userID string) ([]Budget, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]Budget, 0)
	for _, budget := range r.budgets {
		if budget.UserID == userID {
			items = append(items, budget)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func (r *InMemoryRepository) UpdateBudget(_ context.Context, budget *Budget) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	budget.UpdatedAt = time.Now().UTC()
	r.budgets[budget.ID] = *budget
	return nil
}

func (r *InMemoryRepository) DeleteBudget(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.budgets, id)
	return nil
}

func (r *InMemoryRepository) CreateAlert(_ context.Context, alert *BudgetAlert) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	alert.ID = generateID()
	if alert.TriggeredAt.IsZero() {
		alert.TriggeredAt = time.Now().UTC()
	}
	r.alerts[alert.ID] = *alert
	return nil
}

func (r *InMemoryRepository) ListAlerts(_ context.Context, userID string, limit int) ([]BudgetAlert, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := make([]BudgetAlert, 0)
	for _, alert := range r.alerts {
		if alert.UserID == userID {
			items = append(items, alert)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].TriggeredAt.After(items[j].TriggeredAt)
	})
	if limit > 0 && len(items) > limit {
		return items[:limit], nil
	}
	return items, nil
}

func (r *InMemoryRepository) GetPreference(_ context.Context, userID string) (*UserNotificationPref, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pref, ok := r.preferences[userID]
	if !ok {
		return &UserNotificationPref{UserID: userID, RealtimeAlert: true}, nil
	}
	return &pref, nil
}

func (r *InMemoryRepository) UpdatePreference(_ context.Context, pref *UserNotificationPref) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.preferences[pref.UserID] = *pref
	return nil
}
