package dto

import (
	"time"
)

type AnalyticsResponse struct {
	Income  *ItemAnalytics          `json:"income"`
	Expense *ItemAnalytics          `json:"expense"`
	Details []AnalyticsItemResponse `json:"details"`
}

type ItemAnalytics struct {
	Sum       float64 `json:"sum"`
	Avg       float64 `json:"avg"`
	Count     int64   `json:"count"`
	Median    float64 `json:"median"`
	Percent90 float64 `json:"percent90"`
}

type AnalyticsItemResponse struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
