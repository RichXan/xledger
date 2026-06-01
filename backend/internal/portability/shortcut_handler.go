package portability

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
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
	mu               sync.Mutex
	configs          map[string]ShortcutConfig
	confirmed        map[string]QuickAddConfirmResponse
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
	AccountID  *string
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
	Name             string  `json:"name"`
	ExpiresIn        *int    `json:"expires_in,omitempty"`
	DefaultLedgerID  string  `json:"default_ledger_id"`
	DefaultAccountID *string `json:"default_account_id,omitempty"`
	Mode             string  `json:"mode"`
}

type GenerateShortcutResponse struct {
	ShortcutID       string  `json:"shortcut_id"`
	PATToken         string  `json:"pat_token"`
	APIEndpoint      string  `json:"api_endpoint"`
	ShortcutURL      string  `json:"shortcut_url"`
	InstallURL       string  `json:"install_url"`
	QRURL            string  `json:"qr_url"`
	DefaultLedgerID  string  `json:"default_ledger_id"`
	DefaultAccountID *string `json:"default_account_id,omitempty"`
	ExpiresAt        *string `json:"expires_at,omitempty"`
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

type ShortcutConfig struct {
	ID               string
	UserID           string
	Name             string
	DefaultLedgerID  string
	DefaultAccountID *string
	Mode             string
	PATID            string
	CreatedAt        time.Time
	ExpiresAt        *time.Time
}

type QuickAddPreviewRequest struct {
	ShortcutID       string  `json:"shortcut_id"`
	RawText          string  `json:"raw_text" binding:"required"`
	Source           string  `json:"source"`
	IdempotencyKey   string  `json:"idempotency_key"`
	DefaultLedgerID  string  `json:"default_ledger_id"`
	DefaultAccountID *string `json:"default_account_id,omitempty"`
}

type QuickAddSuggestion struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence,omitempty"`
}

type QuickAddPreviewResponse struct {
	ShortcutID         string               `json:"shortcut_id"`
	Amount             float64              `json:"amount"`
	Type               string               `json:"type"`
	OccurredAt         string               `json:"occurred_at"`
	Memo               string               `json:"memo"`
	CategorySuggestion *QuickAddSuggestion  `json:"category_suggestion,omitempty"`
	LedgerSuggestions  []QuickAddSuggestion `json:"ledger_suggestions"`
	AccountSuggestions []QuickAddSuggestion `json:"account_suggestions"`
	NeedsReview        bool                 `json:"needs_review"`
	RawText            string               `json:"raw_text,omitempty"`
}

type QuickAddConfirmRequest struct {
	ShortcutID     string     `json:"shortcut_id"`
	IdempotencyKey string     `json:"idempotency_key"`
	LedgerID       string     `json:"ledger_id" binding:"required"`
	AccountID      *string    `json:"account_id"`
	CategoryID     *string    `json:"category_id"`
	TagIDs         []string   `json:"tag_ids"`
	Type           string     `json:"type" binding:"required,oneof=expense income"`
	Amount         float64    `json:"amount" binding:"required"`
	Memo           string     `json:"memo"`
	OccurredAt     *time.Time `json:"occurred_at"`
}

type QuickAddConfirmResponse struct {
	ID         string  `json:"id"`
	Amount     float64 `json:"amount"`
	Type       string  `json:"type"`
	Memo       string  `json:"memo"`
	Success    bool    `json:"success"`
	OccurredAt string  `json:"occurred_at"`
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
		configs:          make(map[string]ShortcutConfig),
		confirmed:        make(map[string]QuickAddConfirmResponse),
	}
}

