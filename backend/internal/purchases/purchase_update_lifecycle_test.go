package purchases

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestPurchaseUpdateSynchronizesItemsInventorySupplierAndLedgerSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "purchase-update.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatalf("open local database: %v", err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	supplierOne := uuid.New()
	supplierTwo := uuid.New()
	productOne := uuid.New()
	productTwo := uuid.New()
	for _, supplier := range []struct {
		id   uuid.UUID
		code string
		name string
	}{
		{id: supplierOne, code: "SUP-UPDATE-1", name: "Supplier One"},
		{id: supplierTwo, code: "SUP-UPDATE-2", name: "Supplier Two"},
	} {
		if _, err := db.Exec(`INSERT INTO suppliers (id, code, name, current_balance, is_active, created_at, updated_at) VALUES (?, ?, ?, 0, 1, ?, ?)`, supplier.id, supplier.code, supplier.name, now, now); err != nil {
			t.Fatalf("insert supplier: %v", err)
		}
	}
	for _, product := range []struct {
		id   uuid.UUID
		sku  string
		name string
	}{
		{id: productOne, sku: "PUR-UPDATE-1", name: "Product One"},
		{id: productTwo, sku: "PUR-UPDATE-2", name: "Product Two"},
	} {
		if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, is_active, created_at, updated_at) VALUES (?, ?, ?, 0, 200, 1, ?, ?)`, product.id, product.sku, product.name, now, now); err != nil {
			t.Fatalf("insert product: %v", err)
		}
	}

	service := NewService(NewRepository(db), db)
	created, err := service.CreatePurchase(ctx, uuid.Nil, &PurchaseRequest{
		SupplierID:    supplierOne,
		InvoiceNumber: "UPDATE-001",
		PurchaseDate:  time.Now().UTC(),
		Items: []PurchaseItemRequest{{
			ProductID: productOne,
			Quantity:  2,
			UnitCost:  100,
			Condition: "new",
		}},
	})
	if err != nil {
		t.Fatalf("create purchase: %v", err)
	}
	firstItemID := created.Items[0].ID

	updated, err := service.UpdatePurchase(ctx, created.Purchase.ID, &PurchaseUpdateRequest{
		SupplierID: &supplierTwo,
		Items: []PurchaseItemRequest{
			{ID: &firstItemID, ProductID: productOne, Quantity: 1, UnitCost: 120, SellingPrice: 240, Condition: "new"},
			{ProductID: productTwo, Quantity: 3, UnitCost: 50, SellingPrice: 100, Condition: "new"},
		},
	})
	if err != nil {
		t.Fatalf("update pending purchase: %v", err)
	}
	if updated.Purchase.SupplierID != supplierTwo || updated.Purchase.TotalAmount != 270 || len(updated.Items) != 2 {
		t.Fatalf("unexpected pending update: supplier=%s total=%v items=%d", updated.Purchase.SupplierID, updated.Purchase.TotalAmount, len(updated.Items))
	}
	assertSupplierBalance(t, db, supplierOne, 0)
	assertSupplierBalance(t, db, supplierTwo, 270)

	if _, err := service.AddPayment(ctx, created.Purchase.ID, uuid.Nil, 100, "cash"); err != nil {
		t.Fatalf("add payment: %v", err)
	}
	if _, err := service.ReceivePurchase(ctx, created.Purchase.ID, uuid.Nil); err != nil {
		t.Fatalf("receive purchase: %v", err)
	}

	current, err := service.GetPurchase(ctx, created.Purchase.ID)
	if err != nil {
		t.Fatal(err)
	}
	var productOneItem PurchaseItem
	for _, item := range current.Items {
		if item.ProductID == productOne {
			productOneItem = item
		}
	}
	productOneItemID := productOneItem.ID
	updated, err = service.UpdatePurchase(ctx, created.Purchase.ID, &PurchaseUpdateRequest{
		SupplierID: &supplierOne,
		Items: []PurchaseItemRequest{{
			ID: &productOneItemID, ProductID: productOne, Quantity: 2, UnitCost: 120, SellingPrice: 240, Condition: "new",
		}},
	})
	if err != nil {
		t.Fatalf("update received purchase: %v", err)
	}
	if updated.Purchase.TotalAmount != 240 || len(updated.Items) != 1 {
		t.Fatalf("unexpected received update: total=%v items=%d", updated.Purchase.TotalAmount, len(updated.Items))
	}
	var productOneStock, productTwoStock int
	if err := db.Get(&productOneStock, `SELECT COUNT(*) FROM inventory_items WHERE product_id = ?`, productOne); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&productTwoStock, `SELECT COUNT(*) FROM inventory_items WHERE product_id = ?`, productTwo); err != nil {
		t.Fatal(err)
	}
	if productOneStock != 2 || productTwoStock != 0 {
		t.Fatalf("inventory not synchronized: product one=%d product two=%d", productOneStock, productTwoStock)
	}
	var movedInventory, movedPayments int
	if err := db.Get(&movedInventory, `SELECT COUNT(*) FROM inventory_items WHERE product_id = ? AND supplier_id = ?`, productOne, supplierOne); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&movedPayments, `SELECT COUNT(*) FROM payments WHERE purchase_id = ? AND supplier_id = ?`, created.Purchase.ID, supplierOne); err != nil {
		t.Fatal(err)
	}
	if movedInventory != 2 || movedPayments != 1 {
		t.Fatalf("supplier links not moved atomically: inventory=%d payments=%d", movedInventory, movedPayments)
	}
	assertSupplierBalance(t, db, supplierOne, 140)
	assertSupplierBalance(t, db, supplierTwo, 0)

	_, err = service.UpdatePurchase(ctx, created.Purchase.ID, &PurchaseUpdateRequest{Items: []PurchaseItemRequest{{
		ID: &productOneItemID, ProductID: productOne, Quantity: 2, UnitCost: 40, SellingPrice: 240, Condition: "new",
	}}})
	if err != ErrPaymentExceedsTotal {
		t.Fatalf("expected update below paid amount to fail with ErrPaymentExceedsTotal, got %v", err)
	}
	unchangedPaidPurchase, err := service.GetPurchase(ctx, created.Purchase.ID)
	if err != nil {
		t.Fatal(err)
	}
	if unchangedPaidPurchase.Purchase.TotalAmount != 240 || unchangedPaidPurchase.Purchase.PaidAmount != 100 || unchangedPaidPurchase.Items[0].UnitCost != 120 {
		t.Fatalf("rejected edit changed paid purchase: total=%v paid=%v unit_cost=%v", unchangedPaidPurchase.Purchase.TotalAmount, unchangedPaidPurchase.Purchase.PaidAmount, unchangedPaidPurchase.Items[0].UnitCost)
	}
	assertSupplierBalance(t, db, supplierOne, 140)

	if _, err := db.Exec(`UPDATE inventory_items SET status = 'SOLD' WHERE product_id = ?`, productOne); err != nil {
		t.Fatal(err)
	}
	_, err = service.UpdatePurchase(ctx, created.Purchase.ID, &PurchaseUpdateRequest{Items: []PurchaseItemRequest{{
		ID: &productOneItemID, ProductID: productOne, Quantity: 1, UnitCost: 120, SellingPrice: 240, Condition: "new",
	}}})
	if err == nil {
		t.Fatal("expected reducing below used inventory to be rejected")
	}
	unchanged, err := service.GetPurchase(ctx, created.Purchase.ID)
	if err != nil {
		t.Fatal(err)
	}
	if unchanged.Items[0].Quantity != 2 || unchanged.Purchase.TotalAmount != 240 {
		t.Fatalf("failed edit was not rolled back: quantity=%d total=%v", unchanged.Items[0].Quantity, unchanged.Purchase.TotalAmount)
	}
}

func assertSupplierBalance(t *testing.T, db *sqlx.DB, supplierID uuid.UUID, want float64) {
	t.Helper()
	var got float64
	if err := db.Get(&got, `SELECT COALESCE(current_balance, 0) FROM suppliers WHERE id = ?`, supplierID); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("supplier %s balance=%v, want %v", supplierID, got, want)
	}
}

func TestCancelPurchaseReversesSupplierBalanceAndBlocksPaidOrReceivedPurchases(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "purchase-cancel.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatalf("open local database: %v", err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	supplierID, productID := uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO suppliers (id,code,name,current_balance,is_active,created_at,updated_at) VALUES (?,?,?,?,1,?,?)`, supplierID, "SUP-CANCEL", "Supplier Cancel", 0, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id,sku,name,cost_price,selling_price,is_active,created_at,updated_at) VALUES (?,?,?,0,200,1,?,?)`, productID, "PUR-CANCEL", "Product Cancel", now, now); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db), db)
	createPurchase := func(invoice string) *PurchaseResponse {
		t.Helper()
		created, err := service.CreatePurchase(ctx, uuid.Nil, &PurchaseRequest{
			SupplierID: supplierID, InvoiceNumber: invoice, PurchaseDate: time.Now().UTC(),
			Items: []PurchaseItemRequest{{ProductID: productID, Quantity: 2, UnitCost: 100, Condition: "new"}},
		})
		if err != nil {
			t.Fatalf("create purchase %s: %v", invoice, err)
		}
		return created
	}

	unpaid := createPurchase("CANCEL-UNPAID")
	if _, err := service.CancelPurchase(ctx, uuid.Nil, unpaid.Purchase.ID); err != nil {
		t.Fatalf("cancel unpaid purchase: %v", err)
	}
	var ledgerRows int
	if err := db.Get(&ledgerRows, `SELECT COUNT(*) FROM supplier_ledger WHERE reference_id = ?`, unpaid.Purchase.ID); err != nil {
		t.Fatal(err)
	}
	if ledgerRows != 2 {
		t.Fatalf("supplier ledger rows = %d, want original debit and cancellation credit", ledgerRows)
	}
	assertSupplierBalance(t, db, supplierID, 0)

	paid := createPurchase("CANCEL-PAID")
	if _, err := service.AddPayment(ctx, paid.Purchase.ID, uuid.Nil, 40, "cash"); err != nil {
		t.Fatalf("add partial purchase payment: %v", err)
	}
	if _, err := service.CancelPurchase(ctx, uuid.Nil, paid.Purchase.ID); err != ErrInvalidPurchaseStatus {
		t.Fatalf("cancel paid purchase error = %v, want ErrInvalidPurchaseStatus", err)
	}
	paidAfter, err := service.GetPurchase(ctx, paid.Purchase.ID)
	if err != nil {
		t.Fatal(err)
	}
	if paidAfter.Purchase.Status != StatusPending || paidAfter.Purchase.PaidAmount != 40 {
		t.Fatalf("blocked paid purchase changed: status=%s paid=%v", paidAfter.Purchase.Status, paidAfter.Purchase.PaidAmount)
	}
	assertSupplierBalance(t, db, supplierID, 160)

	partiallyReceived := createPurchase("CANCEL-PARTIAL")
	if _, err := db.Exec(`UPDATE purchases SET status='partially_received' WHERE id=?`, partiallyReceived.Purchase.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CancelPurchase(ctx, uuid.Nil, partiallyReceived.Purchase.ID); err != ErrPurchaseAlreadyReceived {
		t.Fatalf("cancel partially received purchase error = %v, want ErrPurchaseAlreadyReceived", err)
	}
	partialAfter, err := service.GetPurchase(ctx, partiallyReceived.Purchase.ID)
	if err != nil {
		t.Fatal(err)
	}
	if partialAfter.Purchase.Status != StatusPartiallyReceived {
		t.Fatalf("blocked partial receipt changed status to %s", partialAfter.Purchase.Status)
	}
}
