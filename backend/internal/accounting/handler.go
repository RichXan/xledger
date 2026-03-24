package accounting

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/classification"
	"xledger/backend/internal/common/httpx"
)

type Handler struct {
	ledgerService      *LedgerService
	accountService     *AccountService
	transactionService *TransactionService
}

type createAccountRequest struct {
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	InitialBalance float64 `json:"initial_balance"`
}

type updateAccountRequest struct {
	Name       *string `json:"name"`
	Type       *string `json:"type"`
	ArchivedAt *bool   `json:"archived"`
}

type createLedgerRequest struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

type updateLedgerRequest struct {
	Name string `json:"name"`
}

type createTransactionRequest struct {
	LedgerID      string     `json:"ledger_id"`
	FromLedgerID  *string    `json:"from_ledger_id"`
	ToLedgerID    *string    `json:"to_ledger_id"`
	AccountID     *string    `json:"account_id"`
	CategoryID    *string    `json:"category_id"`
	TagIDs        []string   `json:"tag_ids"`
	FromAccountID *string    `json:"from_account_id"`
	ToAccountID   *string    `json:"to_account_id"`
	Type          string     `json:"type"`
	Amount        float64    `json:"amount"`
	OccurredAt    *time.Time `json:"occurred_at"`
}

type updateTransactionRequest struct {
	Amount     float64  `json:"amount"`
	Version    *int     `json:"version"`
	CategoryID *string  `json:"category_id"`
	TagIDs     []string `json:"tag_ids"`
}

func NewHandler(ledgerService *LedgerService, accountService *AccountService, transactionService ...*TransactionService) *Handler {
	handler := &Handler{ledgerService: ledgerService, accountService: accountService}
	if len(transactionService) > 0 {
		handler.transactionService = transactionService[0]
	}
	return handler
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	if h.transactionService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var req createTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	occurredAt := time.Time{}
	if req.OccurredAt != nil {
		occurredAt = req.OccurredAt.UTC()
	}
	if req.Type == TransactionTypeTransfer {
		txn, err := h.transactionService.CreateTransfer(c.Request.Context(), userID, TransactionTransferInput{LedgerID: req.LedgerID, FromLedgerID: req.FromLedgerID, ToLedgerID: req.ToLedgerID, FromAccountID: req.FromAccountID, ToAccountID: req.ToAccountID, Amount: req.Amount, OccurredAt: occurredAt})
		if err != nil {
			h.writeError(c, err)
			return
		}
		httpx.JSON(c, http.StatusCreated, "OK", "成功", txn)
		return
	}
	txn, err := h.transactionService.CreateTransaction(c.Request.Context(), userID, TransactionCreateInput{LedgerID: req.LedgerID, AccountID: req.AccountID, CategoryID: req.CategoryID, TagIDs: req.TagIDs, Type: req.Type, Amount: req.Amount, OccurredAt: occurredAt})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, "OK", "成功", txn)
}

func (h *Handler) UpdateTransaction(c *gin.Context) {
	if h.transactionService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var req updateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	txn, err := h.transactionService.EditTransaction(c.Request.Context(), userID, c.Param("id"), TransactionEditInput{Amount: req.Amount, Version: req.Version, HasCategory: req.CategoryID != nil, CategoryID: req.CategoryID, HasTagIDs: req.TagIDs != nil, TagIDs: req.TagIDs})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", txn)
}

func (h *Handler) DeleteTransaction(c *gin.Context) {
	if h.transactionService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var version *int
	if rawVersion := c.Query("version"); rawVersion != "" {
		parsed, parseErr := strconv.Atoi(rawVersion)
		if parseErr != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		version = &parsed
	}
	if err := h.transactionService.DeleteTransaction(c.Request.Context(), userID, c.Param("id"), version); err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"deleted": true})
}

func (h *Handler) ListTransactions(c *gin.Context) {
	if h.transactionService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	query := TransactionQuery{LedgerID: c.Query("ledger_id"), AccountID: c.Query("account_id"), CategoryID: c.Query("category_id"), TagID: c.Query("tag_id")}
	if rawFrom := c.Query("date_from"); rawFrom != "" {
		parsed, err := time.Parse(time.RFC3339, rawFrom)
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		query.OccurredFrom = parsed.UTC()
	}
	if rawTo := c.Query("date_to"); rawTo != "" {
		parsed, err := time.Parse(time.RFC3339, rawTo)
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		query.OccurredTo = parsed.UTC()
	}
	if rawPage := c.Query("page"); rawPage != "" {
		parsed, err := strconv.Atoi(rawPage)
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		query.Page = parsed
	}
	if rawPageSize := c.Query("page_size"); rawPageSize != "" {
		parsed, err := strconv.Atoi(rawPageSize)
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		query.PageSize = parsed
	}
	items, total, err := h.transactionService.ListTransactionsWithTotal(c.Request.Context(), userID, query)
	if err != nil {
		h.writeError(c, err)
		return
	}
	page := query.Page
	pageSize := query.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"items": items, "pagination": gin.H{"page": page, "page_size": pageSize, "total": total, "total_pages": totalPages}})
}

