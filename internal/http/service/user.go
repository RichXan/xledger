package service

import (
	"context"
	"errors"
	"regexp"
	"xLedger/internal/access/model"
	"xLedger/internal/access/repo"

	"github.com/RichXan/xcommon/xerror"
	"github.com/RichXan/xcommon/xlog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

// Register 用户注册
func (s *userService) Register(ctx context.Context, username, password, email string) (*model.User, error) {
	// 验证用户名格式
	if !validateUsername(username) {
		return nil, xerror.UsernameInvalid
	}

	// 验证邮箱格式
	if !validateEmail(email) {
		return nil, xerror.EmailInvalid
	}

	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return nil, xerror.UserExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, xerror.Wrap(err, xerror.CodeSystemError, "failed to check username")
	}

	// 检查邮箱是否已存在
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, xerror.EmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, xerror.Wrap(err, xerror.CodeSystemError, "failed to check email")
	}

	// 创建用户
	user := &model.User{
		Username: username,
		Password: password,
		Email:    email,
		Nickname: username,
		Status:   1,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, xerror.Wrap(err, xerror.CodeSystemError, "failed to create user")
	}

	return user, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, username, password string) (string, error) {
	// 查找用户
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", xerror.UserNotFound
		}
		return "", xerror.Wrap(err, xerror.CodeSystemError, "failed to find user")
	}

	// 检查用户状态
	if user.Status != 1 {
		return "", xerror.UserDisabled
	}

	// 验证密码
	if !user.ComparePassword(password) {
		return "", xerror.PasswordError
	}

	// TODO: 生成访问令牌
	return "", nil
}

// GetUser 获取用户信息
func (s *userService) GetUser(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.userRepo.FindByID(uint64(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xerror.UserNotFound
		}
		return nil, xerror.Wrap(err, xerror.CodeSystemError, "failed to get user")
	}
	return user, nil
}

// UpdateUser ��新用户信息
func (s *userService) UpdateUser(ctx context.Context, userID int64, nickname, avatar, bio string) (*model.User, error) {
	user, err := s.userRepo.FindByID(uint64(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, xerror.UserNotFound
		}
		return nil, xerror.Wrap(err, xerror.CodeSystemError, "failed to get user")
	}

	// 更新用户信息
	user.Nickname = nickname
	user.Avatar = avatar
	// TODO: 添加 bio 字段

	if err := s.userRepo.Update(user); err != nil {
		return nil, xerror.Wrap(err, xerror.CodeSystemError, "failed to update user")
	}

	return user, nil
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(uint64(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return xerror.UserNotFound
		}
		return xerror.Wrap(err, xerror.CodeSystemError, "failed to get user")
	}

	// 验证旧密码
	if !user.ComparePassword(oldPassword) {
		return xerror.PasswordError
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return xerror.Wrap(err, xerror.CodeSystemError, "failed to hash password")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return xerror.Wrap(err, xerror.CodeSystemError, "failed to update password")
	}

	return nil
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.List(offset, pageSize)
}
