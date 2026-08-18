package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditMiddlewareConfig configures the audit middleware
type AuditMiddlewareConfig struct {
	AuditService interface{} // This would be the actual audit service
	SkipPaths    []string    // Paths to skip audit logging
}

// AuditMiddleware logs all requests to the audit system
func AuditMiddleware(config AuditMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit logging for specified paths
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		start := time.Now()

		// Read request body for logging
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create a response writer wrapper to capture response
		w := &responseWriter{ResponseWriter: c.Writer, body: bytes.NewBufferString("")}
		c.Writer = w

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get user and organization context
		userID := GetUserID(c)
		organizationID := GetOrganizationID(c)
		requestID := GetRequestID(c)

		// Determine action based on HTTP method
		action := getActionFromMethod(c.Request.Method)

		// Determine entity type from path
		entityType := getEntityTypeFromPath(c.Request.URL.Path)

		// Determine status
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "failure"
		}

		// Prepare audit log data
		auditData := map[string]interface{}{
			"user_id":         userID,
			"organization_id": organizationID,
			"action":          action,
			"entity_type":     entityType,
			"request_id":      requestID,
			"method":          c.Request.Method,
			"path":            c.Request.URL.Path,
			"query":           c.Request.URL.RawQuery,
			"status_code":     c.Writer.Status(),
			"duration":        duration.Milliseconds(),
			"ip_address":      c.ClientIP(),
			"user_agent":      c.GetHeader("User-Agent"),
			"status":          status,
			"request_body":    string(bodyBytes),
			"response_body":   w.body.String(),
			"error_message":   getErrorMessage(c.Errors),
		}

		// Log to audit service (this would be the actual service call)
		// For now, we'll just log it using Gin's built-in logger
		c.Set("audit_data", auditData)

		// Log the audit data
		c.Set("audit_logged", true)
	}
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// getActionFromMethod converts HTTP method to audit action
func getActionFromMethod(method string) string {
	switch method {
	case "POST":
		return "create"
	case "GET":
		return "view"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

// getEntityTypeFromPath extracts entity type from URL path
func getEntityTypeFromPath(path string) string {
	// Simple extraction based on common patterns
	// This could be enhanced with more sophisticated parsing
	if len(path) > 1 {
		// Remove leading slash and get first segment
		if path[0] == '/' {
			path = path[1:]
		}
		// Find first slash and extract entity type
		for i, char := range path {
			if char == '/' {
				return path[:i]
			}
		}
		return path
	}
	return "unknown"
}

// getErrorMessage extracts error message from gin errors
func getErrorMessage(errors []*gin.Error) string {
	if len(errors) > 0 {
		return errors[0].Error()
	}
	return ""
}

// AuditLogHelper provides helper functions for manual audit logging
type AuditLogHelper struct {
	// This would contain the actual audit service
}

// NewAuditLogHelper creates a new audit log helper
func NewAuditLogHelper() *AuditLogHelper {
	return &AuditLogHelper{}
}

// LogAction logs a specific action manually
func (h *AuditLogHelper) LogAction(c *gin.Context, action, entityType string, entityID uuid.UUID, description string, status string) {
	userID := GetUserID(c)
	organizationID := GetOrganizationID(c)
	requestID := GetRequestID(c)

	auditData := map[string]interface{}{
		"user_id":         userID,
		"organization_id": organizationID,
		"action":          action,
		"entity_type":     entityType,
		"entity_id":       entityID,
		"request_id":      requestID,
		"description":     description,
		"status":          status,
		"ip_address":      c.ClientIP(),
		"user_agent":      c.GetHeader("User-Agent"),
	}

	c.Set("manual_audit_log", auditData)
}

// LogError logs an error event
func (h *AuditLogHelper) LogError(c *gin.Context, action, entityType string, entityID uuid.UUID, description, errorMessage string) {
	h.LogAction(c, action, entityType, entityID, description, "failure")
	
	// Add error message
	if auditData, exists := c.Get("manual_audit_log"); exists {
		if data, ok := auditData.(map[string]interface{}); ok {
			data["error_message"] = errorMessage
		}
	}
}

// GetAuditData retrieves audit data from context
func GetAuditData(c *gin.Context) map[string]interface{} {
	if auditData, exists := c.Get("audit_data"); exists {
		return auditData.(map[string]interface{})
	}
	return nil
}

// GetManualAuditLog retrieves manual audit log from context
func GetManualAuditLog(c *gin.Context) map[string]interface{} {
	if auditData, exists := c.Get("manual_audit_log"); exists {
		return auditData.(map[string]interface{})
	}
	return nil
}

// HasAuditLog checks if audit log was created for this request
func HasAuditLog(c *gin.Context) bool {
	_, hasAutoAudit := c.Get("audit_data")
	_, hasManualAudit := c.Get("manual_audit_log")
	return hasAutoAudit || hasManualAudit
}