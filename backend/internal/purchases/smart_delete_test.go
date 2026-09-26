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
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, customer_return_id TEXT, sale_id TEXT, purchase_id TEXT, supplier_id TEXT, return_number TEXT, status TEXT, source_status TEXT, reason TEXT, notes TEXT, created_by TEXT, return_reason TEXT, return_date TEXT, refund_amount REAL, created_at TEXT, updated_at TEXT);
		CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT, customer_return_id TEXT, sale_id TEXT, sale_item_id TEXT, inventory_item_id TEXT, purchase_item_id TEXT, product_id TEXT, barcode TEXT, serial_number TEXT, return_reason TEXT, return_date TEXT, created_at TEXT, quantity INTEGER, unit_cost REAL, purchase_cost REAL);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, item_code TEXT, product_id TEXT, barcode TEXT, status TEXT, condition TEXT, sold_at TEXT, updated_at TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT UNIQUE, quantity INTEGER, created_at TEXT, updated_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, reference_type TEXT, reference_id TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, reference_type TEXT, reference_id TEXT);
		CREATE TABLE returns (id TEXT PRIMARY KEY, sale_id TEXT, purchase_id TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, invoice_number TEXT, customer_id TEXT, status TEXT, total_amount REAL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, payment_status TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_price REAL, unit_cost REAL, total_amount REAL, tax_amount REAL, created_at TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, sale_id TEXT, paid_amount REAL DEFAULT 0);
		CREATE TABLE barcodes (id TEXT PRIMARY KEY, inventory_item_id TEXT);
		CREATE TABLE reservations (id TEXT PRIMARY KEY, item_id TEXT);
		CREATE TABLE acquisition_items (id TEXT PRIMARY KEY, inventory_item_id TEXT, item_status TEXT, updated_at TEXT);
		CREATE TABLE inspection_items (id TEXT PRIMARY KEY, item_id TEXT);
		CREATE TABLE item_repair_costs (id TEXT PRIMARY KEY, inventory_item_id TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}
	purchaseID, supplierID, productID, inventoryItemID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	itemCode := "ITM-" + purchaseID.String()[:8] + "-001"
	_, err = db.Exec(`INSERT INTO purchases VALUES (?, 'PUR-DEL-01', ?, 'received', 200, 0)`, purchaseID.String(), supplierID.String())
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
	purchaseItem2ID := uuid.New()
	if _, err := db.Exec(`INSERT INTO purchase_items VALUES (?, ?, ?, 1)`, purchaseItem2ID.String(), purchaseID.String(), productID.String()); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory_items (id,item_code,product_id,barcode,status,condition) VALUES (?, ?, ?, 'BC-PUR-1', 'SOLD', 'NEW')`, inventoryItemID.String(), itemCode, productID.String())
	if err != nil {
		t.Fatal(err)
	}
	returnedItemID := uuid.New()
	returnedItemCode := "ITM-" + purchaseID.String()[:8] + "-002"
	if _, err := db.Exec(`INSERT INTO inventory_items (id,item_code,product_id,barcode,status,condition) VALUES (?, ?, ?, 'BC-PUR-2', 'RETURNED', 'NEW')`, returnedItemID.String(), returnedItemCode, productID.String()); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory VALUES (?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, uuid.NewString(), productID.String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,reference_type,reference_id) VALUES (?, ?, ?, 'PURCHASE', 1, 'purchase', ?)`, uuid.NewString(), inventoryItemID.String(), productID.String(), purchaseID.String())
	if err != nil {
		t.Fatal(err)
	}
	saleID := uuid.New()
	if _, err := db.Exec(`INSERT INTO sales (id,invoice_number,status,total_amount,paid_amount,remaining_amount,payment_status,updated_at) VALUES (?,'SALE-USES-PURCHASE-ITEM','completed',100,0,100,'unpaid',CURRENT_TIMESTAMP)`, saleID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id,sale_id,product_id,inventory_item_id,quantity,unit_price,unit_cost,total_amount,tax_amount,created_at) VALUES (?,?,?,?,1,100,50,100,0,CURRENT_TIMESTAMP)`, uuid.NewString(), saleID.String(), productID.String(), inventoryItemID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,reference_type,reference_id) VALUES (?, ?, ?, 'SALE', -1, 'sale', ?)`, uuid.NewString(), inventoryItemID.String(), productID.String(), saleID.String()); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,reference_type,reference_id) VALUES (?, ?, ?, 'PURCHASE', 1, 'purchase', ?)`, uuid.NewString(), returnedItemID.String(), productID.String(), purchaseID.String())
	if err != nil {
		t.Fatal(err)
	}
	supplierReturnID := uuid.New()
	if _, err = db.Exec(`INSERT INTO supplier_returns (id,purchase_id,supplier_id,return_number,status,reason,refund_amount,created_at,updated_at) VALUES (?, ?, ?, 'SRET-1', 'COMPLETED', 'defective', 100, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, supplierReturnID.String(), purchaseID.String(), supplierID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO supplier_return_items (id,supplier_return_id,inventory_item_id,purchase_item_id,product_id,quantity,unit_cost,created_at) VALUES (?, ?, ?, ?, ?, 1, 100, CURRENT_TIMESTAMP)`, uuid.NewString(), supplierReturnID.String(), returnedItemID.String(), purchaseItem2ID.String(), productID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,reference_type,reference_id) VALUES (?, ?, ?, 'SUPPLIER_RETURN', -1, 'supplier_return', ?)`, uuid.NewString(), returnedItemID.String(), productID.String(), supplierReturnID.String()); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO supplier_ledger VALUES (?, ?, 'debit', 'PURCHASE', 200, ?)`, uuid.NewString(), supplierID.String(), purchaseID.String())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_ledger VALUES (?, ?, 'credit', 'SUPPLIER_RETURN', 100, ?)`, uuid.NewString(), supplierID.String(), supplierReturnID.String()); err != nil {
		t.Fatal(err)
	}

	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), purchaseID, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("purchase deletion result = %+v", result)
	}
	var purchases, inventoryItems, movements, ledgers, audits, salesCount int
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
	if err := db.QueryRow(`SELECT COUNT(*) FROM sales WHERE id=?`, saleID.String()).Scan(&salesCount); err != nil {
		t.Fatal(err)
	}
	var remainingMovements int
	if err := db.QueryRow(`SELECT COUNT(*) FROM inventory_movements`).Scan(&remainingMovements); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT quantity FROM inventory WHERE product_id=?`, productID.String()).Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT current_balance FROM suppliers WHERE id=?`, supplierID.String()).Scan(&supplierBalance); err != nil {
		t.Fatal(err)
	}
	var supplierReturns, supplierReturnItems int
	if err := db.QueryRow(`SELECT COUNT(*) FROM supplier_returns WHERE id=?`, supplierReturnID.String()).Scan(&supplierReturns); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM supplier_return_items WHERE supplier_return_id=?`, supplierReturnID.String()).Scan(&supplierReturnItems); err != nil {
		t.Fatal(err)
	}
	if purchases != 0 || inventoryItems != 0 || movements != 0 || remainingMovements != 0 || ledgers != 0 || audits < 1 || salesCount != 0 || quantity != 0 || supplierBalance != 0 || supplierReturns != 0 || supplierReturnItems != 0 {
		t.Fatalf("purchase=%d sale=%d item=%d movement=%d/%d ledger=%d audit=%d stock=%d supplier_balance=%v supplier_return=%d supplier_return_items=%d", purchases, salesCount, inventoryItems, movements, remainingMovements, ledgers, audits, quantity, supplierBalance, supplierReturns, supplierReturnItems)
	}
}

func TestSmartDeleteDraftPurchaseReversesRecordedAndLegacySummaryPayments(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE purchases (id TEXT PRIMARY KEY, invoice_number TEXT, supplier_id TEXT, status TEXT, total_amount REAL, paid_amount REAL, remaining_amount REAL, updated_at TEXT);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, purchase_id TEXT, customer_id TEXT, supplier_id TEXT, sale_id TEXT, amount REAL, payment_status TEXT, created_at TEXT);
		CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL DEFAULT 0, reference_id TEXT, created_at TEXT DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE suppliers (id TEXT PRIMARY KEY, current_balance REAL, updated_at TEXT);
		CREATE TABLE payment_allocation_batches (payment_id TEXT PRIMARY KEY, owner_type TEXT NOT NULL, owner_id TEXT NOT NULL, sale_id TEXT, tracked_at TEXT);
		CREATE TABLE payment_debt_allocations (id TEXT PRIMARY KEY, payment_id TEXT, debt_id TEXT, amount REAL, created_at TEXT);
		CREATE TABLE supplier_debts (id TEXT PRIMARY KEY, supplier_id TEXT, amount REAL, paid_amount REAL DEFAULT 0, is_paid BOOLEAN DEFAULT FALSE);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, reference_id TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
		CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, movement_type TEXT, reference_type TEXT, reference_id TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, reference_type TEXT, reference_id TEXT);
		CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	purchaseID, paidOnlyID, supplierID, paymentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO purchases VALUES (?, 'PUR-PAID-DRAFT', ?, 'pending', 100, 20, 80, CURRENT_TIMESTAMP), (?, 'PUR-PAID-ONLY', ?, 'pending', 100, 25, 75, CURRENT_TIMESTAMP)`, purchaseID, supplierID, paidOnlyID, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO suppliers VALUES (?,80,CURRENT_TIMESTAMP)`, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payments (id,purchase_id,supplier_id,amount,payment_status,created_at) VALUES (?,?,?,20,'completed',CURRENT_TIMESTAMP)`, paymentID, purchaseID, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payment_allocation_batches (payment_id,owner_type,owner_id) VALUES (?,'supplier',?)`, paymentID, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_ledger (id,supplier_id,type,transaction_type,amount,balance,reference_id) VALUES (?,?,'credit','PAYMENT',20,80,?)`, uuid.New(), supplierID, paymentID); err != nil {
		t.Fatal(err)
	}

	service := NewSmartDeleteService(sqlx.NewDb(db, "sqlite"))
	result, err := service.SmartDelete(context.Background(), purchaseID, uuid.Nil)
	if err != nil || result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("linked payment purchase deletion result=%+v err=%v, want deleted", result, err)
	}
	legacyResult, err := service.SmartDelete(context.Background(), paidOnlyID, uuid.Nil)
	if err != nil || legacyResult.Action != "deleted" || !legacyResult.CanProceed {
		t.Fatalf("legacy paid-summary purchase result=%+v err=%v; want hard delete", legacyResult, err)
	}
	var purchases, payments, ledgers, paidPurchaseCount int
	for _, check := range []struct {
		query string
		id    uuid.UUID
		dest  *int
	}{
		{`SELECT COUNT(*) FROM payments WHERE id=?`, paymentID, &payments},
		{`SELECT COUNT(*) FROM supplier_ledger WHERE reference_id=?`, paymentID, &ledgers},
	} {
		if err := db.QueryRow(check.query, check.id).Scan(check.dest); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM purchases`).Scan(&purchases); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM purchases WHERE id=? AND paid_amount=25`, paidOnlyID).Scan(&paidPurchaseCount); err != nil {
		t.Fatal(err)
	}
	if purchases != 0 || payments != 0 || ledgers != 0 || paidPurchaseCount != 0 {
		t.Fatalf("purchase reversal results: purchases=%d payments=%d supplier_ledger=%d unresolved_paid_purchase=%d", purchases, payments, ledgers, paidPurchaseCount)
	}
}

