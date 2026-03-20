package accounting

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	ledgerService  *LedgerService
	accountService *AccountService
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

func NewHandler(ledgerService *LedgerService, accountService *AccountService) *Handler {
	return &Handler{ledgerService: ledgerService, accountService: accountService}
}

func (h *Handler) CreateAccount(c *gin.Context) {
	if h.accountService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": ACCOUNT_INVALID})
		return
	}

	account, err := h.accountService.CreateAccount(c.Request.Context(), userID, AccountCreateInput{
		Name:           req.Name,
		Type:           req.Type,
		InitialBalance: req.InitialBalance,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, account)
}

func (h *Handler) ListAccounts(c *gin.Context) {
	if h.accountService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	accounts, err := h.accountService.ListAccounts(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": accounts})
}

func (h *Handler) GetAccount(c *gin.Context) {
	if h.accountService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	account, err := h.accountService.GetAccount(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, account)
}

func (h *Handler) UpdateAccount(c *gin.Context) {
	if h.accountService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	var req updateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": ACCOUNT_INVALID})
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
	c.JSON(http.StatusOK, account)
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	if h.accountService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	err := h.accountService.DeleteAccount(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) ListLedgers(c *gin.Context) {
	if h.ledgerService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	ledgers, err := h.ledgerService.ListLedgers(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": ledgers})
}

func (h *Handler) CreateLedger(c *gin.Context) {
	if h.ledgerService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	var req createLedgerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": LEDGER_INVALID})
		return
	}

	ledger, err := h.ledgerService.CreateLedger(c.Request.Context(), userID, LedgerCreateInput{Name: req.Name, IsDefault: req.IsDefault})
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ledger)
}

func (h *Handler) UpdateLedger(c *gin.Context) {
	if h.ledgerService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	var req updateLedgerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": LEDGER_INVALID})
		return
	}

	ledger, err := h.ledgerService.UpdateLedger(c.Request.Context(), userID, c.Param("id"), LedgerCreateInput{Name: req.Name})
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, ledger)
}

func (h *Handler) DeleteLedger(c *gin.Context) {
	if h.ledgerService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
		return
	}

	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}

	err := h.ledgerService.DeleteLedger(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch ErrorCode(err) {
	case ACCOUNT_INVALID:
		c.JSON(http.StatusBadRequest, gin.H{"error_code": ACCOUNT_INVALID})
	case ACCOUNT_NOT_FOUND:
		c.JSON(http.StatusNotFound, gin.H{"error_code": ACCOUNT_NOT_FOUND})
	case LEDGER_DEFAULT_IMMUTABLE:
		c.JSON(http.StatusConflict, gin.H{"error_code": LEDGER_DEFAULT_IMMUTABLE})
	case LEDGER_INVALID:
		c.JSON(http.StatusBadRequest, gin.H{"error_code": LEDGER_INVALID})
	case LEDGER_NOT_FOUND:
		c.JSON(http.StatusNotFound, gin.H{"error_code": LEDGER_NOT_FOUND})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
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
