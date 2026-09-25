-- Reconciled against the live Cloud schema on 2026-09-25.
-- Schema only: no existing operational row is inserted, updated or deleted.
-- Run this file explicitly after review; do not infer pending work from
-- schema_migrations, since Cloud contains effects of unrecorded migrations.
BEGIN;

-- Backend barcode routes and local sync both require these tables.
CREATE TABLE IF NOT EXISTS barcodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) NOT NULL UNIQUE,
    product_id UUID REFERENCES products(id),
    inventory_item_id UUID REFERENCES inventory_items(id),
    type VARCHAR(20) NOT NULL CHECK (type IN ('EXTERNAL','INTERNAL','SKU','SERIAL','ITEM_CODE')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    generated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_barcodes_product ON barcodes(product_id);
CREATE INDEX IF NOT EXISTS idx_barcodes_item ON barcodes(inventory_item_id);
CREATE INDEX IF NOT EXISTS idx_barcodes_type ON barcodes(type);

CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    item_id UUID NOT NULL REFERENCES inventory_items(id),
    customer_id UUID REFERENCES customers(id),
    user_id UUID NOT NULL REFERENCES users(id),
    reserved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','expired','converted','cancelled')),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_reservations_item ON reservations(item_id);
CREATE INDEX IF NOT EXISTS idx_reservations_customer ON reservations(customer_id);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);
CREATE INDEX IF NOT EXISTS idx_reservations_expires_at ON reservations(expires_at);

-- A completed return retains its posted accounting/stock facts after cleanup.
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
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
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
    created_at TIMESTAMPTZ NOT NULL
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
    created_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_return_effect_refunds_effect ON return_effect_refunds(return_effect_id);

CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID,
    sale_id UUID REFERENCES sales(id) ON DELETE SET NULL,
    payment_id UUID REFERENCES payments(id) ON DELETE SET NULL,
    provider VARCHAR(40) NOT NULL,
    provider_payment_id VARCHAR(255),
    provider_transaction_id VARCHAR(255),
    status VARCHAR(30) NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency VARCHAR(3) NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL,
    checkout_url TEXT,
    failure_code VARCHAR(100),
    failure_message TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    UNIQUE(provider, idempotency_key)
);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_sale ON payment_transactions(sale_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_order ON payment_transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_provider_id ON payment_transactions(provider, provider_payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_status ON payment_transactions(status);

CREATE TABLE IF NOT EXISTS payment_refunds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_transaction_id UUID NOT NULL REFERENCES payment_transactions(id) ON DELETE RESTRICT,
    provider_refund_id VARCHAR(255),
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(30) NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    reason TEXT,
    failure_message TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_payment_refunds_transaction ON payment_refunds(payment_transaction_id);

CREATE TABLE IF NOT EXISTS payment_webhook_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider VARCHAR(40) NOT NULL,
    provider_event_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100),
    payment_transaction_id UUID REFERENCES payment_transactions(id) ON DELETE SET NULL,
    payload JSONB NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'received',
    error_message TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    UNIQUE(provider, provider_event_id)
);

CREATE TABLE IF NOT EXISTS return_payment_refunds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    return_id UUID NOT NULL UNIQUE REFERENCES returns(id) ON DELETE RESTRICT,
    payment_transaction_id UUID NOT NULL REFERENCES payment_transactions(id) ON DELETE RESTRICT,
    payment_refund_id UUID REFERENCES payment_refunds(id) ON DELETE SET NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_return_payment_refunds_transaction
    ON return_payment_refunds(payment_transaction_id);

CREATE TABLE IF NOT EXISTS customer_debts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    reference_id UUID,
    reference_type VARCHAR(50),
    due_date DATE NOT NULL,
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,
    paid_amount NUMERIC(15,2) NOT NULL DEFAULT 0 CHECK (paid_amount >= 0 AND paid_amount <= amount),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_customer_debts_customer_due
    ON customer_debts(customer_id, is_paid, due_date, created_at);

CREATE TABLE IF NOT EXISTS supplier_debts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    supplier_id UUID NOT NULL REFERENCES suppliers(id),
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    reference_id UUID,
    reference_type VARCHAR(50),
    due_date DATE NOT NULL,
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,
    paid_amount NUMERIC(15,2) NOT NULL DEFAULT 0 CHECK (paid_amount >= 0 AND paid_amount <= amount),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_supplier_debts_supplier_due
    ON supplier_debts(supplier_id, is_paid, due_date, created_at);

CREATE TABLE IF NOT EXISTS debt_collections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    type VARCHAR(40) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    notes TEXT,
    scheduled_date TIMESTAMPTZ NOT NULL,
    completed_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_debt_collections_customer_schedule
    ON debt_collections(customer_id, status, scheduled_date);

CREATE TABLE IF NOT EXISTS supplier_debt_collections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    supplier_id UUID NOT NULL REFERENCES suppliers(id),
    type VARCHAR(40) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    notes TEXT,
    scheduled_date TIMESTAMPTZ NOT NULL,
    completed_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_supplier_debt_collections_supplier_schedule
    ON supplier_debt_collections(supplier_id, status, scheduled_date);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user ON password_reset_tokens(user_id);

-- Current cleanup code already writes this audit snapshot, independently of
-- the older, unrecorded migration 086. It has no FK to cleaned operations.
CREATE TABLE IF NOT EXISTS deleted_operation_snapshots (
    entity_type VARCHAR(80) NOT NULL,
    operation_id VARCHAR(100) NOT NULL,
    business_number VARCHAR(120),
    status VARCHAR(40) NOT NULL,
    snapshot JSONB NOT NULL,
    deleted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (entity_type, operation_id)
);
CREATE INDEX IF NOT EXISTS idx_deleted_operation_snapshots_deleted_at
    ON deleted_operation_snapshots(deleted_at);

