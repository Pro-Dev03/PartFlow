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

type testDependentSaleDeleter struct {
	db       *sqlx.DB
	prepared []uuid.UUID
	deleted  []uuid.UUID
}

func (d *testDependentSaleDeleter) PrepareDelete(_ context.Context, saleID, _ uuid.UUID) error {
	d.prepared = append(d.prepared, saleID)
	return nil
}

func (d *testDependentSaleDeleter) DeleteInTransaction(ctx context.Context, tx *sqlx.Tx, saleID, _ uuid.UUID) error {
	if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM inventory_movements WHERE reference_type='sale' AND reference_id=?`), saleID.String()); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory SET quantity=quantity+1 WHERE product_id=(SELECT product_id FROM sale_items WHERE sale_id=? LIMIT 1)`), saleID.String()); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`UPDATE inventory_items SET status='AVAILABLE' WHERE id IN (SELECT inventory_item_id FROM sale_items WHERE sale_id=?)`), saleID.String()); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM sale_items WHERE sale_id=?`), saleID.String()); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM sales WHERE id=?`), saleID.String()); err != nil {
		return err
	}
	d.deleted = append(d.deleted, saleID)
	return nil
}

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

func TestDeleteReturnReversesLaterSaleBeforeRemovingReturnedInventorySQLite(t *testing.T) {
	db, service, originalSaleID, products := openReturnLifecycleTest(t)
	ctx := context.Background()
	productID := products[0]
	now := time.Now().UTC().Format(time.RFC3339Nano)
	itemID, originalSaleItemID, returnID, returnItemID, laterSaleID, laterSaleItemID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO inventory_items (id,item_code,product_id,status,condition,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`, itemID.String(), "RET-CASCADE-ITEM", productID.String(), "SOLD", "NEW", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,reserved_quantity,created_at,updated_at) VALUES (?,?,0,0,?,?)`, uuid.NewString(), productID.String(), now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id,sale_id,product_id,inventory_item_id,quantity,unit_price,item_total,created_at) VALUES (?,?,?,?,1,100,100,?)`, originalSaleItemID.String(), originalSaleID.String(), productID.String(), itemID.String(), now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO returns (id,return_number,sale_id,total_refund_amount,refund_status,status,reason,return_date,created_at,updated_at,refund_method,debt_adjustment) VALUES (?,?,?,100,'completed','COMPLETED','defective',?,?,?,'CASH',0)`, returnID.String(), "RET-CASCADE-1", originalSaleID.String(), now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO return_items (id,return_id,sale_item_id,product_id,inventory_item_id,quantity,quantity_returned,original_quantity,unit_price,total_refund_amount,resolution,inventory_status,created_at,updated_at) VALUES (?,?,?,?,?,1,1,1,100,100,'RESTOCK','AVAILABLE',?,?)`, returnItemID.String(), returnID.String(), originalSaleItemID.String(), productID.String(), itemID.String(), now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,created_at) VALUES (?,?,?,'RETURN',1,0,1,'return',?,?)`, uuid.NewString(), itemID.String(), productID.String(), returnID.String(), now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id,sale_number,invoice_number,sale_date,subtotal,discount_amount,total_amount,paid_amount,remaining_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, laterSaleID.String(), "S-LATER-RET-CASCADE", "INV-LATER-RET-CASCADE", now, 100, 0, 100, 100, 0, "cash", "paid", "completed", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id,sale_id,product_id,inventory_item_id,quantity,unit_price,item_total,created_at) VALUES (?,?,?,?,1,100,100,?)`, laterSaleItemID.String(), laterSaleID.String(), productID.String(), itemID.String(), now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,created_at) VALUES (?,?,?,'SALE',-1,1,0,'sale',?,?)`, uuid.NewString(), itemID.String(), productID.String(), laterSaleID.String(), now); err != nil {
		t.Fatal(err)
	}
	coordinator := &testDependentSaleDeleter{db: db}
	service.SetDependentSaleDeleteCoordinator(coordinator)

	if err := service.DeleteReturn(ctx, returnID); err != nil {
		t.Fatalf("delete return cascade: %v", err)
	}
	if len(coordinator.prepared) != 1 || coordinator.prepared[0] != laterSaleID || len(coordinator.deleted) != 1 || coordinator.deleted[0] != laterSaleID {
		t.Fatalf("dependent sale lifecycle: prepared=%v deleted=%v", coordinator.prepared, coordinator.deleted)
	}
	var deletedReturn, deletedSale, inventoryQuantity int
	var itemStatus string
	for _, check := range []struct {
		query string
		args  []any
		dest  *int
	}{{`SELECT COUNT(*) FROM returns WHERE id=?`, []any{returnID.String()}, &deletedReturn}, {`SELECT COUNT(*) FROM sales WHERE id=?`, []any{laterSaleID.String()}, &deletedSale}, {`SELECT quantity FROM inventory WHERE product_id=?`, []any{productID.String()}, &inventoryQuantity}} {
		if err := db.QueryRow(check.query, check.args...).Scan(check.dest); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Get(&itemStatus, `SELECT status FROM inventory_items WHERE id=?`, itemID.String()); err != nil {
		t.Fatal(err)
	}
	if deletedReturn != 0 || deletedSale != 0 || inventoryQuantity != 0 || itemStatus != "SOLD" {
		t.Fatalf("cascade result: return=%d later_sale=%d quantity=%d item_status=%s; want both transactions removed and original sold state restored", deletedReturn, deletedSale, inventoryQuantity, itemStatus)
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
	if err := service.DeleteReturnItem(ctx, created.Items[0].ID); err != nil {
		t.Fatalf("deleting the final item should hard-delete the empty return: %v", err)
	}
	var remaining int
	if err := db.Get(&remaining, `SELECT COUNT(*) FROM returns WHERE id=?`, created.Return.ID); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("return remains after deleting final item, count=%d", remaining)
	}
}

