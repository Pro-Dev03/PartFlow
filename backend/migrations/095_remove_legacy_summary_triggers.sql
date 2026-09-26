-- Remove the old migration-033 row triggers. Summary values are maintained
-- explicitly by the application; these triggers used an incompatible summary
-- schema and made transaction inserts/deletes drift or fail.
DROP TRIGGER IF EXISTS trigger_update_daily_sales_summary ON sales;
DROP TRIGGER IF EXISTS trigger_update_daily_debt_summary ON debts;
DROP FUNCTION IF EXISTS update_daily_sales_summary();
DROP FUNCTION IF EXISTS update_daily_debt_summary();
