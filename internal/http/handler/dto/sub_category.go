package dto

import (
	"github.com/RichXan/xcommon/xhttp"
	"gorm.io/gorm"
)

type SubCategoryCreate struct {
	Name string `json:"name" binding:"required"`
}

type SubCategoryUpdate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
	Status   int    `json:"status" validate:"omitempty,oneof=1 2"`
}

type SubCategoryList struct {
	xhttp.PageReq
}

func (dto *SubCategoryList) BuildQuery(db *gorm.DB) *gorm.DB {
	return db
}
