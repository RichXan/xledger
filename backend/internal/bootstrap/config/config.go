package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	SMTPHost       string
	APIAddr        string
	TrustedProxies []string
}

func Load() (Config, error) {
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		return Config{}, errors.New("missing required env var: SMTP_HOST")
	}

	apiAddr := os.Getenv("API_ADDR")
	if apiAddr == "" {
		apiAddr = ":8080"
	}

	trustedProxies := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))

	return Config{
		SMTPHost:       smtpHost,
		APIAddr:        apiAddr,
		TrustedProxies: trustedProxies,
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
