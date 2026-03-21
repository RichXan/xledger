package portability

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	preview *ImportPreviewService
}

func NewHandler(preview *ImportPreviewService) *Handler {
	return &Handler{preview: preview}
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

func (h *Handler) writeError(c *gin.Context, err error) {
	switch ErrorCode(err) {
	case IMPORT_INVALID_FILE:
		c.JSON(http.StatusBadRequest, gin.H{"error_code": IMPORT_INVALID_FILE})
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
