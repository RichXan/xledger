package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserStatus int8

const (
	UserStatusNormal  UserStatus = 1 // 正常状态
	UserStatusDisable UserStatus = 2 // 禁用
)

// User 用户模型
type User struct {
	UUIDModel
	Username string     `gorm:"column:username;size:255;not null" json:"username"`
	Password string     `gorm:"column:password;size:255;not null" json:"-"`
	Email    string     `gorm:"column:email;size:255;not null" json:"email"`
	Nickname string     `gorm:"column:nickname;size:255;not null" json:"nickname"`
	Gender   string     `gorm:"column:gender;size:50;not null" json:"gender"`
	Avatar   string     `gorm:"column:avatar;size:255;default:''" json:"avatar"`
	Status   UserStatus `gorm:"column:status;default:1;not null" json:"status"` // 1-正常，2-禁用
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
