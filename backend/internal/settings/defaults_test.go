package settings

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestNewHandlerSeedsPOSAndPaymentSettings(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE settings (
			id TEXT PRIMARY KEY,
			key TEXT UNIQUE,
			value TEXT,
			value_type TEXT CHECK (value_type IN ('string', 'number', 'boolean', 'json')),
			category TEXT,
			description TEXT,
			is_public INTEGER,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`); err != nil {
		t.Fatal(err)
	}

	NewHandler(db)

	keys := []string{
		"pos_products_per_page",
		"pos_product_view_mode",
		"electronic_payments_enabled",
		"payment_provider",
		"payment_methods",
	}
	for _, key := range keys {
		var value string
		if err := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value); err != nil {
			t.Fatalf("setting %q was not seeded: %v", key, err)
		}
		if value == "" {
			t.Fatalf("setting %q has an empty default", key)
		}
	}
}
