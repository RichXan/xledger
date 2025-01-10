package service

import (
	"context"
	"xledger/internal/access/model"
)

// UserService 用户服务接口
type UserService interface {
	// 用户基本功能
	Register(ctx context.Context, username, password, email string) (*model.User, error)
	Login(ctx context.Context, username, password string) (string, error)
	// Get(ctx context.Context, userID int64) (*model.User, error)
	// Update(ctx context.Context, userID int64, nickname, avatar, bio string) (*model.User, error)
	// ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	// List(ctx context.Context, page, pageSize int) ([]*model.User, int64, error)
}
