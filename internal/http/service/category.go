package service

import (
	"context"
	"errors"
	"xledger/database/model"
	"xledger/database/repo"
	"xledger/internal/http/handler/dto"

	"github.com/RichXan/xcommon/xerror"
	"github.com/RichXan/xcommon/xlog"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// CategoryService 类目服务接口
type CategoryService interface {
	// 类目基本功能
	Create(ctx context.Context, createDto *dto.CategoryCreate) (*model.Category, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, updateDto *dto.CategoryUpdate) (*model.Category, error)
	Get(ctx context.Context, id string) (*model.Category, error)
	List(ctx context.Context, listDto *dto.CategoryList) ([]*model.Category, int64, error)
}

type categoryService struct {
	logger       *xlog.Logger
	categoryRepo repo.CategoryRepository
}

func NewCategoryService(logger *xlog.Logger, categoryRepo repo.CategoryRepository) *categoryService {
	return &categoryService{
		logger:       logger,
		categoryRepo: categoryRepo,
	}
}

func (s *categoryService) Create(ctx context.Context, createDto *dto.CategoryCreate) (*model.Category, error) {
	category := &model.Category{
		Name: createDto.Name,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, errors.New("failed to create category")
	}

	return category, nil
}

func (s *categoryService) Delete(ctx context.Context, id string) error {
	return s.categoryRepo.Delete(id)
}

// Update 更新类目信息
func (s *categoryService) Update(ctx context.Context, updateDto *dto.CategoryUpdate) (*model.Category, error) {
	category, err := s.categoryRepo.GetByQuery(&model.Category{
		UUIDModel: model.UUIDModel{ID: uuid.FromStringOrNil(updateDto.ID)},
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, errors.New("failed to get category")
	}

	// 更新类目信息
	category.Name = updateDto.Name
	category.IsSystem = updateDto.IsSystem

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, xerror.Wrap(err, xerror.CodeUpdateError, "failed to update category")
	}

	return category, nil
}

// Get 获取类目信息
func (s *categoryService) Get(ctx context.Context, id string) (*model.Category, error) {
	category, err := s.categoryRepo.GetByQuery(&model.Category{
		UUIDModel: model.UUIDModel{ID: uuid.FromStringOrNil(id)},
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, errors.New("failed to get category")
	}
	return category, nil
}

func (s *categoryService) List(ctx context.Context, listDto *dto.CategoryList) ([]*model.Category, int64, error) {
	// 获取用户创建的类目和系统默认类目
	return s.categoryRepo.List(listDto.GetOffset(), listDto.GetLimit(), listDto.GetOrder())
}
