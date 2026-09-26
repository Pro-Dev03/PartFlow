package reports

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestFinancialReportsReconcileMultiLineTaxedReturn(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY, sale_date TEXT, created_at TEXT, total_amount REAL,
			tax_amount REAL, subtotal REAL, discount_amount REAL, payment_method TEXT,
			status TEXT, cost_amount REAL, paid_amount REAL, cash_received REAL, change_amount REAL
		);
		CREATE TABLE sale_items (
			id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, quantity INTEGER,
			unit_cost REAL, total_amount REAL, tax_amount REAL
		);
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, category_id TEXT, cost_price REAL, is_active INTEGER);
		CREATE TABLE categories (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, expense_date TEXT, status TEXT);
		CREATE TABLE accounting_returns (
			id TEXT PRIMARY KEY, sale_id TEXT, refund_date TEXT, return_date TEXT, created_at TEXT,
			status TEXT, total_refund_amount REAL, reference_number TEXT, reason TEXT
		);
		CREATE TABLE accounting_return_items (
			id TEXT PRIMARY KEY, return_id TEXT, product_id TEXT, sale_item_id TEXT, inventory_item_id TEXT,
			quantity_returned INTEGER, total_refund_amount REAL, original_cost REAL
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		INSERT INTO categories (id, name) VALUES ('cat-1', 'Parts A'), ('cat-2', 'Parts B');
		INSERT INTO products (id, name, category_id, cost_price, is_active) VALUES
			('00000000-0000-0000-0000-000000000011', 'Part A', 'cat-1', 30, 1),
			('00000000-0000-0000-0000-000000000012', 'Part B', 'cat-2', 100, 1);
		INSERT INTO sales (id, sale_date, created_at, total_amount, tax_amount, subtotal, discount_amount,
			payment_method, status, cost_amount, paid_amount, cash_received, change_amount)
		VALUES ('sale-1', '2026-09-16T10:00:00Z', '2026-09-16T10:00:00Z', 330, 30, 300, 0,
			'cash', 'completed', 160, 330, 330, 0);
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_cost, total_amount, tax_amount) VALUES
			('line-1', 'sale-1', '00000000-0000-0000-0000-000000000011', 2, 30, 110, 10),
			('line-2', 'sale-1', '00000000-0000-0000-0000-000000000012', 1, 100, 220, 20);
		INSERT INTO accounting_returns (id, sale_id, refund_date, return_date, created_at, status, total_refund_amount, reference_number, reason)
		VALUES ('return-1', 'sale-1', '2026-09-16T14:00:00Z', '2026-09-16T14:00:00Z', '2026-09-16T14:00:00Z', 'COMPLETED', 165, 'RET-1', 'damaged');
		INSERT INTO accounting_return_items (id, return_id, product_id, sale_item_id, inventory_item_id, quantity_returned, total_refund_amount, original_cost) VALUES
			('return-line-1', 'return-1', '00000000-0000-0000-0000-000000000011', 'line-1', NULL, 1, 55, 30),
			('return-line-2', 'return-1', '00000000-0000-0000-0000-000000000012', 'line-2', NULL, 1, 110, 100);
	`)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	repo := NewRepository(db)

	sales, err := repo.GetSalesData(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetSalesData: %v", err)
	}
	if !closeReportAmount(sales.TotalRevenue, 150) || !closeReportAmount(sales.TotalTax, 15) || !closeReportAmount(sales.TotalCOGS, 30) || !closeReportAmount(sales.GrossProfit, 120) || sales.TotalItemsSold != 1 {
		t.Fatalf("sales return adjustment did not reconcile: revenue %.2f tax %.2f cogs %.2f gross %.2f items %d", sales.TotalRevenue, sales.TotalTax, sales.TotalCOGS, sales.GrossProfit, sales.TotalItemsSold)
	}
	if !closeReportAmount(sales.ByPaymentMethod["cash"], 150) {
		t.Fatalf("sales payment breakdown = %#v, want tax-exclusive net cash revenue 150", sales.ByPaymentMethod)
	}
	salesProducts := make(map[string]ProductSales, len(sales.TopProducts))
	for _, product := range sales.TopProducts {
		salesProducts[product.ProductID.String()] = product
	}
	productA := salesProducts["00000000-0000-0000-0000-000000000011"]
	productB := salesProducts["00000000-0000-0000-0000-000000000012"]
	if productA.Quantity != 1 || !closeReportAmount(productA.Revenue, 50) || productB.Quantity != 0 || !closeReportAmount(productB.Revenue, 100) {
		t.Fatalf("top products must deduct returns excluding returned tax: A=%#v B=%#v", productA, productB)
	}

	profits, err := repo.GetProfitsData(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetProfitsData: %v", err)
	}
	if !closeReportAmount(profits.TotalRevenue, sales.TotalRevenue) || !closeReportAmount(profits.TotalCOGS, sales.TotalCOGS) || !closeReportAmount(profits.GrossProfit, sales.GrossProfit) {
		t.Fatalf("sales/profit totals disagree: sales=(%.2f, %.2f, %.2f), profit=(%.2f, %.2f, %.2f)", sales.TotalRevenue, sales.TotalCOGS, sales.GrossProfit, profits.TotalRevenue, profits.TotalCOGS, profits.GrossProfit)
	}
	if len(profits.ByDay) != 1 || !closeReportAmount(profits.ByDay[0].Revenue, profits.TotalRevenue) || !closeReportAmount(profits.ByDay[0].COGS, profits.TotalCOGS) {
		t.Fatalf("daily profit does not reconcile: %#v", profits.ByDay)
	}
	if len(profits.ByMonth) != 1 || !closeReportAmount(profits.ByMonth[0].Revenue, profits.TotalRevenue) || !closeReportAmount(profits.ByMonth[0].COGS, profits.TotalCOGS) {
		t.Fatalf("monthly profit does not reconcile: %#v", profits.ByMonth)
	}
	categoryProfit := 0.0
	for _, amount := range profits.ByCategory {
		categoryProfit += amount
	}
	if !closeReportAmount(categoryProfit, profits.GrossProfit) {
		t.Fatalf("category gross profit = %.2f, want report gross profit %.2f", categoryProfit, profits.GrossProfit)
	}
	if !closeReportAmount(profits.ByCategory["Parts A"], 20) || !closeReportAmount(profits.ByCategory["Parts B"], 100) {
		adjustments, adjustmentErr := repo.returnCategoryAdjustments(context.Background(), start, end)
		t.Fatalf("category return allocation is incorrect: %#v, adjustments=%#v, err=%v; want 20 and 100", profits.ByCategory, adjustments, adjustmentErr)
	}

	netSales, err := repo.GetNetSalesData(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetNetSalesData: %v", err)
	}
	if !closeReportAmount(netSales.GrossRevenue, 300) || !closeReportAmount(netSales.TotalRefunded, 165) || !closeReportAmount(netSales.NetRevenue, 150) {
		t.Fatalf("net sales amounts disagree with sales/profit reports: gross %.2f refunded %.2f net %.2f", netSales.GrossRevenue, netSales.TotalRefunded, netSales.NetRevenue)
	}
	netProducts := make(map[string]ProductNetSales, len(netSales.TopReturnedProducts))
	for _, product := range netSales.TopReturnedProducts {
		netProducts[product.ProductID.String()] = product
	}
	if !closeReportAmount(netProducts["00000000-0000-0000-0000-000000000011"].NetRevenue, 50) || !closeReportAmount(netProducts["00000000-0000-0000-0000-000000000012"].NetRevenue, 100) {
		t.Fatalf("top returned product revenue must match tax-exclusive net sales: %#v", netProducts)
	}
	if len(netSales.ByDay) != 1 || !closeReportAmount(netSales.ByDay[0].NetRevenue, netSales.NetRevenue) {
		t.Fatalf("daily net sales does not reconcile: %#v", netSales.ByDay)
	}

	returns, err := repo.GetReturnsData(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetReturnsData: %v", err)
	}
	if returns.TotalReturns != 1 || !closeReportAmount(returns.TotalRefunded, 165) || len(returns.ByMonth) != 1 || !closeReportAmount(returns.ByMonth[0].Amount, returns.TotalRefunded) {
		t.Fatalf("returns summary does not reconcile: %#v", returns)
	}

	tax, err := repo.GetTaxData(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetTaxData: %v", err)
	}
	if !closeReportAmount(tax.ReturnsTotal, 165) || !closeReportAmount(tax.ReturnedTax, 15) || !closeReportAmount(tax.NetTaxCollected, 15) || !closeReportAmount(tax.NetSalesTotal, 165) {
		t.Fatalf("tax summary does not reconcile: returns %.2f returned tax %.2f net tax %.2f net sales %.2f", tax.ReturnsTotal, tax.ReturnedTax, tax.NetTaxCollected, tax.NetSalesTotal)
	}
}

func closeReportAmount(actual, expected float64) bool {
	return math.Abs(actual-expected) < 0.001
}

func TestDebtReportDoesNotMultiplyDebtByPaymentRows(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, remaining_amount REAL, due_date TEXT, status TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT, customer_id TEXT, amount REAL, payment_method TEXT, reference TEXT, notes TEXT, created_at TEXT);
		INSERT INTO customers (id, name) VALUES ('00000000-0000-0000-0000-000000000001', 'Customer One');
		INSERT INTO debts (id, customer_id, amount, remaining_amount, due_date, status) VALUES
			('debt-open', '00000000-0000-0000-0000-000000000001', 100, 70, '2026-09-01', 'partial'),
			('debt-paid', '00000000-0000-0000-0000-000000000001', 50, 0, '2026-09-01', 'paid'),
			('debt-cancelled', '00000000-0000-0000-0000-000000000001', 20, 20, '2026-09-01', 'cancelled');
		INSERT INTO payments (id, transaction_number, customer_id, amount, payment_method, reference, notes, created_at) VALUES
			('payment-1', 'PAY-1', '00000000-0000-0000-0000-000000000001', 10, 'cash', 'ref-1', 'first', '2026-09-19T10:00:00Z'),
			('payment-2', 'PAY-2', '00000000-0000-0000-0000-000000000001', 20, 'cash', 'ref-2', 'second', '2026-09-20T10:00:00Z');
	`)
	if err != nil {
		t.Fatal(err)
	}

	report, err := NewRepository(db).GetDebtsData(context.Background())
	if err != nil {
		t.Fatalf("GetDebtsData: %v", err)
	}
	if report.TotalDebt != 150 || report.TotalPaid != 80 || report.Outstanding != 70 || report.OverdueDebt != 70 {
		t.Fatalf("debt totals include invalid or duplicated entries: total %.2f paid %.2f outstanding %.2f overdue %.2f", report.TotalDebt, report.TotalPaid, report.Outstanding, report.OverdueDebt)
	}
	if len(report.ByCustomer) != 1 {
		t.Fatalf("customer debt rows = %d, want 1", len(report.ByCustomer))
	}
	customer := report.ByCustomer[0]
	if customer.TotalDebt != report.TotalDebt || customer.PaidAmount != report.TotalPaid || customer.Outstanding != report.Outstanding || customer.OverdueAmount != report.OverdueDebt {
		t.Fatalf("customer debt does not reconcile with report totals: %#v vs totals %.2f/%.2f/%.2f/%.2f", customer, report.TotalDebt, report.TotalPaid, report.Outstanding, report.OverdueDebt)
	}
	if len(report.PaymentHistory) != 2 || report.PaymentHistory[0].ReferenceNumber != "ref-2" {
		t.Fatalf("payment history missing or malformed: %#v", report.PaymentHistory)
	}
}

