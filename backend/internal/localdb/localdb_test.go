package localdb

import (
	"path/filepath"
	"testing"
)

func TestOpenInitializesLocalDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partflow.db")
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", path)
	t.Setenv("PARTFLOW_BOOTSTRAP_OWNER_PASSWORD", "test-only-owner-password")

	database, err := Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer database.DB.Close()

	if database.Path != path {
		t.Fatalf("database path = %q, want %q", database.Path, path)
	}

	if err := SetMetadata(database.DB, "operating_mode", "offline"); err != nil {
		t.Fatalf("SetMetadata() error = %v", err)
	}

	var mode string
	if err := database.DB.QueryRow("SELECT value FROM local_metadata WHERE key = ?", "operating_mode").Scan(&mode); err != nil {
		t.Fatalf("read operating mode error = %v", err)
	}
	if mode != "offline" {
		t.Fatalf("operating mode = %q, want offline", mode)
	}

	if metadata, err := GetMetadata(database.DB, "operating_mode"); err != nil || metadata != "offline" {
		t.Fatalf("GetMetadata() = %q, %v; want offline, nil", metadata, err)
	}

	if _, err := EnqueueSyncOperation(database.DB, "sale", "sale-123", "create", `{"amount": 125}`); err != nil {
		t.Fatalf("EnqueueSyncOperation() error = %v", err)
	}

	count, err := GetPendingSyncCount(database.DB)
	if err != nil {
		t.Fatalf("GetPendingSyncCount() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("GetPendingSyncCount() = %d, want 1", count)
	}

	var queueID string
	if err := database.DB.QueryRow("SELECT id FROM sync_queue WHERE entity_id = ? LIMIT 1", "sale-123").Scan(&queueID); err != nil {
		t.Fatalf("read sync queue id error = %v", err)
	}
	if err := MarkSyncOperationFailed(database.DB, queueID, "temporary failure"); err != nil {
		t.Fatalf("MarkSyncOperationFailed() error = %v", err)
	}

	var nextRetryAt string
	if err := database.DB.QueryRow("SELECT next_retry_at FROM sync_queue WHERE id = ?", queueID).Scan(&nextRetryAt); err != nil {
		t.Fatalf("read next_retry_at error = %v", err)
	}
	if nextRetryAt == "" {
		t.Fatal("next_retry_at is empty after failure")
	}

	countAfterFailure, err := GetPendingSyncCount(database.DB)
	if err != nil {
		t.Fatalf("GetPendingSyncCount() after failure error = %v", err)
	}
	if countAfterFailure != 0 {
		t.Fatalf("GetPendingSyncCount() after failure = %d, want 0 while retry is scheduled", countAfterFailure)
	}

	if err := RecordSyncConflict(database.DB, "sale", "sale-123", "sales", "update", "2026-08-30T02:00:00Z", "2026-08-30T03:00:00Z", "remote newer", `{"amount": 125}`); err != nil {
		t.Fatalf("RecordSyncConflict() error = %v", err)
	}

	if err := SaveLocalSession(database.DB, "user-123", "owner@partflow.com", "Owner", "access-token", "refresh-token", "2026-08-31T00:00:00Z"); err != nil {
		t.Fatalf("SaveLocalSession() error = %v", err)
	}

	localSession, err := GetLocalSession(database.DB)
	if err != nil {
		t.Fatalf("GetLocalSession() error = %v", err)
	}
	if localSession == nil || localSession["email"] != "owner@partflow.com" {
		t.Fatalf("GetLocalSession() = %#v, want user session for owner@partflow.com", localSession)
	}

	var conflictCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM sync_conflicts WHERE entity_id = ?", "sale-123").Scan(&conflictCount); err != nil || conflictCount != 1 {
		t.Fatalf("sync_conflicts count = %d, %v; want 1, nil", conflictCount, err)
	}

	// Verify all business tables exist
	requiredTables := []string{
		"local_metadata", "sync_queue", "sync_conflicts", "local_sessions",
		"categories", "brands", "products", "customers", "suppliers",
		"inventory_items", "sales", "sale_items",
		"purchases", "purchase_items", "payments", "debts",
		"expense_categories", "expenses", "returns", "return_items", "warranty_claims",
		"locations", "inventory_movements", "reservations", "customer_ledger",
		"supplier_ledger", "ledger_entries", "inspection_items",
	}

	for _, table := range requiredTables {
		var exists int
		if err := database.DB.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&exists); err != nil || exists != 1 {
			t.Fatalf("table %s does not exist", table)
		}
	}

	for _, table := range []string{"daily_sales_summary", "monthly_sales_summary", "daily_inventory_summary", "daily_debt_summary", "daily_profit_summary"} {
		var exists int
		if err := database.DB.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&exists); err != nil || exists != 1 {
			t.Fatalf("aggregation table %s does not exist in local SQLite schema", table)
		}
	}

	for _, column := range []string{"brand_id", "model", "barcode", "cost_price", "track_serial", "track_individual", "deleted_at", "preferred_supplier_id"} {
		var count int
		if err := database.DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('products') WHERE name = ?", column).Scan(&count); err != nil || count != 1 {
			t.Fatalf("products.%s column missing: %v", column, err)
		}
	}

	for _, column := range []string{"tax_id"} {
		var count int
		if err := database.DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('customers') WHERE name = ?", column).Scan(&count); err != nil || count != 1 {
			t.Fatalf("customers.%s column missing: %v", column, err)
		}
	}

	var userCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", "owner@partflow.com").Scan(&userCount); err != nil || userCount != 1 {
		t.Fatalf("default owner user missing in local SQLite DB: %v", err)
	}
}

