package domain

import "time"

type Item struct {
	ID          int64
	Type        string
	Amount      float64
	Date        time.Time
	Category    string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
