package items_handler

import (
	"context"
	"sales-tracker/internal/domain"
)

type itemsUsecase interface {
	CreateItem(ctx context.Context, item *domain.Item) (int64, error)
	GetItems(ctx context.Context) ([]*domain.Item, error)
	GetItemsWithPagination(ctx context.Context, offset, limit int) ([]*domain.Item, int64, error)
	GetItemByID(ctx context.Context, id int64) (*domain.Item, error)
	UpdateItem(ctx context.Context, id int64, item *domain.Item) error
	DeleteItem(ctx context.Context, id int64) error
}
