package portability

import (
	"context"
	"strings"
	"time"

	"xledger/backend/internal/accounting"
	"xledger/backend/internal/classification"
	"xledger/backend/internal/common/text"
)

type ShortcutAdapter struct {
	txnService    *accounting.TransactionService
	ledgerService *accounting.LedgerService
	catService    *classification.CategoryService
}

func NewShortcutAdapter(
	txnService *accounting.TransactionService,
	ledgerService *accounting.LedgerService,
	catService *classification.CategoryService,
) *ShortcutAdapter {
	return &ShortcutAdapter{
		txnService:    txnService,
		ledgerService: ledgerService,
		catService:    catService,
	}
}

func (a *ShortcutAdapter) CreateTransaction(ctx context.Context, userID string, input ShortcutTransactionInput) (ShortcutTransactionResult, error) {
	createInput := accounting.TransactionCreateInput{
		LedgerID:   input.LedgerID,
		CategoryID: input.CategoryID,
		Type:       input.Type,
		Amount:     input.Amount,
		Memo:       input.Memo,
		OccurredAt: input.OccurredAt,
	}

	result, err := a.txnService.CreateTransaction(ctx, userID, createInput)
	if err != nil {
		return ShortcutTransactionResult{}, err
	}

	return ShortcutTransactionResult{
		ID:     result.ID,
		Amount: result.Amount,
		Type:   result.Type,
		Memo:   result.Memo,
	}, nil
}

func (a *ShortcutAdapter) GetDefaultLedgerID(ctx context.Context, userID string) (string, error) {
	ledgers, err := a.ledgerService.ListLedgers(ctx, userID)
	if err != nil || len(ledgers) == 0 {
		return "", err
	}

	for _, l := range ledgers {
		if l.IsDefault {
			return l.ID, nil
		}
	}
	return ledgers[0].ID, nil
}

func (a *ShortcutAdapter) FindByName(ctx context.Context, userID string, name string) (string, error) {
	if a.catService == nil {
		return "", nil
	}

	categories, err := a.catService.ListCategories(ctx, userID)
	if err != nil {
		return "", err
	}

	name = strings.TrimSpace(name)
	for _, cat := range categories {
		if text.StripEmojiPrefix(cat.Name) == name || cat.Name == name {
			return cat.ID, nil
		}
	}
	return "", nil
}

func (a *ShortcutAdapter) ListNames(ctx context.Context, userID string) ([]string, error) {
	if a.catService == nil {
		return []string{}, nil
	}
	categories, err := a.catService.ListCategories(ctx, userID)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(categories))
	for _, cat := range categories {
		names = append(names, text.StripEmojiPrefix(cat.Name))
	}
	return names, nil
}

type InMemoryCallbackRecorder struct {
	callbacks map[string][]ShortcutCallback
}

func NewInMemoryCallbackRecorder() *InMemoryCallbackRecorder {
	return &InMemoryCallbackRecorder{
		callbacks: make(map[string][]ShortcutCallback),
	}
}

func (r *InMemoryCallbackRecorder) Record(_ context.Context, userID string, callback ShortcutCallback) error {
	callback.ShortcutID = time.Now().Format("20060102150405")
	r.callbacks[userID] = append(r.callbacks[userID], callback)
	return nil
}

func (r *InMemoryCallbackRecorder) List(userID string) []ShortcutCallback {
	return r.callbacks[userID]
}
