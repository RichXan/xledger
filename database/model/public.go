package model

import (
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DefaultOrder = "created_at"

type Tmodel any

type Gorm func(db *gorm.DB) *gorm.DB

type UUIDModel struct {
	ID        string `gorm:"column:id;primarykey;type:varchar(255);comment:主键" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	db *gorm.DB `gorm:"-"`
}

func (my *UUIDModel) BeforeCreate(tx *gorm.DB) (err error) {
	if my.ID == "" {
		uuid, err := uuid.NewV4()
		if err != nil {
			return err
		}
		my.ID = strings.ReplaceAll(uuid.String(), "-", "")
	}

	return
}

func (m *UUIDModel) Gorm(fn ...Gorm) *gorm.DB {
	db := m.db
	for _, f := range fn {
		db = f(db)
	}
	return db
}

// Clauses
func (m *UUIDModel) Clauses() Gorm {
	return func(db *gorm.DB) *gorm.DB {
		return db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		})
	}
}

func (m *UUIDModel) Equal(key string, val any) Gorm {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(key+" = ?", m.ID)
	}
}

type PublicBy struct {
	CreatedBy string `gorm:"column:created_by" json:"created_by"`
	UpdatedBy string `gorm:"column:updated_by" json:"updated_by"`
	DeletedBy string `gorm:"column:deleted_by" json:"deleted_by"`
}
