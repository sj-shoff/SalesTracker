package dto

import "time"

type CreateItemRequest struct {
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
}

type UpdateItemRequest struct {
	Type        string    `json:"type,omitempty"`
	Amount      float64   `json:"amount,omitempty"`
	Date        time.Time `json:"date,omitempty"`
	Category    string    `json:"category,omitempty"`
	Description string    `json:"description,omitempty"`
}

type ItemResponse struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AnalyticsResponse struct {
	Sum       float64 `json:"sum"`
	Avg       float64 `json:"avg"`
	Count     int64   `json:"count"`
	Median    float64 `json:"median"`
	Percent90 float64 `json:"percent90"`
}
