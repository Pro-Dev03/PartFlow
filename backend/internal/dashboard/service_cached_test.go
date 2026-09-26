package dashboard

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	_ "modernc.org/sqlite"
)

func TestInventoryStatusPresentationSeparatesReservedAndReturned(t *testing.T) {
	reservedName, _, reservedHealth := inventoryStatusPresentation("RESERVED", "")
	returnedName, _, returnedHealth := inventoryStatusPresentation("RETURNED", "")

	if reservedName != "محجوز" || reservedHealth != "attention" {
		t.Fatalf("reserved presentation = (%q, %q), want (محجوز, attention)", reservedName, reservedHealth)
	}
	if returnedName != "مرتجع" || returnedHealth != "attention" {
		t.Fatalf("returned presentation = (%q, %q), want (مرتجع, attention)", returnedName, returnedHealth)
	}
}

func TestGetLowStockItemsUsesGeneralInventoryQuantity(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE products (
			id TEXT PRIMARY KEY, name TEXT, min_stock_level INTEGER,
			cost_price REAL, selling_price REAL, preferred_supplier_id TEXT,
			is_active INTEGER, deleted_at TEXT
		);
		CREATE TABLE inventory (product_id TEXT, quantity INTEGER);
		CREATE TABLE inventory_items (id TEXT, product_id TEXT, status TEXT, condition TEXT);
		INSERT INTO products (id, name, min_stock_level, is_active) VALUES ('manual-product', 'Manual Product', 1, 1);
		INSERT INTO inventory (product_id, quantity) VALUES ('manual-product', 5);
	`); err != nil {
		t.Fatalf("seed general inventory: %v", err)
	}

	items, err := (&CachedService{db: sqlx.NewDb(db, "sqlite")}).GetLowStockItems(context.Background())
	if err != nil {
		t.Fatalf("get low stock items: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("manual product with quantity 5 was reported as low stock: %#v", items)
	}
}

func TestInventoryStatusPresentationShowsReversedSeparately(t *testing.T) {
	name, _, health := inventoryStatusPresentation("REVERSED", "")
	if name != "شراء ملغى" || health != "neutral" {
		t.Fatalf("reversed presentation = (%q, %q), want (شراء ملغى, neutral)", name, health)
	}
}

func TestFetchInventoryDistributionExcludesArchivedItems(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE inventory_items (
			id TEXT PRIMARY KEY,
			status TEXT,
			condition TEXT,
			purchase_cost REAL,
			selling_price REAL
		);
		CREATE TABLE sales (
			id TEXT PRIMARY KEY,
			status TEXT,
			total_amount REAL
		);
		INSERT INTO inventory_items (id, status, condition, purchase_cost, selling_price)
		VALUES
			('archived-1', 'ARCHIVED', 'NEW', 200, 241),
			('archived-2', 'ARCHIVED', 'NEW', 200, 241),
			('archived-3', 'ARCHIVED', 'NEW', 200, 241),
			('archived-4', 'ARCHIVED', 'NEW', 200, 241),
			('archived-5', 'ARCHIVED', 'NEW', 200, 241),
			('archived-6', 'ARCHIVED', 'NEW', 200, 241),
			('available-1', 'AVAILABLE', 'NEW', 100, 300),
			('available-2', 'AVAILABLE', 'NEW', 100, 300),
			('used-available', 'AVAILABLE', 'USED', 100, 300);
	`); err != nil {
		t.Fatalf("seed inventory distribution rows: %v", err)
	}

	service := &CachedService{db: sqlx.NewDb(db, "sqlite")}
	result := service.fetchInventoryDistribution(context.Background())
	if result == nil {
		t.Fatal("expected inventory distribution")
	}
	if result.TotalItems != 3 || result.TotalValue != 900 {
		t.Fatalf("distribution totals = (%d, %v), want (3, 900)", result.TotalItems, result.TotalValue)
	}
	var foundUsedAvailable bool
	for _, item := range result.Data {
		if item.Name == "مؤرشف" {
			t.Fatal("archived items must not appear in inventory distribution")
		}
		if item.Name == "قطع مستعملة متاحة" {
			foundUsedAvailable = true
		}
	}
	if !foundUsedAvailable {
		t.Fatal("used available items must appear as a separate distribution category")
	}
}