func (h *ShortcutHandler) GenerateShortcut(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "鏈璇佹垨鍑瘉鏃犳晥", nil)
		return
	}

	var req GenerateShortcutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Name = "蹇嵎璁拌处"
	}

	if req.Name == "" {
		req.Name = "蹇嵎璁拌处"
	}

	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(*req.ExpiresIn) * 24 * time.Hour)
		expiresAt = &t
	}

	token, record, err := h.patService.CreatePAT(c.Request.Context(), userID, req.Name, expiresAt)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "鐢熸垚Token澶辫触", nil)
		return
	}

	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	apiEndpoint := scheme + "://" + c.Request.Host
	shortcutID := "sc_" + record.ID
	defaultLedgerID := strings.TrimSpace(req.DefaultLedgerID)
	if defaultLedgerID == "" && h.ledgerSvc != nil {
		if id, err := h.ledgerSvc.GetDefaultLedgerID(c.Request.Context(), userID); err == nil {
			defaultLedgerID = id
		}
	}
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = "manual"
	}
	config := ShortcutConfig{
		ID:               shortcutID,
		UserID:           userID,
		Name:             req.Name,
		DefaultLedgerID:  defaultLedgerID,
		DefaultAccountID: cleanOptionalString(req.DefaultAccountID),
		Mode:             mode,
		PATID:            record.ID,
		CreatedAt:        time.Now().UTC(),
		ExpiresAt:        record.ExpiresAt,
	}
	h.mu.Lock()
	h.configs[shortcutConfigKey(userID, shortcutID)] = config
	h.mu.Unlock()

	var expiresAtStr *string
	if record.ExpiresAt != nil {
		s := record.ExpiresAt.Format(time.RFC3339)
		expiresAtStr = &s
	}

	httpx.JSON(c, http.StatusOK, "OK", "鎴愬姛", GenerateShortcutResponse{
		ShortcutID:       shortcutID,
		PATToken:         token,
		APIEndpoint:      apiEndpoint,
		ShortcutURL:      "shortcuts://run-shortcut?name=" + req.Name,
		InstallURL:       "shortcuts://run-shortcut?name=" + req.Name,
		QRURL:            apiEndpoint + "/api/shortcuts/" + shortcutID + "/qr",
		DefaultLedgerID:  defaultLedgerID,
		DefaultAccountID: config.DefaultAccountID,
		ExpiresAt:        expiresAtStr,
	})
}

func (h *ShortcutHandler) PreviewQuickAdd(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "鏈璇佹垨鍑瘉鏃犳晥", nil)
		return
	}

	var req QuickAddPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "璇锋眰鍙傛暟涓嶅悎娉? "+err.Error(), nil)
		return
	}

	config := h.findShortcutConfig(userID, req.ShortcutID)
	ledgerID := firstNonEmpty(req.DefaultLedgerID, config.DefaultLedgerID)
	if ledgerID == "" && h.ledgerSvc != nil {
		if id, err := h.ledgerSvc.GetDefaultLedgerID(c.Request.Context(), userID); err == nil {
			ledgerID = id
		}
	}
	accountID := cleanOptionalString(req.DefaultAccountID)
	if accountID == nil {
		accountID = config.DefaultAccountID
	}

	parsed := parseOCRQuickAdd(req.RawText)
	categorySuggestion := h.suggestCategory(c.Request.Context(), userID, parsed.Category)
	httpx.JSON(c, http.StatusOK, "OK", "鎴愬姛", QuickAddPreviewResponse{
		ShortcutID:         firstNonEmpty(req.ShortcutID, config.ID),
		Amount:             parsed.Amount,
		Type:               parsed.Type,
		OccurredAt:         parsed.OccurredAt.Format(time.RFC3339),
		Memo:               parsed.Memo,
		CategorySuggestion: categorySuggestion,
		LedgerSuggestions:  ledgerSuggestions(ledgerID),
		AccountSuggestions: accountSuggestions(accountID),
		NeedsReview:        parsed.Amount <= 0 || categorySuggestion == nil,
		RawText:            req.RawText,
	})
}

