package auth

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/common/httpx"
)

type Handler struct {
	codeService             *CodeService
	oauthService            *OAuthService
	sessionService          *SessionService
	passwordService         *PasswordService
	googleFrontendReturnURL string
}

type sendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type verifyCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type registerRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name"`
}

type passwordLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type updateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required"`
}

func NewHandler(service *CodeService) *Handler {
	return &Handler{codeService: service}
}

func NewHandlerWithServices(codeService *CodeService, oauthService *OAuthService, sessionService *SessionService) *Handler {
	return &Handler{
		codeService:    codeService,
		oauthService:   oauthService,
		sessionService: sessionService,
	}
}

func (h *Handler) SetPasswordService(service *PasswordService) {
	if h == nil {
		return
	}
	h.passwordService = service
}

func (h *Handler) SetGoogleFrontendReturnURL(returnURL string) {
	h.googleFrontendReturnURL = strings.TrimSpace(returnURL)
}

func (h *Handler) HasOAuthService() bool {
	return h != nil && h.oauthService != nil
}

func (h *Handler) HasSessionService() bool {
	return h != nil && h.sessionService != nil
}

func (h *Handler) SendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}

	err := h.codeService.SendCode(c.Request.Context(), req.Email, c.ClientIP())
	if err != nil {
		switch ErrorCode(err) {
		case AUTH_CODE_RATE_LIMIT:
			httpx.JSON(c, http.StatusTooManyRequests, "RATE_LIMITED", "触发限流", nil)
			return
		case AUTH_CODE_SEND_FAILED:
			httpx.JSON(c, http.StatusBadGateway, "INTERNAL_ERROR", "服务内部错误", nil)
			return
		default:
			httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
			return
		}
	}

	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"code_sent": true})
}

func (h *Handler) VerifyCode(c *gin.Context) {
	var req verifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}

	tokens, err := h.codeService.VerifyCode(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		switch ErrorCode(err) {
		case AUTH_CODE_INVALID, AUTH_CODE_EXPIRED:
			httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
			return
		default:
			httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
			return
		}
	}

	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"access_token": tokens.AccessToken, "refresh_token": tokens.RefreshToken})
}

func (h *Handler) GoogleCallback(c *gin.Context) {
	if h.oauthService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
		return
	}

	code := strings.TrimSpace(c.Query("code"))
	if code != "" {
		tokens, err := h.oauthService.GoogleCallbackByCode(c.Request.Context(), c.Query("state"), code)
		if err != nil {
			log.Printf("google oauth callback by code failed: state=%q err=%v", c.Query("state"), err)
			if h.googleFrontendReturnURL != "" {
				target := h.googleFrontendReturnURL + "?error_code=" + url.QueryEscape(AUTH_OAUTH_FAILED)
				if errors.Is(err, ErrGoogleOAuthCodeInvalidOrExpired) {
					target += "&error_reason=" + url.QueryEscape("oauth_code_invalid_or_expired")
				}
				c.Redirect(http.StatusTemporaryRedirect, target)
				return
			}
			switch ErrorCode(err) {
			case AUTH_OAUTH_FAILED:
				c.JSON(http.StatusUnauthorized, gin.H{"error_code": AUTH_OAUTH_FAILED})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
				return
			}
		}
		if h.googleFrontendReturnURL != "" {
			target := h.googleFrontendReturnURL +
				"?access_token=" + url.QueryEscape(tokens.AccessToken) +
				"&refresh_token=" + url.QueryEscape(tokens.RefreshToken)
			c.Redirect(http.StatusTemporaryRedirect, target)
			return
		}
		c.JSON(http.StatusOK, tokens)
		return
	}

	tokens, err := h.oauthService.GoogleCallback(c.Request.Context(), GoogleCallbackInput{
		State: c.Query("state"),
		Nonce: c.Query("nonce"),
	})
	if err != nil {
		log.Printf("google oauth callback failed: state=%q nonce=%q err=%v", c.Query("state"), c.Query("nonce"), err)
		switch ErrorCode(err) {
		case AUTH_OAUTH_FAILED:
			c.JSON(http.StatusUnauthorized, gin.H{"error_code": AUTH_OAUTH_FAILED})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
			return
		}
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) GoogleStart(c *gin.Context) {
	if h.oauthService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
		return
	}
	loginURL, err := h.oauthService.GoogleLoginURL(c.Request.Context())
	if err != nil {
		switch ErrorCode(err) {
		case AUTH_OAUTH_FAILED:
			c.JSON(http.StatusUnauthorized, gin.H{"error_code": AUTH_OAUTH_FAILED})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
			return
		}
	}
	c.Redirect(http.StatusTemporaryRedirect, loginURL)
}

