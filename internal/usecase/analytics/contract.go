package analytics_usecase

import (
	"context"
	"sales-tracker/internal/domain"
	"time"
)

type analyticsRepository interface {
	GetAnalytics(ctx context.Context, from, to time.Time) (*domain.Analytics, error)
}
