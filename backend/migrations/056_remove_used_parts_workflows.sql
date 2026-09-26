-- Retire legacy used-parts history, repair-cost and aging workflows. Product
-- and acquisition inspections remain active workflows in the current app.
DROP TRIGGER IF EXISTS create_item_history_on_link ON acquisition_items;
DROP FUNCTION IF EXISTS create_item_history_entry();
DROP VIEW IF EXISTS used_parts_aging;

DROP TABLE IF EXISTS item_repair_costs;