func TestConcurrentReturnsDoNotExceedSaleItemQuantitySQLite(t *testing.T) {
	db, service, saleID, products := openReturnLifecycleTest(t)
	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	saleItemID := uuid.New()
	if _, err := db.Exec(`INSERT INTO sale_items (id,sale_id,product_id,quantity,unit_price,item_total,created_at) VALUES (?,?,?,?,?,?,?)`, saleItemID.String(), saleID.String(), products[0].String(), 1, 100, 100, now); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			_, err := service.CreateReturn(ctx, uuid.Nil, &ReturnRequest{
				SaleID:                   &saleID,
				ReturnDate:               time.Now().UTC(),
				ReturnType:               "PARTIAL",
				Reason:                   "DEFECTIVE",
				ItemConditionAfterReturn: "NOT_FOR_SALE",
				RefundMethod:             "CASH",
				Items: []ReturnItemRequest{{
					SaleItemID:        &saleItemID,
					ProductID:         &products[0],
					QuantityReturned:  1,
					UnitPrice:         100,
					TotalRefundAmount: 100,
					ReturnedCondition: "DAMAGED",
					Resolution:        "WRITE_OFF",
				}},
			})
			results <- err
		}()
	}
	close(start)

	successes := 0
	for i := 0; i < 2; i++ {
		if err := <-results; err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful concurrent returns = %d, want 1", successes)
	}
	var returned int
	if err := db.Get(&returned, `SELECT COALESCE(SUM(quantity_returned),0) FROM return_items WHERE sale_item_id = ?`, saleItemID.String()); err != nil {
		t.Fatal(err)
	}
	if returned != 1 {
		t.Fatalf("returned quantity = %d, sold quantity = 1", returned)
	}
}

func TestReturnCannotReferenceSaleItemFromAnotherSaleSQLite(t *testing.T) {
	db, service, saleID, products := openReturnLifecycleTest(t)
	ctx := context.Background()
	otherSaleID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO sales (id,sale_number,invoice_number,sale_date,subtotal,discount_amount,total_amount,paid_amount,remaining_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, otherSaleID.String(), "S-OTHER-RETURN", "INV-OTHER-RETURN", now, 100, 0, 100, 100, 0, "cash", "paid", "completed", now, now); err != nil {
		t.Fatal(err)
	}
	saleItemID := uuid.New()
	if _, err := db.Exec(`INSERT INTO sale_items (id,sale_id,product_id,quantity,unit_price,item_total,created_at) VALUES (?,?,?,?,?,?,?)`, saleItemID.String(), otherSaleID.String(), products[1].String(), 1, 100, 100, now); err != nil {
		t.Fatal(err)
	}

	_, err := service.CreateReturn(ctx, uuid.Nil, &ReturnRequest{
		SaleID:                   &saleID,
		ReturnDate:               time.Now().UTC(),
		ReturnType:               "PARTIAL",
		Reason:                   "DEFECTIVE",
		ItemConditionAfterReturn: "NOT_FOR_SALE",
		RefundMethod:             "CASH",
		Items: []ReturnItemRequest{{
			SaleItemID:        &saleItemID,
			ProductID:         &products[1],
			QuantityReturned:  1,
			UnitPrice:         100,
			TotalRefundAmount: 100,
			ReturnedCondition: "DAMAGED",
			Resolution:        "WRITE_OFF",
		}},
	})
	if err != ErrSaleItemNotFound {
		t.Fatalf("cross-sale item return error = %v, want %v", err, ErrSaleItemNotFound)
	}
	var returnCount int
	if err := db.Get(&returnCount, `SELECT COUNT(*) FROM returns`); err != nil {
		t.Fatal(err)
	}
	if returnCount != 0 {
		t.Fatalf("cross-sale return created %d records, want none", returnCount)
	}
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
