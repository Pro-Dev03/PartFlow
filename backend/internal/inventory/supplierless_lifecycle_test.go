package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
	returnservice "github.com/partflow/smart-store/internal/returns"
	salesservice "github.com/partflow/smart-store/internal/sales"
)

func TestSupplierlessOpeningStockSaleAndCustomerReturnPreserveIdentitySQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/supplierless-lifecycle.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	ctx := context.Background()
	productID := uuid.New()
	barcode := "LIFE-BC-001"
	serial := "LIFE-SN-001"
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, productID, "LIFE-001", "Supplierless Lifecycle", 12, 30, now, now); err != nil {
		t.Fatal(err)
	}

	inventoryService := NewService(NewRepository(db), db)
	opening, err := inventoryService.CreateOpeningStock(ctx, &OpeningStockRequest{
		ProductID: &productID, Mode: OpeningStockModeIndividual, Quantity: 1, BusinessDate: "2026-03-01",
		Barcode: &barcode, SerialNumber: &serial, Condition: ConditionUsed, PurchaseCost: 12, SellingPrice: 30,
	}, uuid.New())
	if err != nil {
		t.Fatalf("opening stock: %v", err)
	}
	if opening.Item == nil {
		t.Fatal("opening stock did not return an individual item")
	}
	itemID := opening.Item.ID
	assertSupplierlessIdentity(t, db, itemID, productID, barcode, serial, "AVAILABLE")
	assertInventoryQuantity(t, db, productID, 1)

	saleService := salesservice.NewService(salesservice.NewRepository(db), db)
	sale, err := saleService.CreateSale(ctx, uuid.Nil, &salesservice.CreateSaleRequest{
		Items:         []salesservice.SaleItemRequest{{ProductID: productID, InventoryItemID: &itemID, Quantity: 1, UnitPrice: 30}},
		PaymentAmount: 30,
	})
	if err != nil {
		t.Fatalf("sale: %v", err)
	}
	var saleItemID uuid.UUID
	var saleInventoryItemID uuid.UUID
	var saleSupplierID *string
	if err := db.QueryRow(`SELECT id, inventory_item_id, supplier_id FROM sale_items WHERE sale_id = ?`, sale.ID).Scan(&saleItemID, &saleInventoryItemID, &saleSupplierID); err != nil {
		t.Fatal(err)
	}
	if saleInventoryItemID != itemID || saleSupplierID != nil {
		t.Fatalf("sale linkage = item=%s supplier=%v, want item=%s supplier=NULL", saleInventoryItemID, saleSupplierID, itemID)
	}
	assertSupplierlessIdentity(t, db, itemID, productID, barcode, serial, "SOLD")
	assertInventoryQuantity(t, db, productID, 0)

	returnServiceInstance := returnservice.NewService(returnservice.NewRepository(db))
	returned, err := returnServiceInstance.CreateReturn(ctx, uuid.Nil, &returnservice.ReturnRequest{
		SaleID: &sale.ID, ReturnDate: time.Now().UTC(), ReturnType: "PARTIAL", Reason: "CUSTOMER_CHANGED_MIND",
		ItemConditionAfterReturn: "READY_FOR_SALE", RefundMethod: "CASH",
		Items: []returnservice.ReturnItemRequest{{SaleItemID: &saleItemID, InventoryItemID: &itemID, QuantityReturned: 1, UnitPrice: 30, TotalRefundAmount: 30, ReturnedCondition: "USED", Resolution: "RESTOCK"}},
	})
	if err != nil {
		t.Fatalf("create return: %v", err)
	}
	if _, err := returnServiceInstance.ApproveReturn(ctx, returned.Return.ID); err != nil {
		t.Fatalf("approve return: %v", err)
	}
	if _, err := returnServiceInstance.CompleteReturn(ctx, returned.Return.ID, uuid.Nil); err != nil {
		t.Fatalf("complete return: %v", err)
	}

	assertSupplierlessIdentity(t, db, itemID, productID, barcode, serial, "AVAILABLE")
	assertInventoryQuantity(t, db, productID, 1)
}

func assertSupplierlessIdentity(t *testing.T, db *sqlx.DB, itemID, productID uuid.UUID, barcode, serial, status string) {
	t.Helper()
	var storedProductID, storedBarcode, storedSerial, storedStatus string
	var supplierID *string
	if err := db.QueryRow(`SELECT product_id, barcode, serial_number, status, supplier_id FROM inventory_items WHERE id = ?`, itemID).Scan(&storedProductID, &storedBarcode, &storedSerial, &storedStatus, &supplierID); err != nil {
		t.Fatal(err)
	}
	if storedProductID != productID.String() || storedBarcode != barcode || storedSerial != serial || storedStatus != status || supplierID != nil {
		t.Fatalf("identity = product=%s barcode=%q serial=%q status=%q supplier=%v", storedProductID, storedBarcode, storedSerial, storedStatus, supplierID)
	}
}

func assertInventoryQuantity(t *testing.T, db *sqlx.DB, productID uuid.UUID, want int) {
	t.Helper()
	var quantity int
	if err := db.Get(&quantity, `SELECT quantity FROM inventory WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if quantity != want {
		t.Fatalf("inventory quantity = %d, want %d", quantity, want)
	}
}
