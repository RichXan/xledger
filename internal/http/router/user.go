package router

import (
	"github.com/gin-gonic/gin"
)

// setupUserRoutes 设置用户相关路由
func setupUserRoutes(r *gin.RouterGroup) {
	public := r.Group("")
	{
		public.POST("/register", userHandler.HandleRegister)
		public.POST("/login", userHandler.HandleLogin)
	}
}
