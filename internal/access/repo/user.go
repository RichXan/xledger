package repo

import (
	"xledger/internal/access/model"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(id uint64) error
	FindByID(id uint64) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	List(offset, limit int) ([]*model.User, int64, error)
}
