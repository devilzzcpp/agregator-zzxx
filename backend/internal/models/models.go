package models

import "time"

type Subscription struct {
	ID          string     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ServiceName string     `gorm:"column:service_name;type:text;not null" json:"service_name"`
	Price       int        `gorm:"column:price;not null" json:"price"`
	UserID      string     `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	StartDate   time.Time  `gorm:"column:start_date;type:date;not null" json:"start_date"`
	EndDate     *time.Time `gorm:"column:end_date;type:date" json:"end_date,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:timestamptz;not null;default:now()" json:"updated_at"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}
