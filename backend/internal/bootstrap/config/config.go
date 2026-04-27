package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// yamlConfig mirrors the structure of config/config.yaml.
type yamlConfig struct {
	SMTP struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
		From string `yaml:"from"`
	} `yaml:"smtp"`

	Auth struct {
		CodePepper  string `yaml:"code_pepper"`
		TokenSecret string `yaml:"token_secret"`
	} `yaml:"auth"`

	DatabaseURL string `yaml:"database_url"`
	RedisURL    string `yaml:"redis_url"`
	APIAddr     string `yaml:"api_addr"`
	GinMode     string `yaml:"gin_mode"`

	EnableDevLogin bool `yaml:"enable_dev_login"`

	// Comma-separated list of trusted proxy IPs.
	TrustedProxies string `yaml:"trusted_proxies"`

	GoogleAuth struct {
		ClientID       string `yaml:"client_id"`
		ClientSecret   string `yaml:"client_secret"`
		RedirectURL    string `yaml:"redirect_url"`
		FrontendReturn string `yaml:"frontend_return"`
	} `yaml:"google_auth"`
}

// Config is the application-wide configuration passed to all subsystems.
type Config struct {
	SMTPHost                 string
	SMTPPort                 string
	SMTPUser                 string
	SMTPPass                 string
	SMTPFrom                 string
	AuthCodePepper           string
	AuthTokenSecret          string
	APIAddr                  string
	GinMode                  string
	EnableDevLogin           bool
	TrustedProxies           []string
	DatabaseURL              string
	RedisURL                 string
	GoogleAuthClientID       string
	GoogleAuthClientSecret   string
	GoogleAuthRedirectURL    string
	GoogleAuthFrontendReturn string
}

// Load reads configuration from a YAML file.
//
// The file path is resolved in the following order:
//  1. The CONFIG_FILE environment variable, if set.
//  2. config/config.yaml relative to the current working directory.
func Load() (Config, error) {
	path := strings.TrimSpace(os.Getenv("CONFIG_FILE"))
	explicitPath := path != ""
	if path == "" {
		path = "config/config.yaml"
	}

	var raw yamlConfig
	data, err := os.ReadFile(path)
	if err != nil {
		if explicitPath || !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("cannot read config file %q: %w", path, err)
		}
		raw = yamlConfigFromEnv()
	} else if err := yaml.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("cannot parse config file %q: %w", path, err)
	}

	// Validate required fields.
	if strings.TrimSpace(raw.Auth.CodePepper) == "" {
		return Config{}, errors.New("config: AUTH_CODE_PEPPER (auth.code_pepper) is required")
	}
	if strings.TrimSpace(raw.Auth.TokenSecret) == "" {
		return Config{}, errors.New("config: AUTH_TOKEN_SECRET (auth.token_secret) is required")
	}

	apiAddr := strings.TrimSpace(raw.APIAddr)
	if apiAddr == "" {
		apiAddr = ":8080"
	}

	googleRedirectURL := strings.TrimSpace(raw.GoogleAuth.RedirectURL)
	if googleRedirectURL == "" {
		googleRedirectURL = "http://127.0.0.1:8080/api/auth/google/callback"
	}

	googleFrontendReturn := strings.TrimSpace(raw.GoogleAuth.FrontendReturn)
	if googleFrontendReturn == "" {
		googleFrontendReturn = "http://127.0.0.1:4173/auth/google/callback"
	}

	os.Setenv("AUTH_CODE_PEPPER", strings.TrimSpace(raw.Auth.CodePepper))
	os.Setenv("AUTH_TOKEN_SECRET", strings.TrimSpace(raw.Auth.TokenSecret))

	return Config{
		SMTPHost:                 strings.TrimSpace(raw.SMTP.Host),
		SMTPPort:                 strings.TrimSpace(raw.SMTP.Port),
		SMTPUser:                 strings.TrimSpace(raw.SMTP.User),
		SMTPPass:                 strings.TrimSpace(raw.SMTP.Pass),
		SMTPFrom:                 strings.TrimSpace(raw.SMTP.From),
		AuthCodePepper:           strings.TrimSpace(raw.Auth.CodePepper),
		AuthTokenSecret:          strings.TrimSpace(raw.Auth.TokenSecret),
		APIAddr:                  apiAddr,
		GinMode:                  strings.TrimSpace(raw.GinMode),
		EnableDevLogin:           raw.EnableDevLogin,
		TrustedProxies:           parseTrustedProxies(raw.TrustedProxies),
		DatabaseURL:              strings.TrimSpace(raw.DatabaseURL),
		RedisURL:                 strings.TrimSpace(raw.RedisURL),
		GoogleAuthClientID:       strings.TrimSpace(raw.GoogleAuth.ClientID),
		GoogleAuthClientSecret:   strings.TrimSpace(raw.GoogleAuth.ClientSecret),
		GoogleAuthRedirectURL:    googleRedirectURL,
		GoogleAuthFrontendReturn: googleFrontendReturn,
	}, nil
}

func yamlConfigFromEnv() yamlConfig {
	var raw yamlConfig
	raw.SMTP.Host = os.Getenv("SMTP_HOST")
	raw.SMTP.Port = os.Getenv("SMTP_PORT")
	raw.SMTP.User = os.Getenv("SMTP_USER")
	raw.SMTP.Pass = os.Getenv("SMTP_PASS")
	raw.SMTP.From = os.Getenv("SMTP_FROM")
	raw.Auth.CodePepper = os.Getenv("AUTH_CODE_PEPPER")
	raw.Auth.TokenSecret = os.Getenv("AUTH_TOKEN_SECRET")
	raw.DatabaseURL = os.Getenv("DATABASE_URL")
	raw.RedisURL = os.Getenv("REDIS_URL")
	raw.APIAddr = os.Getenv("API_ADDR")
	raw.GinMode = os.Getenv("GIN_MODE")
	raw.EnableDevLogin = strings.EqualFold(os.Getenv("ENABLE_DEV_LOGIN"), "true") || os.Getenv("ENABLE_DEV_LOGIN") == "1"
	raw.TrustedProxies = os.Getenv("TRUSTED_PROXIES")
	raw.GoogleAuth.ClientID = os.Getenv("GOOGLE_AUTH_CLIENT_ID")
	raw.GoogleAuth.ClientSecret = os.Getenv("GOOGLE_AUTH_CLIENT_SECRET")
	raw.GoogleAuth.RedirectURL = os.Getenv("GOOGLE_AUTH_REDIRECT_URL")
	raw.GoogleAuth.FrontendReturn = os.Getenv("GOOGLE_AUTH_FRONTEND_RETURN")
	return raw
}

func parseTrustedProxies(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{"127.0.0.1", "::1"}
	}

	parts := strings.Split(value, ",")
	proxies := make([]string, 0, len(parts))
	for _, part := range parts {
		proxy := strings.TrimSpace(part)
		if proxy == "" {
			continue
		}
		proxies = append(proxies, proxy)
	}

	if len(proxies) == 0 {
		return []string{"127.0.0.1", "::1"}
	}

	return proxies
}
