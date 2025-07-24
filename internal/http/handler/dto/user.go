package dto

import (
	"github.com/RichXan/xcommon/xhttp"
	"gorm.io/gorm"
)

type UserCreate struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email" validate:"required,email"`
}

type UserUpdate struct {
	ID       string `json:"id"`
	Status   int    `json:"status" validate:"omitempty,oneof=1 2"`
	Nickname string `json:"nickname"`
	Gender   string `json:"gender"`
	Avatar   string `json:"avatar"`
}

type UserList struct {
	xhttp.PageReq
}

func (dto *UserList) BuildQuery(db *gorm.DB) *gorm.DB {
	return db
}

type UserRegister struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserChangePassword struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type UserRefreshToken struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
