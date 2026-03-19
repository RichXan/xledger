package config

import (
	"errors"
	"os"
)

type Config struct {
	SMTPHost string
	APIAddr  string
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

	return Config{SMTPHost: smtpHost, APIAddr: apiAddr}, nil
}
