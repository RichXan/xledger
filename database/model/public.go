package model

import (
	"fmt"
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
	ID        uuid.UUID      `gorm:"primarykey;type:uuid;comment:主键;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除

	db *gorm.DB `gorm:"-"`
}

func (m *UUIDModel) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID, err = uuid.NewV4()
		if err != nil {
			return fmt.Errorf("uuid create with ID failed, %w", err)
		}
	}

	// 将 uuid 转换为 32 位字符串
	uuidString := m.ID.String()
	uuidString = strings.ReplaceAll(uuidString, "-", "")
	m.ID, err = uuid.FromString(uuidString)
	if err != nil {
		return fmt.Errorf("uuid create with ID failed, %w", err)
	}

	return nil
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
