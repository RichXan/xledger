package model

import "time"

type PublicTime struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}
type PublicBy struct {
	CreatedBy string `gorm:"column:created_by" json:"created_by"`
	UpdatedBy string `gorm:"column:updated_by" json:"updated_by"`
	DeletedBy string `gorm:"column:deleted_by" json:"deleted_by"`
}
