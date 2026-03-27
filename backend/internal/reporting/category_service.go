package reporting

import (
	"context"
	"sort"
	"strings"
	"time"

	"xledger/backend/internal/accounting"
)

type CategoryQuery struct {
	From time.Time
	To   time.Time
}

type CategoryStatItem struct {
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
}

type CategoryResult struct {
	Items []CategoryStatItem `json:"items"`
}

type CategoryService struct{ repo *Repository }

func NewCategoryService(repo *Repository) *CategoryService { return &CategoryService{repo: repo} }

func (s *CategoryService) GetCategoryStats(ctx context.Context, userID string, query CategoryQuery) (CategoryResult, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return CategoryResult{}, &contractError{code: STAT_QUERY_INVALID}
	}
	if (!query.From.IsZero() && query.To.IsZero()) || (query.From.IsZero() && !query.To.IsZero()) || (!query.From.IsZero() && query.From.After(query.To)) {
		return CategoryResult{}, &contractError{code: STAT_QUERY_INVALID}
	}
	txnQuery := accounting.TransactionQuery{}
	if !query.From.IsZero() {
		txnQuery.OccurredFrom = query.From
		txnQuery.OccurredTo = query.To
	}
	txns, err := s.repo.ListTransactions(userID, txnQuery)
	if err != nil {
		return CategoryResult{}, err
	}
	agg := map[string]float64{}
	for _, txn := range txns {
		if txn.Type != accounting.TransactionTypeExpense {
			continue
		}
		name := strings.TrimSpace(txn.CategoryName)
		if name == "" {
			if historical, ok := s.repo.HistoricalCategoryName(ctx, userID, strings.TrimSpace(ptrString(txn.CategoryID))); ok {
				name = historical
			}
		}
		if name == "" {
			name = "Uncategorized"
		}
		agg[name] += txn.Amount
	}
	items := make([]CategoryStatItem, 0, len(agg))
	for name, amount := range agg {
		items = append(items, CategoryStatItem{CategoryName: name, Amount: amount})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CategoryName < items[j].CategoryName })
	return CategoryResult{Items: items}, nil
}

func ptrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
