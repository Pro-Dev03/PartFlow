package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestStripTransactionControlStatements(t *testing.T) {
	input := "BEGIN;\nCREATE TABLE demo (id INTEGER);\nCOMMIT;\n"
	want := "CREATE TABLE demo (id INTEGER);\n"
	if got := stripTransactionControlStatements(input); got != want {
		t.Fatalf("stripTransactionControlStatements() = %q, want %q", got, want)
	}
}

func TestStripTransactionControlStatementsKeepsProceduralBlocks(t *testing.T) {
	input := "DO $$\nBEGIN\n  PERFORM 1;\nEND $$;\n"
	if got := stripTransactionControlStatements(input); got != input {
		t.Fatalf("procedural block was modified: got %q, want %q", got, input)
	}
}

func TestFindAppliedMigrationRecognizesLegacyExtension(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES ('001_initial_schema.sql', 'now')`); err != nil {
		t.Fatal(err)
	}
	version, appliedAt, err := findAppliedMigration(db, "001_initial_schema", "001_initial_schema.sql")
	if err != nil {
		t.Fatalf("findAppliedMigration: %v", err)
	}
	if version != "001_initial_schema.sql" || appliedAt != "now" {
		t.Fatalf("got version=%q appliedAt=%q", version, appliedAt)
	}
}

func TestSingleStoreDefaultSkipsTenantIsolationMigration(t *testing.T) {
	if !shouldSkipDefaultMigration("080_tenant_isolation.sql") {
		t.Fatal("default single-store migration run must skip tenant isolation")
	}
	if shouldSkipDefaultMigration("081_single_store_cloud_hardening.sql") {
		t.Fatal("single-store cloud hardening migration must remain in the default run")
	}
	if shouldSkipDefaultMigration("082_remove_duplicate_indexes.sql") {
		t.Fatal("single-store duplicate index cleanup must remain in the default run")
	}
}

func TestDuplicateIndexMigrationDropsOnlyRedundantSupabaseIndexes(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "082_remove_duplicate_indexes.sql")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read duplicate index cleanup migration: %v", err)
	}
	sql := string(contents)

	for _, redundantIndex := range []string{
		"idx_daily_debt_summary_runtime_date",
		"idx_daily_inventory_summary_runtime_date",
		"idx_daily_profit_summary_runtime_date",
		"idx_daily_sales_summary_runtime_date",
		"idx_monthly_debt_summary_runtime_year_month",
		"idx_monthly_inventory_summary_runtime_year_month",
		"idx_monthly_profit_summary_runtime_year_month",
		"idx_monthly_sales_summary_runtime_year_month",
		"idx_supplier_ledger_runtime_type",
		"idx_supplier_ledger_type_normalized",
	} {
		if !strings.Contains(sql, "DROP INDEX IF EXISTS public."+redundantIndex) {
			t.Errorf("migration does not remove redundant index %q", redundantIndex)
		}
	}

	for _, retainedIndex := range []string{
		"idx_daily_debt_date", "idx_daily_inventory_date", "idx_daily_profit_date", "idx_daily_sales_date",
		"idx_monthly_debt_year_month", "idx_monthly_inventory_year_month", "idx_monthly_profit_year_month", "idx_monthly_sales_year_month",
		"idx_supplier_ledger_type",
	} {
		if strings.Contains(sql, "DROP INDEX IF EXISTS public."+retainedIndex+";") {
			t.Errorf("migration drops canonical index %q", retainedIndex)
		}
	}
}

func TestSingleStoreHardeningMigrationCoversSupabaseFindings(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "081_single_store_cloud_hardening.sql")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read single-store hardening migration: %v", err)
	}
	sql := string(contents)

	for _, view := range []string{
		"customer_ledger_view", "supplier_ledger_view", "inventory_ledger_view",
		"seller_balances", "customer_acquisition_summary", "used_parts_aging",
		"returns_summary", "monthly_returns_analysis", "sales_returns_analysis",
	} {
		if !strings.Contains(sql, "'"+view+"'") {
			t.Errorf("migration does not include linter-reported view %q", view)
		}
	}
	for _, function := range []string{
		"update_updated_at_column", "generate_item_code", "generate_internal_barcode",
		"cleanup_expired_idempotency_keys", "sync_role_permissions", "update_customer_balance",
		"update_supplier_balance", "update_inventory_quantity", "calculate_acquisition_total",
		"create_item_history_entry", "generate_return_number", "handle_return_debt_adjustment",
		"validate_return_quantity", "update_return_item_inventory_status",
		"update_inventory_current_state", "update_daily_sales_summary", "update_monthly_sales_summary",
	} {
		if !strings.Contains(sql, "'"+function+"'") {
			t.Errorf("migration does not include linter-reported function %q", function)
		}
	}

	for _, required := range []string{
		"security_invoker = true",
		"REVOKE ALL PRIVILEGES ON TABLE public.%I FROM anon",
		"REVOKE ALL PRIVILEGES ON TABLE public.%I FROM authenticated",
		"SET search_path TO pg_catalog, public, pg_temp",
		"REVOKE CREATE ON SCHEMA public FROM PUBLIC",
		"REVOKE EXECUTE ON FUNCTION %s FROM service_role",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("migration is missing security hardening %q", required)
		}
	}
	if !strings.Contains(sql, "p.proname = 'rls_auto_enable'") {
		t.Fatal("migration does not include the exposed rls_auto_enable function")
	}
}
