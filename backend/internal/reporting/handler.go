package reporting

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/common/httpx"
)

type Handler struct {
	overview *OverviewService
	trend    *TrendService
	category *CategoryService
}

func NewHandler(overview *OverviewService, trend *TrendService, category *CategoryService) *Handler {
	return &Handler{overview: overview, trend: trend, category: category}
}

func (h *Handler) Overview(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	result, err := h.overview.GetOverview(c.Request.Context(), userID, OverviewQuery{LedgerID: c.Query("ledger_id")})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", result)
}

func (h *Handler) Trend(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	from := time.Time{}
	to := time.Time{}
	var err error
	if raw := c.Query("from"); raw != "" {
		from, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
	}
	if raw := c.Query("to"); raw != "" {
		to, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
	}
	timeout := time.Duration(0)
	if raw := c.Query("timeout_ms"); raw != "" {
		ms, convErr := time.ParseDuration(raw + "ms")
		if convErr != nil {
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		}
		timeout = ms
	}
	result, err := h.trend.GetTrend(c.Request.Context(), userID, TrendQuery{From: from, To: to, Granularity: c.DefaultQuery("granularity", "day"), Timezone: c.Query("timezone"), Timeout: timeout})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", result)
}

func (h *Handler) Category(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	result, err := h.category.GetCategoryStats(c.Request.Context(), userID, CategoryQuery{})
	if err != nil {
		h.writeError(c, err)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", result)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch err.Error() {
	case STAT_QUERY_INVALID:
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
	case STAT_TIMEOUT:
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
