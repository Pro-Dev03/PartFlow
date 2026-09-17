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
