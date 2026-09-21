package archive

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestRepositoryListIncludesBusinessHistorySources(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			first_name TEXT,
			last_name TEXT,
			email TEXT
		);
		CREATE TABLE audit_logs (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			action TEXT,
			entity_type TEXT,
			entity_id TEXT,
			description TEXT,
			new_values TEXT,
			changes TEXT,
			status TEXT,
			created_at TEXT
		);
		CREATE TABLE inventory_movements (
			id TEXT PRIMARY KEY,
			item_id TEXT,
			product_id TEXT,
			movement_type TEXT,
			quantity INTEGER,
			before_quantity INTEGER,
			after_quantity INTEGER,
			reference_type TEXT,
			reference_id TEXT,
			reason TEXT,
			created_by TEXT,
			created_at TEXT,
			is_reversed INTEGER DEFAULT 0
		);
		CREATE TABLE item_history (
			id TEXT PRIMARY KEY,
			inventory_item_id TEXT,
			event_type TEXT,
			event_date TEXT,
			reference_type TEXT,
			reference_id TEXT,
			description TEXT,
			metadata TEXT,
			created_by TEXT,
			created_at TEXT
		);
		CREATE TABLE ledger_entries (
			id TEXT PRIMARY KEY,
			ledger_type TEXT,
			entity_id TEXT,
			transaction_type TEXT,
			reference_id TEXT,
			reference_type TEXT,
			amount REAL,
			balance REAL,
			previous_balance REAL,
			description TEXT,
			metadata TEXT,
			created_by TEXT,
			created_at TEXT
		);
	`); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	userID := "user-1"
	createdAt := time.Date(2026, time.September, 18, 14, 0, 0, 0, time.UTC).Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO users (id, first_name, last_name, email) VALUES (?, ?, ?, ?)`, userID, "Ahmed", "Ali", "ahmed@example.com"); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, description, new_values, changes, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, "audit-1", userID, "CREATE_SALE", "sale", "sale-123", "Sale created", "{\"amount\":120}", "{\"amount\":120}", "success", createdAt); err != nil {
		t.Fatalf("insert audit log: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at, is_reversed) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, "movement-1", "item-1", "prod-1", "SALE", -1, 5, 4, "sale", "sale-123", "Sold item", userID, createdAt, 0); err != nil {
		t.Fatalf("insert inventory movement: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO item_history (id, inventory_item_id, event_type, event_date, reference_type, reference_id, description, metadata, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, "item-history-1", "item-1", "sold", createdAt, "sale", "sale-123", "Item sold", "{\"quantity\":1}", userID, createdAt); err != nil {
		t.Fatalf("insert item history: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO ledger_entries (id, ledger_type, entity_id, transaction_type, reference_id, reference_type, amount, balance, previous_balance, description, metadata, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, "ledger-1", "CUSTOMER", "customer-1", "SALE", "sale-123", "sale", 120.0, 120.0, 0.0, "Customer payment", "{\"source\":\"sale\"}", userID, createdAt); err != nil {
		t.Fatalf("insert ledger entry: %v", err)
	}

	repo := NewRepository(db)
	result, total, err := repo.List(context.Background(), ListRequest{Page: 1, PerPage: 25, Search: "sale-123"})
	if err != nil {
		t.Fatalf("list archive: %v", err)
	}
	if total < 4 {
		t.Fatalf("expected at least 4 events, got %d: %+v", total, result)
	}

	inventoryOnly, totalInventory, err := repo.List(context.Background(), ListRequest{Page: 1, PerPage: 25, Section: "inventory"})
	if err != nil {
		t.Fatalf("filter inventory: %v", err)
	}
	if totalInventory == 0 || len(inventoryOnly) == 0 {
		t.Fatalf("expected inventory items in archive, got total=%d results=%d", totalInventory, len(inventoryOnly))
	}

	customerOnly, totalCustomers, err := repo.List(context.Background(), ListRequest{Page: 1, PerPage: 25, Section: "customers"})
	if err != nil {
		t.Fatalf("filter customers: %v", err)
	}
	if totalCustomers == 0 || len(customerOnly) == 0 {
		t.Fatalf("expected customer business history, got total=%d results=%d", totalCustomers, len(customerOnly))
	}

	filtered, filteredTotal, err := repo.List(context.Background(), ListRequest{Page: 1, PerPage: 25, Status: "success"})
	if err != nil {
		t.Fatalf("filter status: %v", err)
	}
	if filteredTotal == 0 || len(filtered) == 0 {
		t.Fatalf("expected status-filtered business history, got total=%d results=%d", filteredTotal, len(filtered))
	}
}

func TestRepositoryPreservesImmutableSnapshotAfterSourceUpdate(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE users (id TEXT PRIMARY KEY, first_name TEXT, last_name TEXT, email TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, description TEXT, new_values TEXT, changes TEXT, status TEXT, created_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT, is_reversed INTEGER DEFAULT 0);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, event_type TEXT, event_date TEXT, reference_type TEXT, reference_id TEXT, description TEXT, metadata TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE ledger_entries (id TEXT PRIMARY KEY, ledger_type TEXT, entity_id TEXT, transaction_type TEXT, reference_id TEXT, reference_type TEXT, amount REAL, balance REAL, previous_balance REAL, description TEXT, metadata TEXT, created_by TEXT, created_at TEXT);
	`); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	const snapshot = `{"sale_id":"sale-1","unit_price":125,"unit_cost":80,"quantity":2,"discount":10,"total":240,"customer_id":"customer-1","invoice_number":"INV-1"}`
	if _, err := db.Exec(`INSERT INTO audit_logs (id, action, entity_type, entity_id, description, new_values, changes, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, "audit-1", "CREATE_SALE", "sale", "sale-1", "sale created", snapshot, snapshot, "success", "2026-09-21T10:00:00Z"); err != nil {
		t.Fatalf("insert snapshot: %v", err)
	}
	if _, err := db.Exec(`UPDATE audit_logs SET new_values = ? WHERE id = ?`, `{"sale_id":"sale-1","unit_price":999,"total":1}`, "missing-update-source"); err != nil {
		t.Fatalf("update current source fixture: %v", err)
	}

	events, total, err := NewRepository(db).List(context.Background(), ListRequest{Page: 1, PerPage: 25, EntityID: "sale-1"})
	if err != nil {
		t.Fatalf("list archive: %v", err)
	}
	if total != 1 || len(events) != 1 {
		t.Fatalf("expected one immutable event, got total=%d events=%d", total, len(events))
	}
	if !strings.Contains(events[0].Details, `"unit_price":125`) || strings.Contains(events[0].Details, `"unit_price":999`) {
		t.Fatalf("archive snapshot changed unexpectedly: %s", events[0].Details)
	}
}

func TestRepositoryIncludesLegacyPartyLedgerReferences(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE users (id TEXT PRIMARY KEY, first_name TEXT, last_name TEXT, email TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, description TEXT, new_values TEXT, changes TEXT, status TEXT, created_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT, is_reversed INTEGER DEFAULT 0);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, event_type TEXT, event_date TEXT, reference_type TEXT, reference_id TEXT, description TEXT, metadata TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE ledger_entries (id TEXT PRIMARY KEY, ledger_type TEXT, entity_id TEXT, transaction_type TEXT, reference_id TEXT, reference_type TEXT, amount REAL, balance REAL, previous_balance REAL, description TEXT, metadata TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL, description TEXT, reference_id TEXT, reference_type TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL, description TEXT, reference_id TEXT, reference_type TEXT, created_by TEXT, created_at TEXT);
	`); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	for _, query := range []string{
		`INSERT INTO customer_ledger VALUES ('customer-ledger-1', 'customer-1', 'debit', 'SALE', 100, 100, 'sale', 'sale-1', 'sale', 'user-1', '2026-09-21T10:00:00Z')`,
		`INSERT INTO supplier_ledger VALUES ('supplier-ledger-1', 'supplier-1', 'debit', 'PURCHASE', 200, 200, 'purchase', 'purchase-1', 'purchase', 'user-1', '2026-09-21T10:00:00Z')`,
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("insert ledger source: %v", err)
		}
	}

	events, total, err := NewRepository(db).List(context.Background(), ListRequest{Page: 1, PerPage: 25, Search: "sale-1"})
	if err != nil {
		t.Fatalf("list customer ledger: %v", err)
	}
	if total != 1 || events[0].ReferenceID != "sale-1" || events[0].Section != "customers" {
		t.Fatalf("customer relationship missing: total=%d events=%+v", total, events)
	}

	events, total, err = NewRepository(db).List(context.Background(), ListRequest{Page: 1, PerPage: 25, Search: "purchase-1"})
	if err != nil {
		t.Fatalf("list supplier ledger: %v", err)
	}
	if total != 1 || events[0].ReferenceID != "purchase-1" || events[0].Section != "suppliers" {
		t.Fatalf("supplier relationship missing: total=%d events=%+v", total, events)
	}
}
