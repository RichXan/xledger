package router

import (
	"github.com/gin-gonic/gin"
)

func CategoryRouter(r *gin.RouterGroup) {
	category := r.Group("category")
	{
		category.POST("", categoryHandler.Create)
		category.DELETE("/:id", categoryHandler.Delete)
		category.PUT("/:id", categoryHandler.Update)
		category.GET("/:id", categoryHandler.Get)
		category.GET("", categoryHandler.List)
	}

	subCategory := r.Group("sub_category")
	{
		subCategory.POST("", subCategoryHandler.Create)
		subCategory.DELETE("/:id", subCategoryHandler.Delete)
		subCategory.PUT("/:id", subCategoryHandler.Update)
		subCategory.GET("/:id", subCategoryHandler.Get)
		subCategory.GET("", subCategoryHandler.List)
	}
}
