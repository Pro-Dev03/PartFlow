-- Persist the fields required by the used-parts inspection workflow.
ALTER TABLE inspections
    ADD COLUMN IF NOT EXISTS condition VARCHAR(20),
    ADD COLUMN IF NOT EXISTS grade VARCHAR(2),
    ADD COLUMN IF NOT EXISTS test_results JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS acquisition_item_id UUID REFERENCES acquisition_items(id),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_inspections_acquisition_item
    ON inspections(acquisition_item_id);
