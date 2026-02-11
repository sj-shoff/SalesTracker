package items_postgres

import (
	"context"
	"database/sql"
	"fmt"

	"sales-tracker/internal/domain"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type ItemsPostgresRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewPostgresRepository(db *dbpg.DB, retries retry.Strategy) *ItemsPostgresRepository {
	return &ItemsPostgresRepository{
		db:      db,
		retries: retries,
	}
}

func (r *ItemsPostgresRepository) CreateItem(ctx context.Context, item *domain.Item) (int64, error) {
	var id int64
	query := `
		INSERT INTO items (type, amount, date, category, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, item.Type, item.Amount, item.Date, item.Category, item.Description)
	if err != nil {
		return 0, fmt.Errorf("failed to create item: %w", err)
	}
	err = row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to scan created item id: %w", err)
	}
	return id, nil
}

func (r *ItemsPostgresRepository) GetItems(ctx context.Context) ([]*domain.Item, error) {
	var items []*domain.Item
	query := `
		SELECT id, type, amount, date, category, description, created_at, updated_at
		FROM items
		ORDER BY date DESC
	`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
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
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating items: %w", err)
	}
	return items, nil
}

func (r *ItemsPostgresRepository) GetItemByID(ctx context.Context, id int64) (*domain.Item, error) {
	item := &domain.Item{}
	query := `
		SELECT id, type, amount, date, category, description, created_at, updated_at
		FROM items
		WHERE id = $1
	`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item by id: %w", err)
	}
	err = row.Scan(
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
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}
	return item, nil
}

func (r *ItemsPostgresRepository) UpdateItem(ctx context.Context, id int64, item *domain.Item) error {
	query := `
		UPDATE items
		SET type = $1, amount = $2, date = $3, category = $4, description = $5, updated_at = now()
		WHERE id = $6
	`
	res, err := r.db.ExecWithRetry(ctx, r.retries, query, item.Type, item.Amount, item.Date, item.Category, item.Description, id)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ItemsPostgresRepository) DeleteItem(ctx context.Context, id int64) error {
	query := `
		DELETE FROM items
		WHERE id = $1
	`
	res, err := r.db.ExecWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
