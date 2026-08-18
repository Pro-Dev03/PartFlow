package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles notification business logic
type Service struct {
	repo *Repository
}

// NewService creates a new notification service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateNotification creates a new notification
func (s *Service) CreateNotification(ctx context.Context, organizationID uuid.UUID, req *NotificationRequest) (*Notification, error) {
	// Validate request
	if err := ValidateNotificationRequest(req); err != nil {
		return nil, err
	}

	// Create notification
	notification := CreateNotification(organizationID, req)

	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

// GetNotification retrieves a notification by ID
func (s *Service) GetNotification(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Notification, error) {
	notification, err := s.repo.GetNotificationByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Check if notification is expired
	if notification.IsExpired() {
		return nil, ErrNotificationExpired
	}

	return notification, nil
}

// ListNotifications retrieves notifications for a user
func (s *Service) ListNotifications(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req NotificationListRequest) ([]Notification, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	notifications, total, err := s.repo.ListNotifications(ctx, organizationID, userID, req)
	if err != nil {
		return nil, 0, err
	}

	// Filter out expired notifications
	var validNotifications []Notification
	for _, notification := range notifications {
		if !notification.IsExpired() {
			validNotifications = append(validNotifications, notification)
		}
	}

	return validNotifications, total, nil
}

// MarkAsRead marks a notification as read
func (s *Service) MarkAsRead(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	notification, err := s.repo.GetNotificationByID(ctx, id, organizationID)
	if err != nil {
		return err
	}

	if notification.IsRead() {
		return ErrNotificationAlreadyRead
	}

	return s.repo.MarkAsRead(ctx, id, organizationID)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *Service) MarkAllAsRead(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID, organizationID)
}

// DeleteNotification deletes a notification
func (s *Service) DeleteNotification(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	return s.repo.DeleteNotification(ctx, id, organizationID)
}

// GetNotificationSummary retrieves notification summary for a user
func (s *Service) GetNotificationSummary(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (*NotificationSummary, error) {
	return s.repo.GetNotificationSummary(ctx, userID, organizationID)
}

// CreateNotificationPreferences creates notification preferences for a user
func (s *Service) CreateNotificationPreferences(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID) (*NotificationPreferences, error) {
	preferences := CreateDefaultPreferences(organizationID, userID)

	if err := s.repo.CreateNotificationPreferences(ctx, preferences); err != nil {
		return nil, fmt.Errorf("failed to create notification preferences: %w", err)
	}

	return preferences, nil
}

// GetNotificationPreferences retrieves notification preferences for a user
func (s *Service) GetNotificationPreferences(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (*NotificationPreferences, error) {
	preferences, err := s.repo.GetNotificationPreferences(ctx, userID, organizationID)
	if err != nil {
		// If preferences don't exist, create default ones
		if err == ErrPreferencesNotFound {
			return s.CreateNotificationPreferences(ctx, organizationID, userID)
		}
		return nil, err
	}

	return preferences, nil
}

// UpdateNotificationPreferences updates notification preferences
func (s *Service) UpdateNotificationPreferences(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, preferences *NotificationPreferences) (*NotificationPreferences, error) {
	// Get existing preferences first
	existing, err := s.repo.GetNotificationPreferences(ctx, userID, organizationID)
	if err != nil {
		return nil, err
	}

	// Update fields
	preferences.ID = existing.ID
	preferences.UserID = existing.UserID
	preferences.OrganizationID = existing.OrganizationID
	preferences.UpdatedAt = time.Now()

	if err := s.repo.UpdateNotificationPreferences(ctx, preferences); err != nil {
		return nil, err
	}

	return preferences, nil
}

// SendLowStockNotification sends low stock notification
func (s *Service) SendLowStockNotification(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, productName string, currentStock int, minStock int) error {
	req := &NotificationRequest{
		UserID:     userID,
		Type:       "low_stock",
		Title:      "Low Stock Alert",
		Message:    fmt.Sprintf("%s is running low. Current stock: %d, Minimum: %d", productName, currentStock, minStock),
		Priority:   "high",
		ActionURL:  "/inventory",
		ActionText: "View Inventory",
		ExpiresIn:  72, // 3 days
		Data:       fmt.Sprintf(`{"product_name": "%s", "current_stock": %d, "min_stock": %d}`, productName, currentStock, minStock),
	}

	_, err := s.CreateNotification(ctx, organizationID, req)
	return err
}

// SendDebtOverdueNotification sends debt overdue notification
func (s *Service) SendDebtOverdueNotification(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, customerName string, amount float64, overdueDays int) error {
	req := &NotificationRequest{
		UserID:     userID,
		Type:       "debt_overdue",
		Title:      "Debt Overdue Alert",
		Message:    fmt.Sprintf("%s has an overdue debt of %.2f for %d days", customerName, amount, overdueDays),
		Priority:   "urgent",
		ActionURL:  "/customers",
		ActionText: "View Customer",
		ExpiresIn:  168, // 7 days
		Data:       fmt.Sprintf(`{"customer_name": "%s", "amount": %.2f, "overdue_days": %d}`, customerName, amount, overdueDays),
	}

	_, err := s.CreateNotification(ctx, organizationID, req)
	return err
}

// SendWarrantyExpiringNotification sends warranty expiring notification
func (s *Service) SendWarrantyExpiringNotification(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, productName string, serialNumber string, daysRemaining int) error {
	req := &NotificationRequest{
		UserID:     userID,
		Type:       "warranty_expiring",
		Title:      "Warranty Expiring Soon",
		Message:    fmt.Sprintf("%s (%s) warranty expires in %d days", productName, serialNumber, daysRemaining),
		Priority:   "medium",
		ActionURL:  "/warranties",
		ActionText: "View Warranty",
		ExpiresIn:  daysRemaining * 24, // Expire when warranty expires
		Data:       fmt.Sprintf(`{"product_name": "%s", "serial_number": "%s", "days_remaining": %d}`, productName, serialNumber, daysRemaining),
	}

	_, err := s.CreateNotification(ctx, organizationID, req)
	return err
}

// SendReturnRequestNotification sends return request notification
func (s *Service) SendReturnRequestNotification(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, customerName string, productName string) error {
	req := &NotificationRequest{
		UserID:     userID,
		Type:       "return_request",
		Title:      "New Return Request",
		Message:    fmt.Sprintf("%s has requested to return %s", customerName, productName),
		Priority:   "high",
		ActionURL:  "/returns",
		ActionText: "View Return",
		ExpiresIn:  168, // 7 days
		Data:       fmt.Sprintf(`{"customer_name": "%s", "product_name": "%s"}`, customerName, productName),
	}

	_, err := s.CreateNotification(ctx, organizationID, req)
	return err
}

// SendExpenseApprovalNotification sends expense approval notification
func (s *Service) SendExpenseApprovalNotification(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, expenseTitle string, amount float64) error {
	req := &NotificationRequest{
		UserID:     userID,
		Type:       "expense_approval",
		Title:      "Expense Approval Required",
		Message:    fmt.Sprintf("Expense '%s' (%.2f) requires approval", expenseTitle, amount),
		Priority:   "medium",
		ActionURL:  "/expenses",
		ActionText: "Review Expense",
		ExpiresIn:  168, // 7 days
		Data:       fmt.Sprintf(`{"expense_title": "%s", "amount": %.2f}`, expenseTitle, amount),
	}

	_, err := s.CreateNotification(ctx, organizationID, req)
	return err
}

// GetUnreadCount retrieves the count of unread notifications for a user
func (s *Service) GetUnreadCount(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID, organizationID)
}
