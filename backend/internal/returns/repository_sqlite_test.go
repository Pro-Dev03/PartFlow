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
	returnRecord := &Return{ReturnNumber: "R-RETURN", SaleID: saleID, CustomerID: customerID, ReturnDate: time.Now().UTC(), ReturnType: "FULL", Status: "PENDING", TotalRefundAmount: 100, RefundMethod: "CASH", Reason: "DEFECTIVE", ItemConditionAfterReturn: "NEEDS_INSPECTION"}
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
