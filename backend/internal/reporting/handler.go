package reporting

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}
	result, err := h.overview.GetOverview(c.Request.Context(), userID, OverviewQuery{LedgerID: c.Query("ledger_id")})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) Trend(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}
	from := time.Time{}
	to := time.Time{}
	var err error
	if raw := c.Query("from"); raw != "" {
		from, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error_code": STAT_QUERY_INVALID})
			return
		}
	}
	if raw := c.Query("to"); raw != "" {
		to, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error_code": STAT_QUERY_INVALID})
			return
		}
	}
	timeout := time.Duration(0)
	if raw := c.Query("timeout_ms"); raw != "" {
		ms, convErr := time.ParseDuration(raw + "ms")
		if convErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error_code": STAT_QUERY_INVALID})
			return
		}
		timeout = ms
	}
	result, err := h.trend.GetTrend(c.Request.Context(), userID, TrendQuery{From: from, To: to, Granularity: c.DefaultQuery("granularity", "day"), Timezone: c.Query("timezone"), Timeout: timeout})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) Category(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error_code": "AUTH_UNAUTHORIZED"})
		return
	}
	result, err := h.category.GetCategoryStats(c.Request.Context(), userID, CategoryQuery{})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch err.Error() {
	case STAT_QUERY_INVALID:
		c.JSON(http.StatusBadRequest, gin.H{"error_code": STAT_QUERY_INVALID})
	case STAT_TIMEOUT:
		c.JSON(http.StatusGatewayTimeout, gin.H{"error_code": STAT_TIMEOUT})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "REPORTING_INTERNAL"})
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
