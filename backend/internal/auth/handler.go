package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error_code": "AUTH_BAD_REQUEST"})
		return
	}

	err := h.codeService.SendCode(c.Request.Context(), req.Email, c.ClientIP())
	if err != nil {
		switch ErrorCode(err) {
		case AUTH_CODE_RATE_LIMIT:
			c.JSON(http.StatusTooManyRequests, gin.H{"error_code": AUTH_CODE_RATE_LIMIT})
			return
		case AUTH_CODE_SEND_FAILED:
			c.JSON(http.StatusBadGateway, gin.H{"error_code": AUTH_CODE_SEND_FAILED})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"code_sent": true})
}

func (h *Handler) VerifyCode(c *gin.Context) {
	var req verifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": "AUTH_BAD_REQUEST"})
		return
	}

	tokens, err := h.codeService.VerifyCode(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		switch ErrorCode(err) {
		case AUTH_CODE_INVALID:
			c.JSON(http.StatusUnauthorized, gin.H{"error_code": AUTH_CODE_INVALID})
			return
		case AUTH_CODE_EXPIRED:
			c.JSON(http.StatusUnauthorized, gin.H{"error_code": AUTH_CODE_EXPIRED})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error_code": "AUTH_INTERNAL"})
			return
		}
	}

	c.JSON(http.StatusOK, tokens)
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
