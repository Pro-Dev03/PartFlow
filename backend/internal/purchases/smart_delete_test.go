package purchases

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestSmartDeleteDeletesPurchaseWithBusinessReferences(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE purchases (id TEXT PRIMARY KEY, supplier_id TEXT, status TEXT, total_amount REAL, paid_amount REAL);
		CREATE TABLE suppliers (id TEXT PRIMARY KEY, current_balance REAL, updated_at TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, reference_id TEXT);
		CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, reference_type TEXT, reference_id TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, reference_type TEXT, reference_id TEXT);
	`); err != nil {
		t.Fatal(err)
	}

	purchaseID := uuid.New().String()
	supplierID := uuid.New().String()
	if _, err := db.Exec(`INSERT INTO purchases VALUES (?, ?, 'received', 100, 100)`, purchaseID, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO suppliers VALUES (?, 100, CURRENT_TIMESTAMP)`, supplierID); err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{
		`INSERT INTO payments VALUES (?, ?)`,
		`INSERT INTO purchase_items VALUES (?, ?)`,
		`INSERT INTO inventory_movements VALUES (?, 'purchase', ?)`,
		`INSERT INTO item_history VALUES (?, 'purchase', ?)`,
	} {
		if _, err := db.Exec(query, uuid.New().String(), purchaseID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO supplier_ledger VALUES (?, ?, 'debit', 'PURCHASE', 100, ?)`, uuid.New().String(), supplierID, purchaseID); err != nil {
		t.Fatal(err)
	}

	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(purchaseID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("smart delete result = %+v, want deleted", result)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM purchases WHERE id = ?`, purchaseID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("purchase was not deleted: %d", count)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM inventory_movements WHERE reference_id = ?`, purchaseID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("purchase inventory history was not cleaned: %d", count)
	}
}
