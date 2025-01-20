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

// SubCategoryService 子类目服务接口
type SubCategoryService interface {
	// 子类目基本功能
	Create(ctx context.Context, createDto *dto.SubCategoryCreate) (*model.SubCategory, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, updateDto *dto.SubCategoryUpdate) (*model.SubCategory, error)
	Get(ctx context.Context, id string) (*model.SubCategory, error)
	List(ctx context.Context, listDto *dto.SubCategoryList) ([]*model.SubCategory, int64, error)
}

type subCategoryService struct {
	logger          *xlog.Logger
	subCategoryRepo repo.SubCategoryRepository
}

func NewSubCategoryService(logger *xlog.Logger, subCategoryRepo repo.SubCategoryRepository) *subCategoryService {
	return &subCategoryService{
		logger:          logger,
		subCategoryRepo: subCategoryRepo,
	}
}

func (s *subCategoryService) Create(ctx context.Context, createDto *dto.SubCategoryCreate) (*model.SubCategory, error) {
	subCategory := &model.SubCategory{
		Name: createDto.Name,
	}

	if err := s.subCategoryRepo.Create(subCategory); err != nil {
		return nil, errors.New("failed to create subCategory")
	}

	return subCategory, nil
}

func (s *subCategoryService) Delete(ctx context.Context, id string) error {
	return s.subCategoryRepo.Delete(id)
}

// Update 更新子类目信息
func (s *subCategoryService) Update(ctx context.Context, updateDto *dto.SubCategoryUpdate) (*model.SubCategory, error) {
	subCategory, err := s.subCategoryRepo.GetByQuery(&model.SubCategory{UUIDModel: model.UUIDModel{ID: uuid.FromStringOrNil(updateDto.ID)}})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subCategory not found")
		}
		return nil, errors.New("failed to get subCategory")
	}

	// 更新子类目信息
	subCategory.Name = updateDto.Name

	if err := s.subCategoryRepo.Update(subCategory); err != nil {
		return nil, xerror.Wrap(err, xerror.CodeUpdateError, "failed to update subCategory")
	}

	return subCategory, nil
}

// Get 获取子类目信息
func (s *subCategoryService) Get(ctx context.Context, id string) (*model.SubCategory, error) {
	subCategory, err := s.subCategoryRepo.GetByQuery(&model.SubCategory{UUIDModel: model.UUIDModel{ID: uuid.FromStringOrNil(id)}})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("subCategory not found")
		}
		return nil, errors.New("failed to get subCategory")
	}
	return subCategory, nil
}

func (s *subCategoryService) List(ctx context.Context, listDto *dto.SubCategoryList) ([]*model.SubCategory, int64, error) {
	return s.subCategoryRepo.List(listDto.GetOffset(), listDto.GetLimit(), listDto.GetOrder())
}
