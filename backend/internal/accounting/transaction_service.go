package accounting

import (
	"context"
	"errors"
	"strings"
	"time"

	"xledger/backend/internal/classification"
)

const (
	TXN_VALIDATION_FAILED = "TXN_VALIDATION_FAILED"
	TXN_NOT_FOUND         = "TXN_NOT_FOUND"
	TXN_CONFLICT          = "TXN_CONFLICT"
)

type TransactionService struct {
	repo        TransactionRepository
	ledgerRepo  LedgerRepository
	accountRepo AccountRepository
	transferSvc *TransferService
	categorySvc transactionCategoryService
	tagSvc      transactionTagService
}

type transactionCategoryService interface {
	ValidateCategorySelectable(ctx context.Context, userID string, categoryID string) error
	RecordCategoryUsage(ctx context.Context, userID string, categoryID string) (string, error)
}

type transactionTagService interface {
	ValidateTagIDs(ctx context.Context, userID string, tagIDs []string) error
	ReplaceTransactionTags(ctx context.Context, userID string, transactionID string, tagIDs []string) error
	RemoveTransactionTags(ctx context.Context, userID string, transactionID string) error
	ListTransactionIDsByTag(ctx context.Context, userID string, tagID string) ([]string, error)
}

func NewTransactionService(repo TransactionRepository, ledgerRepo LedgerRepository, accountRepo AccountRepository, categorySvc transactionCategoryService, tagSvc transactionTagService) *TransactionService {
	return &TransactionService{
		repo:        repo,
		ledgerRepo:  ledgerRepo,
		accountRepo: accountRepo,
		transferSvc: NewTransferService(repo),
		categorySvc: categorySvc,
		tagSvc:      tagSvc,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, userID string, input TransactionCreateInput) (Transaction, error) {
	userID = strings.TrimSpace(userID)
	input.LedgerID = strings.TrimSpace(input.LedgerID)
	input.Type = strings.TrimSpace(input.Type)
	input.AccountID = normalizeOptionalID(input.AccountID)
	input.CategoryID = normalizeOptionalID(input.CategoryID)
	input.TagIDs = normalizeTagIDs(input.TagIDs)
	input.Memo = strings.TrimSpace(input.Memo)
	input.FromAccountID = normalizeOptionalID(input.FromAccountID)
	input.ToAccountID = normalizeOptionalID(input.ToAccountID)
	if userID == "" || input.LedgerID == "" {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}
	ledgerExists, ledgerErr := s.ledgerExists(userID, input.LedgerID)
	if ledgerErr != nil {
		return Transaction{}, ledgerErr
	}
	if !ledgerExists {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	switch input.Type {
	case TransactionTypeIncome, TransactionTypeExpense:
		accountExists, accountErr := s.accountOptionalExists(userID, input.AccountID)
		if accountErr != nil {
			return Transaction{}, accountErr
		}
		if !accountExists {
			return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
		}
	case TransactionTypeTransfer:
		if strings.TrimSpace(ptrString(input.FromAccountID)) == "" || strings.TrimSpace(ptrString(input.ToAccountID)) == "" {
			return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
		}
		fromExists, fromErr := s.accountRequiredExists(userID, input.FromAccountID)
		if fromErr != nil {
			return Transaction{}, fromErr
		}
		toExists, toErr := s.accountRequiredExists(userID, input.ToAccountID)
		if toErr != nil {
			return Transaction{}, toErr
		}
		if !fromExists || !toExists {
			return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
		}
	default:
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	if input.Amount <= 0 {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	}

	if err := s.validateClassification(ctx, userID, input.CategoryID, input.TagIDs); err != nil {
		return Transaction{}, err
	}

	created, err := s.repo.Create(userID, input)
	if err != nil {
		return Transaction{}, err
	}
	if err := s.applyTransactionClassification(ctx, userID, created.ID, &created, input.CategoryID, true, input.TagIDs, true); err != nil {
		_, _ = s.repo.DeleteByIDForUser(userID, created.ID)
		return Transaction{}, err
	}

	return created, nil
}

func (s *TransactionService) CreateForAutomation(userID string, input TransactionCreateInput) (Transaction, error) {
	return s.CreateTransaction(context.Background(), userID, input)
}

func (s *TransactionService) CreateTransfer(ctx context.Context, userID string, input TransactionTransferInput) (Transaction, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	input.LedgerID = strings.TrimSpace(input.LedgerID)
	input.FromLedgerID = normalizeOptionalID(input.FromLedgerID)
	input.ToLedgerID = normalizeOptionalID(input.ToLedgerID)
	input.FromAccountID = normalizeOptionalID(input.FromAccountID)
	input.ToAccountID = normalizeOptionalID(input.ToAccountID)

	if strings.TrimSpace(ptrString(input.FromAccountID)) == "" || strings.TrimSpace(ptrString(input.ToAccountID)) == "" {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	fromLedgerID := strings.TrimSpace(ptrString(input.FromLedgerID))
	if fromLedgerID == "" {
		fromLedgerID = input.LedgerID
	}
	toLedgerID := strings.TrimSpace(ptrString(input.ToLedgerID))
	if toLedgerID == "" {
		toLedgerID = input.LedgerID
	}
	if fromLedgerID == "" || toLedgerID == "" {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	fromLedgerExists, fromLedgerErr := s.ledgerExists(userID, fromLedgerID)
	if fromLedgerErr != nil {
		return Transaction{}, fromLedgerErr
	}
	toLedgerExists, toLedgerErr := s.ledgerExists(userID, toLedgerID)
	if toLedgerErr != nil {
		return Transaction{}, toLedgerErr
	}
	if !fromLedgerExists || !toLedgerExists {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	fromExists, fromErr := s.accountRequiredExists(userID, input.FromAccountID)
	if fromErr != nil {
		return Transaction{}, fromErr
	}
	toExists, toErr := s.accountRequiredExists(userID, input.ToAccountID)
	if toErr != nil {
		return Transaction{}, toErr
	}
	if !fromExists || !toExists {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}

	if input.Amount <= 0 {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	}

	created, ledgers, err := s.transferSvc.Create(userID,
		TransactionCreateInput{
			LedgerID:      fromLedgerID,
			Type:          TransactionTypeTransfer,
			FromAccountID: input.FromAccountID,
			ToAccountID:   input.ToAccountID,
			Amount:        input.Amount,
			OccurredAt:    input.OccurredAt,
		},
		TransactionCreateInput{
			LedgerID:      toLedgerID,
			Type:          TransactionTypeTransfer,
			FromAccountID: input.FromAccountID,
			ToAccountID:   input.ToAccountID,
			Amount:        input.Amount,
			OccurredAt:    input.OccurredAt,
		},
	)
	if err != nil {
		return Transaction{}, s.mapTransferError(err)
	}

	if recalcErr := s.recalculateForLedgers(userID, ledgers); recalcErr != nil {
		return Transaction{}, recalcErr
	}

	return created, nil
}

func (s *TransactionService) EditTransaction(ctx context.Context, userID string, txnID string, input TransactionEditInput) (Transaction, error) {
	userID = strings.TrimSpace(userID)
	txnID = strings.TrimSpace(txnID)
	if userID == "" || txnID == "" || input.Amount <= 0 {
		return Transaction{}, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if input.HasCategory {
		input.CategoryID = normalizeOptionalID(input.CategoryID)
	}
	if input.HasMemo {
		trimmed := strings.TrimSpace(ptrString(input.Memo))
		input.Memo = &trimmed
	}
	if input.HasTagIDs {
		input.TagIDs = normalizeTagIDs(input.TagIDs)
	}

	txn, found, err := s.repo.GetByIDForUser(userID, txnID)
	if err != nil {
		return Transaction{}, err
	}
	if !found {
		return Transaction{}, &contractError{code: TXN_NOT_FOUND}
	}

	if txn.Type == TransactionTypeTransfer {
		edited, ledgers, transferErr := s.transferSvc.Edit(userID, txnID, input.Amount, input.Version)
		if transferErr != nil {
			return Transaction{}, s.mapTransferError(transferErr)
		}
		if recalcErr := s.recalculateForLedgers(userID, ledgers); recalcErr != nil {
			return Transaction{}, recalcErr
		}
		return edited, nil
	}

	categoryID := txn.CategoryID
	if input.HasCategory {
		categoryID = input.CategoryID
	}
	tagIDs := []string{}
	if input.HasTagIDs {
		tagIDs = input.TagIDs
	}
	if err := s.validateClassification(ctx, userID, categoryID, tagIDsIfSet(input.HasTagIDs, tagIDs)); err != nil {
		return Transaction{}, err
	}

	txn.Amount = input.Amount
	if input.HasCategory {
		txn.CategoryID = cloneStringPtr(input.CategoryID)
	}
	if input.HasMemo {
		txn.Memo = strings.TrimSpace(ptrString(input.Memo))
	}
	if input.Version != nil && txn.Version != *input.Version {
		return Transaction{}, &contractError{code: TXN_CONFLICT}
	}
	txn.Version++
	updated, saved, saveErr := s.repo.SaveByIDForUser(userID, txnID, txn)
	if saveErr != nil {
		return Transaction{}, saveErr
	}
	if !saved {
		return Transaction{}, &contractError{code: TXN_NOT_FOUND}
	}
	if err := s.applyTransactionClassification(ctx, userID, txnID, &updated, updated.CategoryID, input.HasCategory, tagIDs, input.HasTagIDs); err != nil {
		return Transaction{}, err
	}

	if err := s.recalculateForLedger(userID, txn.LedgerID); err != nil {
		return Transaction{}, err
	}

	return updated, nil
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, userID string, txnID string, expectedVersion *int) error {
	userID = strings.TrimSpace(userID)
	txnID = strings.TrimSpace(txnID)
	if userID == "" || txnID == "" {
		return &contractError{code: TXN_VALIDATION_FAILED}
	}

	txn, found, err := s.repo.GetByIDForUser(userID, txnID)
	if err != nil {
		return err
	}
	if !found {
		return &contractError{code: TXN_NOT_FOUND}
	}

	if txn.Type == TransactionTypeTransfer {
		ledgers, transferErr := s.transferSvc.Delete(userID, txnID, expectedVersion)
		if transferErr != nil {
			return s.mapTransferError(transferErr)
		}
		return s.recalculateForLedgers(userID, ledgers)
	}

	if expectedVersion != nil && txn.Version != *expectedVersion {
		return &contractError{code: TXN_CONFLICT}
	}

	deleted, deleteErr := s.repo.DeleteByIDForUser(userID, txnID)
	if deleteErr != nil {
		return deleteErr
	}
	if !deleted {
		return &contractError{code: TXN_NOT_FOUND}
	}
	if s.tagSvc != nil {
		if err := s.tagSvc.RemoveTransactionTags(ctx, userID, txnID); err != nil {
			return err
		}
	}

	return s.recalculateForLedger(userID, txn.LedgerID)
}

func (s *TransactionService) ListTransactions(ctx context.Context, userID string, query TransactionQuery) ([]Transaction, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, &contractError{code: TXN_VALIDATION_FAILED}
	}
	query.LedgerID = strings.TrimSpace(query.LedgerID)
	query.AccountID = strings.TrimSpace(query.AccountID)
	query.CategoryID = strings.TrimSpace(query.CategoryID)
	query.TagID = strings.TrimSpace(query.TagID)
	if !query.OccurredFrom.IsZero() && !query.OccurredTo.IsZero() && query.OccurredFrom.After(query.OccurredTo) {
		return nil, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if query.Page < 0 || query.PageSize < 0 {
		return nil, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if query.Page == 0 && query.PageSize > 0 {
		return nil, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if query.Page > 0 && query.PageSize == 0 {
		return nil, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if query.TagID != "" {
		if s.tagSvc == nil {
			return []Transaction{}, nil
		}
		txnIDs, err := s.tagSvc.ListTransactionIDsByTag(ctx, userID, query.TagID)
		if err != nil {
			return nil, err
		}
		query.TransactionIDs = txnIDs
		query.UseTransactionIDs = true
	}
	return s.repo.ListByUser(userID, query)
}

func (s *TransactionService) validateClassification(ctx context.Context, userID string, categoryID *string, tagIDs []string) error {
	if s.categorySvc != nil && strings.TrimSpace(ptrString(categoryID)) != "" {
		if err := s.categorySvc.ValidateCategorySelectable(ctx, userID, ptrString(categoryID)); err != nil {
			return err
		}
	}
	if s.tagSvc != nil && len(tagIDs) > 0 {
		if err := s.tagSvc.ValidateTagIDs(ctx, userID, tagIDs); err != nil {
			return err
		}
	}
	return nil
}

func (s *TransactionService) applyTransactionClassification(ctx context.Context, userID string, txnID string, txn *Transaction, categoryID *string, hasCategory bool, tagIDs []string, hasTagIDs bool) error {
	if hasCategory {
		txn.CategoryName = ""
		if s.categorySvc != nil && strings.TrimSpace(ptrString(categoryID)) != "" {
			name, err := s.categorySvc.RecordCategoryUsage(ctx, userID, ptrString(categoryID))
			if err != nil {
				return err
			}
			txn.CategoryName = name
		}
		updated, saved, err := s.repo.SaveByIDForUser(userID, txnID, *txn)
		if err != nil {
			return err
		}
		if !saved {
			return &contractError{code: TXN_NOT_FOUND}
		}
		*txn = updated
	}
	if hasTagIDs && s.tagSvc != nil {
		if err := s.tagSvc.ReplaceTransactionTags(ctx, userID, txnID, tagIDs); err != nil {
			return err
		}
	}
	return nil
}

func (s *TransactionService) recalculateForLedger(userID string, ledgerID string) error {
	if err := s.repo.MarkBalancesRecalculated(userID, ledgerID); err != nil {
		return err
	}
	if err := s.repo.MarkStatsInputRecalculated(userID, ledgerID); err != nil {
		return err
	}
	return nil
}

func (s *TransactionService) recalculateForLedgers(userID string, ledgerIDs []string) error {
	unique := make([]string, 0, len(ledgerIDs))
	for _, ledgerID := range ledgerIDs {
		trimmed := strings.TrimSpace(ledgerID)
		if trimmed == "" {
			continue
		}
		unique = appendUnique(unique, trimmed)
	}
	for _, ledgerID := range unique {
		if err := s.recalculateForLedger(userID, ledgerID); err != nil {
			return err
		}
	}
	return nil
}

func (s *TransactionService) mapTransferError(err error) error {
	switch {
	case errors.Is(err, errTransferConflict), errors.Is(err, errTransferVersionConflict), errors.Is(err, errTransferBilateralMismatch):
		return &contractError{code: TXN_CONFLICT}
	case errors.Is(err, errTransferNotFound):
		return &contractError{code: TXN_NOT_FOUND}
	default:
		return err
	}
}

func ptrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *TransactionService) ledgerExists(userID string, ledgerID string) (bool, error) {
	if s.ledgerRepo == nil {
		return true, nil
	}
	_, found, err := s.ledgerRepo.GetByIDForUser(userID, ledgerID)
	if err != nil {
		return false, err
	}
	return found, nil
}

func (s *TransactionService) accountOptionalExists(userID string, accountID *string) (bool, error) {
	if strings.TrimSpace(ptrString(accountID)) == "" {
		return true, nil
	}
	return s.accountRequiredExists(userID, accountID)
}

func (s *TransactionService) accountRequiredExists(userID string, accountID *string) (bool, error) {
	if s.accountRepo == nil {
		return true, nil
	}
	account := strings.TrimSpace(ptrString(accountID))
	if account == "" {
		return false, nil
	}
	_, found, err := s.accountRepo.GetByIDForUser(userID, account)
	if err != nil {
		return false, err
	}
	return found, nil
}

func normalizeOptionalID(value *string) *string {
	trimmed := strings.TrimSpace(ptrString(value))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeTagIDs(tagIDs []string) []string {
	if tagIDs == nil {
		return nil
	}
	normalized := make([]string, 0, len(tagIDs))
	seen := map[string]bool{}
	for _, tagID := range tagIDs {
		trimmed := strings.TrimSpace(tagID)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func tagIDsIfSet(hasTagIDs bool, tagIDs []string) []string {
	if !hasTagIDs {
		return nil
	}
	return tagIDs
}

func (s *TransactionService) ListTransactionsWithTotal(ctx context.Context, userID string, query TransactionQuery) ([]Transaction, int, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, 0, &contractError{code: TXN_VALIDATION_FAILED}
	}
	query.LedgerID = strings.TrimSpace(query.LedgerID)
	query.AccountID = strings.TrimSpace(query.AccountID)
	query.CategoryID = strings.TrimSpace(query.CategoryID)
	query.TagID = strings.TrimSpace(query.TagID)
	if !query.OccurredFrom.IsZero() && !query.OccurredTo.IsZero() && query.OccurredFrom.After(query.OccurredTo) {
		return nil, 0, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if query.Page < 0 || query.PageSize < 0 {
		return nil, 0, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if query.Page == 0 && query.PageSize > 0 {
		return nil, 0, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if query.Page > 0 && query.PageSize == 0 {
		return nil, 0, &contractError{code: TXN_VALIDATION_FAILED}
	}
	if query.TagID != "" {
		if s.tagSvc == nil {
			return []Transaction{}, 0, nil
		}
		txnIDs, err := s.tagSvc.ListTransactionIDsByTag(ctx, userID, query.TagID)
		if err != nil {
			return nil, 0, err
		}
		query.TransactionIDs = txnIDs
		query.UseTransactionIDs = true
	}

	total, err := s.repo.CountByUser(userID, query)
	if err != nil {
		return nil, 0, err
	}

	items, err := s.repo.ListByUser(userID, query)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetCategorySpentInPeriod returns total spending for a category within a date range.
func (s *TransactionService) GetCategorySpentInPeriod(ctx context.Context, userID, categoryID string, start, end time.Time) (float64, error) {
	query := TransactionQuery{
		CategoryID:  categoryID,
		OccurredFrom: start,
		OccurredTo:   end,
	}
	txns, err := s.ListTransactions(ctx, userID, query)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, t := range txns {
		if t.Type == "expense" {
			total += t.Amount
		}
	}
	return total, nil
}

var _ transactionCategoryService = (*classification.CategoryService)(nil)
var _ transactionTagService = (*classification.TagService)(nil)
