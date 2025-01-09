package global

import (
	"fmt"
	"os"

	"github.com/RichXan/xcommon/xcache"
	"github.com/RichXan/xcommon/xdatabase"
	"github.com/RichXan/xcommon/xlog"
	"github.com/RichXan/xcommon/xutil"
	"gopkg.in/yaml.v3"
)

var Config *Configuration

// Configuration 配置结构
type Configuration struct {
	System     SystemConfig          `yaml:"system"`
	Log        xlog.LoggerConfig     `yaml:"log"`
	Minio      xdatabase.MinioConfig `yaml:"minio"`
	SMTP       xutil.SMTPConfig      `yaml:"smtp"`
	Social     SocialConfig          `yaml:"social"`
	Server     ServerConfig          `yaml:"server"`
	MySQL      xdatabase.MySQLConfig `yaml:"mysql"`
	Redis      xcache.RedisConfig    `yaml:"redis"`
	Prometheus PrometheusConfig      `yaml:"prometheus"`
	Grafana    GrafanaConfig         `yaml:"grafana"`
	Jaeger     JaegerConfig          `yaml:"jaeger"`
}

type SystemConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Env         string `yaml:"env"`
	Port        int    `yaml:"port"`
	Debug       bool   `yaml:"debug"`
	HTTP        struct {
		ReadTimeout  int `yaml:"read_timeout"`
		WriteTimeout int `yaml:"write_timeout"`
		IdleTimeout  int `yaml:"idle_timeout"`
	} `yaml:"http"`
}

type PrometheusConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	MetricsPath    string `yaml:"metrics_path"`
	ScrapeInterval string `yaml:"scrape_interval"`
}

type GrafanaConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type JaegerConfig struct {
	Host         string  `yaml:"host"`
	Port         int     `yaml:"port"`
	SamplerType  string  `yaml:"sampler_type"`
	SamplerParam float64 `yaml:"sampler_param"`
	LogSpans     bool    `yaml:"log_spans"`
}

type ServerConfig struct {
	HTTPPort int `yaml:"http_port"`
	GRPCPort int `yaml:"grpc_port"`
}

type SMTPConfig struct {
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	From     string   `yaml:"from"`
	User     string   `yaml:"user"`
	Password string   `yaml:"password"`
	ToEmails []string `yaml:"to_emails"`
}

type SocialConfig struct {
	OAuth OAuthConfig `yaml:"oauth"`
}

type OAuthConfig struct {
	CallbackBaseURL string         `yaml:"callback_base_url"`
	StateExpiry     string         `yaml:"state_expiry"`
	AutoCreateUser  bool           `yaml:"auto_create_user"`
	DefaultRole     string         `yaml:"default_role"`
	Security        SecurityConfig `yaml:"security"`
	Providers       struct {
		Github struct {
			ClientID     string   `yaml:"client_id"`
			ClientSecret string   `yaml:"client_secret"`
			Scopes       []string `yaml:"scopes"`
			Enabled      bool     `yaml:"enabled"`
		} `yaml:"github"`
		Google struct {
			ClientID     string   `yaml:"client_id"`
			ClientSecret string   `yaml:"client_secret"`
			Scopes       []string `yaml:"scopes"`
			Enabled      bool     `yaml:"enabled"`
			Extra        struct {
				QRCodeSize int    `yaml:"qrcode_size"`
				Lang       string `yaml:"lang"`
			} `yaml:"extra"`
		} `yaml:"google"`
		Wechat struct {
			ClientID     string   `yaml:"client_id"`
			ClientSecret string   `yaml:"client_secret"`
			Scopes       []string `yaml:"scopes"`
			Enabled      bool     `yaml:"enabled"`
			Extra        struct {
				QRCodeSize int    `yaml:"qrcode_size"`
				Lang       string `yaml:"lang"`
			} `yaml:"extra"`
		} `yaml:"wechat"`
		QQ struct {
			ClientID     string   `yaml:"client_id"`
			ClientSecret string   `yaml:"client_secret"`
			Scopes       []string `yaml:"scopes"`
			Enabled      bool     `yaml:"enabled"`
			Extra        struct {
				Display string `yaml:"display"`
			} `yaml:"extra"`
		} `yaml:"qq"`
		Weibo struct {
			ClientID     string   `yaml:"client_id"`
			ClientSecret string   `yaml:"client_secret"`
			Scopes       []string `yaml:"scopes"`
			Enabled      bool     `yaml:"enabled"`
			Extra        struct {
				Display string `yaml:"display"`
			} `yaml:"extra"`
		} `yaml:"weibo"`
	} `yaml:"providers"`
}

type SecurityConfig struct {
	MaxBindings         int    `yaml:"max_bindings"`
	AllowUnbindLast     bool   `yaml:"allow_unbind_last"`
	AllowMerge          bool   `yaml:"allow_merge"`
	MergeConfirmTimeout string `yaml:"merge_confirm_timeout"`
	EnableIPLimit       bool   `yaml:"enable_ip_limit"`
	IPLimit             struct {
		Window      string   `yaml:"window"`
		MaxRequests int      `yaml:"max_requests"`
		BanDuration string   `yaml:"ban_duration"`
		Whitelist   []string `yaml:"whitelist"`
	} `yaml:"ip_limit"`
}

// LoadConfig 加载配置文件
func LoadConfig(file string) error {
	// 读取配置文件
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read config file error: %v", err)
	}

	// 解析配置
	Config = &Configuration{}
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
