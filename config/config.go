package config

import (
	"github.com/RichXan/xcommon/xcache"
	"github.com/RichXan/xcommon/xdatabase"
	"github.com/RichXan/xcommon/xlog"
	"github.com/RichXan/xcommon/xoauth"
	"github.com/RichXan/xcommon/xutil"
)

// Configuration 配置结构
type Configuration struct {
	System     SystemConfig             `yaml:"system"`
	Log        xlog.LoggerConfig        `yaml:"log"`
	OAuth      xoauth.Config            `yaml:"oauth"`
	Minio      xdatabase.MinioConfig    `yaml:"minio"`
	SMTP       xutil.SMTPConfig         `yaml:"smtp"`
	Social     SocialConfig             `yaml:"social"`
	Server     ServerConfig             `yaml:"server"`
	MySQL      xdatabase.MySQLConfig    `yaml:"mysql"`
	Postgres   xdatabase.PostgresConfig `yaml:"postgres"`
	Redis      xcache.RedisConfig       `yaml:"redis"`
	Prometheus PrometheusConfig         `yaml:"prometheus"`
	Grafana    GrafanaConfig            `yaml:"grafana"`
	Jaeger     JaegerConfig             `yaml:"jaeger"`
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
