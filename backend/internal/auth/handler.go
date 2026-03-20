package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *CodeService
}

type sendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type verifyCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

func NewHandler(service *CodeService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_code": "AUTH_BAD_REQUEST"})
		return
	}

	err := h.service.SendCode(c.Request.Context(), req.Email)
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

	tokens, err := h.service.VerifyCode(c.Request.Context(), req.Email, req.Code)
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
