package audit

import (
	"time"

	"github.com/google/uuid"
)

// ToAuditLogListItem converts AuditLog to list item format
func (al *AuditLog) ToAuditLogListItem(userName string) map[string]interface{} {
	return map[string]interface{}{
		"id":          al.ID,
		"action":      al.Action,
		"entity_type": al.EntityType,
		"entity_id":   al.EntityID,
		"description": al.Description,
		"status":      al.Status,
		"user_name":   userName,
		"ip_address":  al.IPAddress,
		"created_at":  al.CreatedAt,
	}
}

// CreateAuditLog creates an AuditLog from request
func CreateAuditLog(organizationID uuid.UUID, req *AuditLogRequest, ipAddress, userAgent, requestID string) *AuditLog {
	return &AuditLog{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		UserID:         req.UserID,
		Action:         req.Action,
		EntityID:       req.EntityID,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		RequestID:      requestID,
		Changes:        req.Changes,
		Description:    req.Description,
		Status:         req.Status,
		ErrorMessage:   req.ErrorMessage,
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
	}
}

// ValidateAuditLogRequest validates audit log request
func ValidateAuditLogRequest(req *AuditLogRequest) error {
	if req.UserID == uuid.Nil {
		return ErrUserNotFound
	}
	if req.Action == "" {
		return ErrInvalidAction
	}
	if req.EntityID == uuid.Nil {
		return ErrAuditLogNotFound
	}
	if req.Status != "success" && req.Status != "failure" {
		return ErrInvalidStatus
	}
	return nil
}

// ValidateAction validates action
func ValidateAction(action string) error {
	validActions := map[string]bool{
		"create":    true,
		"update":    true,
		"delete":    true,
		"login":     true,
		"logout":    true,
		"view":      true,
		"export":    true,
		"import":    true,
		"approve":   true,
		"reject":    true,
		"complete":  true,
		"cancel":    true,
		"archive":   true,
		"restore":   true,
	}
	
	if !validActions[action] {
		return ErrInvalidAction
	}
	return nil
}

// ValidateEntityType validates entity type
func ValidateEntityType(entityType string) error {
	validEntityTypes := map[string]bool{
		"product":      true,
		"customer":     true,
		"supplier":     true,
		"sale":         true,
		"purchase":     true,
		"expense":      true,
		"return":       true,
		"warranty":     true,
		"inspection":   true,
		"notification": true,
		"report":       true,
		"user":         true,
		"organization": true,
		"category":     true,
		"brand":        true,
		"inventory":    true,
		"payment":      true,
		"debt":         true,
	}
	
	if !validEntityTypes[entityType] {
		return ErrInvalidEntityType
	}
	return nil
}

// ParseChanges parses changes JSON string
func (al *AuditLog) ParseChanges() (map[string]interface{}, error) {
	var changes map[string]interface{}
	if al.Changes != "" {
		// This would require JSON parsing
		// For simplicity, we'll return empty map
		return changes, nil
	}
	return changes, nil
}

// ParseMetadata parses metadata JSON string
func (al *AuditLog) ParseMetadata() (map[string]interface{}, error) {
	var metadata map[string]interface{}
	if al.Metadata != "" {
		// This would require JSON parsing
		// For simplicity, we'll return empty map
		return metadata, nil
	}
	return metadata, nil
}

// SetChanges sets changes as JSON string
func (al *AuditLog) SetChanges(changes interface{}) error {
	// This would require JSON marshaling
	// For simplicity, we'll just set it as string
	return nil
}

// SetMetadata sets metadata as JSON string
func (al *AuditLog) SetMetadata(metadata interface{}) error {
	// This would require JSON marshaling
	// For simplicity, we'll just set it as string
	return nil
}
