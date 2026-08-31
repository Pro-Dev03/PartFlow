package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestRepositoryListInventoryItemsHandlesSQLiteTextTimestamps(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE inventory_items (
			id TEXT PRIMARY KEY,
			product_id TEXT,
			part_type_id TEXT,
			item_code TEXT,
			barcode TEXT,
			serial_number TEXT,
			condition TEXT NOT NULL,
			grade TEXT,
			purchase_cost REAL DEFAULT 0,
			selling_price REAL DEFAULT 0,
			status TEXT NOT NULL,
			location_id TEXT,
			supplier_id TEXT,
			purchase_date TEXT,
			sold_at TEXT,
			notes TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	updatedAt := time.Now().UTC().Add(time.Minute).Format(time.RFC3339)
	productID := uuid.New().String()
	itemID := uuid.New().String()

	_, err = db.Exec(`
		INSERT INTO inventory_items (
			id, product_id, part_type_id, item_code, barcode, serial_number, condition, grade,
			purchase_cost, selling_price, status, location_id, supplier_id, purchase_date, sold_at,
			notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, itemID, productID, nil, "ITEM-001", "BC-001", "SN-001", ConditionNew, GradeExcellent,
		100.0, 150.0, StatusAvailable, nil, nil, nil, nil, nil, createdAt, updatedAt)
	if err != nil {
		t.Fatalf("insert item: %v", err)
	}

	repo := NewRepository(db)
	items, total, err := repo.ListInventoryItems(context.Background(), 10, 0, map[string]interface{}{})
	if err != nil {
		t.Fatalf("ListInventoryItems returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %d, want 1", total)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].CreatedAt.IsZero() {
		t.Fatal("CreatedAt is zero after SQLite text timestamp scan")
	}
	if items[0].UpdatedAt.IsZero() {
		t.Fatal("UpdatedAt is zero after SQLite text timestamp scan")
	}
}
