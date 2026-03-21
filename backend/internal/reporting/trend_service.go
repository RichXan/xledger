package reporting

import (
	"context"
	"sort"
	"strings"
	"time"

	"xledger/backend/internal/accounting"
)

type TrendQuery struct {
	From        time.Time
	To          time.Time
	Granularity string
}

type TrendPoint struct {
	BucketStart time.Time `json:"bucket_start"`
	Income      float64   `json:"income"`
	Expense     float64   `json:"expense"`
}

type TrendResult struct {
	Points []TrendPoint `json:"points"`
}

type TrendService struct{ repo *Repository }

func NewTrendService(repo *Repository) *TrendService { return &TrendService{repo: repo} }

func (s *TrendService) GetTrend(ctx context.Context, userID string, query TrendQuery) (TrendResult, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" || query.From.IsZero() || query.To.IsZero() || query.From.After(query.To) {
		return TrendResult{}, &contractError{code: STAT_QUERY_INVALID}
	}
	if strings.TrimSpace(query.Granularity) == "" {
		query.Granularity = "day"
	}
	txns, err := s.repo.ListTransactions(userID, accounting.TransactionQuery{OccurredFrom: query.From, OccurredTo: query.To})
	if err != nil {
		return TrendResult{}, err
	}

	buckets := map[time.Time]*TrendPoint{}
	for current := truncateTrendBucket(query.From, query.Granularity); current.Before(query.To); current = advanceTrendBucket(current, query.Granularity) {
		copy := current
		buckets[current] = &TrendPoint{BucketStart: copy}
	}
	for _, txn := range txns {
		if txn.Type == accounting.TransactionTypeTransfer {
			continue
		}
		bucket := truncateTrendBucket(txn.OccurredAt, query.Granularity)
		point := buckets[bucket]
		if point == nil {
			continue
		}
		switch txn.Type {
		case accounting.TransactionTypeIncome:
			point.Income += txn.Amount
		case accounting.TransactionTypeExpense:
			point.Expense += txn.Amount
		}
	}
	points := make([]TrendPoint, 0, len(buckets))
	for _, point := range buckets {
		points = append(points, *point)
	}
	sort.Slice(points, func(i, j int) bool { return points[i].BucketStart.Before(points[j].BucketStart) })
	_ = ctx
	return TrendResult{Points: points}, nil
}

func truncateTrendBucket(value time.Time, granularity string) time.Time {
	value = value.UTC()
	switch granularity {
	case "day":
		return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
	default:
		return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
	}
}

func advanceTrendBucket(value time.Time, granularity string) time.Time {
	switch granularity {
	case "day":
		return value.Add(24 * time.Hour)
	default:
		return value.Add(24 * time.Hour)
	}
}
