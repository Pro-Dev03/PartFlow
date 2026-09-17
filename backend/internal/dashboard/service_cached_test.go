package dashboard

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
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