func TestSQLiteCompatibilityHelpers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "compat.db")
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", path)

	database, err := Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer database.DB.Close()

	if _, err := database.DB.Exec("INSERT INTO products (id, sku, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", "prod-1", "SKU-1", "Oil Filter", "Replacement filter", "2026-08-30T00:00:00Z", "2026-08-30T00:00:00Z"); err != nil {
		t.Fatalf("seed product failed: %v", err)
	}

	var matches int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM products WHERE LOWER(name) LIKE LOWER(?)", "%oil%").Scan(&matches); err != nil {
		t.Fatalf("case-insensitive search compatibility failed: %v", err)
	}
	if matches != 1 {
		t.Fatalf("case-insensitive search count = %d, want 1", matches)
	}

	if _, err := database.DB.Exec("CREATE TABLE IF NOT EXISTS inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, reserved_quantity INTEGER, location TEXT, created_at TEXT, updated_at TEXT)"); err != nil {
		t.Fatalf("create compatibility inventory table failed: %v", err)
	}
	if _, err := database.DB.Exec("INSERT INTO inventory (id, product_id, quantity, reserved_quantity, location, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)", "inv-1", "prod-1", 5, 1, "A1", "2026-08-30T00:00:00Z", "2026-08-30T00:00:00Z"); err != nil {
		t.Fatalf("seed compatibility inventory failed: %v", err)
	}

	if err := database.DB.QueryRow("SELECT COUNT(*) FROM inventory WHERE product_id = ?", "prod-1").Scan(&matches); err != nil {
		t.Fatalf("inventory compatibility query failed: %v", err)
	}
	if matches != 1 {
		t.Fatalf("inventory compatibility count = %d, want 1", matches)
	}
}

