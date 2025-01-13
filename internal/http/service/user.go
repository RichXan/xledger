package service

import (
	"context"
	"errors"
	"regexp"
	"xledger/database/model"
	"xledger/database/repo"
	"xledger/internal/http/handler/dto"

	"github.com/RichXan/xcommon/xerror"
	"github.com/RichXan/xcommon/xlog"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	// 用户基本功能
	Create(ctx context.Context, createDto *dto.UserCreate) (*model.User, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, updateDto *dto.UserUpdate) (*model.User, error)
	Get(ctx context.Context, id string) (*model.User, error)
	List(ctx context.Context, listDto *dto.UserList) ([]*model.User, int64, error)
}

type userService struct {
	logger   *xlog.Logger
	userRepo repo.UserRepository
}

func NewUserService(logger *xlog.Logger, userRepo repo.UserRepository) *userService {
	return &userService{
		logger:   logger,
		userRepo: userRepo,
	}
}

func (s *userService) Create(ctx context.Context, createDto *dto.UserCreate) (*model.User, error) {
	// 验证username
	if !validateUsername(createDto.Username) {
		s.logger.Error().Msg("invalid username, username must be 4-16 characters long and can only contain letters, numbers, underscores, and hyphens")
		return nil, xerror.Wrap(xerror.ParamError, xerror.CodeParamError, "invalid username")
	}

	// 验证email
	if !validateEmail(createDto.Email) {
		return nil, xerror.Wrap(xerror.ParamError, xerror.CodeParamError, "invalid email")
	}

	user := &model.User{
		Username: createDto.Username,
		Password: createDto.Password,
		Email:    createDto.Email,
		Status:   model.UserStatusNormal,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	return s.userRepo.Delete(id)
}

// Update 更新用户信息
func (s *userService) Update(ctx context.Context, updateDto *dto.UserUpdate) (*model.User, error) {
	user, err := s.userRepo.GetByID(updateDto.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to get user")
	}

	// 更新用户信息
	user.Nickname = updateDto.Nickname
	user.Avatar = updateDto.Avatar
	if updateDto.Status != 0 {
		user.Status = model.UserStatus(updateDto.Status)
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, xerror.Wrap(err, xerror.CodeUpdateError, "failed to update user")
	}

	return user, nil
}

// Get 获取用户信息
func (s *userService) Get(ctx context.Context, id string) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to get user")
	}
	return user, nil
}

func (s *userService) List(ctx context.Context, listDto *dto.UserList) ([]*model.User, int64, error) {
	return s.userRepo.List(listDto.GetOffset(), listDto.GetLimit(), listDto.GetOrderString())
}

// validateEmail 验证邮箱格式
func validateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

// validateUsername 验证用户名格式
func validateUsername(username string) bool {
	pattern := `^[a-zA-Z0-9_-]{4,16}$`
	match, _ := regexp.MatchString(pattern, username)
	return match
}
