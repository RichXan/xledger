// backend/internal/push/service.go
package push

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var vapIDPublicKey, vapIDPrivateKey string

func init() {
	// 生成 VAPID 密钥对（生产环境应该持久化）
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	vapIDPrivateKey = base64.URLEncoding.EncodeToString(privateKey.D.Bytes())
	vapIDPublicKey = base64.URLEncoding.EncodeToString(
		elliptic.MarshalCompressed(privateKey.PublicKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y),
	)
}

func GetVAPIDPublicKey() string {
	return vapIDPublicKey
}

// 存储订阅（内存-map，生产环境用 Redis）
var subscriptionStore = make(map[string]PushSubscription)

type PushSubscription struct {
	UserID    string `json:"user_id"`
	Endpoint  string `json:"endpoint"`
	P256dh    string `json:"keys.p256dh"`
	Auth      string `json:"keys.auth"`
	CreatedAt int64  `json:"created_at"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Subscribe(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "未认证"})
		return
	}

	var sub struct {
		Endpoint string `json:"endpoint" binding:"required"`
		Keys      struct {
			P256dh string `json:"p256dh" binding:"required"`
			Auth   string `json:"auth" binding:"required"`
		} `json:"keys" binding:"required"`
	}
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "请求格式错误"})
		return
	}

	key := fmt.Sprintf("%s:%s", userID, sub.Endpoint)
	subscriptionStore[key] = PushSubscription{
		UserID:    userID.(string),
		Endpoint:  sub.Endpoint,
		P256dh:    sub.Keys.P256dh,
		Auth:      sub.Keys.Auth,
		CreatedAt: time.Now().Unix(),
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "订阅成功"})
}

func (s *Service) Unsubscribe(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "未认证"})
		return
	}

	var sub struct {
		Endpoint string `json:"endpoint"`
	}
	c.ShouldBindJSON(&sub)

	key := fmt.Sprintf("%s:%s", userID, sub.Endpoint)
	delete(subscriptionStore, key)

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "已退订"})
}

func (s *Service) SendPushNotification(userID, title, body, tag string) error {
	// 查找用户的订阅
	for _, sub := range subscriptionStore {
		if sub.UserID != userID {
			continue
		}
		// 使用 Web Push 发送
		// 注意：生产环境应该使用 web-push 库
		_ = sub // 实际发送逻辑省略
	}
	return nil
}