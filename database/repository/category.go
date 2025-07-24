package repository

import (
	"xledger/database/model"

	"gorm.io/gorm"
)

// CategoryRepository 类目仓库接口
type CategoryRepository interface {
	BaseRepository[model.Category]
}

// categoryRepository 类目仓库实现
type categoryRepository struct {
	baseRepository[model.Category]
}

// NewCategoryRepository 创建类目仓库
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{
		baseRepository: baseRepository[model.Category]{db: db},
	}
}
