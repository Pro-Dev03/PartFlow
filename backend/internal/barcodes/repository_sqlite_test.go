package barcodes

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestSQLiteBarcodeAndLookup(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "barcodes.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	pid := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := database.DB.Exec(`INSERT INTO products (id,sku,name,barcode,selling_price,cost_price,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`, pid.String(), "SKU-BAR", "Barcode item", "BC-1", 25, 10, now, now); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	b := &Barcode{ID: uuid.New(), Code: "BC-1", Type: BarcodeTypeInternal, ProductID: &pid, IsActive: true, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := repo.CreateBarcode(ctx, b); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetBarcodeByCode(ctx, "BC-1")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ProductID == nil || *loaded.ProductID != pid {
		t.Fatalf("barcode mismatch: %#v", loaded)
	}
	product, err := repo.GetProductByBarcode(ctx, "BC-1")
	if err != nil {
		t.Fatal(err)
	}
	if product.ID != pid || product.Stock != 0 {
		t.Fatalf("product mismatch: %#v", product)
	}
}

func TestSQLiteResolveBarcodePreservesLifecycleReferences(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "barcode-lifecycle.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	productID, itemID, purchaseID, saleID, returnID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := database.DB.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatal(err)
	}
	statements := []struct {
		query string
		args  []interface{}
	}{
		{`INSERT INTO products (id,sku,name,barcode,selling_price,cost_price,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`, []interface{}{productID.String(), "SKU-GPU", "RTX 3060", "FNX-GPU-000421", 880, 400, now, now}},
		{`INSERT INTO inventory_items (id,product_id,item_code,barcode,serial_number,condition,purchase_cost,selling_price,status,purchase_date,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, []interface{}{itemID.String(), productID.String(), "IT-GPU-1", "FNX-GPU-000421", "SN-GPU-1", "USED", 400, 880, "SOLD", now, now, now}},
		{`INSERT INTO purchases (id,purchase_number,total_amount,status,created_at,updated_at) VALUES (?,?,?,?,?,?)`, []interface{}{purchaseID.String(), "PUR-GPU", 400, "received", now, now}},
		{`INSERT INTO purchase_items (id,purchase_id,product_id,quantity,unit_price,item_total,created_at) VALUES (?,?,?,?,?,?,?)`, []interface{}{uuid.New().String(), purchaseID.String(), productID.String(), 1, 400, 400, now}},
		{`INSERT INTO sales (id,sale_number,total_amount,status,created_at,updated_at) VALUES (?,?,?,?,?,?)`, []interface{}{saleID.String(), "SALE-GPU", 880, "completed", now, now}},
		{`INSERT INTO sale_items (id,sale_id,inventory_item_id,product_id,quantity,unit_price,item_total,created_at) VALUES (?,?,?,?,?,?,?,?)`, []interface{}{uuid.New().String(), saleID.String(), itemID.String(), productID.String(), 1, 880, 880, now}},
		{`INSERT INTO returns (id,return_number,sale_id,purchase_id,total_refund_amount,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`, []interface{}{returnID.String(), "RET-GPU", saleID.String(), purchaseID.String(), 880, "completed", now, now}},
		{`INSERT INTO return_items (id,return_id,product_id,quantity,unit_price,total_refund_amount,created_at) VALUES (?,?,?,?,?,?,?)`, []interface{}{uuid.New().String(), returnID.String(), productID.String(), 1, 880, 880, now}},
	}
	for _, statement := range statements {
		if _, err := database.DB.Exec(statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := database.DB.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatal(err)
	}
	var itemCount int
	if err := db.Get(&itemCount, `SELECT COUNT(*) FROM inventory_items WHERE barcode = ?`, "FNX-GPU-000421"); err != nil || itemCount != 1 {
		t.Fatalf("seeded inventory item count = %d, err = %v", itemCount, err)
	}

	resolution, err := NewRepository(db).ResolveBarcode(ctx, "FNX-GPU-000421")
	if err != nil {
		t.Fatal(err)
	}
	if resolution.InventoryItem == nil || resolution.InventoryItem.ID != itemID || resolution.InventoryItem.ProductID != productID {
		t.Fatalf("inventory identity mismatch: %#v", resolution.InventoryItem)
	}
	if resolution.Product == nil || resolution.Product.ID != productID {
		t.Fatalf("product identity mismatch: %#v", resolution.Product)
	}
	if len(resolution.SaleIDs) != 1 || resolution.SaleIDs[0] != saleID {
		t.Fatalf("sale references = %#v, want %s", resolution.SaleIDs, saleID)
	}
	if len(resolution.PurchaseIDs) != 1 || resolution.PurchaseIDs[0] != purchaseID {
		t.Fatalf("purchase references = %#v, want %s", resolution.PurchaseIDs, purchaseID)
	}
	if len(resolution.ReturnIDs) != 1 || resolution.ReturnIDs[0] != returnID {
		t.Fatalf("return references = %#v, want %s", resolution.ReturnIDs, returnID)
	}
}

func TestSQLiteResolveBarcodeUsesDirectLifecycleBarcodeRows(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "barcode-direct-lifecycle.db"))
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	productID, itemID, purchaseID, saleID, returnID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	code := "FNX-GPU-000421"
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := database.DB.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatal(err)
	}
	statements := []struct {
		query string
		args  []interface{}
	}{
		{`INSERT INTO products (id,sku,name,barcode,selling_price,cost_price,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`, []interface{}{productID.String(), "SKU-GPU", "RTX 3060", "", 880, 400, now, now}},
		{`INSERT INTO inventory_items (id,product_id,item_code,barcode,serial_number,condition,purchase_cost,selling_price,status,purchase_date,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, []interface{}{itemID.String(), productID.String(), "IT-GPU-1", "", "SN-GPU-1", "USED", 400, 880, "SOLD", now, now, now}},
		{`INSERT INTO purchases (id,purchase_number,total_amount,status,created_at,updated_at) VALUES (?,?,?,?,?,?)`, []interface{}{purchaseID.String(), "PUR-GPU", 400, "received", now, now}},
		{`INSERT INTO purchase_items (id,purchase_id,product_id,barcode,quantity,unit_price,item_total,created_at) VALUES (?,?,?,?,?,?,?,?)`, []interface{}{uuid.New().String(), purchaseID.String(), productID.String(), code, 1, 400, 400, now}},
		{`INSERT INTO sales (id,sale_number,total_amount,status,created_at,updated_at) VALUES (?,?,?,?,?,?)`, []interface{}{saleID.String(), "SALE-GPU", 880, "completed", now, now}},
		{`INSERT INTO sale_items (id,sale_id,inventory_item_id,product_id,quantity,unit_price,item_total,created_at) VALUES (?,?,?,?,?,?,?,?)`, []interface{}{uuid.New().String(), saleID.String(), itemID.String(), productID.String(), 1, 880, 880, now}},
		{`INSERT INTO returns (id,return_number,sale_id,purchase_id,total_refund_amount,status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`, []interface{}{returnID.String(), "RET-GPU", saleID.String(), purchaseID.String(), 880, "completed", now, now}},
		{`INSERT INTO return_items (id,return_id,product_id,inventory_item_id,serial_number,barcode,quantity_returned,original_quantity,unit_price,total_refund_amount,original_condition,returned_condition,resolution,inventory_status,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, []interface{}{uuid.New().String(), returnID.String(), productID.String(), itemID.String(), "SN-GPU-1", code, 1, 1, 880, 880, "USED", "USED", "refund", "SOLD", now, now}},
	}
	for _, statement := range statements {
		if _, err := database.DB.Exec(statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := database.DB.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatal(err)
	}

	resolution, err := NewRepository(db).ResolveBarcode(ctx, code)
	if err != nil {
		t.Fatalf("resolve direct barcode lifecycle rows: %v", err)
	}
	if resolution.Product == nil || resolution.Product.ID != productID {
		t.Fatalf("product should resolve from lifecycle barcode rows: %#v", resolution.Product)
	}
	if resolution.InventoryItem == nil || resolution.InventoryItem.ID != itemID {
		t.Fatalf("inventory item should resolve from lifecycle barcode rows: %#v", resolution.InventoryItem)
	}
	if len(resolution.PurchaseIDs) != 1 || resolution.PurchaseIDs[0] != purchaseID {
		t.Fatalf("purchase references = %#v, want %s", resolution.PurchaseIDs, purchaseID)
	}
	if len(resolution.ReturnIDs) != 1 || resolution.ReturnIDs[0] != returnID {
		t.Fatalf("return references = %#v, want %s", resolution.ReturnIDs, returnID)
	}
	if len(resolution.SaleIDs) != 1 || resolution.SaleIDs[0] != saleID {
		t.Fatalf("sale references = %#v, want %s", resolution.SaleIDs, saleID)
	}
}
