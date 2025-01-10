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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	// 用户基本功能
	Create(ctx context.Context, createDto *dto.UserCreate) (*model.User, error)
	Delete(ctx context.Context, userID uint64) error
	Update(ctx context.Context, updateDto *dto.UserUpdate) (*model.User, error)
	Get(ctx context.Context, userID uint64) (*model.User, error)
	List(ctx context.Context, listDto *dto.UserList) ([]*model.User, int64, error)
	// Register(ctx context.Context, username, password, email string) (*model.User, error)
	// Login(ctx context.Context, username, password string) (string, error)
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

func (s *userService) Delete(ctx context.Context, userID uint64) error {
	return s.userRepo.Delete(userID)
}

// Update 更新用户信息
func (s *userService) Update(ctx context.Context, updateDto *dto.UserUpdate) (*model.User, error) {
	user, err := s.userRepo.FindByID(updateDto.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to get user")
	}

	// 更新用户信息
	user.Nickname = updateDto.Nickname
	user.Avatar = updateDto.Avatar

	if err := s.userRepo.Update(user); err != nil {
		return nil, xerror.Wrap(err, xerror.CodeSystemError, "failed to update user")
	}

	return user, nil
}

// Get 获取用户信息
func (s *userService) Get(ctx context.Context, userID uint64) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to get user")
	}
	return user, nil
}

func (s *userService) List(ctx context.Context, listDto *dto.UserList) ([]*model.User, int64, error) {
	return s.userRepo.List(listDto.Page.Page, listDto.Page.PageSize)
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return errors.New("failed to get user")
	}

	// 验证旧密码
	if !user.ComparePassword(oldPassword) {
		return errors.New("password error")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// 更新密码
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	offset := (page - 1) * pageSize
	return s.userRepo.List(offset, pageSize)
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, username, password, email string) (*model.User, error) {
	// 验证用户名格式
	if !validateUsername(username) {
		return nil, errors.New("invalid username")
	}

	// 验证邮箱格式
	if !validateEmail(email) {
		return nil, errors.New("invalid email")
	}

	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return nil, errors.New("username exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("failed to check username")
	}

	// 检查邮箱是否已存在
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, errors.New("email exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("failed to check email")
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
			return "", xerror.New(xerror.CodeParamError, ", user not found")
		}
		return "", errors.New("failed to find user")
	}

	// 检查用户状态
	if user.Status != 1 {
		return "", errors.New("user disabled")
	}

	// 验证密码
	if !user.ComparePassword(password) {
		return "", errors.New("password error")
	}

	// TODO: 生成访问令牌
	return "", nil
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
