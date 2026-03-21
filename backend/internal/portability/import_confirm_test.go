package portability

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestImportConfirm_UsesIdempotencyBeforeTripleDedup(t *testing.T) {
	repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	service := NewImportConfirmService(repo)
	handler := NewHandler(NewImportPreviewService(), service, nil, nil)
	r := gin.New()
	r.POST("/import/csv/confirm", withUser("user-1"), handler.ImportConfirm)

	payload := `{"rows":[{"date":"2026-03-01","amount":12.5,"description":"lunch"},{"date":"2026-03-01","amount":12.5,"description":"lunch"}]}`
	first := performConfirm(t, r, payload, "import-1", "user-1", http.StatusOK)
	firstData := confirmDataMap(t, first)
	if firstData["success_count"].(float64) != 1 || firstData["skip_count"].(float64) != 1 {
		t.Fatalf("expected first import to write one row then dedup one row, got %#v", first)
	}
	second := performConfirm(t, r, payload, "import-1", "user-1", http.StatusConflict)
	if second["code"] != "BUSINESS_RULE_VIOLATION" {
		t.Fatalf("expected duplicate request business rule violation, got %#v", second)
	}
	if repo.StoredRowCount("user-1") != 1 {
		t.Fatalf("expected idempotency to prevent second-write amplification, got %d rows", repo.StoredRowCount("user-1"))
	}
}

func TestImportConfirm_ReturnsSuccessSkipFailBreakdown(t *testing.T) {
	repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	service := NewImportConfirmService(repo)
	result, err := service.Confirm("user-1", "import-2", ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}, {Date: "2026-03-01", Amount: 10, Description: "ok"}, {Date: "", Amount: 5, Description: "bad"}}})
	if err == nil {
		t.Fatalf("expected partial failure error, got nil result=%#v", result)
	}
	if result.SuccessCount != 1 || result.SkipCount != 1 || result.FailCount != 1 {
		t.Fatalf("expected 1/1/1 breakdown, got %#v", result)
	}
}

func TestImportConfirm_DuplicateRequest_ReturnsIMPORT_DUPLICATE_REQUEST(t *testing.T) {
	repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	service := NewImportConfirmService(repo)
	_, _ = service.Confirm("user-1", "import-3", ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}}})
	_, err := service.Confirm("user-1", "import-3", ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}}})
	if ErrorCode(err) != IMPORT_DUPLICATE_REQUEST {
		t.Fatalf("expected %s, got %q", IMPORT_DUPLICATE_REQUEST, ErrorCode(err))
	}
}

func TestImportConfirm_PartialFailure_ReturnsIMPORT_PARTIAL_FAILED(t *testing.T) {
	repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	service := NewImportConfirmService(repo)
	_, err := service.Confirm("user-1", "import-4", ImportConfirmRequest{Rows: []ImportRow{{Date: "", Amount: 10, Description: "missing date"}}})
	if ErrorCode(err) != IMPORT_PARTIAL_FAILED {
		t.Fatalf("expected %s, got %q", IMPORT_PARTIAL_FAILED, ErrorCode(err))
	}
}

func TestImportConfirm_ResponseIncludesPerRowReason(t *testing.T) {
	repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	service := NewImportConfirmService(repo)
	result, _ := service.Confirm("user-1", "import-5", ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}, {Date: "", Amount: 9, Description: "bad"}}})
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 row results, got %#v", result.Rows)
	}
	if result.Rows[1].Reason == "" {
		t.Fatalf("expected failure row reason, got %#v", result.Rows[1])
	}
}

func TestImportConfirm_AcceptsAccessAndPAT(t *testing.T) {
	payload := `{"rows":[{"date":"2026-03-01","amount":12.5,"description":"lunch"}]}`
	for _, tokenType := range []string{"access", "pat"} {
		repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
		service := NewImportConfirmService(repo)
		handler := NewHandler(NewImportPreviewService(), service, nil, nil)
		r := gin.New()
		r.POST("/import/csv/confirm", withUser("user-1"), handler.ImportConfirm)
		resp := performConfirm(t, r, payload, "confirm-"+tokenType, "user-1", http.StatusOK)
		if confirmDataMap(t, resp)["success_count"].(float64) != 1 {
			t.Fatalf("expected success for %s token, got %#v", tokenType, resp)
		}
	}
}

func TestImportConfirm_MissingIdempotencyKey_ReturnsBadRequest(t *testing.T) {
	repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	service := NewImportConfirmService(repo)
	handler := NewHandler(NewImportPreviewService(), service, nil, nil)
	r := gin.New()
	r.POST("/import/csv/confirm", withUser("user-1"), handler.ImportConfirm)
	payload := `{"rows":[{"date":"2026-03-01","amount":12.5,"description":"lunch"}]}`
	resp := performConfirm(t, r, payload, "", "user-1", http.StatusBadRequest)
	if resp["code"] != "VALIDATION_ERROR" {
		t.Fatalf("expected validation error for missing key, got %#v", resp)
	}
}

func TestImportConfirm_IdempotencyWindowBoundary24h(t *testing.T) {
	now := time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	repo := NewRepository(func() time.Time { return now })
	service := NewImportConfirmService(repo)
	_, err := service.Confirm("user-1", "import-6", ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}}})
	if err != nil {
		t.Fatalf("unexpected first confirm error: %v", err)
	}
	repo.SetNow(func() time.Time { return now.Add(24 * time.Hour) })
	result, err := service.Confirm("user-1", "import-6", ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-02", Amount: 20, Description: "next-day"}}})
	if err != nil {
		t.Fatalf("expected 24h boundary replay to be allowed, got %v", err)
	}
	if result.SuccessCount != 1 {
		t.Fatalf("expected new write after 24h boundary, got %#v", result)
	}
}

func confirmDataMap(t *testing.T, payload map[string]any) map[string]any {
	t.Helper()
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data map in %#v", payload)
	}
	return data
}

func performConfirm(t *testing.T, handler http.Handler, payload string, idempotencyKey string, userID string, wantStatus int) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/import/csv/confirm", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		req.Header.Set("X-Idempotency-Key", idempotencyKey)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("expected status %d, got %d body=%s", wantStatus, rec.Code, rec.Body.String())
	}
	var payloadMap map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payloadMap); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	return payloadMap
}
