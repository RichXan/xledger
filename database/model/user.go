package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserStatus int8

const (
	UserStatusNormal  UserStatus = 0 // 正常状态
	UserStatusDisable UserStatus = 1 // 禁用
)

// User 用户模型
type User struct {
	UUIDModel
	Username string     `gorm:"size:255;not null" json:"username"`
	Password string     `gorm:"size:255;not null" json:"-"`
	Email    string     `gorm:"size:255;not null" json:"email"`
	Nickname string     `gorm:"size:255;not null" json:"nickname"`
	Gender   string     `gorm:"size:50;not null" json:"gender"`
	Avatar   string     `gorm:"size:255;default:''" json:"avatar"`
	Status   UserStatus `gorm:"default:0;not null" json:"status"` // 0-正常，1-禁用
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
