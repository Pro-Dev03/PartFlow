package audit

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Repository handles audit log data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new audit log repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateAuditLog creates a new audit log entry
func (r *Repository) CreateAuditLog(ctx context.Context, auditLog *AuditLog) error {
	if auditLog.ID == uuid.Nil {
		auditLog.ID = uuid.New()
	}
	auditLog.CreatedAt = auditLog.CreatedAt.Round(0)
	query := `
		INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id,
			ip_address, user_agent, request_id, changes, description, status, error_message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.ExecContext(ctx, query,
		auditLog.ID, auditLog.UserID, auditLog.Action, auditLog.EntityType, auditLog.EntityID,
		auditLog.IPAddress, auditLog.UserAgent, auditLog.RequestID, auditLog.Changes,
		auditLog.Description, auditLog.Status, auditLog.ErrorMessage, auditLog.Metadata,
		auditLog.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// GetAuditLogByID retrieves an audit log by ID
func (r *Repository) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*AuditLog, error) {
	var auditLog AuditLog
	query := `
		SELECT id, COALESCE(NULLIF(user_id, ''), '00000000-0000-0000-0000-000000000000') AS user_id, action, entity_type,
			COALESCE(NULLIF(entity_id, ''), '00000000-0000-0000-0000-000000000000') AS entity_id,
			COALESCE(ip_address, '') AS ip_address, COALESCE(user_agent, '') AS user_agent, COALESCE(request_id, '') AS request_id, COALESCE(changes, '') AS changes, COALESCE(new_values, changes, '') AS new_values, COALESCE(description, '') AS description, COALESCE(status, 'success') AS status, COALESCE(error_message, '') AS error_message, COALESCE(metadata, '{}') AS metadata, created_at
		FROM audit_logs
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &auditLog, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAuditLogNotFound
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}
	return &auditLog, nil
}

