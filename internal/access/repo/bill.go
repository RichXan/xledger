package repo

import (
	"context"
	"time"

	"xledger/internal/access/model"
)

// BillRepository 账单仓储接口
type BillRepository interface {
	// Create 创建账单
	Create(ctx context.Context, bill *model.Bill) error

	// Update 更新账单
	Update(ctx context.Context, bill *model.Bill) error

	// Delete 删除账单
	Delete(ctx context.Context, id uint) error

	// Get 获取账单详情
	Get(ctx context.Context, id uint) (*model.Bill, error)

	// List 获取账单列表
	List(ctx context.Context, params ListBillParams) ([]*model.Bill, int64, error)

	// GetStats 获取账单统计
	GetStats(ctx context.Context, params BillStatsParams) (*BillStats, error)
}

// ListBillParams 获取账单列表参数
type ListBillParams struct {
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