-- Columns verified missing from the live Cloud schema and used by Backend.
ALTER TABLE products ADD COLUMN IF NOT EXISTS purchase_price NUMERIC(14,2);
ALTER TABLE products ADD COLUMN IF NOT EXISTS currency VARCHAR(3);
ALTER TABLE sales ADD COLUMN IF NOT EXISTS sale_number VARCHAR(100);
ALTER TABLE sales ADD COLUMN IF NOT EXISTS remaining_amount NUMERIC(14,2);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_sale_number_unique
    ON sales(sale_number) WHERE sale_number IS NOT NULL;
-- Historic allocations have no recorded status. Leave them NULL instead of
-- assigning an invented financial state; new writes provide status explicitly.
ALTER TABLE sale_payment_allocations ADD COLUMN IF NOT EXISTS status VARCHAR(20);
CREATE INDEX IF NOT EXISTS idx_sale_payment_allocations_status ON sale_payment_allocations(status);
ALTER TABLE inventory_movements ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE inventory_movements ADD COLUMN IF NOT EXISTS reversed_by UUID;
ALTER TABLE inventory_movements ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMPTZ;
ALTER TABLE inventory_movements ADD COLUMN IF NOT EXISTS reversal_reason TEXT;
-- Archive's customer ledger source reads these fields, while the current
-- Cloud ledger only has the older type column. Existing values stay intact.
ALTER TABLE customer_ledger ADD COLUMN IF NOT EXISTS transaction_type VARCHAR(20);
ALTER TABLE customer_ledger ADD COLUMN IF NOT EXISTS created_by UUID;

-- Cloud migration 035 renamed this column to item_id, while Local and the
-- Backend use inventory_item_id. Rename metadata only: existing IDs stay put.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema = 'public' AND table_name = 'trade_ins' AND column_name = 'item_id')
       AND EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_schema = 'public' AND table_name = 'trade_ins' AND column_name = 'inventory_item_id') THEN
        RAISE EXCEPTION 'trade_ins has both item_id and inventory_item_id; reconcile manually';
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_schema = 'public' AND table_name = 'trade_ins' AND column_name = 'item_id') THEN
        ALTER TABLE trade_ins RENAME COLUMN item_id TO inventory_item_id;
    END IF;
END $$;
ALTER TABLE trade_ins ALTER COLUMN inventory_item_id DROP NOT NULL;
ALTER TABLE trade_ins ALTER COLUMN purchase_price TYPE NUMERIC(14,2)
    USING purchase_price::NUMERIC(14,2);

-- Read models used by reports and return cleanup. Both combine live returns
-- with posted effects that intentionally outlive the operational return row.
CREATE OR REPLACE VIEW accounting_returns AS
SELECT r.id, r.return_number, r.reference_number, r.sale_id, r.purchase_id, r.customer_id,
       r.total_refund_amount,
       CASE WHEN r.refund_date IS NOT NULL OR UPPER(COALESCE(r.status, '')) = 'COMPLETED'
            THEN 'refunded'::VARCHAR(30) ELSE 'pending'::VARCHAR(30) END AS refund_status,
       r.status, r.return_date, r.refund_date, r.reason, r.refund_method,
       r.debt_id, r.debt_adjustment, r.customer_credit, r.created_at, r.updated_at,
       r.return_type, r.is_warranty_claim, r.item_condition_after_return
FROM returns r
UNION ALL
SELECT id, NULL::VARCHAR(50),
       CASE WHEN is_reversal THEN 'REV-POSTED'::VARCHAR(50) ELSE NULL::VARCHAR(50) END,
       sale_id, purchase_id, customer_id, total_refund_amount, 'refunded'::VARCHAR(30),
       status, return_date, refund_date, NULL::TEXT, refund_method,
       debt_id, debt_adjustment, customer_credit, created_at, updated_at,
       CASE WHEN EXISTS (SELECT 1 FROM return_effect_items ri WHERE ri.return_effect_id = return_effects.id)
                 AND (SELECT COALESCE(SUM(ri.quantity_returned), 0)
                      FROM return_effect_items ri WHERE ri.return_effect_id = return_effects.id)
                     >= (SELECT COALESCE(SUM(COALESCE(ri.original_quantity, ri.quantity_returned)), 0)
                         FROM return_effect_items ri WHERE ri.return_effect_id = return_effects.id)
            THEN 'FULL'::VARCHAR(30)
            WHEN EXISTS (SELECT 1 FROM return_effect_items ri WHERE ri.return_effect_id = return_effects.id)
            THEN 'QUANTITY_PARTIAL'::VARCHAR(30) ELSE NULL::VARCHAR(30) END,
       FALSE, NULL::TEXT
FROM return_effects;

CREATE OR REPLACE VIEW accounting_return_items AS
SELECT id, return_id, sale_item_id, product_id, inventory_item_id,
       serial_number, barcode, quantity_returned, original_quantity, unit_price,
       total_refund_amount, original_cost, resolution, inventory_status, created_at
FROM return_items
UNION ALL
SELECT id, return_effect_id, sale_item_id, product_id, inventory_item_id,
       serial_number, barcode, quantity_returned, original_quantity, unit_price,
       total_refund_amount, original_cost, resolution, inventory_status, created_at
FROM return_effect_items;

COMMIT;
