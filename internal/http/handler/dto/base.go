package dto

import (
	"github.com/RichXan/xcommon/xhttp"
	"gorm.io/gorm"
)

// UpdateDto 定义更新接口
type UpdateDto struct {
	IDs []string `json:"ids" form:"ids" binding:"required"`
}

// DeletesDto 定义删除接口
type DeletesDto struct {
	IDs []string `json:"ids" form:"ids" binding:"required"`
}

// ListDto 定义列表查询接口
type ListDto interface {
	BuildQuery(db *gorm.DB) *gorm.DB
	GetPageReq() xhttp.PageReq
}
