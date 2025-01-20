package repo

import (
	"xledger/database/model"

	"gorm.io/gorm"
)

// userRepository 用户仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.db.Model(user).Create(user).Error
}

// Update 更新用户
func (r *userRepository) Update(user *model.User) error {
	return r.db.Model(user).Updates(user).Error
}

// Delete 删除用户
func (r *userRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.User{}).Error
}

// GetByQuery 根据查询条件查找用户
func (r *userRepository) GetByQuery(query *model.User) (*model.User, error) {
	var user model.User
	err := r.db.Where(query).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表
func (r *userRepository) List(offset, limit int, order string) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	err := r.db.Model(&model.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Order(order).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
