package errors

import "errors"

// Бизнес-ошибки
var (
	ErrItemNotFound      = errors.New("item not found")
	ErrInvalidItemType   = errors.New("invalid item type")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidDateRange  = errors.New("invalid date range")
	ErrInvalidInput      = errors.New("invalid input")
	ErrMissingParameter  = errors.New("missing required parameter")
	ErrUnsupportedFormat = errors.New("unsupported date format")
	ErrPeriodTooLarge    = errors.New("date range exceeds maximum allowed period")
)

// Технические ошибки
var (
	ErrDatabase = errors.New("database error")
	ErrInternal = errors.New("internal error")
	ErrTimeout  = errors.New("operation timeout")
)
