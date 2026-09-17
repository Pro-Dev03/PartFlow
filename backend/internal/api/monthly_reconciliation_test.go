package api

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
	"github.com/partflow/smart-store/internal/reports"
)

func TestMonthlyAggregationMatchesProfitReportAfterTaxAndReturnsSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/monthly-reconciliation.db")
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()

	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	month := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	start := month
	end := month.AddDate(0, 1, 0)
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	productID := uuid.New()
	saleID := uuid.New()
	saleItemID := uuid.New()
	returnID := uuid.New()
	returnItemID := uuid.New()
	expenseID := uuid.New()

	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, 'MONTHLY-001', 'Monthly Product', 10.50, 100.25, ?, ?)`, productID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id, sale_number, invoice_number, total_amount, tax_amount, cost_amount, paid_amount, remaining_amount, payment_method, payment_status, status, sale_date, created_at, updated_at) VALUES (?, 'MONTHLY-SALE', 'MONTHLY-INV', 100.25, 0.25, 10.50, 100.25, 0, 'cash', 'paid', 'completed', '2026-08-15', ?, ?)`, saleID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_price, item_total, unit_cost, tax_amount, total_amount, created_at) VALUES (?, ?, ?, 1, 100.25, 100.25, 10.50, 0.25, 100.25, ?)`, saleItemID, saleID, productID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO returns (id, return_number, sale_id, return_date, return_type, status, total_refund_amount, refund_method, created_at, updated_at) VALUES (?, 'MONTHLY-RETURN', ?, '2026-08-20', 'PARTIAL', 'COMPLETED', 20, 'CASH', ?, ?)`, returnID, saleID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO return_items (id, return_id, sale_item_id, product_id, quantity_returned, original_quantity, unit_price, total_refund_amount, original_cost, created_at, updated_at) VALUES (?, ?, ?, ?, 1, 1, 20, 20, 5, ?, ?)`, returnItemID, returnID, saleItemID, productID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO expenses (id, title, amount, expense_date, status, created_at, updated_at) VALUES (?, 'Monthly expense', 7.25, '2026-08-25', 'approved', ?, ?)`, expenseID, now, now); err != nil {
		t.Fatal(err)
	}

	h := NewAggregationHandler(db)
	if err := h.refreshSQLiteMonth(ctx, month); err != nil {
		t.Fatalf("refresh monthly summaries: %v", err)
	}

	var summary struct {
		TotalRevenue float64 `db:"total_revenue"`
		TotalCost    float64 `db:"total_cost"`
		GrossProfit  float64 `db:"gross_profit"`
		NetProfit    float64 `db:"net_profit"`
	}
	if err := db.Get(&summary, `SELECT total_revenue, total_cost, gross_profit, net_profit FROM monthly_profit_summary WHERE year = 2026 AND month = 8`); err != nil {
		t.Fatal(err)
	}
	var monthlySalesRevenue, monthlySalesProfit float64
	if err := db.QueryRow(`SELECT total_revenue, total_profit FROM monthly_sales_summary WHERE year = 2026 AND month = 8`).Scan(&monthlySalesRevenue, &monthlySalesProfit); err != nil {
		t.Fatal(err)
	}

	report, err := reports.NewRepository(db).GetProfitsData(ctx, start, end)
	if err != nil {
		t.Fatalf("profit report: %v", err)
	}
	if summary.TotalRevenue != report.TotalRevenue || summary.TotalCost != report.TotalCOGS || summary.GrossProfit != report.GrossProfit || summary.NetProfit != report.NetProfit {
		t.Fatalf("monthly profit mismatch: summary=%+v report={revenue:%v cost:%v gross:%v net:%v}", summary, report.TotalRevenue, report.TotalCOGS, report.GrossProfit, report.NetProfit)
	}
	salesReport, err := reports.NewRepository(db).GetSalesData(ctx, start, end)
	if err != nil {
		t.Fatalf("sales report: %v", err)
	}
	if monthlySalesRevenue != salesReport.TotalRevenue || monthlySalesProfit != salesReport.GrossProfit {
		t.Fatalf("monthly sales mismatch: summary={revenue:%v profit:%v} report={revenue:%v profit:%v}", monthlySalesRevenue, monthlySalesProfit, salesReport.TotalRevenue, salesReport.GrossProfit)
	}
	taxReport, err := reports.NewRepository(db).GetTaxData(ctx, start, end)
	if err != nil {
		t.Fatalf("tax report: %v", err)
	}
	if taxReport.TaxCollected != 0.25 {
		t.Fatalf("monthly tax = %v, want 0.25", taxReport.TaxCollected)
	}
}
