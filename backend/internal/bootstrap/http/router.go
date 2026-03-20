package http

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"xledger/backend/internal/auth"
)

func NewRouter(trustedProxies []string) (*gin.Engine, error) {
	r := gin.New()
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	repo := auth.NewInMemoryRepository(time.Now)
	sender := auth.NewSMTPMailSender(auth.SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("SMTP_FROM"),
	})
	service := auth.NewCodeService(repo, sender, nil, time.Now, nil)
	handler := auth.NewHandler(service)

	authGroup := r.Group("/api/auth")
	authGroup.POST("/send-code", handler.SendCode)
	authGroup.POST("/verify-code", handler.VerifyCode)

	return r, nil
}
