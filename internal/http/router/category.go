package router

import (
	"github.com/gin-gonic/gin"
)

func CategoryRouter(r *gin.RouterGroup) {
	category := r.Group("categories")
	{
		category.POST("", categoryHandler.Create)
		category.DELETE("/:category_id", categoryHandler.Delete)
		category.PUT("/:category_id", categoryHandler.Update)
		category.GET("/:category_id", categoryHandler.Get)
		category.GET("", categoryHandler.List)
	}

	subCategory := r.Group("sub-categories")
	{
		subCategory.POST("", subCategoryHandler.Create)
		subCategory.DELETE("/:sub_category_id", subCategoryHandler.Delete)
		subCategory.PUT("/:sub_category_id", subCategoryHandler.Update)
		subCategory.GET("/:sub_category_id", subCategoryHandler.Get)
		subCategory.GET("", subCategoryHandler.List)
	}
}
