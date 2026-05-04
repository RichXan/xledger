package portability

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/common/httpx"
)

type Handler struct {
	preview  *ImportPreviewService
	confirm  *ImportConfirmService
	export   *ExportService
	pat      *PATService
	shortcut *ShortcutHandler
}

func NewHandler(preview *ImportPreviewService, confirm *ImportConfirmService, export *ExportService, pat *PATService, shortcut ...*ShortcutHandler) *Handler {
	h := &Handler{preview: preview, confirm: confirm, export: export, pat: pat}
	if len(shortcut) > 0 {
		h.shortcut = shortcut[0]
	}
	return h
}

func (h *Handler) SetShortcutHandler(shortcut *ShortcutHandler) {
	h.shortcut = shortcut
}

func (h *Handler) GetPATService() *PATService {
	if h == nil {
		return nil
	}
	return h.pat
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
		parsed, err := ParseImportRowsFromCSV(opened)
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
	query.LedgerID = strings.TrimSpace(c.Query("ledger_id"))
	query.AccountID = strings.TrimSpace(c.Query("account_id"))
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
	if strings.EqualFold(query.Format, "json") {
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
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "生成Token失败", nil)
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
	return httpx.UserIDFromContext(c)
}

func (h *Handler) GenerateShortcut(c *gin.Context) {
	if h.shortcut == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "快捷指令服务未配置", nil)
		return
	}
	h.shortcut.GenerateShortcut(c)
}

func (h *Handler) QuickAdd(c *gin.Context) {
	if h.shortcut == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "快捷指令服务未配置", nil)
		return
	}
	h.shortcut.QuickAdd(c)
}

func (h *Handler) ListCategories(c *gin.Context) {
	if h.shortcut == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "快捷指令服务未配置", nil)
		return
	}
	h.shortcut.ListCategories(c)
}
