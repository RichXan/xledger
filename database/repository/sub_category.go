package repository

import (
	"xledger/database/model"

	"gorm.io/gorm"
)

// SubCategoryRepository 子类目仓库接口
type SubCategoryRepository interface {
	BaseRepository[model.SubCategory]
}

// subCategoryRepository 子类目仓库实现
type subCategoryRepository struct {
	baseRepository[model.SubCategory]
}

// NewSubCategoryRepository 创建子类目仓库
func NewSubCategoryRepository(db *gorm.DB) SubCategoryRepository {
	return &subCategoryRepository{
		baseRepository: baseRepository[model.SubCategory]{db: db},
	}
}