func (h *Handler) Me(c *gin.Context) {
	token, err := h.requireAccessToken(c)
	if err != nil || token.Type != "access" || !time.Now().UTC().Before(token.ExpiresAt) {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	displayName := ""
	if h.sessionService != nil {
		name, ok, nameErr := h.sessionService.GetUserDisplayName(c.Request.Context(), token.Email)
		if nameErr == nil && ok {
			displayName = strings.TrimSpace(name)
		}
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"email": token.Email, "name": displayName})
}

func (h *Handler) RegisterWithPassword(c *gin.Context) {
	if h.passwordService == nil || h.sessionService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}

	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}

	record, err := h.passwordService.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		switch ErrorCode(err) {
		case AUTH_USER_EXISTS:
			httpx.JSON(c, http.StatusConflict, "VALIDATION_ERROR", "用户已存在", nil)
			return
		case AUTH_PASSWORD_WEAK, AUTH_PROFILE_BAD_INPUT:
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		default:
			httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
			return
		}
	}

	tokens, issueErr := h.sessionService.IssueSessionWithProfile(c.Request.Context(), record.Email, record.DisplayName)
	if issueErr != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}

	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func (h *Handler) LoginWithPassword(c *gin.Context) {
	if h.passwordService == nil || h.sessionService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}

	var req passwordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}

	record, err := h.passwordService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch ErrorCode(err) {
		case AUTH_USER_NOT_FOUND, AUTH_PASSWORD_INVALID:
			httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
			return
		default:
			httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
			return
		}
	}

	tokens, issueErr := h.sessionService.IssueSessionWithProfile(c.Request.Context(), record.Email, record.DisplayName)
	if issueErr != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}

	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	if h.passwordService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}

	token, err := h.requireAccessToken(c)
	if err != nil || token.Type != "access" || !time.Now().UTC().Before(token.ExpiresAt) {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}

	var req changePasswordRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}

	if changeErr := h.passwordService.ChangePassword(c.Request.Context(), token.Email, req.OldPassword, req.NewPassword); changeErr != nil {
		switch ErrorCode(changeErr) {
		case AUTH_PASSWORD_INVALID:
			httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
			return
		case AUTH_PASSWORD_WEAK, AUTH_PROFILE_BAD_INPUT:
			httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
			return
		default:
			httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
			return
		}
	}

	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"changed": true})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	if h.passwordService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}

	token, err := h.requireAccessToken(c)
	if err != nil || token.Type != "access" || !time.Now().UTC().Before(token.ExpiresAt) {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}

	var req updateProfileRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		httpx.JSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "请求参数不合法", nil)
		return
	}

	if updateErr := h.passwordService.UpdateDisplayName(c.Request.Context(), token.Email, req.DisplayName); updateErr != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}

	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{
		"email": token.Email,
		"name":  strings.TrimSpace(req.DisplayName),
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	if h.sessionService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
		return
	}

	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": AUTH_BAD_REQUEST})
		return
	}

	tokens, err := h.sessionService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch ErrorCode(err) {
		case AUTH_BAD_REQUEST:
			c.JSON(http.StatusBadRequest, gin.H{"error_code": AUTH_BAD_REQUEST})
			return
		case AUTH_REFRESH_EXPIRED:
			c.JSON(http.StatusUnauthorized, gin.H{"error_code": AUTH_REFRESH_EXPIRED})
			return
		case AUTH_REFRESH_REVOKED:
			c.JSON(http.StatusUnauthorized, gin.H{"error_code": AUTH_REFRESH_REVOKED})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
			return
		}
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) DevLogin(c *gin.Context) {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("GIN_MODE")), "release") {
		httpx.JSON(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "资源不存在", nil)
		return
	}
	if h.sessionService == nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}

	email := strings.TrimSpace(c.Query("email"))
	if email == "" {
		email = "demo@example.com"
	}

	tokens, err := h.sessionService.IssueSession(c.Request.Context(), email)
	if err != nil {
		httpx.JSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
		return
	}

	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"email":         strings.TrimSpace(strings.ToLower(email)),
	})
}

func (h *Handler) Logout(c *gin.Context) {
	if h.sessionService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
		return
	}

	refreshToken := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(refreshToken), "bearer ") {
		refreshToken = strings.TrimSpace(refreshToken[len("Bearer "):])
	}
	if refreshToken == "" {
		var req logoutRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			refreshToken = strings.TrimSpace(req.RefreshToken)
		}
	}

	err := h.sessionService.Logout(c.Request.Context(), refreshToken)
	if err != nil {
		switch ErrorCode(err) {
		case AUTH_BAD_REQUEST:
			c.JSON(http.StatusBadRequest, gin.H{"error_code": AUTH_BAD_REQUEST})
			return
		case AUTH_UNAUTHORIZED:
			c.JSON(http.StatusUnauthorized, gin.H{"error_code": AUTH_UNAUTHORIZED})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"logged_out": true})
}

func (h *Handler) requireAccessToken(c *gin.Context) (SessionToken, error) {
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		authorization = strings.TrimSpace(authorization[len("Bearer "):])
	}
	return ParseSessionToken(authorization)
}
