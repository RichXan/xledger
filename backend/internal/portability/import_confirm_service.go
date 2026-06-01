package portability

import (
	"context"
	"strings"
	"time"
)

const (
	IMPORT_DUPLICATE_REQUEST = "IMPORT_DUPLICATE_REQUEST"
	IMPORT_PARTIAL_FAILED    = "IMPORT_PARTIAL_FAILED"

	IMPORT_JOB_RUNNING   = "running"
	IMPORT_JOB_SUCCEEDED = "succeeded"
	IMPORT_JOB_FAILED    = "failed"
)

type ImportRow struct {
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Type        string  `json:"type,omitempty"`
	Category    string  `json:"category,omitempty"`
	CategoryID  string  `json:"category_id,omitempty"`
	Account     string  `json:"account,omitempty"`
	Ledger      string  `json:"ledger,omitempty"`
	AccountID   string  `json:"account_id,omitempty"`
	LedgerID    string  `json:"ledger_id,omitempty"`
}

type ImportConfirmRequest struct {
	DefaultAccountID string      `json:"default_account_id,omitempty"`
	DefaultLedgerID  string      `json:"default_ledger_id,omitempty"`
	Rows             []ImportRow `json:"rows"`
}

type ImportConfirmRowResult struct {
	RowIndex int    `json:"row_index"`
	Status   string `json:"status"`
	Reason   string `json:"reason,omitempty"`
}

type ImportConfirmResponse struct {
	SuccessCount int                      `json:"success_count"`
	SkipCount    int                      `json:"skip_count"`
	FailCount    int                      `json:"fail_count"`
	Rows         []ImportConfirmRowResult `json:"rows"`
}

type ImportJobStatusResponse struct {
	JobID         string                   `json:"job_id"`
	Status        string                   `json:"status"`
	TotalRows     int                      `json:"total_rows"`
	ProcessedRows int                      `json:"processed_rows"`
	SuccessCount  int                      `json:"success_count"`
	SkipCount     int                      `json:"skip_count"`
	FailCount     int                      `json:"fail_count"`
	Rows          []ImportConfirmRowResult `json:"rows,omitempty"`
	ErrorCode     string                   `json:"error_code,omitempty"`
}

func normalizeImportTransactionType(value string) string {
	if normalized := normalizeImportType(value); normalized != "" {
		return normalized
	}
	return "expense"
}

func shouldReplayImportJob(job importJob) bool {
	return strings.TrimSpace(job.ErrorCode) == "" || job.Response.SuccessCount > 0 || job.Response.SkipCount > 0
}

type ImportConfirmRepository interface {
	FindJob(userID string, path string, idempotencyKey string) (importJob, bool)
	SaveJob(job importJob)
	HasTriple(userID string, row storedImportRow) bool
	SaveRow(userID string, row storedImportRow)
	StoredRowCount(userID string) int
	Now() time.Time
}

type ImportConfirmService struct {
	repo           ImportConfirmRepository
	categorySyncer ImportCategorySyncer
}

type importReferenceResolver interface {
	ResolveImportReferences(userID string, row ImportRow) ImportRow
}

type ImportCategorySyncer interface {
	FindOrCreateImportCategory(ctx context.Context, userID string, name string) (id string, displayName string, err error)
}

func NewImportConfirmService(repo ImportConfirmRepository, categorySyncer ...ImportCategorySyncer) *ImportConfirmService {
	service := &ImportConfirmService{repo: repo}
	if len(categorySyncer) > 0 {
		service.categorySyncer = categorySyncer[0]
	}
	return service
}

func (s *ImportConfirmService) Confirm(userID string, idempotencyKey string, req ImportConfirmRequest) (ImportConfirmResponse, error) {
	return s.confirm(context.Background(), userID, idempotencyKey, req)
}

