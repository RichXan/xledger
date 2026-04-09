package accounting

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/classification"
	"xledger/backend/internal/common/httpx"
	"xledger/backend/internal/common/text"
)

type QuickAddHandler struct {
	transactionService *TransactionService
	ledgerService      *LedgerService
	categoryService    *classification.CategoryService
}

type QuickAddRequest struct {
	Amount    float64 `json:"amount" binding:"required"`
	Type      string  `json:"type" binding:"required,oneof=expense income"`
	Category  string  `json:"category"`
	Memo      string  `json:"memo"`
	Timestamp string  `json:"timestamp"`
}

type QuickAddResponse struct {
	ID           string  `json:"id"`
	Amount       float64 `json:"amount"`
	Type         string  `json:"type"`
	CategoryName string  `json:"category_name,omitempty"`
	Memo         string  `json:"memo,omitempty"`
	OccurredAt   string  `json:"occurred_at"`
}

type QuickAddCategoriesResponse struct {
	Categories []string `json:"categories"`
}

func NewQuickAddHandler(txnSvc *TransactionService, ledgerSvc *LedgerService, catSvc *classification.CategoryService) *QuickAddHandler {
	return &QuickAddHandler{
		transactionService: txnSvc,
		ledgerService:      ledgerSvc,
		categoryService:    catSvc,
	}
}

func (h *QuickAddHandler) QuickAdd(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}

	var req QuickAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法: "+err.Error(), nil)
		return
	}

	if req.Amount <= 0 {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "金额必须大于0", nil)
		return
	}

	ledgers, err := h.ledgerService.ListLedgers(c.Request.Context(), userID)
	if err != nil || len(ledgers) == 0 {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "无法获取默认账本", nil)
		return
	}
	ledgerID := ledgers[0].ID
	for _, l := range ledgers {
		if l.IsDefault {
			ledgerID = l.ID
			break
		}
	}

	var categoryID *string
	var categoryName string
	if req.Category != "" && h.categoryService != nil {
		id, err := h.categoryService.FindByName(c.Request.Context(), userID, req.Category)
		if err == nil && id != "" {
			categoryID = &id
			categoryName = req.Category
		}
	}

	occurredAt := time.Now().UTC()
	if req.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			occurredAt = parsed
		}
	}

	input := TransactionCreateInput{
		LedgerID:   ledgerID,
		CategoryID: categoryID,
		Type:       req.Type,
		Amount:     req.Amount,
		Memo:       req.Memo,
		OccurredAt: occurredAt,
	}

	result, err := h.transactionService.CreateTransaction(c.Request.Context(), userID, input)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "创建交易失败", nil)
		return
	}

	httpx.JSON(c, http.StatusCreated, "OK", "成功", QuickAddResponse{
		ID:           result.ID,
		Amount:       result.Amount,
		Type:         result.Type,
		CategoryName: categoryName,
		Memo:         result.Memo,
		OccurredAt:   occurredAt.Format(time.RFC3339),
	})
}

func (h *QuickAddHandler) ListCategories(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}

	if h.categoryService == nil {
		httpx.JSON(c, http.StatusOK, "OK", "成功", QuickAddCategoriesResponse{Categories: []string{}})
		return
	}

	categories, err := h.categoryService.ListCategories(c.Request.Context(), userID)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "获取分类列表失败", nil)
		return
	}

	names := make([]string, 0, len(categories))
	for _, cat := range categories {
		names = append(names, text.StripEmojiPrefix(cat.Name))
	}

	httpx.JSON(c, http.StatusOK, "OK", "成功", QuickAddCategoriesResponse{
		Categories: names,
	})
}
