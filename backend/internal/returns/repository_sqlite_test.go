package returns

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
	"github.com/partflow/smart-store/internal/paymenttransactions"
)

type recordingRefundProcessor struct {
	calls  int
	fail   bool
	status string
}

func (p *recordingRefundProcessor) RefundForReturn(context.Context, uuid.UUID, uuid.UUID, int64, *uuid.UUID) (paymenttransactions.ReturnRefundResult, error) {
	p.calls++
	if p.fail {
		return paymenttransactions.ReturnRefundResult{}, fmt.Errorf("provider refund failed")
	}
	status := p.status
	if status == "" {
		status = "refunded"
	}
	return paymenttransactions.ReturnRefundResult{RefundID: uuid.New().String(), Status: status}, nil
}

func TestReturnInventoryStatusMapsAllCustomerChoices(t *testing.T) {
	returnRecord := &Return{ItemConditionAfterReturn: "READY_FOR_SALE"}
	tests := []struct {
		name       string
		condition  string
		resolution string
		want       string
	}{
		{name: "ready for sale", condition: "READY_FOR_SALE", resolution: "RESTOCK", want: "AVAILABLE"},
		{name: "not for sale", condition: "NOT_FOR_SALE", resolution: "WRITE_OFF", want: "ARCHIVED"},
		{name: "supplier return", condition: "RETURN_TO_SUPPLIER", resolution: "SUPPLIER_RETURN", want: "RETURNED"},
		{name: "needs repair", condition: "NEEDS_REPAIR", resolution: "REPAIR", want: "IN_REPAIR"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			returnRecord.ItemConditionAfterReturn = test.condition
			if got := returnInventoryStatus(returnRecord, ReturnItem{Resolution: test.resolution}); got != test.want {
				t.Fatalf("returnInventoryStatus() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSQLiteReturnRoundTrip(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "returns.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	productID, customerID, saleID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := database.DB.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, productID.String(), "SKU-RETURN", "Return product", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO customers (id,code,name,created_at,updated_at) VALUES (?,?,?,?,?)`, customerID.String(), "C-RETURN", "Return customer", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO sales (id,sale_number,total_amount,created_at,updated_at) VALUES (?,?,?,?,?)`, saleID.String(), "S-RETURN", 100, now, now); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	returnRecord := &Return{ReturnNumber: "R-RETURN", SaleID: saleID, CustomerID: customerID, ReturnDate: time.Now().UTC(), ReturnType: "FULL", Status: "PENDING", TotalRefundAmount: 100, RefundMethod: "CASH", Reason: "DEFECTIVE", ItemConditionAfterReturn: "READY_FOR_SALE"}
	if err := repo.CreateReturn(ctx, returnRecord); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetReturnByID(ctx, returnRecord.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ReturnNumber != returnRecord.ReturnNumber || loaded.CustomerID != customerID {
		t.Fatalf("loaded return mismatch: %#v", loaded)
	}
	item := &ReturnItem{ReturnID: returnRecord.ID, ProductID: &productID, QuantityReturned: 1, UnitPrice: 100, TotalRefundAmount: 100, ReturnedCondition: "DAMAGED"}
	if err := repo.CreateReturnItem(ctx, item); err != nil {
		t.Fatal(err)
	}
	items, err := repo.GetReturnItems(ctx, returnRecord.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ProductID == nil || *items[0].ProductID != productID || items[0].ProductName != "Return product" {
		t.Fatalf("items mismatch: %#v", items)
	}
	listed, count, err := repo.ListReturns(ctx, ReturnListRequest{Page: 1, PerPage: 20, Search: "R-RETURN"})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 || len(listed) != 1 {
		t.Fatalf("list count/items = %d/%d", count, len(listed))
	}
}

func TestElectronicReturnCompletesOnceAndRecordsRefundState(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "electronic-return.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	userID, saleID, productID, transactionID, saleItemID, inventoryItemID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := database.DB.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, productID.String(), "SKU-E-RETURN", "Electronic return product", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO sales (id,sale_number,invoice_number,sale_date,subtotal,discount_amount,total_amount,paid_amount,remaining_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, saleID.String(), "S-E-RETURN", "INV-E-RETURN", now, 100, 0, 100, 100, 0, "electronic", "paid", "completed", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO payment_transactions (id,sale_id,provider,status,amount_minor,currency,idempotency_key,metadata,created_at,updated_at,paid_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, transactionID.String(), saleID.String(), "cardcom", "paid", 10000, "ILS", "e-return-sale", "{}", now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO inventory_items (id,product_id,item_code,barcode,serial_number,condition,purchase_cost,selling_price,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, inventoryItemID.String(), productID.String(), "ITM-E-RETURN", "BAR-E-RETURN", "SN-E-RETURN", "NEW", 40, 100, "SOLD", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO sale_items (id,sale_id,product_id,inventory_item_id,quantity,unit_price,item_total,unit_cost,tax_amount,total_amount,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, saleItemID.String(), saleID.String(), productID.String(), inventoryItemID.String(), 1, 100, 100, 40, 0, 100, now); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db))
	processor := &recordingRefundProcessor{}
	service.SetElectronicRefundProcessor(processor)
	response, err := service.CreateReturn(ctx, userID, &ReturnRequest{
		SaleID: &saleID, ReturnDate: time.Now().UTC(), ReturnType: "FULL", Reason: "CUSTOMER_CHANGED_MIND", ItemConditionAfterReturn: "NOT_FOR_SALE", RefundMethod: "CASH",
		Items: []ReturnItemRequest{{SaleItemID: &saleItemID, ProductID: &productID, InventoryItemID: &inventoryItemID, QuantityReturned: 1, UnitPrice: 100, TotalRefundAmount: 100, ReturnedCondition: "DAMAGED", Resolution: "WRITE_OFF"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApproveReturn(ctx, response.Return.ID); err != nil {
		t.Fatal(err)
	}
	completed, err := service.CompleteReturn(ctx, response.Return.ID, userID)
	if err != nil {
		t.Fatal(err)
	}
	if processor.calls != 1 || completed.Return.Status != "COMPLETED" || completed.Return.RefundStatus != "refunded" {
		t.Fatalf("refund calls/status = %d/%s/%s", processor.calls, completed.Return.Status, completed.Return.RefundStatus)
	}
	if _, err := service.CompleteReturn(ctx, response.Return.ID, userID); err == nil {
		t.Fatal("expected duplicate completion to be rejected")
	}
	var refundStateCount int
	if err := db.GetContext(ctx, &refundStateCount, `SELECT COUNT(*) FROM return_payment_refunds WHERE return_id = ?`, response.Return.ID.String()); err != nil {
		t.Fatal(err)
	}
	if refundStateCount != 1 {
		t.Fatalf("refund state rows = %d, want 1", refundStateCount)
	}
	var inventoryStatus string
	if err := db.GetContext(ctx, &inventoryStatus, `SELECT status FROM inventory_items WHERE id = ?`, inventoryItemID.String()); err != nil {
		t.Fatal(err)
	}
	if inventoryStatus != "ARCHIVED" {
		t.Fatalf("returned inventory status = %s, want ARCHIVED", inventoryStatus)
	}
}

func TestManualReturnDoesNotCallRefundProvider(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "manual-return.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	saleID, productID := uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, _ = database.DB.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, productID.String(), "SKU-M-RETURN", "Manual return product", now, now)
	_, err = database.DB.Exec(`INSERT INTO sales (id,sale_number,invoice_number,sale_date,subtotal,discount_amount,total_amount,paid_amount,remaining_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, saleID.String(), "S-M-RETURN", "INV-M-RETURN", now, 50, 0, 50, 50, 0, "cash", "paid", "completed", now, now)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db))
	processor := &recordingRefundProcessor{}
	service.SetElectronicRefundProcessor(processor)
	response, err := service.CreateReturn(ctx, uuid.New(), &ReturnRequest{SaleID: &saleID, ReturnDate: time.Now().UTC(), ReturnType: "FULL", Reason: "CUSTOMER_CHANGED_MIND", ItemConditionAfterReturn: "NOT_FOR_SALE", RefundMethod: "CASH", Items: []ReturnItemRequest{{ProductID: &productID, QuantityReturned: 1, UnitPrice: 50, TotalRefundAmount: 50, ReturnedCondition: "DAMAGED", Resolution: "WRITE_OFF"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApproveReturn(ctx, response.Return.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CompleteReturn(ctx, response.Return.ID, uuid.Nil); err != nil {
		t.Fatal(err)
	}
	if processor.calls != 0 {
		t.Fatalf("manual return invoked provider %d times", processor.calls)
	}
}

func TestElectronicReturnRefundFailureKeepsReturnUncompleted(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "failed-electronic-return.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	saleID, productID := uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, _ = database.DB.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, productID.String(), "SKU-F-RETURN", "Failed refund product", now, now)
	_, err = database.DB.Exec(`INSERT INTO sales (id,sale_number,invoice_number,sale_date,subtotal,discount_amount,total_amount,paid_amount,remaining_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, saleID.String(), "S-F-RETURN", "INV-F-RETURN", now, 75, 0, 75, 75, 0, "electronic", "paid", "completed", now, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.DB.Exec(`INSERT INTO payment_transactions (id,sale_id,provider,status,amount_minor,currency,idempotency_key,metadata,created_at,updated_at,paid_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, uuid.New().String(), saleID.String(), "cardcom", "paid", 7500, "ILS", "failed-return-sale", "{}", now, now, now)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db))
	service.SetElectronicRefundProcessor(&recordingRefundProcessor{fail: true})
	response, err := service.CreateReturn(ctx, uuid.New(), &ReturnRequest{SaleID: &saleID, ReturnDate: time.Now().UTC(), ReturnType: "FULL", Reason: "CUSTOMER_CHANGED_MIND", ItemConditionAfterReturn: "NOT_FOR_SALE", RefundMethod: "CASH", Items: []ReturnItemRequest{{ProductID: &productID, QuantityReturned: 1, UnitPrice: 75, TotalRefundAmount: 75, ReturnedCondition: "DAMAGED", Resolution: "WRITE_OFF"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApproveReturn(ctx, response.Return.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CompleteReturn(ctx, response.Return.ID, uuid.Nil); err == nil {
		t.Fatal("expected provider refund failure")
	}
	loaded, err := service.GetReturn(ctx, response.Return.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Return.Status == "COMPLETED" || loaded.Return.RefundStatus != "failed" {
		t.Fatalf("return status/refund status = %s/%s", loaded.Return.Status, loaded.Return.RefundStatus)
	}
}

func TestElectronicPartialReturnPersistsPartialRefundState(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "partial-electronic-return.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	saleID, returnID, transactionID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, _ = database.DB.Exec(`INSERT INTO sales (id,sale_number,invoice_number,sale_date,subtotal,discount_amount,total_amount,paid_amount,remaining_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, saleID.String(), "S-P-RETURN", "INV-P-RETURN", now, 100, 0, 100, 100, 0, "electronic", "paid", "completed", now, now)
	_, err = database.DB.Exec(`INSERT INTO payment_transactions (id,sale_id,provider,status,amount_minor,currency,idempotency_key,metadata,created_at,updated_at,paid_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, transactionID.String(), saleID.String(), "cardcom", "paid", 10000, "ILS", "partial-return-sale", "{}", now, now, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.DB.Exec(`INSERT INTO returns (id,return_number,sale_id,return_date,return_type,status,total_refund_amount,refund_method,reason,item_condition_after_return,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, returnID.String(), "R-P-RETURN", saleID.String(), now, "PARTIAL", "APPROVED", 40, "CASH", "CUSTOMER_CHANGED_MIND", "WRITE_OFF", now, now)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db))
	processor := &recordingRefundProcessor{status: "partially_refunded"}
	service.SetElectronicRefundProcessor(processor)
	if err := service.processElectronicRefund(ctx, &Return{ID: returnID, SaleID: saleID, TotalRefundAmount: 40}, uuid.Nil); err != nil {
		t.Fatal(err)
	}
	if err := service.processElectronicRefund(ctx, &Return{ID: returnID, SaleID: saleID, TotalRefundAmount: 40, RefundStatus: "partially_refunded"}, uuid.Nil); err != nil {
		t.Fatal(err)
	}
	if processor.calls != 1 {
		t.Fatalf("partial refund provider calls = %d, want 1", processor.calls)
	}
	var status string
	if err := db.GetContext(ctx, &status, `SELECT status FROM return_payment_refunds WHERE return_id = ?`, returnID.String()); err != nil {
		t.Fatal(err)
	}
	if status != "partially_refunded" {
		t.Fatalf("partial refund state = %s", status)
	}
}

func TestSQLiteGetSaleInfoAllowsWalkInCustomer(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "walk-in-sale.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()

	db := sqlx.NewDb(database.DB, "sqlite")
	saleID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := database.DB.Exec(`INSERT INTO sales (id,sale_number,invoice_number,sale_date,subtotal,discount_amount,total_amount,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?)`, saleID.String(), "S-WALK-IN", "INV-WALK-IN", now, 500, 0, 500, now, now); err != nil {
		t.Fatal(err)
	}

	sale, err := NewRepository(db).GetSaleInfo(context.Background(), saleID)
	if err != nil {
		t.Fatalf("GetSaleInfo should allow a walk-in sale: %v", err)
	}
	if sale.CustomerID != uuid.Nil {
		t.Fatalf("expected no customer for walk-in sale, got %s", sale.CustomerID)
	}
}

func TestRepositoryGetReturnStatisticsWorksOnLegacySQLiteSchema(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "returns-statistics-legacy.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	ctx := context.Background()
	db := sqlx.NewDb(database.DB, "sqlite")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := database.DB.Exec(`INSERT INTO returns (id, return_number, status, reason, total_refund_amount, return_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), "RET-LEGACY-001", "COMPLETED", "DEFECTIVE", 123.45, now, now, now); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	stats, err := repo.GetReturnStatistics(ctx)
	if err != nil {
		t.Fatalf("GetReturnStatistics should work on legacy SQLite schema: %v", err)
	}
	if got, ok := stats["total_returns"].(int); !ok || got != 1 {
		t.Fatalf("expected total_returns=1, got %#v", stats["total_returns"])
	}
	if got, ok := stats["completed_returns"].(int); !ok || got != 1 {
		t.Fatalf("expected completed_returns=1, got %#v", stats["completed_returns"])
	}
	if got, ok := stats["defective_returns"].(int); !ok || got != 1 {
		t.Fatalf("expected defective_returns=1, got %#v", stats["defective_returns"])
	}
	if got, ok := stats["total_refunded"].(float64); !ok || got != 123.45 {
		t.Fatalf("expected total_refunded=123.45, got %#v", stats["total_refunded"])
	}
}

func TestServiceCreateReturnLinksActiveDebtForDebtAdjustment(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "returns-debt.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	userID, customerID, saleID, productID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	debtID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := database.DB.Exec(`INSERT INTO customers (id,code,name,created_at,updated_at) VALUES (?,?,?,?,?)`, customerID.String(), "C-DEBT", "Debt customer", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, productID.String(), "SKU-DEBT", "Debt product", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO sales (id,sale_number,invoice_number,customer_id,sale_date,total_amount,paid_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, saleID.String(), "S-DEBT", "INV-DEBT", customerID.String(), now, 500.0, 0.0, "cash", "debt", "completed", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO debts (id,customer_id,amount,paid_amount,remaining_amount,due_date,status,notes,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`, debtID.String(), customerID.String(), 500.0, 0.0, 500.0, time.Now().Add(30*24*time.Hour).Format(time.RFC3339Nano), "pending", "Sale debt", now, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db))
	resp, err := service.CreateReturn(ctx, userID, &ReturnRequest{
		SaleID:                   &saleID,
		CustomerID:               &customerID,
		ReturnDate:               time.Now().UTC(),
		ReturnType:               "PARTIAL",
		Reason:                   "DEFECTIVE",
		ItemConditionAfterReturn: "READY_FOR_SALE",
		RefundMethod:             "DEBT_ADJUSTMENT",
		Items: []ReturnItemRequest{{
			ProductID:         &productID,
			QuantityReturned:  1,
			UnitPrice:         120,
			TotalRefundAmount: 120,
			ReturnedCondition: "DAMAGED",
			Resolution:        "REPAIR",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Return.DebtID == nil || *resp.Return.DebtID != debtID {
		t.Fatalf("expected return to link the active debt, got %#v", resp.Return.DebtID)
	}
}

func TestServiceReverseReturnKeepsOriginalReturnNumber(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "reverse-return.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()

	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	userID := uuid.New()
	returnRecord := &Return{
		ID:                       uuid.New(),
		SaleID:                   uuid.New(),
		CustomerID:               uuid.New(),
		ReturnNumber:             "RET-0002",
		ReferenceNumber:          "REF-0002",
		ReturnDate:               time.Now().UTC(),
		ReturnType:               "FULL",
		Status:                   "COMPLETED",
		TotalRefundAmount:        120,
		RefundMethod:             "CASH",
		Reason:                   "CUSTOMER_CHANGED_MIND",
		ItemConditionAfterReturn: "SELLABLE",
		CreatedBy:                &userID,
		CreatedAt:                time.Now().UTC(),
		UpdatedAt:                time.Now().UTC(),
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := database.DB.Exec(`INSERT INTO customers (id,code,name,created_at,updated_at) VALUES (?,?,?,?,?)`, returnRecord.CustomerID.String(), "C-REVERSE", "Reverse customer", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO sales (id,sale_number,invoice_number,sale_date,total_amount,paid_amount,remaining_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, returnRecord.SaleID.String(), "S-REVERSE", "INV-REVERSE", now, 120, 120, 0, "cash", "paid", "completed", now, now); err != nil {
		t.Fatal(err)
	}

	if err := NewRepository(db).CreateReturn(ctx, returnRecord); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db))
	_, err = service.ReverseReturn(ctx, returnRecord.ID, userID)
	if err != nil {
		t.Fatalf("ReverseReturn() returned error: %v", err)
	}

	var count int
	if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM returns`); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected original return to remain as single row after reversal, got %d rows", count)
	}

	updated, err := service.GetReturn(ctx, returnRecord.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Return.Status != "CANCELLED" {
		t.Fatalf("expected return status CANCELLED after reverse, got %s", updated.Return.Status)
	}
	if updated.Return.ReturnNumber != "RET-0002" {
		t.Fatalf("expected original return number to be preserved, got %s", updated.Return.ReturnNumber)
	}
}

func TestSQLiteDebtAdjustmentCapsDebtAndCreatesCustomerCredit(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "returns-credit.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	userID, customerID, saleID, productID, debtID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	_, err = database.DB.Exec(`INSERT INTO customers (id,code,name,created_at,updated_at) VALUES (?,?,?,?,?)`, customerID.String(), "C-CREDIT", "Credit customer", now, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.DB.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, productID.String(), "SKU-CREDIT", "Credit product", now, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.DB.Exec(`INSERT INTO sales (id,sale_number,invoice_number,customer_id,sale_date,total_amount,paid_amount,remaining_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, saleID.String(), "S-CREDIT", "INV-CREDIT", customerID.String(), now, 110.0, 40.0, 70.0, "credit", "partial", "completed", now, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.DB.Exec(`INSERT INTO debts (id,customer_id,sale_id,amount,paid_amount,remaining_amount,due_date,status,notes,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, debtID.String(), customerID.String(), saleID.String(), 110.0, 40.0, 70.0, time.Now().Add(24*time.Hour).Format(time.RFC3339Nano), "partial", "Credit sale", now, now)
	if err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db))
	resp, err := service.CreateReturn(ctx, userID, &ReturnRequest{
		SaleID: &saleID, CustomerID: &customerID, ReturnDate: time.Now().UTC(),
		ReturnType: "PARTIAL", Reason: "CUSTOMER_CHANGED_MIND", ItemConditionAfterReturn: "READY_FOR_SALE",
		RefundMethod: "DEBT_ADJUSTMENT", DebtID: &debtID,
		Items: []ReturnItemRequest{{ProductID: &productID, QuantityReturned: 1, UnitPrice: 100, TotalRefundAmount: 100, ReturnedCondition: "DAMAGED", Resolution: "REPAIR"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApproveReturn(ctx, resp.Return.ID); err != nil {
		t.Fatal(err)
	}
	completed, err := service.CompleteReturn(ctx, resp.Return.ID, userID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Return.DebtAdjustment != 70 || completed.Return.CustomerCredit != 30 {
		t.Fatalf("return adjustment = %v, credit = %v; want 70 and 30", completed.Return.DebtAdjustment, completed.Return.CustomerCredit)
	}
	var remaining float64
	if err := db.GetContext(ctx, &remaining, `SELECT remaining_amount FROM debts WHERE id = ?`, debtID.String()); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("remaining debt = %v, want 0", remaining)
	}
	var ledgerCount int
	if err := db.GetContext(ctx, &ledgerCount, `SELECT COUNT(*) FROM customer_ledger WHERE reference_id = ?`, resp.Return.ID.String()); err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 1 {
		t.Fatalf("customer return ledger rows = %d, want 1", ledgerCount)
	}
	if _, err := service.CompleteReturn(ctx, resp.Return.ID, userID); err == nil {
		t.Fatal("expected completing the same return twice to fail")
	}
}

func TestServiceCreateSupplierReturnBridgeAcceptsNormalizedSupplierReturnCondition(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "returns-supplier-bridge-normalized.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()

	productID, supplierID, purchaseID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := database.DB.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, productID.String(), "SKU-NORMALIZED", "Normalized bridge item", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO suppliers (id, code, name, email, phone, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, supplierID.String(), "SUP-002", "Normalized Supplier", "supplier2@example.com", "0500000001", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID.String(), "PUR-002", supplierID.String(), 100.0, 0.0, "completed", now, now); err != nil {
		t.Fatal(err)
	}
	purchaseItemID := uuid.New()
	if _, err := database.DB.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, purchaseItemID.String(), purchaseID.String(), productID.String(), 1, 100.0, 100.0, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db))
	returnRecord := &Return{ReturnNumber: "RET-NORMALIZED", ItemConditionAfterReturn: "SUPPLIER_RETURN"}
	item := ReturnItem{ProductID: &productID, QuantityReturned: 1, UnitPrice: 100.0}
	if err := service.createSupplierReturnBridge(ctx, returnRecord, []ReturnItem{item}); err != nil {
		t.Fatal(err)
	}

	var supplierReturnCount int
	if err := db.GetContext(ctx, &supplierReturnCount, `SELECT COUNT(*) FROM supplier_returns WHERE status = 'NEEDS_SOURCE_DATA' AND source_status = 'NEEDS_SOURCE_DATA'`); err != nil {
		t.Fatal(err)
	}
	if supplierReturnCount != 1 {
		t.Fatalf("expected normalized supplier return condition to create an unresolved bridge row, got %d", supplierReturnCount)
	}
}

func TestServiceCompleteReturnCreatesSupplierReturnBridge(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "returns-supplier-bridge.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()

	userID, customerID, productID, supplierID, purchaseID, saleID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := database.DB.Exec(`INSERT INTO customers (id,code,name,created_at,updated_at) VALUES (?,?,?,?,?)`, customerID.String(), "C-SUP", "Supplier bridge customer", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, productID.String(), "SKU-SUP", "Supplier bridge product", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO suppliers (id, code, name, email, phone, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, supplierID.String(), "SUP-001", "Bridge Supplier", "supplier@example.com", "0500000000", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID.String(), "PUR-001", supplierID.String(), 100.0, 0.0, "completed", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, uuid.NewString(), purchaseID.String(), productID.String(), 1, 100.0, 100.0, now); err != nil {
		t.Fatal(err)
	}
	inventoryID := uuid.New()
	if _, err := database.DB.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, serial_number, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, inventoryID.String(), productID.String(), "ITM-"+purchaseID.String()[:8]+"-001", "BAR-001", "SN-001", "NEW", 100.0, 100.0, "AVAILABLE", supplierID.String(), now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO sales (id,sale_number,invoice_number,customer_id,sale_date,total_amount,paid_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, saleID.String(), "S-SUP", "INV-SUP", customerID.String(), now, 100.0, 100.0, "cash", "paid", "completed", now, now); err != nil {
		t.Fatal(err)
	}
	saleItemID := uuid.New()
	if _, err := database.DB.Exec(`INSERT INTO sale_items (id, sale_id, product_id, inventory_item_id, quantity, unit_price, item_total, unit_cost, tax_amount, total_amount, supplier_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, saleItemID.String(), saleID.String(), productID.String(), inventoryID.String(), 1, 100.0, 100.0, 100.0, 0.0, 100.0, supplierID.String(), now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db))
	resp, err := service.CreateReturn(ctx, userID, &ReturnRequest{
		SaleID:                   &saleID,
		CustomerID:               &customerID,
		ReturnDate:               time.Now().UTC(),
		ReturnType:               "PARTIAL",
		Reason:                   "DEFECTIVE",
		ItemConditionAfterReturn: "RETURN_TO_SUPPLIER",
		RefundMethod:             "CASH",
		Items: []ReturnItemRequest{{
			SaleItemID:        &saleItemID,
			ProductID:         &productID,
			QuantityReturned:  1,
			UnitPrice:         100,
			TotalRefundAmount: 100,
			ReturnedCondition: "DEFECTIVE",
			Resolution:        "SUPPLIER_RETURN",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.ApproveReturn(ctx, resp.Return.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CompleteReturn(ctx, resp.Return.ID, userID); err != nil {
		t.Fatal(err)
	}

	var supplierReturnCount int
	if err := db.GetContext(ctx, &supplierReturnCount, `SELECT COUNT(*) FROM supplier_returns WHERE purchase_id = ?`, purchaseID.String()); err != nil {
		t.Fatal(err)
	}
	if supplierReturnCount != 1 {
		t.Fatalf("expected supplier return bridge row to be created, got %d", supplierReturnCount)
	}
	var supplierReturnItemCount int
	if err := db.GetContext(ctx, &supplierReturnItemCount, `SELECT COUNT(*) FROM supplier_return_items WHERE supplier_return_id IN (SELECT id FROM supplier_returns WHERE purchase_id = ?)`, purchaseID.String()); err != nil {
		t.Fatal(err)
	}
	if supplierReturnItemCount != 1 {
		t.Fatalf("expected supplier return bridge item to be created, got %d", supplierReturnItemCount)
	}
}

func TestServiceCreateSupplierReturnBridgeStoresSourceLinksAndPreventsDuplicates(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "returns-supplier-bridge-links.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()

	userID, customerID, productID, supplierID, purchaseID, saleID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := database.DB.Exec(`INSERT INTO customers (id,code,name,created_at,updated_at) VALUES (?,?,?,?,?)`, customerID.String(), "C-LINK", "Linked customer", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO products (id,sku,name,created_at,updated_at) VALUES (?,?,?,?,?)`, productID.String(), "SKU-LINK", "Linked product", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO suppliers (id, code, name, email, phone, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, supplierID.String(), "SUP-LINK", "Linked Supplier", "supplier@example.com", "0500000000", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID.String(), "PUR-LINK", supplierID.String(), 100.0, 0.0, "completed", now, now); err != nil {
		t.Fatal(err)
	}
	purchaseItemID := uuid.New()
	if _, err := database.DB.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, purchaseItemID.String(), purchaseID.String(), productID.String(), 1, 100.0, 100.0, now); err != nil {
		t.Fatal(err)
	}
	inventoryID := uuid.New()
	if _, err := database.DB.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, serial_number, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, inventoryID.String(), productID.String(), "ITM-"+purchaseID.String()[:8]+"-001", "BAR-LINK-001", "SN-LINK-001", "NEW", 100.0, 100.0, "AVAILABLE", supplierID.String(), now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`INSERT INTO sales (id,sale_number,invoice_number,customer_id,sale_date,total_amount,paid_amount,payment_method,payment_status,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, saleID.String(), "S-LINK", "INV-LINK", customerID.String(), now, 100.0, 100.0, "cash", "paid", "completed", now, now); err != nil {
		t.Fatal(err)
	}
	saleItemID := uuid.New()
	if _, err := database.DB.Exec(`INSERT INTO sale_items (id, sale_id, product_id, inventory_item_id, quantity, unit_price, item_total, unit_cost, tax_amount, total_amount, supplier_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, saleItemID.String(), saleID.String(), productID.String(), inventoryID.String(), 1, 100.0, 100.0, 100.0, 0.0, 100.0, supplierID.String(), now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db))
	returnRecord := &Return{
		ID:                       uuid.New(),
		ReturnNumber:             "RET-LINK-001",
		SaleID:                   saleID,
		CustomerID:               customerID,
		ReturnDate:               time.Now().UTC(),
		Reason:                   "DEFECTIVE",
		ItemConditionAfterReturn: "RETURN_TO_SUPPLIER",
		CreatedBy:                &userID,
	}
	if _, err := database.DB.Exec(`INSERT INTO returns (id, return_number, customer_id, sale_id, purchase_id, total_refund_amount, refund_status, status, reason, return_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, returnRecord.ID.String(), returnRecord.ReturnNumber, customerID.String(), saleID.String(), purchaseID.String(), 100.0, "pending", "completed", returnRecord.Reason, now, now, now); err != nil {
		t.Fatal(err)
	}
	item := ReturnItem{
		ProductID:        &productID,
		SaleItemID:       &saleItemID,
		InventoryItemID:  &inventoryID,
		SerialNumber:     "SN-LINK-001",
		Barcode:          "BAR-LINK-001",
		QuantityReturned: 1,
		UnitPrice:        100,
		OriginalCost:     ptrFloat64(100),
		Resolution:       "SUPPLIER_RETURN",
	}

	if err := service.createSupplierReturnBridge(ctx, returnRecord, []ReturnItem{item}); err != nil {
		t.Fatal(err)
	}
	if err := service.createSupplierReturnBridge(ctx, returnRecord, []ReturnItem{item}); err != nil {
		t.Fatal(err)
	}

	var supplierReturnCount int
	if err := db.GetContext(ctx, &supplierReturnCount, `SELECT COUNT(*) FROM supplier_returns WHERE customer_return_id = ? AND purchase_id = ? AND supplier_id = ?`, returnRecord.ID.String(), purchaseID.String(), supplierID.String()); err != nil {
		t.Fatal(err)
	}
	if supplierReturnCount != 1 {
		t.Fatalf("expected one active supplier return request per customer return, got %d", supplierReturnCount)
	}

	var itemCount int
	if err := db.GetContext(ctx, &itemCount, `SELECT COUNT(*) FROM supplier_return_items WHERE customer_return_id = ? AND inventory_item_id = ? AND purchase_item_id = ?`, returnRecord.ID.String(), inventoryID.String(), purchaseItemID.String()); err != nil {
		t.Fatal(err)
	}
	if itemCount != 1 {
		t.Fatalf("expected one linked supplier return item for the same inventory item, got %d", itemCount)
	}

	var actualCustomerReturnID, actualPurchaseID, actualSupplierID, actualSaleID, actualInventoryItemID, actualSerial, actualBarcode, actualReturnReason string
	var actualQuantity int
	var actualPurchaseCost float64
	if err := db.GetContext(ctx, &actualCustomerReturnID, `SELECT customer_return_id FROM supplier_returns WHERE customer_return_id = ?`, returnRecord.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &actualPurchaseID, `SELECT purchase_id FROM supplier_returns WHERE customer_return_id = ?`, returnRecord.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &actualSupplierID, `SELECT supplier_id FROM supplier_returns WHERE customer_return_id = ?`, returnRecord.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &actualSaleID, `SELECT sale_id FROM supplier_returns WHERE customer_return_id = ?`, returnRecord.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &actualInventoryItemID, `SELECT inventory_item_id FROM supplier_return_items WHERE customer_return_id = ? LIMIT 1`, returnRecord.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &actualSerial, `SELECT serial_number FROM supplier_return_items WHERE customer_return_id = ? LIMIT 1`, returnRecord.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &actualBarcode, `SELECT barcode FROM supplier_return_items WHERE customer_return_id = ? LIMIT 1`, returnRecord.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &actualReturnReason, `SELECT return_reason FROM supplier_return_items WHERE customer_return_id = ? LIMIT 1`, returnRecord.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &actualQuantity, `SELECT quantity FROM supplier_return_items WHERE customer_return_id = ? LIMIT 1`, returnRecord.ID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.GetContext(ctx, &actualPurchaseCost, `SELECT purchase_cost FROM supplier_return_items WHERE customer_return_id = ? LIMIT 1`, returnRecord.ID.String()); err != nil {
		t.Fatal(err)
	}

	if actualCustomerReturnID != returnRecord.ID.String() || actualPurchaseID != purchaseID.String() || actualSupplierID != supplierID.String() || actualSaleID != saleID.String() || actualInventoryItemID != inventoryID.String() || actualSerial != "SN-LINK-001" || actualBarcode != "BAR-LINK-001" || actualReturnReason != "DEFECTIVE" || actualQuantity != 1 || actualPurchaseCost != 100.0 {
		t.Fatalf("unexpected bridge linkage: customer_return_id=%s purchase_id=%s supplier_id=%s sale_id=%s inventory_item_id=%s serial=%s barcode=%s reason=%s qty=%d purchase_cost=%v", actualCustomerReturnID, actualPurchaseID, actualSupplierID, actualSaleID, actualInventoryItemID, actualSerial, actualBarcode, actualReturnReason, actualQuantity, actualPurchaseCost)
	}
}

func ptrFloat64(v float64) *float64 { return &v }

func TestNormalizeReturnConditionUsesDatabaseValues(t *testing.T) {
	cases := map[string]string{
		"READY_FOR_SALE":     "SELLABLE",
		"NOT_FOR_SALE":       "WRITE_OFF",
		"RETURN_TO_SUPPLIER": "SUPPLIER_RETURN",
		"NEEDS_REPAIR":       "NEEDS_REPAIR",
	}

	for input, expected := range cases {
		if actual := normalizeReturnCondition(input); actual != expected {
			t.Errorf("normalizeReturnCondition(%q) = %q, want %q", input, actual, expected)
		}
	}
}
