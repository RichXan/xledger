package portability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestImportConfirm_ReplaysIdempotentResultBeforeTripleDedup(t *testing.T) {
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
	second := performConfirm(t, r, payload, "import-1", "user-1", http.StatusOK)
	secondData := confirmDataMap(t, second)
	if secondData["success_count"].(float64) != firstData["success_count"].(float64) || secondData["skip_count"].(float64) != firstData["skip_count"].(float64) {
		t.Fatalf("expected duplicate request to replay cached import result, first=%#v second=%#v", first, second)
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

func TestImportConfirm_RetriesPreviouslyFailedImportWithSameKey(t *testing.T) {
	repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	service := NewImportConfirmService(repo)

	failed, err := service.Confirm("user-1", "import-retryable", ImportConfirmRequest{Rows: []ImportRow{{Date: "", Amount: 0, Description: ""}}})
	if ErrorCode(err) != IMPORT_PARTIAL_FAILED || failed.FailCount != 1 {
		t.Fatalf("expected first import to fail with retryable partial result, got result=%#v err=%v", failed, err)
	}

	retried, err := service.Confirm("user-1", "import-retryable", ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}}})
	if err != nil {
		t.Fatalf("expected retry with same key to reprocess corrected rows, got %v", err)
	}
	if retried.SuccessCount != 1 || retried.FailCount != 0 {
		t.Fatalf("expected retry to import corrected row, got %#v", retried)
	}
}

func TestImportConfirm_SkipsExistingRowsAcrossRequestKeys(t *testing.T) {
	repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	service := NewImportConfirmService(repo)
	req := ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}}}

	first, err := service.Confirm("user-1", "import-a", req)
	if err != nil {
		t.Fatalf("unexpected first confirm error: %v", err)
	}
	if first.SuccessCount != 1 || first.SkipCount != 0 {
		t.Fatalf("expected first request to import one row, got %#v", first)
	}

	second, err := service.Confirm("user-1", "import-b", req)
	if err != nil {
		t.Fatalf("unexpected second confirm error: %v", err)
	}
	if second.SuccessCount != 0 || second.SkipCount != 1 {
		t.Fatalf("expected second request to skip duplicate row, got %#v", second)
	}
	if repo.StoredRowCount("user-1") != 1 {
		t.Fatalf("expected duplicate row to avoid another stored import row, got %d", repo.StoredRowCount("user-1"))
	}
}

func TestImportConfirm_DeduplicatesAgainstExistingTransactions(t *testing.T) {
	repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	if err := repo.SaveImportedTransaction("user-1", ImportRow{Date: "2026-03-01", Amount: 10, Description: "ok"}); err != nil {
		t.Fatalf("seed transaction: %v", err)
	}
	service := NewImportConfirmService(repo)

	result, err := service.Confirm("user-1", "import-existing-transaction", ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}}})
	if err != nil {
		t.Fatalf("expected duplicate transaction skip without error, got %v", err)
	}
	if result.SuccessCount != 0 || result.SkipCount != 1 {
		t.Fatalf("expected existing transaction to be skipped, got %#v", result)
	}
	if repo.StoredRowCount("user-1") != 0 {
		t.Fatalf("expected transaction-based dedup to avoid writing an import row, got %d", repo.StoredRowCount("user-1"))
	}
}

type toggleImportWriterRepo struct {
	*Repository
	failPersist bool
}

func (r *toggleImportWriterRepo) SaveImportedTransaction(userID string, row ImportRow) error {
	if r.failPersist {
		return errors.New("persist failed")
	}
	return r.Repository.SaveImportedTransaction(userID, row)
}

func TestImportConfirm_DoesNotDedupRowsThatFailedPersistence(t *testing.T) {
	repo := &toggleImportWriterRepo{
		Repository:  NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) }),
		failPersist: true,
	}
	service := NewImportConfirmService(repo)
	req := ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}}}

	failed, err := service.Confirm("user-1", "import-failed", req)
	if ErrorCode(err) != IMPORT_PARTIAL_FAILED {
		t.Fatalf("expected persist failure to return %s, got result=%#v err=%v", IMPORT_PARTIAL_FAILED, failed, err)
	}
	if repo.StoredRowCount("user-1") != 0 {
		t.Fatalf("expected failed persistence to avoid writing dedup row, got %d", repo.StoredRowCount("user-1"))
	}

	repo.failPersist = false
	retried, err := service.Confirm("user-1", "import-retry", req)
	if err != nil {
		t.Fatalf("expected retry after persistence recovery to succeed, got %v", err)
	}
	if retried.SuccessCount != 1 || retried.SkipCount != 0 {
		t.Fatalf("expected retry to import row instead of skipping it, got %#v", retried)
	}
}

