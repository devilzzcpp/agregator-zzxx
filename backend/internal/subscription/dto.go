package subscription

import "time"

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" binding:"required"`
	Price       int     `json:"price" binding:"required"`
	UserID      string  `json:"user_id" binding:"required,uuid"`
	StartDate   string  `json:"start_date" binding:"required"`
	EndDate     *string `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string `json:"service_name,omitempty"`
	Price       *int    `json:"price,omitempty"`
	UserID      *string `json:"user_id,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}

type ListSubscriptionsQuery struct {
	UserID      string `form:"user_id"`
	ServiceName string `form:"service_name"`
	FromMonth   string `form:"from"`
	ToMonth     string `form:"to"`
	Limit       int    `form:"limit"`
	Offset      int    `form:"offset"`
}

type TotalSubscriptionsQuery struct {
	UserID      string    `form:"user_id"`
	ServiceName string    `form:"service_name"`
	From        time.Time `form:"-"`
	To          time.Time `form:"-"`
}

type SubscriptionResponse struct {
	ID          string  `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type SubscriptionsTotalResponse struct {
	TotalPrice int    `json:"total_price"`
	Currency   string `json:"currency"`
	FromMonth  string `json:"from"`
	ToMonth    string `json:"to"`
	Months     int    `json:"months"`
}
