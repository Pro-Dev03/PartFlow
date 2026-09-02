package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Repository handles notification data operations
type Repository struct {
	db *sqlx.DB
}

type localNotificationRow struct {
	ID         string         `db:"id"`
	UserID     string         `db:"user_id"`
	Type       string         `db:"type"`
	Title      string         `db:"title"`
	Message    string         `db:"message"`
	Data       string         `db:"data"`
	Priority   string         `db:"priority"`
	Status     string         `db:"status"`
	ActionURL  string         `db:"action_url"`
	ActionText string         `db:"action_text"`
	CreatedAt  string         `db:"created_at"`
	UpdatedAt  string         `db:"updated_at"`
	ExpiresAt  sql.NullString `db:"expires_at"`
	ReadAt     sql.NullString `db:"read_at"`
}

func (r localNotificationRow) model() (*Notification, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(r.UserID)
	if err != nil {
		return nil, err
	}
	created, err := dbutil.ParseTimestamp(r.CreatedAt)
	if err != nil {
		return nil, err
	}
	updated, err := dbutil.ParseTimestamp(r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	n := &Notification{ID: id, UserID: uid, Type: r.Type, Title: r.Title, Message: r.Message, Data: r.Data, Priority: r.Priority, Status: r.Status, ActionURL: r.ActionURL, ActionText: r.ActionText, CreatedAt: created, UpdatedAt: updated}
	if r.ExpiresAt.Valid {
		t, e := dbutil.ParseTimestamp(r.ExpiresAt.String)
		if e != nil {
			return nil, e
		}
		n.ExpiresAt = &t
	}
	if r.ReadAt.Valid {
		t, e := dbutil.ParseTimestamp(r.ReadAt.String)
		if e != nil {
			return nil, e
		}
		n.ReadAt = &t
	}
	return n, nil
}

// NewRepository creates a new notification repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateNotification creates a new notification
func (r *Repository) CreateNotification(ctx context.Context, notification *Notification) error {
	if dbutil.IsSQLite(r.db) {
		if notification.ID == uuid.Nil {
			notification.ID = uuid.New()
		}
		if notification.CreatedAt.IsZero() {
			notification.CreatedAt = time.Now().UTC()
		}
		if notification.UpdatedAt.IsZero() {
			notification.UpdatedAt = notification.CreatedAt
		}
		_, err := r.db.ExecContext(ctx, `INSERT INTO notifications (id,user_id,type,title,message,data,priority,status,action_url,action_text,expires_at,created_at,updated_at,read_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, notification.ID.String(), notification.UserID.String(), notification.Type, notification.Title, notification.Message, notification.Data, notification.Priority, notification.Status, notification.ActionURL, notification.ActionText, notification.ExpiresAt, notification.CreatedAt.Format(time.RFC3339Nano), notification.UpdatedAt.Format(time.RFC3339Nano), notification.ReadAt)
		if err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
		return nil
	}
	query := `
		INSERT INTO notifications (user_id, type, title, message, data, priority, status, action_url, action_text, expires_at, created_at, updated_at, read_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		notification.UserID, notification.Type, notification.Title, notification.Message, notification.Data,
		notification.Priority, notification.Status, notification.ActionURL, notification.ActionText,
		notification.ExpiresAt, notification.CreatedAt, notification.UpdatedAt, notification.ReadAt,
	).Scan(&notification.ID, &notification.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

// GetNotificationByID retrieves a notification by ID
func (r *Repository) GetNotificationByID(ctx context.Context, id uuid.UUID) (*Notification, error) {
	if dbutil.IsSQLite(r.db) {
		var row localNotificationRow
		err := r.db.GetContext(ctx, &row, `SELECT id,user_id,type,title,message,COALESCE(data,'{}') AS data,priority,status,COALESCE(action_url,'') AS action_url,COALESCE(action_text,'') AS action_text,created_at,updated_at,expires_at,read_at FROM notifications WHERE id = ?`, id.String())
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrNotificationNotFound
			}
			return nil, fmt.Errorf("failed to get notification: %w", err)
		}
		n, err := row.model()
		if err != nil {
			return nil, fmt.Errorf("failed to parse notification: %w", err)
		}
		return n, nil
	}
	var notification Notification
	query := `
		SELECT id, user_id, type, title, message, COALESCE(CAST(data AS TEXT), '{}') AS data, priority, status,
			COALESCE(action_url, '') AS action_url, COALESCE(action_text, '') AS action_text,
			expires_at, created_at, read_at
		FROM notifications
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &notification, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	return &notification, nil
}

// ListNotifications retrieves notifications with pagination and filters
func (r *Repository) ListNotifications(ctx context.Context, userID uuid.UUID, req NotificationListRequest) ([]Notification, int, error) {
	if dbutil.IsSQLite(r.db) {
		if req.Page < 1 {
			req.Page = 1
		}
		if req.PerPage < 1 {
			req.PerPage = 20
		}
		where := " WHERE user_id = ? AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)"
		args := []interface{}{userID.String()}
		if req.Type != "" {
			where += " AND type = ?"
			args = append(args, req.Type)
		}
		if req.Status != "" {
			where += " AND status = ?"
			args = append(args, req.Status)
		}
		if req.Priority != "" {
			where += " AND priority = ?"
			args = append(args, req.Priority)
		}
		if req.StartDate != nil {
			where += " AND created_at >= ?"
			args = append(args, req.StartDate.Format(time.RFC3339Nano))
		}
		if req.EndDate != nil {
			where += " AND created_at <= ?"
			args = append(args, req.EndDate.Format(time.RFC3339Nano))
		}
		var count int
		countQuery := "SELECT COUNT(*) FROM notifications" + where
		if err := r.db.GetContext(ctx, &count, r.db.Rebind(countQuery), args...); err != nil {
			return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
		}
		sortBy := "created_at"
		for _, allowed := range []string{"created_at", "priority", "status", "type"} {
			if req.SortBy == allowed {
				sortBy = allowed
			}
		}
		sortOrder := "DESC"
		if req.SortOrder == "ASC" {
			sortOrder = "ASC"
		}
		query := "SELECT id,user_id,type,title,message,COALESCE(data,'{}') AS data,priority,status,COALESCE(action_url,'') AS action_url,COALESCE(action_text,'') AS action_text,created_at,updated_at,expires_at,read_at FROM notifications" + where + " ORDER BY " + sortBy + " " + sortOrder + " LIMIT ? OFFSET ?"
		args = append(args, req.PerPage, (req.Page-1)*req.PerPage)
		var rows []localNotificationRow
		if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
			return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
		}
		notifications := make([]Notification, 0, len(rows))
		for _, row := range rows {
			n, err := row.model()
			if err != nil {
				return nil, 0, fmt.Errorf("failed to parse notification: %w", err)
			}
			notifications = append(notifications, *n)
		}
		return notifications, count, nil
	}
	var notifications []Notification
	var count int

	// Build base query
	baseQuery := `
		SELECT id, user_id, type, title, message, COALESCE(CAST(data AS TEXT), '{}') AS data, priority, status,
			COALESCE(action_url, '') AS action_url, COALESCE(action_text, '') AS action_text,
			expires_at, created_at, read_at
		FROM notifications
		WHERE user_id = $1
	`

	countQuery := `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1
	`

	args := []interface{}{userID}
	argCount := 1

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
	countQuery += " AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)"
	args = append(args, time.Now())

	// Get total count
	countArgs := args[:len(args)-1]
	err := r.db.GetContext(ctx, &count, countQuery, countArgs...)
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
	if dbutil.IsSQLite(r.db) {
		now := time.Now().UTC()
		result, err := r.db.ExecContext(ctx, `UPDATE notifications SET status=?, read_at=?, updated_at=? WHERE id=?`, notification.Status, notification.ReadAt, now.Format(time.RFC3339Nano), notification.ID.String())
		if err != nil {
			return fmt.Errorf("failed to update notification: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return ErrNotificationNotFound
		}
		notification.UpdatedAt = now
		return nil
	}
	query := `
		UPDATE notifications
		SET status = $2, read_at = $3
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		notification.ID, notification.Status, notification.ReadAt,
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
func (r *Repository) DeleteNotification(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	arg := interface{}(id)
	if dbutil.IsSQLite(r.db) {
		query = `DELETE FROM notifications WHERE id = ?`
		arg = id.String()
	}
	result, err := r.db.ExecContext(ctx, query, arg)
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
func (r *Repository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if dbutil.IsSQLite(r.db) {
		result, err := r.db.ExecContext(ctx, `UPDATE notifications SET status='read', read_at=?, updated_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano), id.String())
		if err != nil {
			return fmt.Errorf("failed to mark notification as read: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return ErrNotificationNotFound
		}
		return nil
	}
	query := `
		UPDATE notifications
		SET status = 'read', read_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING read_at
	`

	var readAt time.Time
	err := r.db.QueryRowContext(ctx, query, id).Scan(&readAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotificationNotFound
		}
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *Repository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET status = 'read', read_at = CURRENT_TIMESTAMP
		WHERE user_id = $1
	`

	arg := interface{}(userID)
	if dbutil.IsSQLite(r.db) {
		query = `UPDATE notifications SET status = 'read', read_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?`
		arg = userID.String()
	}
	_, err := r.db.ExecContext(ctx, query, arg)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

// GetNotificationSummary retrieves notification summary for a user
func (r *Repository) GetNotificationSummary(ctx context.Context, userID uuid.UUID) (*NotificationSummary, error) {
	var summary NotificationSummary

	// Total notifications
	err := r.db.GetContext(ctx, &summary.Total,
		`SELECT COUNT(*) FROM notifications
		 WHERE user_id = $1
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total notifications: %w", err)
	}

	// Unread notifications
	err = r.db.GetContext(ctx, &summary.Unread,
		`SELECT COUNT(*) FROM notifications
		 WHERE user_id = $1 AND status = 'unread'
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}

	// Priority notifications
	err = r.db.GetContext(ctx, &summary.Priority,
		`SELECT COUNT(*) FROM notifications
		 WHERE user_id = $1 AND priority IN ('high', 'urgent')
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get priority notifications: %w", err)
	}

	// Urgent notifications
	err = r.db.GetContext(ctx, &summary.Urgent,
		`SELECT COUNT(*) FROM notifications
		 WHERE user_id = $1 AND priority = 'urgent'
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get urgent notifications: %w", err)
	}

	// By type
	summary.ByType = make(map[string]int)
	rows, err := r.db.QueryContext(ctx,
		`SELECT type, COUNT(*) FROM notifications
		 WHERE user_id = $1
		 AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		 GROUP BY type`,
		userID)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate notifications by type: %w", err)
	}

	return &summary, nil
}

// CreateNotificationPreferences creates notification preferences for a user
func (r *Repository) CreateNotificationPreferences(ctx context.Context, preferences *NotificationPreferences) error {
	if dbutil.IsSQLite(r.db) {
		if preferences.ID == uuid.Nil {
			preferences.ID = uuid.New()
		}
		if preferences.CreatedAt.IsZero() {
			preferences.CreatedAt = time.Now().UTC()
		}
		if preferences.UpdatedAt.IsZero() {
			preferences.UpdatedAt = preferences.CreatedAt
		}
		_, err := r.db.ExecContext(ctx, `INSERT INTO notification_preferences (id,user_id,email_enabled,push_enabled,low_stock,debt_overdue,return_requests,expense_approval,sales_updates,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, preferences.ID.String(), preferences.UserID.String(), preferences.EmailEnabled, preferences.PushEnabled, preferences.LowStock, preferences.DebtOverdue, preferences.ReturnRequests, preferences.ExpenseApproval, preferences.SalesUpdates, preferences.CreatedAt.Format(time.RFC3339Nano), preferences.UpdatedAt.Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("failed to create notification preferences: %w", err)
		}
		return nil
	}
	query := `
		INSERT INTO notification_preferences (user_id, email_enabled, push_enabled, low_stock, debt_overdue, return_requests, expense_approval, sales_updates, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		preferences.UserID, preferences.EmailEnabled, preferences.PushEnabled, preferences.LowStock,
		preferences.DebtOverdue, preferences.ReturnRequests, preferences.ExpenseApproval,
		preferences.SalesUpdates, preferences.CreatedAt, preferences.UpdatedAt,
	).Scan(&preferences.ID, &preferences.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create notification preferences: %w", err)
	}
	return nil
}

// GetNotificationPreferences retrieves notification preferences for a user
func (r *Repository) GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error) {
	if dbutil.IsSQLite(r.db) {
		var row struct {
			ID              string `db:"id"`
			UserID          string `db:"user_id"`
			EmailEnabled    bool   `db:"email_enabled"`
			PushEnabled     bool   `db:"push_enabled"`
			LowStock        bool   `db:"low_stock"`
			DebtOverdue     bool   `db:"debt_overdue"`
			ReturnRequests  bool   `db:"return_requests"`
			ExpenseApproval bool   `db:"expense_approval"`
			SalesUpdates    bool   `db:"sales_updates"`
			CreatedAt       string `db:"created_at"`
			UpdatedAt       string `db:"updated_at"`
		}
		err := r.db.GetContext(ctx, &row, `SELECT id,user_id,email_enabled,push_enabled,low_stock,debt_overdue,return_requests,expense_approval,sales_updates,created_at,updated_at FROM notification_preferences WHERE user_id = ?`, userID.String())
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrPreferencesNotFound
			}
			return nil, fmt.Errorf("failed to get notification preferences: %w", err)
		}
		id, e := uuid.Parse(row.ID)
		if e != nil {
			return nil, e
		}
		uid, e := uuid.Parse(row.UserID)
		if e != nil {
			return nil, e
		}
		created, e := dbutil.ParseTimestamp(row.CreatedAt)
		if e != nil {
			return nil, e
		}
		updated, e := dbutil.ParseTimestamp(row.UpdatedAt)
		if e != nil {
			return nil, e
		}
		return &NotificationPreferences{ID: id, UserID: uid, EmailEnabled: row.EmailEnabled, PushEnabled: row.PushEnabled, LowStock: row.LowStock, DebtOverdue: row.DebtOverdue, ReturnRequests: row.ReturnRequests, ExpenseApproval: row.ExpenseApproval, SalesUpdates: row.SalesUpdates, CreatedAt: created, UpdatedAt: updated}, nil
	}
	var preferences NotificationPreferences
	query := `
		SELECT id, user_id, email_enabled, push_enabled, low_stock, debt_overdue, return_requests, expense_approval, sales_updates, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	err := r.db.GetContext(ctx, &preferences, query, userID)
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
	if dbutil.IsSQLite(r.db) {
		now := time.Now().UTC()
		result, err := r.db.ExecContext(ctx, `UPDATE notification_preferences SET email_enabled=?,push_enabled=?,low_stock=?,debt_overdue=?,return_requests=?,expense_approval=?,sales_updates=?,updated_at=? WHERE id=?`, preferences.EmailEnabled, preferences.PushEnabled, preferences.LowStock, preferences.DebtOverdue, preferences.ReturnRequests, preferences.ExpenseApproval, preferences.SalesUpdates, now.Format(time.RFC3339Nano), preferences.ID.String())
		if err != nil {
			return fmt.Errorf("failed to update notification preferences: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return ErrPreferencesNotFound
		}
		preferences.UpdatedAt = now
		return nil
	}
	query := `
		UPDATE notification_preferences
		SET email_enabled = $2, push_enabled = $3, low_stock = $4, debt_overdue = $5,
			return_requests = $6, expense_approval = $7, sales_updates = $8, updated_at = $9
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		preferences.ID, preferences.EmailEnabled, preferences.PushEnabled, preferences.LowStock,
		preferences.DebtOverdue, preferences.ReturnRequests,
		preferences.ExpenseApproval, preferences.SalesUpdates, preferences.UpdatedAt,
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
func (r *Repository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	var err error

	if dbutil.IsSQLite(r.db) {
		err = r.db.GetContext(ctx, &count,
			`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND status = 'unread' AND (expires_at IS NULL OR expires_at > datetime('now'))`,
			userID)
	} else {
		err = r.db.GetContext(ctx, &count,
			`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND status = 'unread' AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)`,
			userID)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}
