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
	codeService := auth.NewCodeService(repo, sender, nil, time.Now, nil)
	sessionService := auth.NewSessionService(repo, nil, time.Now)
	oauthService := auth.NewOAuthService(repo, sessionService, time.Now)

	return auth.NewHandlerWithServices(codeService, oauthService, sessionService)
}
