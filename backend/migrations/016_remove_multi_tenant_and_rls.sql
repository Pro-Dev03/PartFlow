-- Convert the database to the agreed single-store model. This migration is
-- intentionally catalog-driven: earlier releases did not all have the same
-- optional tables or organization_id columns, and later migrations add some
-- of those tables only after version 016.

-- Remove all policies and disable RLS only on relations that exist now.
DO $$
DECLARE
    policy_row RECORD;
    table_row RECORD;
BEGIN
    FOR policy_row IN
        SELECT schemaname, tablename, policyname
        FROM pg_policies
        WHERE schemaname = current_schema()
    LOOP
        EXECUTE format('DROP POLICY IF EXISTS %I ON %I.%I',
            policy_row.policyname, policy_row.schemaname, policy_row.tablename);
    END LOOP;

    FOR table_row IN
        SELECT schemaname, tablename
        FROM pg_tables
        WHERE schemaname = current_schema()
    LOOP
        EXECUTE format('ALTER TABLE %I.%I DISABLE ROW LEVEL SECURITY',
            table_row.schemaname, table_row.tablename);
    END LOOP;
END $$;

-- Drop tenant ownership columns wherever an earlier tenant-enabled schema
-- added them. CASCADE removes only constraints/indexes/views dependent on that
-- column; the business rows themselves remain untouched.
DO $$
DECLARE
    column_row RECORD;
BEGIN
    FOR column_row IN
        SELECT table_schema, table_name
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND column_name = 'organization_id'
    LOOP
        EXECUTE format('ALTER TABLE %I.%I DROP COLUMN IF EXISTS organization_id CASCADE',
            column_row.table_schema, column_row.table_name);
    END LOOP;
END $$;

-- Restore single-store uniqueness only when the target table and column exist
-- and no equivalent one-column unique index/constraint is already present.
DO $$
DECLARE
    spec RECORD;
    relation_oid OID;
    attribute_number SMALLINT;
BEGIN
    FOR spec IN
        SELECT * FROM (VALUES
            ('categories', 'name', 'categories_name_key'),
            ('products', 'sku', 'products_sku_key'),
            ('customers', 'code', 'customers_code_key'),
            ('suppliers', 'code', 'suppliers_code_key'),
            ('sales', 'invoice_number', 'sales_invoice_number_key'),
            ('purchases', 'invoice_number', 'purchases_invoice_number_key'),
            ('brands', 'name', 'brands_name_key'),
            ('part_types', 'name_ar', 'part_types_name_ar_key'),
            ('barcodes', 'code', 'barcodes_code_key'),
            ('idempotency_keys', 'idempotency_key', 'idempotency_keys_idempotency_key_key')
        ) AS unique_specs(table_name, column_name, constraint_name)
    LOOP
        relation_oid := to_regclass(format('%I.%I', current_schema(), spec.table_name));
        IF relation_oid IS NULL THEN
            CONTINUE;
        END IF;

        SELECT attnum INTO attribute_number
        FROM pg_attribute
        WHERE attrelid = relation_oid
          AND attname = spec.column_name
          AND NOT attisdropped;
        IF attribute_number IS NULL THEN
            CONTINUE;
        END IF;

        IF EXISTS (
            SELECT 1
            FROM pg_index
            WHERE indrelid = relation_oid
              AND indisunique
              AND indpred IS NULL
              AND indexprs IS NULL
              AND indnatts = 1
              AND indkey[0] = attribute_number
        ) THEN
            CONTINUE;
        END IF;

        IF EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conrelid = relation_oid
              AND conname = spec.constraint_name
        ) THEN
            CONTINUE;
        END IF;

        EXECUTE format('ALTER TABLE %I.%I ADD CONSTRAINT %I UNIQUE (%I)',
            current_schema(), spec.table_name, spec.constraint_name, spec.column_name);
    END LOOP;
END $$;

DROP TABLE IF EXISTS organizations CASCADE;
DROP INDEX IF EXISTS idx_sales_organization_date;
CREATE INDEX IF NOT EXISTS idx_sales_date ON sales(sale_date);

DO $$
DECLARE
    relation_row RECORD;
BEGIN
    FOR relation_row IN
        SELECT c.relname, c.relrowsecurity
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = current_schema()
          AND c.relkind IN ('r', 'p')
          AND c.relrowsecurity
    LOOP
        RAISE NOTICE 'RLS remains enabled on table: %', relation_row.relname;
    END LOOP;
END $$;