func TestSaleDayContractUsesSaleDateAcrossChartAndActivity(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY,
			total_amount REAL,
			tax_amount REAL,
			cost_amount REAL,
			status TEXT,
			sale_date TEXT,
			created_at TEXT
		);
		CREATE TABLE purchases (id TEXT PRIMARY KEY, total_amount REAL, status TEXT, created_at TEXT);
		INSERT INTO sales (id, total_amount, tax_amount, cost_amount, status, sale_date, created_at) VALUES
			('before-midnight', 880, 0, 0, 'completed', '2026-09-16', '2026-09-16T20:58:00Z'),
			('at-2359', 880, 0, 0, 'completed', '2026-09-16', '2026-09-16T20:59:00Z'),
			('at-0000', 880, 0, 0, 'completed', '2026-09-17', '2026-09-16T21:00:00Z'),
			('after-midnight', 880, 0, 0, 'completed', '2026-09-17', '2026-09-16T21:01:00Z');
	`); err != nil {
		t.Fatalf("seed sale-day contract rows: %v", err)
	}

	service := &CachedService{db: sqlx.NewDb(db, "sqlite")}
	chart := service.fetchSQLiteSalesChart(context.Background(), "2026-09-16")
	if len(chart) != 2 || chart[0].Name != "2026-09-16" || chart[0].Sales != 1760 || chart[1].Name != "2026-09-17" || chart[1].Sales != 1760 {
		t.Fatalf("chart sale days = %#v, want 1760 on each official sale date", chart)
	}

	activityPage, err := service.GetActivity(context.Background(), 1, 10, "sale")
	if err != nil {
		t.Fatalf("get activity: %v", err)
	}
	if len(activityPage.Items) != 4 {
		t.Fatalf("activity count = %d, want 4", len(activityPage.Items))
	}
	for _, item := range activityPage.Items {
		if item.Type != "sale" || item.SaleDate == "" {
			t.Fatalf("activity item missing official sale date: %#v", item)
		}
	}
}

func TestGetActivityReturnsEmptyPageWhenActivityTablesAreMissing(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	service := &CachedService{db: sqlx.NewDb(db, "sqlite")}
	activityPage, err := service.GetActivity(context.Background(), 2, 5, "")
	if err != nil {
		t.Fatalf("get activity from empty database: %v", err)
	}
	if activityPage.Page != 2 || activityPage.PerPage != 5 || activityPage.Total != 0 || activityPage.TotalPages != 0 {
		t.Fatalf("empty activity page = %#v, want page 2 with no results", activityPage)
	}
	if activityPage.Items == nil {
		t.Fatal("empty activity items must be an initialized list")
	}
}

func TestGetActivitySupportsLegacySchemasWithoutOptionalColumns(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL, status TEXT, created_at TEXT);
		CREATE TABLE purchases (id TEXT PRIMARY KEY, total_amount REAL, created_at TEXT);
		INSERT INTO sales (id, total_amount, status, created_at) VALUES ('legacy-sale', 120, 'completed', '2026-09-21T10:00:00Z');
		INSERT INTO purchases (id, total_amount, created_at) VALUES ('legacy-purchase', 80, '2026-09-21T09:00:00Z');
	`); err != nil {
		t.Fatalf("seed legacy activity schema: %v", err)
	}

	activityPage, err := (&CachedService{db: sqlx.NewDb(db, "sqlite")}).GetActivity(context.Background(), 1, 10, "")
	if err != nil {
		t.Fatalf("get activity from legacy schema: %v", err)
	}
	if len(activityPage.Items) != 2 {
		t.Fatalf("legacy activity count = %d, want 2", len(activityPage.Items))
	}
}

func TestSQLiteSalesChartUsesCreatedAtOnlyForLegacyRows(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY,
			total_amount REAL,
			status TEXT,
			created_at TEXT
		);
		INSERT INTO sales (id, total_amount, status, created_at)
		VALUES ('legacy-sale', 880, 'completed', '2026-09-16T20:59:00Z');
	`); err != nil {
		t.Fatalf("seed legacy sale row: %v", err)
	}

	service := &CachedService{db: sqlx.NewDb(db, "sqlite")}
	chart := service.fetchSQLiteSalesChart(context.Background(), "2026-09-16")
	if len(chart) != 1 || chart[0].Name != "2026-09-16" || chart[0].Sales != 880 {
		t.Fatalf("legacy chart = %#v, want one fallback point on 2026-09-16", chart)
	}
}

func TestSQLiteSalesChartFallsBackToProductCostWhenStoredCostIsZero(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY, total_amount REAL, tax_amount REAL, cost_amount REAL,
			status TEXT, sale_date TEXT, created_at TEXT
		);
		CREATE TABLE products (id TEXT PRIMARY KEY, cost_price REAL);
		CREATE TABLE sale_items (
			id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT,
			quantity REAL, unit_cost REAL
		);
		INSERT INTO products (id, cost_price) VALUES ('product', 20);
		INSERT INTO sales (id, total_amount, tax_amount, cost_amount, status, sale_date, created_at)
		VALUES ('sale', 27, 2, 0, 'completed', '2026-09-26', '2026-09-26T08:00:00Z');
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_cost)
		VALUES ('line', 'sale', 'product', 1, 0);
	`); err != nil {
		t.Fatalf("seed sale cost fallback: %v", err)
	}

	chart := (&CachedService{db: sqlx.NewDb(db, "sqlite")}).fetchSQLiteSalesChart(context.Background(), "2026-09-26")
	if len(chart) != 1 || chart[0].Sales != 25 || chart[0].Profit != 5 {
		t.Fatalf("gross profit chart = %#v, want sales 25 and gross profit 5 using product cost 20", chart)
	}
}

