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

func TestInventoryItemFromMapHandlesNonStringScalars(t *testing.T) {
	id := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	productID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	locationID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	createdAt := "2024-01-02T03:04:05Z"
	updatedAt := "2024-01-03T03:04:05Z"

	record := map[string]any{
		"id":                 []byte(id.String()),
		"product_id":         []byte(productID.String()),
		"part_type_id":       nil,
		"item_code":          []byte("ITEM-001"),
		"barcode":            []byte("BC-001"),
		"serial_number":      []byte("SN-001"),
		"condition":          "NEW",
		"grade":              "EXCELLENT",
		"purchase_cost":      float64(99.5),
		"selling_price":      int64(120),
		"status":             "AVAILABLE",
		"location_id":        []byte(locationID.String()),
		"supplier_id":        nil,
		"purchase_date":      createdAt,
		"sold_at":            nil,
		"notes":              []byte("ok"),
		"created_at":         createdAt,
		"updated_at":         updatedAt,
		"current_quantity":   int64(2),
		"available_quantity": int64(1),
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("inventoryItemFromMap panicked: %v", r)
		}
	}()

	item, err := inventoryItemFromMap(record)
	if err != nil {
		t.Fatalf("inventoryItemFromMap returned error: %v", err)
	}
	if item.ID != id {
		t.Fatalf("item ID mismatch: got %s, want %s", item.ID, id)
	}
	if item.ProductID == nil || *item.ProductID != productID {
		t.Fatalf("product ID mismatch: %#v", item.ProductID)
	}
	if item.LocationID == nil || *item.LocationID != locationID {
		t.Fatalf("location ID mismatch: %#v", item.LocationID)
	}
	if item.PurchaseCost != 99.5 {
		t.Fatalf("purchase_cost = %v, want 99.5", item.PurchaseCost)
	}
	if item.CurrentQuantity != 2 || item.AvailableQuantity != 1 {
		t.Fatalf("quantities incorrect: %+v", item)
	}
}
