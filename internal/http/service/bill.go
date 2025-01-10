package service

import (
	"context"
	"time"

	"xledger/internal/access/model"
)

// BillService 账单服务接口
type BillService interface {
	// CreateBill 创建账单
	CreateBill(ctx context.Context, bill *model.Bill) error

	// UpdateBill 更新账单
	UpdateBill(ctx context.Context, bill *model.Bill) error

	// DeleteBill 删除账单
	DeleteBill(ctx context.Context, id uint) error

	// GetBill 获取账单详情
	GetBill(ctx context.Context, id uint) (*model.Bill, error)

	// ListBills 获取账单列表
	ListBills(ctx context.Context, params ListBillsParams) ([]*model.Bill, int64, error)

	// GetBillStats 获取账单统计
	GetBillStats(ctx context.Context, params BillStatsParams) (*BillStats, error)
}

// ListBillsParams 获取账单列表参数
type ListBillsParams struct {
	UserID     uint
	Type       model.BillType
	CategoryID uint
	TagIDs     []uint
	StartDate  time.Time
	EndDate    time.Time
	Page       int
	PageSize   int
}

// BillStatsParams 账单统计参数
type BillStatsParams struct {
	UserID     uint
	Type       model.BillType
	CategoryID uint
	StartDate  time.Time
	EndDate    time.Time
}

// BillStats 账单统计结果
type BillStats struct {
	TotalAmount   float64        `json:"total_amount"`   // 总金额
	CategoryStats []CategoryStat `json:"category_stats"` // 分类统计
	DailyStats    []DailyStat    `json:"daily_stats"`    // 每日统计
}

// CategoryStat 分类统计
type CategoryStat struct {
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
	Percentage   float64 `json:"percentage"`
}

// DailyStat 每日统计
type DailyStat struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
}
