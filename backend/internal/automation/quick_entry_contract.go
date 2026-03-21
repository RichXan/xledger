package automation

import "xledger/backend/internal/accounting"

const (
	QE_PAT_INVALID     = "QE_PAT_INVALID"
	QE_LLM_UNAVAILABLE = "QE_LLM_UNAVAILABLE"
	QE_PARSE_FAILED    = "QE_PARSE_FAILED"
	QE_TIMEOUT         = "QE_TIMEOUT"
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

type QuickEntryRequest struct {
	UserID         string
	PATID          string
	IdempotencyKey string
	Text           string
	PATInvalid     bool
	LLMUnavailable bool
	ParseFailed    bool
	TimedOut       bool
	ParsedAmount   float64
	ParsedType     string
	AccountHint    string
}

type QuickEntryResult struct {
	TransactionID     string  `json:"transaction_id"`
	Amount            float64 `json:"amount"`
	Type              string  `json:"type"`
	AccountID         *string `json:"account_id,omitempty"`
	AccountHintStatus string  `json:"account_hint_status,omitempty"`
	WroteTransaction  bool    `json:"wrote_transaction"`
}

type transactionCreator interface {
	CreateForAutomation(userID string, input accounting.TransactionCreateInput) (accounting.Transaction, error)
}
