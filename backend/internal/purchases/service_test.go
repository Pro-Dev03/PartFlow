package purchases

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/supplierreturns"
	_ "modernc.org/sqlite"
)

func TestReceivePurchaseUsesUniqueSerialNumbersForDuplicateInputOnSQLite(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "purchase-serial-duplicate.sqlite")
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
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, serial_number TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, category_id TEXT, item_code TEXT UNIQUE, barcode TEXT UNIQUE, serial_number TEXT UNIQUE, condition TEXT, grade TEXT, purchase_cost REAL, selling_price REAL, status TEXT, supplier_id TEXT, purchase_date TEXT, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT, purchase_item_id TEXT, return_number TEXT UNIQUE, supplier_id TEXT, status TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
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
	if _, err := xdb.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, tax_amount, total_amount, paid_amount, remaining_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID, "INV-serial-duplicate", supplierID, 0.0, 100.0, 100.0, 0.0, "pending", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, serial_number, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, uuid.New(), purchaseID, productID, 2, 50.0, 100.0, "SN-DUPLICATE", now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(xdb), xdb)
	purchase, err := service.ReceivePurchase(context.Background(), purchaseID, userID)
	if err != nil {
		t.Fatalf("ReceivePurchase should create unique serials when the input serial is reused across a multi-quantity item: %v", err)
	}
	if purchase.Purchase.Status != "received" {
		t.Fatalf("expected purchase status to become received, got %s", purchase.Purchase.Status)
	}

	var count int
	if err := xdb.Get(&count, `SELECT COUNT(*) FROM inventory_items WHERE product_id = ?`, productID); err != nil {
		t.Fatalf("count inventory items: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 inventory rows created, got %d", count)
	}

	var serials []string
	if err := xdb.Select(&serials, `SELECT serial_number FROM inventory_items WHERE product_id = ? ORDER BY created_at`, productID); err != nil {
		t.Fatalf("read serial numbers: %v", err)
	}
	if len(serials) != 2 {
		t.Fatalf("expected 2 serial numbers, got %d", len(serials))
	}
	if serials[0] == serials[1] {
		t.Fatalf("expected unique serial numbers after receive, got duplicate values: %v", serials)
	}
}

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
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, serial_number TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, category_id TEXT, item_code TEXT UNIQUE, barcode TEXT UNIQUE, serial_number TEXT, condition TEXT, grade TEXT, purchase_cost REAL, selling_price REAL, status TEXT, supplier_id TEXT, purchase_date TEXT, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
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

