package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles notification data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new notification repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateNotification creates a new notification
func (r *Repository) CreateNotification(ctx context.Context, notification *Notification) error {
	query := `
		INSERT INTO notifications (id, organization_id, user_id, type, title, message, 
			data, priority, status, action_url, action_text, expires_at, created_at, updated_at, read_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		notification.ID, notification.OrganizationID, notification.UserID, notification.Type,
		notification.Title, notification.Message, notification.Data, notification.Priority,
		notification.Status, notification.ActionURL, notification.ActionText, notification.ExpiresAt,
		notification.CreatedAt, notification.UpdatedAt, notification.ReadAt,
	).Scan(&notification.ID, &notification.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

// GetNotificationByID retrieves a notification by ID
func (r *Repository) GetNotificationByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Notification, error) {
	var notification Notification
	query := `
		SELECT id, organization_id, user_id, type, title, message, 
			data, priority, status, action_url, action_text, expires_at, created_at, read_at
		FROM notifications
		WHERE id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &notification, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	return &notification, nil
}

// ListNotifications retrieves notifications with pagination and filters
func (r *Repository) ListNotifications(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req NotificationListRequest) ([]Notification, int, error) {
	var notifications []Notification
	var count int
	
	// Build base query
	baseQuery := `
		SELECT id, organization_id, user_id, type, title, message, 
			data, priority, status, action_url, action_text, expires_at, created_at, read_at
		FROM notifications
		WHERE organization_id = $1 AND user_id = $2
	`
	countQuery := `SELECT COUNT(*) FROM notifications WHERE organization_id = $1 AND user_id = $2`
	
	args := []interface{}{organizationID, userID}
	argCount := 2
	
	// Add filters
	if req.Type != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, req.Type)
	}
	
	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}
	
	if req.Priority != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND priority = $%d", argCount)
		countQuery += fmt.Sprintf(" AND priority = $%d", argCount)
		args = append(args, req.Priority)
	}
	
	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *req.StartDate)
	}
	
	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *req.EndDate)
	}
	
	// Filter out expired notifications
	argCount++
	baseQuery += fmt.Sprintf(" AND (expires_at IS NULL OR expires_at > $%d)", argCount)
	countQuery += fmt.Sprintf(" AND (expires_at IS NULL OR expires_at > $%d)", argCount)
	args = append(args, time.Now())
	
	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}
	
	// Add sorting
	sortBy := "created_at"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "DESC"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
	
	// Add pagination
	offset := (req.Page - 1) * req.PerPage
	argCount++
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.PerPage, offset)
	
	err = r.db.SelectContext(ctx, &notifications, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}
	
	return notifications, count, nil
}

