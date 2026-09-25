BEGIN;

-- Retain an audit-safe snapshot when an operational row is removed for cleanup.
-- These snapshots intentionally have no foreign keys to operational tables so
-- later cleanup of the source rows cannot create broken references.
CREATE TABLE IF NOT EXISTS deleted_operation_snapshots (
    entity_type VARCHAR(80) NOT NULL,
    operation_id VARCHAR(100) NOT NULL,
    business_number VARCHAR(120),
    status VARCHAR(40) NOT NULL,
    snapshot JSONB NOT NULL,
    deleted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (entity_type, operation_id)
);

CREATE INDEX IF NOT EXISTS idx_deleted_operation_snapshots_deleted_at
    ON deleted_operation_snapshots(deleted_at);

COMMIT;
