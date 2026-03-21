package reporting

import (
	"context"
	"strings"

	"xledger/backend/internal/accounting"
)

const (
	STAT_QUERY_INVALID = "STAT_QUERY_INVALID"
	STAT_TIMEOUT       = "STAT_TIMEOUT"
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
}

type OverviewResult struct {
	TotalAssets float64 `json:"total_assets"`
	Income      float64 `json:"income"`
	Expense     float64 `json:"expense"`
	Net         float64 `json:"net"`
}

type OverviewService struct{ repo *Repository }

func NewOverviewService(repo *Repository) *OverviewService { return &OverviewService{repo: repo} }

func (s *OverviewService) GetOverview(ctx context.Context, userID string, query OverviewQuery) (OverviewResult, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return OverviewResult{}, &contractError{code: STAT_QUERY_INVALID}
	}
	accounts, err := s.repo.ListAccounts(userID)
	if err != nil {
		return OverviewResult{}, err
	}
	txns, err := s.repo.ListTransactions(userID, accounting.TransactionQuery{LedgerID: strings.TrimSpace(query.LedgerID)})
	if err != nil {
		return OverviewResult{}, err
	}

	result := OverviewResult{}
	for _, account := range accounts {
		result.TotalAssets += account.InitialBalance
	}
	for _, txn := range txns {
		if txn.Type == accounting.TransactionTypeTransfer {
			continue
		}
		switch txn.Type {
		case accounting.TransactionTypeIncome:
			result.Income += txn.Amount
		case accounting.TransactionTypeExpense:
			result.Expense += txn.Amount
		}
	}
	result.Net = result.Income - result.Expense
	_ = ctx
	return result, nil
}
