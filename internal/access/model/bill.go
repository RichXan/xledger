package model

import (
	"time"
)

// BillType 账单类型
type BillType string

const (
	BillTypeIncome  BillType = "income"  // 收入
	BillTypeExpense BillType = "expense" // 支出
)

// Bill 账单记录
type Bill struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"`                 // 用户ID
	Amount      float64   `gorm:"type:decimal(10,2)" json:"amount"`     // 金额
	Type        BillType  `gorm:"type:varchar(10)" json:"type"`         // 类型：收入/支出
	CategoryID  uint      `gorm:"index" json:"category_id"`             // 分类ID
	Description string    `gorm:"type:varchar(255)" json:"description"` // 描述
	BillDate    time.Time `gorm:"index" json:"bill_date"`               // 账单日期
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BillCategory 账单分类
type BillCategory struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `gorm:"type:varchar(50)" json:"name"`  // 分类名称
	Type      BillType  `gorm:"type:varchar(10)" json:"type"`  // 适用类型：收入/支出
	Icon      string    `gorm:"type:varchar(100)" json:"icon"` // 图标
	Sort      int       `gorm:"default:0" json:"sort"`         // 排序
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BillTag 账单标签
type BillTag struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`          // 用户ID
	Name      string    `gorm:"type:varchar(50)" json:"name"`  // 标签名称
	Color     string    `gorm:"type:varchar(20)" json:"color"` // 标签颜色
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BillTagRelation 账单-标签关联表
type BillTagRelation struct {
	BillID uint `gorm:"primarykey" json:"bill_id"`
	TagID  uint `gorm:"primarykey" json:"tag_id"`
}
