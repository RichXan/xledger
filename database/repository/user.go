package repository

import (
	"xledger/database/model"

	"gorm.io/gorm"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	BaseRepository[model.User]
}

// userRepository 用户仓库实现
type userRepository struct {
	baseRepository[model.User]
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		baseRepository: baseRepository[model.User]{db: db},
	}
}
