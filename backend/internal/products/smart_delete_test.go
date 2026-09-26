package products

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestDeleteProductCascadeRemovesRemainingStockSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "product-hard-delete.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO products (id,sku,name,cost_price,selling_price,is_active,created_at,updated_at) VALUES (?,?,?,10,20,1,?,?)`, productID, "PROD-HARD-DELETE", "Delete product", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,reserved_quantity,created_at,updated_at) VALUES (?,?,7,0,?,?)`, uuid.New(), productID, now, now); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db))
	if err := service.DeleteProduct(context.Background(), productID); err != nil {
		t.Fatalf("delete product with aggregate stock: %v", err)
	}
	var productCount, inventoryCount int
	if err := db.Get(&productCount, `SELECT COUNT(*) FROM products WHERE id=?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&inventoryCount, `SELECT COUNT(*) FROM inventory WHERE product_id=?`, productID); err != nil {
		t.Fatal(err)
	}
	if productCount != 0 || inventoryCount != 0 {
		t.Fatalf("product delete left products=%d inventory=%d; want 0/0", productCount, inventoryCount)
	}
}

func TestDeleteProductCascadeReversesSaleAndDeletesInvoiceSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "product-sale-hard-delete.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID, saleID := uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO products (id,sku,name,cost_price,selling_price,is_active,created_at,updated_at) VALUES (?,?,?,10,20,1,?,?)`, productID, "PROD-SALE-DELETE", "Delete sold product", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,reserved_quantity,created_at,updated_at) VALUES (?,?,0,0,?,?)`, uuid.New(), productID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id,sale_number,invoice_number,total_amount,paid_amount,remaining_amount,payment_method,status,created_at,updated_at) VALUES (?,?,?,20,20,0,'cash','completed',?,?)`, saleID, "SALE-PROD-DELETE", "SALE-PROD-DELETE", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id,sale_id,product_id,quantity,unit_price,item_total,created_at) VALUES (?,?,?,1,20,20,?)`, uuid.New(), saleID, productID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payments (id,transaction_number,customer_id,amount,payment_method,sale_id,created_at) VALUES (?,?,NULL,20,'cash',?,?)`, uuid.New(), "PAY-PROD-DELETE", saleID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,reason,created_at) VALUES (?,?, 'SALE',-1,1,0,'sale',?,'sale test',?)`, uuid.New(), productID, saleID, now); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db))
	if err := service.DeleteProduct(context.Background(), productID); err != nil {
		t.Fatalf("delete product with sale history: %v", err)
	}
	var productCount, saleCount, saleItemCount, movementCount int
	for _, check := range []struct {
		query string
		args  []any
		into  *int
	}{
		{`SELECT COUNT(*) FROM products WHERE id=?`, []any{productID}, &productCount},
		{`SELECT COUNT(*) FROM sales WHERE id=?`, []any{saleID}, &saleCount},
		{`SELECT COUNT(*) FROM sale_items WHERE sale_id=?`, []any{saleID}, &saleItemCount},
		{`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=?`, []any{saleID}, &movementCount},
	} {
		if err := db.Get(check.into, check.query, check.args...); err != nil {
			t.Fatal(err)
		}
	}
	if productCount != 0 || saleCount != 0 || saleItemCount != 0 || movementCount != 0 {
		t.Fatalf("product cascade left product/sale/lines/movements=%d/%d/%d/%d; want all zero", productCount, saleCount, saleItemCount, movementCount)
	}
}

func TestDeleteProductCascadeReversesAcquisitionStockSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "product-acquisition-hard-delete.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID, supplierID, acquisitionID, inventoryItemID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO suppliers (id,code,name,credit_limit,current_balance,is_active,created_at,updated_at) VALUES (?,?,?,0,0,1,?,?)`, supplierID, "SUP-PROD-ACQ", "Acquisition supplier", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id,sku,name,cost_price,selling_price,is_active,created_at,updated_at) VALUES (?,?,?,5,10,1,?,?)`, productID, "PROD-ACQ-DELETE", "Acquired product", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES (?,?,1,?,?)`, uuid.New(), productID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id,product_id,item_code,status,purchase_cost,selling_price,created_at,updated_at) VALUES (?,?,?,'AVAILABLE',5,10,?,?)`, inventoryItemID, productID, "ACQ-PROD-001", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO acquisitions (id,type,acquisition_date,supplier_id,total_cost,paid_amount,payment_status,status,created_at,updated_at) VALUES (?,'SUPPLIER',?,?,5,0,'payable','acquired',?,?)`, acquisitionID, now, supplierID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO acquisition_items (id,acquisition_id,product_id,inventory_item_id,unit_cost,total_cost,item_status,created_at,updated_at) VALUES (?,?,?,?,5,5,'available',?,?)`, uuid.New(), acquisitionID, productID, inventoryItemID, now, now); err != nil {
		t.Fatal(err)
	}
	if err := NewService(NewRepository(db)).DeleteProduct(context.Background(), productID); err != nil {
		t.Fatalf("delete product with acquisition history: %v", err)
	}
	var productCount, acquisitionCount, itemCount, stock int
	for _, check := range []struct {
		query string
		args  []any
		into  *int
	}{
		{`SELECT COUNT(*) FROM products WHERE id=?`, []any{productID}, &productCount},
		{`SELECT COUNT(*) FROM acquisitions WHERE id=?`, []any{acquisitionID}, &acquisitionCount},
		{`SELECT COUNT(*) FROM inventory_items WHERE id=?`, []any{inventoryItemID}, &itemCount},
		{`SELECT COALESCE((SELECT quantity FROM inventory WHERE product_id=?),0)`, []any{productID}, &stock},
	} {
		if err := db.Get(check.into, check.query, check.args...); err != nil {
			t.Fatal(err)
		}
	}
	if productCount != 0 || acquisitionCount != 0 || itemCount != 0 || stock != 0 {
		t.Fatalf("product acquisition delete left product/acquisition/item/stock=%d/%d/%d/%d; want all zero", productCount, acquisitionCount, itemCount, stock)
	}
}
