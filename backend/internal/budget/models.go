package budget

import (
	"time"

	"github.com/google/uuid"
)

func generateID() string {
	return uuid.New().String()
}

type Budget struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	UserID     string    `json:"user_id" gorm:"index;not null"`
	CategoryID string    `json:"category_id" gorm:"index"`
	Amount     float64   `json:"amount" gorm:"not null"`
	Period     string    `json:"period" gorm:"not null"` // "monthly"
	AlertAt    float64   `json:"alert_at"`               // threshold percentage (0-100)
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type BudgetAlert struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"index;not null"`
	BudgetID     string    `json:"budget_id" gorm:"index;not null"`
	TriggeredAt  time.Time `json:"triggered_at"`
	AlertType    string    `json:"alert_type"` // "threshold", "over_budget"
	SpentAmount  float64   `json:"spent_amount"`
	BudgetAmount float64   `json:"budget_amount"`
	Message      string    `json:"message"`
}

type UserNotificationPref struct {
	UserID        string `json:"user_id" gorm:"primaryKey"`
	RealtimeAlert bool   `json:"realtime_alert" gorm:"default:true"`
	DailyDigest   bool   `json:"daily_digest" gorm:"default:false"`
	WeeklyDigest  bool   `json:"weekly_digest" gorm:"default:false"`
	PushEndpoint  string `json:"push_endpoint"`
	PushKey       string `json:"push_key"`
}

type BudgetWithUsage struct {
	Budget
	Spent     float64 `json:"spent"`
	Remaining float64 `json:"remaining"`
	Percent   float64 `json:"percent"`
}
