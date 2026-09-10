package purchases

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestReceivePurchaseSkipsDuplicateInventoryItemsOnSQLite(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "purchase-receive.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	statements := []string{
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT NOT NULL, phone TEXT)`,
		`CREATE TABLE users (id TEXT PRIMARY KEY)`,
		`CREATE TABLE products (id TEXT PRIMARY KEY, selling_price REAL DEFAULT 0)`,
		`CREATE TABLE purchases (id TEXT PRIMARY KEY, purchase_number TEXT NOT NULL UNIQUE, supplier_id TEXT, tax_amount REAL DEFAULT 0, total_amount REAL NOT NULL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, status TEXT NOT NULL, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, item_code TEXT UNIQUE, barcode TEXT UNIQUE, serial_number TEXT, condition TEXT, grade TEXT, purchase_cost REAL, selling_price REAL, status TEXT, supplier_id TEXT, purchase_date TEXT, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT, purchase_item_id TEXT, return_number TEXT UNIQUE, supplier_id TEXT, customer_id TEXT, total_amount REAL DEFAULT 0, status TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT NOT NULL, purchase_item_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, total_amount REAL NOT NULL, created_at TEXT NOT NULL)`,
	}
	xdb := sqlx.NewDb(db, "sqlite")
	for _, statement := range statements {
		if _, err := xdb.Exec(statement); err != nil {
			t.Fatalf("create test table: %v", err)
		}
	}

	supplierID := uuid.New()
	productID := uuid.New()
	userID := uuid.New()
	purchaseID := uuid.New()
	now := time.Now()

	if _, err := xdb.Exec(`INSERT INTO suppliers (id, name, phone) VALUES (?, ?, ?)`, supplierID, "Supplier", "000"); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO users (id) VALUES (?)`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO products (id, selling_price) VALUES (?, ?)`, productID, 50.0); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, tax_amount, total_amount, paid_amount, remaining_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID, "INV-duplicate", supplierID, 0.0, 100.0, 100.0, 0.0, "pending", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, uuid.New(), purchaseID, productID, 2, 50.0, 100.0, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, uuid.New(), productID, "ITM-"+purchaseID.String()[:8]+"-001", "BC-"+purchaseID.String()[:8]+"-001", "NEW", 50.0, 50.0, "AVAILABLE", supplierID, now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, uuid.New(), productID, "ITM-"+purchaseID.String()[:8]+"-002", "BC-"+purchaseID.String()[:8]+"-002", "NEW", 50.0, 50.0, "AVAILABLE", supplierID, now, now, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(xdb), xdb)

	purchase, err := service.ReceivePurchase(context.Background(), purchaseID, userID)
	if err != nil {
		t.Fatalf("ReceivePurchase should be idempotent and not fail when inventory already exists: %v", err)
	}
	if purchase.Purchase.Status != "received" {
		t.Fatalf("expected purchase status to become received, got %s", purchase.Purchase.Status)
	}

	var count int
	if err := xdb.Get(&count, `SELECT COUNT(*) FROM inventory_items WHERE item_code LIKE ?`, "ITM-"+purchaseID.String()[:8]+"-%"); err != nil {
		t.Fatalf("count inventory items: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected existing inventory items to remain unchanged, got %d", count)
	}
}

func TestCreatePurchaseWritesLedgerAndAuditAtomicallyOnSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	statements := []string{
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT NOT NULL, phone TEXT)`,
		`CREATE TABLE users (id TEXT PRIMARY KEY)`,
		`CREATE TABLE purchases (id TEXT PRIMARY KEY, purchase_number TEXT NOT NULL UNIQUE, supplier_id TEXT, total_amount REAL NOT NULL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, status TEXT NOT NULL, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT NOT NULL, type TEXT, transaction_type TEXT, amount REAL NOT NULL, balance REAL NOT NULL, description TEXT, reference_id TEXT, reference_type TEXT, created_by TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT NOT NULL, entity_type TEXT NOT NULL, entity_id TEXT NOT NULL, new_values TEXT, created_at TEXT NOT NULL)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("create test table: %v", err)
		}
	}
	supplierID := uuid.New()
	productID := uuid.New()
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO suppliers (id, name, phone) VALUES (?, ?, ?)`, supplierID, "Supplier", "000"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (id) VALUES (?)`, userID); err != nil {
		t.Fatal(err)
	}

	xdb := sqlx.NewDb(db, "sqlite")
	service := NewService(NewRepository(xdb), xdb)
	req := &PurchaseRequest{
		SupplierID:    supplierID,
		InvoiceNumber: "INV-1",
		PurchaseDate:  time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC),
		Items: []PurchaseItemRequest{{
			ProductID: productID,
			Quantity:  2,
			UnitCost:  10,
			Condition: "new",
		}},
	}
	response, err := service.CreatePurchase(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("create purchase: %v", err)
	}
	if response.Purchase.TotalAmount != 20 {
		t.Fatalf("unexpected purchase total: %v", response.Purchase.TotalAmount)
	}

	var ledger struct {
		Type    string  `db:"type"`
		Amount  float64 `db:"amount"`
		Balance float64 `db:"balance"`
	}
	if err := xdb.Get(&ledger, `SELECT type, amount, balance FROM supplier_ledger WHERE reference_id = ?`, response.Purchase.ID); err != nil {
		t.Fatalf("read supplier ledger: %v", err)
	}
	if ledger.Type != "debit" || ledger.Amount != 20 || ledger.Balance != 20 {
		t.Fatalf("unexpected supplier ledger entry: %+v", ledger)
	}
	var auditCount int
	if err := xdb.Get(&auditCount, `SELECT COUNT(*) FROM audit_logs WHERE entity_id = ? AND action = 'CREATE_PURCHASE'`, response.Purchase.ID); err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected one purchase audit log, got %d", auditCount)
	}
}