func TestReceivePurchaseUpdatesEachProductForMultiItemPurchaseOnSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "purchase-multi-item.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	xdb := sqlx.NewDb(db, "sqlite")
	for _, statement := range []string{
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT NOT NULL, phone TEXT)`,
		`CREATE TABLE users (id TEXT PRIMARY KEY)`,
		`CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, selling_price REAL DEFAULT 0)`,
		`CREATE TABLE purchases (id TEXT PRIMARY KEY, purchase_number TEXT NOT NULL UNIQUE, supplier_id TEXT, tax_amount REAL DEFAULT 0, total_amount REAL NOT NULL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, status TEXT NOT NULL, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, serial_number TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, quantity INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, category_id TEXT, item_code TEXT UNIQUE, barcode TEXT UNIQUE, serial_number TEXT, condition TEXT, grade TEXT, purchase_cost REAL, selling_price REAL, status TEXT, supplier_id TEXT, purchase_date TEXT, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT, status TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT NOT NULL, purchase_item_id TEXT NOT NULL, quantity INTEGER NOT NULL, created_at TEXT NOT NULL)`,
	} {
		if _, err := xdb.Exec(statement); err != nil {
			t.Fatalf("create test table: %v", err)
		}
	}

	supplierID, userID := uuid.New(), uuid.New()
	productA, productB := uuid.New(), uuid.New()
	purchaseID := uuid.New()
	now := time.Now()
	for _, args := range [][3]interface{}{
		{supplierID, "Supplier", "000"},
	} {
		if _, err := xdb.Exec(`INSERT INTO suppliers (id, name, phone) VALUES (?, ?, ?)`, args[:]...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := xdb.Exec(`INSERT INTO users (id) VALUES (?)`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO products (id, name, selling_price) VALUES (?, ?, ?), (?, ?, ?)`, productA, "Product A", 100, productB, "Product B", 200); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, total_amount, paid_amount, remaining_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID, "INV-MULTI", supplierID, 1100, 1, 1099, "pending", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?)`, uuid.New(), purchaseID, productA, 5, 100, 500, now, uuid.New(), purchaseID, productB, 3, 200, 600, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES (?, ?, 0, ?, ?), (?, ?, 0, ?, ?)`, uuid.New(), productA, now, now, uuid.New(), productB, now, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(xdb), xdb)
	received, err := service.ReceivePurchase(context.Background(), purchaseID, userID)
	if err != nil {
		t.Fatalf("receive multi-item purchase: %v", err)
	}
	if len(received.Items) != 2 || received.Items[0].ProductName != "Product A" || received.Items[1].ProductName != "Product B" {
		t.Fatalf("received purchase item names = %#v, want Product A and Product B", received.Items)
	}

	for _, check := range []struct {
		productID uuid.UUID
		quantity  int
	}{
		{productA, 5},
		{productB, 3},
	} {
		var aggregate, itemCount int
		if err := xdb.Get(&aggregate, `SELECT quantity FROM inventory WHERE product_id = ?`, check.productID); err != nil {
			t.Fatal(err)
		}
		if err := xdb.Get(&itemCount, `SELECT COUNT(*) FROM inventory_items WHERE product_id = ?`, check.productID); err != nil {
			t.Fatal(err)
		}
		if aggregate != check.quantity || itemCount != check.quantity {
			t.Fatalf("product %s received aggregate=%d items=%d, want %d", check.productID, aggregate, itemCount, check.quantity)
		}
	}
}

func TestListSummariesHandlesNullPurchaseDateOnSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	statements := []string{
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT NOT NULL, phone TEXT)`,
		`CREATE TABLE purchases (id TEXT PRIMARY KEY, purchase_number TEXT NOT NULL UNIQUE, supplier_id TEXT, purchase_date TEXT, tax_amount REAL DEFAULT 0, total_amount REAL NOT NULL, paid_amount REAL DEFAULT 0, status TEXT NOT NULL, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, serial_number TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, item_code TEXT, status TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT, supplier_return_id TEXT, status TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT NOT NULL, purchase_item_id TEXT NOT NULL, quantity INTEGER NOT NULL, created_at TEXT NOT NULL)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("create test table: %v", err)
		}
	}

	supplierID := uuid.New()
	purchaseID := uuid.New()
	if _, err := db.Exec(`INSERT INTO suppliers (id, name) VALUES (?, ?)`, supplierID, "Supplier"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, purchase_date, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID, "INV-NULL-DATE", supplierID, nil, 100.0, 0.0, "received", nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, serial_number, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, uuid.New(), purchaseID, uuid.New(), 10, 10.0, 100.0, nil, time.Now()); err != nil {
		t.Fatal(err)
	}

	xdb := sqlx.NewDb(db, "sqlite")
	repo := NewRepository(xdb)
	items, total, err := repo.ListSummaries(context.Background(), PurchaseListRequest{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("ListSummaries should tolerate NULL purchase dates on SQLite: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected one summary row, got total=%d len=%d", total, len(items))
	}
	if items[0].PurchaseDate.IsZero() {
		t.Fatalf("expected a non-zero purchase date fallback when the DB value is NULL")
	}
	if items[0].TotalItems != 10 {
		t.Fatalf("expected total item quantity 10, got %d", items[0].TotalItems)
	}
}

