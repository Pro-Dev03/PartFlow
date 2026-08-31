package products

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func boolPtr(v bool) *bool { return &v }

func TestListProducts_LowStockFilterUsesInventoryItems(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE products (
			id TEXT PRIMARY KEY,
			sku TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			category_id TEXT,
			brand_id TEXT,
			preferred_supplier_id TEXT,
			model TEXT,
			barcode TEXT,
			cost_price REAL DEFAULT 0,
			selling_price REAL DEFAULT 0,
			track_serial INTEGER NOT NULL DEFAULT 0,
			track_individual INTEGER NOT NULL DEFAULT 0,
			min_stock_level INTEGER DEFAULT 0,
			warranty_days INTEGER DEFAULT 0,
			is_active INTEGER NOT NULL DEFAULT 1,
			deleted_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE inventory_items (
			id TEXT PRIMARY KEY,
			product_id TEXT NOT NULL,
			item_code TEXT NOT NULL UNIQUE,
			barcode TEXT,
			serial_number TEXT,
			condition TEXT,
			purchase_cost REAL DEFAULT 0,
			selling_price REAL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'AVAILABLE',
			supplier_id TEXT,
			purchase_date TEXT,
			sold_at TEXT,
			notes TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	productID := uuid.NewString()
	_, err = db.Exec(`
		INSERT INTO products (
			id, sku, name, description, category_id, brand_id, preferred_supplier_id,
			model, barcode, cost_price, selling_price, track_serial, track_individual,
			min_stock_level, warranty_days, is_active, deleted_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NULL, ?, ?)
	`, productID, "SKU-LOW-1", "Low Stock Product", "Test", nil, nil, nil, "Model A", "BAR-LOW-1", 10.0, 25.0, 0, 0, 3, 0, now, now)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}

	_, err = db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, uuid.NewString(), productID, "ITEM-LOW-1", "BAR-LOW-1", "AVAILABLE", now, now)
	if err != nil {
		t.Fatalf("insert inventory item: %v", err)
	}

	repo := NewRepository(db)
	products, total, err := repo.ListProducts(context.Background(), &ProductListRequest{Page: 1, PerPage: 20, LowStockOnly: boolPtr(true)})
	if err != nil {
		t.Fatalf("ListProducts returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
	if products[0].Name != "Low Stock Product" {
		t.Fatalf("unexpected product name: %s", products[0].Name)
	}
}
