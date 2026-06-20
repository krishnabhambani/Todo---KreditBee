package apperrors

import "net/http"

// AppError is a custom error type that carries an HTTP status code
// along with the user-facing error message.
type AppError struct {
	StatusCode int
	Message    string
	Err        error // Underlying original error, if any, for internal logging
}

// Error implements the standard error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// -----------------------------------------------------------------------------
// Constructors
// -----------------------------------------------------------------------------

func NewBadRequest(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

func NewUnauthorized(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	}
}

func NewForbidden(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusForbidden,
		Message:    message,
	}
}

func NewNotFound(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Message:    message,
	}
}

func NewInternal(err error, message string) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
		Err:        err,
	}
}

