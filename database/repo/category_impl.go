package repo

import (
	"xledger/database/model"

	"gorm.io/gorm"
)

// categoryRepository 类目仓库实现
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建类目仓库
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// Create 创建类目
func (r *categoryRepository) Create(category *model.Category) error {
	return r.db.Model(category).Create(category).Error
}

// Update 更新类目
func (r *categoryRepository) Update(category *model.Category) error {
	return r.db.Model(category).Updates(category).Error
}

// Delete 删除类目
func (r *categoryRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Category{}).Error
}

// GetByQuery 根据查询条件查找类目
func (r *categoryRepository) GetByQuery(query *model.Category) (*model.Category, error) {
	var category model.Category
	err := r.db.Where(query).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// List 获取类目列表
func (r *categoryRepository) List(offset, limit int, order string) ([]*model.Category, int64, error) {
	var categories []*model.Category
	var total int64

	err := r.db.Model(&model.Category{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Order(order).Find(&categories).Error
	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}
