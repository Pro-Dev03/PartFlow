package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Err        error  `json:"-"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new application error
func New(code int, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an existing error with application context
func Wrap(err error, code int, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Err:        err,
		StatusCode: statusCode,
	}
}

// Common error types
var (
	ErrBadRequest       = New(400, "Bad request", http.StatusBadRequest)
	ErrUnauthorized     = New(401, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden        = New(403, "Forbidden", http.StatusForbidden)
	ErrNotFound         = New(404, "Resource not found", http.StatusNotFound)
	ErrConflict         = New(409, "Resource already exists", http.StatusConflict)
	ErrValidation       = New(422, "Validation failed", http.StatusUnprocessableEntity)
	ErrInternalServer   = New(500, "Internal server error", http.StatusInternalServerError)
	ErrServiceUnavailable = New(503, "Service unavailable", http.StatusServiceUnavailable)
)

// Database errors
func ErrDatabase(err error) *AppError {
	return Wrap(err, 500, "Database error", http.StatusInternalServerError)
}

func ErrNotFound(resource string) *AppError {
	return New(404, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func ErrAlreadyExists(resource string) *AppError {
	return New(409, fmt.Sprintf("%s already exists", resource), http.StatusConflict)
}

// Validation errors
func ErrValidationFailed(field string, message string) *AppError {
	return New(422, fmt.Sprintf("Validation failed for %s: %s", field, message), http.StatusUnprocessableEntity)
}

// Authentication errors
func ErrInvalidCredentials() *AppError {
	return New(401, "Invalid credentials", http.StatusUnauthorized)
}

func ErrTokenExpired() *AppError {
	return New(401, "Token expired", http.StatusUnauthorized)
}

func ErrInvalidToken() *AppError {
	return New(401, "Invalid token", http.StatusUnauthorized)
}

// Authorization errors
func ErrInsufficientPermissions() *AppError {
	return New(403, "Insufficient permissions", http.StatusForbidden)
}

// Business logic errors
func ErrInvalidState(action string) *AppError {
	return New(422, fmt.Sprintf("Cannot %s in current state", action), http.StatusUnprocessableEntity)
}

func ErrQuotaExceeded(resource string) *AppError {
	return New(429, fmt.Sprintf("%s quota exceeded", resource), http.StatusTooManyRequests)
}
