package router

import (
	"github.com/gin-gonic/gin"
)

// UserRoutes 设置用户相关路由
func UserRouter(r *gin.RouterGroup) {
	user := r.Group("users")
	{
		user.POST("", userHandler.Create)
		user.DELETE("/:user_id", userHandler.Delete)
		user.PUT("/:user_id", userHandler.Update)
		user.GET("/:user_id", userHandler.Get)
		user.GET("", userHandler.List)
	}
}
