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
	if len(items) != 1 || items[0].ProductID == nil || *items[0].ProductID != productID {
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
