package budget

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "xledger/backend/internal/common/httpx"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) ListBudgets(c *gin.Context) {
    userID, ok := c.Get("user_id")
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    budgets, err := h.service.GetUserBudgets(c.Request.Context(), userID.(string))
    if err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }

    httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"budgets": budgets})
}

func (h *Handler) CreateBudget(c *gin.Context) {
    userID, ok := c.Get("user_id")
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    var req struct {
        CategoryID string  `json:"category_id" binding:"required"`
        Amount    float64 `json:"amount" binding:"required"`
        AlertAt   float64 `json:"alert_at"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
        return
    }

    budget, err := h.service.CreateBudget(c.Request.Context(), userID.(string), req.CategoryID, req.Amount, req.AlertAt)
    if err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }

    httpx.JSON(c, http.StatusCreated, "OK", "成功", budget)
}

func (h *Handler) UpdateBudget(c *gin.Context) {
    id := c.Param("id")

    var req struct {
        Amount  float64 `json:"amount" binding:"required"`
        AlertAt float64 `json:"alert_at"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
        return
    }

    budget, err := h.service.UpdateBudget(c.Request.Context(), id, req.Amount, req.AlertAt)
    if err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }

    httpx.JSON(c, http.StatusOK, "OK", "成功", budget)
}

func (h *Handler) DeleteBudget(c *gin.Context) {
    id := c.Param("id")

    if err := h.service.DeleteBudget(c.Request.Context(), id); err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }

    httpx.JSON(c, http.StatusOK, "OK", "成功", nil)
}

func (h *Handler) ListAlerts(c *gin.Context) {
    userID, ok := c.Get("user_id")
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    limit := 20
    if l := c.Query("limit"); l != "" {
        if parsed, err := strconv.Atoi(l); err == nil {
            limit = parsed
        }
    }

    alerts, err := h.service.ListAlerts(c.Request.Context(), userID.(string), limit)
    if err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }

    httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"alerts": alerts})
}

func (h *Handler) GetPreferences(c *gin.Context) {
    userID, ok := c.Get("user_id")
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    prefs, err := h.service.GetPreference(c.Request.Context(), userID.(string))
    if err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }
    httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"prefs": prefs})
}

func (h *Handler) UpdatePreferences(c *gin.Context) {
    userID, ok := c.Get("user_id")
    if !ok {
        httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证", nil)
        return
    }

    var req struct {
        RealtimeAlert bool `json:"realtime_alert"`
        DailyDigest   bool `json:"daily_digest"`
        WeeklyDigest  bool `json:"weekly_digest"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
        return
    }

    pref := &UserNotificationPref{
        UserID:         userID.(string),
        RealtimeAlert:  req.RealtimeAlert,
        DailyDigest:    req.DailyDigest,
        WeeklyDigest:   req.WeeklyDigest,
    }
    if err := h.service.UpdatePreference(c.Request.Context(), pref); err != nil {
        httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
        return
    }
    httpx.JSON(c, http.StatusOK, "OK", "成功", nil)
}
