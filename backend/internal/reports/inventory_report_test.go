package reports

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestInventoryReportConditionsMatchAvailableStockSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/inventory-report.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	bulkID, serialID, legacyID := uuid.New(), uuid.New(), uuid.New()

	for _, product := range []struct {
		id    uuid.UUID
		sku   string
		cost  float64
		price float64
	}{
		{bulkID, "REPORT-BULK", 10, 20},
		{serialID, "REPORT-SERIAL", 20, 30},
		{legacyID, "REPORT-LEGACY", 30, 40},
	} {
		if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`, product.id, product.sku, product.sku, product.cost, product.price, now, now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, created_at, updated_at)
		VALUES (?, ?, 4, ?, ?)`, uuid.New(), bulkID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, created_at, updated_at)
		VALUES (?, ?, 2, ?, ?)`, uuid.New(), serialID, now, now); err != nil {
		t.Fatal(err)
	}
	for index, condition := range []string{"NEW", "REFURBISHED"} {
		if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, condition, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, 'AVAILABLE', ?, ?)`, uuid.New(), serialID, "REPORT-SERIAL-"+string(rune('1'+index)), condition, now, now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, condition, status, created_at, updated_at)
		VALUES (?, ?, 'REPORT-LEGACY-1', 'NEW', 'AVAILABLE', ?, ?)`, uuid.New(), legacyID, now, now); err != nil {
		t.Fatal(err)
	}

	report, err := NewRepository(db).GetInventoryData(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if report.TotalItems != 7 || report.TotalValue != 110 {
		t.Fatalf("inventory totals = (%d, %.2f), want (7, 110)", report.TotalItems, report.TotalValue)
	}
	if report.ByCondition["NEW"] != 6 || report.ByCondition["REFURBISHED"] != 1 {
		t.Fatalf("condition counts = %#v", report.ByCondition)
	}
	if report.Valuation.ByCondition["NEW"] != 90 || report.Valuation.ByCondition["REFURBISHED"] != 20 {
		t.Fatalf("condition values = %#v", report.Valuation.ByCondition)
	}
	categoryTotal := 0
	for _, count := range report.ByCategory {
		categoryTotal += count
	}
	if categoryTotal != report.TotalItems {
		t.Fatalf("category stock totals = %d, want total available stock %d (categories %#v)", categoryTotal, report.TotalItems, report.ByCategory)
	}
}