func TestSalesChartReconcilesProductCostFallbackAndCompletedReturns(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE sales (
			id TEXT PRIMARY KEY, total_amount REAL, tax_amount REAL, cost_amount REAL,
			status TEXT, sale_date TEXT, created_at TEXT
		);
		CREATE TABLE products (id TEXT PRIMARY KEY, cost_price REAL);
		CREATE TABLE sale_items (
			id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT,
			quantity REAL, unit_cost REAL, total_amount REAL, tax_amount REAL
		);
		CREATE TABLE returns (
			id TEXT PRIMARY KEY, reference_number TEXT, sale_id TEXT,
			total_refund_amount REAL, status TEXT, return_date TEXT, refund_date TEXT, created_at TEXT
		);
		CREATE TABLE return_items (
			id TEXT PRIMARY KEY, return_id TEXT, sale_item_id TEXT, product_id TEXT,
			quantity_returned REAL, total_refund_amount REAL, original_cost REAL
		);
		INSERT INTO products (id, cost_price) VALUES ('product', 20);
		INSERT INTO sales (id, total_amount, tax_amount, cost_amount, status, sale_date, created_at)
		VALUES ('sale', 54, 4, 0, 'completed', '2026-09-26', '2026-09-26T10:00:00Z');
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_cost, total_amount, tax_amount)
		VALUES ('line', 'sale', 'product', 2, 0, 54, 4);
		INSERT INTO returns (id, reference_number, sale_id, total_refund_amount, status, return_date, refund_date, created_at)
		VALUES ('return', 'RET-1', 'sale', 27, 'COMPLETED', '2026-09-27', '2026-09-27', '2026-09-27T08:00:00Z');
		INSERT INTO return_items (id, return_id, sale_item_id, product_id, quantity_returned, total_refund_amount, original_cost)
		VALUES ('return-line', 'return', 'line', 'product', 1, 27, 20);
	`)
	if err != nil {
		t.Fatalf("seed sales chart reconciliation rows: %v", err)
	}

	location, err := accounting.StoreLocation()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, location)
	chart := (&CachedService{db: db}).fetchSalesChartForRange(context.Background(), now, 2)
	if len(chart) != 2 {
		t.Fatalf("chart = %#v, want sale day and return day", chart)
	}
	if chart[0].Name != "2026-09-26" || chart[0].Sales != 50 || chart[0].Profit != 10 {
		t.Fatalf("sale-day chart = %#v, want net sales 50 and gross profit 10 using product cost fallback", chart[0])
	}
	if chart[1].Name != "2026-09-27" || chart[1].Sales != -25 || chart[1].Profit != -5 {
		t.Fatalf("return-day chart = %#v, want net sales -25 and gross profit -5 after partial return", chart[1])
	}
}

func TestInventoryStatusPresentationSeparatesUsedSold(t *testing.T) {
	name, _, health := inventoryStatusPresentation("SOLD", "USED")
	if name != "قطع مستعملة مباعة" || health != "neutral" {
		t.Fatalf("used sold presentation = (%q, %q), want (قطع مستعملة مباعة, neutral)", name, health)
	}
}

func TestInventoryStatusPresentationSeparatesUsedAvailable(t *testing.T) {
	name, _, health := inventoryStatusPresentation("AVAILABLE", "USED")
	if name != "قطع مستعملة متاحة" || health != "good" {
		t.Fatalf("used available presentation = (%q, %q), want (قطع مستعملة متاحة, good)", name, health)
	}
}

func TestRecentSaleActivityIncludesSellerNameAndTime(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			first_name TEXT,
			last_name TEXT
		);
		CREATE TABLE sales (
			id TEXT PRIMARY KEY,
			total_amount REAL,
			status TEXT,
			user_id TEXT,
			sale_date TEXT,
			created_at TEXT
		);
		CREATE TABLE purchases (
			id TEXT PRIMARY KEY,
			total_amount REAL,
			status TEXT,
			created_at TEXT
		);
		INSERT INTO users (id, first_name, last_name) VALUES ('user-1', 'أحمد', 'الزيدي');
		INSERT INTO sales (id, total_amount, status, user_id, sale_date, created_at) VALUES
			('sale-1', 300, 'completed', 'user-1', '2026-09-16', '2026-09-16T10:30:00Z');
	`); err != nil {
		t.Fatalf("seed sale activity rows: %v", err)
	}

	service := &CachedService{db: sqlx.NewDb(db, "sqlite")}
	activityPage, err := service.GetActivity(context.Background(), 1, 10, "sale")
	if err != nil {
		t.Fatalf("get activity: %v", err)
	}
	if len(activityPage.Items) != 1 {
		t.Fatalf("activity count = %d, want 1", len(activityPage.Items))
	}
	item := activityPage.Items[0]
	if item.SellerName != "أحمد الزيدي" {
		t.Fatalf("seller name = %q, want %q", item.SellerName, "أحمد الزيدي")
	}
	if item.Time == "" {
		t.Fatal("activity time must be populated for sales")
	}
}
