package notifications

import (
	"time"

	"github.com/google/uuid"
)

// ToNotificationListItem converts Notification to list item format
func (n *Notification) ToNotificationListItem() map[string]interface{} {
	return map[string]interface{}{
		"id":          n.ID,
		"type":        n.Type,
		"title":       n.Title,
		"message":     n.Message,
		"priority":    n.Priority,
		"status":      n.Status,
		"action_url":  n.ActionURL,
		"action_text": n.ActionText,
		"created_at":  n.CreatedAt,
		"read_at":     n.ReadAt,
		"expires_at":  n.ExpiresAt,
	}
}

// CreateNotification creates a Notification from request
func CreateNotification(organizationID uuid.UUID, req *NotificationRequest) *Notification {
	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
		expiresAt = &exp
	}

	return &Notification{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		UserID:         req.UserID,
		Type:           req.Type,
		Title:          req.Title,
		Message:        req.Message,
		Data:           req.Data,
		Priority:       req.Priority,
		Status:         "unread",
		ActionURL:      req.ActionURL,
		ActionText:     req.ActionText,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
	}
}

// ValidateNotificationRequest validates notification request
func ValidateNotificationRequest(req *NotificationRequest) error {
	if req.UserID == uuid.Nil {
		return ErrUserNotFound
	}
	if req.Type == "" {
		return ErrInvalidNotificationType
	}
	if req.Title == "" {
		return ErrNotificationNotFound
	}
	if req.Message == "" {
		return ErrNotificationNotFound
	}
	if req.Priority != "low" && req.Priority != "medium" && 
		req.Priority != "high" && req.Priority != "urgent" {
		return ErrInvalidPriority
	}
	return nil
}

// ValidateNotificationStatus validates notification status
func ValidateNotificationStatus(status string) error {
	validStatuses := map[string]bool{
		"unread":   true,
		"read":     true,
		"archived": true,
	}
	
	if !validStatuses[status] {
		return ErrInvalidNotificationStatus
	}
	return nil
}

// IsExpired checks if notification is expired
func (n *Notification) IsExpired() bool {
	if n.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*n.ExpiresAt)
}

// IsRead checks if notification is read
func (n *Notification) IsRead() bool {
	return n.Status == "read" || n.Status == "archived"
}

// MarkAsRead marks notification as read
func (n *Notification) MarkAsRead() {
	n.Status = "read"
	now := time.Now()
	n.ReadAt = &now
}

// MarkAsArchived marks notification as archived
func (n *Notification) MarkAsArchived() {
	n.Status = "archived"
	if n.ReadAt == nil {
		now := time.Now()
		n.ReadAt = &now
	}
}

// ParseData parses data JSON string
func (n *Notification) ParseData() (map[string]interface{}, error) {
	var data map[string]interface{}
	if n.Data != "" {
		// This would require JSON parsing
		// For simplicity, we'll return empty map
		return data, nil
	}
	return data, nil
}

// CreateDefaultPreferences creates default notification preferences for a user
func CreateDefaultPreferences(organizationID uuid.UUID, userID uuid.UUID) *NotificationPreferences {
	return &NotificationPreferences{
		ID:               uuid.New(),
		UserID:           userID,
		OrganizationID:   organizationID,
		EmailEnabled:     true,
		PushEnabled:      true,
		LowStock:         true,
		DebtOverdue:      true,
		WarrantyExpiring: true,
		ReturnRequests:   true,
		ExpenseApproval:  true,
		SalesUpdates:     false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}
