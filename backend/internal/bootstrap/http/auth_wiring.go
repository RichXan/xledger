package http

import (
	"os"
	"strings"
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
	redirectURL := strings.TrimSpace(os.Getenv("GOOGLE_AUTH_REDIRECT_URL"))
	if redirectURL == "" {
		redirectURL = "http://127.0.0.1:8080/api/auth/google/callback"
	}
	oauthService.SetGoogleProvider(auth.NewGoogleOAuthProvider(auth.GoogleOAuthConfig{
		ClientID:     os.Getenv("GOOGLE_AUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_AUTH_CLIENT_SECRET"),
		RedirectURL:  redirectURL,
	}))

	handler := auth.NewHandlerWithServices(codeService, oauthService, sessionService)
	frontendReturn := strings.TrimSpace(os.Getenv("GOOGLE_AUTH_FRONTEND_RETURN"))
	if frontendReturn == "" {
		frontendReturn = "http://127.0.0.1:4173/auth/google/callback"
	}
	handler.SetGoogleFrontendReturnURL(frontendReturn)
	return handler
}
