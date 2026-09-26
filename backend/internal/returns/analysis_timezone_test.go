package returns

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	_ "modernc.org/sqlite"
)

func TestReturnAnalysisUsesStoreLocalMonthWithoutMultiplyingSales(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}

	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY,
			sale_date TEXT,
			created_at TEXT,
			status TEXT,
			total_amount REAL,
			tax_amount REAL,
			cost_amount REAL,
			gross_profit REAL
		);
		CREATE TABLE accounting_returns (
			id TEXT PRIMARY KEY,
			sale_id TEXT,
			customer_id TEXT,
			return_date TEXT,
			refund_date TEXT,
			created_at TEXT,
			status TEXT,
			total_refund_amount REAL,
			return_type TEXT,
			reason TEXT,
			is_warranty_claim INTEGER,
			item_condition_after_return TEXT,
			reference_number TEXT
		);
		INSERT INTO sales VALUES
			('sale-aug-local', '2026-09-01T02:00:00Z', '2026-09-01T02:00:00Z', 'COMPLETED', 100, 0, 20, 80),
			('sale-sep-local', '2026-09-01T05:00:00Z', '2026-09-01T05:00:00Z', 'COMPLETED', 50, 0, 10, 40);
		INSERT INTO accounting_returns VALUES
			('return-1', 'sale-sep-local', 'customer-1', '2026-09-01T03:30:00Z', NULL, '2026-09-01T03:30:00Z', 'COMPLETED', 5, 'PARTIAL', 'DAMAGED', 0, 'SELLABLE', 'REF-1'),
			('return-2', 'sale-sep-local', 'customer-1', '2026-09-01T03:45:00Z', NULL, '2026-09-01T03:45:00Z', 'COMPLETED', 7, 'PARTIAL', 'DAMAGED', 0, 'SELLABLE', 'REF-2');
	`)
	if err != nil {
		t.Fatalf("create return analysis fixture: %v", err)
	}

	repo := NewRepository(db)
	monthly, err := repo.GetMonthlyReturnsAnalysis(context.Background())
	if err != nil {
		t.Fatalf("monthly returns analysis: %v", err)
	}
	if len(monthly) != 1 || monthly[0].Month.Format("2006-01-02 15:04:05 -07:00") != "2026-08-01 00:00:00 -04:00" {
		t.Fatalf("monthly return date should be the store-local month start: %+v", monthly)
	}
	if monthly[0].TotalReturns != 2 || monthly[0].UniqueCustomers != 1 || monthly[0].TotalRefundAmount != 12 || monthly[0].AvgRefundAmount != 6 {
		t.Fatalf("monthly return totals mismatch: %+v", monthly[0])
	}

	comparison, err := repo.GetSalesReturnsAnalysis(context.Background())
	if err != nil {
		t.Fatalf("sales and returns analysis: %v", err)
	}
	if len(comparison) != 2 {
		t.Fatalf("sales and returns analysis returned %d months, want 2: %+v", len(comparison), comparison)
	}
	byMonth := map[string]SalesReturnsAnalysis{}
	for _, month := range comparison {
		byMonth[month.Month.Format("2006-01")] = month
	}
	august := byMonth["2026-08"]
	if august.TotalSales != 1 || august.GrossSales != 100 || august.ReturnsAmount != 12 || august.ReturnCount != 2 || august.NetSales != 88 {
		t.Fatalf("August totals should use the store calendar and count both returns once: %+v", august)
	}
	september := byMonth["2026-09"]
	if september.TotalSales != 1 || september.GrossSales != 50 || september.ReturnsAmount != 0 || september.ReturnCount != 0 || september.NetSales != 50 {
		t.Fatalf("September sales were multiplied by returns recorded in August: %+v", september)
	}

	if err := accounting.ConfigureStoreTimezone("Asia/Tokyo"); err != nil {
		t.Fatal(err)
	}
	monthly, err = repo.GetMonthlyReturnsAnalysis(context.Background())
	if err != nil {
		t.Fatalf("monthly returns analysis after changing timezone: %v", err)
	}
	if len(monthly) != 1 || monthly[0].Month.Format("2006-01-02 15:04:05 -07:00") != "2026-09-01 00:00:00 +09:00" {
		t.Fatalf("timezone change was not reflected in monthly report: %+v", monthly)
	}
	comparison, err = repo.GetSalesReturnsAnalysis(context.Background())
	if err != nil {
		t.Fatalf("sales and returns analysis after changing timezone: %v", err)
	}
	if len(comparison) != 1 || comparison[0].TotalSales != 2 || comparison[0].GrossSales != 150 || comparison[0].ReturnsAmount != 12 || comparison[0].ReturnCount != 2 || comparison[0].NetSales != 138 {
		t.Fatalf("timezone change was not reflected in sales/returns analysis: %+v", comparison)
	}
}
