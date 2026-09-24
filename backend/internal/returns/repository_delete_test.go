package returns

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/reports"
	salesrepo "github.com/partflow/smart-store/internal/sales"
	_ "modernc.org/sqlite"
)

func TestDeleteReturnRemovesUnprocessedReturnAndWritesDeletionAudit(t *testing.T) {
	db := openReturnDeleteDB(t)
	returnID := uuid.New().String()
	if _, err := db.Exec(`INSERT INTO returns (id, return_number, status, refund_status, return_date, created_at) VALUES (?, 'RET-DELETE-1', 'PENDING', 'pending', '2026-09-20', '2026-09-20')`, returnID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO return_items (id, return_id, quantity_returned, unit_price, total_refund_amount, created_at) VALUES ('item-1', ?, 1, 10, 10, '2026-09-20')`, returnID); err != nil {
		t.Fatal(err)
	}

	if err := NewRepository(db).DeleteReturn(context.Background(), uuid.MustParse(returnID)); err != nil {
		t.Fatalf("delete unprocessed return: %v", err)
	}

	var returnCount, itemCount, auditCount, effectCount int
	for query, destination := range map[string]*int{
		`SELECT COUNT(*) FROM returns WHERE id=?`:                               &returnCount,
		`SELECT COUNT(*) FROM return_items WHERE return_id=?`:                   &itemCount,
		`SELECT COUNT(*) FROM audit_logs WHERE entity_id=? AND action='DELETE'`: &auditCount,
		`SELECT COUNT(*) FROM return_effects WHERE id=?`:                        &effectCount,
	} {
		args := []any{returnID}
		if query == `SELECT COUNT(*) FROM audit_logs WHERE entity_id=? AND action='DELETE'` {
			args = []any{returnID}
		}
		if err := db.Get(destination, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	if returnCount != 0 || itemCount != 0 || auditCount != 1 || effectCount != 0 {
		t.Fatalf("delete results: return=%d items=%d delete_audit=%d effects=%d; want 0, 0, 1, 0", returnCount, itemCount, auditCount, effectCount)
	}
}

func TestDeleteCompletedReturnPreservesPostedEffectsAndReports(t *testing.T) {
	db := openReturnDeleteDB(t)
	ctx := context.Background()
	returnID := uuid.New().String()
	customerID := uuid.New().String()
	productID := uuid.New().String()
	saleID := uuid.New().String()
	saleItemID := uuid.New().String()
	inventoryItemID := uuid.New().String()
	debtID := uuid.New().String()
	transactionID := uuid.New().String()
	refundID := uuid.New().String()
	storeDate := time.Now().UTC().Format("2006-01-02")

	insert := func(query string, args ...any) {
		t.Helper()
		if _, execErr := db.Exec(query, args...); execErr != nil {
			t.Fatalf("insert completed return fixture: %v", execErr)
		}
	}
	insert(`INSERT INTO products (id, name, cost_price) VALUES (?, 'Part', 50)`, productID)
	insert(`INSERT INTO sales (id, sale_date, total_amount, tax_amount, cost_amount, paid_amount, payment_method, status, customer_id, created_at) VALUES (?, ?, 1000, 0, 500, 1000, 'cash', 'completed', ?, ?)`, saleID, storeDate, customerID, storeDate)
	insert(`INSERT INTO inventory_items (id, product_id, status, purchase_cost, created_at, updated_at) VALUES (?, ?, 'AVAILABLE', 50, ?, ?)`, inventoryItemID, productID, storeDate, storeDate)
	insert(`INSERT INTO sale_items (id, sale_id, product_id, inventory_item_id, quantity, unit_price, unit_cost, total_amount, tax_amount, created_at) VALUES (?, ?, ?, ?, 10, 100, 50, 1000, 0, ?)`, saleItemID, saleID, productID, inventoryItemID, storeDate)
	insert(`INSERT INTO returns (id, return_number, reference_number, sale_id, customer_id, total_refund_amount, refund_status, status, return_date, refund_date, reason, refund_method, debt_id, debt_adjustment, customer_credit, created_at) VALUES (?, 'RET-POSTED-1', 'REF-1', ?, ?, 200, 'refunded', 'COMPLETED', ?, ?, 'DAMAGED', 'CASH', ?, 200, 0, ?)`, returnID, saleID, customerID, storeDate, storeDate, debtID, storeDate)
	insert(`INSERT INTO return_items (id, return_id, sale_item_id, product_id, inventory_item_id, quantity_returned, original_quantity, unit_price, total_refund_amount, original_cost, created_at) VALUES ('return-item-1', ?, ?, ?, ?, 2, 10, 100, 200, 50, ?)`, returnID, saleItemID, productID, inventoryItemID, storeDate)
	insert(`INSERT INTO inventory (id, product_id, quantity) VALUES ('inventory-1', ?, 5)`, productID)
	insert(`INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, reference_type, reference_id, created_at) VALUES ('movement-1', ?, ?, 'RETURN', 2, 'return', ?, ?)`, inventoryItemID, productID, returnID, storeDate)
	insert(`INSERT INTO customer_ledger (id, customer_id, type, amount, balance, reference_type, reference_id, created_at) VALUES ('customer-ledger-1', ?, 'credit', 200, 800, 'return', ?, ?)`, customerID, returnID, storeDate)
	insert(`INSERT INTO customers (id, current_balance, updated_at) VALUES (?, 800, ?)`, customerID, storeDate)
	insert(`INSERT INTO debts (id, amount, paid_amount, remaining_amount, status, updated_at) VALUES (?, 1000, 200, 800, 'partial', ?)`, debtID, storeDate)
	insert(`INSERT INTO payment_transactions (id, amount_minor, status, updated_at) VALUES (?, 100000, 'partially_refunded', ?)`, transactionID, storeDate)
	insert(`INSERT INTO payment_refunds (id, payment_transaction_id, amount_minor, status, created_at) VALUES (?, ?, 20000, 'refunded', ?)`, refundID, transactionID, storeDate)
	insert(`INSERT INTO return_payment_refunds (id, return_id, payment_transaction_id, payment_refund_id, amount_minor) VALUES ('return-refund-link-1', ?, ?, ?, 20000)`, returnID, transactionID, refundID)
	insert(`INSERT INTO ledger_entries (id, ledger_type, entity_id, transaction_type, amount, reference_type, reference_id, created_at) VALUES ('ledger-entry-1', 'customer', ?, 'RETURN', 200, 'return', ?, ?)`, customerID, returnID, storeDate)
	insert(`INSERT INTO item_history (id, inventory_item_id, event_type, event_date, reference_type, reference_id) VALUES ('item-history-1', 'inventory-item-1', 'RETURNED', ?, 'return', ?)`, storeDate, returnID)
	insert(`INSERT INTO return_refunds (id, return_id, amount, refund_date) VALUES ('return-refund-record-1', ?, 200, ?)`, returnID, storeDate)
	insert(`INSERT INTO return_inspection (id, return_item_id) VALUES ('inspection-1', 'return-item-1')`)
	insert(`INSERT INTO return_audit_log (id, return_id) VALUES ('return-audit-1', ?)`, returnID)

	repo := reports.NewRepository(db)
	start, err := time.Parse("2006-01-02", storeDate)
	if err != nil {
		t.Fatal(err)
	}
	end := start.AddDate(0, 0, 1)
	beforeSales, err := repo.GetSalesData(ctx, start, end)
	if err != nil {
		t.Fatalf("sales report before deletion: %v", err)
	}
	beforeProfit, err := repo.GetProfitsData(ctx, start, end)
	if err != nil {
		t.Fatalf("profit report before deletion: %v", err)
	}
	beforeReturns, err := repo.GetReturnsData(ctx, start, end)
	if err != nil {
		t.Fatalf("returns report before deletion: %v", err)
	}
	returnsRepo := NewRepository(db)
	beforeStats, err := returnsRepo.GetReturnStatistics(ctx)
	if err != nil {
		t.Fatalf("return statistics before deletion: %v", err)
	}
	beforeMonthly, err := returnsRepo.GetMonthlyReturnsAnalysis(ctx)
	if err != nil {
		t.Fatalf("monthly returns analysis before deletion: %v", err)
	}
	beforeSalesReturns, err := returnsRepo.GetSalesReturnsAnalysis(ctx)
	if err != nil {
		t.Fatalf("sales returns analysis before deletion: %v", err)
	}

	if err := NewRepository(db).DeleteReturn(ctx, uuid.MustParse(returnID)); err != nil {
		t.Fatalf("delete completed return: %v", err)
	}

	afterSales, err := repo.GetSalesData(ctx, start, end)
	if err != nil {
		t.Fatalf("sales report after deletion: %v", err)
	}
	afterProfit, err := repo.GetProfitsData(ctx, start, end)
	if err != nil {
		t.Fatalf("profit report after deletion: %v", err)
	}
	afterReturns, err := repo.GetReturnsData(ctx, start, end)
	if err != nil {
		t.Fatalf("returns report after deletion: %v", err)
	}
	afterStats, err := returnsRepo.GetReturnStatistics(ctx)
	if err != nil {
		t.Fatalf("return statistics after deletion: %v", err)
	}
	afterMonthly, err := returnsRepo.GetMonthlyReturnsAnalysis(ctx)
	if err != nil {
		t.Fatalf("monthly returns analysis after deletion: %v", err)
	}
	afterSalesReturns, err := returnsRepo.GetSalesReturnsAnalysis(ctx)
	if err != nil {
		t.Fatalf("sales returns analysis after deletion: %v", err)
	}
	if beforeSales.TotalRevenue != 800 || beforeSales.TotalCOGS != 400 || beforeSales.TotalItemsSold != 8 {
		t.Fatalf("pre-delete sales totals = revenue %v, cogs %v, items %v; want 800, 400, 8", beforeSales.TotalRevenue, beforeSales.TotalCOGS, beforeSales.TotalItemsSold)
	}
	if beforeProfit.TotalRevenue != 800 || beforeProfit.TotalCOGS != 400 || beforeProfit.NetProfit != 400 {
		t.Fatalf("pre-delete profit totals = revenue %v, cogs %v, net %v; want 800, 400, 400", beforeProfit.TotalRevenue, beforeProfit.TotalCOGS, beforeProfit.NetProfit)
	}
	if beforeReturns.TotalRefunded != 200 || beforeReturns.TotalReturns != 1 {
		t.Fatalf("pre-delete return totals = amount %v, count %d; want 200, 1", beforeReturns.TotalRefunded, beforeReturns.TotalReturns)
	}
	if afterSales.TotalRevenue != beforeSales.TotalRevenue || afterSales.TotalCOGS != beforeSales.TotalCOGS || afterSales.TotalItemsSold != beforeSales.TotalItemsSold {
		t.Fatalf("sales report changed after deleting admin record: before=%+v after=%+v", beforeSales, afterSales)
	}
	if afterProfit.TotalRevenue != beforeProfit.TotalRevenue || afterProfit.TotalCOGS != beforeProfit.TotalCOGS || afterProfit.NetProfit != beforeProfit.NetProfit {
		t.Fatalf("profit report changed after deleting admin record: before=%+v after=%+v", beforeProfit, afterProfit)
	}
	if afterReturns.TotalRefunded != beforeReturns.TotalRefunded || afterReturns.TotalReturns != beforeReturns.TotalReturns {
		t.Fatalf("financial returns report changed after deleting admin record: before=%+v after=%+v", beforeReturns, afterReturns)
	}
	if beforeStats["total_refunded"] != float64(200) || beforeStats["total_returns"] != 1 || afterStats["total_refunded"] != beforeStats["total_refunded"] || afterStats["total_returns"] != beforeStats["total_returns"] {
		t.Fatalf("return page statistics changed after deletion: before=%+v after=%+v", beforeStats, afterStats)
	}
	if len(beforeMonthly) != 1 || len(afterMonthly) != 1 || beforeMonthly[0].TotalRefundAmount != 200 || afterMonthly[0].TotalRefundAmount != beforeMonthly[0].TotalRefundAmount {
		t.Fatalf("monthly return report changed after deletion: before=%+v after=%+v", beforeMonthly, afterMonthly)
	}
	if len(beforeSalesReturns) != 1 || len(afterSalesReturns) != 1 || beforeSalesReturns[0].NetSales != 800 || afterSalesReturns[0].NetSales != beforeSalesReturns[0].NetSales || afterSalesReturns[0].ReturnsAmount != beforeSalesReturns[0].ReturnsAmount {
		t.Fatalf("sales/returns analysis changed after deletion: before=%+v after=%+v", beforeSalesReturns, afterSalesReturns)
	}
	saleDetails, err := salesrepo.NewRepository(db).GetSaleItems(ctx, uuid.MustParse(saleID))
	if err != nil {
		t.Fatalf("sale invoice details after return deletion: %v", err)
	}
	if len(saleDetails) != 1 || saleDetails[0].ReturnedQuantity != 2 || saleDetails[0].RemainingQuantity != 8 {
		t.Fatalf("sale history lost returned quantities: %+v", saleDetails)
	}

	var adminReturns, adminItems, inspectionCount, returnRefundCount, effectRefundCount, deleteAuditCount int
	var stock, debtRemaining, customerBalance, refundCount, movementCount, customerLedgerCount, effectCount, effectItemCount, saleCount, saleItemCount int
	var paymentStatus string
	var returnedUnitStatus string
	checks := []struct {
		query string
		args  []any
		dest  any
	}{
		{`SELECT COUNT(*) FROM returns WHERE id=?`, []any{returnID}, &adminReturns},
		{`SELECT COUNT(*) FROM return_items WHERE return_id=?`, []any{returnID}, &adminItems},
		{`SELECT COUNT(*) FROM return_inspection WHERE return_item_id='return-item-1'`, nil, &inspectionCount},
		{`SELECT COUNT(*) FROM return_refunds WHERE return_id=?`, []any{returnID}, &returnRefundCount},
		{`SELECT COUNT(*) FROM return_effect_refunds WHERE return_effect_id=?`, []any{returnID}, &effectRefundCount},
		{`SELECT COUNT(*) FROM audit_logs WHERE entity_id=? AND action='DELETE' AND entity_type='return'`, []any{returnID}, &deleteAuditCount},
		{`SELECT quantity FROM inventory WHERE id='inventory-1'`, nil, &stock},
		{`SELECT status FROM inventory_items WHERE id=?`, []any{inventoryItemID}, &returnedUnitStatus},
		{`SELECT remaining_amount FROM debts WHERE id=?`, []any{debtID}, &debtRemaining},
		{`SELECT current_balance FROM customers WHERE id=?`, []any{customerID}, &customerBalance},
		{`SELECT COUNT(*) FROM payment_refunds WHERE id=?`, []any{refundID}, &refundCount},
		{`SELECT status FROM payment_transactions WHERE id=?`, []any{transactionID}, &paymentStatus},
		{`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=? AND reference_type='return_effect'`, []any{returnID}, &movementCount},
		{`SELECT COUNT(*) FROM customer_ledger WHERE reference_id=? AND reference_type='return_effect'`, []any{returnID}, &customerLedgerCount},
		{`SELECT COUNT(*) FROM return_effects WHERE id=?`, []any{returnID}, &effectCount},
		{`SELECT COUNT(*) FROM return_effect_items WHERE return_effect_id=?`, []any{returnID}, &effectItemCount},
		{`SELECT COUNT(*) FROM sales WHERE id=?`, []any{saleID}, &saleCount},
		{`SELECT COUNT(*) FROM sale_items WHERE id=?`, []any{saleItemID}, &saleItemCount},
	}
	for _, check := range checks {
		if err := db.Get(check.dest, check.query, check.args...); err != nil {
			t.Fatalf("check %q: %v", check.query, err)
		}
	}
	if adminReturns != 0 || adminItems != 0 || inspectionCount != 0 || returnRefundCount != 0 || effectRefundCount != 1 || deleteAuditCount != 1 {
		t.Fatalf("admin/effect cleanup results return=%d items=%d inspections=%d refunds=%d effect_refunds=%d deletion_audit=%d", adminReturns, adminItems, inspectionCount, returnRefundCount, effectRefundCount, deleteAuditCount)
	}
	if stock != 5 || returnedUnitStatus != "AVAILABLE" || debtRemaining != 800 || customerBalance != 800 || refundCount != 1 || paymentStatus != "partially_refunded" || movementCount != 1 || customerLedgerCount != 1 {
		t.Fatalf("posted effects changed: stock=%d unit-status=%s debt=%d balance=%d refund=%d payment=%s movements=%d customer-ledger=%d", stock, returnedUnitStatus, debtRemaining, customerBalance, refundCount, paymentStatus, movementCount, customerLedgerCount)
	}
	if effectCount != 1 || effectItemCount != 1 || saleCount != 1 || saleItemCount != 1 {
		t.Fatalf("ledger/sales detail result effects=%d effect_items=%d sales=%d sale_items=%d", effectCount, effectItemCount, saleCount, saleItemCount)
	}
}

func TestDeleteCompletedReturnRollsBackWhenAuditWriteFails(t *testing.T) {
	db := openReturnDeleteDB(t)
	returnID := uuid.New().String()
	if _, err := db.Exec(`INSERT INTO returns (id, return_number, status, total_refund_amount, return_date, created_at) VALUES (?, 'RET-ROLLBACK', 'COMPLETED', 20, '2026-09-20', '2026-09-20')`, returnID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO return_items (id, return_id, quantity_returned, total_refund_amount, created_at) VALUES ('rollback-item', ?, 1, 20, '2026-09-20')`, returnID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_return_delete_audit BEFORE INSERT ON audit_logs WHEN NEW.action='DELETE' BEGIN SELECT RAISE(ABORT, 'audit unavailable'); END`); err != nil {
		t.Fatal(err)
	}
	if err := NewRepository(db).DeleteReturn(context.Background(), uuid.MustParse(returnID)); err == nil {
		t.Fatal("expected audit insert failure")
	}
	var returnsCount, itemsCount, effectsCount int
	_ = db.Get(&returnsCount, `SELECT COUNT(*) FROM returns WHERE id=?`, returnID)
	_ = db.Get(&itemsCount, `SELECT COUNT(*) FROM return_items WHERE return_id=?`, returnID)
	_ = db.Get(&effectsCount, `SELECT COUNT(*) FROM return_effects WHERE id=?`, returnID)
	if returnsCount != 1 || itemsCount != 1 || effectsCount != 0 {
		t.Fatalf("transaction did not roll back: returns=%d items=%d effects=%d", returnsCount, itemsCount, effectsCount)
	}
}

func openReturnDeleteDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE returns (id TEXT PRIMARY KEY, return_number TEXT, reference_number TEXT, sale_id TEXT, purchase_id TEXT, customer_id TEXT, total_refund_amount REAL DEFAULT 0, refund_status TEXT, status TEXT, return_date TEXT, refund_date TEXT, reason TEXT, refund_method TEXT, debt_id TEXT, debt_adjustment REAL DEFAULT 0, customer_credit REAL DEFAULT 0, created_at TEXT, updated_at TEXT, return_type TEXT, is_warranty_claim INTEGER DEFAULT 0, item_condition_after_return TEXT);
		CREATE TABLE return_items (id TEXT PRIMARY KEY, return_id TEXT, sale_item_id TEXT, product_id TEXT, inventory_item_id TEXT, serial_number TEXT, barcode TEXT, quantity_returned INTEGER DEFAULT 0, original_quantity INTEGER, unit_price REAL DEFAULT 0, total_refund_amount REAL DEFAULT 0, original_cost REAL, resolution TEXT, inventory_status TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE return_effects (id TEXT PRIMARY KEY, sale_id TEXT, purchase_id TEXT, customer_id TEXT, total_refund_amount REAL, status TEXT, return_date TEXT, refund_date TEXT, refund_method TEXT, debt_id TEXT, debt_adjustment REAL, customer_credit REAL, is_reversal INTEGER DEFAULT 0, created_at TEXT, updated_at TEXT);
		CREATE TABLE return_effect_items (id TEXT PRIMARY KEY, return_effect_id TEXT, sale_item_id TEXT, product_id TEXT, inventory_item_id TEXT, serial_number TEXT, barcode TEXT, quantity_returned INTEGER, original_quantity INTEGER, unit_price REAL, total_refund_amount REAL, original_cost REAL, resolution TEXT, inventory_status TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE return_effect_refunds (id TEXT PRIMARY KEY, return_effect_id TEXT, refund_type TEXT, amount REAL, refund_date TEXT, payment_method TEXT, transaction_reference TEXT, debt_id TEXT, debt_reduction_amount REAL, created_at TEXT);
		CREATE VIEW accounting_returns AS SELECT id, return_number, reference_number, sale_id, purchase_id, customer_id, total_refund_amount, refund_status, status, return_date, refund_date, reason, refund_method, debt_id, debt_adjustment, customer_credit, created_at, updated_at, return_type, is_warranty_claim, item_condition_after_return FROM returns UNION ALL SELECT id, NULL AS return_number, CASE WHEN is_reversal<>0 THEN 'REV-POSTED' ELSE NULL END AS reference_number, sale_id, purchase_id, customer_id, total_refund_amount, 'refunded' AS refund_status, status, return_date, refund_date, NULL AS reason, refund_method, debt_id, debt_adjustment, customer_credit, created_at, updated_at, CASE WHEN EXISTS (SELECT 1 FROM return_effect_items ri WHERE ri.return_effect_id=return_effects.id) AND (SELECT COALESCE(SUM(ri.quantity_returned),0) FROM return_effect_items ri WHERE ri.return_effect_id=return_effects.id)>=(SELECT COALESCE(SUM(COALESCE(ri.original_quantity,ri.quantity_returned)),0) FROM return_effect_items ri WHERE ri.return_effect_id=return_effects.id) THEN 'FULL' WHEN EXISTS (SELECT 1 FROM return_effect_items ri WHERE ri.return_effect_id=return_effects.id) THEN 'QUANTITY_PARTIAL' ELSE NULL END, 0, NULL FROM return_effects;
		CREATE VIEW accounting_return_items AS SELECT id, return_id, sale_item_id, product_id, inventory_item_id, serial_number, barcode, quantity_returned, original_quantity, unit_price, total_refund_amount, original_cost, resolution, inventory_status, created_at FROM return_items UNION ALL SELECT id, return_effect_id, sale_item_id, product_id, inventory_item_id, serial_number, barcode, quantity_returned, original_quantity, unit_price, total_refund_amount, original_cost, resolution, inventory_status, created_at FROM return_effect_items;
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, updated_at TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT, sold_at TEXT, updated_at TEXT, created_at TEXT, purchase_cost REAL, serial_number TEXT, barcode TEXT, supplier_id TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL, reference_id TEXT, reference_type TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL, updated_at TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, amount REAL, paid_amount REAL, remaining_amount REAL, status TEXT, updated_at TEXT);
		CREATE TABLE ledger_entries (id TEXT PRIMARY KEY, ledger_type TEXT, entity_id TEXT, transaction_type TEXT, amount REAL, reference_id TEXT, reference_type TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, event_type TEXT, event_date TEXT, reference_type TEXT, reference_id TEXT);
		CREATE TABLE payment_transactions (id TEXT PRIMARY KEY, amount_minor INTEGER, status TEXT, updated_at TEXT);
		CREATE TABLE payment_refunds (id TEXT PRIMARY KEY, payment_transaction_id TEXT, amount_minor INTEGER, status TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE return_payment_refunds (id TEXT PRIMARY KEY, return_id TEXT, payment_transaction_id TEXT, payment_refund_id TEXT, amount_minor INTEGER);
		CREATE TABLE return_refunds (id TEXT PRIMARY KEY, return_id TEXT, refund_type TEXT, amount REAL, refund_date TEXT, payment_method TEXT, transaction_reference TEXT, debt_id TEXT, debt_reduction_amount REAL, created_at TEXT);
		CREATE TABLE return_inspection (id TEXT PRIMARY KEY, return_item_id TEXT);
		CREATE TABLE return_audit_log (id TEXT PRIMARY KEY, return_id TEXT);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, supplier_id TEXT, customer_return_id TEXT);
		CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT, customer_return_id TEXT, purchase_item_id TEXT);
		CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, reference_id TEXT);
		CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT, current_balance REAL, updated_at TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, cost_price REAL, is_active INTEGER DEFAULT 1);
		CREATE TABLE sales (id TEXT PRIMARY KEY, invoice_number TEXT, sale_number TEXT, customer_id TEXT, sale_date TEXT, created_at TEXT, total_amount REAL, tax_amount REAL DEFAULT 0, discount_amount REAL DEFAULT 0, paid_amount REAL DEFAULT 0, cash_received REAL DEFAULT 0, change_amount REAL DEFAULT 0, payment_method TEXT, payment_status TEXT, status TEXT, cost_amount REAL, gross_profit REAL DEFAULT 0, user_id TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_price REAL, unit_cost REAL, total_amount REAL, tax_amount REAL DEFAULT 0, discount_amount REAL DEFAULT 0, supplier_id TEXT, barcode TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, expense_date TEXT, status TEXT, title TEXT, description TEXT);
		CREATE TABLE categories (id TEXT PRIMARY KEY, name TEXT);
		CREATE VIEW monthly_returns_analysis AS SELECT date(r.return_date, 'start of month') AS month, COUNT(DISTINCT r.id) AS total_returns, COUNT(DISTINCT r.customer_id) AS unique_customers, COALESCE(SUM(r.total_refund_amount),0) AS total_refund_amount, COALESCE(AVG(r.total_refund_amount),0) AS avg_refund_amount, COUNT(CASE WHEN UPPER(COALESCE(r.return_type,''))='FULL' THEN 1 END) AS full_returns, COUNT(CASE WHEN UPPER(COALESCE(r.return_type,''))='PARTIAL' THEN 1 END) AS partial_returns, COUNT(CASE WHEN UPPER(COALESCE(r.return_type,''))='QUANTITY_PARTIAL' THEN 1 END) AS quantity_partial_returns, COUNT(CASE WHEN UPPER(COALESCE(r.reason,''))='DEFECTIVE' THEN 1 END) AS defective_returns, COUNT(CASE WHEN UPPER(COALESCE(r.reason,''))='WARRANTY' THEN 1 END) AS warranty_returns, COUNT(CASE WHEN r.is_warranty_claim<>0 THEN 1 END) AS warranty_claims, SUM(CASE WHEN r.item_condition_after_return='SELLABLE' THEN 1 ELSE 0 END) AS sellable_items, SUM(CASE WHEN r.item_condition_after_return='NEEDS_REPAIR' THEN 1 ELSE 0 END) AS repair_needed, SUM(CASE WHEN r.item_condition_after_return='WRITE_OFF' THEN 1 ELSE 0 END) AS written_off FROM accounting_returns r WHERE UPPER(COALESCE(r.status,''))='COMPLETED' GROUP BY date(r.return_date,'start of month');
		CREATE VIEW sales_returns_analysis AS WITH returns_by_sale_month AS (SELECT sale_id, date(return_date,'start of month') AS month, SUM(total_refund_amount) AS returns_amount, COUNT(*) AS return_count FROM accounting_returns WHERE UPPER(COALESCE(status,''))='COMPLETED' AND sale_id IS NOT NULL GROUP BY sale_id, date(return_date,'start of month')) SELECT date(s.sale_date,'start of month') AS month, COUNT(DISTINCT s.id) AS total_sales, COALESCE(SUM(s.total_amount),0) AS gross_sales, COALESCE(SUM(s.cost_amount),0) AS total_cost, COALESCE(SUM(s.gross_profit),0) AS gross_profit, COALESCE(SUM(r.returns_amount),0) AS returns_amount, COALESCE(SUM(r.return_count),0) AS return_count, COALESCE(SUM(s.total_amount),0)-COALESCE(SUM(r.returns_amount),0) AS net_sales FROM sales s LEFT JOIN returns_by_sale_month r ON r.sale_id=s.id AND r.month=date(s.sale_date,'start of month') WHERE s.status='completed' GROUP BY date(s.sale_date,'start of month');
	`)
	if err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	return db
}
