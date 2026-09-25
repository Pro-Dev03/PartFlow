-- Final additive reconciliation against live Cloud schema and active Backend.
-- No existing row is inserted, updated, deleted, or backfilled.
BEGIN;

-- Local enforces this business identity. Cloud has no duplicates as checked
-- before rollout; the partial index permits historical NULL links.
CREATE UNIQUE INDEX IF NOT EXISTS idx_supplier_return_items_customer_return_item
    ON supplier_return_items(customer_return_id, inventory_item_id)
    WHERE customer_return_id IS NOT NULL AND inventory_item_id IS NOT NULL;

-- Backend customer/supplier payment lists use these foreign-key columns.
CREATE INDEX IF NOT EXISTS idx_payments_customer ON payments(customer_id);
CREATE INDEX IF NOT EXISTS idx_payments_supplier ON payments(supplier_id);

-- These updated_at triggers were defined by older migrations but are absent
-- from the live schema. They maintain new edits only, with no row backfill.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'public.inventory_items'::regclass AND tgname = 'update_inventory_items_updated_at' AND NOT tgisinternal) THEN
        CREATE TRIGGER update_inventory_items_updated_at BEFORE UPDATE ON inventory_items
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'public.locations'::regclass AND tgname = 'update_locations_updated_at' AND NOT tgisinternal) THEN
        CREATE TRIGGER update_locations_updated_at BEFORE UPDATE ON locations
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'public.expense_categories'::regclass AND tgname = 'update_expense_categories_updated_at' AND NOT tgisinternal) THEN
        CREATE TRIGGER update_expense_categories_updated_at BEFORE UPDATE ON expense_categories
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'public.reservations'::regclass AND tgname = 'update_reservations_updated_at' AND NOT tgisinternal) THEN
        CREATE TRIGGER update_reservations_updated_at BEFORE UPDATE ON reservations
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'public.barcodes'::regclass AND tgname = 'update_barcodes_updated_at' AND NOT tgisinternal) THEN
        CREATE TRIGGER update_barcodes_updated_at BEFORE UPDATE ON barcodes
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- Analytical views must include posted return effects after an operational
-- return is cleaned. The operational returns_summary view remains unchanged.
CREATE OR REPLACE VIEW monthly_returns_analysis AS
SELECT DATE_TRUNC('month', r.return_date) AS month,
       COUNT(DISTINCT r.id) AS total_returns,
       COUNT(DISTINCT r.customer_id) AS unique_customers,
       COALESCE(SUM(r.total_refund_amount), 0) AS total_refund_amount,
       COALESCE(AVG(r.total_refund_amount), 0) AS avg_refund_amount,
       COUNT(CASE WHEN UPPER(COALESCE(r.return_type, '')) = 'FULL' THEN 1 END) AS full_returns,
       COUNT(CASE WHEN UPPER(COALESCE(r.return_type, '')) = 'PARTIAL' THEN 1 END) AS partial_returns,
       COUNT(CASE WHEN UPPER(COALESCE(r.return_type, '')) = 'QUANTITY_PARTIAL' THEN 1 END) AS quantity_partial_returns,
       COUNT(CASE WHEN UPPER(COALESCE(r.reason, '')) = 'DEFECTIVE' THEN 1 END) AS defective_returns,
       COUNT(CASE WHEN UPPER(COALESCE(r.reason, '')) = 'WARRANTY' THEN 1 END) AS warranty_returns,
       COUNT(CASE WHEN r.is_warranty_claim = TRUE THEN 1 END) AS warranty_claims,
       SUM(CASE WHEN r.item_condition_after_return = 'SELLABLE' THEN 1 ELSE 0 END) AS sellable_items,
       SUM(CASE WHEN r.item_condition_after_return = 'NEEDS_REPAIR' THEN 1 ELSE 0 END) AS repair_needed,
       SUM(CASE WHEN r.item_condition_after_return = 'WRITE_OFF' THEN 1 ELSE 0 END) AS written_off
FROM accounting_returns r
WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
GROUP BY DATE_TRUNC('month', r.return_date)
ORDER BY month DESC;

CREATE OR REPLACE VIEW sales_returns_analysis AS
WITH returns_by_sale_month AS (
    SELECT sale_id, DATE_TRUNC('month', return_date) AS month,
           SUM(total_refund_amount) AS returns_amount, COUNT(*) AS return_count
    FROM accounting_returns
    WHERE UPPER(COALESCE(status, '')) = 'COMPLETED' AND sale_id IS NOT NULL
    GROUP BY sale_id, DATE_TRUNC('month', return_date)
)
SELECT DATE_TRUNC('month', s.sale_date) AS month,
       COUNT(DISTINCT s.id) AS total_sales,
       COALESCE(SUM(s.total_amount), 0) AS gross_sales,
       COALESCE(SUM(s.cost_amount), 0) AS total_cost,
       COALESCE(SUM(s.gross_profit), 0) AS gross_profit,
       COALESCE(SUM(r.returns_amount), 0) AS returns_amount,
       COALESCE(SUM(r.return_count), 0)::BIGINT AS return_count,
       COALESCE(SUM(s.total_amount), 0) - COALESCE(SUM(r.returns_amount), 0) AS net_sales
FROM sales s
LEFT JOIN returns_by_sale_month r ON r.sale_id = s.id AND r.month = DATE_TRUNC('month', s.sale_date)
WHERE s.status = 'completed'
GROUP BY DATE_TRUNC('month', s.sale_date)
ORDER BY month DESC;

COMMIT;
