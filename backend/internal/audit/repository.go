package audit

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
	query := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, entity_id, 
			ip_address, user_agent, request_id, changes, description, status, error_message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		auditLog.ID, auditLog.OrganizationID, auditLog.UserID, auditLog.Action, auditLog.EntityType,
		auditLog.EntityID, auditLog.IPAddress, auditLog.UserAgent, auditLog.RequestID, auditLog.Changes,
		auditLog.Description, auditLog.Status, auditLog.ErrorMessage, auditLog.Metadata,
		auditLog.CreatedAt,
	).Scan(&auditLog.ID, &auditLog.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// GetAuditLogByID retrieves an audit log by ID
func (r *Repository) GetAuditLogByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*AuditLog, error) {
	var auditLog AuditLog
	query := `
		SELECT id, organization_id, user_id, action, entity_type, entity_id, 
			ip_address, user_agent, request_id, changes, description, status, error_message, metadata, created_at
		FROM audit_logs
		WHERE id = $1 AND organization_id = $2
	`
	
	err := r.db.GetContext(ctx, &auditLog, query, id, organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAuditLogNotFound
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}
	return &auditLog, nil
}

// ListAuditLogs retrieves audit logs with pagination and filters
func (r *Repository) ListAuditLogs(ctx context.Context, organizationID uuid.UUID, req AuditLogListRequest) ([]AuditLog, int, error) {
	var auditLogs []AuditLog
	var count int
	
	// Build base query
	baseQuery := `
		SELECT id, organization_id, user_id, action, entity_type, entity_id, 
			ip_address, user_agent, request_id, changes, description, status, error_message, metadata, created_at
		FROM audit_logs
		WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
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
		baseQuery += fmt.Sprintf(" AND (description ILIKE $%d OR error_message ILIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (description ILIKE $%d OR error_message ILIKE $%d)", argCount, argCount)
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
	
	err = r.db.SelectContext(ctx, &auditLogs, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}
	
	return auditLogs, count, nil
}

// GetAuditLogSummary retrieves audit log summary statistics
func (r *Repository) GetAuditLogSummary(ctx context.Context, organizationID uuid.UUID) (*AuditLogSummary, error) {
	var summary AuditLogSummary
	
	// Total logs
	err := r.db.GetContext(ctx, &summary.TotalLogs, 
		`SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total logs: %w", err)
	}
	
	// Success logs
	err = r.db.GetContext(ctx, &summary.SuccessLogs, 
		`SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1 AND status = 'success'`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get success logs: %w", err)
	}
	
	// Failure logs
	err = r.db.GetContext(ctx, &summary.FailureLogs, 
		`SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1 AND status = 'failure'`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get failure logs: %w", err)
	}
	
	// This week logs
	err = r.db.GetContext(ctx, &summary.ThisWeek, 
		`SELECT COUNT(*) FROM audit_logs 
		 WHERE organization_id = $1 
		 AND created_at >= DATE_TRUNC('week', CURRENT_DATE)`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get this week logs: %w", err)
	}
	
	// This month logs
	err = r.db.GetContext(ctx, &summary.ThisMonth, 
		`SELECT COUNT(*) FROM audit_logs 
		 WHERE organization_id = $1 
		 AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', CURRENT_DATE)`, 
		organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get this month logs: %w", err)
	}
	
	// By action
	summary.ByAction = make(map[string]int)
	rows, err := r.db.QueryContext(ctx, 
		`SELECT action, COUNT(*) FROM audit_logs WHERE organization_id = $1 GROUP BY action`, 
		organizationID)
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
	
	// By entity type
	summary.ByEntityType = make(map[string]int)
	rows, err = r.db.QueryContext(ctx, 
		`SELECT entity_type, COUNT(*) FROM audit_logs WHERE organization_id = $1 GROUP BY entity_type`, 
		organizationID)
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
	
	// By user
	summary.ByUser = make(map[string]int)
	rows, err = r.db.QueryContext(ctx, 
		`SELECT user_id, COUNT(*) FROM audit_logs WHERE organization_id = $1 GROUP BY user_id`, 
		organizationID)
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
	
	// Recent activity (last 10 entries)
	summary.RecentActivity = []AuditLogEntry{}
	rows, err = r.db.QueryContext(ctx, 
		`SELECT al.id, al.action, al.entity_type, al.entity_id, al.description, al.status, 
			al.user_id, u.first_name || ' ' || u.last_name as user_name, al.created_at
		 FROM audit_logs al
		 LEFT JOIN users u ON al.user_id = u.id
		 WHERE al.organization_id = $1
		 ORDER BY al.created_at DESC
		 LIMIT 10`,
		organizationID)
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
