package router

import (
	"net/http"
	"xledger/database/repo"
	"xledger/global"
	"xledger/internal/http/handler"
	"xledger/internal/http/middleware"
	"xledger/internal/http/service"

	"github.com/RichXan/xcommon/xlog"
	"github.com/RichXan/xcommon/xmiddleware"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	userHandler *handler.UserHandler
)

var (
	userService service.UserService
)

var (
	userRepo repo.UserRepository
)

// Setup 设置路由
func Setup(
	tracer opentracing.Tracer,
	logger *xlog.Logger,
	db *gorm.DB,
) *gin.Engine {
	// 初始化服务依赖
	initRepo(db)
	initService(logger)
	initHandler(logger)

	r := gin.New()

	// 使用中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(xmiddleware.Cors())
	r.Use(xmiddleware.RequestID())
	r.Use(xmiddleware.Logger(logger, global.Config.System.Debug))
	r.Use(xmiddleware.TracingMiddleware(tracer))
	r.Use(middleware.MetricsMiddleware())
	r.Use(xmiddleware.TimeFormat)

	// 健康检查
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API版本
	v1 := r.Group("/api/v1")
	{
		setupUserRoutes(v1)
	}

	return r
}

func initRepo(db *gorm.DB) {
	userRepo = repo.NewUserRepository(db)
}

func initService(logger *xlog.Logger) {
	userService = service.NewUserService(logger, userRepo)
}

func initHandler(logger *xlog.Logger) {
	userHandler = handler.NewUserHandler(logger, userService)
}
