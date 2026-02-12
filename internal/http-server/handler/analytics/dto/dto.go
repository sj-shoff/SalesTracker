package dto

import "sales-tracker/internal/domain"

type AnalyticsResponse struct {
	Sum       float64       `json:"sum"`
	Avg       float64       `json:"avg"`
	Count     int64         `json:"count"`
	Median    float64       `json:"median"`
	Percent90 float64       `json:"percent90"`
	Details   []domain.Item `json:"details"`
}
