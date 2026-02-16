package dto

import "time"

type CreateItemRequest struct {
	Type        string  `json:"type" validate:"required,oneof=income expense"`
	Amount      float64 `json:"amount" validate:"gte=0"`
	Date        string  `json:"date" validate:"required"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
}

type UpdateItemRequest struct {
	Type        string   `json:"type,omitempty" validate:"omitempty,oneof=income expense"`
	Amount      *float64 `json:"amount,omitempty" validate:"omitempty,gte=0"`
	Date        string   `json:"date,omitempty"`
	Category    string   `json:"category,omitempty"`
	Description string   `json:"description,omitempty"`
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

type ItemsResponse struct {
	Items []*ItemResponse `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}
