package errors

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/partflow/smart-store/pkg/logger"
	"github.com/partflow/smart-store/pkg/response"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeNotFound     ErrorType = "not_found"
	ErrorTypeConflict     ErrorType = "conflict"
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	ErrorTypeForbidden    ErrorType = "forbidden"
	ErrorTypeInternal     ErrorType = "internal"
	ErrorTypeDatabase     ErrorType = "database"
	ErrorTypeExternal     ErrorType = "external"
	ErrorTypeBusiness     ErrorType = "business"
)

// AppError represents an application error
type AppError struct {
	Type       ErrorType              `json:"type"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Err        error                  `json:"-"`
	Context    map[string]interface{} `json:"context,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code, message string, statusCode int, err error) *AppError {
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		Context:    make(map[string]interface{}),
	}
}

// Common error constructors
func BadRequest(message string, err error) *AppError {
	return NewAppError(ErrorTypeValidation, "BAD_REQUEST", message, http.StatusBadRequest, err)
}

func Unauthorized(message string, err error) *AppError {
	return NewAppError(ErrorTypeUnauthorized, "UNAUTHORIZED", message, http.StatusUnauthorized, err)
}

func Forbidden(message string, err error) *AppError {
	return NewAppError(ErrorTypeForbidden, "FORBIDDEN", message, http.StatusForbidden, err)
}

func NotFound(message string, err error) *AppError {
	return NewAppError(ErrorTypeNotFound, "NOT_FOUND", message, http.StatusNotFound, err)
}

func Conflict(message string, err error) *AppError {
	return NewAppError(ErrorTypeConflict, "CONFLICT", message, http.StatusConflict, err)
}

func InternalServerError(message string, err error) *AppError {
	return NewAppError(ErrorTypeInternal, "INTERNAL_SERVER_ERROR", message, http.StatusInternalServerError, err)
}

func ServiceUnavailable(message string, err error) *AppError {
	return NewAppError(ErrorTypeExternal, "SERVICE_UNAVAILABLE", message, http.StatusServiceUnavailable, err)
}

// Business logic errors
func ValidationError(message string, err error) *AppError {
	return NewAppError(ErrorTypeValidation, "VALIDATION_ERROR", message, http.StatusBadRequest, err)
}

func InsufficientStockError(message string, err error) *AppError {
	return NewAppError(ErrorTypeBusiness, "INSUFFICIENT_STOCK", message, http.StatusConflict, err)
}

func InvalidStatusError(message string, err error) *AppError {
	return NewAppError(ErrorTypeBusiness, "INVALID_STATUS", message, http.StatusBadRequest, err)
}

func DuplicateResourceError(message string, err error) *AppError {
	return NewAppError(ErrorTypeConflict, "DUPLICATE_RESOURCE", message, http.StatusConflict, err)
}

// Specific error constructors for handler functions
func NewValidationError(message string, err error) *AppError {
	return NewAppError(ErrorTypeValidation, "VALIDATION_ERROR", message, http.StatusBadRequest, err)
}

func NewNotFoundError(resource string, err error) *AppError {
	return NewAppError(ErrorTypeNotFound, "NOT_FOUND", fmt.Sprintf("%s not found", resource), http.StatusNotFound, err)
}

func NewConflictError(message string, err error) *AppError {
	return NewAppError(ErrorTypeConflict, "CONFLICT", message, http.StatusConflict, err)
}

func NewUnauthorizedError(message string, err error) *AppError {
	return NewAppError(ErrorTypeUnauthorized, "UNAUTHORIZED", message, http.StatusUnauthorized, err)
}

func NewForbiddenError(message string, err error) *AppError {
	return NewAppError(ErrorTypeForbidden, "FORBIDDEN", message, http.StatusForbidden, err)
}

func NewInternalError(message string, err error) *AppError {
	return NewAppError(ErrorTypeInternal, "INTERNAL_SERVER_ERROR", message, http.StatusInternalServerError, err)
}

func NewDatabaseError(message string, err error) *AppError {
	return NewAppError(ErrorTypeDatabase, "DATABASE_ERROR", message, http.StatusInternalServerError, err)
}

func NewBusinessError(message string, err error) *AppError {
	return NewAppError(ErrorTypeBusiness, "BUSINESS_ERROR", message, http.StatusBadRequest, err)
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// WithRequestID adds request ID to the error
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError converts an error to AppError if possible
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return InternalServerError("An unexpected error occurred", err)
}

// HandleError handles errors in a standardized way
func HandleError(c *gin.Context, err error) {
	requestID := getRequestID(c)
	
	// Check if it's an AppError
	appErr, isAppErr := GetAppError(err), IsAppError(err)
	
	if isAppErr {
		// Add request ID if not set
		if appErr.RequestID == "" {
			appErr.WithRequestID(requestID)
		}
		
		// Log the error based on type
		logAppError(c, appErr)
		
		// Send response
		response.Error(c, appErr.StatusCode, appErr.StatusCode, appErr.Message, getErrorMessage(err))
		return
	}
	
	// Handle generic errors
	handleGenericError(c, err, requestID)
}

