package supplierreturns

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestSupplierReturnCreditsLedgerAndRemovesInventorySQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/supplier-return-lifecycle.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	ctx := context.Background()
	supplierID, productID, purchaseID := uuid.New(), uuid.New(), uuid.New()
	purchaseItemID, inventoryItemID := uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := db.Exec(`INSERT INTO suppliers (id, code, name, current_balance, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, 1, ?, ?)`, supplierID, "SUP-RETURN", "Return Supplier", 100, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, productID, "RET-001", "Return Product", 100, 150, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, purchase_date, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID, "PUR-RETURN", supplierID, now, 100, 0, "received", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, purchaseItemID, purchaseID, productID, 1, 100, 100, now); err != nil {
		t.Fatal(err)
	}
	itemCode := "ITM-" + purchaseID.String()[:8] + "-001"
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, inventoryItemID, productID, itemCode, "RET-BC-001", "USED", 100, 150, "AVAILABLE", supplierID, now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, reserved_quantity, created_at, updated_at) VALUES (?, ?, ?, 0, ?, ?)`, uuid.New(), productID, 1, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, description, reference_id, created_at) VALUES (?, ?, 'debit', 'PURCHASE', ?, ?, 'Purchase', ?, ?)`, uuid.New(), supplierID, 100, 100, purchaseID, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(db)
	_, err = service.Create(ctx, uuid.New(), CreateRequest{
		PurchaseID: purchaseID, Reason: "defective", PurchaseItemID: uuid.New(), Quantity: 1,
	})
	if err == nil {
		t.Fatal("creating a return with an unavailable purchase item should fail")
	}
	var returnCount int
	if err := db.Get(&returnCount, `SELECT COUNT(*) FROM supplier_returns WHERE purchase_id = ?`, purchaseID); err != nil {
		t.Fatal(err)
	}
	if returnCount != 0 {
		t.Fatalf("failed supplier return creation left %d header rows, want 0", returnCount)
	}

	created, err := service.Create(ctx, uuid.New(), CreateRequest{
		PurchaseID: purchaseID, Reason: "defective", PurchaseItemID: purchaseItemID, Quantity: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Complete(ctx, created.ID, uuid.Nil); err != nil {
		t.Fatal(err)
	}

	var status string
	if err := db.Get(&status, `SELECT status FROM inventory_items WHERE id = ?`, inventoryItemID); err != nil {
		t.Fatal(err)
	}
	if status != "RETURNED" {
		t.Fatalf("inventory status = %q, want RETURNED", status)
	}
	var creditCount int
	if err := db.Get(&creditCount, `SELECT COUNT(*) FROM supplier_ledger WHERE supplier_id = ? AND transaction_type = 'SUPPLIER_RETURN'`, supplierID); err != nil {
		t.Fatal(err)
	}
	if creditCount != 1 {
		t.Fatalf("supplier return credits = %d, want 1", creditCount)
	}
	var refund float64
	if err := db.Get(&refund, `SELECT refund_amount FROM supplier_returns WHERE id = ?`, created.ID); err != nil {
		t.Fatal(err)
	}
	if refund != 100 {
		t.Fatalf("refund amount = %v, want 100", refund)
	}
	// A mismatch must roll back the whole cleanup without changing the live
	// balance or deleting any part of the operation.
	if _, err := db.Exec(`UPDATE supplier_ledger SET amount = 99 WHERE reference_id = ? AND transaction_type = 'SUPPLIER_RETURN'`, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(ctx, created.ID, true); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("mismatched completed return cleanup error = %v, want explicit ledger mismatch", err)
	}
	var retainedReturnCount, retainedItemCount, retainedCreditCount, retainedMovementCount, snapshotCount int
	if err := db.Get(&retainedReturnCount, `SELECT COUNT(*) FROM supplier_returns WHERE id = ? AND status = 'COMPLETED'`, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&retainedItemCount, `SELECT COUNT(*) FROM supplier_return_items WHERE supplier_return_id = ?`, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&retainedCreditCount, `SELECT COUNT(*) FROM supplier_ledger WHERE supplier_id = ? AND transaction_type = 'SUPPLIER_RETURN'`, supplierID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&retainedMovementCount, `SELECT COUNT(*) FROM inventory_movements WHERE reference_id = ? AND movement_type = 'SUPPLIER_RETURN'`, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&snapshotCount, `SELECT COUNT(*) FROM deleted_operation_snapshots WHERE entity_type = 'supplier_return' AND operation_id = ?`, created.ID); err != nil {
		t.Fatal(err)
	}
	var balanceAfterRejectedDelete float64
	if err := db.Get(&balanceAfterRejectedDelete, `SELECT current_balance FROM suppliers WHERE id = ?`, supplierID); err != nil {
		t.Fatal(err)
	}
	if retainedReturnCount != 1 || retainedItemCount != 1 || retainedCreditCount != 1 || retainedMovementCount != 1 || snapshotCount != 0 || balanceAfterRejectedDelete != 0 {
		t.Fatalf("mismatched cleanup changed data: returns=%d items=%d credits=%d movements=%d snapshots=%d balance=%v", retainedReturnCount, retainedItemCount, retainedCreditCount, retainedMovementCount, snapshotCount, balanceAfterRejectedDelete)
	}
	if _, err := db.Exec(`UPDATE supplier_ledger SET amount = 100 WHERE reference_id = ? AND transaction_type = 'SUPPLIER_RETURN'`, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.Delete(ctx, created.ID, true); err != nil {
		t.Fatalf("clean up completed, fully posted supplier return: %v", err)
	}
	var deletedReturnCount, deletedItemCount, keptCreditCount, keptMovementCount, keptSnapshotCount, auditCount int
	if err := db.Get(&deletedReturnCount, `SELECT COUNT(*) FROM supplier_returns WHERE id = ?`, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&deletedItemCount, `SELECT COUNT(*) FROM supplier_return_items WHERE supplier_return_id = ?`, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&keptCreditCount, `SELECT COUNT(*) FROM supplier_ledger WHERE reference_id = ? AND reference_type = 'supplier_return_effect'`, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&keptMovementCount, `SELECT COUNT(*) FROM inventory_movements WHERE reference_id = ? AND reference_type = 'supplier_return_effect'`, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&keptSnapshotCount, `SELECT COUNT(*) FROM deleted_operation_snapshots WHERE entity_type = 'supplier_return' AND operation_id = ?`, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&auditCount, `SELECT COUNT(*) FROM audit_logs WHERE entity_type = 'supplier_return' AND entity_id = ? AND action = 'DELETE'`, created.ID); err != nil {
		t.Fatal(err)
	}
	var balanceAfterCleanup float64
	var inventoryAfterCleanup int
	if err := db.Get(&balanceAfterCleanup, `SELECT current_balance FROM suppliers WHERE id = ?`, supplierID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&inventoryAfterCleanup, `SELECT COALESCE(SUM(quantity), 0) FROM inventory WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if deletedReturnCount != 0 || deletedItemCount != 0 || keptCreditCount != 1 || keptMovementCount != 1 || keptSnapshotCount != 1 || auditCount != 1 {
		t.Fatalf("cleanup retained wrong operational/history rows: returns=%d items=%d credits=%d movements=%d snapshots=%d audit=%d", deletedReturnCount, deletedItemCount, keptCreditCount, keptMovementCount, keptSnapshotCount, auditCount)
	}
	if balanceAfterCleanup != 0 || inventoryAfterCleanup != 0 {
		t.Fatalf("cleanup changed settled balances: supplier=%v inventory=%d, want 0/0", balanceAfterCleanup, inventoryAfterCleanup)
	}
}

func TestSupplierReturnCompleteAllowsMultipleInventoryRowsForSameProduct(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/supplier-return-multi-inventory.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	ctx := context.Background()
	supplierID, productID, purchaseID := uuid.New(), uuid.New(), uuid.New()
	purchaseItemID, inventoryItemID1, inventoryItemID2 := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := db.Exec(`INSERT INTO suppliers (id, code, name, current_balance, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, 1, ?, ?)`, supplierID, "SUP-MULTI", "Multi Inventory Supplier", 0, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, productID, "RET-002", "Return Product Multi", 100, 150, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, purchase_date, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID, "PUR-MULTI", supplierID, now, 200, 0, "received", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, purchaseItemID, purchaseID, productID, 2, 100, 200, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, inventoryItemID1, productID, "ITM-"+purchaseID.String()[:8]+"-001", "RET-BC-001", "NEW", 100, 150, "AVAILABLE", supplierID, now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, inventoryItemID2, productID, "ITM-"+purchaseID.String()[:8]+"-002", "RET-BC-002", "NEW", 100, 150, "AVAILABLE", supplierID, now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, reserved_quantity, created_at, updated_at) VALUES (?, ?, ?, 0, ?, ?), (?, ?, ?, 0, ?, ?)`, uuid.New(), productID, 1, now, now, uuid.New(), productID, 1, now, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(db)
	created, err := service.Create(ctx, uuid.New(), CreateRequest{PurchaseID: purchaseID, Reason: "defective"})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.AddItem(ctx, created.ID, AddItemRequest{PurchaseItemID: purchaseItemID, Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	if err := service.Complete(ctx, created.ID, uuid.Nil); err != nil {
		t.Fatalf("complete supplier return with multiple inventory rows should succeed: %v", err)
	}

	var status string
	if err := db.Get(&status, `SELECT status FROM inventory_items WHERE id = ?`, inventoryItemID1); err != nil {
		t.Fatal(err)
	}
	if status != "RETURNED" {
		t.Fatalf("first inventory item status = %q, want RETURNED", status)
	}
	var total int
	if err := db.Get(&total, `SELECT COALESCE(SUM(quantity), 0) FROM inventory WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("aggregate inventory after supplier return = %d, want 1", total)
	}
}

func TestSupplierReturnCompleteUsesAvailableInventoryItemsWhenAggregateIsStale(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/supplier-return-stale-aggregate.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	ctx := context.Background()
	supplierID, productID, purchaseID := uuid.New(), uuid.New(), uuid.New()
	purchaseItemID, inventoryItemID1, inventoryItemID2, inventoryItemID3, inventoryItemID4 := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := db.Exec(`INSERT INTO suppliers (id, code, name, current_balance, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, 1, ?, ?)`, supplierID, "SUP-STALE", "Stale Aggregate Supplier", 0, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, productID, "RET-003", "Stale Aggregate Product", 100, 150, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, purchase_date, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID, "PUR-STALE", supplierID, now, 400, 0, "received", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, purchaseItemID, purchaseID, productID, 4, 100, 400, now); err != nil {
		t.Fatal(err)
	}
	for _, itemID := range []uuid.UUID{inventoryItemID1, inventoryItemID2, inventoryItemID3, inventoryItemID4} {
		if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, itemID, productID, "ITM-"+purchaseID.String()[:8]+"-00"+uuid.NewString()[:2], "BC-"+itemID.String()[:8], "NEW", 100, 150, "AVAILABLE", supplierID, now, now, now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, reserved_quantity, created_at, updated_at) VALUES (?, ?, ?, 0, ?, ?)`, uuid.New(), productID, 0, now, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(db)
	created, err := service.Create(ctx, uuid.New(), CreateRequest{PurchaseID: purchaseID, Reason: "defective"})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.AddItem(ctx, created.ID, AddItemRequest{PurchaseItemID: purchaseItemID, Quantity: 4}); err != nil {
		t.Fatal(err)
	}
	if err := service.Complete(ctx, created.ID, uuid.Nil); err != nil {
		t.Fatalf("complete supplier return should succeed when inventory_items have stock even if inventory aggregate is stale: %v", err)
	}

	var returnedCount int
	if err := db.Get(&returnedCount, `SELECT COUNT(*) FROM inventory_items WHERE product_id = ? AND status = 'RETURNED'`, productID); err != nil {
		t.Fatal(err)
	}
	if returnedCount != 4 {
		t.Fatalf("returned inventory items = %d, want 4", returnedCount)
	}
	var resultStatus string
	if err := db.Get(&resultStatus, `SELECT status FROM supplier_returns WHERE id = ?`, created.ID); err != nil {
		t.Fatal(err)
	}
	if resultStatus != "COMPLETED" {
		t.Fatalf("supplier return status = %q, want COMPLETED", resultStatus)
	}
}