func (s *ImportConfirmService) StartBackgroundConfirm(userID string, idempotencyKey string, req ImportConfirmRequest) (ImportJobStatusResponse, error) {
	userID = strings.TrimSpace(userID)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		return ImportJobStatusResponse{}, &contractError{code: IMPORT_DUPLICATE_REQUEST}
	}
	if job, found := s.repo.FindJob(userID, "/api/import/csv/confirm", idempotencyKey); found {
		status := normalizeImportJobStatus(job)
		if status == IMPORT_JOB_RUNNING || status == IMPORT_JOB_SUCCEEDED {
			return importJobStatusResponse(job), nil
		}
	}

	job := importJob{
		UserID:         userID,
		Path:           "/api/import/csv/confirm",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      s.repo.Now(),
		Response:       ImportConfirmResponse{Rows: []ImportConfirmRowResult{}},
		Status:         IMPORT_JOB_RUNNING,
		TotalRows:      len(req.Rows),
		ProcessedRows:  0,
	}
	s.repo.SaveJob(job)
	go func() {
		_, _ = s.confirm(context.Background(), userID, idempotencyKey, req)
	}()
	return importJobStatusResponse(job), nil
}

func (s *ImportConfirmService) GetJobStatus(userID string, idempotencyKey string) (ImportJobStatusResponse, bool) {
	job, found := s.repo.FindJob(strings.TrimSpace(userID), "/api/import/csv/confirm", strings.TrimSpace(idempotencyKey))
	if !found {
		return ImportJobStatusResponse{}, false
	}
	return importJobStatusResponse(job), true
}

func (s *ImportConfirmService) confirm(ctx context.Context, userID string, idempotencyKey string, req ImportConfirmRequest) (ImportConfirmResponse, error) {
	userID = strings.TrimSpace(userID)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		return ImportConfirmResponse{}, &contractError{code: IMPORT_DUPLICATE_REQUEST}
	}
	if job, found := s.repo.FindJob(userID, "/api/import/csv/confirm", idempotencyKey); found && normalizeImportJobStatus(job) != IMPORT_JOB_RUNNING && shouldReplayImportJob(job) {
		return job.Response, nil
	}
	result := ImportConfirmResponse{Rows: make([]ImportConfirmRowResult, 0, len(req.Rows))}
	totalRows := len(req.Rows)
	s.repo.SaveJob(importJob{
		UserID:         userID,
		Path:           "/api/import/csv/confirm",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      s.repo.Now(),
		Response:       result,
		Status:         IMPORT_JOB_RUNNING,
		TotalRows:      totalRows,
		ProcessedRows:  0,
	})
	txnWriter, hasTxnWriter := s.repo.(interface {
		SaveImportedTransaction(userID string, row ImportRow) error
	})
	syncedCategories := map[string]struct {
		id   string
		name string
	}{}
	saveRunningProgress := func(processedRows int) {
		if totalRows == 0 || (processedRows != totalRows && processedRows%25 != 0) {
			return
		}
		s.repo.SaveJob(importJob{
			UserID:         userID,
			Path:           "/api/import/csv/confirm",
			IdempotencyKey: idempotencyKey,
			CreatedAt:      s.repo.Now(),
			Response:       result,
			Status:         IMPORT_JOB_RUNNING,
			TotalRows:      totalRows,
			ProcessedRows:  processedRows,
		})
	}
	for idx, row := range req.Rows {
		processedRows := idx + 1
		if strings.TrimSpace(row.AccountID) == "" {
			row.AccountID = strings.TrimSpace(req.DefaultAccountID)
		}
		if strings.TrimSpace(row.LedgerID) == "" {
			row.LedgerID = strings.TrimSpace(req.DefaultLedgerID)
		}
		if resolver, ok := s.repo.(importReferenceResolver); ok {
			row = resolver.ResolveImportReferences(userID, row)
		}
		trimmedDate := strings.TrimSpace(row.Date)
		trimmedDescription := strings.TrimSpace(row.Description)
		if trimmedDate == "" || row.Amount <= 0 || trimmedDescription == "" {
			result.FailCount++
			result.Rows = append(result.Rows, ImportConfirmRowResult{RowIndex: idx, Status: "failed", Reason: "invalid_row"})
			saveRunningProgress(processedRows)
			continue
		}
		normalizedType := normalizeImportTransactionType(row.Type)
		row.Type = normalizedType
		stored := storedImportRow{Date: trimmedDate, Amount: row.Amount, Description: trimmedDescription, Type: normalizedType}
		if s.repo.HasTriple(userID, stored) {
			result.SkipCount++
			result.Rows = append(result.Rows, ImportConfirmRowResult{RowIndex: idx, Status: "skipped", Reason: "duplicate_transaction"})
			saveRunningProgress(processedRows)
			continue
		}
		if s.categorySyncer != nil && strings.TrimSpace(row.Category) != "" && strings.TrimSpace(row.CategoryID) == "" {
			categoryKey := strings.ToLower(strings.TrimSpace(row.Category))
			if synced, ok := syncedCategories[categoryKey]; ok {
				row.CategoryID = synced.id
				row.Category = synced.name
			} else {
				categoryID, categoryName, err := s.categorySyncer.FindOrCreateImportCategory(ctx, userID, row.Category)
				if err != nil {
					result.FailCount++
					result.Rows = append(result.Rows, ImportConfirmRowResult{RowIndex: idx, Status: "failed", Reason: "persist_failed"})
					saveRunningProgress(processedRows)
					continue
				}
				row.CategoryID = strings.TrimSpace(categoryID)
				if strings.TrimSpace(categoryName) != "" {
					row.Category = strings.TrimSpace(categoryName)
				}
				syncedCategories[categoryKey] = struct {
					id   string
					name string
				}{id: row.CategoryID, name: row.Category}
			}
		}
		if hasTxnWriter {
			if err := txnWriter.SaveImportedTransaction(userID, row); err != nil {
				result.FailCount++
				result.Rows = append(result.Rows, ImportConfirmRowResult{RowIndex: idx, Status: "failed", Reason: "persist_failed"})
				saveRunningProgress(processedRows)
				continue
			}
		}
		s.repo.SaveRow(userID, stored)
		result.SuccessCount++
		result.Rows = append(result.Rows, ImportConfirmRowResult{RowIndex: idx, Status: "success"})
		saveRunningProgress(processedRows)
	}
	errCode := ""
	status := IMPORT_JOB_SUCCEEDED
	if result.FailCount > 0 {
		errCode = IMPORT_PARTIAL_FAILED
		status = IMPORT_JOB_FAILED
	}
	s.repo.SaveJob(importJob{UserID: userID, Path: "/api/import/csv/confirm", IdempotencyKey: idempotencyKey, CreatedAt: s.repo.Now(), Response: result, ErrorCode: errCode, Status: status, TotalRows: totalRows, ProcessedRows: totalRows})
	if errCode != "" {
		return result, &contractError{code: errCode}
	}
	return result, nil
}

