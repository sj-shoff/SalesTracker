package items_postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sales-tracker/internal/domain"
	customErr "sales-tracker/internal/domain/errors"

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
		return 0, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	err = row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: no rows returned", customErr.ErrDatabase)
		}
		return 0, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
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
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	return items, nil
}

func (r *ItemsPostgresRepository) GetItemsWithPagination(ctx context.Context, offset, limit int) ([]*domain.Item, int64, error) {
	var items []*domain.Item
	var total int64

	countQuery := `SELECT COUNT(*) FROM items`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}

	err = row.Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}

	query := `
		SELECT id, type, amount, date, category, description, created_at, updated_at
		FROM items
		ORDER BY date DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
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
			return nil, 0, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}

	return items, total, nil
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
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrItemNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
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
		return fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	if rows == 0 {
		return customErr.ErrItemNotFound
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
		return fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	if rows == 0 {
		return customErr.ErrItemNotFound
	}
	return nil
}
