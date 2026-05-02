package portability

import (
	"context"
	"strings"
	"time"
)

const (
	IMPORT_DUPLICATE_REQUEST = "IMPORT_DUPLICATE_REQUEST"
	IMPORT_PARTIAL_FAILED    = "IMPORT_PARTIAL_FAILED"
)

type ImportRow struct {
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Type        string  `json:"type,omitempty"`
	Category    string  `json:"category,omitempty"`
}

type ImportConfirmRequest struct {
	Rows []ImportRow `json:"rows"`
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

type ImportConfirmRepository interface {
	FindJob(userID string, path string, idempotencyKey string) (importJob, bool)
	SaveJob(job importJob)
	HasTriple(userID string, row storedImportRow) bool
	SaveRow(userID string, row storedImportRow)
	StoredRowCount(userID string) int
	Now() time.Time
}

type ImportConfirmService struct {
	repo ImportConfirmRepository
}

func NewImportConfirmService(repo ImportConfirmRepository) *ImportConfirmService {
	return &ImportConfirmService{repo: repo}
}

func (s *ImportConfirmService) Confirm(userID string, idempotencyKey string, req ImportConfirmRequest) (ImportConfirmResponse, error) {
	userID = strings.TrimSpace(userID)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		return ImportConfirmResponse{}, &contractError{code: IMPORT_DUPLICATE_REQUEST}
	}
	if job, found := s.repo.FindJob(userID, "/api/import/csv/confirm", idempotencyKey); found {
		return job.Response, &contractError{code: IMPORT_DUPLICATE_REQUEST}
	}
	result := ImportConfirmResponse{Rows: make([]ImportConfirmRowResult, 0, len(req.Rows))}
	txnWriter, hasTxnWriter := s.repo.(interface {
		SaveImportedTransaction(userID string, row ImportRow) error
	})
	for idx, row := range req.Rows {
		trimmedDate := strings.TrimSpace(row.Date)
		trimmedDescription := strings.TrimSpace(row.Description)
		if trimmedDate == "" || row.Amount <= 0 || trimmedDescription == "" {
			result.FailCount++
			result.Rows = append(result.Rows, ImportConfirmRowResult{RowIndex: idx, Status: "failed", Reason: "invalid_row"})
			continue
		}
		stored := storedImportRow{Date: trimmedDate, Amount: row.Amount, Description: trimmedDescription}
		if s.repo.HasTriple(userID, stored) {
			result.SkipCount++
			result.Rows = append(result.Rows, ImportConfirmRowResult{RowIndex: idx, Status: "skipped", Reason: "duplicate: amount+date+description match"})
			continue
		}
		if hasTxnWriter {
			if err := txnWriter.SaveImportedTransaction(userID, row); err != nil {
				result.FailCount++
				result.Rows = append(result.Rows, ImportConfirmRowResult{RowIndex: idx, Status: "failed", Reason: "persist_failed"})
				continue
			}
		}
		s.repo.SaveRow(userID, stored)
		result.SuccessCount++
		result.Rows = append(result.Rows, ImportConfirmRowResult{RowIndex: idx, Status: "success"})
	}
	errCode := ""
	if result.FailCount > 0 {
		errCode = IMPORT_PARTIAL_FAILED
	}
	s.repo.SaveJob(importJob{UserID: userID, Path: "/api/import/csv/confirm", IdempotencyKey: idempotencyKey, CreatedAt: s.repo.Now(), Response: result, ErrorCode: errCode})
	if errCode != "" {
		return result, &contractError{code: errCode}
	}
	return result, nil
}

func (s *ImportConfirmService) ConfirmContext(ctx context.Context, userID string, idempotencyKey string, req ImportConfirmRequest) (ImportConfirmResponse, error) {
	_ = ctx
	return s.Confirm(userID, idempotencyKey, req)
}
