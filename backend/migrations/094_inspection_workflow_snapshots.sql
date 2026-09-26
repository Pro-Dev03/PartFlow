BEGIN;

CREATE TABLE IF NOT EXISTS inspection_workflow_snapshots (
    inspection_id UUID PRIMARY KEY REFERENCES inspections(id) ON DELETE CASCADE,
    inventory_item_id UUID,
    inventory_status_before TEXT,
    acquisition_item_id UUID,
    acquisition_inventory_item_id_before UUID,
    acquisition_inspection_id_before UUID,
    acquisition_inspection_status_before TEXT,
    acquisition_item_status_before TEXT,
    acquisition_id UUID,
    acquisition_status_before TEXT,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
