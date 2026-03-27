package reporting

import (
	"context"
	"sort"
	"strings"
	"time"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/common/timex"
)

type TrendQuery struct {
	From        time.Time
	To          time.Time
	Granularity string
	Timezone    string
	Timeout     time.Duration
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
	if query.Granularity != "day" && query.Granularity != "month" {
		return TrendResult{}, &contractError{code: STAT_QUERY_INVALID}
	}
	if query.To.Sub(query.From) > 370*24*time.Hour {
		return TrendResult{}, &contractError{code: STAT_QUERY_INVALID}
	}
	loc, ok := timex.ParseUserTZ(strings.TrimSpace(query.Timezone))
	if !ok {
		return TrendResult{}, &contractError{code: STAT_QUERY_INVALID}
	}
	txns, err := s.listTransactions(ctx, userID, accounting.TransactionQuery{OccurredFrom: query.From, OccurredTo: query.To}, query.Timeout)
	if err != nil {
		return TrendResult{}, err
	}

	fromBoundary := localizeTrendBoundary(query.From, loc)
	toBoundary := localizeTrendBoundary(query.To, loc)
	buckets := map[time.Time]*TrendPoint{}
	for current := truncateTrendBucket(fromBoundary, query.Granularity, loc); current.Before(toBoundary); current = advanceTrendBucket(current, query.Granularity) {
		copy := current
		buckets[current] = &TrendPoint{BucketStart: copy}
	}
	for _, txn := range txns {
		if txn.Type == accounting.TransactionTypeTransfer {
			continue
		}
		bucket := truncateTrendBucket(txn.OccurredAt, query.Granularity, loc)
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

func (s *TrendService) listTransactions(ctx context.Context, userID string, query accounting.TransactionQuery, timeout time.Duration) ([]accounting.Transaction, error) {
	if timeout <= 0 {
		return s.repo.ListTransactions(userID, query)
	}
	type result struct {
		txns []accounting.Transaction
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		txns, err := s.repo.ListTransactions(userID, query)
		ch <- result{txns: txns, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, &contractError{code: STAT_TIMEOUT}
	case res := <-ch:
		return res.txns, res.err
	case <-time.After(timeout):
		return nil, &contractError{code: STAT_TIMEOUT}
	}
}

func truncateTrendBucket(value time.Time, granularity string, loc *time.Location) time.Time {
	value = value.In(loc)
	switch granularity {
	case "month":
		return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, loc)
	case "day":
		return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, loc)
	default:
		return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, loc)
	}
}

func advanceTrendBucket(value time.Time, granularity string) time.Time {
	switch granularity {
	case "month":
		return value.AddDate(0, 1, 0)
	case "day":
		return value.Add(24 * time.Hour)
	default:
		return value.Add(24 * time.Hour)
	}
}

func localizeTrendBoundary(value time.Time, loc *time.Location) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), value.Hour(), value.Minute(), value.Second(), value.Nanosecond(), loc)
}