func TestSeedLocalSnapshotFromOnlineData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.db")
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", path)

	database, err := Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer database.DB.Close()

	dataset := map[string]any{
		"products": []map[string]any{{
			"id":            "prod-1",
			"sku":           "SKU-1",
			"name":          "Oil Filter",
			"selling_price": 120.5,
			"created_at":    "2026-08-30T00:00:00Z",
			"updated_at":    "2026-08-30T00:00:00Z",
		}},
		"customers": []map[string]any{{
			"id":         "cust-1",
			"code":       "C-001",
			"name":       "Ahmad",
			"created_at": "2026-08-30T00:00:00Z",
			"updated_at": "2026-08-30T00:00:00Z",
		}},
		"customer_ledger": []map[string]any{{
			"id": "customer-ledger-1", "customer_id": "cust-1", "type": "debit", "amount": 25,
			"balance": 25, "description": "Invoice", "created_at": "2026-08-30T00:00:00Z",
		}},
		"locations": []map[string]any{{
			"id": "location-1", "name": "Main", "type": "warehouse", "is_active": true,
			"created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
		}},
	}

	if err := SeedLocalSnapshot(database.DB, dataset); err != nil {
		t.Fatalf("SeedLocalSnapshot() error = %v", err)
	}

	var productCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM products WHERE id = ?", "prod-1").Scan(&productCount); err != nil || productCount != 1 {
		t.Fatalf("products row not synced: count=%d err=%v", productCount, err)
	}

	var customerCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM customers WHERE id = ?", "cust-1").Scan(&customerCount); err != nil || customerCount != 1 {
		t.Fatalf("customers row not synced: count=%d err=%v", customerCount, err)
	}

	var ledgerCount, locationCount int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM customer_ledger WHERE id = ?", "customer-ledger-1").Scan(&ledgerCount); err != nil || ledgerCount != 1 {
		t.Fatalf("customer ledger row not synced: count=%d err=%v", ledgerCount, err)
	}
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM locations WHERE id = ?", "location-1").Scan(&locationCount); err != nil || locationCount != 1 {
		t.Fatalf("location row not synced: count=%d err=%v", locationCount, err)
	}
}

func TestSeedLocalSnapshotCloudSchemaAliases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cloud-shape.db")
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", path)

	database, err := Open()
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer database.DB.Close()

	dataset := map[string]any{
		"categories": []map[string]any{{
			"id": "cat-1", "name": "Filters", "created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
		}},
		"products": []map[string]any{{
			"id": "prod-1", "sku": "SKU-1", "name": "Oil Filter", "barcode": "123", "category_id": "cat-1",
			"cost_price": 50, "selling_price": 80, "created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
		}},
		"customers": []map[string]any{{
			"id": "cust-1", "code": "C-001", "name": "Ahmad", "created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
		}},
		"sales": []map[string]any{{
			"id": "sale-1", "invoice_number": "INV-1", "customer_id": "cust-1", "sale_date": "2026-08-30",
			"subtotal": 80, "total_amount": 80, "paid_amount": 80, "remaining_amount": 0, "status": "completed",
			"created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
		}},
		"sale_items": []map[string]any{{
			"id": "sale-item-1", "sale_id": "sale-1", "product_id": "prod-1", "quantity": 1, "unit_price": 80,
			"total_amount": 80, "created_at": "2026-08-30T00:00:00Z",
		}},
		"expenses": []map[string]any{{
			"id": "expense-1", "reference_number": "EXP-1", "category": "Rent", "amount": 100,
			"description": "Monthly rent", "expense_date": "2026-08-30", "created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
		}},
		"debts": []map[string]any{{
			"id": "debt-1", "customer_id": "cust-1", "amount": 25, "remaining_amount": 25,
			"due_date": "2026-09-30", "status": "pending", "created_at": "2026-08-30T00:00:00Z", "updated_at": "2026-08-30T00:00:00Z",
		}},
	}

	if err := SeedLocalSnapshot(database.DB, dataset); err != nil {
		t.Fatalf("SeedLocalSnapshot() error = %v", err)
	}

	var saleNumber, productBarcode, expenseTitle string
	if err := database.DB.QueryRow("SELECT sale_number FROM sales WHERE id = ?", "sale-1").Scan(&saleNumber); err != nil || saleNumber != "INV-1" {
		t.Fatalf("sale alias not synced: value=%q err=%v", saleNumber, err)
	}
	if err := database.DB.QueryRow("SELECT barcode FROM products WHERE id = ?", "prod-1").Scan(&productBarcode); err != nil || productBarcode != "123" {
		t.Fatalf("product barcode not synced: value=%q err=%v", productBarcode, err)
	}
	if err := database.DB.QueryRow("SELECT title FROM expenses WHERE id = ?", "expense-1").Scan(&expenseTitle); err != nil || expenseTitle != "Monthly rent" {
		t.Fatalf("expense title fallback not synced: value=%q err=%v", expenseTitle, err)
	}

	var debtPaid float64
	if err := database.DB.QueryRow("SELECT paid_amount FROM debts WHERE id = ?", "debt-1").Scan(&debtPaid); err != nil || debtPaid != 0 {
		t.Fatalf("debt paid amount default incorrect: value=%v err=%v", debtPaid, err)
	}
}
