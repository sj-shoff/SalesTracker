package domain

import "time"

type Item struct {
	ID          int64
	Type        string    `validate:"required,oneof=income expense"`
	Amount      float64   `validate:"gte=0"`
	Date        time.Time `validate:"required"`
	Category    string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
