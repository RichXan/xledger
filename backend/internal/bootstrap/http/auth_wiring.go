package http

import (
	"os"
	"time"

	"xledger/backend/internal/auth"
	"xledger/backend/internal/classification"
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
	templateService := classification.NewTemplateService(classification.NewInMemoryTemplateRepository())
	sessionService := auth.NewSessionService(repo, &auth.SessionServiceOptions{
		PostLoginBootstrap: templateService.EnsureUserDefaults,
	}, time.Now)
	codeService := auth.NewCodeService(repo, sender, auth.NewSessionTokenIssuer(sessionService), time.Now, nil)
	oauthService := auth.NewOAuthService(repo, sessionService, time.Now)

	return auth.NewHandlerWithServices(codeService, oauthService, sessionService)
}
