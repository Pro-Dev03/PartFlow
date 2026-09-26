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
	nowTime := time.Now().UTC()
	now := nowTime.Format(time.RFC3339Nano)
	oldStockDate := nowTime.AddDate(0, 0, -45).Format(time.RFC3339Nano)
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
	if _, err := db.Exec(`UPDATE products SET min_stock_level = 7 WHERE id = ?`, bulkID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, created_at, updated_at)
		VALUES (?, ?, 4, ?, ?)`, uuid.New(), bulkID, oldStockDate, oldStockDate); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, location, created_at, updated_at)
		VALUES (?, ?, 2, 'SHELF-B', ?, ?)`, uuid.New(), bulkID, oldStockDate, oldStockDate); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id, product_id, movement_type, quantity, before_quantity, after_quantity, created_at)
		VALUES (?, ?, 'ADJUSTMENT', 4, 0, 4, ?)`, uuid.New(), bulkID, oldStockDate); err != nil {
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
	saleID := uuid.New()
	if _, err := db.Exec(`INSERT INTO sales (id, sale_number, total_amount, sale_date, status, created_at, updated_at)
		VALUES (?, ?, 30, ?, 'completed', ?, ?)`, saleID, "REPORT-SALE-1", nowTime.Format("2006-01-02"), now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_price, item_total, created_at)
		VALUES (?, ?, ?, 1, 30, 30, ?)`, uuid.New(), saleID, serialID, now); err != nil {
		t.Fatal(err)
	}

	report, err := NewRepository(db).GetInventoryData(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if report.TotalItems != 9 || report.TotalValue != 130 {
		t.Fatalf("inventory totals = (%d, %.2f), want (9, 130)", report.TotalItems, report.TotalValue)
	}
	if report.ByCondition["NEW"] != 8 || report.ByCondition["REFURBISHED"] != 1 {
		t.Fatalf("condition counts = %#v", report.ByCondition)
	}
	if report.Valuation.ByCondition["NEW"] != 110 || report.Valuation.ByCondition["REFURBISHED"] != 20 {
		t.Fatalf("condition values = %#v", report.Valuation.ByCondition)
	}
	categoryTotal := 0
	for _, count := range report.ByCategory {
		categoryTotal += count
	}
	if categoryTotal != report.TotalItems {
		t.Fatalf("category stock totals = %d, want total available stock %d (categories %#v)", categoryTotal, report.TotalItems, report.ByCategory)
	}
	bulkInventoryRows := 0
	for _, item := range report.Items {
		if item.ProductID == bulkID {
			bulkInventoryRows++
			if item.CurrentStock != 6 {
				t.Errorf("bulk inventory row stock = %d, want 6", item.CurrentStock)
			}
		}
	}
	if bulkInventoryRows != 1 {
		t.Errorf("bulk product rows = %d, want exactly one product row", bulkInventoryRows)
	}
	bulkLowStockRows := 0
	for _, item := range report.LowStockItems {
		if item.ProductID == bulkID {
			bulkLowStockRows++
			if item.CurrentStock != 6 {
				t.Errorf("bulk low stock = %d, want summed stock 6", item.CurrentStock)
			}
		}
	}
	if bulkLowStockRows != 1 {
		t.Errorf("bulk low-stock rows = %d, want exactly one row", bulkLowStockRows)
	}

	var bulkOverstock *OverstockItem
	for index := range report.OverstockItems {
		if report.OverstockItems[index].ProductID == bulkID {
			bulkOverstock = &report.OverstockItems[index]
			break
		}
	}
	if bulkOverstock == nil || bulkOverstock.CurrentStock != 6 {
		t.Fatalf("aggregate-only stock missing from overstock report: %#v", report.OverstockItems)
	}
	if bulkOverstock.AvgMonthlySales != 0 || bulkOverstock.MonthsOfSupply != nil {
		t.Fatalf("no-sale overstock metrics = (%.2f, %v), want zero monthly sales and undefined supply", bulkOverstock.AvgMonthlySales, bulkOverstock.MonthsOfSupply)
	}
	var serialOverstock *OverstockItem
	for index := range report.OverstockItems {
		if report.OverstockItems[index].ProductID == serialID {
			serialOverstock = &report.OverstockItems[index]
			break
		}
	}
	if serialOverstock == nil {
		t.Fatalf("slow-selling product missing from overstock report: %#v", report.OverstockItems)
	}
	if serialOverstock.AvgMonthlySales < 0.33 || serialOverstock.AvgMonthlySales > 0.34 || serialOverstock.MonthsOfSupply == nil || *serialOverstock.MonthsOfSupply < 5.99 || *serialOverstock.MonthsOfSupply > 6.01 {
		t.Fatalf("slow-selling metrics = (%.3f, %v), want average 0.333 and supply 6 months", serialOverstock.AvgMonthlySales, serialOverstock.MonthsOfSupply)
	}

	var bulkStagnant *StagnantItem
	for index := range report.StagnantItems {
		if report.StagnantItems[index].ProductID == bulkID {
			bulkStagnant = &report.StagnantItems[index]
			break
		}
	}
	if bulkStagnant == nil {
		t.Fatalf("aggregate-only stock missing from stagnant report: %#v", report.StagnantItems)
	}
	if bulkStagnant.CurrentStock != 6 || bulkStagnant.Value != 60 || bulkStagnant.DaysSinceSale < 30 {
		t.Fatalf("aggregate stagnant item = %#v, want stock 6, value 60, and at least 30 days old", bulkStagnant)
	}
}
