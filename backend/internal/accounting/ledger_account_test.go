package accounting

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDeleteDefaultLedger_ReturnsLEDGER_DEFAULT_IMMUTABLE(t *testing.T) {
	repo := NewInMemoryLedgerRepository()
	service := NewLedgerService(repo)

	ledger, err := repo.Create("user-1", LedgerCreateInput{Name: "Default", IsDefault: true})
	if err != nil {
		t.Fatalf("seed default ledger: %v", err)
	}

	err = service.DeleteLedger(context.Background(), "user-1", ledger.ID)
	if ErrorCode(err) != LEDGER_DEFAULT_IMMUTABLE {
		t.Fatalf("expected %s, got %q", LEDGER_DEFAULT_IMMUTABLE, ErrorCode(err))
	}
}

func TestCreateSecondDefaultLedger_ReturnsLEDGER_INVALID(t *testing.T) {
	repo := NewInMemoryLedgerRepository()
	service := NewLedgerService(repo)

	_, err := service.CreateLedger(context.Background(), "user-1", LedgerCreateInput{Name: "Default", IsDefault: true})
	if err != nil {
		t.Fatalf("create first default ledger: %v", err)
	}

	_, err = service.CreateLedger(context.Background(), "user-1", LedgerCreateInput{Name: "Another Default", IsDefault: true})
	if ErrorCode(err) != LEDGER_INVALID {
		t.Fatalf("expected %s, got %q", LEDGER_INVALID, ErrorCode(err))
	}
}

func TestAccountCRUD_OwnershipScopedByUser(t *testing.T) {
	repo := NewInMemoryAccountRepository()
	service := NewAccountService(repo)

	created, err := service.CreateAccount(context.Background(), "owner-user", AccountCreateInput{
		Name:           "Cash",
		Type:           "asset",
		InitialBalance: 10,
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	if _, err := service.GetAccount(context.Background(), "owner-user", created.ID); err != nil {
		t.Fatalf("owner should read account: %v", err)
	}

	if _, err := service.GetAccount(context.Background(), "other-user", created.ID); ErrorCode(err) != ACCOUNT_NOT_FOUND {
		t.Fatalf("other user get expected %s, got %q", ACCOUNT_NOT_FOUND, ErrorCode(err))
	}

	_, err = service.UpdateAccount(context.Background(), "other-user", created.ID, AccountUpdateInput{Name: "Renamed"})
	if ErrorCode(err) != ACCOUNT_NOT_FOUND {
		t.Fatalf("other user update expected %s, got %q", ACCOUNT_NOT_FOUND, ErrorCode(err))
	}

	err = service.DeleteAccount(context.Background(), "other-user", created.ID)
	if ErrorCode(err) != ACCOUNT_NOT_FOUND {
		t.Fatalf("other user delete expected %s, got %q", ACCOUNT_NOT_FOUND, ErrorCode(err))
	}

	if _, err := service.GetAccount(context.Background(), "owner-user", created.ID); err != nil {
		t.Fatalf("owner account should still exist: %v", err)
	}
}

func TestAccountGet_NotFound_ReturnsACCOUNT_NOT_FOUND(t *testing.T) {
	repo := NewInMemoryAccountRepository()
	service := NewAccountService(repo)

	_, err := service.GetAccount(context.Background(), "user-1", "missing-account")
	if ErrorCode(err) != ACCOUNT_NOT_FOUND {
		t.Fatalf("expected %s, got %q", ACCOUNT_NOT_FOUND, ErrorCode(err))
	}
}

func TestAccountCreate_InvalidPayload_ReturnsACCOUNT_INVALID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountRepo := NewInMemoryAccountRepository()
	handler := NewHandler(nil, NewAccountService(accountRepo))

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	r.POST("/accounts", handler.CreateAccount)

	req := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(`{"name":"","type":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body["code"] != "VALIDATION_ERROR" {
		t.Fatalf("expected code=VALIDATION_ERROR, got %#v", body)
	}
	if body["data"] != nil {
		t.Fatalf("expected nil data, got %#v", body["data"])
	}
}

func TestListAccounts_IncludesCurrentBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ledgerRepo := NewInMemoryLedgerRepository()
	accountRepo := NewInMemoryAccountRepository()
	transactionRepo := NewInMemoryTransactionRepository()
	ledgerService := NewLedgerService(ledgerRepo)
	accountService := NewAccountService(accountRepo)
	transactionService := NewTransactionService(transactionRepo, ledgerRepo, accountRepo, nil, nil)

	ledger, err := ledgerService.CreateLedger(context.Background(), "user-1", LedgerCreateInput{Name: "Default", IsDefault: true})
	if err != nil {
		t.Fatalf("seed ledger: %v", err)
	}
	account, err := accountService.CreateAccount(context.Background(), "user-1", AccountCreateInput{
		Name:           "Cash",
		Type:           "cash",
		InitialBalance: 1000,
	})
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if _, err := transactionService.CreateTransaction(context.Background(), "user-1", TransactionCreateInput{
		LedgerID:  ledger.ID,
		AccountID: &account.ID,
		Type:      TransactionTypeExpense,
		Amount:    125,
	}); err != nil {
		t.Fatalf("seed transaction: %v", err)
	}

	handler := NewHandler(ledgerService, accountService, transactionService)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Next()
	})
	r.GET("/accounts", handler.ListAccounts)

	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			Items []struct {
				ID             string  `json:"id"`
				InitialBalance float64 `json:"initial_balance"`
				CurrentBalance float64 `json:"current_balance"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if len(body.Data.Items) != 1 {
		t.Fatalf("expected 1 account, got %#v", body.Data.Items)
	}
	if body.Data.Items[0].InitialBalance != 1000 {
		t.Fatalf("expected initial balance 1000, got %v", body.Data.Items[0].InitialBalance)
	}
	if body.Data.Items[0].CurrentBalance != 875 {
		t.Fatalf("expected current balance 875, got %v", body.Data.Items[0].CurrentBalance)
	}
}
