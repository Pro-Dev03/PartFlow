-- Return debt, payment and ledger effects are applied atomically by the
-- application service. Older deployments may still have the legacy row
-- triggers from migration 034 because editing an already-applied migration
-- does not rerun it. Drop those triggers and their now-unused functions.
DO $$
BEGIN
    IF to_regclass('public.returns') IS NOT NULL THEN
        DROP TRIGGER IF EXISTS handle_return_debt_trigger ON public.returns;
        DROP TRIGGER IF EXISTS adjust_debt_trigger ON public.returns;
    END IF;
END;
$$;

DROP FUNCTION IF EXISTS public.handle_return_debt_adjustment();
DROP FUNCTION IF EXISTS public.adjust_debt_for_return();
