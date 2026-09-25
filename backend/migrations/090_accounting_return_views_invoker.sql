-- The accounting views were created after the earlier view-hardening migration.
-- Check underlying table permissions and RLS as the querying role, and keep
-- direct Supabase API roles from reading business data through these views.
BEGIN;

ALTER VIEW public.accounting_returns SET (security_invoker = true);
ALTER VIEW public.accounting_return_items SET (security_invoker = true);

REVOKE ALL PRIVILEGES ON TABLE public.accounting_returns FROM PUBLIC;
REVOKE ALL PRIVILEGES ON TABLE public.accounting_return_items FROM PUBLIC;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'anon') THEN
        REVOKE ALL PRIVILEGES ON TABLE public.accounting_returns FROM anon;
        REVOKE ALL PRIVILEGES ON TABLE public.accounting_return_items FROM anon;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated') THEN
        REVOKE ALL PRIVILEGES ON TABLE public.accounting_returns FROM authenticated;
        REVOKE ALL PRIVILEGES ON TABLE public.accounting_return_items FROM authenticated;
    END IF;
END $$;

COMMIT;
