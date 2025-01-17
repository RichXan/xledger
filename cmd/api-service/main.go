package main

import (
	"fmt"
	"log"
	"os"

	"xledger/cmd"
	"xledger/global"
	"xledger/internal/http"

	"github.com/RichXan/xcommon/xdatabase"
	"github.com/RichXan/xcommon/xlog"
	"github.com/RichXan/xcommon/xoauth"

	"github.com/urfave/cli/v2"
)

func main() {
	cmd.Welcome()
	app := &cli.App{
		Name:    "xledger",
		Usage:   "xledger service",
		Version: cmd.VERSION,
		Authors: cmd.Authors,
		Flags:   cmd.Flags,
		Before: func(c *cli.Context) error {
			// 打印版本信息
			if c.Bool("debug") {
				fmt.Printf("Version: %s\n", cmd.VERSION)
				fmt.Printf("BuildTime: %s\n", cmd.BuildTime)
				fmt.Printf("GitCommit: %s\n", cmd.GitCommit)
			}
			return nil
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:   "start",
			Usage:  "start http server",
			Action: startHttpServer,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func startHttpServer(c *cli.Context) error {
	// 加载配置
	if err := global.LoadConfig(c.String("config")); err != nil {
		return fmt.Errorf("load config error: %v", err)
	}

	// 根据环境设置覆盖配置
	global.Config.System.Env = c.String("env")
	if c.Bool("debug") {
		global.Config.Log.Level = "debug"
	}

	// 初始化日志
	logger := xlog.NewLogger(global.Config.Log)

	// 初始化数据库
	db, err := xdatabase.NewPostgresGormDb(&global.Config.Postgres)
	if err != nil {
		panic(fmt.Errorf("init database error: %v", err))
	}

	logger.Info().Msgf("global.Config.OAuth: %+v", global.Config.OAuth)
	claims, err := xoauth.NewClaimsWithKeyPairFromPEM(&global.Config.OAuth)
	if err != nil {
		panic(fmt.Errorf("init OAuth error: %v", err))
	}
	// 初始化Redis
	// redisClient, err := xcache.NewRedisClient(global.Config.Redis.MasterName, global.Config.Redis.Addresses, global.Config.Redis.Password, logger)
	// if err != nil {
	// 	panic(fmt.Errorf("init redis error: %v", err))
	// }
	// redisSimpleClient, ok := redisClient.Client().(*redis.Client)
	// if !ok {
	// 	panic(fmt.Errorf("redis client is not a simple client"))
	// }

	// 打印启动信息
	logger.Info().
		Str("version", cmd.VERSION).
		Str("env", global.Config.System.Env).
		Bool("debug", c.Bool("debug")).
		Msg("Starting service...")

	// 启动HTTP服务
	http.Start(logger, db, claims)
	return nil
}
