package repo

import (
	"xledger/database/model"

	"gorm.io/gorm"
)

// SubCategoryRepository 子类目仓库接口
type SubCategoryRepository interface {
	Create(subCategory *model.SubCategory) error
	Update(subCategory *model.SubCategory) error
	Delete(id string) error
	GetByQuery(query *model.SubCategory) (*model.SubCategory, error)
	List(offset, limit int, order string) ([]*model.SubCategory, int64, error)
}

// subCategoryRepository 子类目仓库实现
type subCategoryRepository struct {
	db *gorm.DB
}

// NewSubCategoryRepository 创建子类目仓库
func NewSubCategoryRepository(db *gorm.DB) SubCategoryRepository {
	return &subCategoryRepository{db: db}
}

// Create 创建子类目
func (r *subCategoryRepository) Create(subCategory *model.SubCategory) error {
	return r.db.Model(subCategory).Create(subCategory).Error
}

// Update 更新子类目
func (r *subCategoryRepository) Update(subCategory *model.SubCategory) error {
	return r.db.Model(subCategory).Updates(subCategory).Error
}

// Delete 删除子类目
func (r *subCategoryRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.SubCategory{}).Error
}

// GetByQuery 根据查询条件查找子类目
func (r *subCategoryRepository) GetByQuery(query *model.SubCategory) (*model.SubCategory, error) {
	var subCategory model.SubCategory
	err := r.db.Where(query).First(&subCategory).Error
	if err != nil {
		return nil, err
	}
	return &subCategory, nil
}

// List 获取子类目列表
func (r *subCategoryRepository) List(offset, limit int, order string) ([]*model.SubCategory, int64, error) {
	var subCategories []*model.SubCategory
	var total int64

	err := r.db.Model(&model.SubCategory{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Order(order).Find(&subCategories).Error
	if err != nil {
		return nil, 0, err
	}

	return subCategories, total, nil
}
