package repo

import (
	"xledger/database/model"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(id string) error
	List(offset, limit int, order string) ([]*model.User, int64, error)
	GetByQuery(query *model.User) (*model.User, error)
}
