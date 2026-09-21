package supplierreturns

import (
	"context"
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
	created, err := service.Create(ctx, uuid.New(), CreateRequest{PurchaseID: purchaseID, Reason: "defective"})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.AddItem(ctx, created.ID, AddItemRequest{PurchaseItemID: purchaseItemID, Quantity: 1}); err != nil {
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
