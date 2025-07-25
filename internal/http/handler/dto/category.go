package dto

import (
	"github.com/RichXan/xcommon/xhttp"
	"gorm.io/gorm"
)

type CategoryCreate struct {
	Name   string `json:"name" binding:"required"`
	Type   string `json:"type" binding:"required,oneof=income expense transfer"`
	UserID string `json:"user_id" binding:"omitempty"`
}

type CategoryUpdate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsSystem bool   `json:"is_system" validate:"omitempty"`
}

type CategoryList struct {
	xhttp.PageReq
	UserID string `json:"user_id" validate:"omitempty"`
}

func (dto *CategoryList) BuildQuery(db *gorm.DB) *gorm.DB {
	if dto.UserID != "" {
		db = db.Where("user_id = ?", dto.UserID)
	}
	return db
}
