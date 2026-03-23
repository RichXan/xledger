package auth

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/common/httpx"
)

type Handler struct {
	codeService    *CodeService
	oauthService   *OAuthService
	sessionService *SessionService
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

	tokens, err := h.oauthService.GoogleCallback(c.Request.Context(), GoogleCallbackInput{
		State: c.Query("state"),
		Nonce: c.Query("nonce"),
	})
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

	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) Me(c *gin.Context) {
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		authorization = strings.TrimSpace(authorization[len("Bearer "):])
	}
	token, err := ParseSessionToken(authorization)
	if err != nil || token.Type != "access" || !time.Now().UTC().Before(token.ExpiresAt) {
		httpx.JSON(c, http.StatusUnauthorized, "AUTH_REQUIRED", "未认证或凭证无效", nil)
		return
	}
	httpx.JSON(c, http.StatusOK, "OK", "成功", gin.H{"email": token.Email})
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
