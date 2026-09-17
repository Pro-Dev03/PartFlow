-- Durable Customer Return -> Supplier Return requests.
-- Missing source data is represented by a real request, never by a read-time projection.
ALTER TABLE supplier_returns ADD COLUMN IF NOT EXISTS customer_return_id UUID REFERENCES returns(id);
ALTER TABLE supplier_returns ADD COLUMN IF NOT EXISTS sale_id UUID REFERENCES sales(id);
ALTER TABLE supplier_returns ADD COLUMN IF NOT EXISTS return_reason TEXT;
ALTER TABLE supplier_returns ADD COLUMN IF NOT EXISTS return_date TIMESTAMP WITH TIME ZONE;
ALTER TABLE supplier_returns
    ADD COLUMN IF NOT EXISTS source_status VARCHAR(30) NOT NULL DEFAULT 'RESOLVED';

ALTER TABLE supplier_return_items ADD COLUMN IF NOT EXISTS customer_return_id UUID REFERENCES returns(id);
ALTER TABLE supplier_return_items ADD COLUMN IF NOT EXISTS sale_id UUID REFERENCES sales(id);
ALTER TABLE supplier_return_items ADD COLUMN IF NOT EXISTS sale_item_id UUID REFERENCES sale_items(id);
ALTER TABLE supplier_return_items ADD COLUMN IF NOT EXISTS inventory_item_id UUID REFERENCES inventory_items(id);
ALTER TABLE supplier_return_items ADD COLUMN IF NOT EXISTS barcode VARCHAR(100);
ALTER TABLE supplier_return_items ADD COLUMN IF NOT EXISTS serial_number VARCHAR(100);
ALTER TABLE supplier_return_items ADD COLUMN IF NOT EXISTS purchase_cost DECIMAL(10,2) NOT NULL DEFAULT 0;
ALTER TABLE supplier_return_items ADD COLUMN IF NOT EXISTS return_reason TEXT;
ALTER TABLE supplier_return_items ADD COLUMN IF NOT EXISTS return_date TIMESTAMP WITH TIME ZONE;

ALTER TABLE purchase_items ADD COLUMN IF NOT EXISTS barcode VARCHAR(100);
ALTER TABLE purchase_items ADD COLUMN IF NOT EXISTS serial_number VARCHAR(100);
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS barcode VARCHAR(100);
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS serial_number VARCHAR(100);

ALTER TABLE supplier_returns
    ALTER COLUMN purchase_id DROP NOT NULL,
    ALTER COLUMN supplier_id DROP NOT NULL;

ALTER TABLE supplier_returns DROP CONSTRAINT IF EXISTS supplier_returns_status_check;
ALTER TABLE supplier_returns ADD CONSTRAINT supplier_returns_status_check
    CHECK (status IN ('DRAFT','PENDING','SHIPPED','RECEIVED','COMPLETED','REJECTED','NEEDS_SOURCE_DATA'));

CREATE INDEX IF NOT EXISTS idx_supplier_returns_source_status
    ON supplier_returns(source_status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_supplier_returns_customer_return_once
    ON supplier_returns(customer_return_id)
    WHERE customer_return_id IS NOT NULL;

-- Backfill one durable request per historical customer return. The source chain
-- is intentionally anchored by sale_item + inventory_item; product-only guesses
-- are not acceptable.
WITH source_rows AS (
    SELECT DISTINCT ON (r.id)
        r.id AS customer_return_id,
        r.sale_id,
        ri.sale_item_id,
        ri.inventory_item_id,
        pi.id AS purchase_item_id,
        pi.purchase_id,
        p.supplier_id,
        COALESCE(NULLIF(ri.barcode, ''), ii.barcode) AS barcode,
        COALESCE(NULLIF(ri.serial_number, ''), ii.serial_number) AS serial_number,
        COALESCE(ri.quantity_returned, 0) AS quantity,
        COALESCE(ri.original_cost, ii.purchase_cost, pi.unit_price, 0) AS purchase_cost,
        COALESCE(r.reason, 'CUSTOMER_RETURN') AS reason,
        COALESCE(r.return_date, r.created_at) AS return_date,
        r.return_number,
        r.created_by
    FROM returns r
    LEFT JOIN return_items ri ON ri.return_id = r.id
    LEFT JOIN sale_items si ON si.id = ri.sale_item_id
    LEFT JOIN inventory_items ii ON ii.id = ri.inventory_item_id
    LEFT JOIN purchase_items pi ON pi.product_id = ri.product_id
        AND (pi.serial_number = ii.serial_number OR pi.barcode = ii.barcode
             OR ii.item_code LIKE 'ITM-' || LEFT(REPLACE(pi.purchase_id::text, '-', ''), 8) || '-%')
    LEFT JOIN purchases p ON p.id = pi.purchase_id
    WHERE UPPER(COALESCE(r.item_condition_after_return, '')) IN ('RETURN_TO_SUPPLIER', 'SUPPLIER_RETURN')
    ORDER BY r.id, (p.id IS NOT NULL) DESC, (pi.id IS NOT NULL) DESC, pi.created_at DESC
), inserted_requests AS (
    INSERT INTO supplier_returns
        (customer_return_id, sale_id, purchase_id, supplier_id, return_number, status,
         source_status, reason, notes, created_by, return_reason, return_date)
    SELECT s.customer_return_id, s.sale_id, s.purchase_id, s.supplier_id,
        'SRET-' || UPPER(SUBSTRING(REPLACE(s.customer_return_id::text, '-', ''), 1, 10)),
        CASE WHEN s.purchase_id IS NULL OR s.supplier_id IS NULL THEN 'NEEDS_SOURCE_DATA' ELSE 'PENDING' END,
        CASE WHEN s.purchase_id IS NULL OR s.supplier_id IS NULL THEN 'NEEDS_SOURCE_DATA' ELSE 'RESOLVED' END,
        s.reason, 'Historical reconciliation', s.created_by, s.reason, s.return_date
    FROM source_rows s
    WHERE NOT EXISTS (SELECT 1 FROM supplier_returns sr WHERE sr.customer_return_id = s.customer_return_id)
    RETURNING id, customer_return_id
)
INSERT INTO supplier_return_items
    (customer_return_id, sale_id, sale_item_id, inventory_item_id, supplier_return_id,
     purchase_item_id, product_id, quantity, unit_cost, barcode, serial_number,
     purchase_cost, return_reason, return_date)
SELECT s.customer_return_id, s.sale_id, s.sale_item_id, s.inventory_item_id,
    sr.id, s.purchase_item_id, ri.product_id, s.quantity, s.purchase_cost,
    s.barcode, s.serial_number, s.purchase_cost, s.reason, s.return_date
FROM source_rows s
JOIN supplier_returns sr ON sr.customer_return_id = s.customer_return_id
JOIN return_items ri ON ri.return_id = s.customer_return_id
WHERE s.purchase_item_id IS NOT NULL
  AND s.purchase_id IS NOT NULL
  AND s.supplier_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM supplier_return_items sri
      WHERE sri.customer_return_id = s.customer_return_id
        AND sri.inventory_item_id = s.inventory_item_id
  );