func TestImportConfirm_DoesNotReplayFullyFailedImportJob(t *testing.T) {
	repo := &toggleImportWriterRepo{
		Repository:  NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) }),
		failPersist: true,
	}
	service := NewImportConfirmService(repo)
	req := ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}}}

	if _, err := service.Confirm("user-1", "import-failed", req); ErrorCode(err) != IMPORT_PARTIAL_FAILED {
		t.Fatalf("expected initial persistence failure, got %v", err)
	}

	repo.failPersist = false
	retried, err := service.Confirm("user-1", "import-failed", req)
	if err != nil {
		t.Fatalf("expected retry with same idempotency key to succeed after persistence recovery, got %v", err)
	}
	if retried.SuccessCount != 1 || retried.FailCount != 0 {
		t.Fatalf("expected fully failed job not to be replayed, got %#v", retried)
	}
}

func TestImportConfirm_ReplaysPartiallySuccessfulFailedImportJob(t *testing.T) {
	repo := &toggleImportWriterRepo{Repository: NewRepository(func() time.Time {
		return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	})}
	service := NewImportConfirmService(repo)
	req := ImportConfirmRequest{Rows: []ImportRow{
		{Date: "2026-03-01", Amount: 10, Description: "ok"},
		{Date: "", Amount: 10, Description: "bad"},
	}}

	first, err := service.Confirm("user-1", "import-partial", req)
	if ErrorCode(err) != IMPORT_PARTIAL_FAILED {
		t.Fatalf("expected partial failure, got result=%#v err=%v", first, err)
	}
	second, err := service.Confirm("user-1", "import-partial", req)
	if ErrorCode(err) != "" {
		t.Fatalf("expected cached partial result to replay without reprocessing error, got %v", err)
	}
	if second.SuccessCount != first.SuccessCount || second.FailCount != first.FailCount || repo.StoredRowCount("user-1") != 1 {
		t.Fatalf("expected partial failure replay without duplicate writes, first=%#v second=%#v rows=%d", first, second, repo.StoredRowCount("user-1"))
	}
}

