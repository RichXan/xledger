package portability

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"xledger/backend/internal/accounting"
)

const (
	EXPORT_INVALID_RANGE = "EXPORT_INVALID_RANGE"
	EXPORT_TIMEOUT       = "EXPORT_TIMEOUT"
)

type ExportQuery struct {
	Format  string
	From    time.Time
	To      time.Time
	Timeout time.Duration
}

type ExportRepository struct {
	items     []accounting.Transaction
	listFn    func(userID string) ([]accounting.Transaction, error)
	historyFn func(context.Context, string, string) (string, bool)
}

func NewExportRepository(txnRepo accounting.TransactionRepository, categoryHistory interface {
	GetHistoricalCategoryName(context.Context, string, string) (string, bool)
}) *ExportRepository {
	return &ExportRepository{
		listFn: func(userID string) ([]accounting.Transaction, error) {
			return txnRepo.ListByUser(userID, accounting.TransactionQuery{})
		},
		historyFn: categoryHistory.GetHistoricalCategoryName,
	}
}

type ExportService struct {
	repo *ExportRepository
}

func NewExportService(repo *ExportRepository) *ExportService { return &ExportService{repo: repo} }

func (s *ExportService) Export(ctx context.Context, userID string, query ExportQuery) (string, error) {
	format := strings.TrimSpace(strings.ToLower(query.Format))
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "json" {
		return "", &contractError{code: EXPORT_INVALID_RANGE}
	}
	if !query.From.IsZero() && !query.To.IsZero() && query.From.After(query.To) {
		return "", &contractError{code: EXPORT_INVALID_RANGE}
	}
	txns, err := s.list(ctx, userID, query.Timeout)
	if err != nil {
		return "", err
	}
	filtered := make([]accounting.Transaction, 0, len(txns))
	for _, txn := range txns {
		if !query.From.IsZero() && txn.OccurredAt.Before(query.From) {
			continue
		}
		if !query.To.IsZero() && txn.OccurredAt.After(query.To) {
			continue
		}
		if strings.TrimSpace(txn.CategoryName) == "" && s.repo.historyFn != nil {
			if historical, ok := s.repo.historyFn(ctx, userID, exportPtrString(txn.CategoryID)); ok {
				txn.CategoryName = historical
			}
		}
		filtered = append(filtered, txn)
	}
	if format == "json" {
		payload, err := json.Marshal(filtered)
		if err != nil {
			return "", err
		}
		return string(payload), nil
	}
	builder := &strings.Builder{}
	writer := csv.NewWriter(builder)
	if err := writer.Write([]string{"occurred_at", "amount", "type", "category_name"}); err != nil {
		return "", err
	}
	for _, txn := range filtered {
		if err := writer.Write([]string{txn.OccurredAt.UTC().Format(time.RFC3339), strconv.FormatFloat(txn.Amount, 'f', -1, 64), txn.Type, txn.CategoryName}); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func (s *ExportService) list(ctx context.Context, userID string, timeout time.Duration) ([]accounting.Transaction, error) {
	if timeout <= 0 {
		return s.rawList(userID)
	}
	type result struct {
		items []accounting.Transaction
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		items, err := s.rawList(userID)
		ch <- result{items: items, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, &contractError{code: EXPORT_TIMEOUT}
	case res := <-ch:
		return res.items, res.err
	case <-time.After(timeout):
		return nil, &contractError{code: EXPORT_TIMEOUT}
	}
}

func (s *ExportService) rawList(userID string) ([]accounting.Transaction, error) {
	if s.repo.listFn != nil {
		return s.repo.listFn(userID)
	}
	return append([]accounting.Transaction(nil), s.repo.items...), nil
}

func exportPtrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
