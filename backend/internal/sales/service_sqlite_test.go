package sales

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestCreateSaleSQLiteUsesLocalSchema(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT NOT NULL, purchase_price REAL DEFAULT 0);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, purchase_cost REAL DEFAULT 0, condition TEXT, status TEXT, supplier_id TEXT, sold_at TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER DEFAULT 0, updated_at TEXT, created_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, sale_number TEXT, invoice_number TEXT, sale_date TEXT, customer_id TEXT, user_id TEXT, subtotal REAL, tax_amount REAL, discount_amount REAL, total_amount REAL, cost_amount REAL, gross_profit REAL, net_profit REAL, paid_amount REAL, remaining_amount REAL, payment_method TEXT, payment_status TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_price REAL, unit_cost REAL, discount_amount REAL, tax_amount REAL, total_amount REAL, supplier_id TEXT, created_at TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT NOT NULL UNIQUE, sale_id TEXT, customer_id TEXT, amount REAL, payment_method TEXT, payment_status TEXT, created_by TEXT, payment_date TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, amount REAL, balance REAL, reference_id TEXT, description TEXT, created_at TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL DEFAULT 0, credit_limit REAL DEFAULT 0, updated_at TEXT);
		CREATE TABLE acquisition_items (id TEXT PRIMARY KEY, inventory_item_id TEXT, item_status TEXT, updated_at TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, event_type TEXT, event_date TEXT, reference_type TEXT, reference_id TEXT, description TEXT, metadata TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}
	productID, itemID := uuid.New(), uuid.New()
	_, err = db.Exec(`INSERT INTO products (id,name,purchase_price) VALUES ($1,'Widget',10); INSERT INTO inventory_items (id,product_id,purchase_cost,status,created_at,updated_at) VALUES ($2,$1,10,'AVAILABLE',$3,$3); INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES ($4,$1,1,$3,$3)`, productID, itemID, now, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(NewRepository(db), db)
	price := 25.0
	sale, err := svc.CreateSale(context.Background(), uuid.Nil, &CreateSaleRequest{Items: []SaleItemRequest{{ProductID: productID, Quantity: 1, UnitPrice: price}}})
	if err != nil {
		t.Fatalf("CreateSale failed: %v", err)
	}
	if sale.TotalAmount != price {
		t.Fatalf("total = %v, want %v", sale.TotalAmount, price)
	}
	var status string
	if err := db.Get(&status, `SELECT status FROM inventory_items WHERE id = $1`, itemID); err != nil {
		t.Fatal(err)
	}
	if status != "SOLD" {
		t.Fatalf("inventory status = %s, want SOLD", status)
	}
}
