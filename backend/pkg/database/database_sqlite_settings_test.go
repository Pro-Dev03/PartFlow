package database

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestEnsureRequiredSchemaCreatesSettingsTableInSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if err := ensureRequiredSchema(db); err != nil {
		t.Fatalf("ensureRequiredSchema returned error: %v", err)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'settings'`); err != nil {
		t.Fatalf("check settings table: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected settings table to exist, got %d rows", count)
	}

	var taxRate string
	if err := db.Get(&taxRate, `SELECT value FROM settings WHERE key = 'tax_rate' LIMIT 1`); err != nil {
		t.Fatalf("expected tax_rate setting row: %v", err)
	}
	if taxRate != "0" {
		t.Fatalf("expected tax_rate default to 0, got %q", taxRate)
	}

	var defaultProfitMargin string
	if err := db.Get(&defaultProfitMargin, `SELECT value FROM settings WHERE key = 'default_profit_margin' LIMIT 1`); err != nil {
		t.Fatalf("expected default_profit_margin setting row: %v", err)
	}
	if defaultProfitMargin != "30" {
		t.Fatalf("expected default_profit_margin default to 30, got %q", defaultProfitMargin)
	}
}
