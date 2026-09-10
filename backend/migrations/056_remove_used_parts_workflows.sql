-- Keep used-parts acquisition and sale, but remove the retired inspection,
-- item-history, repair-cost, rejection, and aging workflows.
DROP TRIGGER IF EXISTS create_item_history_on_link ON acquisition_items;
DROP FUNCTION IF EXISTS create_item_history_entry();
DROP VIEW IF EXISTS used_parts_aging;

DROP TABLE IF EXISTS inspection_items;
DROP TABLE IF EXISTS item_repair_costs;
DROP TABLE IF EXISTS item_history;
DROP TABLE IF EXISTS inspections;

ALTER TABLE acquisition_items
  DROP COLUMN IF EXISTS inspection_id,
  DROP COLUMN IF EXISTS inspection_status;

ALTER TABLE acquisition_items
  ALTER COLUMN item_status SET DEFAULT 'available';

UPDATE acquisition_items
SET item_status = 'available'
WHERE item_status IN ('inspection', 'rejected', 'for_parts', 'archived');