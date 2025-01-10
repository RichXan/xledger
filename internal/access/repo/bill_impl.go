package repo

import (
	"context"
	"errors"

	"xledger/internal/access/model"

	"gorm.io/gorm"
)

// billRepository 账单仓储实现
type billRepository struct {
	db *gorm.DB
}

// NewBillRepository 创建账单仓储实例
func NewBillRepository(db *gorm.DB) BillRepository {
	return &billRepository{db: db}
}

// Create 创建账单
func (r *billRepository) Create(ctx context.Context, bill *model.Bill) error {
	return r.db.WithContext(ctx).Create(bill).Error
}

// Update 更新账单
func (r *billRepository) Update(ctx context.Context, bill *model.Bill) error {
	return r.db.WithContext(ctx).Save(bill).Error
}

// Delete 删除账单
func (r *billRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Bill{}, id).Error
}

// Get 获取账单详情
func (r *billRepository) Get(ctx context.Context, id uint) (*model.Bill, error) {
	var bill model.Bill
	err := r.db.WithContext(ctx).First(&bill, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &bill, nil
}

// List 获取账单列表
func (r *billRepository) List(ctx context.Context, params ListBillParams) ([]*model.Bill, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Bill{})

	// 添加查询条件
	if params.UserID > 0 {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.CategoryID > 0 {
		query = query.Where("category_id = ?", params.CategoryID)
	}
	if !params.StartDate.IsZero() {
		query = query.Where("bill_date >= ?", params.StartDate)
	}
	if !params.EndDate.IsZero() {
		query = query.Where("bill_date <= ?", params.EndDate)
	}
	if len(params.TagIDs) > 0 {
		query = query.Joins("JOIN bill_tag_relations ON bills.id = bill_tag_relations.bill_id").
			Where("bill_tag_relations.tag_id IN ?", params.TagIDs)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	var bills []*model.Bill
	err := query.Offset(offset).Limit(params.PageSize).
		Order("bill_date DESC, id DESC").
		Find(&bills).Error
	if err != nil {
		return nil, 0, err
	}

	return bills, total, nil
}

// GetStats 获取账单统计
func (r *billRepository) GetStats(ctx context.Context, params BillStatsParams) (*BillStats, error) {
	query := r.db.WithContext(ctx).Model(&model.Bill{})

	// 添加查询条件
	if params.UserID > 0 {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.CategoryID > 0 {
		query = query.Where("category_id = ?", params.CategoryID)
	}
	if !params.StartDate.IsZero() {
		query = query.Where("bill_date >= ?", params.StartDate)
	}
	if !params.EndDate.IsZero() {
		query = query.Where("bill_date <= ?", params.EndDate)
	}

	stats := &BillStats{}

	// 计算总金额
	var totalAmount float64
	err := query.Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount).Error
	if err != nil {
		return nil, err
	}
	stats.TotalAmount = totalAmount

	// 分类统计
	var categoryStats []CategoryStat
	err = query.Select("category_id, SUM(amount) as amount").
		Group("category_id").
		Scan(&categoryStats).Error
	if err != nil {
		return nil, err
	}

	// 计算百分比
	for i := range categoryStats {
		if totalAmount > 0 {
			categoryStats[i].Percentage = categoryStats[i].Amount / totalAmount * 100
		}
	}
	stats.CategoryStats = categoryStats

	// 每日统计
	var dailyStats []DailyStat
	err = query.Select("DATE(bill_date) as date, SUM(amount) as amount").
		Group("DATE(bill_date)").
		Order("date").
		Scan(&dailyStats).Error
	if err != nil {
		return nil, err
	}
	stats.DailyStats = dailyStats

	return stats, nil
}
