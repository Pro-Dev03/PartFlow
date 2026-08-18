package audit

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service handles audit log business logic
type Service struct {
	repo *Repository
}

// NewService creates a new audit log service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateAuditLog creates a new audit log entry
func (s *Service) CreateAuditLog(ctx context.Context, organizationID uuid.UUID, req *AuditLogRequest, ipAddress, userAgent, requestID string) (*AuditLog, error) {
	// Validate request
	if err := ValidateAuditLogRequest(req); err != nil {
		return nil, err
	}

	// Create audit log
	auditLog := CreateAuditLog(organizationID, req, ipAddress, userAgent, requestID)

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return auditLog, nil
}

// GetAuditLog retrieves an audit log by ID
func (s *Service) GetAuditLog(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*AuditLog, error) {
	return s.repo.GetAuditLogByID(ctx, id, organizationID)
}

// ListAuditLogs retrieves audit logs with pagination and filters
func (s *Service) ListAuditLogs(ctx context.Context, organizationID uuid.UUID, req AuditLogListRequest) ([]map[string]interface{}, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	auditLogs, total, err := s.repo.ListAuditLogs(ctx, organizationID, req)
	if err != nil {
		return nil, 0, err
	}

	// Convert to list items
	var result []map[string]interface{}
	for _, auditLog := range auditLogs {
		userName, err := s.repo.GetUserName(ctx, auditLog.UserID)
		if err != nil {
			userName = "Unknown"
		}
		result = append(result, auditLog.ToAuditLogListItem(userName))
	}

	return result, total, nil
}

// GetAuditLogSummary retrieves audit log summary statistics
func (s *Service) GetAuditLogSummary(ctx context.Context, organizationID uuid.UUID) (*AuditLogSummary, error) {
	return s.repo.GetAuditLogSummary(ctx, organizationID)
}

// LogAction logs an action with automatic context
func (s *Service) LogAction(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, action, entityType string, entityID uuid.UUID, description string, status string) error {
	req := &AuditLogRequest{
		UserID:      userID,
		Action:      action,
		EntityID:    entityID,
		Description: description,
		Status:      status,
	}

	_, err := s.CreateAuditLog(ctx, organizationID, req, "", "", "")
	return err
}

// LogActionWithChanges logs an action with changes
func (s *Service) LogActionWithChanges(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, action, entityType string, entityID uuid.UUID, description string, status string, changes interface{}) error {
	// Convert changes to JSON string (simplified)
	changesStr := fmt.Sprintf("%v", changes)
	
	req := &AuditLogRequest{
		UserID:      userID,
		Action:      action,
		EntityID:    entityID,
		Changes:     changesStr,
		Description: description,
		Status:      status,
	}

	_, err := s.CreateAuditLog(ctx, organizationID, req, "", "", "")
	return err
}

// LogLogin logs a login action
func (s *Service) LogLogin(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, ipAddress, userAgent string) error {
	req := &AuditLogRequest{
		UserID:      userID,
		Action:      "login",
		EntityID:    userID,
		Description: "User logged in",
		Status:      "success",
	}

	_, err := s.CreateAuditLog(ctx, organizationID, req, ipAddress, userAgent, "")
	return err
}

// LogLogout logs a logout action
func (s *Service) LogLogout(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, ipAddress, userAgent string) error {
	req := &AuditLogRequest{
		UserID:      userID,
		Action:      "logout",
		EntityID:    userID,
		Description: "User logged out",
		Status:      "success",
	}

	_, err := s.CreateAuditLog(ctx, organizationID, req, ipAddress, userAgent, "")
	return err
}

// LogEntityCreate logs an entity creation
func (s *Service) LogEntityCreate(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, entityType string, entityID uuid.UUID, description string) error {
	return s.LogAction(ctx, organizationID, userID, "create", entityType, entityID, description, "success")
}

// LogEntityUpdate logs an entity update
func (s *Service) LogEntityUpdate(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, entityType string, entityID uuid.UUID, description string, changes interface{}) error {
	return s.LogActionWithChanges(ctx, organizationID, userID, "update", entityType, entityID, description, "success", changes)
}

// LogEntityDelete logs an entity deletion
func (s *Service) LogEntityDelete(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, entityType string, entityID uuid.UUID, description string) error {
	return s.LogAction(ctx, organizationID, userID, "delete", entityType, entityID, description, "success")
}

// LogError logs an error
func (s *Service) LogError(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, action, entityType string, entityID uuid.UUID, description, errorMessage string) error {
	req := &AuditLogRequest{
		UserID:      userID,
		Action:      action,
		EntityID:    entityID,
		Description: description,
		Status:      "failure",
		ErrorMessage: errorMessage,
	}

	_, err := s.CreateAuditLog(ctx, organizationID, req, "", "", "")
	return err
}
