-- Remove redundant indexes reported by the Supabase database linter.
-- Keep the canonical indexes created by the original schema migrations.
BEGIN;

DROP INDEX IF EXISTS public.idx_daily_debt_summary_runtime_date;
DROP INDEX IF EXISTS public.idx_daily_inventory_summary_runtime_date;
DROP INDEX IF EXISTS public.idx_daily_profit_summary_runtime_date;
DROP INDEX IF EXISTS public.idx_daily_sales_summary_runtime_date;

DROP INDEX IF EXISTS public.idx_monthly_debt_summary_runtime_year_month;
DROP INDEX IF EXISTS public.idx_monthly_inventory_summary_runtime_year_month;
DROP INDEX IF EXISTS public.idx_monthly_profit_summary_runtime_year_month;
DROP INDEX IF EXISTS public.idx_monthly_sales_summary_runtime_year_month;

DROP INDEX IF EXISTS public.idx_supplier_ledger_runtime_type;
DROP INDEX IF EXISTS public.idx_supplier_ledger_type_normalized;

COMMIT;
