package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	pb "xLedger/internal/micro/proto/user"

	"github.com/RichXan/xcommon/xlog"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go-micro.dev/v4"
	"gorm.io/gorm"
)

const userServiceName = "user-service"

// StartUserService 启动用户服务
func StartUserService(logger *xlog.Logger, db *gorm.DB) error {
	// 创建服务
	service := micro.NewService(
		micro.Name(userServiceName),
		micro.Version("latest"),
	)

	// 初始化服务
	service.Init()

	// 注册处理器
	if err := pb.RegisterUserServiceHandler(service.Server(), NewUserHandler(logger, db)); err != nil {
		return fmt.Errorf("failed to register handler: %v", err)
	}

	// 启动 metrics 服务器
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9090", nil); err != nil {
			logger.Error().Err(err).Msg("metrics server error")
		}
	}()

	// 运行服务
	if err := service.Run(); err != nil {
		return fmt.Errorf("failed to run service: %v", err)
	}

	return nil
}

// UserHandler 用户服务处理器
type UserHandler struct {
	logger *xlog.Logger
	db     *gorm.DB
}

// NewUserHandler 创建用户服务处理器
func NewUserHandler(logger *xlog.Logger, db *gorm.DB) *UserHandler {
	return &UserHandler{logger: logger, db: db}
}

// Register 处理用户注册
func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest, rsp *pb.RegisterResponse) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Register")
	defer span.Finish()

	h.logger.Info().
		Str("username", req.Username).
		Str("email", req.Email).
		Msg("Processing user registration")

	// TODO: 实现用户注册逻辑

	return nil
}

// Login 处理用户登录
func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest, rsp *pb.LoginResponse) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Login")
	defer span.Finish()

	h.logger.Info().
		Str("username", req.Username).
		Msg("Processing user login")

	// TODO: 实现用户登录逻辑

	return nil
}

// GetUserInfo 获取用户信息
func (h *UserHandler) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest, rsp *pb.GetUserInfoResponse) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetUserInfo")
	defer span.Finish()

	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", req.UserId).Msg("Invalid user ID format")
		return fmt.Errorf("invalid user ID format: %v", err)
	}

	h.logger.Info().
		Uint64("user_id", userID).
		Msg("Getting user info")

	// TODO: 实现获取用户信息逻辑

	return nil
}

// UpdateUserInfo 更新用户信息
func (h *UserHandler) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest, rsp *pb.UpdateUserInfoResponse) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UpdateUserInfo")
	defer span.Finish()

	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", req.UserId).Msg("Invalid user ID format")
		return fmt.Errorf("invalid user ID format: %v", err)
	}

	h.logger.Info().
		Uint64("user_id", userID).
		Msg("Updating user info")

	// TODO: 实现更新用户信息逻辑

	return nil
}