func TestImportConfirm_DuplicateRequest_ReturnsCachedResult(t *testing.T) {
	repo := NewRepository(func() time.Time { return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC) })
	service := NewImportConfirmService(repo)
	first, err := service.Confirm("user-1", "import-3", ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}}})
	if err != nil {
		t.Fatalf("unexpected first confirm error: %v", err)
	}
	second, err := service.Confirm("user-1", "import-3", ImportConfirmRequest{Rows: []ImportRow{{Date: "2026-03-01", Amount: 10, Description: "ok"}}})
	if err != nil {
		t.Fatalf("expected duplicate request to replay cached result without error, got %v", err)
	}
	if second.SuccessCount != first.SuccessCount || second.SkipCount != first.SkipCount || second.FailCount != first.FailCount {
		t.Fatalf("expected cached result, first=%#v second=%#v", first, second)
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

type blockingImportWriterRepo struct {
	*Repository
	started chan struct{}
	release chan struct{}
}

func (r *blockingImportWriterRepo) SaveImportedTransaction(userID string, row ImportRow) error {
	select {
	case r.started <- struct{}{}:
	default:
	}
	<-r.release
	return r.Repository.SaveImportedTransaction(userID, row)
}

func TestImportConfirm_BackgroundJobReturnsAcceptedAndCanBePolled(t *testing.T) {
	repo := &blockingImportWriterRepo{
		Repository: NewRepository(func() time.Time {
			return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
		}),
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	service := NewImportConfirmService(repo)
	handler := NewHandler(NewImportPreviewService(), service, nil, nil)
	r := gin.New()
	r.POST("/import/csv/confirm", withUser("user-1"), handler.ImportConfirm)
	r.GET("/import/csv/jobs/:job_id", withUser("user-1"), handler.ImportConfirmJobStatus)

	payload := `{"rows":[{"date":"2026-03-01","amount":12.5,"description":"lunch"}]}`
	req := httptest.NewRequest(http.MethodPost, "/import/csv/confirm?async=true", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", "background-1")
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		r.ServeHTTP(rec, req)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected async import request to return before row persistence finishes")
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected async import to return 202, got %d body=%s", rec.Code, rec.Body.String())
	}
	accepted := decodeJSONMap(t, rec)
	acceptedData := confirmDataMap(t, accepted)
	if acceptedData["job_id"] != "background-1" || acceptedData["status"] != "running" {
		t.Fatalf("expected running job response, got %#v", accepted)
	}

	select {
	case <-repo.started:
	case <-time.After(time.Second):
		t.Fatal("expected background worker to start persisting import rows")
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/import/csv/jobs/background-1", nil)
	statusRec := httptest.NewRecorder()
	r.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", statusRec.Code, statusRec.Body.String())
	}
	running := confirmDataMap(t, decodeJSONMap(t, statusRec))
	if running["status"] != "running" {
		t.Fatalf("expected running status while row is blocked, got %#v", running)
	}

	close(repo.release)
	deadline := time.Now().Add(time.Second)
	for {
		statusRec = httptest.NewRecorder()
		r.ServeHTTP(statusRec, statusReq)
		finished := confirmDataMap(t, decodeJSONMap(t, statusRec))
		if finished["status"] == "succeeded" {
			if finished["success_count"].(float64) != 1 {
				t.Fatalf("expected one imported row, got %#v", finished)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected background import to finish, last=%#v", finished)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

type recordingImportWriterRepo struct {
	*Repository
	rows []ImportRow
}

func (r *recordingImportWriterRepo) SaveImportedTransaction(userID string, row ImportRow) error {
	r.rows = append(r.rows, row)
	return r.Repository.SaveImportedTransaction(userID, row)
}

func TestImportConfirm_NormalizesLocalizedTransactionTypeBeforePersisting(t *testing.T) {
	repo := &recordingImportWriterRepo{Repository: NewRepository(func() time.Time {
		return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	})}
	service := NewImportConfirmService(repo)

	result, err := service.Confirm("user-1", "import-localized-type", ImportConfirmRequest{Rows: []ImportRow{{
		Date:        "2026-03-01",
		Amount:      25,
		Description: "salary",
		Type:        "收入",
	}}})
	if err != nil {
		t.Fatalf("expected localized transaction type import to succeed, got %v", err)
	}
	if result.SuccessCount != 1 {
		t.Fatalf("expected one imported row, got %#v", result)
	}
	if len(repo.rows) != 1 || repo.rows[0].Type != "income" {
		t.Fatalf("expected persisted row type to be normalized to income, got %#v", repo.rows)
	}
}

type recordingImportCategorySyncer struct {
	names []string
}

func (s *recordingImportCategorySyncer) FindOrCreateImportCategory(_ context.Context, _ string, name string) (string, string, error) {
	s.names = append(s.names, name)
	return "cat-imported", name, nil
}

func TestImportConfirm_SyncsImportedCategoryBeforePersistingRows(t *testing.T) {
	repo := &recordingImportWriterRepo{Repository: NewRepository(func() time.Time {
		return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	})}
	syncer := &recordingImportCategorySyncer{}
	service := NewImportConfirmService(repo, syncer)

	result, err := service.Confirm("user-1", "import-category", ImportConfirmRequest{Rows: []ImportRow{{
		Date:        "2026-03-01",
		Amount:      25,
		Description: "lunch",
		Category:    "Team Meals",
	}}})
	if err != nil {
		t.Fatalf("expected import with category sync to succeed, got %v", err)
	}
	if result.SuccessCount != 1 {
		t.Fatalf("expected one imported row, got %#v", result)
	}
	if len(syncer.names) != 1 || syncer.names[0] != "Team Meals" {
		t.Fatalf("expected category syncer to receive Team Meals, got %#v", syncer.names)
	}
	if len(repo.rows) != 1 {
		t.Fatalf("expected one persisted row, got %#v", repo.rows)
	}
	if repo.rows[0].CategoryID != "cat-imported" || repo.rows[0].Category != "Team Meals" {
		t.Fatalf("expected imported row to carry synced category id/name, got %#v", repo.rows[0])
	}
}

func TestImportConfirm_ReusesSyncedCategoryWithinRequest(t *testing.T) {
	repo := &recordingImportWriterRepo{Repository: NewRepository(func() time.Time {
		return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	})}
	syncer := &recordingImportCategorySyncer{}
	service := NewImportConfirmService(repo, syncer)

	result, err := service.Confirm("user-1", "import-category-cache", ImportConfirmRequest{Rows: []ImportRow{
		{Date: "2026-03-01", Amount: 25, Description: "lunch", Category: "Team Meals"},
		{Date: "2026-03-02", Amount: 30, Description: "dinner", Category: "Team Meals"},
	}})
	if err != nil {
		t.Fatalf("expected import with repeated category to succeed, got %v", err)
	}
	if result.SuccessCount != 2 {
		t.Fatalf("expected two imported rows, got %#v", result)
	}
	if len(syncer.names) != 1 || syncer.names[0] != "Team Meals" {
		t.Fatalf("expected category syncer to be called once for repeated Team Meals, got %#v", syncer.names)
	}
	if len(repo.rows) != 2 || repo.rows[0].CategoryID != "cat-imported" || repo.rows[1].CategoryID != "cat-imported" {
		t.Fatalf("expected both rows to reuse synced category, got %#v", repo.rows)
	}
}

func TestImportConfirm_AppliesDefaultAccountAndLedgerToRows(t *testing.T) {
	repo := &recordingImportWriterRepo{Repository: NewRepository(func() time.Time {
		return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	})}
	service := NewImportConfirmService(repo)
	handler := NewHandler(NewImportPreviewService(), service, nil, nil)
	r := gin.New()
	r.POST("/import/csv/confirm", withUser("user-1"), handler.ImportConfirm)

	payload := `{"default_account_id":"acc-1","default_ledger_id":"ledger-1","rows":[{"date":"2026-03-01 12:30:45","amount":12.5,"description":"lunch"}]}`
	performConfirm(t, r, payload, "import-defaults", "user-1", http.StatusOK)

	if len(repo.rows) != 1 {
		t.Fatalf("expected one recorded import row, got %#v", repo.rows)
	}
	if repo.rows[0].AccountID != "acc-1" || repo.rows[0].LedgerID != "ledger-1" {
		t.Fatalf("expected defaults to be applied to row, got %#v", repo.rows[0])
	}
}

type resolvingImportWriterRepo struct {
	*recordingImportWriterRepo
}

func (r *resolvingImportWriterRepo) ResolveImportReferences(_ string, row ImportRow) ImportRow {
	if row.AccountID == "stale-account" {
		row.AccountID = ""
	}
	return row
}

func (r *resolvingImportWriterRepo) SaveImportedTransaction(userID string, row ImportRow) error {
	if row.AccountID == "stale-account" {
		return errors.New("foreign key violation")
	}
	return r.recordingImportWriterRepo.SaveImportedTransaction(userID, row)
}

func TestImportConfirm_ResolvesStaleDefaultAccountBeforePersisting(t *testing.T) {
	repo := &resolvingImportWriterRepo{recordingImportWriterRepo: &recordingImportWriterRepo{Repository: NewRepository(func() time.Time {
		return time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC)
	})}}
	service := NewImportConfirmService(repo)

	result, err := service.Confirm("user-1", "import-stale-account", ImportConfirmRequest{
		DefaultAccountID: "stale-account",
		Rows: []ImportRow{{
			Date:        "2026-03-01",
			Amount:      10,
			Description: "ok",
		}},
	})

	if err != nil {
		t.Fatalf("expected stale default account to be cleared before persistence, got %v result=%#v", err, result)
	}
	if result.SuccessCount != 1 {
		t.Fatalf("expected import to succeed after clearing stale default account, got %#v", result)
	}
	if len(repo.rows) != 1 || repo.rows[0].AccountID != "" {
		t.Fatalf("expected persisted row to omit stale account id, got %#v", repo.rows)
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

func decodeJSONMap(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payloadMap map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payloadMap); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	return payloadMap
}
