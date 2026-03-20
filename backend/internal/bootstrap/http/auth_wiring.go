package http

import (
	"os"
	"time"

	"xledger/backend/internal/auth"
)

func newDefaultAuthHandlerFromEnv() *auth.Handler {
	repo := auth.NewInMemoryRepository(time.Now)
	sender := auth.NewSMTPMailSender(auth.SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("SMTP_FROM"),
	})
	service := auth.NewCodeService(repo, sender, nil, time.Now, nil)

	return auth.NewHandler(service)
}