// UpdateNotification updates a notification
func (r *Repository) UpdateNotification(ctx context.Context, notification *Notification) error {
	query := `
		UPDATE notifications
		SET status = $2, read_at = $3
		WHERE id = $1 AND organization_id = $4
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		notification.ID, notification.Status, notification.ReadAt, notification.OrganizationID,
	).Scan(&notification.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotificationNotFound
		}
		return fmt.Errorf("failed to update notification: %w", err)
	}
	return nil
}

// DeleteNotification deletes a notification
func (r *Repository) DeleteNotification(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1 AND organization_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotificationNotFound
	}
	
	return nil
}

// MarkAsRead marks notification as read
func (r *Repository) MarkAsRead(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET status = 'read', read_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND organization_id = $2
		RETURNING read_at
	`
	
	var readAt time.Time
	err := r.db.QueryRowContext(ctx, query, id, organizationID).Scan(&readAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotificationNotFound
		}
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *Repository) MarkAllAsRead(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET status = 'read', read_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND organization_id = $2 AND status = 'unread'
	`
	
	_, err := r.db.ExecContext(ctx, query, userID, organizationID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

// GetNotificationSummary retrieves notification summary for a user
func (r *Repository) GetNotificationSummary(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (*NotificationSummary, error) {
	var summary NotificationSummary
	
	// Total notifications
	err := r.db.GetContext(ctx, &summary.Total, 
		`SELECT COUNT(*) FROM notifications 
		 WHERE user_id = $1 AND organization_id = $2 
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`,
		userID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total notifications: %w", err)
	}
	
	// Unread notifications
	err = r.db.GetContext(ctx, &summary.Unread, 
		`SELECT COUNT(*) FROM notifications 
		 WHERE user_id = $1 AND organization_id = $2 AND status = 'unread'
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`,
		userID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}
	
	// Priority notifications
	err = r.db.GetContext(ctx, &summary.Priority, 
		`SELECT COUNT(*) FROM notifications 
		 WHERE user_id = $1 AND organization_id = $2 AND priority IN ('high', 'urgent')
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`,
		userID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get priority notifications: %w", err)
	}
	
	// Urgent notifications
	err = r.db.GetContext(ctx, &summary.Urgent, 
		`SELECT COUNT(*) FROM notifications 
		 WHERE user_id = $1 AND organization_id = $2 AND priority = 'urgent'
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`,
		userID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get urgent notifications: %w", err)
	}
	
	// By type
	summary.ByType = make(map[string]int)
	rows, err := r.db.QueryContext(ctx, 
		`SELECT type, COUNT(*) FROM notifications 
		 WHERE user_id = $1 AND organization_id = $2 
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		 GROUP BY type`,
		userID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications by type: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var notifType string
		var count int
		if err := rows.Scan(&notifType, &count); err != nil {
			continue
		}
		summary.ByType[notifType] = count
	}
	
	return &summary, nil
}

// CreateNotificationPreferences creates notification preferences for a user
func (r *Repository) CreateNotificationPreferences(ctx context.Context, preferences *NotificationPreferences) error {
	query := `
		INSERT INTO notification_preferences (id, user_id, organization_id, email_enabled, push_enabled,
			low_stock, debt_overdue, warranty_expiring, return_requests, expense_approval, sales_updates, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		preferences.ID, preferences.UserID, preferences.OrganizationID, preferences.EmailEnabled,
		preferences.PushEnabled, preferences.LowStock, preferences.DebtOverdue, preferences.WarrantyExpiring,
		preferences.ReturnRequests, preferences.ExpenseApproval, preferences.SalesUpdates,
		preferences.CreatedAt,
	).Scan(&preferences.ID, &preferences.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create notification preferences: %w", err)
	}
	return nil
}

// GetNotificationPreferences retrieves notification preferences for a user
func (r *Repository) GetNotificationPreferences(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (*NotificationPreferences, error) {
	var preferences NotificationPreferences
	query := `
		SELECT id, user_id, organization_id, email_enabled, push_enabled,
			low_stock, debt_overdue, warranty_expiring, return_requests, expense_approval, sales_updates, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &preferences, query, userID, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPreferencesNotFound
		}
		return nil, fmt.Errorf("failed to get notification preferences: %w", err)
	}
	return &preferences, nil
}

// UpdateNotificationPreferences updates notification preferences
func (r *Repository) UpdateNotificationPreferences(ctx context.Context, preferences *NotificationPreferences) error {
	query := `
		UPDATE notification_preferences
		SET email_enabled = $2, push_enabled = $3, low_stock = $4, debt_overdue = $5,
			warranty_expiring = $6, return_requests = $7, expense_approval = $8, sales_updates = $9, updated_at = $10
		WHERE id = $1 AND organization_id = $11
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		preferences.ID, preferences.EmailEnabled, preferences.PushEnabled, preferences.LowStock,
		preferences.DebtOverdue, preferences.WarrantyExpiring, preferences.ReturnRequests,
		preferences.ExpenseApproval, preferences.SalesUpdates, preferences.UpdatedAt,
		preferences.OrganizationID,
	).Scan(&preferences.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPreferencesNotFound
		}
		return fmt.Errorf("failed to update notification preferences: %w", err)
	}
	return nil
}

// GetUnreadCount retrieves the count of unread notifications for a user
func (r *Repository) GetUnreadCount(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, 
		`SELECT COUNT(*) FROM notifications 
		 WHERE user_id = $1 AND organization_id = $2 AND status = 'unread'
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`,
		userID, organizationID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}
