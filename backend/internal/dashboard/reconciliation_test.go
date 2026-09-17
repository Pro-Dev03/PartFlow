package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/localdb"
	"github.com/partflow/smart-store/internal/reports"
)

func TestDashboardAndReportsReconcileDailySalesAndDebtSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/dashboard-reconciliation.db")
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()

	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	now := time.Now().UTC()
	storeDate, err := accounting.StoreDate(now)
	if err != nil {
		t.Fatal(err)
	}
	customerID, saleID, productID := uuid.New(), uuid.New(), uuid.New()

	if _, err := db.Exec(`INSERT INTO customers (id, code, name, credit_limit, current_balance, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, customerID, "REC-CUSTOMER", "Reconciliation Customer", 1000, 123456.78, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, productID, "REC-PRODUCT", "Reconciliation Product", 0.12, 123456.78, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id, sale_number, customer_id, total_amount, tax_amount, paid_amount, remaining_amount, payment_method, status, created_at, updated_at, sale_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, saleID, "REC-SALE", customerID, 123456.78, 0.78, 123456.78, 0, "cash", "completed", now, now, storeDate); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, 1, ?, ?, ?)`, uuid.New(), saleID, productID, 123456.78, 123456.78, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO debts (id, customer_id, sale_id, amount, paid_amount, remaining_amount, due_date, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, uuid.New(), customerID, saleID, 123456.78, 0, 123456.78, now.AddDate(0, 0, 30), "pending", now, now); err != nil {
		t.Fatal(err)
	}

	dashboardStats, err := NewCachedService(db).fetchFromDatabaseAt(ctx, now)
	if err != nil {
		t.Fatalf("dashboard stats: %v", err)
	}
	reportRepo := reports.NewRepository(db)
	startDay, err := time.Parse("2006-01-02", storeDate)
	if err != nil {
		t.Fatal(err)
	}
	start := startDay
	end := start.Add(24 * time.Hour)
	salesReport, err := reportRepo.GetSalesData(ctx, start, end)
	if err != nil {
		t.Fatalf("sales report: %v", err)
	}
	debtReport, err := reportRepo.GetDebtsData(ctx)
	if err != nil {
		t.Fatalf("debt report: %v", err)
	}

	if len(salesReport.ByDay) != 1 || salesReport.ByDay[0].Revenue != dashboardStats.TodaySales {
		t.Fatalf("daily sales mismatch: dashboard=%v report=%v", dashboardStats.TodaySales, salesReport.ByDay)
	}
	if debtReport.Outstanding != dashboardStats.OutstandingDebts {
		t.Fatalf("outstanding debt mismatch: dashboard=%v report=%v", dashboardStats.OutstandingDebts, debtReport.Outstanding)
	}
	if dashboardStats.TodaySales != 123456 {
		t.Fatalf("today sales = %v, want 123456 after tax", dashboardStats.TodaySales)
	}
}
