package database

import (
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestEnsureRequiredSchemaCreatesMissingObjects(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, item_code TEXT, barcode TEXT, serial_number TEXT, status TEXT)`); err != nil {
		t.Fatalf("create inventory table: %v", err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlite")
	if err := ensureRequiredSchema(sqlxDB); err != nil {
		t.Fatalf("ensureRequiredSchema() error = %v", err)
	}

	if ok, err := columnExists(sqlxDB, "inventory_items", "part_type_id"); err != nil {
		t.Fatalf("columnExists inventory_items: %v", err)
	} else if !ok {
		t.Fatal("expected inventory_items.part_type_id to be created")
	}

	if ok, err := tableExists(sqlxDB, "held_sales"); err != nil {
		t.Fatalf("tableExists held_sales: %v", err)
	} else if !ok {
		t.Fatal("expected held_sales table to be created")
	}
}

func TestTenantIsolationEnabledRequiresExplicitOptIn(t *testing.T) {
	for _, test := range []struct {
		value string
		want  bool
	}{
		{value: "", want: false},
		{value: "false", want: false},
		{value: "true", want: true},
		{value: "1", want: true},
	} {
		t.Setenv("PARTFLOW_TENANT_RLS_ENABLED", test.value)
		if got := TenantIsolationEnabled(); got != test.want {
			t.Errorf("TenantIsolationEnabled() with %q = %t, want %t", test.value, got, test.want)
		}
	}
}
