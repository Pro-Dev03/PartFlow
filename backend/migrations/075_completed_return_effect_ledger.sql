-- Keep the posted accounting and stock effect of a completed customer return
-- after its operational return request is physically deleted. These rows are
-- ledger facts, not an archived return workflow record.
CREATE TABLE IF NOT EXISTS return_effects (
    id UUID PRIMARY KEY,
    sale_id UUID,
    purchase_id UUID,
    customer_id UUID,
    total_refund_amount NUMERIC(14,2) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'COMPLETED',
    return_date DATE,
    refund_date DATE,
    refund_method VARCHAR(50),
    debt_id UUID,
    debt_adjustment NUMERIC(14,2) NOT NULL DEFAULT 0,
    customer_credit NUMERIC(14,2) NOT NULL DEFAULT 0,
    is_reversal BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_return_effects_date ON return_effects(return_date);

CREATE TABLE IF NOT EXISTS return_effect_items (
    id UUID PRIMARY KEY,
    return_effect_id UUID NOT NULL REFERENCES return_effects(id) ON DELETE CASCADE,
    sale_item_id UUID,
    product_id UUID,
    inventory_item_id UUID,
    serial_number VARCHAR(100),
    barcode VARCHAR(100),
    quantity_returned INTEGER NOT NULL DEFAULT 0,
    original_quantity INTEGER,
    unit_price NUMERIC(14,2) NOT NULL DEFAULT 0,
    total_refund_amount NUMERIC(14,2) NOT NULL DEFAULT 0,
    original_cost NUMERIC(14,2) NOT NULL DEFAULT 0,
    resolution VARCHAR(30),
    inventory_status VARCHAR(30),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_return_effect_items_effect ON return_effect_items(return_effect_id);
CREATE INDEX IF NOT EXISTS idx_return_effect_items_sale_item ON return_effect_items(sale_item_id);
CREATE INDEX IF NOT EXISTS idx_return_effect_items_product ON return_effect_items(product_id);

CREATE TABLE IF NOT EXISTS return_effect_refunds (
    id UUID PRIMARY KEY,
    return_effect_id UUID NOT NULL REFERENCES return_effects(id) ON DELETE CASCADE,
    refund_type VARCHAR(50),
    amount NUMERIC(14,2) NOT NULL DEFAULT 0,
    refund_date DATE,
    payment_method VARCHAR(50),
    transaction_reference VARCHAR(100),
    debt_id UUID,
    debt_reduction_amount NUMERIC(14,2),
    created_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_return_effect_refunds_effect ON return_effect_refunds(return_effect_id);

CREATE OR REPLACE VIEW accounting_returns AS
SELECT r.id, r.return_number, r.reference_number, r.sale_id, r.purchase_id, r.customer_id,
       r.total_refund_amount,
       CASE
           WHEN r.refund_date IS NOT NULL OR UPPER(COALESCE(r.status, '')) = 'COMPLETED'
               THEN 'refunded'::VARCHAR(30)
           ELSE 'pending'::VARCHAR(30)
       END AS refund_status,
       r.status, r.return_date, r.refund_date,
       r.reason, r.refund_method, r.debt_id, r.debt_adjustment, r.customer_credit, r.created_at, r.updated_at,
       r.return_type, r.is_warranty_claim, r.item_condition_after_return
FROM returns r
UNION ALL
SELECT id, NULL::VARCHAR(50) AS return_number,
       CASE WHEN is_reversal THEN 'REV-POSTED'::VARCHAR(50) ELSE NULL::VARCHAR(50) END AS reference_number,
       sale_id, purchase_id, customer_id, total_refund_amount, 'refunded'::VARCHAR(30) AS refund_status,
       status, return_date, refund_date, NULL::TEXT AS reason, refund_method,
       debt_id, debt_adjustment, customer_credit, created_at, updated_at,
       CASE WHEN EXISTS (SELECT 1 FROM return_effect_items ri WHERE ri.return_effect_id = return_effects.id)
            AND (SELECT COALESCE(SUM(ri.quantity_returned), 0) FROM return_effect_items ri WHERE ri.return_effect_id = return_effects.id)
                >= (SELECT COALESCE(SUM(COALESCE(ri.original_quantity, ri.quantity_returned)), 0) FROM return_effect_items ri WHERE ri.return_effect_id = return_effects.id)
            THEN 'FULL'::VARCHAR(30)
            WHEN EXISTS (SELECT 1 FROM return_effect_items ri WHERE ri.return_effect_id = return_effects.id)
            THEN 'QUANTITY_PARTIAL'::VARCHAR(30)
            ELSE NULL::VARCHAR(30) END AS return_type,
       FALSE AS is_warranty_claim, NULL::TEXT AS item_condition_after_return
FROM return_effects;

CREATE OR REPLACE VIEW accounting_return_items AS
SELECT id, return_id, sale_item_id, product_id, inventory_item_id,
       serial_number, barcode, quantity_returned, original_quantity, unit_price,
       total_refund_amount, original_cost, resolution, inventory_status, created_at
FROM return_items
UNION ALL
SELECT id, return_effect_id AS return_id, sale_item_id, product_id, inventory_item_id,
       serial_number, barcode, quantity_returned, original_quantity, unit_price,
       total_refund_amount, original_cost, resolution, inventory_status, created_at
FROM return_effect_items;

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
       COALESCE(SUM(r.return_count), 0) AS return_count,
       COALESCE(SUM(s.total_amount), 0) - COALESCE(SUM(r.returns_amount), 0) AS net_sales
FROM sales s
LEFT JOIN returns_by_sale_month r ON r.sale_id = s.id AND r.month = DATE_TRUNC('month', s.sale_date)
WHERE s.status = 'completed'
GROUP BY DATE_TRUNC('month', s.sale_date)
ORDER BY month DESC;
