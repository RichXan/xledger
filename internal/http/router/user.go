package router

import (
	"github.com/gin-gonic/gin"
)

// setupUserRoutes 设置用户相关路由
func setupUserRoutes(r *gin.RouterGroup) {
	user := r.Group("user")
	{
		user.POST("", userHandler.Create)
		user.DELETE("/:id", userHandler.Delete)
		user.PUT("/:id", userHandler.Update)
		user.GET("/:id", userHandler.Get)
		user.GET("", userHandler.List)
		user.POST("/login", userHandler.Login)     // 登录
		user.POST("/refresh", userHandler.Refresh) // 刷新token
	}
}
