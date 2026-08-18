package notifications

import (
	"github.com/google/uuid"
)

// ValidateNotificationType validates notification type
func ValidateNotificationType(notificationType string) error {
	validTypes := map[string]bool{
		"low_stock":          true,
		"debt_overdue":       true,
		"warranty_expiring":  true,
		"return_request":     true,
		"expense_approval":   true,
		"sales_update":       true,
		"purchase":           true,
		"general":            true,
	}
	
	if !validTypes[notificationType] {
		return ErrInvalidNotificationType
	}
	return nil
}

// ValidatePriority validates priority
func ValidatePriority(priority string) error {
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
		"urgent": true,
	}
	
	if !validPriorities[priority] {
		return ErrInvalidPriority
	}
	return nil
}

// ValidateUserID validates user ID
func ValidateUserID(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrUserNotFound
	}
	return nil
}
