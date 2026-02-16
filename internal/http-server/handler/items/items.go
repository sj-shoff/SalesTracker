package items_handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"sales-tracker/internal/domain"
	customErr "sales-tracker/internal/domain/errors"
	"sales-tracker/internal/http-server/handler/items/dto"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/zlog"
)

type ItemsHandler struct {
	itemsUsecase     itemsUsecase
	analyticsUsecase analyticsUsecase
	logger           *zlog.Zerolog
}

func NewHandler(itemsUsecase itemsUsecase, analyticsUsecase analyticsUsecase, logger *zlog.Zerolog) *ItemsHandler {
	return &ItemsHandler{
		itemsUsecase:     itemsUsecase,
		analyticsUsecase: analyticsUsecase,
		logger:           logger,
	}
}

func (h *ItemsHandler) writeError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	s := "internal"
	switch {
	case errors.Is(err, customErr.ErrInvalidInput),
		errors.Is(err, customErr.ErrInvalidAmount),
		errors.Is(err, customErr.ErrInvalidItemType),
		errors.Is(err, customErr.ErrInvalidDateRange),
		errors.Is(err, customErr.ErrMissingParameter):
		code = http.StatusBadRequest
		s = "bad_request"
	case errors.Is(err, customErr.ErrItemNotFound):
		code = http.StatusNotFound
		s = "not_found"
	case errors.Is(err, customErr.ErrDatabase):
		code = http.StatusInternalServerError
		s = "database_error"
	}
	http.Error(w, s, code)
}

