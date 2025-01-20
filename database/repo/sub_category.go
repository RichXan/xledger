package repo

import "xledger/database/model"

// SubCategoryRepository 子类目仓库接口
type SubCategoryRepository interface {
	Create(subCategory *model.SubCategory) error
	Update(subCategory *model.SubCategory) error
	Delete(id string) error
	GetByQuery(query *model.SubCategory) (*model.SubCategory, error)
	List(offset, limit int, order string) ([]*model.SubCategory, int64, error)
}
