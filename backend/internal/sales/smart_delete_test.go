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
		CREATE TABLE sales (id TEXT PRIMARY KEY, invoice_number TEXT, customer_id TEXT, status TEXT, total_amount REAL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, payment_status TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, created_at TEXT DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT, sold_at TEXT, updated_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, reference_type TEXT, reference_id TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT UNIQUE, quantity INTEGER, created_at TEXT, updated_at TEXT);
		CREATE TABLE returns (id TEXT PRIMARY KEY, return_number TEXT, sale_id TEXT, status TEXT, refund_method TEXT, customer_id TEXT, debt_id TEXT, debt_adjustment REAL DEFAULT 0, total_refund_amount REAL DEFAULT 0, return_date TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE return_items (id TEXT PRIMARY KEY, return_id TEXT, sale_item_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity_returned INTEGER DEFAULT 0);
		CREATE TABLE payment_transactions (id TEXT PRIMARY KEY, order_id TEXT, sale_id TEXT, payment_id TEXT, provider TEXT, provider_payment_id TEXT, provider_transaction_id TEXT, status TEXT, amount_minor INTEGER, currency TEXT, idempotency_key TEXT, checkout_url TEXT, failure_code TEXT, failure_message TEXT, metadata TEXT, created_at DATETIME, updated_at DATETIME, paid_at DATETIME, cancelled_at DATETIME);
		CREATE TABLE debts (id TEXT PRIMARY KEY, sale_id TEXT, customer_id TEXT, amount REAL DEFAULT 0, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, due_date TEXT, status TEXT DEFAULT 'pending', created_at TEXT DEFAULT CURRENT_TIMESTAMP, updated_at TEXT DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE payments (id TEXT PRIMARY KEY, sale_id TEXT, customer_id TEXT, supplier_id TEXT, amount REAL, payment_status TEXT, payment_date TEXT, created_at TEXT DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL DEFAULT 0, reference_id TEXT, reference_type TEXT, created_at TEXT DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL, updated_at TEXT);
		CREATE TABLE sale_payment_allocations (id TEXT PRIMARY KEY, sale_id TEXT);
		CREATE TABLE payment_allocation_batches (payment_id TEXT PRIMARY KEY, owner_type TEXT NOT NULL, owner_id TEXT NOT NULL, sale_id TEXT, tracked_at TEXT);
		CREATE TABLE payment_debt_allocations (id TEXT PRIMARY KEY, payment_id TEXT NOT NULL, debt_id TEXT NOT NULL, amount REAL NOT NULL, created_at TEXT, UNIQUE(payment_id,debt_id));
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
	_, err = db.Exec(`INSERT INTO sale_items (id,sale_id,product_id,quantity) VALUES (?, ?, ?, 1)`, uuid.NewString(), saleID, productID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory_items VALUES (?, ?, 'SOLD', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, itemID, productID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE sale_items SET inventory_item_id=? WHERE sale_id=?`, itemID, saleID); err != nil {
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
	_, err = db.Exec(`INSERT INTO customer_ledger (id,customer_id,type,amount,reference_id,reference_type) VALUES (?, ?, 'debit', 120, ?, 'sale')`, uuid.NewString(), customerID, saleID)
	if err != nil {
		t.Fatal(err)
	}
	return db, saleID, itemID, customerID
}

func TestSmartDeleteCompletedSaleReversesAndPhysicallyDeletesAtomically(t *testing.T) {
	db, saleID, itemID, customerID := newSaleDeleteTestDB(t)
	for _, statement := range []string{
		`CREATE TABLE financial_transactions (id TEXT PRIMARY KEY, sale_id TEXT, type TEXT, amount REAL)`,
		`CREATE TABLE profit_entries (id TEXT PRIMARY KEY, sale_id TEXT, revenue REAL, cost REAL, gross_profit REAL, net_profit REAL)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO financial_transactions VALUES (?,?,'sale',120)`, uuid.NewString(), saleID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO profit_entries VALUES (?,?,120,40,80,80)`, uuid.NewString(), saleID); err != nil {
		t.Fatal(err)
	}
	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("delete result = %+v, want deleted", result)
	}
	var saleCount, movementCount, auditCount, financialCount, profitCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sales WHERE id=?`, saleID).Scan(&saleCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=?`, saleID).Scan(&movementCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE entity_id=? AND action='DELETE'`, saleID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM financial_transactions WHERE sale_id=?`, saleID).Scan(&financialCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM profit_entries WHERE sale_id=?`, saleID).Scan(&profitCount); err != nil {
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
	if saleCount != 0 || movementCount != 0 || auditCount != 1 || financialCount != 0 || profitCount != 0 || itemStatus != "AVAILABLE" || quantity != 3 || balance != 0 {
		t.Fatalf("sale=%d movements=%d audit=%d financial=%d profit=%d item=%s stock=%d balance=%v", saleCount, movementCount, auditCount, financialCount, profitCount, itemStatus, quantity, balance)
	}
}

func TestSmartDeleteLegacyCompletedSaleReconstructsMissingInventoryMovementSQLite(t *testing.T) {
	db, saleID, itemID, _ := newSaleDeleteTestDB(t)
	if _, err := db.Exec(`DELETE FROM inventory_movements WHERE reference_id=? AND movement_type='SALE'`, saleID); err != nil {
		t.Fatal(err)
	}
	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatalf("delete legacy sale with missing movement: %v", err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("delete result = %+v, want deleted", result)
	}
	var itemStatus string
	var quantity int
	var saleCount int
	if err := db.QueryRow(`SELECT status FROM inventory_items WHERE id=?`, itemID).Scan(&itemStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT quantity FROM inventory`).Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM sales WHERE id=?`, saleID).Scan(&saleCount); err != nil {
		t.Fatal(err)
	}
	if itemStatus != "AVAILABLE" || quantity != 3 || saleCount != 0 {
		t.Fatalf("legacy reversal left item=%s inventory=%d sale_count=%d; want AVAILABLE, 3, 0", itemStatus, quantity, saleCount)
	}
}

func TestSmartDeleteRecalculatesAffectedShiftTotals(t *testing.T) {
	db, saleID, _, _ := newSaleDeleteTestDB(t)
	userID := uuid.NewString()
	if _, err := db.Exec(`ALTER TABLE sales ADD COLUMN user_id TEXT`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`ALTER TABLE sales ADD COLUMN created_at TEXT`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE sales SET user_id=?, created_at=? WHERE id=?`, userID, "2026-09-25T10:00:00Z", saleID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id, invoice_number, customer_id, status, total_amount, paid_amount, user_id, created_at) VALUES (?, 'INV-KEEP', NULL, 'completed', 45, 0, ?, ?)`, uuid.NewString(), userID, "2026-09-25T11:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE pos_shifts (id TEXT PRIMARY KEY, user_id TEXT, opened_at TEXT, closed_at TEXT, sales_total REAL, sale_count INTEGER)`); err != nil {
		t.Fatal(err)
	}
	shiftID, unrelatedShiftID := uuid.NewString(), uuid.NewString()
	if _, err := db.Exec(`INSERT INTO pos_shifts VALUES (?, ?, ?, ?, 165, 2), (?, ?, ?, ?, 999, 9)`,
		shiftID, userID, "2026-09-25T08:00:00Z", "2026-09-25T12:00:00Z",
		unrelatedShiftID, userID, "2026-09-25T12:00:00Z", "2026-09-25T16:00:00Z"); err != nil {
		t.Fatal(err)
	}

	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" {
		t.Fatalf("delete result = %+v, want deleted", result)
	}
	var affectedTotal float64
	var affectedCount int
	if err := db.QueryRow(`SELECT sales_total, sale_count FROM pos_shifts WHERE id=?`, shiftID).Scan(&affectedTotal, &affectedCount); err != nil {
		t.Fatal(err)
	}
	var unrelatedTotal float64
	var unrelatedCount int
	if err := db.QueryRow(`SELECT sales_total, sale_count FROM pos_shifts WHERE id=?`, unrelatedShiftID).Scan(&unrelatedTotal, &unrelatedCount); err != nil {
		t.Fatal(err)
	}
	if affectedTotal != 45 || affectedCount != 1 {
		t.Fatalf("affected shift = total %v, count %d; want 45/1", affectedTotal, affectedCount)
	}
	if unrelatedTotal != 999 || unrelatedCount != 9 {
		t.Fatalf("unrelated shift changed to total %v, count %d", unrelatedTotal, unrelatedCount)
	}
}