func (h *ItemsHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode request")
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	if req.Type == "" {
		h.logger.Warn().Msg("Missing required field: type")
		h.writeError(w, customErr.ErrMissingParameter)
		return
	}
	if req.Amount <= 0 {
		h.logger.Warn().Float64("amount", req.Amount).Msg("Invalid amount")
		h.writeError(w, customErr.ErrInvalidAmount)
		return
	}
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		h.logger.Warn().Err(err).Str("date", req.Date).Msg("Invalid date format")
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	item := &domain.Item{
		Type:        req.Type,
		Amount:      req.Amount,
		Date:        date,
		Category:    req.Category,
		Description: req.Description,
	}
	id, err := h.itemsUsecase.CreateItem(r.Context(), item)
	if err != nil {
		h.logger.Error().Err(err).Msg("CreateItem failed")
		h.writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
	h.logger.Info().Int64("id", id).Msg("Item created")
}

func (h *ItemsHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	page := 1
	limit := 25
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err != nil || p <= 0 {
			h.logger.Warn().Str("page", pageStr).Msg("Invalid page parameter")
			h.writeError(w, customErr.ErrInvalidInput)
			return
		} else {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err != nil || l <= 0 || l > 100 {
			h.logger.Warn().Str("limit", limitStr).Msg("Invalid limit parameter")
			h.writeError(w, customErr.ErrInvalidInput)
			return
		} else {
			limit = l
		}
	}
	offset := (page - 1) * limit
	items, total, err := h.itemsUsecase.GetItemsWithPagination(r.Context(), offset, limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetItems failed")
		h.writeError(w, err)
		return
	}
	resp := dto.ItemsResponse{
		Items: make([]*dto.ItemResponse, len(items)),
		Total: total,
		Page:  page,
		Limit: limit,
	}
	for i, it := range items {
		resp.Items[i] = &dto.ItemResponse{
			ID:          it.ID,
			Type:        it.Type,
			Amount:      it.Amount,
			Date:        it.Date,
			Category:    it.Category,
			Description: it.Description,
			CreatedAt:   it.CreatedAt,
			UpdatedAt:   it.UpdatedAt,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	h.logger.Info().Int("count", len(items)).Int64("total", total).Msg("Items retrieved with pagination")
}

func (h *ItemsHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.logger.Error().Err(err).Str("id", idStr).Msg("Invalid ID")
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	item, err := h.itemsUsecase.GetItemByID(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Int64("id", id).Msg("Failed to get item")
		h.writeError(w, err)
		return
	}
	resp := dto.ItemResponse{
		ID:          item.ID,
		Type:        item.Type,
		Amount:      item.Amount,
		Date:        item.Date,
		Category:    item.Category,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	h.logger.Info().Int64("id", id).Msg("Item retrieved")
}

func (h *ItemsHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.logger.Error().Err(err).Str("id", idStr).Msg("Invalid ID")
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	var req dto.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode request")
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		h.logger.Error().Err(err).Msg("Validation failed")
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	item, err := h.itemsUsecase.GetItemByID(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetItemByID failed")
		h.writeError(w, err)
		return
	}
	if req.Type != "" {
		item.Type = req.Type
	}
	if req.Amount != nil {
		item.Amount = *req.Amount
	}
	if req.Date != "" {
		date, err := time.Parse(time.RFC3339, req.Date)
		if err != nil {
			h.writeError(w, customErr.ErrInvalidInput)
			return
		}
		item.Date = date
	}
	if req.Category != "" {
		item.Category = req.Category
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	err = h.itemsUsecase.UpdateItem(r.Context(), id, item)
	if err != nil {
		h.logger.Error().Err(err).Msg("UpdateItem failed")
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	h.logger.Info().Int64("id", id).Msg("Item updated")
}

func (h *ItemsHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.logger.Error().Err(err).Str("id", idStr).Msg("Invalid ID")
		h.writeError(w, customErr.ErrInvalidInput)
		return
	}
	err = h.itemsUsecase.DeleteItem(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("DeleteItem failed")
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	h.logger.Info().Int64("id", id).Msg("Item deleted")
}

func (h *ItemsHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("Exporting items to CSV")

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var from, to time.Time
	var err error

	if fromStr != "" && toStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			h.logger.Warn().Err(err).Str("from", fromStr).Msg("Invalid from date format")
			h.writeError(w, customErr.ErrUnsupportedFormat)
			return
		}
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			h.logger.Warn().Err(err).Str("to", toStr).Msg("Invalid to date format")
			h.writeError(w, customErr.ErrUnsupportedFormat)
			return
		}
	} else {
		from = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		to = time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	analytics, err := h.analyticsUsecase.GetAnalytics(r.Context(), from, to)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get analytics for export")
		h.writeError(w, err)
		return
	}

	items := analytics.Details

	filename := fmt.Sprintf("sales_tracker_%s_%s.csv",
		from.Format("2006-01-02"),
		to.Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	w.Write([]byte("\xef\xbb\xbf"))

	writer := csv.NewWriter(w)
	writer.Comma = ';'
	writer.UseCRLF = true

	writer.Write([]string{"ОТЧЁТ SalesTracker"})
	writer.Write([]string{fmt.Sprintf("Период: %s - %s",
		from.Format("02.01.2006 15:04"),
		to.Format("02.01.2006 15:04"))})
	writer.Write([]string{""})

	writer.Write([]string{"АНАЛИТИКА ДОХОДОВ"})
	writer.Write([]string{"Сумма", "Среднее", "Количество", "Медиана", "90-й перцентиль"})
	writer.Write([]string{
		fmt.Sprintf("%.2f ₽", analytics.Income.Sum),
		fmt.Sprintf("%.2f ₽", analytics.Income.Avg),
		strconv.FormatInt(analytics.Income.Count, 10),
		fmt.Sprintf("%.2f ₽", analytics.Income.Median),
		fmt.Sprintf("%.2f ₽", analytics.Income.Percent90),
	})
	writer.Write([]string{""})

	writer.Write([]string{"АНАЛИТИКА РАСХОДОВ"})
	writer.Write([]string{"Сумма", "Среднее", "Количество", "Медиана", "90-й перцентиль"})
	writer.Write([]string{
		fmt.Sprintf("%.2f ₽", analytics.Expense.Sum),
		fmt.Sprintf("%.2f ₽", analytics.Expense.Avg),
		strconv.FormatInt(analytics.Expense.Count, 10),
		fmt.Sprintf("%.2f ₽", analytics.Expense.Median),
		fmt.Sprintf("%.2f ₽", analytics.Expense.Percent90),
	})
	writer.Write([]string{""})

	writer.Write([]string{"ОПЕРАЦИИ"})
	headers := []string{"ID", "Тип", "Сумма", "Дата", "Категория", "Описание", "Создано", "Обновлено"}
	writer.Write(headers)

	for _, item := range items {
		row := []string{
			strconv.FormatInt(item.ID, 10),
			h.getTypeLabel(item.Type),
			fmt.Sprintf("%.2f", item.Amount),
			item.Date.Format("02.01.2006 15:04"),
			item.Category,
			item.Description,
			item.CreatedAt.Format("02.01.2006 15:04"),
			item.UpdatedAt.Format("02.01.2006 15:04"),
		}
		writer.Write(row)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		h.logger.Error().Err(err).Msg("Failed to flush CSV")
		return
	}

	h.logger.Info().
		Int("count", len(items)).
		Time("from", from).
		Time("to", to).
		Msg("Report exported to CSV")
}

func (h *ItemsHandler) getTypeLabel(typeStr string) string {
	if typeStr == "income" {
		return "Доход"
	}
	return "Расход"
}
