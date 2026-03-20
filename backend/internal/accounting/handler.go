package accounting

import (
	"net/http"
	"strings"

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
	case LEDGER_NOT_FOUND:
		c.JSON(http.StatusNotFound, gin.H{"error_code": LEDGER_NOT_FOUND})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "ACCOUNTING_INTERNAL"})
	}
}

func userIDFromContext(c *gin.Context) (string, bool) {
	if value, exists := c.Get("user_id"); exists {
		if userID, ok := value.(string); ok && strings.TrimSpace(userID) != "" {
			return userID, true
		}
	}

	headerUserID := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if headerUserID != "" {
		return headerUserID, true
	}
	return "", false
}
