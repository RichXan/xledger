package reporting

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"xledger/backend/internal/accounting"
)

type KeywordQuery struct {
	From  time.Time
	To    time.Time
	Limit int
}

type KeywordStatItem struct {
	Text   string  `json:"text"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

type KeywordResult struct {
	Items []KeywordStatItem `json:"items"`
}

type KeywordService struct{ repo *Repository }

func NewKeywordService(repo *Repository) *KeywordService { return &KeywordService{repo: repo} }

func (s *KeywordService) GetKeywordStats(ctx context.Context, userID string, query KeywordQuery) (KeywordResult, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return KeywordResult{}, &contractError{code: STAT_QUERY_INVALID}
	}
	if (!query.From.IsZero() && query.To.IsZero()) || (query.From.IsZero() && !query.To.IsZero()) || (!query.From.IsZero() && query.From.After(query.To)) {
		return KeywordResult{}, &contractError{code: STAT_QUERY_INVALID}
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	txnQuery := accounting.TransactionQuery{}
	if !query.From.IsZero() {
		txnQuery.OccurredFrom = query.From
		txnQuery.OccurredTo = query.To
	}
	txns, err := s.repo.ListTransactions(userID, txnQuery)
	if err != nil {
		return KeywordResult{}, err
	}

	type aggregate struct {
		amount float64
		count  int
	}
	aggregates := map[string]*aggregate{}
	for _, txn := range txns {
		if txn.Type != accounting.TransactionTypeExpense {
			continue
		}
		terms := uniqueTerms(append(tokenizeKeywords(txn.Memo), tokenizeKeywords(txn.CategoryName)...))
		if len(terms) == 0 {
			terms = []string{"Uncategorized"}
		}
		for _, term := range terms {
			item := aggregates[term]
			if item == nil {
				item = &aggregate{}
				aggregates[term] = item
			}
			item.amount += txn.Amount
			item.count++
		}
	}

	items := make([]KeywordStatItem, 0, len(aggregates))
	for text, aggregate := range aggregates {
		items = append(items, KeywordStatItem{Text: text, Amount: aggregate.amount, Count: aggregate.count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Amount == items[j].Amount {
			if items[i].Count == items[j].Count {
				return items[i].Text < items[j].Text
			}
			return items[i].Count > items[j].Count
		}
		return items[i].Amount > items[j].Amount
	})
	if len(items) > limit {
		items = items[:limit]
	}

	_ = ctx
	return KeywordResult{Items: items}, nil
}

func uniqueTerms(terms []string) []string {
	seen := map[string]bool{}
	unique := make([]string, 0, len(terms))
	for _, term := range terms {
		if seen[term] {
			continue
		}
		seen[term] = true
		unique = append(unique, term)
	}
	return unique
}

func tokenizeKeywords(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}
	}
	var builder strings.Builder
	terms := []string{}
	flush := func() {
		token := strings.TrimSpace(builder.String())
		builder.Reset()
		if token == "" || isKeywordStopWord(token) {
			return
		}
		terms = append(terms, token)
	}

	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return terms
}

func isKeywordStopWord(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return true
	}
	if _, err := strconv.ParseFloat(normalized, 64); err == nil {
		return true
	}
	switch normalized {
	case "memo", "note", "notes", "remark", "remarks", "expense", "income", "transfer",
		"备注", "支出", "收入", "转账", "无", "其它", "其他":
		return true
	default:
		return false
	}
}