func (h *ShortcutHandler) ConfirmQuickAdd(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "鏈璇佹垨鍑瘉鏃犳晥", nil)
		return
	}

	var req QuickAddConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "璇锋眰鍙傛暟涓嶅悎娉? "+err.Error(), nil)
		return
	}
	if req.Amount <= 0 {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "閲戦蹇呴』澶т簬0", nil)
		return
	}

	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey != "" {
		if cached, found := h.findConfirmed(userID, idempotencyKey); found {
			httpx.JSON(c, http.StatusCreated, "OK", "鎴愬姛", cached)
			return
		}
	}

	occurredAt := time.Now().UTC()
	if req.OccurredAt != nil {
		occurredAt = req.OccurredAt.UTC()
	}
	result, err := h.transactionSvc.CreateTransaction(c.Request.Context(), userID, ShortcutTransactionInput{
		LedgerID:   req.LedgerID,
		AccountID:  cleanOptionalString(req.AccountID),
		CategoryID: cleanOptionalString(req.CategoryID),
		Type:       req.Type,
		Amount:     req.Amount,
		Memo:       strings.TrimSpace(req.Memo),
		OccurredAt: occurredAt,
	})
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "鍒涘缓浜ゆ槗澶辫触", nil)
		return
	}
	response := QuickAddConfirmResponse{
		ID:         result.ID,
		Amount:     result.Amount,
		Type:       result.Type,
		Memo:       result.Memo,
		Success:    true,
		OccurredAt: occurredAt.Format(time.RFC3339),
	}
	if idempotencyKey != "" {
		h.rememberConfirmed(userID, idempotencyKey, response)
	}
	if h.callbackRecorder != nil {
		_ = h.callbackRecorder.Record(c.Request.Context(), userID, ShortcutCallback{
			ShortcutID: req.ShortcutID,
			ActionType: "quick_add_confirm",
			Amount:     req.Amount,
			Type:       req.Type,
			Memo:       req.Memo,
			Success:    true,
		})
	}
	httpx.JSON(c, http.StatusCreated, "OK", "鎴愬姛", response)
}

func (h *ShortcutHandler) QuickAdd(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "鏈璇佹垨鍑瘉鏃犳晥", nil)
		return
	}

	var req QuickAddV2Request
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "璇锋眰鍙傛暟涓嶅悎娉? "+err.Error(), nil)
		return
	}

	if req.Amount <= 0 {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "閲戦蹇呴』澶т簬0", nil)
		return
	}

	ledgerID, err := h.ledgerSvc.GetDefaultLedgerID(c.Request.Context(), userID)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "鏃犳硶鑾峰彇榛樿璐︽湰", nil)
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
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "鍒涘缓浜ゆ槗澶辫触", nil)
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

	httpx.JSON(c, http.StatusCreated, "OK", "鎴愬姛", QuickAddV2Response{
		ID:      result.ID,
		Amount:  result.Amount,
		Type:    result.Type,
		Memo:    result.Memo,
		Success: true,
	})
}

func (h *ShortcutHandler) findShortcutConfig(userID string, shortcutID string) ShortcutConfig {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.configs[shortcutConfigKey(userID, shortcutID)]
}

func (h *ShortcutHandler) findConfirmed(userID string, idempotencyKey string) (QuickAddConfirmResponse, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	result, found := h.confirmed[shortcutConfigKey(userID, idempotencyKey)]
	return result, found
}

func (h *ShortcutHandler) rememberConfirmed(userID string, idempotencyKey string, response QuickAddConfirmResponse) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.confirmed[shortcutConfigKey(userID, idempotencyKey)] = response
}

func shortcutConfigKey(userID string, id string) string {
	return strings.TrimSpace(userID) + ":" + strings.TrimSpace(id)
}

func cleanOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

type parsedQuickAddOCR struct {
	Amount     float64
	Type       string
	OccurredAt time.Time
	Memo       string
	Category   string
}

func parseOCRQuickAdd(raw string) parsedQuickAddOCR {
	normalized := strings.TrimSpace(raw)
	amount := parseFirstAmount(normalized)
	when := parseFirstTime(normalized)
	payChannel := "OCR"
	lower := strings.ToLower(normalized)
	if strings.Contains(normalized, "微信") || strings.Contains(lower, "wechat") {
		payChannel = "WeChat Pay"
	} else if strings.Contains(normalized, "支付宝") || strings.Contains(lower, "alipay") {
		payChannel = "Alipay"
	}
	merchant := parseMerchant(normalized)
	memo := strings.TrimSpace(strings.Join(nonEmptyStrings(merchant, payChannel), " - "))
	if memo == "" {
		memo = "OCR quick add"
	}
	category := ""
	if strings.Contains(normalized, "咖啡") || strings.Contains(normalized, "餐") || strings.Contains(normalized, "饭") ||
		strings.Contains(lower, "coffee") || strings.Contains(lower, "restaurant") || strings.Contains(lower, "meal") {
		category = "Food"
	}
	return parsedQuickAddOCR{
		Amount:     amount,
		Type:       "expense",
		OccurredAt: when,
		Memo:       memo,
		Category:   category,
	}
}

