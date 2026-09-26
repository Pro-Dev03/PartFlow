BEGIN;

-- Preserve item status before and after an adjustment so deleting that
-- adjustment can reverse both its quantity and status effects.
ALTER TABLE inventory_movements
    ADD COLUMN IF NOT EXISTS before_status TEXT;

ALTER TABLE inventory_movements
    ADD COLUMN IF NOT EXISTS after_status TEXT;

COMMIT;
