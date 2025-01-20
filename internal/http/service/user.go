package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"xledger/database/model"
	"xledger/database/repo"
	"xledger/internal/http/handler/dto"

	"github.com/RichXan/xcommon/xerror"
	"github.com/RichXan/xcommon/xlog"
	"github.com/RichXan/xcommon/xoauth"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	// 用户基本功能
	Create(ctx context.Context, createDto *dto.UserCreate) (*model.User, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, updateDto *dto.UserUpdate) (*model.User, error)
	Get(ctx context.Context, id string) (*model.User, error)
	Login(ctx context.Context, loginDto *dto.UserLogin) (*xoauth.TokenPair, error)
	RefreshToken(ctx context.Context, refreshTokenDto *dto.UserRefreshToken) (*xoauth.TokenPair, error)
	List(ctx context.Context, listDto *dto.UserList) ([]*model.User, int64, error)
	Register(ctx context.Context, registerDto *dto.UserRegister) error
}

type userService struct {
	logger    *xlog.Logger
	jwtClaims xoauth.Claim
	userRepo  repo.UserRepository
}

func NewUserService(logger *xlog.Logger, jwtClaims xoauth.Claim, userRepo repo.UserRepository) *userService {
	return &userService{
		logger:    logger,
		jwtClaims: jwtClaims,
		userRepo:  userRepo,
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
	user, err := s.userRepo.GetByQuery(&model.User{UUIDModel: model.UUIDModel{ID: uuid.FromStringOrNil(updateDto.ID)}})
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
	user, err := s.userRepo.GetByQuery(&model.User{UUIDModel: model.UUIDModel{ID: uuid.FromStringOrNil(id)}})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to get user")
	}
	return user, nil
}

func (s *userService) List(ctx context.Context, listDto *dto.UserList) ([]*model.User, int64, error) {
	return s.userRepo.List(listDto.GetOffset(), listDto.GetLimit(), listDto.GetOrder())
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

func (s *userService) Login(ctx context.Context, loginDto *dto.UserLogin) (*xoauth.TokenPair, error) {
	user, err := s.userRepo.GetByQuery(&model.User{Username: loginDto.Username})
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginDto.Password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	// 生成JWT token
	return s.jwtClaims.GenerateTokenPair(xoauth.Info{
		UserID:   user.ID.String(),
		Username: user.Username,
	})
}

func (s *userService) Register(ctx context.Context, registerDto *dto.UserRegister) error {
	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByQuery(&model.User{Username: registerDto.Username})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check username: %v", err)
	}
	if existingUser != nil {
		return errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if registerDto.Email != "" {
		existingUser, err = s.userRepo.GetByQuery(&model.User{Email: registerDto.Email})
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check email: %v", err)
		}
		if existingUser != nil {
			return errors.New("email already exists")
		}
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerDto.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// 创建用户
	user := &model.User{
		Username: registerDto.Username,
		Password: string(hashedPassword),
		Email:    registerDto.Email,
	}

	if err := s.userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshTokenDto *dto.UserRefreshToken) (*xoauth.TokenPair, error) {
	return s.jwtClaims.RefreshTokenPair(refreshTokenDto.RefreshToken)
}
