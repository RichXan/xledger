package global

import (
	"fmt"
	"os"
	"xledger/config"

	"gopkg.in/yaml.v3"
)

var Config *config.Configuration

// LoadConfig 加载配置文件
func LoadConfig(file string) error {
	// 读取配置文件
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read config file error: %v", err)
	}

	// 解析配置
	Config = &config.Configuration{}
	if err := yaml.Unmarshal(data, Config); err != nil {
		return fmt.Errorf("unmarshal config error: %v", err)
	}

	// 验证配置
	if err := validateConfig(); err != nil {
		return fmt.Errorf("validate config error: %v", err)
	}

	return nil
}

// validateConfig 验证配置
func validateConfig() error {
	// 验证系统配置
	if Config.System.Name == "" {
		return fmt.Errorf("system name is required")
	}
	if Config.System.Port <= 0 {
		return fmt.Errorf("invalid system port")
	}

	// 验证日志配置
	if Config.Log.Level == "" {
		Config.Log.Level = "info" // 设置默认值
	}
	if Config.Log.Directory == "" {
		Config.Log.Directory = "logs" // 设置默认值
	}
	if Config.Log.MaxSize <= 0 {
		Config.Log.MaxSize = 100 // 默认100MB
	}
	if Config.Log.MaxBackups <= 0 {
		Config.Log.MaxBackups = 10 // 默认保留10个备份
	}

	// 验证服务器配置
	if Config.Server.HTTPPort <= 0 {
		return fmt.Errorf("invalid HTTP port")
	}
	if Config.Server.GRPCPort <= 0 {
		return fmt.Errorf("invalid gRPC port")
	}

	// 验证MySQL配置
	if Config.MySQL.Path == "" {
		Config.MySQL.Path = "localhost:3306" // 设置默认值
	}
	if Config.MySQL.Username == "" {
		return fmt.Errorf("MySQL username is required")
	}
	if Config.MySQL.Password == "" {
		return fmt.Errorf("MySQL password is required")
	}
	if Config.MySQL.Database == "" {
		return fmt.Errorf("MySQL database is required")
	}
	if Config.MySQL.MaxIdleConns <= 0 {
		Config.MySQL.MaxIdleConns = 10 // 设置默认值
	}
	if Config.MySQL.MaxOpenConns <= 0 {
		Config.MySQL.MaxOpenConns = 100 // 设置默认值
	}
	if Config.MySQL.ConnMaxLifetime <= 0 {
		Config.MySQL.ConnMaxLifetime = 300 // 设置默认值
	}

	// 验证Jaeger配置
	if Config.Jaeger.Host == "" {
		Config.Jaeger.Host = "localhost" // 设置默认值
	}
	if Config.Jaeger.Port <= 0 {
		Config.Jaeger.Port = 6831 // 设置默认值
	}
	if Config.Jaeger.SamplerType == "" {
		Config.Jaeger.SamplerType = "const" // 设置默认值
	}
	if Config.Jaeger.SamplerParam == 0 {
		Config.Jaeger.SamplerParam = 1 // 设置默认值
	}

	return nil
}
