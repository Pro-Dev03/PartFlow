package main

import (
	"database/sql"
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
