package portability

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type recordingShortcutTxnCreator struct {
	userID string
	input  ShortcutTransactionInput
}

func (r *recordingShortcutTxnCreator) CreateTransaction(_ context.Context, userID string, input ShortcutTransactionInput) (ShortcutTransactionResult, error) {
	r.userID = userID
	r.input = input
	return ShortcutTransactionResult{ID: "txn-1", Amount: input.Amount, Type: input.Type, Memo: input.Memo}, nil
}

type staticShortcutLookups struct{}

func (staticShortcutLookups) GetDefaultLedgerID(context.Context, string) (string, error) {
	return "ledger-default", nil
}

func (staticShortcutLookups) FindByName(_ context.Context, _ string, name string) (string, error) {
	if strings.EqualFold(strings.TrimSpace(name), "Food") {
		return "cat-food", nil
	}
	return "", nil
}

func (staticShortcutLookups) ListNames(context.Context, string) ([]string, error) {
	return []string{"Food", "Transport"}, nil
}

func performShortcutRequest(handler gin.HandlerFunc, userID string, payload string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/target", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/target", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeShortcutData(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var envelope struct {
		Code string         `json:"code"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	return envelope.Data
}

func TestShortcutGeneratePreviewConfirm_BindsCurrentUserAndConfiguredLedgerAccount(t *testing.T) {
	patService := NewPATService(func() time.Time { return time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC) })
	txn := &recordingShortcutTxnCreator{}
	handler := NewShortcutHandler(patService, txn, staticShortcutLookups{}, staticShortcutLookups{}, NewInMemoryCallbackRecorder())

	generate := performShortcutRequest(handler.GenerateShortcut, "user-a", `{
        "name":"Xledger OCR Test",
		"expires_in":90,
		"default_ledger_id":"ledger-daily",
		"default_account_id":"acct-wechat",
		"mode":"ocr_confirm"
	}`)
	if generate.Code != http.StatusOK {
		t.Fatalf("generate status=%d body=%s", generate.Code, generate.Body.String())
	}
	generated := decodeShortcutData(t, generate)
	shortcutID, _ := generated["shortcut_id"].(string)
	if shortcutID == "" {
		t.Fatalf("expected shortcut_id in generate response: %#v", generated)
	}

	preview := performShortcutRequest(handler.PreviewQuickAdd, "user-a", `{
		"shortcut_id":"`+shortcutID+`",
		"raw_text":"WeChat Pay\nMerchant: Luckin Coffee\nAmount 35.00\nTime 2026-06-01 12:30",
		"source":"ios_shortcuts_ocr",
		"idempotency_key":"qe-1"
	}`)
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	previewed := decodeShortcutData(t, preview)
	if previewed["amount"] != float64(35) || previewed["memo"] != "Luckin Coffee - WeChat Pay" {
		t.Fatalf("unexpected preview data: %#v", previewed)
	}
	ledgerSuggestions := previewed["ledger_suggestions"].([]any)
	accountSuggestions := previewed["account_suggestions"].([]any)
	if ledgerSuggestions[0].(map[string]any)["id"] != "ledger-daily" {
		t.Fatalf("expected configured ledger suggestion, got %#v", ledgerSuggestions)
	}
	if accountSuggestions[0].(map[string]any)["id"] != "acct-wechat" {
		t.Fatalf("expected configured account suggestion, got %#v", accountSuggestions)
	}

	confirm := performShortcutRequest(handler.ConfirmQuickAdd, "user-a", `{
		"shortcut_id":"`+shortcutID+`",
		"idempotency_key":"qe-1",
		"ledger_id":"ledger-daily",
		"account_id":"acct-wechat",
		"category_id":"cat-food",
		"type":"expense",
		"amount":35,
        "memo":"Luckin Coffee - WeChat Pay",
		"occurred_at":"2026-06-01T12:30:00+08:00"
	}`)
	if confirm.Code != http.StatusCreated {
		t.Fatalf("confirm status=%d body=%s", confirm.Code, confirm.Body.String())
	}
	if txn.userID != "user-a" {
		t.Fatalf("expected current user to be used, got %q", txn.userID)
	}
	if txn.input.LedgerID != "ledger-daily" || txn.input.AccountID == nil || *txn.input.AccountID != "acct-wechat" {
		t.Fatalf("expected configured ledger/account, got %#v", txn.input)
	}
}
