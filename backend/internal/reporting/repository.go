package reporting

import (
	"context"
	"strings"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
)

type Repository struct {
	accountRepo     accounting.AccountRepository
	txnRepo         accounting.TransactionRepository
	categoryHistory categoryHistoryLookup
}

type categoryHistoryLookup interface {
	GetHistoricalCategoryName(ctx context.Context, userID string, categoryID string) (string, bool)
}

func NewRepository(accountRepo accounting.AccountRepository, txnRepo accounting.TransactionRepository, categoryHistory categoryHistoryLookup) *Repository {
	return &Repository{accountRepo: accountRepo, txnRepo: txnRepo, categoryHistory: categoryHistory}
}

func (r *Repository) ListAccounts(userID string) ([]accounting.Account, error) {
	return r.accountRepo.ListByUser(strings.TrimSpace(userID))
}

func (r *Repository) ListTransactions(userID string, query accounting.TransactionQuery) ([]accounting.Transaction, error) {
	return r.txnRepo.ListByUser(strings.TrimSpace(userID), query)
}

func (r *Repository) HistoricalCategoryName(ctx context.Context, userID string, categoryID string) (string, bool) {
	if r.categoryHistory == nil {
		return "", false
	}
	return r.categoryHistory.GetHistoricalCategoryName(ctx, userID, categoryID)
}

var _ categoryHistoryLookup = (*classification.CategoryService)(nil)
