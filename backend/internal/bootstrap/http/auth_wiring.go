package http

import (
	"time"

	"xledger/backend/internal/auth"
	"xledger/backend/internal/bootstrap/config"
	"xledger/backend/internal/classification"
)

func newDefaultAuthHandlerWithConfig() (*auth.Handler, config.Config) {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	repo := auth.NewInMemoryRepository(time.Now)
	var sender auth.SMTPSender = auth.NewSMTPMailSender(auth.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUser,
		Password: cfg.SMTPPass,
		From:     cfg.SMTPFrom,
	})
	if isPlaceholderSMTP(cfg.SMTPHost, cfg.SMTPPass) {
		sender = auth.NewDevMailSender()
	}
	templateService := classification.NewTemplateService(classification.NewInMemoryTemplateRepository())
	sessionService := auth.NewSessionService(repo, &auth.SessionServiceOptions{
		PostLoginBootstrap: templateService.EnsureUserDefaults,
	}, time.Now)
	codeService := auth.NewCodeService(repo, sender, auth.NewSessionTokenIssuer(sessionService), time.Now, nil)
	oauthService := auth.NewOAuthService(repo, sessionService, time.Now)
	oauthService.SetGoogleProvider(auth.NewGoogleOAuthProvider(auth.GoogleOAuthConfig{
		ClientID:     cfg.GoogleAuthClientID,
		ClientSecret: cfg.GoogleAuthClientSecret,
		RedirectURL:  cfg.GoogleAuthRedirectURL,
	}))

	handler := auth.NewHandlerWithServices(codeService, oauthService, sessionService)
	handler.SetPasswordService(auth.NewPasswordService(repo))
	handler.SetGoogleFrontendReturnURL(cfg.GoogleAuthFrontendReturn)
	handler.SetDevLoginEnabled(cfg.EnableDevLogin)
	return handler, cfg
}
