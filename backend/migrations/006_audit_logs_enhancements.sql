-- ============================================
-- Audit Logs Enhancements
-- ============================================

-- Add new columns to audit_logs table
ALTER TABLE audit_logs 
ADD COLUMN IF NOT EXISTS request_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS changes TEXT,
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'success',
ADD COLUMN IF NOT EXISTS error_message TEXT,
ADD COLUMN IF NOT EXISTS metadata TEXT;

-- Add indexes for new columns
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);

-- Add check constraint for status
ALTER TABLE audit_logs 
ADD CONSTRAINT chk_audit_logs_status 
CHECK (status IN ('success', 'failure'));

-- Update existing records to have default values
UPDATE audit_logs 
SET status = 'success',
    description = COALESCE(description, 'Audit log entry'),
    changes = COALESCE(changes, '{}'::text)
WHERE status IS NULL OR description IS NULL;

-- Drop old JSONB columns if they exist (they're replaced by changes text column)
-- ALTER TABLE audit_logs DROP COLUMN IF EXISTS old_values;
-- ALTER TABLE audit_logs DROP COLUMN IF EXISTS new_values;