func TestProfitTimelineIncludesDaysAndMonthsWithOnlyExpenses(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE sales (id TEXT PRIMARY KEY, sale_date TEXT, created_at TEXT, total_amount REAL, tax_amount REAL, status TEXT, cost_amount REAL);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, quantity INTEGER, unit_cost REAL, total_amount REAL, tax_amount REAL);
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT, category_id TEXT);
		CREATE TABLE categories (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE expenses (id TEXT PRIMARY KEY, amount REAL, expense_date TEXT, status TEXT);
		INSERT INTO products (id, name) VALUES ('00000000-0000-0000-0000-000000000011', 'Part');
		INSERT INTO sales (id, sale_date, created_at, total_amount, tax_amount, status, cost_amount)
		VALUES ('sale-1', '2026-09-29T10:00:00Z', '2026-09-29T10:00:00Z', 100, 0, 'completed', 40);
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_cost, total_amount, tax_amount)
		VALUES ('line-1', 'sale-1', '00000000-0000-0000-0000-000000000011', 1, 40, 100, 0);
		INSERT INTO expenses (id, amount, expense_date, status) VALUES
		('expense-sep', 10, '2026-09-30', 'approved'),
		('expense-oct', 20, '2026-10-01', 'approved');
	`)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	report, err := NewRepository(db).GetProfitsData(context.Background(), start, end)
	if err != nil {
		t.Fatal(err)
	}
	if !closeReportAmount(report.TotalExpenses, 30) || !closeReportAmount(report.NetProfit, 30) {
		t.Fatalf("period totals = expenses %.2f net %.2f, want 30 and 30", report.TotalExpenses, report.NetProfit)
	}
	if len(report.ByDay) != 3 || !closeReportAmount(report.ByDay[1].NetProfit, -10) || !closeReportAmount(report.ByDay[2].NetProfit, -20) {
		t.Fatalf("daily trend omitted expense-only dates: %#v", report.ByDay)
	}
	if len(report.ByMonth) != 2 || !closeReportAmount(report.ByMonth[0].NetProfit, 50) || !closeReportAmount(report.ByMonth[1].NetProfit, -20) {
		t.Fatalf("monthly trend omitted expense-only months: %#v", report.ByMonth)
	}
}
