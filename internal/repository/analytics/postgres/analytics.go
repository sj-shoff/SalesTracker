package analytics_postgres

import (
	"context"
	"fmt"
	"sales-tracker/internal/domain"
	customErr "sales-tracker/internal/domain/errors"
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
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	err = row.Scan(
		&analytics.Sum,
		&analytics.Avg,
		&analytics.Count,
		&analytics.Median,
		&analytics.Percent90,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}

	detailsQuery := `
		SELECT id, type, amount, date, category, description, created_at, updated_at
		FROM items
		WHERE date BETWEEN $1 AND $2
		ORDER BY date DESC
	`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, detailsQuery, from, to)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	defer rows.Close()

	for rows.Next() {
		item := &domain.Item{}
		err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Amount,
			&item.Date,
			&item.Category,
			&item.Description,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
		}
		analytics.Details = append(analytics.Details, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}

	return analytics, nil
}
