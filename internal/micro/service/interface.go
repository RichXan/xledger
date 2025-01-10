package service

import (
	"context"
	"xledger/database/model"
	"xledger/internal/http/handler/dto"
)

// UserService 用户服务接口
type UserService interface {
	// 用户基本功能
	Create(ctx context.Context, createDto *dto.UserCreate) (*model.User, error)
	Delete(ctx context.Context, userID uint64) error
	Update(ctx context.Context, updateDto *dto.UserUpdate) (*model.User, error)
	Get(ctx context.Context, userID uint64) (*model.User, error)
	List(ctx context.Context, listDto *dto.UserList) ([]*model.User, int64, error)
}
