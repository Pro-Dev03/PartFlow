-- Single-store Supabase hardening. This migration deliberately does not add
-- tenant_id columns or enable tenant RLS; account and subscription decisions
-- remain in the authenticated PartFlow API.
BEGIN;

-- Keep PostgreSQL compatible with the columns currently consumed by the API
-- when a Supabase project was initialized from an older schema snapshot.
DO $$
BEGIN
    IF to_regclass('public.sales') IS NOT NULL THEN
        ALTER TABLE public.sales
            ADD COLUMN IF NOT EXISTS cost_amount NUMERIC(15,2) DEFAULT 0,
            ADD COLUMN IF NOT EXISTS gross_profit NUMERIC(15,2) DEFAULT 0,
            ADD COLUMN IF NOT EXISTS net_profit NUMERIC(15,2) DEFAULT 0,
            ADD COLUMN IF NOT EXISTS cash_received NUMERIC(12,2) NOT NULL DEFAULT 0,
            ADD COLUMN IF NOT EXISTS change_amount NUMERIC(12,2) NOT NULL DEFAULT 0;
    END IF;

    IF to_regclass('public.sale_items') IS NOT NULL THEN
        ALTER TABLE public.sale_items
            ADD COLUMN IF NOT EXISTS unit_cost NUMERIC(15,2) DEFAULT 0,
            ADD COLUMN IF NOT EXISTS barcode VARCHAR(100);
        CREATE INDEX IF NOT EXISTS idx_sale_items_barcode ON public.sale_items(barcode);
    END IF;

    IF to_regclass('public.products') IS NOT NULL THEN
        ALTER TABLE public.products ADD COLUMN IF NOT EXISTS category_id UUID;
    END IF;

    IF to_regclass('public.inventory_items') IS NOT NULL THEN
        ALTER TABLE public.inventory_items ADD COLUMN IF NOT EXISTS category_id UUID;
        IF to_regclass('public.products') IS NOT NULL THEN
            UPDATE public.inventory_items AS ii
            SET category_id = p.category_id
            FROM public.products AS p
            WHERE ii.product_id = p.id
              AND ii.category_id IS NULL
              AND p.category_id IS NOT NULL;
        END IF;
    END IF;
END;
$$;

-- A normal view executes with its owner's privileges, which can bypass RLS
-- applied to its base tables. Make application views invoker-security on PG15+
-- and remove direct PostgREST access for public/anon/authenticated roles.
DO $$
DECLARE
    view_record RECORD;
    pg_version INTEGER := current_setting('server_version_num')::INTEGER;
BEGIN
    FOR view_record IN
        SELECT c.relname
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public'
          AND c.relkind = 'v'
          AND c.relname = ANY (ARRAY[
              'customer_ledger_view', 'supplier_ledger_view', 'inventory_ledger_view',
              'seller_balances', 'customer_acquisition_summary', 'used_parts_aging',
              'returns_summary', 'monthly_returns_analysis', 'sales_returns_analysis',
              'accounting_returns', 'accounting_return_items'
          ])
    LOOP
        IF pg_version >= 150000 THEN
            EXECUTE format('ALTER VIEW public.%I SET (security_invoker = true)', view_record.relname);
        END IF;
        EXECUTE format('REVOKE ALL PRIVILEGES ON TABLE public.%I FROM PUBLIC', view_record.relname);
        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
            EXECUTE format('REVOKE ALL PRIVILEGES ON TABLE public.%I FROM anon', view_record.relname);
        END IF;
        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
            EXECUTE format('REVOKE ALL PRIVILEGES ON TABLE public.%I FROM authenticated', view_record.relname);
        END IF;
    END LOOP;
END;
$$;

-- Pin the search path for legacy application routines to prevent objects in
-- writable schemas from shadowing the tables/functions they resolve.
DO $$
DECLARE
    function_record RECORD;
BEGIN
    FOR function_record IN
        SELECT p.oid::regprocedure AS signature
        FROM pg_proc p
        JOIN pg_namespace n ON n.oid = p.pronamespace
        WHERE n.nspname = 'public'
          AND p.prokind = 'f'
          AND p.proname = ANY (ARRAY[
              'update_updated_at_column', 'generate_item_code', 'generate_internal_barcode',
              'cleanup_expired_idempotency_keys', 'sync_role_permissions',
              'update_customer_balance', 'update_supplier_balance', 'update_inventory_quantity',
              'calculate_acquisition_total', 'create_item_history_entry', 'generate_return_number',
              'handle_return_debt_adjustment', 'validate_return_quantity',
              'update_return_item_inventory_status', 'update_inventory_current_state',
              'update_daily_sales_summary', 'update_monthly_sales_summary'
          ])
    LOOP
        EXECUTE format('ALTER FUNCTION %s SET search_path TO pg_catalog, public', function_record.signature);
    END LOOP;
END;
$$;

-- This helper is not a client RPC. Do not let Supabase API roles invoke a
-- SECURITY DEFINER routine that changes RLS state on arbitrary public tables.
DO $$
DECLARE
    function_record RECORD;
BEGIN
    FOR function_record IN
        SELECT p.oid::regprocedure AS signature
        FROM pg_proc p
        JOIN pg_namespace n ON n.oid = p.pronamespace
        WHERE n.nspname = 'public' AND p.proname = 'rls_auto_enable'
    LOOP
        EXECUTE format('REVOKE EXECUTE ON FUNCTION %s FROM PUBLIC', function_record.signature);
        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
            EXECUTE format('REVOKE EXECUTE ON FUNCTION %s FROM anon', function_record.signature);
        END IF;
        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
            EXECUTE format('REVOKE EXECUTE ON FUNCTION %s FROM authenticated', function_record.signature);
        END IF;
    END LOOP;
END;
$$;

COMMIT;
