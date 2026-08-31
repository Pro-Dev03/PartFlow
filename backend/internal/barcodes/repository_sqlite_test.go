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
