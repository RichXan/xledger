package reporting

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/common/timex"
)

const trendCacheTTL = 5 * time.Minute

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

type TrendService struct {
	repo  *Repository
	cache Cache
}

// NewTrendService creates a TrendService. cache may be nil (no caching).
func NewTrendService(repo *Repository, cache Cache) *TrendService {
	return &TrendService{repo: repo, cache: cache}
}

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

	cacheKey := fmt.Sprintf("rep:trend:%s:%s:%s:%s:%s",
		userID, query.Granularity, query.Timezone,
		query.From.Format(time.RFC3339), query.To.Format(time.RFC3339),
	)

	// Cache-Aside: probe cache first
	if s.cache != nil {
		if data, ok, err := s.cache.Get(cacheKey); ok && err == nil {
			var result TrendResult
			if json.Unmarshal(data, &result) == nil {
				return result, nil
			}
		}
	}

	txnQuery := accounting.TransactionQuery{OccurredFrom: query.From, OccurredTo: query.To}

	// Goroutine+select for timeout protection
	type aggResult struct {
		rows []accounting.TrendRow
		err  error
	}
	ch := make(chan aggResult, 1)
	go func() {
		rows, err := s.repo.GetTrendStats(userID, txnQuery, query.Granularity, loc)
		ch <- aggResult{rows: rows, err: err}
	}()

	var aggRows []accounting.TrendRow
	if query.Timeout > 0 {
		select {
		case <-ctx.Done():
			return TrendResult{}, &contractError{code: STAT_TIMEOUT}
		case res := <-ch:
			if res.err != nil {
				return TrendResult{}, res.err
			}
			aggRows = res.rows
		case <-time.After(query.Timeout):
			return TrendResult{}, &contractError{code: STAT_TIMEOUT}
		}
	} else {
		res := <-ch
		if res.err != nil {
			return TrendResult{}, res.err
		}
		aggRows = res.rows
	}

	// Build all buckets (zero-fill missing ones so callers get a complete series)
	fromBoundary := localizeTrendBoundary(query.From, loc)
	toBoundary := localizeTrendBoundary(query.To, loc)
	buckets := map[time.Time]*TrendPoint{}
	for cur := truncateTrendBucket(fromBoundary, query.Granularity, loc); cur.Before(toBoundary); cur = advanceTrendBucket(cur, query.Granularity) {
		snap := cur
		buckets[snap] = &TrendPoint{BucketStart: snap}
	}

	// Fill from SQL results into the pre-built bucket map
	for _, row := range aggRows {
		bucket := truncateTrendBucket(row.BucketStart, query.Granularity, loc)
		if pt := buckets[bucket]; pt != nil {
			pt.Income = row.Income
			pt.Expense = row.Expense
		}
	}

	points := make([]TrendPoint, 0, len(buckets))
	for _, pt := range buckets {
		points = append(points, *pt)
	}
	// sort ascending by BucketStart
	for i := 0; i < len(points)-1; i++ {
		for j := i + 1; j < len(points); j++ {
			if points[j].BucketStart.Before(points[i].BucketStart) {
				points[i], points[j] = points[j], points[i]
			}
		}
	}

	result := TrendResult{Points: points}

	// Write-back to cache
	if s.cache != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = s.cache.Set(cacheKey, data, trendCacheTTL)
		}
	}

	_ = ctx
	return result, nil
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
