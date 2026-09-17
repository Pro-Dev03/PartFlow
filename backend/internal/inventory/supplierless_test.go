package inventory

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestCreateUsedInventoryItemWithoutSupplierPreservesIdentitySQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/supplierless.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, barcode, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "USED-001", "Supplierless Used", "USED-BC-001", 50, 90); err != nil {
		t.Fatal(err)
	}

	barcode := "USED-BC-001"
	serial := "USED-SN-001"
	service := NewService(NewRepository(db), db)
	item, err := service.CreateInventoryItem(context.Background(), &InventoryItemRequest{
		ProductID:    &productID,
		Quantity:     1,
		Barcode:      &barcode,
		SerialNumber: &serial,
		Condition:    ConditionUsed,
		PurchaseCost: 50,
		SellingPrice: 90,
		Status:       StatusAvailable,
		SupplierID:   nil,
	}, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if item.ProductID == nil || *item.ProductID != productID || item.Barcode == nil || *item.Barcode != barcode || item.SerialNumber == nil || *item.SerialNumber != serial || item.SupplierID != nil || item.Condition != string(ConditionUsed) {
		t.Fatalf("supplierless identity = %+v", item)
	}

	var supplierID *string
	var storedBarcode, storedSerial, condition string
	if err := db.QueryRow(`SELECT supplier_id, barcode, serial_number, condition FROM inventory_items WHERE id = ?`, item.ID).Scan(&supplierID, &storedBarcode, &storedSerial, &condition); err != nil {
		t.Fatal(err)
	}
	if supplierID != nil || storedBarcode != barcode || storedSerial != serial || condition != string(ConditionUsed) {
		t.Fatalf("stored supplierless identity = supplier=%v barcode=%q serial=%q condition=%q", supplierID, storedBarcode, storedSerial, condition)
	}
}
