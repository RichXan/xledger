package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	SMTPHost string
}

func Load() (Config, error) {
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		return Config{}, errors.New("missing required env var: SMTP_HOST")
	}

	return Config{SMTPHost: smtpHost}, nil
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	return cfg
}
