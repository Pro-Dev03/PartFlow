package api

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/aggregations"
	_ "modernc.org/sqlite"
)

func TestDailySalesSummaryQueryUsesSQLiteSafeSyntax(t *testing.T) {
	query := buildDailySalesSummaryQuery()
	if strings.Contains(query, "::date") {
		t.Fatalf("daily sales query should not use PostgreSQL-only cast syntax: %s", query)
	}
	if strings.Contains(query, "NOW()") {
		t.Fatalf("daily sales query should not use PostgreSQL NOW(): %s", query)
	}
	if !strings.Contains(query, "LEFT JOIN daily_sales_summary s ON s.date = dates.summary_date") {
		t.Fatalf("daily sales query should join the real daily summary table by date: %s", query)
	}
}

func TestMonthlySalesSummaryQueryUsesSQLiteSafeSyntax(t *testing.T) {
	query := buildMonthlySalesSummaryQuery()
	if strings.Contains(query, "::int") {
		t.Fatalf("monthly sales query should not use PostgreSQL-only cast syntax: %s", query)
	}
	if strings.Contains(query, "NOW()") {
		t.Fatalf("monthly sales query should not use PostgreSQL NOW(): %s", query)
	}
	if !strings.Contains(query, "LEFT JOIN monthly_sales_summary s ON s.year = dates.summary_year AND s.month = dates.summary_month") {
		t.Fatalf("monthly sales query should join the real monthly summary table by year/month: %s", query)
	}
}

func TestAggregationQueriesExecuteOnSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE daily_sales_summary (
		date TEXT PRIMARY KEY,
		total_sales INTEGER DEFAULT 0,
		total_revenue REAL DEFAULT 0,
		total_profit REAL DEFAULT 0,
		total_customers INTEGER DEFAULT 0,
		average_order_value REAL DEFAULT 0,
		total_items_sold INTEGER DEFAULT 0,
		cash_sales REAL DEFAULT 0,
		card_sales REAL DEFAULT 0,
		debt_sales REAL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create daily_sales_summary: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE monthly_sales_summary (
		year INTEGER NOT NULL,
		month INTEGER NOT NULL,
		total_sales INTEGER DEFAULT 0,
		total_revenue REAL DEFAULT 0,
		total_profit REAL DEFAULT 0,
		total_customers INTEGER DEFAULT 0,
		average_order_value REAL DEFAULT 0,
		total_items_sold INTEGER DEFAULT 0,
		cash_sales REAL DEFAULT 0,
		card_sales REAL DEFAULT 0,
		debt_sales REAL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (year, month)
	)`); err != nil {
		t.Fatalf("create monthly_sales_summary: %v", err)
	}

	if err := db.QueryRow(buildDailySalesSummaryQuery(), "2026-08-30", "2026-08-30").Scan(new(string), new(int), new(float64), new(float64), new(int), new(float64), new(int), new(float64), new(float64), new(float64), new(string)); err != nil {
		t.Fatalf("daily aggregation query failed on sqlite: %v", err)
	}
	if err := db.QueryRow(buildMonthlySalesSummaryQuery(), 2026, 8, 2026, 8).Scan(new(int), new(int), new(int), new(float64), new(float64), new(int), new(float64), new(int), new(float64), new(float64), new(float64), new(string)); err != nil {
		t.Fatalf("monthly aggregation query failed on sqlite: %v", err)
	}
}

func TestDailySalesSummaryStructScansOnSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE daily_sales_summary (
		date TEXT PRIMARY KEY,
		total_sales INTEGER DEFAULT 0,
		total_revenue REAL DEFAULT 0,
		total_profit REAL DEFAULT 0,
		total_customers INTEGER DEFAULT 0,
		average_order_value REAL DEFAULT 0,
		total_items_sold INTEGER DEFAULT 0,
		cash_sales REAL DEFAULT 0,
		card_sales REAL DEFAULT 0,
		debt_sales REAL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create daily_sales_summary: %v", err)
	}

	xdb := sqlx.NewDb(db, "sqlite")
	var summary aggregations.DailySalesSummary
	err = xdb.Get(&summary, buildDailySalesSummaryQuery(), "2026-08-30", "2026-08-30")
	if err != nil {
		t.Fatalf("daily sales summary should scan on sqlite: %v", err)
	}
	if summary.Date == "" {
		t.Fatalf("daily sales summary date should be populated")
	}
}
