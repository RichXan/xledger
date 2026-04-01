package portability

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/common/httpx"
)

type ShortcutHandler struct {
	patService       *PATService
	transactionSvc   ShortcutTransactionCreator
	ledgerSvc        ShortcutLedgerLookup
	categorySvc      ShortcutCategoryLookup
	callbackRecorder ShortcutCallbackRecorder
}

type ShortcutTransactionCreator interface {
	CreateTransaction(ctx context.Context, userID string, input ShortcutTransactionInput) (ShortcutTransactionResult, error)
}

type ShortcutLedgerLookup interface {
	GetDefaultLedgerID(ctx context.Context, userID string) (string, error)
}

type ShortcutCategoryLookup interface {
	FindByName(ctx context.Context, userID string, name string) (string, error)
	ListNames(ctx context.Context, userID string) ([]string, error)
}

type ShortcutCallbackRecorder interface {
	Record(ctx context.Context, userID string, callback ShortcutCallback) error
}

type ShortcutTransactionInput struct {
	LedgerID   string
	CategoryID *string
	Type       string
	Amount     float64
	Memo       string
	OccurredAt time.Time
}

type ShortcutTransactionResult struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
	Memo   string  `json:"memo"`
}

type ShortcutCallback struct {
	ShortcutID   string  `json:"shortcut_id"`
	ActionType   string  `json:"action_type"`
	Amount       float64 `json:"amount"`
	Type         string  `json:"type"`
	Category     string  `json:"category"`
	Memo         string  `json:"memo"`
	SourceApp    string  `json:"source_app"`
	OCRText      string  `json:"ocr_text"`
	Success      bool    `json:"success"`
	ErrorMessage string  `json:"error_message"`
}

type GenerateShortcutRequest struct {
	Name      string `json:"name"`
	ExpiresIn *int   `json:"expires_in,omitempty"`
}

type GenerateShortcutResponse struct {
	PATToken    string  `json:"pat_token"`
	APIEndpoint string  `json:"api_endpoint"`
	ShortcutURL string  `json:"shortcut_url"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type QuickAddV2Request struct {
	Amount   float64 `json:"amount" binding:"required"`
	Type     string  `json:"type" binding:"required,oneof=expense income"`
	Category string  `json:"category"`
	Memo     string  `json:"memo"`
}

type QuickAddV2Response struct {
	ID      string  `json:"id"`
	Amount  float64 `json:"amount"`
	Type    string  `json:"type"`
	Memo    string  `json:"memo"`
	Success bool    `json:"success"`
}

func NewShortcutHandler(
	patService *PATService,
	txnSvc ShortcutTransactionCreator,
	ledgerSvc ShortcutLedgerLookup,
	catSvc ShortcutCategoryLookup,
	callbackSvc ShortcutCallbackRecorder,
) *ShortcutHandler {
	return &ShortcutHandler{
		patService:       patService,
		transactionSvc:   txnSvc,
		ledgerSvc:        ledgerSvc,
		categorySvc:      catSvc,
		callbackRecorder: callbackSvc,
	}
}

func (h *ShortcutHandler) GenerateShortcut(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}

	var req GenerateShortcutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Name = "快捷记账"
	}

	if req.Name == "" {
		req.Name = "快捷记账"
	}

	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(*req.ExpiresIn) * 24 * time.Hour)
		expiresAt = &t
	}

	token, record, err := h.patService.CreatePAT(c.Request.Context(), userID, req.Name, expiresAt)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "生成Token失败", nil)
		return
	}

	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	apiEndpoint := scheme + "://" + c.Request.Host

	var expiresAtStr *string
	if record.ExpiresAt != nil {
		s := record.ExpiresAt.Format(time.RFC3339)
		expiresAtStr = &s
	}

	httpx.JSON(c, http.StatusOK, "OK", "成功", GenerateShortcutResponse{
		PATToken:    token,
		APIEndpoint: apiEndpoint,
		ExpiresAt:   expiresAtStr,
	})
}

func (h *ShortcutHandler) QuickAdd(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}

	var req QuickAddV2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法: "+err.Error(), nil)
		return
	}

	if req.Amount <= 0 {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "金额必须大于0", nil)
		return
	}

	ledgerID, err := h.ledgerSvc.GetDefaultLedgerID(c.Request.Context(), userID)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "无法获取默认账本", nil)
		return
	}

	var categoryID *string
	if req.Category != "" && h.categorySvc != nil {
		id, err := h.categorySvc.FindByName(c.Request.Context(), userID, req.Category)
		if err == nil && id != "" {
			categoryID = &id
		}
	}

	input := ShortcutTransactionInput{
		LedgerID:   ledgerID,
		CategoryID: categoryID,
		Type:       req.Type,
		Amount:     req.Amount,
		Memo:       req.Memo,
		OccurredAt: time.Now().UTC(),
	}

	result, err := h.transactionSvc.CreateTransaction(c.Request.Context(), userID, input)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "创建交易失败", nil)
		return
	}

	if h.callbackRecorder != nil {
		_ = h.callbackRecorder.Record(c.Request.Context(), userID, ShortcutCallback{
			ActionType: "quick_add",
			Amount:     req.Amount,
			Type:       req.Type,
			Category:   req.Category,
			Memo:       req.Memo,
			Success:    true,
		})
	}

	httpx.JSON(c, http.StatusCreated, "OK", "成功", QuickAddV2Response{
		ID:      result.ID,
		Amount:  result.Amount,
		Type:    result.Type,
		Memo:    result.Memo,
		Success: true,
	})
}

func (h *ShortcutHandler) ListCategories(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}

	if h.categorySvc == nil {
		httpx.JSON(c, http.StatusOK, "OK", "成功", []string{})
		return
	}

	names, err := h.categorySvc.ListNames(c.Request.Context(), userID)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "获取分类列表失败", nil)
		return
	}

	httpx.JSON(c, http.StatusOK, "OK", "成功", names)
}
