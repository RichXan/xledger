package reporting

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"xledger/backend/internal/accounting"
)

const (
	STAT_QUERY_INVALID = "STAT_QUERY_INVALID"
	STAT_TIMEOUT       = "STAT_TIMEOUT"
	overviewCacheTTL   = 5 * time.Minute
)

type contractError struct{ code string }

func (e *contractError) Error() string { return e.code }

func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	if typed, ok := err.(*contractError); ok {
		return typed.code
	}
	return err.Error()
}

type OverviewQuery struct {
	LedgerID string
	From     time.Time
	To       time.Time
}

type OverviewResult struct {
	TotalAssets float64 `json:"total_assets"`
	Income      float64 `json:"income"`
	Expense     float64 `json:"expense"`
	Net         float64 `json:"net"`
}

type OverviewService struct {
	repo  *Repository
	cache Cache
}

// NewOverviewService creates an OverviewService. cache may be nil (no caching).
func NewOverviewService(repo *Repository, cache Cache) *OverviewService {
	return &OverviewService{repo: repo, cache: cache}
}

func (s *OverviewService) GetOverview(ctx context.Context, userID string, query OverviewQuery) (OverviewResult, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return OverviewResult{}, &contractError{code: STAT_QUERY_INVALID}
	}
	if (!query.From.IsZero() && query.To.IsZero()) || (query.From.IsZero() && !query.To.IsZero()) || (!query.From.IsZero() && query.From.After(query.To)) {
		return OverviewResult{}, &contractError{code: STAT_QUERY_INVALID}
	}

	cacheKey := fmt.Sprintf("rep:overview:%s:%s:%s:%s",
		userID, query.LedgerID,
		query.From.Format(time.RFC3339), query.To.Format(time.RFC3339),
	)

	// Cache-Aside: probe cache first
	if s.cache != nil {
		if data, ok, err := s.cache.Get(cacheKey); ok && err == nil {
			var result OverviewResult
			if json.Unmarshal(data, &result) == nil {
				return result, nil
			}
		}
	}

	// TotalAssets still requires account listing (no SQL agg today)
	var totalAssets float64
	if s.repo.accountRepo != nil {
		accounts, err := s.repo.ListAccounts(userID)
		if err != nil {
			return OverviewResult{}, err
		}
		for _, a := range accounts {
			totalAssets += a.InitialBalance
		}
	}

	// SQL aggregation for income/expense — O(1) DB round-trip
	txnQuery := accounting.TransactionQuery{LedgerID: strings.TrimSpace(query.LedgerID)}
	if !query.From.IsZero() {
		txnQuery.OccurredFrom = query.From
		txnQuery.OccurredTo = query.To
	}
	income, expense, err := s.repo.GetOverviewStats(userID, txnQuery)
	if err != nil {
		return OverviewResult{}, err
	}

	result := OverviewResult{
		TotalAssets: totalAssets,
		Income:      income,
		Expense:     expense,
		Net:         income - expense,
	}

	// Write-back
	if s.cache != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = s.cache.Set(cacheKey, data, overviewCacheTTL)
		}
	}

	_ = ctx
	return result, nil
}
