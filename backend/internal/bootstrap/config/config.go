package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	SMTPHost                 string
	SMTPPort                 string
	SMTPUser                 string
	SMTPPass                 string
	SMTPFrom                 string
	AuthCodePepper           string
	APIAddr                  string
	TrustedProxies           []string
	DatabaseURL              string
	RedisURL                 string
	GoogleAuthClientID       string
	GoogleAuthClientSecret   string
	GoogleAuthRedirectURL    string
	GoogleAuthFrontendReturn string
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
	googleClientID := strings.TrimSpace(os.Getenv("GOOGLE_AUTH_CLIENT_ID"))
	googleClientSecret := strings.TrimSpace(os.Getenv("GOOGLE_AUTH_CLIENT_SECRET"))
	googleRedirectURL := strings.TrimSpace(os.Getenv("GOOGLE_AUTH_REDIRECT_URL"))
	if googleRedirectURL == "" {
		googleRedirectURL = "http://127.0.0.1:8080/api/auth/google/callback"
	}
	googleFrontendReturn := strings.TrimSpace(os.Getenv("GOOGLE_AUTH_FRONTEND_RETURN"))
	if googleFrontendReturn == "" {
		googleFrontendReturn = "http://127.0.0.1:4173/auth/google/callback"
	}

	return Config{
		SMTPHost:                 smtpHost,
		SMTPPort:                 strings.TrimSpace(os.Getenv("SMTP_PORT")),
		SMTPUser:                 strings.TrimSpace(os.Getenv("SMTP_USER")),
		SMTPPass:                 strings.TrimSpace(os.Getenv("SMTP_PASS")),
		SMTPFrom:                 strings.TrimSpace(os.Getenv("SMTP_FROM")),
		AuthCodePepper:           authCodePepper,
		APIAddr:                  apiAddr,
		TrustedProxies:           trustedProxies,
		DatabaseURL:              databaseURL,
		RedisURL:                 redisURL,
		GoogleAuthClientID:       googleClientID,
		GoogleAuthClientSecret:   googleClientSecret,
		GoogleAuthRedirectURL:    googleRedirectURL,
		GoogleAuthFrontendReturn: googleFrontendReturn,
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
