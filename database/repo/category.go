package repo

import "xledger/database/model"

// CategoryRepository 类目仓库接口
type CategoryRepository interface {
	Create(category *model.Category) error
	Update(category *model.Category) error
	Delete(id string) error
	GetByQuery(query *model.Category) (*model.Category, error)
	List(offset, limit int, order string) ([]*model.Category, int64, error)
}
