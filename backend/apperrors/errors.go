package apperrors

import (
	"fmt"
	"net/http"
)

// ErrorCode is the standardized application error code.
type ErrorCode string

const (
	ErrBadRequest   ErrorCode = "400"
	ErrUnauthorized ErrorCode = "401"
	ErrForbidden    ErrorCode = "403"
	ErrNotFound     ErrorCode = "404"
	ErrConflict     ErrorCode = "409"
	ErrInternal     ErrorCode = "500"
	ErrValidation   ErrorCode = "422"
)

// AppError is the core application error type used across services and controllers.
type AppError struct {
	Code       ErrorCode
	StatusCode int
	Message    string
	Details    map[string]string
	Err        error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// New creates a new AppError with the given code, status and message.
func New(code ErrorCode, statusCode int, message string) *AppError {
	return &AppError{Code: code, StatusCode: statusCode, Message: message}
}

// Newf creates a new AppError with a formatted message.
func Newf(code ErrorCode, statusCode int, format string, args ...any) *AppError {
	return &AppError{Code: code, StatusCode: statusCode, Message: fmt.Sprintf(format, args...)}
}

// BadRequest creates a validation-style error.
func BadRequest(message string) *AppError {
	return New(ErrBadRequest, http.StatusBadRequest, message)
}

// Unauthorized creates an unauthorized error.
func Unauthorized(message string) *AppError {
	return New(ErrUnauthorized, http.StatusUnauthorized, message)
}

// Forbidden creates a forbidden error.
func Forbidden(message string) *AppError {
	return New(ErrForbidden, http.StatusForbidden, message)
}

// NotFound creates a not found error.
func NotFound(message string) *AppError {
	return New(ErrNotFound, http.StatusNotFound, message)
}

// Conflict creates a conflict error.
func Conflict(message string) *AppError {
	return New(ErrConflict, http.StatusConflict, message)
}

// Internal creates an internal server error.
func Internal(message string) *AppError {
	return New(ErrInternal, http.StatusInternalServerError, message)
}

// Validation creates a validation error.
func Validation(message string) *AppError {
	return New(ErrValidation, http.StatusUnprocessableEntity, message)
}

// Backward-compatible constructors for the existing app.
func NewBadRequest(message string) *AppError {
	return BadRequest(message)
}

func NewUnauthorized(message string) *AppError {
	return Unauthorized(message)
}

func NewForbidden(message string) *AppError {
	return Forbidden(message)
}

func NewNotFound(message string) *AppError {
	return NotFound(message)
}

func NewInternal(err error, message string) *AppError {
	return &AppError{Code: ErrInternal, StatusCode: http.StatusInternalServerError, Message: message, Err: err}
}
