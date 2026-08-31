package acquisitions

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestCreateAcquisitionCreatesLinkedLocalInventory(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/acquisition.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID, customerID, userID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := local.DB.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, 10, 20, ?, ?)`, productID, "SKU-ACQ", "Used part", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := local.DB.Exec(`INSERT INTO customers (id, code, name, is_active, created_at, updated_at) VALUES (?, ?, ?, 1, ?, ?)`, customerID, "C-ACQ", "Seller", now, now); err != nil {
		t.Fatal(err)
	}
	service := NewService(db)
	productUUID := uuid.MustParse(productID)
	customerUUID := uuid.MustParse(customerID)
	result, err := service.CreateAcquisition(context.Background(), &AcquisitionRequest{
		Type: TypeCustomer, AcquisitionDate: time.Now().UTC(), CustomerID: &customerUUID,
		Items: []AcquisitionItemRequest{{ProductID: productUUID, Condition: "used", Grade: "good", UnitCost: 10}},
	}, uuid.MustParse(userID))
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalCost != 10 {
		t.Fatalf("total cost = %v, want 10", result.TotalCost)
	}
	var count int
	if err := local.DB.QueryRow(`SELECT COUNT(*) FROM inventory_items WHERE product_id = ?`, productID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("inventory item count = %d, want 1", count)
	}
	var linked, aggregate int
	if err := local.DB.QueryRow(`SELECT COUNT(*) FROM acquisition_items WHERE acquisition_id = ? AND inventory_item_id IS NOT NULL`, result.ID.String()).Scan(&linked); err != nil {
		t.Fatal(err)
	}
	if linked != 1 {
		t.Fatalf("linked acquisition items = %d, want 1", linked)
	}
	if err := local.DB.QueryRow(`SELECT quantity FROM inventory WHERE product_id = ?`, productID).Scan(&aggregate); err != nil {
		t.Fatal(err)
	}
	if aggregate != 1 {
		t.Fatalf("aggregate quantity = %d, want 1", aggregate)
	}
}
