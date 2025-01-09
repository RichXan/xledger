package cmd

import "github.com/urfave/cli/v2"

var (
	VERSION = "0.1.0"
	// BuildTime 构建时间
	BuildTime = "unknown"
	// GitCommit Git提交哈希
	GitCommit = "unknown"
)

var (
	Authors = []*cli.Author{
		{
			Name:  "xan",
			Email: "rich4xan@gmail.com",
		},
	}

	Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Value:   "config/config.yml",
			Usage:   "配置文件路径",
			EnvVars: []string{"xledger_CONFIG"},
		},
		&cli.StringFlag{
			Name:    "env",
			Aliases: []string{"e"},
			Value:   "development",
			Usage:   "运行环境 (development|testing|production)",
			EnvVars: []string{"xledger_ENV"},
		},
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "是否开启调试模式",
			EnvVars: []string{"xledger_DEBUG"},
		},
	}
)