func TestCleanSalesHistoryDeletesEligibleAndLegacyPaidSales(t *testing.T) {
	db, saleID, _, _ := newSaleDeleteTestDB(t)
	for _, column := range []string{
		`sale_date TEXT DEFAULT '2026-09-25'`,
		`subtotal REAL DEFAULT 0`,
		`tax_amount REAL DEFAULT 0`,
		`discount_amount REAL DEFAULT 0`,
		`cost_amount REAL DEFAULT 0`,
		`gross_profit REAL DEFAULT 0`,
		`net_profit REAL DEFAULT 0`,
		`payment_method TEXT`,
		`notes TEXT`,
		`created_at TEXT DEFAULT '2026-09-25T10:00:00Z'`,
	} {
		if _, err := db.Exec(`ALTER TABLE sales ADD COLUMN ` + column); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`ALTER TABLE customers ADD COLUMN name TEXT`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE sales SET updated_at='2026-09-25T10:00:00Z' WHERE id=?`, saleID); err != nil {
		t.Fatal(err)
	}
	blockedSaleID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO sales (id, invoice_number, customer_id, status, total_amount, paid_amount, sale_date, created_at, updated_at, payment_status) VALUES (?, 'INV-BLOCKED', NULL, 'completed', 25, 25, '2026-09-25', '2026-09-25T11:00:00Z', '2026-09-25T11:00:00Z', 'paid')`, blockedSaleID); err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlite")
	service := NewService(NewRepository(sqlxDB), sqlxDB)
	summary, err := service.CleanSalesHistory(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sales`).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if summary.Total != 2 || summary.Deleted != 2 || summary.Blocked != 0 || summary.Failed != 0 || remaining != 0 {
		t.Fatalf("cleanup summary=%+v remaining sales=%d; want total 2, deleted 2, blocked 0, failed 0, remaining 0", summary, remaining)
	}
}

func TestSmartDeleteSaleReversesLinkedReturnAtomically(t *testing.T) {
	db, saleID, itemID, _ := newSaleDeleteTestDB(t)
	returnID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO returns (id,return_number,sale_id,status,return_date,created_at,updated_at) VALUES (?, 'RET-1', ?, 'COMPLETED', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, returnID, saleID); err != nil {
		t.Fatal(err)
	}
	var productID, saleItemID string
	if err := db.QueryRow(`SELECT product_id FROM inventory_items WHERE id=?`, itemID).Scan(&productID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT id FROM sale_items WHERE sale_id=?`, saleID).Scan(&saleItemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE inventory SET quantity=3 WHERE product_id=?`, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE inventory_items SET status='AVAILABLE', sold_at=NULL WHERE id=?`, itemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO return_items (id,return_id,sale_item_id,product_id,inventory_item_id,quantity_returned) VALUES (?, ?, ?, ?, ?, 1)`, uuid.NewString(), returnID, saleItemID, productID, itemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,reference_type,reference_id) VALUES (?, ?, ?, 'RETURN', 1, 'return', ?)`, uuid.NewString(), itemID, productID, returnID); err != nil {
		t.Fatal(err)
	}
	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("delete result = %+v, want deleted", result)
	}
	var saleCount, returnCount, returnItemCount, stock, returnMovementCount int
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
	if err := db.QueryRow(`SELECT COUNT(*) FROM returns WHERE id=?`, returnID).Scan(&returnCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM return_items WHERE return_id=?`, returnID).Scan(&returnItemCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT quantity FROM inventory WHERE product_id=?`, productID).Scan(&stock); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=?`, returnID).Scan(&returnMovementCount); err != nil {
		t.Fatal(err)
	}
	if saleCount != 0 || returnCount != 0 || returnItemCount != 0 || itemStatus != "AVAILABLE" || movements != 0 || returnMovementCount != 0 || stock != 3 {
		t.Fatalf("sale cascade result: sale=%d return=%d return items=%d item=%s sale movements=%d return movements=%d stock=%d; want rows removed and original stock 3 restored", saleCount, returnCount, returnItemCount, itemStatus, movements, returnMovementCount, stock)
	}
}

func TestSmartDeleteSaleDeletesSalesThatConsumedReturnedInventoryFirstSQLite(t *testing.T) {
	db, originalSaleID, itemID, _ := newSaleDeleteTestDB(t)
	var productID, originalSaleItemID string
	if err := db.QueryRow(`SELECT product_id FROM inventory_items WHERE id=?`, itemID).Scan(&productID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT id FROM sale_items WHERE sale_id=?`, originalSaleID).Scan(&originalSaleItemID); err != nil {
		t.Fatal(err)
	}
	returnID, laterSaleID, laterSaleItemID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := "2026-09-27T10:00:00Z"
	if _, err := db.Exec(`INSERT INTO returns (id,return_number,sale_id,status,refund_method,customer_id,total_refund_amount,return_date,created_at,updated_at) VALUES (?, 'RET-CASCADE', ?, 'COMPLETED', 'CASH', '', 100, ?, ?, ?)`, returnID, originalSaleID, now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO return_items (id,return_id,sale_item_id,product_id,inventory_item_id,quantity_returned) VALUES (?, ?, ?, ?, ?, 1)`, uuid.NewString(), returnID, originalSaleItemID, productID, itemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE inventory_items SET status='AVAILABLE', sold_at=NULL WHERE id=?`, itemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,reference_type,reference_id) VALUES (?, ?, ?, 'RETURN', 1, 'return', ?)`, uuid.NewString(), itemID, productID, returnID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id,invoice_number,status,total_amount,paid_amount,remaining_amount,payment_status,updated_at) VALUES (?, 'INV-LATER', 'completed', 100, 100, 0, 'paid', ?)`, laterSaleID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id,sale_id,product_id,inventory_item_id,quantity,created_at) VALUES (?, ?, ?, ?, 1, ?)`, laterSaleItemID, laterSaleID, productID, itemID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,reference_type,reference_id) VALUES (?, ?, ?, 'SALE', -1, 'sale', ?)`, uuid.NewString(), itemID, productID, laterSaleID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE inventory SET quantity=2 WHERE product_id=?`, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE inventory_items SET status='SOLD', sold_at=CURRENT_TIMESTAMP WHERE id=?`, itemID); err != nil {
		t.Fatal(err)
	}

	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(originalSaleID), uuid.Nil)
	if err != nil {
		t.Fatalf("delete sale with later sale of returned item: %v", err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("delete result=%+v, want deleted", result)
	}
	var salesRemaining, returnsRemaining, stock int
	var status string
	if err := db.QueryRow(`SELECT COUNT(*) FROM sales WHERE id IN (?,?)`, originalSaleID, laterSaleID).Scan(&salesRemaining); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM returns WHERE id=?`, returnID).Scan(&returnsRemaining); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT quantity FROM inventory WHERE product_id=?`, productID).Scan(&stock); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT status FROM inventory_items WHERE id=?`, itemID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if salesRemaining != 0 || returnsRemaining != 0 || stock != 3 || status != "AVAILABLE" {
		t.Fatalf("cascade left sales=%d returns=%d stock=%d item=%s; want 0, 0, 3, AVAILABLE", salesRemaining, returnsRemaining, stock, status)
	}
}

func TestSmartDeleteSaleReversesRecordedStockDeltaWhenLegacyInvoiceQuantityDiffersSQLite(t *testing.T) {
	db, saleID, itemID, _ := newSaleDeleteTestDB(t)
	if _, err := db.Exec(`UPDATE inventory_movements SET quantity=-2 WHERE reference_id=?`, saleID); err != nil {
		t.Fatal(err)
	}

	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("delete result = %+v, want deleted", result)
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
	if saleCount != 0 || stock != 4 || itemStatus != "AVAILABLE" {
		t.Fatalf("reversal result: sale=%d stock=%d item=%s; want 0, 4, AVAILABLE", saleCount, stock, itemStatus)
	}
}

func TestSmartDeleteSaleReversesAndHardDeletesProviderHistorySQLite(t *testing.T) {
	db, saleID, _, _ := newSaleDeleteTestDB(t)
	if _, err := db.Exec(`CREATE TABLE payment_refunds (id TEXT PRIMARY KEY,payment_transaction_id TEXT,status TEXT); CREATE TABLE payment_webhook_events (id TEXT PRIMARY KEY,payment_transaction_id TEXT,status TEXT)`); err != nil {
		t.Fatal(err)
	}
	transactionID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO payment_transactions (id,sale_id,provider,status,amount_minor,currency,idempotency_key,metadata,created_at,updated_at) VALUES (?,?,'mock','refunded',2500,'ILS',?,'{}',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, transactionID, saleID, uuid.NewString()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payment_refunds VALUES (?,?,'refunded')`, uuid.NewString(), transactionID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payment_webhook_events VALUES (?,?,'processed')`, uuid.NewString(), transactionID); err != nil {
		t.Fatal(err)
	}
	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" {
		t.Fatalf("provider-linked sale delete result=%+v", result)
	}
	var salesCount, transactionCount, refundCount, webhookCount int
	for _, check := range []struct {
		query string
		arg   string
		dest  *int
	}{
		{`SELECT COUNT(*) FROM sales WHERE id=?`, saleID, &salesCount},
		{`SELECT COUNT(*) FROM payment_transactions WHERE id=?`, transactionID, &transactionCount},
		{`SELECT COUNT(*) FROM payment_refunds WHERE payment_transaction_id=?`, transactionID, &refundCount},
		{`SELECT COUNT(*) FROM payment_webhook_events WHERE payment_transaction_id=?`, transactionID, &webhookCount},
	} {
		if err := db.QueryRow(check.query, check.arg).Scan(check.dest); err != nil {
			t.Fatal(err)
		}
	}
	if salesCount != 0 || transactionCount != 0 || refundCount != 0 || webhookCount != 0 {
		t.Fatalf("sale provider dependencies remain: sales=%d tx=%d refunds=%d webhooks=%d", salesCount, transactionCount, refundCount, webhookCount)
	}
}

func TestSmartDeleteSaleReconstructsLegacyDebtPaymentAndReversesIt(t *testing.T) {
	db, saleID, _, customerID := newSaleDeleteTestDB(t)
	debtID := uuid.NewString()
	paymentID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO debts (id,sale_id,customer_id,amount,paid_amount,remaining_amount,due_date,status,created_at) VALUES (?,?,?,120,20,100,'2026-10-01','partial','2026-09-25T09:00:00Z')`, debtID, saleID, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payments (id,customer_id,amount,payment_status,created_at) VALUES (?,?,20,'completed','2026-09-25T11:00:00Z')`, paymentID, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_ledger (id,customer_id,type,amount,balance,reference_id,reference_type) VALUES (?,?, 'credit',20,100,?,'payment')`, uuid.NewString(), customerID, paymentID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE customers SET current_balance=100 WHERE id=?`, customerID); err != nil {
		t.Fatal(err)
	}
	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("delete result = %+v, want deleted", result)
	}
	var salesCount, debtCount, paymentCount, allocationCount, ledgerCount int
	var customerBalance float64
	checks := []struct {
		query string
		arg   string
		dest  any
	}{
		{`SELECT COUNT(*) FROM sales WHERE id=?`, saleID, &salesCount},
		{`SELECT COUNT(*) FROM debts WHERE id=?`, debtID, &debtCount},
		{`SELECT COUNT(*) FROM payments WHERE id=?`, paymentID, &paymentCount},
		{`SELECT COUNT(*) FROM payment_debt_allocations WHERE debt_id=?`, debtID, &allocationCount},
		{`SELECT COUNT(*) FROM customer_ledger WHERE customer_id=?`, customerID, &ledgerCount},
		{`SELECT current_balance FROM customers WHERE id=?`, customerID, &customerBalance},
	}
	for _, check := range checks {
		if err := db.QueryRow(check.query, check.arg).Scan(check.dest); err != nil {
			t.Fatalf("check %q: %v", check.query, err)
		}
	}
	if salesCount != 0 || debtCount != 0 || paymentCount != 0 || allocationCount != 0 || ledgerCount != 0 || customerBalance != 0 {
		t.Fatalf("legacy sale debt/payment reversal left rows or balances: sale=%d debt=%d payment=%d allocations=%d ledger=%d balance=%v", salesCount, debtCount, paymentCount, allocationCount, ledgerCount, customerBalance)
	}
}

func TestSmartDeleteSaleReversesRecordedPaymentAndDeletesAtomically(t *testing.T) {
	db, saleID, itemID, customerID := newSaleDeleteTestDB(t)
	if _, err := db.Exec(`UPDATE sales SET paid_amount=20, remaining_amount=100, payment_status='partial' WHERE id=?`, saleID); err != nil {
		t.Fatal(err)
	}
	paymentID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO payments (id,sale_id,customer_id,amount,payment_status) VALUES (?,?,?,20,'completed')`, paymentID, saleID, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payment_allocation_batches (payment_id,owner_type,owner_id) VALUES (?,'customer',?)`, paymentID, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_ledger (id,customer_id,type,amount,balance,reference_id,reference_type) VALUES (? ,?,'credit',20,100,?,'payment')`, uuid.NewString(), customerID, paymentID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE customers SET current_balance=100 WHERE id=?`, customerID); err != nil {
		t.Fatal(err)
	}

	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), uuid.MustParse(saleID), uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("delete result = %+v, want deleted", result)
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
	if saleCount != 0 || paymentCount != 0 || stock != 3 || itemStatus != "AVAILABLE" {
		t.Fatalf("reversal/deletion results: sale=%d payments=%d stock=%d item=%s", saleCount, paymentCount, stock, itemStatus)
	}
}