func TestCreatePurchaseAndReceivePreserveBarcodeContinuityOnSQLite(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "purchase-barcode-continuity.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	statements := []string{
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT NOT NULL, phone TEXT)`,
		`CREATE TABLE users (id TEXT PRIMARY KEY)`,
		`CREATE TABLE products (id TEXT PRIMARY KEY, barcode TEXT, selling_price REAL DEFAULT 0, cost_price REAL DEFAULT 0)`,
		`CREATE TABLE purchases (id TEXT PRIMARY KEY, purchase_number TEXT NOT NULL UNIQUE, supplier_id TEXT, tax_amount REAL DEFAULT 0, total_amount REAL NOT NULL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, status TEXT NOT NULL, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, product_id TEXT NOT NULL, barcode TEXT, serial_number TEXT, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, category_id TEXT, item_code TEXT UNIQUE, barcode TEXT UNIQUE, serial_number TEXT, condition TEXT, grade TEXT, purchase_cost REAL, selling_price REAL, status TEXT, supplier_id TEXT, purchase_date TEXT, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT, purchase_item_id TEXT, return_number TEXT UNIQUE, supplier_id TEXT, customer_id TEXT, total_amount REAL DEFAULT 0, status TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT NOT NULL, purchase_item_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, total_amount REAL NOT NULL, created_at TEXT NOT NULL)`,
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
	productBarcode := "FNX-GPU-000421"
	now := time.Now()
	if _, err := db.Exec(`INSERT INTO suppliers (id, name, phone) VALUES (?, ?, ?)`, supplierID, "Supplier", "000"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (id) VALUES (?)`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id, barcode, selling_price, cost_price) VALUES (?, ?, ?, ?)`, productID, productBarcode, 100.0, 60.0); err != nil {
		t.Fatal(err)
	}

	xdb := sqlx.NewDb(db, "sqlite")
	service := NewService(NewRepository(xdb), xdb)
	req := &PurchaseRequest{
		SupplierID:    supplierID,
		InvoiceNumber: "INV-BARCODE-1",
		PurchaseDate:  time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC),
		Items: []PurchaseItemRequest{{
			ProductID: productID,
			Barcode:   productBarcode,
			Quantity:  1,
			UnitCost:  60,
			Condition: "new",
		}},
	}
	response, err := service.CreatePurchase(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("create purchase: %v", err)
	}
	if _, err := db.Exec(`UPDATE purchases SET paid_amount = total_amount, remaining_amount = 0 WHERE id = ?`, response.Purchase.ID); err != nil {
		t.Fatalf("mark purchase as paid: %v", err)
	}

	var storedBarcode string
	if err := xdb.Get(&storedBarcode, `SELECT barcode FROM purchase_items WHERE purchase_id = ? LIMIT 1`, response.Purchase.ID); err != nil {
		t.Fatalf("read stored purchase barcode: %v", err)
	}
	if storedBarcode != productBarcode {
		t.Fatalf("purchase item barcode mismatch: got %q want %q", storedBarcode, productBarcode)
	}

	purchase, err := service.GetPurchase(context.Background(), response.Purchase.ID)
	if err != nil {
		t.Fatalf("get purchase: %v", err)
	}
	if purchase == nil || purchase.Purchase.ID == uuid.Nil {
		t.Fatal("purchase missing after create")
	}

	if _, err := service.ReceivePurchase(context.Background(), response.Purchase.ID, userID); err != nil {
		t.Fatalf("receive purchase: %v", err)
	}

	var inventoryBarcode string
	if err := xdb.Get(&inventoryBarcode, `SELECT barcode FROM inventory_items WHERE product_id = ? LIMIT 1`, productID); err != nil {
		t.Fatalf("read inventory barcode: %v", err)
	}
	if inventoryBarcode != productBarcode {
		t.Fatalf("inventory barcode should preserve purchase barcode, got %q want %q", inventoryBarcode, productBarcode)
	}

	var itemCount int
	if err := xdb.Get(&itemCount, `SELECT COUNT(*) FROM inventory_items WHERE product_id = ?`, productID); err != nil {
		t.Fatalf("count inventory rows: %v", err)
	}
	if itemCount != 1 {
		t.Fatalf("expected 1 received item, got %d", itemCount)
	}

	_ = now
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
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, serial_number TEXT, created_at TEXT NOT NULL)`,
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

