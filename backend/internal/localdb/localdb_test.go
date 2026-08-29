package localdb

import (
	"path/filepath"
	"testing"
)

func TestOpenInitializesLocalDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partflow.db")
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", path)

	database, err := Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer database.DB.Close()

	if database.Path != path {
		t.Fatalf("database path = %q, want %q", database.Path, path)
	}

	if err := SetMetadata(database.DB, "operating_mode", "offline"); err != nil {
		t.Fatalf("SetMetadata() error = %v", err)
	}

	var mode string
	if err := database.DB.QueryRow("SELECT value FROM local_metadata WHERE key = ?", "operating_mode").Scan(&mode); err != nil {
		t.Fatalf("read operating mode error = %v", err)
	}
	if mode != "offline" {
		t.Fatalf("operating mode = %q, want offline", mode)
	}

	// Verify all business tables exist
	requiredTables := []string{
		"local_metadata", "sync_queue",
		"categories", "products", "customers", "suppliers",
		"inventory_items", "sales", "sale_items",
		"purchases", "purchase_items", "payments", "debts",
	}

	for _, table := range requiredTables {
		var exists int
		if err := database.DB.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&exists); err != nil || exists != 1 {
			t.Fatalf("table %s does not exist", table)
		}
	}
}