func parseFirstAmount(raw string) float64 {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:amount|total|paid|pay|price|支付金额|实付|金额)\s*[:：]?\s*[￥¥$]?\s*([0-9]+(?:\.[0-9]+)?)`),
		regexp.MustCompile(`[￥¥$]\s*([0-9]+(?:\.[0-9]+)?)`),
		regexp.MustCompile(`(?m)(?:^|\s)([0-9]+(?:\.[0-9]+)?)(?:\s|$)`),
	}
	for _, pattern := range patterns {
		if match := pattern.FindStringSubmatch(raw); len(match) == 2 {
			amount, _ := strconv.ParseFloat(match[1], 64)
			if amount > 0 {
				return amount
			}
		}
	}
	return 0
}

func parseFirstTime(raw string) time.Time {
	timePattern := regexp.MustCompile(`([0-9]{4}-[0-9]{2}-[0-9]{2})\s+([0-9]{2}:[0-9]{2}(?::[0-9]{2})?)`)
	if match := timePattern.FindStringSubmatch(raw); len(match) == 3 {
		layout := "2006-01-02 15:04"
		value := match[1] + " " + match[2]
		if strings.Count(match[2], ":") == 2 {
			layout = "2006-01-02 15:04:05"
		}
		if parsed, err := time.ParseInLocation(layout, value, time.FixedZone("CST", 8*60*60)); err == nil {
			return parsed
		}
	}
	return time.Now().UTC()
}

func parseMerchant(raw string) string {
	labels := []string{"merchant", "store", "seller", "payee", "收款方", "商户", "商家", "对方"}
	for _, label := range labels {
		pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(label) + `\s*[:：]\s*([^\n\r]+)`)
		if match := pattern.FindStringSubmatch(raw); len(match) == 2 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return result
}

func (h *ShortcutHandler) suggestCategory(ctx context.Context, userID string, category string) *QuickAddSuggestion {
	category = strings.TrimSpace(category)
	if category == "" || h.categorySvc == nil {
		return nil
	}
	id, err := h.categorySvc.FindByName(ctx, userID, category)
	if err != nil || id == "" {
		return nil
	}
	return &QuickAddSuggestion{ID: id, Name: category, Reason: "OCR keyword match", Confidence: 0.82}
}

func ledgerSuggestions(id string) []QuickAddSuggestion {
	if strings.TrimSpace(id) == "" {
		return []QuickAddSuggestion{}
	}
	return []QuickAddSuggestion{{ID: strings.TrimSpace(id), Name: "Default ledger", Reason: "Shortcut setup", Confidence: 1}}
}

func accountSuggestions(id *string) []QuickAddSuggestion {
	if id == nil || strings.TrimSpace(*id) == "" {
		return []QuickAddSuggestion{{ID: "", Name: "No account", Reason: "Account is optional", Confidence: 1}}
	}
	name := "Default account"
	if strings.Contains(strings.ToLower(*id), "wechat") || strings.Contains(*id, "微信") {
		name = "WeChat Wallet"
	}
	return []QuickAddSuggestion{{ID: strings.TrimSpace(*id), Name: name, Reason: "Shortcut setup", Confidence: 1}}
}

func (h *ShortcutHandler) ListCategories(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "鏈璇佹垨鍑瘉鏃犳晥", nil)
		return
	}

	if h.categorySvc == nil {
		httpx.JSON(c, http.StatusOK, "OK", "鎴愬姛", []string{})
		return
	}

	names, err := h.categorySvc.ListNames(c.Request.Context(), userID)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "鑾峰彇鍒嗙被鍒楄〃澶辫触", nil)
		return
	}

	httpx.JSON(c, http.StatusOK, "OK", "鎴愬姛", names)
}
