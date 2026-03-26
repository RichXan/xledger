package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrGoogleOAuthCodeInvalidOrExpired = errors.New("google oauth code invalid or expired")

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type GoogleOAuthProvider interface {
	IsEnabled() bool
	BuildAuthURL(state string) string
	ResolveProfile(ctx context.Context, code string) (GoogleOAuthProfile, error)
}

type GoogleOAuthProfile struct {
	Email string
	Name  string
}

type googleOAuthProvider struct {
	cfg    GoogleOAuthConfig
	client *http.Client
}

func NewGoogleOAuthProvider(cfg GoogleOAuthConfig) GoogleOAuthProvider {
	cfg.ClientID = strings.TrimSpace(cfg.ClientID)
	cfg.ClientSecret = strings.TrimSpace(cfg.ClientSecret)
	cfg.RedirectURL = strings.TrimSpace(cfg.RedirectURL)
	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURL == "" {
		return &googleOAuthProvider{cfg: cfg, client: &http.Client{Timeout: 8 * time.Second}}
	}
	return &googleOAuthProvider{cfg: cfg, client: &http.Client{Timeout: 8 * time.Second}}
}

func (p *googleOAuthProvider) IsEnabled() bool {
	return p != nil && p.cfg.ClientID != "" && p.cfg.ClientSecret != "" && p.cfg.RedirectURL != ""
}

func (p *googleOAuthProvider) BuildAuthURL(state string) string {
	q := url.Values{}
	q.Set("client_id", p.cfg.ClientID)
	q.Set("redirect_uri", p.cfg.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", strings.TrimSpace(state))
	q.Set("nonce", strings.TrimSpace(state))
	q.Set("prompt", "select_account")
	return "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode()
}

func (p *googleOAuthProvider) ResolveProfile(ctx context.Context, code string) (GoogleOAuthProfile, error) {
	if !p.IsEnabled() {
		return GoogleOAuthProfile{}, errors.New("google oauth not configured")
	}
	accessToken, err := p.exchangeCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return GoogleOAuthProfile{}, err
	}
	if accessToken == "" {
		return GoogleOAuthProfile{}, errors.New("missing access token from google")
	}
	return p.fetchProfile(ctx, accessToken)
}

func (p *googleOAuthProvider) exchangeCode(ctx context.Context, code string) (string, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", p.cfg.ClientID)
	form.Set("client_secret", p.cfg.ClientSecret)
	form.Set("redirect_uri", p.cfg.RedirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if resp.StatusCode/100 != 2 {
		if payload.Error == "" {
			payload.Error = resp.Status
		}
		if strings.EqualFold(strings.TrimSpace(payload.Error), "invalid_grant") {
			if strings.TrimSpace(payload.ErrorDescription) != "" {
				return "", fmt.Errorf("%w: %s", ErrGoogleOAuthCodeInvalidOrExpired, strings.TrimSpace(payload.ErrorDescription))
			}
			return "", ErrGoogleOAuthCodeInvalidOrExpired
		}
		if strings.TrimSpace(payload.ErrorDescription) != "" {
			return "", fmt.Errorf("google token exchange failed: %s (%s)", payload.Error, strings.TrimSpace(payload.ErrorDescription))
		}
		return "", fmt.Errorf("google token exchange failed: %s", payload.Error)
	}
	return strings.TrimSpace(payload.AccessToken), nil
}

func (p *googleOAuthProvider) fetchProfile(ctx context.Context, accessToken string) (GoogleOAuthProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return GoogleOAuthProfile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return GoogleOAuthProfile{}, err
	}
	defer resp.Body.Close()

	var payload struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return GoogleOAuthProfile{}, err
	}
	if resp.StatusCode/100 != 2 {
		return GoogleOAuthProfile{}, fmt.Errorf("google userinfo failed: %s", resp.Status)
	}
	if !payload.EmailVerified {
		return GoogleOAuthProfile{}, errors.New("google account email is not verified")
	}
	return GoogleOAuthProfile{
		Email: strings.TrimSpace(strings.ToLower(payload.Email)),
		Name:  strings.TrimSpace(payload.Name),
	}, nil
}
