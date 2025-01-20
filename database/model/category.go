package model

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

const (
	CategoryTypeIncome   = "income"
	CategoryTypeExpense  = "expense"
	CategoryTypeTransfer = "transfer"
)

const (
	CategoryIsSystem = true
	CategoryIsNotSystem = false
)

type Category struct {
	UUIDModel
	Name     string    `gorm:"column:name;size:50;not null;unique"`
	UserID   uuid.UUID `gorm:"column:user_id"`
	Type     string    `gorm:"column:type;size:50;not null"`
	IsSystem bool      `gorm:"column:is_system;not null;default:false"`
}

func (Category) TableName() string {
	return "category"
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Category) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

type SubCategory struct {
	UUIDModel
	CategoryID uuid.UUID `gorm:"column:category_id"`
	Name       string    `gorm:"column:name;size:50;not null"`
	UserID     uuid.UUID `gorm:"column:user_id"`
	IsSystem   bool      `gorm:"column:is_system;not null;default:false"`
}

func (SubCategory) TableName() string {
	return "subcategory"
}

func (s *SubCategory) BeforeCreate(tx *gorm.DB) error {
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return nil
}

func (s *SubCategory) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}
