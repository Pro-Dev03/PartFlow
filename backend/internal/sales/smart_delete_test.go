package sales

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func newSaleDeleteTestDB(t *testing.T) (*sql.DB, string, string, string) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE sales (id TEXT PRIMARY KEY, invoice_number TEXT, customer_id TEXT, status TEXT, total_amount REAL, paid_amount REAL DEFAULT 0, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, quantity INTEGER);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT, sold_at TEXT, updated_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, reference_type TEXT, reference_id TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT UNIQUE, quantity INTEGER, created_at TEXT, updated_at TEXT);
		CREATE TABLE returns (id TEXT PRIMARY KEY, sale_id TEXT);
		CREATE TABLE payment_transactions (id TEXT PRIMARY KEY, sale_id TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, sale_id TEXT, paid_amount REAL);
		CREATE TABLE payments (id TEXT PRIMARY KEY, sale_id TEXT, customer_id TEXT);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, amount REAL, reference_id TEXT, reference_type TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL, updated_at TEXT);
		CREATE TABLE sale_payment_allocations (id TEXT PRIMARY KEY, sale_id TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, reference_id TEXT, reference_type TEXT);
		CREATE TABLE ledger_entries (id TEXT PRIMARY KEY, reference_id TEXT, reference_type TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}
	saleID, productID, itemID, customerID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	_, err = db.Exec(`INSERT INTO sales (id, invoice_number, customer_id, status, total_amount, paid_amount) VALUES (?, 'INV-DEL-01', ?, 'completed', 120, 0)`, saleID, customerID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO sale_items VALUES (?, ?, ?, 1)`, uuid.NewString(), saleID, productID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory_items VALUES (?, ?, 'SOLD', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, itemID, productID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory_movements VALUES (?, ?, NULL, 'SALE', -1, 'sale', ?)`, uuid.NewString(), itemID, saleID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory VALUES (?, ?, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, uuid.NewString(), productID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO customers VALUES (?, 120, CURRENT_TIMESTAMP)`, customerID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO customer_ledger VALUES (?, ?, 'debit', 120, ?, 'sale')`, uuid.NewString(), customerID, saleID)
	if err != nil {
		t.Fatal(err)
	}
	return db, saleID, itemID, customerID
}

func TestSmartDeleteCompletedSaleReversesAndPhysicallyDeletesAtomically(t *testing.T) {
	db, saleID, itemID, customerID := newSaleDeleteTestDB(t)
	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("delete result = %+v, want deleted", result)
	}
	var saleCount, movementCount, auditCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sales WHERE id=?`, saleID).Scan(&saleCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=?`, saleID).Scan(&movementCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE entity_id=? AND action='DELETE'`, saleID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	var itemStatus string
	var quantity int
	var balance float64
	if err := db.QueryRow(`SELECT status FROM inventory_items WHERE id=?`, itemID).Scan(&itemStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT quantity FROM inventory`).Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT current_balance FROM customers WHERE id=?`, customerID).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if saleCount != 0 || movementCount != 0 || auditCount != 1 || itemStatus != "AVAILABLE" || quantity != 3 || balance != 0 {
		t.Fatalf("sale=%d movements=%d audit=%d item=%s stock=%d balance=%v", saleCount, movementCount, auditCount, itemStatus, quantity, balance)
	}
}

func TestSmartDeleteSaleWithReturnIsBlockedWithoutPartialChanges(t *testing.T) {
	db, saleID, itemID, _ := newSaleDeleteTestDB(t)
	if _, err := db.Exec(`INSERT INTO returns (id,sale_id) VALUES (?,?)`, uuid.NewString(), saleID); err != nil {
		t.Fatal(err)
	}
	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "blocked" || result.CanProceed {
		t.Fatalf("delete result = %+v, want blocked", result)
	}
	var saleCount int
	var itemStatus string
	var movements int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sales WHERE id=?`, saleID).Scan(&saleCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT status FROM inventory_items WHERE id=?`, itemID).Scan(&itemStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=?`, saleID).Scan(&movements); err != nil {
		t.Fatal(err)
	}
	if saleCount != 1 || itemStatus != "SOLD" || movements != 1 {
		t.Fatalf("blocked deletion changed data: sale=%d item=%s movements=%d", saleCount, itemStatus, movements)
	}
}

func TestSmartDeleteSaleWithMismatchedStockMovementsIsBlocked(t *testing.T) {
	db, saleID, itemID, _ := newSaleDeleteTestDB(t)
	if _, err := db.Exec(`UPDATE inventory_movements SET quantity=-2 WHERE reference_id=?`, saleID); err != nil {
		t.Fatal(err)
	}

	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "blocked" || result.CanProceed {
		t.Fatalf("delete result = %+v, want blocked", result)
	}
	var saleCount, stock int
	var itemStatus string
	if err := db.QueryRow(`SELECT COUNT(*) FROM sales WHERE id=?`, saleID).Scan(&saleCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT quantity FROM inventory`).Scan(&stock); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT status FROM inventory_items WHERE id=?`, itemID).Scan(&itemStatus); err != nil {
		t.Fatal(err)
	}
	if saleCount != 1 || stock != 2 || itemStatus != "SOLD" {
		t.Fatalf("blocked deletion changed data: sale=%d stock=%d item=%s", saleCount, stock, itemStatus)
	}
}

func TestSmartDeleteSaleWithRecordedPaymentIsBlockedWithoutPartialChanges(t *testing.T) {
	db, saleID, itemID, _ := newSaleDeleteTestDB(t)
	if _, err := db.Exec(`UPDATE sales SET paid_amount=20 WHERE id=?`, saleID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payments (id,sale_id) VALUES (?,?)`, uuid.NewString(), saleID); err != nil {
		t.Fatal(err)
	}

	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "blocked" || result.CanProceed {
		t.Fatalf("delete result = %+v, want blocked", result)
	}
	var saleCount, paymentCount, stock int
	var itemStatus string
	if err := db.QueryRow(`SELECT COUNT(*) FROM sales WHERE id=?`, saleID).Scan(&saleCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM payments WHERE sale_id=?`, saleID).Scan(&paymentCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT quantity FROM inventory`).Scan(&stock); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT status FROM inventory_items WHERE id=?`, itemID).Scan(&itemStatus); err != nil {
		t.Fatal(err)
	}
	if saleCount != 1 || paymentCount != 1 || stock != 2 || itemStatus != "SOLD" {
		t.Fatalf("blocked deletion changed data: sale=%d payments=%d stock=%d item=%s", saleCount, paymentCount, stock, itemStatus)
	}
}