func TestSmartDeleteHardDeletesCancelledAndReversedPurchases(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE suppliers (id TEXT PRIMARY KEY, current_balance REAL DEFAULT 0, updated_at TEXT);
		CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, reference_id TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, supplier_id TEXT, purchase_id TEXT, amount REAL, payment_method TEXT, created_at TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT, item_code TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, movement_type TEXT, reference_type TEXT, reference_id TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, reference_type TEXT, reference_id TEXT);
		CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE purchases (id TEXT PRIMARY KEY, supplier_id TEXT, invoice_number TEXT, status TEXT, total_amount REAL, paid_amount REAL);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT);
		CREATE TABLE purchase_reversals (id TEXT PRIMARY KEY, purchase_id TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	reversedID, cancelledID := uuid.New(), uuid.New()
	supplierID := uuid.New()
	if _, err := db.Exec(`INSERT INTO suppliers (id,current_balance,updated_at) VALUES (?,0,CURRENT_TIMESTAMP)`, supplierID); err != nil {
		t.Fatal(err)
	}
	for _, purchase := range []struct {
		id     uuid.UUID
		status string
	}{{reversedID, "reversed"}, {cancelledID, "cancelled"}} {
		if _, err := db.Exec(`INSERT INTO purchases (id, supplier_id, invoice_number, status, total_amount, paid_amount) VALUES (?, ?, ?, ?, 100, 0)`, purchase.id, supplierID, "INV-"+purchase.status, purchase.status); err != nil {
			t.Fatal(err)
		}
	}

	service := NewSmartDeleteService(sqlx.NewDb(db, "sqlite"))
	for _, purchaseID := range []uuid.UUID{reversedID, cancelledID} {
		result, err := service.SmartDelete(context.Background(), purchaseID, uuid.Nil)
		if err != nil {
			t.Fatal(err)
		}
		if result.Action != "deleted" || !result.CanProceed {
			t.Fatalf("purchase %s deletion result = %+v, want deleted", purchaseID, result)
		}
	}
	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM purchases`).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("cancelled/reversed purchases remaining = %d, want 0", remaining)
	}
}

func TestSmartDeletePurchaseReversesReservationAdjustmentAndTransferEffectsSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	if _, err := db.Exec(`
		CREATE TABLE purchases (id TEXT PRIMARY KEY, supplier_id TEXT, invoice_number TEXT, status TEXT, total_amount REAL, paid_amount REAL DEFAULT 0, updated_at TEXT);
		CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT, product_id TEXT);
		CREATE TABLE suppliers (id TEXT PRIMARY KEY, current_balance REAL DEFAULT 0, updated_at TEXT);
		CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, reference_id TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, reserved_quantity INTEGER DEFAULT 0, updated_at TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, item_code TEXT, product_id TEXT, status TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, before_status TEXT, after_status TEXT, reference_type TEXT, reference_id TEXT);
		CREATE TABLE reservations (id TEXT PRIMARY KEY, item_id TEXT, status TEXT);
		CREATE TABLE returns (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, purchase_id TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	purchaseID, supplierID, productID := uuid.New(), uuid.New(), uuid.New()
	itemAdjustmentID, reservedItemID, convertedItemID := uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO purchases (id,supplier_id,invoice_number,status,total_amount,paid_amount,updated_at) VALUES (?,?,'PUR-REVERSE-ITEM-STATE','received',50,0,CURRENT_TIMESTAMP)`, purchaseID, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO suppliers (id,current_balance,updated_at) VALUES (?,50,CURRENT_TIMESTAMP)`, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_ledger (id,supplier_id,type,transaction_type,amount,reference_id) VALUES (?,?,'debit','PURCHASE',50,?)`, uuid.New(), supplierID, purchaseID); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		id     uuid.UUID
		suffix string
		status string
	}{{itemAdjustmentID, "001", "DAMAGED"}, {reservedItemID, "002", "RESERVED"}, {convertedItemID, "003", "SOLD"}} {
		itemCode := "ITM-" + purchaseID.String()[:8] + "-" + item.suffix
		if _, err := db.Exec(`INSERT INTO inventory_items (id,item_code,product_id,status) VALUES (?,?,?,?)`, item.id, itemCode, productID, item.status); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO purchase_items (id,purchase_id,product_id) VALUES (?,?,?)`, uuid.New(), purchaseID, productID); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,reference_type,reference_id) VALUES (?,?,?,'PURCHASE',1,'purchase',?)`, uuid.New(), item.id, productID, purchaseID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,reserved_quantity,updated_at) VALUES (?,?,6,1,CURRENT_TIMESTAMP)`, uuid.New(), productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,before_quantity,after_quantity,before_status,after_status,reference_type,reference_id) VALUES (?,?,?,'ADJUSTMENT',-1,8,7,'AVAILABLE','DAMAGED','adjustment',?)`, uuid.New(), itemAdjustmentID, productID, itemAdjustmentID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO reservations (id,item_id,status) VALUES (?,?,'active'), (?,?,'converted')`, uuid.New(), reservedItemID, uuid.New(), convertedItemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id) VALUES (?,?,?,'TRANSFER',1,1,1,'transfer',?),(?,?,?,'SALE',1,7,6,'reservation',(SELECT id FROM reservations WHERE item_id=? AND status='converted'))`, uuid.New(), reservedItemID, productID, reservedItemID, uuid.New(), convertedItemID, productID, convertedItemID); err != nil {
		t.Fatal(err)
	}

	result, err := NewSmartDeleteService(sqlx.NewDb(db, "sqlite")).SmartDelete(context.Background(), purchaseID, uuid.Nil)
	if err != nil || result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("purchase with reservation/adjustment/transfer history result=%+v err=%v", result, err)
	}
	var purchases, items, reservations, movements int
	var stock, reserved int
	var supplierBalance float64
	for _, check := range []struct {
		query string
		dest  *int
	}{{`SELECT COUNT(*) FROM purchases`, &purchases}, {`SELECT COUNT(*) FROM inventory_items`, &items}, {`SELECT COUNT(*) FROM reservations`, &reservations}, {`SELECT COUNT(*) FROM inventory_movements`, &movements}} {
		if err := db.QueryRow(check.query).Scan(check.dest); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.QueryRow(`SELECT quantity,reserved_quantity FROM inventory WHERE product_id=?`, productID).Scan(&stock, &reserved); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT current_balance FROM suppliers WHERE id=?`, supplierID).Scan(&supplierBalance); err != nil {
		t.Fatal(err)
	}
	if purchases != 0 || items != 0 || reservations != 0 || movements != 0 || stock != 5 || reserved != 0 || supplierBalance != 0 {
		t.Fatalf("remaining purchase dependencies: purchases=%d items=%d reservations=%d movements=%d stock=%d reserved=%d supplierBalance=%v; want zero rows, stock=5, reserved=0, balance=0", purchases, items, reservations, movements, stock, reserved, supplierBalance)
	}
}
