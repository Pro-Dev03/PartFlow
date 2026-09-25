package purchases

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestSmartDeleteReceivedPurchaseReversesStockAndRemovesTransaction(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE purchases (id TEXT PRIMARY KEY, invoice_number TEXT, supplier_id TEXT, status TEXT, total_amount REAL, paid_amount REAL);
		CREATE TABLE suppliers (id TEXT PRIMARY KEY, current_balance REAL, updated_at TEXT);
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, reference_id TEXT);
		CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT, product_id TEXT, quantity INTEGER);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, item_code TEXT, product_id TEXT, barcode TEXT, status TEXT, condition TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT UNIQUE, quantity INTEGER, created_at TEXT, updated_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, reference_type TEXT, reference_id TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, reference_type TEXT, reference_id TEXT);
		CREATE TABLE returns (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE barcodes (id TEXT PRIMARY KEY, inventory_item_id TEXT);
		CREATE TABLE reservations (id TEXT PRIMARY KEY, item_id TEXT);
		CREATE TABLE acquisition_items (id TEXT PRIMARY KEY, inventory_item_id TEXT);
		CREATE TABLE inspection_items (id TEXT PRIMARY KEY, item_id TEXT);
		CREATE TABLE item_repair_costs (id TEXT PRIMARY KEY, inventory_item_id TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}
	purchaseID, supplierID, productID, inventoryItemID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	itemCode := "ITM-" + purchaseID.String()[:8] + "-001"
	_, err = db.Exec(`INSERT INTO purchases VALUES (?, 'PUR-DEL-01', ?, 'received', 100, 100)`, purchaseID.String(), supplierID.String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO suppliers VALUES (?, 100, CURRENT_TIMESTAMP)`, supplierID.String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO products VALUES (?, 'Test part')`, productID.String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO purchase_items VALUES (?, ?, ?, 1)`, uuid.NewString(), purchaseID.String(), productID.String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory_items VALUES (?, ?, ?, 'BC-PUR-1', 'AVAILABLE', 'NEW')`, inventoryItemID.String(), itemCode, productID.String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory VALUES (?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, uuid.NewString(), productID.String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory_movements VALUES (?, ?, ?, 'PURCHASE', 'purchase', ?)`, uuid.NewString(), inventoryItemID.String(), productID.String(), purchaseID.String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO payments VALUES (?, ?)`, uuid.NewString(), purchaseID.String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO supplier_ledger VALUES (?, ?, 'debit', 'PURCHASE', 100, ?)`, uuid.NewString(), supplierID.String(), purchaseID.String())
	if err != nil {
		t.Fatal(err)
	}

	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), purchaseID, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("purchase deletion result = %+v", result)
	}
	var purchases, inventoryItems, movements, ledgers, audits int
	var quantity int
	var supplierBalance float64
	for _, check := range []struct {
		query string
		dest  *int
	}{
		{`SELECT COUNT(*) FROM purchases WHERE id=?`, &purchases},
		{`SELECT COUNT(*) FROM inventory_items WHERE id=?`, &inventoryItems},
		{`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=?`, &movements},
		{`SELECT COUNT(*) FROM supplier_ledger WHERE reference_id=?`, &ledgers},
		{`SELECT COUNT(*) FROM audit_logs WHERE entity_id=? AND action='DELETE'`, &audits},
	} {
		if err := db.QueryRow(check.query, purchaseID.String()).Scan(check.dest); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.QueryRow(`SELECT quantity FROM inventory WHERE product_id=?`, productID.String()).Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT current_balance FROM suppliers WHERE id=?`, supplierID.String()).Scan(&supplierBalance); err != nil {
		t.Fatal(err)
	}
	if purchases != 0 || inventoryItems != 0 || movements != 0 || ledgers != 0 || audits != 1 || quantity != 0 || supplierBalance != 0 {
		t.Fatalf("purchase=%d item=%d movement=%d ledger=%d audit=%d stock=%d supplier_balance=%v", purchases, inventoryItems, movements, ledgers, audits, quantity, supplierBalance)
	}
}

func TestSmartDeleteBlocksCancelledAndReversedPurchases(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE purchases (id TEXT PRIMARY KEY, status TEXT, total_amount REAL, paid_amount REAL);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	reversedID, cancelledID := uuid.New(), uuid.New()
	for _, purchase := range []struct {
		id     uuid.UUID
		status string
	}{{reversedID, "reversed"}, {cancelledID, "cancelled"}} {
		if _, err := db.Exec(`INSERT INTO purchases (id, status, total_amount, paid_amount) VALUES (?, ?, 100, 0)`, purchase.id, purchase.status); err != nil {
			t.Fatal(err)
		}
	}

	service := NewSmartDeleteService(sqlx.NewDb(db, "sqlite"))
	for _, purchaseID := range []uuid.UUID{reversedID, cancelledID} {
		result, err := service.SmartDelete(context.Background(), purchaseID, uuid.Nil)
		if err != nil {
			t.Fatal(err)
		}
		if result.Action != "blocked" || result.CanProceed {
			t.Fatalf("purchase %s deletion result = %+v, want blocked", purchaseID, result)
		}
	}
	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM purchases`).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 2 {
		t.Fatalf("cancelled/reversed purchases remaining = %d, want 2", remaining)
	}
}
