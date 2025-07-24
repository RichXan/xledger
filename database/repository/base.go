package repository

import (
	"xledger/internal/http/handler/dto"

	"gorm.io/gorm"
)

// IBaseRepository 基础仓储接口
type BaseRepository[T any] interface {
	// 支持事务的CRUD方法
	Create(model *T, tx ...*gorm.DB) error
	NormalCreate(model *T, tx ...*gorm.DB) error
	// upsert
	Upsert(model *T, tx ...*gorm.DB) error
	Update(model *T, tx ...*gorm.DB) error
	Deletes(models []*T, tx ...*gorm.DB) error
	ForceDeletes(models []*T, tx ...*gorm.DB) error
	GetByQuery(query *T, tx ...*gorm.DB) (*T, error)
	List(listDto dto.ListDto, tx ...*gorm.DB) ([]*T, int64, error)
	// 事务相关方法
	Begin() *gorm.DB
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB) error
	Transaction(fc func(tx *gorm.DB) error) error
}

// BaseRepository 基础仓储实现
type baseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) BaseRepository[T] {
	return &baseRepository[T]{db: db}
}

// Create 创建
func (r *baseRepository[T]) Create(model *T, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Model(model).Create(model).Error
}

// NormalCreate 创建
func (r *baseRepository[T]) NormalCreate(model *T, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Create(model).Error
}

// Upsert 存在更新，不存在插入
func (r *baseRepository[T]) Upsert(model *T, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Model(model).Where(model).Updates(model).Error
}

// Update 更新
func (r *baseRepository[T]) Update(model *T, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Model(model).Updates(model).Error
}

// Delete 删除
func (r *baseRepository[T]) Deletes(models []*T, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	return db.Where(models).Delete([]T{}).Error
}

// GetByQuery 根据查询条件查找
func (r *baseRepository[T]) GetByQuery(query *T, tx ...*gorm.DB) (*T, error) {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	var model T
	err := db.Where(query).First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}

// List 通用列表查询方法
func (r *baseRepository[T]) List(listDto dto.ListDto, tx ...*gorm.DB) ([]*T, int64, error) {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	var items []*T
	var total int64

	dbModel := db.Model(new(T))

	// 应用查询条件
	dbModel = listDto.BuildQuery(dbModel)

	// 获取总数
	err := dbModel.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	pageReq := listDto.GetPageReq()
	err = dbModel.Offset(pageReq.GetOffset()).
		Limit(pageReq.GetLimit()).
		Order(pageReq.GetOrder()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ForceDeletes 强制删除多个
func (r *baseRepository[T]) ForceDeletes(models []*T, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	return db.Unscoped().Where(models).Delete([]T{}).Error
}

// Begin 开始事务
func (r *baseRepository[T]) Begin() *gorm.DB {
	return r.db.Begin()
}

// Commit 提交事务
func (r *baseRepository[T]) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

// Rollback 回滚事务
func (r *baseRepository[T]) Rollback(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// Transaction 事务包装方法
func (r *baseRepository[T]) Transaction(fc func(tx *gorm.DB) error) error {
	return r.db.Transaction(fc)
}