func TestPurchaseLifecycleSupplierBalanceAndReturnLedgerSQLite(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "purchase-lifecycle.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	xdb := sqlx.NewDb(db, "sqlite")
	for _, statement := range []string{
		`CREATE TABLE users (id TEXT PRIMARY KEY)`,
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, code TEXT, name TEXT, phone TEXT, current_balance REAL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE products (id TEXT PRIMARY KEY, sku TEXT, name TEXT, selling_price REAL DEFAULT 0, cost_price REAL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE purchases (id TEXT PRIMARY KEY, purchase_number TEXT NOT NULL UNIQUE, supplier_id TEXT NOT NULL, purchase_date TEXT, tax_amount REAL DEFAULT 0, total_amount REAL NOT NULL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, status TEXT NOT NULL, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, barcode TEXT, serial_number TEXT, condition TEXT, grade TEXT, notes TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, quantity INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, category_id TEXT, item_code TEXT NOT NULL, barcode TEXT, serial_number TEXT, condition TEXT, grade TEXT, purchase_cost REAL DEFAULT 0, selling_price REAL DEFAULT 0, status TEXT NOT NULL, supplier_id TEXT, purchase_date TEXT, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT NOT NULL UNIQUE, purchase_id TEXT, supplier_id TEXT, amount REAL NOT NULL, payment_method TEXT, payment_status TEXT, created_by TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT NOT NULL, type TEXT, transaction_type TEXT, amount REAL NOT NULL, balance REAL NOT NULL, description TEXT, reference_id TEXT, reference_type TEXT, created_by TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL, supplier_id TEXT NOT NULL, return_number TEXT NOT NULL UNIQUE, status TEXT NOT NULL, reason TEXT NOT NULL, refund_amount REAL DEFAULT 0, notes TEXT, created_by TEXT, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT NOT NULL, purchase_item_id TEXT NOT NULL, product_id TEXT NOT NULL, inventory_item_id TEXT, quantity INTEGER NOT NULL, unit_cost REAL NOT NULL, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT NOT NULL, entity_type TEXT NOT NULL, entity_id TEXT NOT NULL, new_values TEXT, created_at TEXT NOT NULL)`,
	} {
		if _, err := xdb.Exec(statement); err != nil {
			t.Fatalf("create table: %v (%s)", err, statement)
		}
	}

	ctx := context.Background()
	supplierID := uuid.New()
	productID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()
	purchaseDate := now.Add(-time.Hour)

	if _, err := xdb.Exec(`INSERT INTO suppliers (id, code, name, phone, current_balance, created_at, updated_at) VALUES (?, ?, ?, ?, 0, ?, ?)`, supplierID, "SUP-LIFE", "Lifecycle Supplier", "000", now.Format(time.RFC3339), now.Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO users (id) VALUES (?)`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO products (id, sku, name, selling_price, cost_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, productID, "SKU-001", "Lifecycle Product", 150.0, 100.0, now.Format(time.RFC3339), now.Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}

	purchaseSvc := NewService(NewRepository(xdb), xdb)
	returnSvc := supplierreturns.NewService(xdb)

	response, err := purchaseSvc.CreatePurchase(ctx, userID, &PurchaseRequest{
		SupplierID:    supplierID,
		InvoiceNumber: "INV-LIFECYCLE-001",
		PurchaseDate:  purchaseDate,
		Items: []PurchaseItemRequest{{
			ProductID:    productID,
			Quantity:     3,
			UnitCost:     100,
			SellingPrice: 150,
			Condition:    "new",
		}},
	})
	if err != nil {
		t.Fatalf("create purchase: %v", err)
	}
	if response.Purchase.TotalAmount != 300 {
		t.Fatalf("purchase total = %v, want 300", response.Purchase.TotalAmount)
	}
	if response.Purchase.PaidAmount != 0 || response.Remaining != 300 {
		t.Fatalf("unpaid projection = paid %.2f remaining %.2f, want 0/300", response.Purchase.PaidAmount, response.Remaining)
	}

	partial, err := purchaseSvc.AddPayment(ctx, response.Purchase.ID, userID, 150, "cash")
	if err != nil {
		t.Fatalf("partial payment: %v", err)
	}
	if partial.Purchase.PaidAmount != 150 || partial.Remaining != 150 {
		t.Fatalf("partial payment projection = paid %.2f remaining %.2f, want 150/150", partial.Purchase.PaidAmount, partial.Remaining)
	}
	full, err := purchaseSvc.AddPayment(ctx, response.Purchase.ID, userID, 150, "cash")
	if err != nil {
		t.Fatalf("full payment: %v", err)
	}
	if full.Purchase.PaidAmount != 300 || full.Remaining != 0 {
		t.Fatalf("full payment projection = paid %.2f remaining %.2f, want 300/0", full.Purchase.PaidAmount, full.Remaining)
	}
	if _, err := purchaseSvc.AddPayment(ctx, response.Purchase.ID, userID, 0.01, "cash"); !errors.Is(err, ErrPaymentExceedsTotal) {
		t.Fatalf("overpayment error = %v, want %v", err, ErrPaymentExceedsTotal)
	}
	var paymentCount int
	if err := xdb.Get(&paymentCount, `SELECT COUNT(*) FROM payments WHERE purchase_id = ?`, response.Purchase.ID); err != nil {
		t.Fatalf("read payments after rejected overpayment: %v", err)
	}
	if paymentCount != 2 {
		t.Fatalf("rejected overpayment created a payment: count=%d, want 2", paymentCount)
	}
	if _, err := purchaseSvc.ReceivePurchase(ctx, response.Purchase.ID, userID); err != nil {
		t.Fatalf("receive purchase: %v", err)
	}

	var balance float64
	if err := xdb.Get(&balance, `SELECT current_balance FROM suppliers WHERE id = ?`, supplierID); err != nil {
		t.Fatalf("read supplier balance after partial payment: %v", err)
	}
	if balance != 0 {
		t.Fatalf("supplier balance after full payment = %v, want 0", balance)
	}

	var purchaseItemID string
	if err := xdb.Get(&purchaseItemID, `SELECT id FROM purchase_items WHERE purchase_id = ? LIMIT 1`, response.Purchase.ID); err != nil {
		t.Fatalf("read purchase item id: %v", err)
	}

	created, err := returnSvc.Create(ctx, uuid.New(), supplierreturns.CreateRequest{PurchaseID: response.Purchase.ID, Reason: "defective"})
	if err != nil {
		t.Fatalf("create supplier return: %v", err)
	}
	if err := returnSvc.AddItem(ctx, created.ID, supplierreturns.AddItemRequest{PurchaseItemID: uuid.MustParse(purchaseItemID), Quantity: 1}); err != nil {
		t.Fatalf("add supplier return item: %v", err)
	}
	if err := returnSvc.Complete(ctx, created.ID, userID); err != nil {
		t.Fatalf("complete supplier return: %v", err)
	}

	if err := xdb.Get(&balance, `SELECT current_balance FROM suppliers WHERE id = ?`, supplierID); err != nil {
		t.Fatalf("read supplier balance after supplier return: %v", err)
	}
	if balance != -100 {
		t.Fatalf("supplier balance after supplier return = %v, want -100", balance)
	}

	var ledgerCount int
	if err := xdb.Get(&ledgerCount, `SELECT COUNT(*) FROM supplier_ledger WHERE supplier_id = ? AND transaction_type IN ('PURCHASE', 'PAYMENT', 'SUPPLIER_RETURN')`, supplierID); err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 4 {
		t.Fatalf("ledger entry count = %d, want 4 (purchase + two payments + supplier return)", ledgerCount)
	}

	var refundAmount float64
	if err := xdb.Get(&refundAmount, `SELECT refund_amount FROM supplier_returns WHERE id = ?`, created.ID); err != nil {
		t.Fatal(err)
	}
	if refundAmount != 100 {
		t.Fatalf("supplier return refund_amount = %v, want 100", refundAmount)
	}
}
