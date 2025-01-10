package handler

import (
	"net/http"
	"strconv"
	"time"

	"xledger/internal/access/model"
	"xledger/internal/http/service"

	"github.com/gin-gonic/gin"
)

type BillHandler struct {
	billService service.BillService
}

func NewBillHandler(billService service.BillService) *BillHandler {
	return &BillHandler{
		billService: billService,
	}
}

// RegisterRoutes 注册路由
func (h *BillHandler) RegisterRoutes(r *gin.RouterGroup) {
	bills := r.Group("/bills")
	{
		bills.POST("", h.CreateBill)
		bills.PUT("/:id", h.UpdateBill)
		bills.DELETE("/:id", h.DeleteBill)
		bills.GET("/:id", h.GetBill)
		bills.GET("", h.ListBills)
		bills.GET("/stats", h.GetBillStats)
	}
}

// CreateBill 创建账单
func (h *BillHandler) CreateBill(c *gin.Context) {
	var bill model.Bill
	if err := c.ShouldBindJSON(&bill); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置用户ID（从认证中获取）
	userID := uint(1) // TODO: 从认证中获取真实用户ID
	bill.UserID = userID

	if err := h.billService.CreateBill(c.Request.Context(), &bill); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bill)
}

// UpdateBill 更新账单
func (h *BillHandler) UpdateBill(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var bill model.Bill
	if err := c.ShouldBindJSON(&bill); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bill.ID = uint(id)
	if err := h.billService.UpdateBill(c.Request.Context(), &bill); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bill)
}

// DeleteBill 删除账单
func (h *BillHandler) DeleteBill(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.billService.DeleteBill(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetBill 获取账单详情
func (h *BillHandler) GetBill(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	bill, err := h.billService.GetBill(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if bill == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, bill)
}

// ListBills 获取账单列表
func (h *BillHandler) ListBills(c *gin.Context) {
	params := service.ListBillsParams{
		UserID:   uint(1), // TODO: 从认证中获取真实用户ID
		Page:     1,
		PageSize: 10,
	}

	// 解析查询参数
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			params.PageSize = ps
		}
	}
	if billType := c.Query("type"); billType != "" {
		params.Type = model.BillType(billType)
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		if cid, err := strconv.ParseUint(categoryID, 10, 32); err == nil {
			params.CategoryID = uint(cid)
		}
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if sd, err := time.Parse("2006-01-02", startDate); err == nil {
			params.StartDate = sd
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if ed, err := time.Parse("2006-01-02", endDate); err == nil {
			params.EndDate = ed
		}
	}

	bills, total, err := h.billService.ListBills(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": bills,
	})
}

// GetBillStats 获取账单统计
func (h *BillHandler) GetBillStats(c *gin.Context) {
	params := service.BillStatsParams{
		UserID: uint(1), // TODO: 从认证中获取真实用户ID
	}

	// 解析查询参数
	if billType := c.Query("type"); billType != "" {
		params.Type = model.BillType(billType)
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		if cid, err := strconv.ParseUint(categoryID, 10, 32); err == nil {
			params.CategoryID = uint(cid)
		}
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if sd, err := time.Parse("2006-01-02", startDate); err == nil {
			params.StartDate = sd
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if ed, err := time.Parse("2006-01-02", endDate); err == nil {
			params.EndDate = ed
		}
	}

	stats, err := h.billService.GetBillStats(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
