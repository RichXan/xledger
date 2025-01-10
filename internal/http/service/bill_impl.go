package service

import (
	"context"

	"xledger/internal/access/model"
	"xledger/internal/access/repo"
)

// billService 账单服务实现
type billService struct {
	billRepo repo.BillRepository
}

// NewBillService 创建账单服务实例
func NewBillService(billRepo repo.BillRepository) BillService {
	return &billService{billRepo: billRepo}
}

// CreateBill 创建账单
func (s *billService) CreateBill(ctx context.Context, bill *model.Bill) error {
	return s.billRepo.Create(ctx, bill)
}

// UpdateBill 更新账单
func (s *billService) UpdateBill(ctx context.Context, bill *model.Bill) error {
	return s.billRepo.Update(ctx, bill)
}

// DeleteBill 删除账单
func (s *billService) DeleteBill(ctx context.Context, id uint) error {
	return s.billRepo.Delete(ctx, id)
}

// GetBill 获取账单详情
func (s *billService) GetBill(ctx context.Context, id uint) (*model.Bill, error) {
	return s.billRepo.Get(ctx, id)
}

// ListBills 获取账单列表
func (s *billService) ListBills(ctx context.Context, params ListBillsParams) ([]*model.Bill, int64, error) {
	repoParams := repo.ListBillParams{
		UserID:     params.UserID,
		Type:       params.Type,
		CategoryID: params.CategoryID,
		TagIDs:     params.TagIDs,
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}
	return s.billRepo.List(ctx, repoParams)
}

// GetBillStats 获取账单统计
func (s *billService) GetBillStats(ctx context.Context, params BillStatsParams) (*BillStats, error) {
	repoParams := repo.BillStatsParams{
		UserID:     params.UserID,
		Type:       params.Type,
		CategoryID: params.CategoryID,
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
	}

	repoStats, err := s.billRepo.GetStats(ctx, repoParams)
	if err != nil {
		return nil, err
	}

	// 转换分类统计数据
	categoryStats := make([]CategoryStat, len(repoStats.CategoryStats))
	for i, stat := range repoStats.CategoryStats {
		categoryStats[i] = CategoryStat{
			CategoryID:   stat.CategoryID,
			CategoryName: stat.CategoryName,
			Amount:       stat.Amount,
			Percentage:   stat.Percentage,
		}
	}

	// 转换每日统计数据
	dailyStats := make([]DailyStat, len(repoStats.DailyStats))
	for i, stat := range repoStats.DailyStats {
		dailyStats[i] = DailyStat{
			Date:   stat.Date,
			Amount: stat.Amount,
		}
	}

	return &BillStats{
		TotalAmount:   repoStats.TotalAmount,
		CategoryStats: categoryStats,
		DailyStats:    dailyStats,
	}, nil
}
