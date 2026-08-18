package audit

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	Action         string     `json:"action" db:"action"` // create, update, delete, login, logout, etc.
	EntityType     string     `json:"entity_type" db:"entity_type"` // product, customer, sale, etc.
	EntityID       uuid.UUID  `json:"entity_id" db:"entity_id"`
	IPAddress     string     `json:"ip_address" db:"ip_address"`
	UserAgent      string     `json:"user_agent" db:"user_agent"`
	RequestID      string     `json:"request_id" db:"request_id"`
	Changes        string     `json:"changes" db:"changes"` // JSON string with before/after values
	Description    string     `json:"description" db:"description"`
	Status         string     `json:"status" db:"status"` // success, failure
	ErrorMessage   string     `json:"error_message" db:"error_message"`
	Metadata       string     `json:"metadata" db:"metadata"` // JSON string with additional metadata
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// AuditLogRequest represents audit log creation request
type AuditLogRequest struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	Action      string    `json:"action" binding:"required"`
	EntityID    uuid.UUID `json:"entity_id" binding:"required"`
	Changes     string    `json:"changes"`
	Description string    `json:"description"`
	Status      string    `json:"status" binding:"required,oneof=success failure"`
	ErrorMessage string  `json:"error_message"`
	Metadata    string    `json:"metadata"`
}

// AuditLogListRequest represents audit log list query parameters
type AuditLogListRequest struct {
	Page         int        `form:"page" binding:"min=1"`
	PerPage      int        `form:"per_page" binding:"min=1,max=100"`
	UserID       *uuid.UUID `form:"user_id"`
	Action       string     `form:"action"`
	EntityID     *uuid.UUID `form:"entity_id"`
	EntityType   string     `form:"entity_type"`
	Status       string     `form:"status" binding:"omitempty,oneof=success failure"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	Search       string     `form:"search"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}

// AuditLogSummary represents audit log summary statistics
type AuditLogSummary struct {
	TotalLogs       int              `json:"total_logs"`
	SuccessLogs    int              `json:"success_logs"`
	FailureLogs    int              `json:"failure_logs"`
	ByAction       map[string]int    `json:"by_action"`
	ByEntityType   map[string]int    `json:"by_entity_type"`
	ByUser        map[string]int    `json:"by_user"`
	ThisWeek       int              `json:"this_week"`
	ThisMonth      int              `json:"this_month"`
	RecentActivity []AuditLogEntry  `json:"recent_activity"`
}

// AuditLogEntry represents a simplified audit log entry
type AuditLogEntry struct {
	ID          uuid.UUID `json:"id"`
	Action      string    `json:"action"`
	EntityType  string    `json:"entity_type"`
	EntityID    uuid.UUID `json:"entity_id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name"`
	CreatedAt   time.Time `json:"created_at"`
}
