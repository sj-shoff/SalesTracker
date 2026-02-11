package analytics_postgres

import (
	"context"
	"fmt"
	"sales-tracker/internal/domain"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type AnalyticsPostgresRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewAnalyticsPostgresRepository(db *dbpg.DB, retries retry.Strategy) *AnalyticsPostgresRepository {
	return &AnalyticsPostgresRepository{
		db:      db,
		retries: retries,
	}
}

func (r *AnalyticsPostgresRepository) GetAnalytics(ctx context.Context, from, to time.Time) (*domain.Analytics, error) {
	analytics := &domain.Analytics{}
	query := `
		SELECT 
			COALESCE(SUM(amount), 0) AS sum,
			COALESCE(AVG(amount), 0) AS avg,
			COUNT(*) AS count,
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount), 0) AS median,
			COALESCE(PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY amount), 0) AS percent90
		FROM items
		WHERE date BETWEEN $1 AND $2
	`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}
	err = row.Scan(
		&analytics.Sum,
		&analytics.Avg,
		&analytics.Count,
		&analytics.Median,
		&analytics.Percent90,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan analytics: %w", err)
	}
	return analytics, nil
}