func (s *ImportConfirmService) ConfirmContext(ctx context.Context, userID string, idempotencyKey string, req ImportConfirmRequest) (ImportConfirmResponse, error) {
	return s.confirm(ctx, userID, idempotencyKey, req)
}

func normalizeImportJobStatus(job importJob) string {
	switch strings.TrimSpace(job.Status) {
	case IMPORT_JOB_RUNNING, IMPORT_JOB_SUCCEEDED, IMPORT_JOB_FAILED:
		return strings.TrimSpace(job.Status)
	}
	if strings.TrimSpace(job.ErrorCode) != "" {
		return IMPORT_JOB_FAILED
	}
	return IMPORT_JOB_SUCCEEDED
}

func importJobStatusResponse(job importJob) ImportJobStatusResponse {
	status := normalizeImportJobStatus(job)
	processedRows := job.ProcessedRows
	if processedRows == 0 && status != IMPORT_JOB_RUNNING {
		processedRows = job.TotalRows
	}
	return ImportJobStatusResponse{
		JobID:         job.IdempotencyKey,
		Status:        status,
		TotalRows:     job.TotalRows,
		ProcessedRows: processedRows,
		SuccessCount:  job.Response.SuccessCount,
		SkipCount:     job.Response.SkipCount,
		FailCount:     job.Response.FailCount,
		Rows:          job.Response.Rows,
		ErrorCode:     strings.TrimSpace(job.ErrorCode),
	}
}
