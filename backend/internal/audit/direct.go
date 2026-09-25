package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// RecordDirect writes a best-effort audit entry for handlers that own the
// business operation but do not depend on the audit service package.
func RecordDirect(ctx context.Context, db *sqlx.DB, userID uuid.UUID, action, entityType string, entityID uuid.UUID, description string) error {
	query := `INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, description, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.ExecContext(ctx, db.Rebind(query), uuid.New(), userID, action, entityType, entityID, description, "success", time.Now().UTC())
	return err
}
