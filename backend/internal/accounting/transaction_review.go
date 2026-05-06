package accounting

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	ReviewReasonUncategorized = "uncategorized"
	ReviewReasonDuplicate     = "duplicate"
	ReviewReasonLarge         = "large"
)

type TransactionReviewSummary struct {
	Review        int `json:"review"`
	Uncategorized int `json:"uncategorized"`
	Duplicates    int `json:"duplicates"`
	Large         int `json:"large"`
}

type TransactionReviewItem struct {
	Transaction Transaction `json:"transaction"`
	Reasons     []string    `json:"reasons"`
}

type TransactionReviewQuery struct {
	TransactionQuery
	Reason string
}

func (s *TransactionService) GetReviewSummary(ctx context.Context, userID string, query TransactionQuery) (TransactionReviewSummary, error) {
	items, err := s.ListTransactions(ctx, userID, query)
	if err != nil {
		return TransactionReviewSummary{}, err
	}
	duplicateIDs, duplicateGroups := buildReviewDuplicateIDs(items)
	summary := TransactionReviewSummary{Duplicates: duplicateGroups}
	for _, txn := range items {
		reasons := reviewReasonsForTransaction(txn, duplicateIDs)
		if len(reasons) > 0 {
			summary.Review++
		}
		for _, reason := range reasons {
			switch reason {
			case ReviewReasonUncategorized:
				summary.Uncategorized++
			case ReviewReasonLarge:
				summary.Large++
			}
		}
	}
	return summary, nil
}

func (s *TransactionService) ListReviewItems(ctx context.Context, userID string, query TransactionReviewQuery) ([]TransactionReviewItem, int, error) {
	reason := strings.TrimSpace(query.Reason)
	if reason != "" && reason != "all" && reason != ReviewReasonUncategorized && reason != ReviewReasonDuplicate && reason != ReviewReasonLarge {
		return nil, 0, &contractError{code: TXN_VALIDATION_FAILED}
	}
	listQuery := query.TransactionQuery
	page, pageSize := listQuery.Page, listQuery.PageSize
	listQuery.Page = 0
	listQuery.PageSize = 0
	items, err := s.ListTransactions(ctx, userID, listQuery)
	if err != nil {
		return nil, 0, err
	}
	duplicateIDs, _ := buildReviewDuplicateIDs(items)
	reviewItems := make([]TransactionReviewItem, 0, len(items))
	for _, txn := range items {
		reasons := reviewReasonsForTransaction(txn, duplicateIDs)
		if len(reasons) == 0 {
			continue
		}
		if reason != "" && reason != "all" && !reviewReasonsContain(reasons, reason) {
			continue
		}
		reviewItems = append(reviewItems, TransactionReviewItem{Transaction: txn, Reasons: reasons})
	}
	total := len(reviewItems)
	if page > 0 && pageSize > 0 {
		start := (page - 1) * pageSize
		if start >= len(reviewItems) {
			return []TransactionReviewItem{}, total, nil
		}
		end := start + pageSize
		if end > len(reviewItems) {
			end = len(reviewItems)
		}
		reviewItems = reviewItems[start:end]
	}
	return reviewItems, total, nil
}

func buildReviewDuplicateIDs(transactions []Transaction) (map[string]bool, int) {
	groups := map[string][]Transaction{}
	for _, txn := range transactions {
		key := reviewDuplicateKey(txn)
		groups[key] = append(groups[key], txn)
	}
	ids := map[string]bool{}
	groupCount := 0
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		groupCount++
		for _, txn := range group {
			ids[txn.ID] = true
		}
	}
	return ids, groupCount
}

func reviewReasonsForTransaction(txn Transaction, duplicateIDs map[string]bool) []string {
	reasons := []string{}
	if strings.TrimSpace(txn.CategoryName) == "" && txn.CategoryID == nil {
		reasons = append(reasons, ReviewReasonUncategorized)
	}
	if duplicateIDs[txn.ID] {
		reasons = append(reasons, ReviewReasonDuplicate)
	}
	if txn.Type == TransactionTypeExpense && math.Abs(txn.Amount) >= 1000 {
		reasons = append(reasons, ReviewReasonLarge)
	}
	return reasons
}

func reviewDuplicateKey(txn Transaction) string {
	memo := strings.ToLower(strings.TrimSpace(txn.Memo))
	category := strings.ToLower(strings.TrimSpace(txn.CategoryName))
	date := txn.OccurredAt.In(time.UTC).Format("2006-01-02")
	label := memo
	if label == "" {
		label = category
	}
	return fmt.Sprintf("%s|%.2f|%s|%s", txn.Type, math.Abs(txn.Amount), date, label)
}

func reviewReasonsContain(reasons []string, reason string) bool {
	for _, current := range reasons {
		if current == reason {
			return true
		}
	}
	return false
}
