package audit

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateAuditLogPreservesEntityType(t *testing.T) {
	req := &AuditLogRequest{
		UserID:      uuid.New(),
		Action:      "create",
		EntityType:  "inventory",
		EntityID:    uuid.New(),
		Description: "inventory item created",
		Status:      "success",
	}

	log := CreateAuditLog(req, "127.0.0.1", "test-agent", "req-1")
	if log.EntityType != "inventory" {
		t.Fatalf("expected entity type to be preserved, got %q", log.EntityType)
	}
}