func (h *Handler) CreateAccount(c *gin.Context) {
	if h.accountService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	account, err := h.accountService.CreateAccount(c.Request.Context(), userID, AccountCreateInput{Name: req.Name, Type: req.Type, InitialBalance: req.InitialBalance})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, "OK", "成功", account)
}

func (h *Handler) ListAccounts(c *gin.Context) {
	if h.accountService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	accounts, err := h.accountService.ListAccounts(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	page, pageSize := 1, len(accounts)
	if rawPage := c.Query("page"); rawPage != "" {
		if parsed, err := strconv.Atoi(rawPage); err == nil {
			page = parsed
		}
	}
	if rawPageSize := c.Query("page_size"); rawPageSize != "" {
		if parsed, err := strconv.Atoi(rawPageSize); err == nil {
			pageSize = parsed
		}
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"items": accounts, "pagination": gin.H{"page": page, "page_size": pageSize, "total": len(accounts), "total_pages": 1}})
}

func (h *Handler) GetAccount(c *gin.Context) {
	if h.accountService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	account, err := h.accountService.GetAccount(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", account)
}

func (h *Handler) UpdateAccount(c *gin.Context) {
	if h.accountService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var req updateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	input := AccountUpdateInput{}
	if req.Name != nil {
		input.HasName = true
		input.Name = *req.Name
	}
	if req.Type != nil {
		input.HasType = true
		input.Type = *req.Type
	}
	if req.ArchivedAt != nil {
		input.HasArchive = true
		input.Archive = *req.ArchivedAt
	}
	account, err := h.accountService.UpdateAccount(c.Request.Context(), userID, c.Param("id"), input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", account)
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	if h.accountService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	if err := h.accountService.DeleteAccount(c.Request.Context(), userID, c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"deleted": true})
}

func (h *Handler) ListLedgers(c *gin.Context) {
	if h.ledgerService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	ledgers, err := h.ledgerService.ListLedgers(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"items": ledgers, "pagination": gin.H{"page": 1, "page_size": len(ledgers), "total": len(ledgers), "total_pages": 1}})
}

func (h *Handler) CreateLedger(c *gin.Context) {
	if h.ledgerService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var req createLedgerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	ledger, err := h.ledgerService.CreateLedger(c.Request.Context(), userID, LedgerCreateInput{Name: req.Name, IsDefault: req.IsDefault})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusCreated, "OK", "成功", ledger)
}

func (h *Handler) UpdateLedger(c *gin.Context) {
	if h.ledgerService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	var req updateLedgerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}
	ledger, err := h.ledgerService.UpdateLedger(c.Request.Context(), userID, c.Param("id"), LedgerCreateInput{Name: req.Name})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", ledger)
}

func (h *Handler) DeleteLedger(c *gin.Context) {
	if h.ledgerService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	if err := h.ledgerService.DeleteLedger(c.Request.Context(), userID, c.Param("id")); err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"deleted": true})
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch ErrorCode(err) {
	case ACCOUNT_INVALID, LEDGER_INVALID, TXN_VALIDATION_FAILED:
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
	case ACCOUNT_NOT_FOUND, LEDGER_NOT_FOUND, TXN_NOT_FOUND:
		httpx.JSON(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "资源不存在", nil)
	case LEDGER_DEFAULT_IMMUTABLE, TXN_CONFLICT:
		httpx.JSON(c, http.StatusConflict, "BUSINESS_RULE_VIOLATION", "业务规则不满足", nil)
	default:
		switch classification.ErrorCode(err) {
		case classification.CAT_INVALID, classification.CAT_ARCHIVED, classification.TAG_INVALID:
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		case classification.CAT_NOT_FOUND, classification.TAG_NOT_FOUND:
			httpx.JSON(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "资源不存在", nil)
		default:
			httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		}
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
