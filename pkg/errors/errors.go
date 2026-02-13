package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// AppError represents an application error
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Detail     string `json:"detail,omitempty"`
	Internal   error  `json:"-"`
	StackTrace string `json:"-"`
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// Unwrap returns the internal error
func (e *AppError) Unwrap() error {
	return e.Internal
}

// Common error codes
const (
	ErrCodeOK           = http.StatusOK
	ErrCodeBadRequest   = http.StatusBadRequest
	ErrCodeUnauthorized = http.StatusUnauthorized
	ErrCodeForbidden    = http.StatusForbidden
	ErrCodeNotFound     = http.StatusNotFound
	ErrCodeConflict     = http.StatusConflict
	ErrCodeInternal     = http.StatusInternalServerError
	ErrCodeServiceUnavailable = http.StatusServiceUnavailable
)

// Common errors
var (
	ErrNotFound = &AppError{
		Code:    ErrCodeNotFound,
		Message: "Resource not found",
	}

	ErrUnauthorized = &AppError{
		Code:    ErrCodeUnauthorized,
		Message: "Authentication required",
	}

	ErrForbidden = &AppError{
		Code:    ErrCodeForbidden,
		Message: "Access denied",
	}

	ErrInvalidInput = &AppError{
		Code:    ErrCodeBadRequest,
		Message: "Invalid input provided",
	}

	ErrInternalServer = &AppError{
		Code:    ErrCodeInternal,
		Message: "An unexpected error occurred",
	}

	ErrServiceUnavailable = &AppError{
		Code:    ErrCodeServiceUnavailable,
		Message: "Service temporarily unavailable",
	}
)

// NewAppError creates a new AppError
func NewAppError(code int, message string, detail string, internal error) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Detail:    detail,
		Internal:  internal,
		StackTrace: middleware.GetStackDir(2),
	}
}

// WrapError wraps an existing error with context
func WrapError(err error, code int, message string) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return &AppError{
			Code:      code,
			Message:   message,
			Internal:  err,
			StackTrace: middleware.GetStackDir(2),
		}
	}
	return &AppError{
		Code:      code,
		Message:   message,
		Detail:    err.Error(),
		Internal:  err,
		StackTrace: middleware.GetStackDir(2),
	}
}

// NotFoundError creates a not found error
func NotFoundError(resource string, id string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Detail:  fmt.Sprintf("ID: %s", id),
	}
}

// ValidationError creates a validation error
func ValidationErrors(fields []string) *AppError {
	return &AppError{
		Code:    ErrCodeBadRequest,
		Message: "Validation failed",
		Detail:  fmt.Sprintf("Invalid fields: %v", fields),
	}
}

// ConflictError creates a conflict error
func ConflictError(resource string, id string) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: fmt.Sprintf("%s already exists", resource),
		Detail:  fmt.Sprintf("ID: %s", id),
	}
}

// Is checks if the error is of the given type
func Is(err error, target *AppError) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == target.Code
	}
	return false
}

// HTTPStatus returns the HTTP status code for the error
func HTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return http.StatusInternalServerError
}

// Response creates an error response
func Response(err error) map[string]interface{} {
	var appErr *AppError
	if errors.As(err, &appErr) {
		response := map[string]interface{}{
			"error":   appErr.Message,
			"code":    appErr.Code,
		}
		if appErr.Detail != "" {
			response["details"] = appErr.Detail
		}
		return response
	}
	return map[string]interface{}{
		"error":   "Internal Server Error",
		"code":    http.StatusInternalServerError,
		"details": err.Error(),
	}
}