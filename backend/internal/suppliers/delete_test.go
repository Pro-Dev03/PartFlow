package suppliers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestDeleteSupplierDoesNotDependOnBusinessHistory(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE suppliers (id TEXT PRIMARY KEY, is_active INTEGER, updated_at TEXT);
		CREATE TABLE purchases (id TEXT PRIMARY KEY, supplier_id TEXT);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, supplier_id TEXT);
		CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT);
	`); err != nil {
		t.Fatal(err)
	}

	supplierID := uuid.New().String()
	if _, err := db.Exec(`INSERT INTO suppliers VALUES (?, 1, CURRENT_TIMESTAMP)`, supplierID); err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{
		`INSERT INTO purchases VALUES (?, ?)`,
		`INSERT INTO supplier_returns VALUES (?, ?)`,
		`INSERT INTO supplier_ledger VALUES (?, ?)`,
	} {
		if _, err := db.Exec(query, uuid.New().String(), supplierID); err != nil {
			t.Fatal(err)
		}
	}

	if err := NewRepository(db).Delete(context.Background(), uuid.MustParse(supplierID)); err != nil {
		t.Fatalf("delete supplier with business history: %v", err)
	}
	var active int
	if err := db.Get(&active, `SELECT is_active FROM suppliers WHERE id = ?`, supplierID); err != nil {
		t.Fatal(err)
	}
	if active != 0 {
		t.Fatalf("supplier was not deleted/deactivated: is_active=%d", active)
	}
}
