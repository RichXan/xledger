package portability

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/common/httpx"
)

type Handler struct {
	preview *ImportPreviewService
	confirm *ImportConfirmService
	export  *ExportService
	pat     *PATService
}

func NewHandler(preview *ImportPreviewService, confirm *ImportConfirmService, export *ExportService, pat *PATService) *Handler {
	return &Handler{preview: preview, confirm: confirm, export: export, pat: pat}
}

func (h *Handler) ImportPreview(c *gin.Context) {
	if _, ok := userIDFromContext(c); !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	opened, err := file.Open()
	if err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	defer opened.Close()
	result, err := h.preview.PreviewCSV(opened)
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", result)
}

func (h *Handler) ImportConfirm(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	if h.confirm == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	var req ImportConfirmRequest
	if strings.Contains(strings.ToLower(strings.TrimSpace(c.GetHeader("Content-Type"))), "multipart/form-data") {
		file, err := c.FormFile("file")
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		opened, err := file.Open()
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		defer opened.Close()
		parsed, err := parseImportRowsFromCSV(opened)
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		req = parsed
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
	}
	idempotencyKey := strings.TrimSpace(c.GetHeader("X-Idempotency-Key"))
	if idempotencyKey == "" {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	result, err := h.confirm.ConfirmContext(c.Request.Context(), userID, idempotencyKey, req)
	if err != nil {
		h.writeConfirmError(c, err, result)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", result)
}

func parseImportRowsFromCSV(reader io.Reader) (ImportConfirmRequest, error) {
	records, err := csv.NewReader(reader).ReadAll()
	if err != nil || len(records) < 2 {
		return ImportConfirmRequest{}, err
	}
	headerMap := map[string]int{}
	for idx, name := range records[0] {
		trimmed := strings.TrimSpace(name)
		trimmed = strings.TrimPrefix(trimmed, "\uFEFF")
		headerMap[trimmed] = idx
	}

	getIndex := func(candidates ...string) int {
		for _, candidate := range candidates {
			if idx, ok := headerMap[candidate]; ok {
				return idx
			}
		}
		return -1
	}

	dateIdx := getIndex("时间", "date", "Date", "time", "occurred_at")
	typeIdx := getIndex("类型", "type", "Type")
	purposeIdx := getIndex("用途/来源", "category", "Category")
	amountIdx := getIndex("金额", "amount", "Amount")
	noteIdx := getIndex("备注", "description", "memo", "note")
	signAmountIdx := getIndex("金额正负处理")

	if dateIdx < 0 || amountIdx < 0 {
		return ImportConfirmRequest{}, io.EOF
	}

	rows := make([]ImportRow, 0, len(records)-1)
	for _, raw := range records[1:] {
		date := csvCell(raw, dateIdx)
		amountRaw := csvCell(raw, amountIdx)
		if strings.TrimSpace(date) == "" || strings.TrimSpace(amountRaw) == "" {
			continue
		}

		amount, parseErr := parseImportAmount(amountRaw)
		if parseErr != nil {
			continue
		}

		rowType := normalizeImportType(csvCell(raw, typeIdx))
		signAmount := csvCell(raw, signAmountIdx)
		if rowType == "" && strings.Contains(signAmount, "-") {
			rowType = "expense"
		}
		if rowType == "" && strings.Contains(signAmount, "+") {
			rowType = "income"
		}
		if rowType == "" {
			rowType = "expense"
		}

		category := strings.TrimSpace(csvCell(raw, purposeIdx))
		description := strings.TrimSpace(csvCell(raw, noteIdx))
		if description == "" {
			description = category
		}

		rows = append(rows, ImportRow{
			Date:        strings.TrimSpace(date),
			Amount:      math.Abs(amount),
			Description: description,
			Type:        rowType,
			Category:    category,
		})
	}
	if len(rows) == 0 {
		return ImportConfirmRequest{}, io.EOF
	}
	return ImportConfirmRequest{Rows: rows}, nil
}

func csvCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func parseImportAmount(raw string) (float64, error) {
	replacer := strings.NewReplacer("¥", "", "￥", "", ",", "", " ", "")
	normalized := replacer.Replace(strings.TrimSpace(raw))
	return strconv.ParseFloat(normalized, 64)
}

func normalizeImportType(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(lower, "income"), strings.Contains(lower, "收入"):
		return "income"
	case strings.Contains(lower, "expense"), strings.Contains(lower, "支出"):
		return "expense"
	default:
		return ""
	}
}

func (h *Handler) Export(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	if h.export == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	query := ExportQuery{Format: c.DefaultQuery("format", "csv")}
	if raw := c.Query("from"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		query.From = parsed
	}
	if raw := c.Query("to"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		query.To = parsed
	}
	if raw := c.Query("timeout_ms"); raw != "" {
		parsed, err := time.ParseDuration(raw + "ms")
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		query.Timeout = parsed
	}
	content, err := h.export.Export(c.Request.Context(), userID, query)
	if err != nil {
		h.writeExportError(c, err)
		return
	}
	if query.Format == "json" {
		httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"content": json.RawMessage(content)})
		return
	}
	c.Data(http.StatusOK, "text/csv", []byte(content))
}

func (h *Handler) ListPATs(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	if h.pat == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	items := h.pat.ListPATs(c.Request.Context(), userID)
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"items": items, "pagination": gin.H{"page": 1, "page_size": len(items), "total": len(items), "total_pages": 1}})
}

func (h *Handler) CreatePAT(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	if h.pat == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	plain, record, err := h.pat.CreatePAT(c.Request.Context(), userID, "default", nil)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"token": plain, "id": record.ID, "expires_at": record.ExpiresAt})
}

func (h *Handler) RevokePAT(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	if h.pat == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	if err := h.pat.RevokePAT(c.Request.Context(), userID, c.Param("id")); err != nil {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"revoked": true})
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch ErrorCode(err) {
	case IMPORT_INVALID_FILE:
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
	default:
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
	}
}

func (h *Handler) writeConfirmError(c *gin.Context, err error, result ImportConfirmResponse) {
	switch ErrorCode(err) {
	case IMPORT_DUPLICATE_REQUEST:
		httpx.JSON(c, http.StatusConflict, "BUSINESS_RULE_VIOLATION", "业务规则不满足", nil)
	case IMPORT_PARTIAL_FAILED:
		httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"success_count": result.SuccessCount, "skip_count": result.SkipCount, "fail_count": result.FailCount, "rows": result.Rows})
	default:
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
	}
}

func (h *Handler) writeExportError(c *gin.Context, err error) {
	switch ErrorCode(err) {
	case EXPORT_INVALID_RANGE:
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
	case EXPORT_TIMEOUT:
		httpx.JSON(c, http.StatusGatewayTimeout, "INTERNAL_ERROR", "服务内部错误", nil)
	default:
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
	}
}

func userIDFromContext(c *gin.Context) (string, bool) {
	if value, exists := c.Get("user_id"); exists {
		if userID, ok := value.(string); ok && userID != "" {
			return userID, true
		}
	}
	return "", false
}
