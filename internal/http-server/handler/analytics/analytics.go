package analytics_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	customErr "sales-tracker/internal/domain/errors"
	"sales-tracker/internal/http-server/handler/analytics/dto"

	"github.com/wb-go/wbf/zlog"
)

type AnalyticsHandler struct {
	analyticsUsecase analyticsUsecase
	logger           *zlog.Zerolog
}

func NewHandler(analyticsUsecase analyticsUsecase, logger *zlog.Zerolog) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsUsecase: analyticsUsecase,
		logger:           logger,
	}
}

func (h *AnalyticsHandler) writeError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := map[string]string{"error": "internal_error"}
	switch {
	case errors.Is(err, customErr.ErrInvalidInput),
		errors.Is(err, customErr.ErrInvalidDateRange),
		errors.Is(err, customErr.ErrMissingParameter),
		errors.Is(err, customErr.ErrUnsupportedFormat):
		resp["error"] = "invalid_input"
		if statusCode == 0 {
			statusCode = http.StatusBadRequest
		}
	case errors.Is(err, customErr.ErrItemNotFound):
		resp["error"] = "not_found"
		if statusCode == 0 {
			statusCode = http.StatusNotFound
		}
	case errors.Is(err, customErr.ErrDatabase):
		resp["error"] = "database_error"
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}
	default:
		resp["error"] = err.Error()
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		h.logger.Warn().Str("from", fromStr).Str("to", toStr).Msg("Missing required parameters")
		h.writeError(w, customErr.ErrMissingParameter, http.StatusBadRequest)
		return
	}
	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		h.logger.Warn().Err(err).Str("from", fromStr).Msg("Invalid from date format")
		h.writeError(w, customErr.ErrUnsupportedFormat, http.StatusBadRequest)
		return
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		h.logger.Warn().Err(err).Str("to", toStr).Msg("Invalid to date format")
		h.writeError(w, customErr.ErrUnsupportedFormat, http.StatusBadRequest)
		return
	}
	if from.After(to) {
		h.logger.Warn().Time("from", from).Time("to", to).Msg("Invalid date range: from > to")
		h.writeError(w, customErr.ErrInvalidDateRange, http.StatusBadRequest)
		return
	}
	maxPeriod := 365 * 24 * time.Hour
	if to.Sub(from) > maxPeriod {
		h.logger.Warn().Dur("period", to.Sub(from)).Msg("Date range exceeds maximum allowed period")
		h.writeError(w, customErr.ErrPeriodTooLarge, http.StatusBadRequest)
		return
	}
	h.logger.Info().
		Str("from", fromStr).
		Str("to", toStr).
		Time("from_parsed", from).
		Time("to_parsed", to).
		Msg("Analytics request received")
	an, err := h.analyticsUsecase.GetAnalytics(r.Context(), from, to)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get analytics")
		if errors.Is(err, customErr.ErrInvalidDateRange) ||
			errors.Is(err, customErr.ErrMissingParameter) ||
			errors.Is(err, customErr.ErrUnsupportedFormat) {
			h.writeError(w, err, http.StatusBadRequest)
		} else {
			h.writeError(w, err, http.StatusInternalServerError)
		}
		return
	}

	details := make([]dto.AnalyticsItemResponse, len(an.Details))
	for i, item := range an.Details {
		details[i] = dto.AnalyticsItemResponse{
			ID:          item.ID,
			Type:        item.Type,
			Amount:      item.Amount,
			Date:        item.Date,
			Category:    item.Category,
			Description: item.Description,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	resp := dto.AnalyticsResponse{
		Income: &dto.ItemAnalytics{
			Sum:       an.Income.Sum,
			Avg:       an.Income.Avg,
			Count:     an.Income.Count,
			Median:    an.Income.Median,
			Percent90: an.Income.Percent90,
		},
		Expense: &dto.ItemAnalytics{
			Sum:       an.Expense.Sum,
			Avg:       an.Expense.Avg,
			Count:     an.Expense.Count,
			Median:    an.Expense.Median,
			Percent90: an.Expense.Percent90,
		},
		Details: details,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
