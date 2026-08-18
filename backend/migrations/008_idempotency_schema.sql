-- PartFlow Idempotency Schema
-- Support for idempotent API requests

-- ============================================
-- Idempotency Keys Table
-- ============================================
CREATE TABLE IF NOT EXISTS idempotency_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    idempotency_key VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    request_hash VARCHAR(64) NOT NULL,
    response_code INTEGER NOT NULL,
    response_body JSONB NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, idempotency_key)
);

CREATE INDEX idx_idempotency_keys_organization ON idempotency_keys(organization_id);
CREATE INDEX idx_idempotency_keys_key ON idempotency_keys(idempotency_key);
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);

-- ============================================
-- Comments
-- ============================================
COMMENT ON TABLE idempotency_keys IS 'Stores idempotency keys to prevent duplicate operations';
COMMENT ON COLUMN idempotency_keys.idempotency_key IS 'Unique key provided by client for idempotency';
COMMENT ON COLUMN idempotency_keys.request_hash IS 'Hash of the request body to detect changes';
COMMENT ON COLUMN idempotency_keys.response_code IS 'HTTP status code of the original response';
COMMENT ON COLUMN idempotency_keys.response_body IS 'Response body to return for duplicate requests';
COMMENT ON COLUMN idempotency_keys.expires_at IS 'When this idempotency key expires';

-- ============================================
-- Cleanup Function
-- ============================================
CREATE OR REPLACE FUNCTION cleanup_expired_idempotency_keys()
RETURNS void AS $$
BEGIN
    DELETE FROM idempotency_keys WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Trigger for automatic cleanup
-- ============================================
-- Note: This would typically be called by a scheduled job
-- For now, we'll rely on application-level cleanup