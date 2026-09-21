-- Archive filters read these source columns directly through UNION ALL.
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX IF NOT EXISTS idx_inventory_movements_reference_id ON inventory_movements(reference_id);
CREATE INDEX IF NOT EXISTS idx_item_history_reference_id ON item_history(reference_id);