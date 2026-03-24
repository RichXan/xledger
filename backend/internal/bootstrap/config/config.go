package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	SMTPHost       string
	AuthCodePepper string
	APIAddr        string
	TrustedProxies []string
	DatabaseURL    string
	RedisURL       string
}

func Load() (Config, error) {
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		return Config{}, errors.New("missing required env var: SMTP_HOST")
	}

	authCodePepper := strings.TrimSpace(os.Getenv("AUTH_CODE_PEPPER"))
	if authCodePepper == "" {
		return Config{}, errors.New("missing required env var: AUTH_CODE_PEPPER")
	}

	apiAddr := os.Getenv("API_ADDR")
	if apiAddr == "" {
		apiAddr = ":8080"
	}

	trustedProxies := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))

	return Config{
		SMTPHost:       smtpHost,
		AuthCodePepper: authCodePepper,
		APIAddr:        apiAddr,
		TrustedProxies: trustedProxies,
		DatabaseURL:    databaseURL,
		RedisURL:       redisURL,
	}, nil
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