// ListAuditLogs retrieves audit logs with pagination and filters
func (r *Repository) ListAuditLogs(ctx context.Context, req AuditLogListRequest) ([]AuditLog, int, error) {
	var auditLogs []AuditLog
	var count int

	// Build base query
	baseQuery := `
		SELECT id, COALESCE(NULLIF(user_id, ''), '00000000-0000-0000-0000-000000000000') AS user_id, action, entity_type,
			COALESCE(NULLIF(entity_id, ''), '00000000-0000-0000-0000-000000000000') AS entity_id,
			COALESCE(ip_address, '') AS ip_address, COALESCE(user_agent, '') AS user_agent, COALESCE(request_id, '') AS request_id, COALESCE(changes, '') AS changes, COALESCE(new_values, changes, '') AS new_values, COALESCE(description, '') AS description, COALESCE(status, 'success') AS status, COALESCE(error_message, '') AS error_message, COALESCE(metadata, '{}') AS metadata, created_at
		FROM audit_logs
		WHERE 1=1
	`

	countQuery := `
		SELECT COUNT(*)
		FROM audit_logs
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if req.UserID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *req.UserID)
	}

	if req.Action != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND action = $%d", argCount)
		countQuery += fmt.Sprintf(" AND action = $%d", argCount)
		args = append(args, req.Action)
	}

	if req.EntityID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND entity_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND entity_id = $%d", argCount)
		args = append(args, *req.EntityID)
	}

	if req.EntityType != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND entity_type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND entity_type = $%d", argCount)
		args = append(args, req.EntityType)
	}

	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
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

	if req.Search != "" {
		argCount++
		like := "ILIKE"
		if dbutil.IsSQLite(r.db) {
			like = "LIKE"
		}
		baseQuery += fmt.Sprintf(" AND (description %s $%d OR error_message %s $%d)", like, argCount, like, argCount)
		countQuery += fmt.Sprintf(" AND (description %s $%d OR error_message %s $%d)", like, argCount, like, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
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

	rows, err := r.db.QueryxContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		record := map[string]any{}
		if err := rows.MapScan(record); err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}
		log, err := auditLogFromRecord(record)
		if err != nil {
			return nil, 0, err
		}
		auditLogs = append(auditLogs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate audit logs: %w", err)
	}

	return auditLogs, count, nil
}

func auditLogFromRecord(record map[string]any) (AuditLog, error) {
	parseID := func(value any) uuid.UUID {
		id, err := uuid.Parse(strings.TrimSpace(fmt.Sprint(value)))
		if err != nil {
			return uuid.Nil
		}
		return id
	}
	valueString := func(value any) string {
		if value == nil {
			return ""
		}
		return fmt.Sprint(value)
	}
	createdAt, err := dbutil.ParseTimestamp(record["created_at"])
	if err != nil {
		return AuditLog{}, fmt.Errorf("failed to parse audit timestamp: %w", err)
	}
	return AuditLog{
		ID:           parseID(record["id"]),
		UserID:       parseID(record["user_id"]),
		Action:       valueString(record["action"]),
		EntityType:   valueString(record["entity_type"]),
		EntityID:     parseID(record["entity_id"]),
		IPAddress:    valueString(record["ip_address"]),
		UserAgent:    valueString(record["user_agent"]),
		RequestID:    valueString(record["request_id"]),
		Changes:      valueString(record["changes"]),
		NewValues:    valueString(record["new_values"]),
		Description:  valueString(record["description"]),
		Status:       valueString(record["status"]),
		ErrorMessage: valueString(record["error_message"]),
		Metadata:     valueString(record["metadata"]),
		CreatedAt:    createdAt,
	}, nil
}

// GetAuditLogSummary retrieves audit log summary statistics
func (r *Repository) GetAuditLogSummary(ctx context.Context) (*AuditLogSummary, error) {
	var summary AuditLogSummary

	// Total logs
	err := r.db.GetContext(ctx, &summary.TotalLogs,
		`SELECT COUNT(*) FROM audit_logs`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total logs: %w", err)
	}

	// Success logs
	err = r.db.GetContext(ctx, &summary.SuccessLogs,
		`SELECT COUNT(*) FROM audit_logs WHERE status = 'success'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get success logs: %w", err)
	}

	// Failure logs
	err = r.db.GetContext(ctx, &summary.FailureLogs,
		`SELECT COUNT(*) FROM audit_logs WHERE status = 'failure'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get failure logs: %w", err)
	}

	// This week logs
	weekStart, _, err := accounting.StoreWeekBounds(accounting.StoreNow())
	if err != nil {
		return nil, fmt.Errorf("calculate store week boundary: %w", err)
	}
	monthStart, monthEnd, err := accounting.StoreMonthBounds(accounting.StoreNow())
	if err != nil {
		return nil, fmt.Errorf("calculate store month boundary: %w", err)
	}
	weekQuery := `SELECT COUNT(*) FROM audit_logs WHERE created_at >= $1`
	monthQuery := `SELECT COUNT(*) FROM audit_logs WHERE created_at >= $1 AND created_at < $2`
	if dbutil.IsSQLite(r.db) {
		weekQuery = `SELECT COUNT(*) FROM audit_logs WHERE datetime(created_at) >= datetime(?)`
		monthQuery = `SELECT COUNT(*) FROM audit_logs WHERE datetime(created_at) >= datetime(?) AND datetime(created_at) < datetime(?)`
	}
	err = r.db.GetContext(ctx, &summary.ThisWeek, weekQuery, weekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get this week logs: %w", err)
	}

	// This month logs
	err = r.db.GetContext(ctx, &summary.ThisMonth, monthQuery, monthStart, monthEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month logs: %w", err)
	}

	// By action
	summary.ByAction = make(map[string]int)
	rows, err := r.db.QueryContext(ctx,
		`SELECT action, COUNT(*) FROM audit_logs GROUP BY action`)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs by action: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			continue
		}
		summary.ByAction[action] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate logs by action: %w", err)
	}

	// By entity type
	summary.ByEntityType = make(map[string]int)
	rows, err = r.db.QueryContext(ctx,
		`SELECT entity_type, COUNT(*) FROM audit_logs GROUP BY entity_type`)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs by entity type: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entityType string
		var count int
		if err := rows.Scan(&entityType, &count); err != nil {
			continue
		}
		summary.ByEntityType[entityType] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate logs by entity type: %w", err)
	}

	// By user
	summary.ByUser = make(map[string]int)
	rows, err = r.db.QueryContext(ctx,
		`SELECT user_id, COUNT(*) FROM audit_logs GROUP BY user_id`)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs by user: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID uuid.UUID
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			continue
		}
		summary.ByUser[userID.String()] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate logs by user: %w", err)
	}

	// Recent activity (last 10 entries)
	summary.RecentActivity = []AuditLogEntry{}
	rows, err = r.db.QueryContext(ctx,
		`SELECT al.id, al.action, al.entity_type, al.entity_id, al.description, al.status,
			al.user_id, u.first_name || ' ' || u.last_name as user_name, al.created_at
		 FROM audit_logs al
		 LEFT JOIN users u ON al.user_id = u.id
		 ORDER BY al.created_at DESC
		 LIMIT 10`)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entry AuditLogEntry
		if err := rows.Scan(&entry.ID, &entry.Action, &entry.EntityType, &entry.EntityID,
			&entry.Description, &entry.Status, &entry.UserID, &entry.UserName, &entry.CreatedAt); err != nil {
			continue
		}
		summary.RecentActivity = append(summary.RecentActivity, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate recent activity: %w", err)
	}

	return &summary, nil
}

// GetUserName retrieves user name by ID
func (r *Repository) GetUserName(ctx context.Context, userID uuid.UUID) (string, error) {
	var name string
	query := `SELECT first_name || ' ' || last_name as name FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &name, query, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user name: %w", err)
	}
	return name, nil
}
