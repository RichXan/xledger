package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserStatus int8

const (
	UserStatusDisable UserStatus = 0 // 禁用
	UserStatusNormal  UserStatus = 1 // 正常状态
)

// User 用户模型
type User struct {
	gorm.Model
	Username string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password string     `gorm:"size:100;not null" json:"-"`
	Email    string     `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Nickname string     `gorm:"size:50;not null" json:"nickname"`
	Avatar   string     `gorm:"size:255;default:''" json:"avatar"`
	Status   UserStatus `gorm:"default:1;not null" json:"status"` // 0-禁用，1-正常
}

func (User) TableName() string {
	return "user"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// ComparePassword 比较密码
func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
