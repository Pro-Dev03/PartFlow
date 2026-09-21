package sales

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func TestSmartDeleteDeletesCompletedSaleWithoutReversal(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE sales (id TEXT PRIMARY KEY, status TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, reference_type TEXT, reference_id TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, reference_type TEXT, reference_id TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, reference_type TEXT, reference_id TEXT);
	`); err != nil {
		t.Fatal(err)
	}

	saleID := uuid.New().String()
	if _, err := db.Exec(`INSERT INTO sales (id, status) VALUES (?, 'completed')`, saleID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items VALUES (?, ?)`, uuid.New().String(), saleID); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"customer_ledger", "inventory_movements", "item_history"} {
		if _, err := db.Exec("INSERT INTO "+table+" (id, reference_type, reference_id) VALUES (?, 'sale', ?)", uuid.New().String(), saleID); err != nil {
			t.Fatal(err)
		}
	}

	result, err := NewSmartDeleteService(db).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("smart delete result = %+v, want deleted", result)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sales WHERE id = ?`, saleID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("completed sale was not deleted: %d", count)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM inventory_movements WHERE reference_id = ?`, saleID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("sale inventory history was not cleaned: %d", count)
	}
}
