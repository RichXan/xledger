// backend/internal/push/handler.go
package push

import "github.com/gin-gonic/gin"

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Subscribe(c *gin.Context) {
	h.service.Subscribe(c)
}

func (h *Handler) Unsubscribe(c *gin.Context) {
	h.service.Unsubscribe(c)
}

func (h *Handler) GetVAPIDKey(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    "OK",
		"message": "成功",
		"data": gin.H{
			"publicKey": GetVAPIDPublicKey(),
		},
	})
}
