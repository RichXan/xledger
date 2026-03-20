package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(trustedProxies []string) *gin.Engine {
	r := gin.New()
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		panic(err)
	}
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
