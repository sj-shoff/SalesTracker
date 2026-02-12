package analytics_usecase

import (
	"context"
	"errors"
	"fmt"
	"sales-tracker/internal/domain"
	customErr "sales-tracker/internal/domain/errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/zlog"
)

type Service struct {
	repo     analyticsRepository
	logger   *zlog.Zerolog
	validate *validator.Validate
}

func NewService(repo analyticsRepository, logger *zlog.Zerolog) *Service {
	return &Service{
		repo:     repo,
		logger:   logger,
		validate: validator.New(),
	}
}

func (s *Service) GetAnalytics(ctx context.Context, from, to time.Time) (*domain.Analytics, error) {
	if from.IsZero() || to.IsZero() {
		return nil, customErr.ErrMissingParameter
	}

	if from.After(to) {
		return nil, customErr.ErrInvalidDateRange
	}

	maxPeriod := 365 * 24 * time.Hour
	if to.Sub(from) > maxPeriod {
		return nil, customErr.ErrPeriodTooLarge
	}

	s.logger.Info().Time("from", from).Time("to", to).Msg("Getting analytics")
	anal, err := s.repo.GetAnalytics(ctx, from, to)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get analytics")
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, customErr.ErrDatabase
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Msg("Analytics retrieved")
	return anal, nil
}
