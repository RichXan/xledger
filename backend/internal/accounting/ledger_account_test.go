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

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body["error_code"] != ACCOUNT_INVALID {
		t.Fatalf("expected error_code=%s, got %q", ACCOUNT_INVALID, body["error_code"])
	}
}
