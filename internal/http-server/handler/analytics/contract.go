package analytics_handler

import (
	"context"
	"sales-tracker/internal/domain"
	"time"
)

type analyticsUsecase interface {
	GetAnalytics(ctx context.Context, from, to time.Time) (*domain.Analytics, error)
}
