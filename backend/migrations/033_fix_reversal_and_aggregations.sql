-- Reversal fields required by the transaction lifecycle code. Summary tables
-- and their legacy functions already exist in 002_architecture_principles.sql
-- with the column names consumed by internal/api/aggregations.go. Do not
-- replace them with a second incompatible schema or attach write-time summary
-- triggers: reports are rebuilt by the application and hard-delete services
-- explicitly rebase affected summaries.

ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS reversed_by UUID REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS reversal_reason TEXT;

ALTER TABLE sales
    ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS reversed_by UUID REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS reversal_reason TEXT;

ALTER TABLE purchases
    ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS reversed_by UUID REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS reversal_reason TEXT;
