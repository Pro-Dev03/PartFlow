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

	if reservedName != "محجوز" || reservedHealth != "low" {
		t.Fatalf("reserved presentation = (%q, %q), want (محجوز, low)", reservedName, reservedHealth)
	}
	if returnedName != "مرتجع" || returnedHealth != "low" {
		t.Fatalf("returned presentation = (%q, %q), want (مرتجع, low)", returnedName, returnedHealth)
	}
}

func TestInventoryStatusPresentationShowsReversedSeparately(t *testing.T) {
	name, _, health := inventoryStatusPresentation("REVERSED", "")
	if name != "شراء ملغى" || health != "low" {
		t.Fatalf("reversed presentation = (%q, %q), want (شراء ملغى, low)", name, health)
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
			('available-2', 'AVAILABLE', 'NEW', 100, 300);
	`); err != nil {
		t.Fatalf("seed inventory distribution rows: %v", err)
	}

	service := &CachedService{db: sqlx.NewDb(db, "sqlite")}
	result := service.fetchInventoryDistribution(context.Background())
	if result == nil {
		t.Fatal("expected inventory distribution")
	}
	if result.TotalItems != 2 || result.TotalValue != 600 {
		t.Fatalf("distribution totals = (%d, %v), want (2, 600)", result.TotalItems, result.TotalValue)
	}
	for _, item := range result.Data {
		if item.Name == "مؤرشف" {
			t.Fatal("archived items must not appear in inventory distribution")
		}
	}
}

func TestInventoryStatusPresentationSeparatesUsedSold(t *testing.T) {
	name, _, health := inventoryStatusPresentation("SOLD", "USED")
	if name != "قطع مستعملة مباعة" || health != "low" {
		t.Fatalf("used sold presentation = (%q, %q), want (قطع مستعملة مباعة, low)", name, health)
	}
}

func TestInventoryStatusPresentationSeparatesUsedAvailable(t *testing.T) {
	name, _, health := inventoryStatusPresentation("AVAILABLE", "USED")
	if name != "قطع مستعملة متاحة" || health != "good" {
		t.Fatalf("used available presentation = (%q, %q), want (قطع مستعملة متاحة, good)", name, health)
	}
}
