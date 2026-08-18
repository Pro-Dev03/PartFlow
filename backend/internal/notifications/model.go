package notifications

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents a notification
type Notification struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	Type           string     `json:"type" db:"type"` // low_stock, debt_overdue, warranty_expiring, return_request, expense_approval, etc.
	Title          string     `json:"title" db:"title"`
	Message        string     `json:"message" db:"message"`
	Data           string     `json:"data" db:"data"` // JSON string with additional data
	Priority       string     `json:"priority" db:"priority"` // low, medium, high, urgent
	Status         string     `json:"status" db:"status"` // unread, read, archived
	ActionURL      string     `json:"action_url" db:"action_url"`
	ActionText     string     `json:"action_text" db:"action_text"`
	ExpiresAt      *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	ReadAt         *time.Time `json:"read_at" db:"read_at"`
}

// NotificationRequest represents notification creation request
type NotificationRequest struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	Type        string    `json:"type" binding:"required"`
	Title       string    `json:"title" binding:"required"`
	Message     string    `json:"message" binding:"required"`
	Data        string    `json:"data"`
	Priority    string    `json:"priority" binding:"required,oneof=low medium high urgent"`
	ActionURL   string    `json:"action_url"`
	ActionText  string    `json:"action_text"`
	ExpiresIn  int       `json:"expires_in"` // in hours
}

// NotificationUpdateRequest represents notification update request
type NotificationUpdateRequest struct {
	Status string `json:"status" binding:"omitempty,oneof=unread read archived"`
}

// NotificationListRequest represents notification list query parameters
type NotificationListRequest struct {
	Page      int    `form:"page" binding:"min=1"`
	PerPage   int    `form:"per_page" binding:"min=1,max=100"`
	Type      string `form:"type"`
	Status    string `form:"status" binding:"omitempty,oneof=unread read archived"`
	Priority  string `form:"priority" binding:"omitempty,oneof=low medium high urgent"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	EmailEnabled   bool      `json:"email_enabled" db:"email_enabled"`
	PushEnabled    bool      `json:"push_enabled" db:"push_enabled"`
	LowStock       bool      `json:"low_stock" db:"low_stock"`
	DebtOverdue    bool      `json:"debt_overdue" db:"debt_overdue"`
	WarrantyExpiring bool    `json:"warranty_expiring" db:"warranty_expiring"`
	ReturnRequests bool      `json:"return_requests" db:"return_requests"`
	ExpenseApproval bool     `json:"expense_approval" db:"expense_approval"`
	SalesUpdates   bool      `json:"sales_updates" db:"sales_updates"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationSummary represents notification summary
type NotificationSummary struct {
	Total       int `json:"total"`
	Unread      int `json:"unread"`
	Priority    int `json:"priority"`
	Urgent      int `json:"urgent"`
	ByType      map[string]int `json:"by_type"`
}
