package suppliers

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestDeleteSupplierHardDeletesSupplierSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "supplier-hard-delete.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	supplierID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO suppliers (id,code,name,credit_limit,current_balance,is_active,created_at,updated_at) VALUES (?,?,?,0,0,1,?,?)`, supplierID, "SUP-DELETE-01", "Delete Supplier", now, now); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db), db)
	if err := service.DeleteSupplier(context.Background(), supplierID); err != nil {
		t.Fatalf("delete supplier: %v", err)
	}
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM suppliers WHERE id=?`, supplierID); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("supplier rows after hard delete = %d, want 0", count)
	}
}

func TestDeleteSupplierReversesPurchaseStockAndHardDeletesCascadeSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "supplier-purchase-hard-delete.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	supplierID, productID, purchaseID, itemID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO suppliers (id,code,name,credit_limit,current_balance,is_active,created_at,updated_at) VALUES (?,?,?,0,10,1,?,?)`, supplierID, "SUP-PURCHASE-DELETE", "Supplier with purchase", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id,sku,name,cost_price,selling_price,is_active,created_at,updated_at) VALUES (?,?,?,10,20,1,?,?)`, productID, "PROD-SUPPLIER-DELETE", "Purchased product", now, now); err != nil {
		t.Fatal(err)
	}
	purchaseNumber := "PUR-SUPPLIER-DELETE"
	if _, err := db.Exec(`INSERT INTO purchases (id,purchase_number,invoice_number,supplier_id,total_amount,paid_amount,remaining_amount,status,created_at,updated_at) VALUES (?,?,?,?,10,0,10,'received',?,?)`, purchaseID, purchaseNumber, purchaseNumber, supplierID, now, now); err != nil {
		t.Fatal(err)
	}
	itemCode := "ITM-" + purchaseID.String()[:8] + "-001"
	if _, err := db.Exec(`INSERT INTO inventory_items (id,product_id,item_code,status,supplier_id,purchase_date,purchase_cost,created_at,updated_at) VALUES (?,?,?,'AVAILABLE',?,?,10,?,?)`, itemID, productID, itemCode, supplierID, now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,reserved_quantity,created_at,updated_at) VALUES (?,?,1,0,?,?)`, uuid.New(), productID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchase_items (id,purchase_id,product_id,quantity,unit_price,item_total,created_at) VALUES (?,?,?,1,10,10,?)`, uuid.New(), purchaseID, productID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,created_at) VALUES (?,?,?,'PURCHASE',1,0,1,'purchase',?,?)`, uuid.New(), itemID, productID, purchaseID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_ledger (id,supplier_id,type,transaction_type,amount,balance,description,reference_id,reference_type,created_at) VALUES (?,?, 'debit','PURCHASE',10,10,'purchase',?,'purchase',?)`, uuid.New(), supplierID, purchaseID, now); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db), db)
	if err := service.DeleteSupplier(context.Background(), supplierID); err != nil {
		t.Fatalf("delete supplier and purchase cascade: %v", err)
	}
	var suppliersCount, purchasesCount, itemsCount, stock, productCount int
	for _, check := range []struct {
		query string
		args  []any
		into  *int
	}{
		{`SELECT COUNT(*) FROM suppliers WHERE id=?`, []any{supplierID}, &suppliersCount},
		{`SELECT COUNT(*) FROM purchases WHERE id=?`, []any{purchaseID}, &purchasesCount},
		{`SELECT COUNT(*) FROM inventory_items WHERE id=?`, []any{itemID}, &itemsCount},
		{`SELECT quantity FROM inventory WHERE product_id=?`, []any{productID}, &stock},
		{`SELECT COUNT(*) FROM products WHERE id=?`, []any{productID}, &productCount},
	} {
		if err := db.Get(check.into, check.query, check.args...); err != nil {
			t.Fatal(err)
		}
	}
	if suppliersCount != 0 || purchasesCount != 0 || itemsCount != 0 || stock != 0 || productCount != 1 {
		t.Fatalf("cascade left supplier/purchase/items/stock/product=%d/%d/%d/%d/%d; want 0/0/0/0/1", suppliersCount, purchasesCount, itemsCount, stock, productCount)
	}
}
