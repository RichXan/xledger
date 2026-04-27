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
	redirectURL := cfg.GoogleAuthRedirectURL
	if redirectURL == "" {
		redirectURL = "http://127.0.0.1:8080/api/auth/google/callback"
	}
	oauthService.SetGoogleProvider(auth.NewGoogleOAuthProvider(auth.GoogleOAuthConfig{
		ClientID:     cfg.GoogleAuthClientID,
		ClientSecret: cfg.GoogleAuthClientSecret,
		RedirectURL:  redirectURL,
	}))

	handler := auth.NewHandlerWithServices(codeService, oauthService, sessionService)
	handler.SetPasswordService(auth.NewPasswordService(repo))
	frontendReturn := cfg.GoogleAuthFrontendReturn
	if frontendReturn == "" {
		frontendReturn = "http://127.0.0.1:4173/auth/google/callback"
	}
	handler.SetGoogleFrontendReturnURL(frontendReturn)
	handler.SetDevLoginEnabled(cfg.EnableDevLogin)
	return handler, cfg
}
