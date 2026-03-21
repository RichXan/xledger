package automation

import (
	"context"
	"strings"
	"sync"
	"time"

	"xledger/backend/internal/accounting"
)

type QuickEntryAdapter struct {
	mu              sync.Mutex
	now             func() time.Time
	creator         transactionCreator
	defaultLedgerID string
	idempotency     map[string]quickEntryJob
}

type quickEntryJob struct {
	createdAt time.Time
	result    QuickEntryResult
}

func NewQuickEntryAdapter(now func() time.Time, creator transactionCreator, defaultLedgerID string) *QuickEntryAdapter {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &QuickEntryAdapter{now: now, creator: creator, defaultLedgerID: strings.TrimSpace(defaultLedgerID), idempotency: map[string]quickEntryJob{}}
}

func (a *QuickEntryAdapter) SetNow(now func() time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if now != nil {
		a.now = now
	}
}

func (a *QuickEntryAdapter) Process(ctx context.Context, req QuickEntryRequest) (QuickEntryResult, error) {
	if req.PATInvalid {
		return QuickEntryResult{}, &contractError{code: QE_PAT_INVALID}
	}
	if req.LLMUnavailable {
		return QuickEntryResult{}, &contractError{code: QE_LLM_UNAVAILABLE}
	}
	if req.ParseFailed {
		return QuickEntryResult{}, &contractError{code: QE_PARSE_FAILED}
	}
	if req.TimedOut {
		return QuickEntryResult{}, &contractError{code: QE_TIMEOUT}
	}

	jobKey := a.idempotencyKey(req)
	if job, ok := a.loadJob(jobKey); ok {
		return job.result, nil
	}

	input := accounting.TransactionCreateInput{
		LedgerID:   a.defaultLedgerID,
		Type:       strings.TrimSpace(req.ParsedType),
		Amount:     req.ParsedAmount,
		OccurredAt: a.now(),
	}
	result := QuickEntryResult{Amount: req.ParsedAmount, Type: strings.TrimSpace(req.ParsedType)}
	if strings.TrimSpace(req.AccountHint) != "" {
		result.AccountHintStatus = "account_hint_ambiguous"
	}
	created, err := a.creator.CreateForAutomation(strings.TrimSpace(req.UserID), input)
	if err != nil {
		return QuickEntryResult{}, err
	}
	result.TransactionID = created.ID
	result.AccountID = created.AccountID
	result.WroteTransaction = true
	if result.AccountID != nil {
		result.AccountHintStatus = ""
	}
	a.storeJob(jobKey, result)
	_ = ctx
	return result, nil
}

func (a *QuickEntryAdapter) idempotencyKey(req QuickEntryRequest) string {
	return strings.TrimSpace(req.PATID) + "|" + normalizeText(req.Text) + "|" + strings.TrimSpace(req.IdempotencyKey)
}

func normalizeText(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

func (a *QuickEntryAdapter) loadJob(key string) (quickEntryJob, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	job, ok := a.idempotency[key]
	if !ok {
		return quickEntryJob{}, false
	}
	if a.now().Sub(job.createdAt) >= 24*time.Hour {
		delete(a.idempotency, key)
		return quickEntryJob{}, false
	}
	return job, true
}

func (a *QuickEntryAdapter) storeJob(key string, result QuickEntryResult) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.idempotency[key] = quickEntryJob{createdAt: a.now(), result: result}
}
