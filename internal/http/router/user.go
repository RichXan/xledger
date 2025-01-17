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
	}
	// 登录
	user.POST("/login", userHandler.Login)
	// 刷新token
	// user.POST("/refresh", userHandler.Refresh)
}
