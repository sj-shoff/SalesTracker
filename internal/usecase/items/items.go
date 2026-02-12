package items_usecase

import (
	"context"
	"errors"
	"fmt"
	"sales-tracker/internal/domain"
	customErr "sales-tracker/internal/domain/errors"

	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/zlog"
)

type Service struct {
	repo     itemsRepository
	logger   *zlog.Zerolog
	validate *validator.Validate
}

func NewService(repo itemsRepository, logger *zlog.Zerolog) *Service {
	return &Service{
		repo:     repo,
		logger:   logger,
		validate: validator.New(),
	}
}

func (s *Service) CreateItem(ctx context.Context, item *domain.Item) (int64, error) {
	if err := s.validate.Struct(item); err != nil {
		s.logger.Error().Err(err).Msg("Validation failed")
		return 0, fmt.Errorf("%w: %v", customErr.ErrInvalidInput, err)
	}
	s.logger.Info().Msg("Creating item")
	id, err := s.repo.CreateItem(ctx, item)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create item")
		if errors.Is(err, customErr.ErrDatabase) {
			return 0, customErr.ErrDatabase
		}
		return 0, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int64("id", id).Msg("Item created")
	return id, nil
}

func (s *Service) GetItems(ctx context.Context) ([]*domain.Item, error) {
	s.logger.Info().Msg("Getting items")
	items, err := s.repo.GetItems(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get items")
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, customErr.ErrDatabase
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int("count", len(items)).Msg("Items retrieved")
	return items, nil
}

func (s *Service) GetItemsWithPagination(ctx context.Context, offset, limit int) ([]*domain.Item, int64, error) {
	s.logger.Info().Int("offset", offset).Int("limit", limit).Msg("Getting items with pagination")
	items, total, err := s.repo.GetItemsWithPagination(ctx, offset, limit)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get items with pagination")
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, 0, customErr.ErrDatabase
		}
		return nil, 0, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int("count", len(items)).Int64("total", total).Msg("Items retrieved with pagination")
	return items, total, nil
}

func (s *Service) GetItemByID(ctx context.Context, id int64) (*domain.Item, error) {
	if id <= 0 {
		return nil, customErr.ErrInvalidInput
	}

	s.logger.Info().Int64("id", id).Msg("Getting item")
	item, err := s.repo.GetItemByID(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("Failed to get item")
		if errors.Is(err, customErr.ErrItemNotFound) {
			return nil, customErr.ErrItemNotFound
		}
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, customErr.ErrDatabase
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int64("id", id).Msg("Item retrieved")
	return item, nil
}

func (s *Service) UpdateItem(ctx context.Context, id int64, item *domain.Item) error {
	if id <= 0 {
		return customErr.ErrInvalidInput
	}

	if err := s.validate.Struct(item); err != nil {
		s.logger.Error().Err(err).Msg("Validation failed")
		return fmt.Errorf("%w: %v", customErr.ErrInvalidInput, err)
	}
	s.logger.Info().Int64("id", id).Msg("Updating item")
	err := s.repo.UpdateItem(ctx, id, item)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("Failed to update item")
		if errors.Is(err, customErr.ErrItemNotFound) {
			return customErr.ErrItemNotFound
		}
		if errors.Is(err, customErr.ErrDatabase) {
			return customErr.ErrDatabase
		}
		return fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int64("id", id).Msg("Item updated")
	return nil
}

func (s *Service) DeleteItem(ctx context.Context, id int64) error {
	if id <= 0 {
		return customErr.ErrInvalidInput
	}

	s.logger.Info().Int64("id", id).Msg("Deleting item")
	err := s.repo.DeleteItem(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", id).Msg("Failed to delete item")
		if errors.Is(err, customErr.ErrItemNotFound) {
			return customErr.ErrItemNotFound
		}
		if errors.Is(err, customErr.ErrDatabase) {
			return customErr.ErrDatabase
		}
		return fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int64("id", id).Msg("Item deleted")
	return nil
}
