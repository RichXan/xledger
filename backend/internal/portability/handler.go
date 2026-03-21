package portability

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	preview *ImportPreviewService
	confirm *ImportConfirmService
}

func NewHandler(preview *ImportPreviewService, confirm *ImportConfirmService) *Handler {
	return &Handler{preview: preview, confirm: confirm}
}

func (h *Handler) ImportPreview(c *gin.Context) {
	if _, ok := userIDFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": IMPORT_INVALID_FILE})
		return
	}
	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": IMPORT_INVALID_FILE})
		return
	}
	defer opened.Close()

	result, err := h.preview.PreviewCSV(opened)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ImportConfirm(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}
	if h.confirm == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "PORTABILITY_INTERNAL"})
		return
	}
	var req ImportConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": IMPORT_DUPLICATE_REQUEST})
		return
	}
	idempotencyKey := c.GetHeader("X-Idempotency-Key")
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": IMPORT_DUPLICATE_REQUEST})
		return
	}
	result, err := h.confirm.ConfirmContext(c.Request.Context(), userID, idempotencyKey, req)
	if err != nil {
		h.writeConfirmError(c, err, result)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch ErrorCode(err) {
	case IMPORT_INVALID_FILE:
		c.JSON(http.StatusBadRequest, gin.H{"error_code": IMPORT_INVALID_FILE})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "PORTABILITY_INTERNAL"})
	}
}

func (h *Handler) writeConfirmError(c *gin.Context, err error, result ImportConfirmResponse) {
	switch ErrorCode(err) {
	case IMPORT_DUPLICATE_REQUEST:
		c.JSON(http.StatusConflict, gin.H{"error_code": IMPORT_DUPLICATE_REQUEST})
	case IMPORT_PARTIAL_FAILED:
		c.JSON(http.StatusOK, gin.H{
			"error_code":    IMPORT_PARTIAL_FAILED,
			"success_count": result.SuccessCount,
			"skip_count":    result.SkipCount,
			"fail_count":    result.FailCount,
			"rows":          result.Rows,
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "PORTABILITY_INTERNAL"})
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