// logAppError logs application errors based on type
func logAppError(c *gin.Context, appErr *AppError) {
	fields := map[string]interface{}{
		"error_type":   appErr.Type,
		"status_code":  appErr.StatusCode,
		"request_id":   appErr.RequestID,
		"path":         c.Request.URL.Path,
		"method":       c.Request.Method,
	}
	
	// Add context fields
	for k, v := range appErr.Context {
		fields[k] = v
	}
	
	switch appErr.Type {
	case ErrorTypeValidation:
		logger.Warn("Validation error", fields)
	case ErrorTypeNotFound:
		logger.Info("Resource not found", fields)
	case ErrorTypeConflict:
		logger.Warn("Conflict error", fields)
	case ErrorTypeUnauthorized:
		logger.Warn("Unauthorized access attempt", fields)
	case ErrorTypeForbidden:
		logger.Warn("Forbidden access attempt", fields)
	case ErrorTypeDatabase:
		logger.Error("Database error", appErr.Err, fields)
	case ErrorTypeExternal:
		logger.Error("External service error", appErr.Err, fields)
	case ErrorTypeBusiness:
		logger.Info("Business logic error", fields)
	case ErrorTypeInternal:
		logger.Error("Internal server error", appErr.Err, fields)
	}
}

// handleGenericError handles generic (non-AppError) errors
func handleGenericError(c *gin.Context, err error, requestID string) {
	fields := map[string]interface{}{
		"error_type":  "generic",
		"request_id":  requestID,
		"path":        c.Request.URL.Path,
		"method":      c.Request.Method,
		"error":       err.Error(),
	}
	
	// Log the error
	logger.Error("Unhandled error", err, fields)
	
	// Send generic error response
	response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Internal server error", "An unexpected error occurred")
}

// ErrorHandlerMiddleware is a middleware that handles panics and errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := getRequestID(c)
				
				// Log panic with stack trace
				fields := map[string]interface{}{
					"error_type":  "panic",
					"request_id":  requestID,
					"path":        c.Request.URL.Path,
					"method":      c.Request.Method,
					"stack_trace": string(debug.Stack()),
				}
				
				logger.Error("Panic recovered", fmt.Errorf("%v", err), fields)
				
				// Send error response
				response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Internal server error", "An unexpected error occurred")
				c.Abort()
			}
		}()
		
		c.Next()
		
		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				HandleError(c, ginErr.Err)
				return
			}
		}
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapErrorWithType wraps an error with specific error type
func WrapErrorWithType(err error, errorType ErrorType, message string, statusCode int) error {
	if err == nil {
		return nil
	}
	return &AppError{
		Type:       errorType,
		Code:       string(errorType),
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		Context:    make(map[string]interface{}),
	}
}

// getRequestID gets request ID from context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// getErrorMessage extracts error message
func getErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	
	// Check if it's an AppError
	if appErr, ok := err.(*AppError); ok {
		if appErr.Err != nil {
			return appErr.Err.Error()
		}
		return appErr.Message
	}
	
	return err.Error()
}

// ValidateRequest validates request and returns appropriate error
func ValidateRequest(err error) error {
	if err == nil {
		return nil
	}
	return NewValidationError("Invalid request parameters", err)
}

// ValidateResource checks if resource exists
func ValidateResource(resource string, exists bool) error {
	if !exists {
		return NewNotFoundError(resource, nil)
	}
	return nil
}

// CheckPermission checks if user has permission
func CheckPermission(resource, action string, hasPermission bool) error {
	if !hasPermission {
		return NewForbiddenError(
			fmt.Sprintf("You don't have permission to %s %s", action, resource),
			nil,
		)
	}
	return nil
}

// CheckOwnership checks if user owns the resource
func CheckOwnership(resource string, isOwner bool) error {
	if !isOwner {
		return NewForbiddenError(
			fmt.Sprintf("You don't own this %s", resource),
			nil,
		)
	}
	return nil
}

// RetryableError checks if error is retryable
func RetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	appErr, isAppErr := GetAppError(err), IsAppError(err)
	if !isAppErr {
		return false
	}
	
	switch appErr.Type {
	case ErrorTypeDatabase, ErrorTypeExternal:
		return true
	default:
		return false
	}
}

// GetUserFriendlyMessage returns a user-friendly error message
func GetUserFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}
	
	appErr, isAppErr := GetAppError(err), IsAppError(err)
	if !isAppErr {
		return "An unexpected error occurred. Please try again."
	}
	
	switch appErr.Type {
	case ErrorTypeValidation:
		return appErr.Message
	case ErrorTypeNotFound:
		return appErr.Message
	case ErrorTypeConflict:
		return appErr.Message
	case ErrorTypeUnauthorized:
		return "Please log in to continue"
	case ErrorTypeForbidden:
		return "You don't have permission to perform this action"
	case ErrorTypeBusiness:
		return appErr.Message
	default:
		return "An unexpected error occurred. Please try again."
	}
}

// ParseDatabaseError parses database errors and converts to AppError
func ParseDatabaseError(err error, operation string) error {
	if err == nil {
		return nil
	}
	
	errMsg := strings.ToLower(err.Error())
	
	if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique") {
		return NewConflictError("A record with this information already exists", err)
	}
	
	if strings.Contains(errMsg, "foreign key") {
		return NewBusinessError("This operation conflicts with related records", err)
	}
	
	if strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "timeout") {
		return NewDatabaseError("Database connection error", err)
	}
	
	return NewDatabaseError(fmt.Sprintf("Database error during %s", operation), err)
}
