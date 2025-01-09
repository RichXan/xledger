package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xLedger/global"
	"xLedger/internal/http/router"

	"github.com/RichXan/xcommon/xlog"
	"github.com/RichXan/xcommon/xutil"
	"gorm.io/gorm"

	"github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

var (
	srv    *http.Server
	logger *xlog.Logger
)

// initJaeger 初始化Jaeger
func initJaeger() (opentracing.Tracer, error) {
	cfg := jaegercfg.Configuration{
		ServiceName: global.Config.System.Name,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  global.Config.Jaeger.SamplerType,
			Param: global.Config.Jaeger.SamplerParam,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: global.Config.Jaeger.LogSpans,
			LocalAgentHostPort: fmt.Sprintf("%s:%d",
				global.Config.Jaeger.Host,
				global.Config.Jaeger.Port,
			),
		},
	}

	tracer, _, err := cfg.NewTracer()
	if err != nil {
		return nil, err
	}

	opentracing.SetGlobalTracer(tracer)
	return tracer, nil
}

// Start 启动HTTP服务
func Start(l *xlog.Logger, db *gorm.DB) {
	logger = l

	// 初始化Jaeger
	tracer, err := initJaeger()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize Jaeger")
		os.Exit(1)
	}

	r := router.Setup(tracer, logger, db)
	srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", global.Config.Server.HTTPPort),
		Handler: r,
	}

	// 启动HTTP服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Failed to start server")
			os.Exit(1)
		}
	}()

	logger.Info().Int("port", global.Config.Server.HTTPPort).Msg("Server started")
	sendEmail(logger)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
		os.Exit(1)
	}

	logger.Info().Msg("Server exited")
}

func sendEmail(logger *xlog.Logger) {
	// 初始化SMTP
	smtpClient := xutil.NewSMTPClient(global.Config.SMTP)
	subject := fmt.Sprintf("%s started successfully", global.Config.System.Name)
	body := fmt.Sprintf(" server name: %s\n server description: %s\n server port: %d\n server env: %s\n server version: %s\n",
		global.Config.System.Name,
		global.Config.System.Description,
		global.Config.Server.HTTPPort,
		global.Config.System.Env,
		global.Config.System.Version,
	)

	err := smtpClient.SendEmail(xutil.EmailParams{
		Subject:  subject,
		Body:     body,
		BodyType: xutil.PLAIN,
	})
	if err != nil {
		logger.Error().Err(err).Msg("send email error")
	}
	logger.Info().Msg("send email success")
}
