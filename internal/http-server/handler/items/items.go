package items_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sales-tracker/internal/domain"
	customErr "sales-tracker/internal/domain/errors"
	"sales-tracker/internal/http-server/handler/items/dto"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wb-go/wbf/zlog"
)

type ItemsHandler struct {
	itemsUsecase itemsUsecase
	logger       *zlog.Zerolog
}

func NewHandler(itemsUsecase itemsUsecase, logger *zlog.Zerolog) *ItemsHandler {
	return &ItemsHandler{
		itemsUsecase: itemsUsecase,
		logger:       logger,
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

	// Валидация обязательных полей
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
	if req.Date.IsZero() {
		h.logger.Warn().Msg("Missing required field: date")
		h.writeError(w, customErr.ErrMissingParameter)
		return
	}

	item := &domain.Item{
		Type:        req.Type,
		Amount:      req.Amount,
		Date:        req.Date,
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

	item, err := h.itemsUsecase.GetItemByID(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("GetItemByID failed")
		h.writeError(w, err)
		return
	}

	if req.Type != "" {
		item.Type = req.Type
	}
	if req.Amount != 0 {
		item.Amount = req.Amount
	}
	if !req.Date.IsZero() {
		item.Date = req.Date
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
