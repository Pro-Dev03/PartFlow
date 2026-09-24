package returns

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestCreateReturnRollsBackParentAndItemsWhenAnItemFailsSQLite(t *testing.T) {
	db, service, saleID, products := openReturnLifecycleTest(t)
	if _, err := db.Exec(`CREATE TRIGGER fail_selected_return_item BEFORE INSERT ON return_items WHEN NEW.barcode = 'FAIL' BEGIN SELECT RAISE(ABORT, 'forced item failure'); END`); err != nil {
		t.Fatal(err)
	}

	_, err := service.CreateReturn(context.Background(), uuid.Nil, &ReturnRequest{
		SaleID:                   &saleID,
		ReturnDate:               time.Now().UTC(),
		ReturnType:               "PARTIAL",
		Reason:                   "DEFECTIVE",
		ItemConditionAfterReturn: "NOT_FOR_SALE",
		RefundMethod:             "CASH",
		Items: []ReturnItemRequest{
			{ProductID: &products[0], Barcode: "OK", QuantityReturned: 1, UnitPrice: 100, TotalRefundAmount: 100, ReturnedCondition: "DAMAGED", Resolution: "WRITE_OFF"},
			{ProductID: &products[1], Barcode: "FAIL", QuantityReturned: 1, UnitPrice: 50, TotalRefundAmount: 50, ReturnedCondition: "DAMAGED", Resolution: "WRITE_OFF"},
		},
	})
	if err == nil {
		t.Fatal("expected the forced second item failure")
	}
	for _, table := range []string{"returns", "return_items"} {
		var count int
		if err := db.Get(&count, `SELECT COUNT(*) FROM `+table); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s retained %d partial rows after rollback", table, count)
		}
	}
}

func TestReturnItemMutationsRecalculateParentTotalAtomicallySQLite(t *testing.T) {
	db, service, saleID, products := openReturnLifecycleTest(t)
	ctx := context.Background()
	created, err := service.CreateReturn(ctx, uuid.Nil, &ReturnRequest{
		SaleID:                   &saleID,
		ReturnDate:               time.Now().UTC(),
		ReturnType:               "PARTIAL",
		Reason:                   "DEFECTIVE",
		ItemConditionAfterReturn: "NOT_FOR_SALE",
		RefundMethod:             "CASH",
		Items: []ReturnItemRequest{
			{ProductID: &products[0], QuantityReturned: 1, UnitPrice: 100, TotalRefundAmount: 100, ReturnedCondition: "DAMAGED", Resolution: "WRITE_OFF"},
			{ProductID: &products[1], QuantityReturned: 1, UnitPrice: 60, TotalRefundAmount: 60, ReturnedCondition: "DAMAGED", Resolution: "WRITE_OFF"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertReturnTotal(t, db, created.Return.ID, 160)

	first, err := service.UpdateReturnItem(ctx, created.Items[0].ID, ReturnItemRequest{QuantityReturned: 2, UnitPrice: 75})
	if err != nil {
		t.Fatal(err)
	}
	if first.TotalRefundAmount != 150 {
		t.Fatalf("updated item total=%v, want 150", first.TotalRefundAmount)
	}
	assertReturnTotal(t, db, created.Return.ID, 210)

	if err := service.DeleteReturnItem(ctx, created.Items[1].ID); err != nil {
		t.Fatal(err)
	}
	assertReturnTotal(t, db, created.Return.ID, 150)
	if err := service.DeleteReturnItem(ctx, created.Items[0].ID); err != ErrNoItems {
		t.Fatalf("deleting the final item error=%v, want %v", err, ErrNoItems)
	}
	assertReturnTotal(t, db, created.Return.ID, 150)
}

func openReturnLifecycleTest(t *testing.T) (*sqlx.DB, *Service, uuid.UUID, []uuid.UUID) {
	t.Helper()
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "return-lifecycle.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = local.DB.Close() })
	db := sqlx.NewDb(local.DB, "sqlite")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	products := []uuid.UUID{uuid.New(), uuid.New()}
	for index, productID := range products {
		if _, err := db.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, productID.String(), "RET-LIFE-"+string(rune('A'+index)), "Return lifecycle product", now, now); err != nil {
			t.Fatal(err)
		}
	}
	saleID := uuid.New()
	if _, err := db.Exec(`INSERT INTO sales (id,sale_number,invoice_number,sale_date,subtotal,discount_amount,total_amount,paid_amount,remaining_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, saleID.String(), "S-RET-LIFE", "INV-RET-LIFE", now, 500, 0, 500, 500, 0, "cash", "paid", "completed", now, now); err != nil {
		t.Fatal(err)
	}
	return db, NewService(NewRepository(db)), saleID, products
}

func assertReturnTotal(t *testing.T, db *sqlx.DB, returnID uuid.UUID, want float64) {
	t.Helper()
	var got float64
	if err := db.Get(&got, `SELECT total_refund_amount FROM returns WHERE id = ?`, returnID.String()); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("return total=%v, want %v", got, want)
	}
}
