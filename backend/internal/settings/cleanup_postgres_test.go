package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/internal/debts"
	"github.com/partflow/smart-store/internal/inventory"
	"github.com/partflow/smart-store/internal/payments"
	"github.com/partflow/smart-store/internal/products"
	"github.com/partflow/smart-store/internal/purchases"
	"github.com/partflow/smart-store/internal/reports"
	"github.com/partflow/smart-store/internal/returns"
	"github.com/partflow/smart-store/internal/sales"
	"github.com/partflow/smart-store/internal/supplierreturns"
)

// These PostgreSQL integration tests are opt-in and refuse non-loopback or
// non-test databases. They create a random schema and drop it when complete.
func openIsolatedPostgresCleanupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	rawDSN := strings.TrimSpace(os.Getenv("PARTFLOW_TEST_POSTGRES_DSN"))
	if rawDSN == "" {
		t.Skip("set PARTFLOW_TEST_POSTGRES_DSN to a loopback test database to run PostgreSQL cleanup integration tests")
	}
	parsed, err := url.Parse(rawDSN)
	if err != nil {
		t.Fatalf("parse PostgreSQL test DSN: %v", err)
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "localhost" && host != "127.0.0.1" && host != "::1" {
		t.Fatalf("refusing PostgreSQL integration test against non-loopback host %q", host)
	}
	databaseName := strings.TrimPrefix(parsed.Path, "/")
	if !strings.Contains(strings.ToLower(databaseName), "test") && !strings.HasSuffix(strings.ToLower(databaseName), "_e2e") {
		t.Fatalf("refusing PostgreSQL integration test against database %q; database name must include 'test' or end in '_e2e'", databaseName)
	}

	baseDB, err := database.Connect(rawDSN)
	if err != nil {
		t.Fatalf("connect to isolated PostgreSQL test database: %v", err)
	}
	schema := "partflow_cleanup_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := baseDB.Exec(`CREATE SCHEMA "` + schema + `"`); err != nil {
		baseDB.Close()
		t.Fatalf("create isolated PostgreSQL schema: %v", err)
	}
	parsed.Query().Set("search_path", schema)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	testDB, err := database.Connect(parsed.String())
	if err != nil {
		_, _ = baseDB.Exec(`DROP SCHEMA "` + schema + `" CASCADE`)
		baseDB.Close()
		t.Fatalf("connect to isolated PostgreSQL schema: %v", err)
	}
	testDB.SetMaxOpenConns(5)
	t.Cleanup(func() {
		_ = testDB.Close()
		_, _ = baseDB.Exec(`DROP SCHEMA "` + schema + `" CASCADE`)
		_ = baseDB.Close()
	})
	return testDB
}

func TestPostgresHistoricalCleanupUsesStoreTimezoneAndReversesPurchasePayments(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE purchases (
			id UUID PRIMARY KEY, invoice_number TEXT, supplier_id UUID, status TEXT NOT NULL,
			total_amount NUMERIC(12,2), paid_amount NUMERIC(12,2), remaining_amount NUMERIC(12,2),
			purchase_date TIMESTAMPTZ, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, reversed_at TIMESTAMPTZ
		);
		CREATE TABLE suppliers (id UUID PRIMARY KEY, current_balance NUMERIC(12,2), updated_at TIMESTAMPTZ);
		CREATE TABLE payments (id UUID PRIMARY KEY, purchase_id UUID, customer_id UUID, supplier_id UUID, sale_id UUID,
			amount NUMERIC(12,2) NOT NULL DEFAULT 0, status TEXT, payment_status TEXT, payment_date TIMESTAMPTZ, created_at TIMESTAMPTZ);
		CREATE TABLE payment_allocation_batches (payment_id UUID PRIMARY KEY, owner_type TEXT NOT NULL, owner_id UUID NOT NULL, sale_id UUID, tracked_at TIMESTAMPTZ DEFAULT NOW());
		CREATE TABLE payment_debt_allocations (id UUID PRIMARY KEY, payment_id UUID NOT NULL, debt_id UUID NOT NULL, amount NUMERIC(12,2) NOT NULL, created_at TIMESTAMPTZ DEFAULT NOW());
		CREATE TABLE supplier_ledger (
			id UUID PRIMARY KEY, supplier_id UUID, transaction_type TEXT, type TEXT, amount NUMERIC(12,2), balance NUMERIC(12,2),
			reference_id UUID, reference_type TEXT, created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE TABLE supplier_returns (id UUID PRIMARY KEY, purchase_id UUID);
		CREATE TABLE supplier_return_items (id UUID PRIMARY KEY, supplier_return_id UUID);
		CREATE TABLE inventory_movements (id UUID PRIMARY KEY, reference_id UUID, reference_type TEXT);
		CREATE TABLE item_history (id UUID PRIMARY KEY, reference_id UUID, reference_type TEXT);
		CREATE TABLE purchase_items (id UUID PRIMARY KEY, purchase_id UUID);
		CREATE TABLE returns (id UUID PRIMARY KEY, purchase_id UUID, return_date DATE NOT NULL);
	`); err != nil {
		t.Fatalf("create PostgreSQL purchase cleanup schema: %v", err)
	}
	purchaseID, outsidePurchaseID, supplierID, paymentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO suppliers (id,current_balance,updated_at) VALUES ($1,0,NOW())`, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO purchases (id, invoice_number, supplier_id, status, total_amount, paid_amount, remaining_amount, purchase_date, created_at, updated_at)
		VALUES ($1, 'P-LOCAL-DATE', $3, 'pending', 100, 0, 100, '2026-05-01T02:30:00Z', '2026-05-01T02:30:00Z', NOW()),
		       ($2, 'P-OUTSIDE-DATE', $3, 'pending', 200, 0, 200, '2026-05-01T04:30:00Z', '2026-05-01T04:30:00Z', NOW())`, purchaseID, outsidePurchaseID, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payments (id,purchase_id,supplier_id,amount,status,payment_status,payment_date,created_at) VALUES ($1,$2,$3,0,'pending','pending',NOW(),NOW())`, paymentID, purchaseID, supplierID); err != nil {
		t.Fatal(err)
	}
	preview, err := loadHistoricalCleanupPreview(context.Background(), db, false, "purchases", "2026-04-30", "2026-04-30")
	if err != nil {
		t.Fatalf("preview PostgreSQL purchase cleanup: %v", err)
	}
	if preview.CandidateCount != 1 || len(preview.CandidateIDs) != 1 || preview.CandidateIDs[0] != purchaseID.String() || preview.Timezone != "America/New_York" {
		t.Fatalf("PostgreSQL store-date preview=%+v; want only the April 30 store-date purchase", preview)
	}
	dateReturnID := uuid.New()
	if _, err := db.Exec(`INSERT INTO returns (id,return_date) VALUES ($1,'2026-04-30')`, dateReturnID); err != nil {
		t.Fatal(err)
	}
	datePreview, err := loadHistoricalCleanupPreview(context.Background(), db, false, "returns", "2026-04-30", "2026-04-30")
	if err != nil || datePreview.CandidateCount != 1 || datePreview.CandidateIDs[0] != dateReturnID.String() {
		t.Fatalf("PostgreSQL DATE-column preview=%+v err=%v; want the store date without a timezone conversion", datePreview, err)
	}

	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(fmt.Sprintf(`{"type":"purchases","start_date":"2026-04-30","end_date":"2026-04-30","candidate_ids":[%q]}`, purchaseID.String())))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("PostgreSQL historical cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 1 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 {
		t.Fatalf("PostgreSQL purchase cleanup result=%+v; want deletion after reversing linked pending payment", envelope.Data)
	}
	var purchaseCount, paymentCount int
	if err := db.Get(&purchaseCount, `SELECT COUNT(*) FROM purchases WHERE id=$1`, purchaseID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&paymentCount, `SELECT COUNT(*) FROM payments WHERE id=$1`, paymentID); err != nil {
		t.Fatal(err)
	}
	if purchaseCount != 0 || paymentCount != 0 {
		t.Fatalf("PostgreSQL purchase cleanup left source/payment rows: purchases=%d payments=%d", purchaseCount, paymentCount)
	}
}

func TestPostgresHistoricalCleanupDeletesOnlyUnreferencedDirectoriesAtomically(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("UTC"); err != nil {
		t.Fatal(err)
	}
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE customers (id UUID PRIMARY KEY, created_at TIMESTAMPTZ NOT NULL);
		CREATE TABLE suppliers (id UUID PRIMARY KEY, created_at TIMESTAMPTZ NOT NULL);
		CREATE TABLE sales (id UUID PRIMARY KEY, customer_id UUID);
		CREATE TABLE purchases (id UUID PRIMARY KEY, supplier_id UUID);
	`); err != nil {
		t.Fatalf("create PostgreSQL directory cleanup schema: %v", err)
	}
	freeCustomerID, linkedCustomerID := uuid.New(), uuid.New()
	freeSupplierID, linkedSupplierID := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{freeCustomerID, linkedCustomerID} {
		if _, err := db.Exec(`INSERT INTO customers VALUES ($1, '2020-01-01T00:00:00Z')`, id); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []uuid.UUID{freeSupplierID, linkedSupplierID} {
		if _, err := db.Exec(`INSERT INTO suppliers VALUES ($1, '2020-01-01T00:00:00Z')`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO sales VALUES ($1, $2)`, uuid.New(), linkedCustomerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchases VALUES ($1, $2)`, uuid.New(), linkedSupplierID); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		key string
		ids []string
	}{
		{key: "customers", ids: []string{freeCustomerID.String(), linkedCustomerID.String()}},
		{key: "suppliers", ids: []string{freeSupplierID.String(), linkedSupplierID.String()}},
	} {
		spec := historicalCleanupSpecs[tc.key]
		dateExpression, found, err := historicalCleanupDateColumn(context.Background(), db, false, spec)
		if err != nil || !found {
			t.Fatalf("resolve %s store-date column: found=%v err=%v", tc.key, found, err)
		}
		deleted, blocked, err := deleteUnreferencedHistoricalDirectories(context.Background(), db, false, spec, tc.ids, dateExpression, "", "2020-01-01", "2020-01-02")
		if err != nil {
			t.Fatalf("cleanup PostgreSQL %s: %v", tc.key, err)
		}
		if deleted != 1 || blocked != 1 {
			t.Fatalf("PostgreSQL %s cleanup deleted=%d blocked=%d; want 1/1", tc.key, deleted, blocked)
		}
	}
	var customersLeft, suppliersLeft, linkedCustomerLeft, linkedSupplierLeft int
	for _, check := range []struct {
		query string
		id    uuid.UUID
		dest  *int
	}{
		{`SELECT COUNT(*) FROM customers`, uuid.Nil, &customersLeft},
		{`SELECT COUNT(*) FROM suppliers`, uuid.Nil, &suppliersLeft},
		{`SELECT COUNT(*) FROM customers WHERE id=$1`, linkedCustomerID, &linkedCustomerLeft},
		{`SELECT COUNT(*) FROM suppliers WHERE id=$1`, linkedSupplierID, &linkedSupplierLeft},
	} {
		var err error
		if check.id == uuid.Nil {
			err = db.Get(check.dest, check.query)
		} else {
			err = db.Get(check.dest, check.query, check.id)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if customersLeft != 1 || suppliersLeft != 1 || linkedCustomerLeft != 1 || linkedSupplierLeft != 1 {
		t.Fatalf("remaining customers=%d suppliers=%d linked customer=%d linked supplier=%d; expected only referenced directory rows", customersLeft, suppliersLeft, linkedCustomerLeft, linkedSupplierLeft)
	}
}

func TestPostgresHistoricalHeldSaleCleanupIsScopedToCurrentUser(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`CREATE TABLE held_sales (id UUID PRIMARY KEY, user_id UUID NOT NULL, items JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL)`); err != nil {
		t.Fatalf("create PostgreSQL held sales schema: %v", err)
	}
	userID, otherUserID := uuid.New(), uuid.New()
	ownedID, foreignID := uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO held_sales (id,user_id,items,created_at) VALUES ($1,$2,'[]','2026-04-30T12:00:00Z'),($3,$4,'[]','2026-04-30T12:00:00Z')`, ownedID, userID, foreignID, otherUserID); err != nil {
		t.Fatal(err)
	}
	preview, err := loadHistoricalCleanupPreviewForUser(context.Background(), db, false, "held_sales", "2026-04-30", "2026-04-30", userID)
	if err != nil || preview.CandidateCount != 1 || len(preview.CandidateIDs) != 1 || preview.CandidateIDs[0] != ownedID.String() {
		t.Fatalf("PostgreSQL held sale preview=%+v err=%v; want only current user's record", preview, err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(fmt.Sprintf(`{"type":"held_sales","start_date":"2026-04-30","end_date":"2026-04-30","candidate_ids":[%q,%q]}`, ownedID, foreignID)))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", userID)
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("PostgreSQL held sale cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 1 || envelope.Data.Blocked != 1 || envelope.Data.Failed != 0 {
		t.Fatalf("PostgreSQL held sale cleanup result=%+v; want one deletion and one blocked foreign row", envelope.Data)
	}
	var ownRemaining, otherRemaining int
	if err := db.Get(&ownRemaining, `SELECT COUNT(*) FROM held_sales WHERE user_id=$1`, userID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&otherRemaining, `SELECT COUNT(*) FROM held_sales WHERE user_id=$1`, otherUserID); err != nil {
		t.Fatal(err)
	}
	if ownRemaining != 0 || otherRemaining != 1 {
		t.Fatalf("PostgreSQL remaining held sales: current user=%d other user=%d", ownRemaining, otherRemaining)
	}
}

func TestPostgresHistoricalInventoryCleanupReversesStockAndDeletesHistory(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("UTC"); err != nil {
		t.Fatal(err)
	}
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE inventory_items (id UUID PRIMARY KEY, product_id UUID NOT NULL, status TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL);
		CREATE TABLE inventory (product_id UUID PRIMARY KEY, quantity INTEGER NOT NULL, reserved_quantity INTEGER NOT NULL DEFAULT 0, updated_at TIMESTAMPTZ);
		CREATE TABLE inventory_movements (id UUID PRIMARY KEY, item_id UUID);
		CREATE TABLE reservations (id UUID PRIMARY KEY, item_id UUID);
		CREATE TABLE audit_logs (id UUID PRIMARY KEY, user_id UUID, action TEXT, entity_type TEXT, entity_id UUID, new_values JSONB, created_at TIMESTAMPTZ);
		CREATE TABLE sale_items (sale_id UUID, inventory_item_id UUID);
		CREATE TABLE return_items (inventory_item_id UUID);
		CREATE TABLE supplier_return_items (inventory_item_id UUID);
		CREATE TABLE acquisition_items (inventory_item_id UUID);
		CREATE TABLE trade_ins (inventory_item_id UUID);
		CREATE TABLE item_repair_costs (inventory_item_id UUID);
		CREATE TABLE item_history (inventory_item_id UUID);
		CREATE TABLE barcodes (inventory_item_id UUID);
		CREATE TABLE inspections (inventory_item_id UUID);
		CREATE TABLE item_specification_values (inventory_item_id UUID);
	`); err != nil {
		t.Fatalf("create PostgreSQL inventory cleanup schema: %v", err)
	}
	eligibleID, protectedID, productID := uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO inventory_items (id,product_id,status,created_at) VALUES ($1,$3,'AVAILABLE','2026-04-30T12:00:00Z'),($2,$3,'AVAILABLE','2026-04-30T12:00:00Z')`, eligibleID, protectedID, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (product_id,quantity,reserved_quantity) VALUES ($1,2,0)`, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id) VALUES ($1,$2)`, uuid.New(), protectedID); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(fmt.Sprintf(`{"type":"inventory_items","start_date":"2026-04-30","end_date":"2026-04-30","candidate_ids":[%q,%q]}`, eligibleID, protectedID)))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("PostgreSQL inventory cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 2 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 {
		t.Fatalf("PostgreSQL inventory cleanup result=%+v; want both inventory rows hard-deleted with linked effects reversed", envelope.Data)
	}
	var itemCount, quantity, movementCount, auditCount int
	if err := db.Get(&itemCount, `SELECT COUNT(*) FROM inventory_items`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&quantity, `SELECT quantity FROM inventory WHERE product_id=$1`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&movementCount, `SELECT COUNT(*) FROM inventory_movements WHERE item_id=$1`, protectedID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&auditCount, `SELECT COUNT(*) FROM audit_logs WHERE entity_id=$1`, eligibleID); err != nil {
		t.Fatal(err)
	}
	if itemCount != 0 || quantity != 0 || movementCount != 0 || auditCount != 1 {
		t.Fatalf("PostgreSQL inventory cleanup: items=%d stock=%d protected movements=%d deletion audits=%d", itemCount, quantity, movementCount, auditCount)
	}
}

func TestPostgresHistoricalInspectionCleanupReversesAndDeletesAllStates(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE inventory_items (id UUID PRIMARY KEY, serial_number TEXT, status TEXT, sold_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
		CREATE TABLE acquisition_items (id UUID PRIMARY KEY, serial_number TEXT);
		CREATE TABLE inspection_items (id UUID PRIMARY KEY, inspection_id UUID);
		CREATE TABLE inspection_workflow_snapshots (
			inspection_id UUID PRIMARY KEY, inventory_item_id UUID, inventory_status_before TEXT,
			acquisition_item_id UUID, acquisition_inventory_item_id_before UUID, acquisition_inspection_id_before UUID,
			acquisition_inspection_status_before TEXT, acquisition_item_status_before TEXT, acquisition_id UUID,
			acquisition_status_before TEXT, captured_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE TABLE inspections (
			id UUID PRIMARY KEY, product_id UUID, inventory_item_id UUID, inspection_date DATE NOT NULL,
			inspector_id UUID NOT NULL, result TEXT NOT NULL, condition TEXT, grade TEXT, notes TEXT,
			images JSONB NOT NULL DEFAULT '[]'::jsonb, test_results JSONB NOT NULL DEFAULT '{}'::jsonb,
			acquisition_item_id UUID, created_at TIMESTAMPTZ NOT NULL, updated_at TIMESTAMPTZ NOT NULL
		);
	`); err != nil {
		t.Fatalf("create PostgreSQL inspection cleanup schema: %v", err)
	}
	standaloneID, linkedID, completedID := uuid.New(), uuid.New(), uuid.New()
	itemID, inspectorID := uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO inventory_items (id,serial_number,status,updated_at) VALUES ($1,'SER-1','INSPECTION',NOW())`, itemID); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		id     uuid.UUID
		result string
		itemID any
	}{{standaloneID, "pending", nil}, {linkedID, "pending", itemID}, {completedID, "passed", nil}} {
		if _, err := db.Exec(`INSERT INTO inspections (id,inspection_date,inspector_id,result,inventory_item_id,created_at,updated_at) VALUES ($1,'2026-04-30',$2,$3,$4,'2026-04-30T12:00:00Z','2026-04-30T12:00:00Z')`, row.id, inspectorID, row.result, row.itemID); err != nil {
			t.Fatal(err)
		}
	}
	preview, err := loadHistoricalCleanupPreview(context.Background(), db, false, "inspections", "2026-04-30", "2026-04-30")
	if err != nil || preview.CandidateCount != 3 {
		t.Fatalf("PostgreSQL inspection cleanup preview=%+v err=%v; want all three removable states", preview, err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(fmt.Sprintf(`{"type":"inspections","start_date":"2026-04-30","end_date":"2026-04-30","candidate_ids":[%q,%q,%q]}`, standaloneID, linkedID, completedID)))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("PostgreSQL inspection cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 3 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 {
		t.Fatalf("PostgreSQL inspection cleanup result=%+v; want all three inspections deleted", envelope.Data)
	}
	var remaining int
	if err := db.Get(&remaining, `SELECT COUNT(*) FROM inspections`); err != nil {
		t.Fatal(err)
	}
	var itemStatus string
	if err := db.Get(&itemStatus, `SELECT status FROM inventory_items WHERE id=$1`, itemID); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 || itemStatus != "AVAILABLE" {
		t.Fatalf("PostgreSQL inspection cleanup left inspections=%d item_status=%s; want 0 and AVAILABLE", remaining, itemStatus)
	}
}

func TestPostgresSmartDeleteSaleReversesStockCustomerLedgerAndReports(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE sales (id UUID PRIMARY KEY, invoice_number TEXT, customer_id UUID, user_id UUID, created_at TIMESTAMPTZ, status TEXT, payment_status TEXT, total_amount NUMERIC(12,2), paid_amount NUMERIC(12,2) DEFAULT 0, updated_at TIMESTAMPTZ);
		CREATE TABLE sale_items (id UUID PRIMARY KEY, sale_id UUID, product_id UUID, quantity INTEGER);
		CREATE TABLE inventory_items (id UUID PRIMARY KEY, product_id UUID, status TEXT, sold_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
		CREATE TABLE inventory_movements (id UUID PRIMARY KEY, item_id UUID, product_id UUID, movement_type TEXT, quantity INTEGER, reference_type TEXT, reference_id UUID);
		CREATE TABLE inventory (id UUID PRIMARY KEY, product_id UUID UNIQUE, quantity INTEGER, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
		CREATE TABLE returns (id UUID PRIMARY KEY, sale_id UUID);
		CREATE TABLE accounting_returns (id UUID PRIMARY KEY, sale_id UUID);
		CREATE TABLE payment_transactions (id UUID PRIMARY KEY, sale_id UUID, created_at TIMESTAMPTZ DEFAULT NOW());
		CREATE TABLE debts (id UUID PRIMARY KEY, sale_id UUID, paid_amount NUMERIC(12,2));
		CREATE TABLE payments (id UUID PRIMARY KEY, sale_id UUID, customer_id UUID);
		CREATE TABLE customer_ledger (id UUID PRIMARY KEY, customer_id UUID, type TEXT, amount NUMERIC(12,2), reference_id UUID, reference_type TEXT);
		CREATE TABLE customers (id UUID PRIMARY KEY, current_balance NUMERIC(12,2), updated_at TIMESTAMPTZ);
		CREATE TABLE sale_payment_allocations (id UUID PRIMARY KEY, sale_id UUID);
		CREATE TABLE item_history (id UUID PRIMARY KEY, reference_id UUID, reference_type TEXT);
		CREATE TABLE ledger_entries (id UUID PRIMARY KEY, reference_id UUID, reference_type TEXT);
		CREATE TABLE audit_logs (id UUID PRIMARY KEY, user_id UUID, action TEXT, entity_type TEXT, entity_id UUID, new_values JSONB, created_at TIMESTAMPTZ);
		CREATE TABLE pos_shifts (id UUID PRIMARY KEY, user_id UUID, opened_at TIMESTAMPTZ, closed_at TIMESTAMPTZ, sales_total NUMERIC(12,2), sale_count INTEGER);
	`); err != nil {
		t.Fatalf("create PostgreSQL sale deletion schema: %v", err)
	}
	saleID, productID, itemID, customerID, userID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO sales (id,invoice_number,customer_id,user_id,created_at,status,payment_status,total_amount,paid_amount) VALUES ($1,'INV-PG-DEL',$2,$3,'2026-04-30T10:00:00Z','completed','unpaid',120,0)`, saleID, customerID, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id,invoice_number,user_id,created_at,status,payment_status,total_amount,paid_amount) VALUES ($1,'INV-PG-KEEP',$2,'2026-04-30T11:00:00Z','completed','unpaid',45,0)`, uuid.New(), userID); err != nil {
		t.Fatal(err)
	}
	shiftID, unrelatedShiftID := uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO pos_shifts (id,user_id,opened_at,closed_at,sales_total,sale_count) VALUES ($1,$3,'2026-04-30T08:00:00Z','2026-04-30T12:00:00Z',165,2),($2,$3,'2026-04-30T12:00:00Z','2026-04-30T16:00:00Z',999,9)`, shiftID, unrelatedShiftID, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id,sale_id,product_id,quantity) VALUES ($1,$2,$3,1)`, uuid.New(), saleID, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id,product_id,status,sold_at,updated_at) VALUES ($1,$2,'SOLD',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, itemID, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,reference_type,reference_id) VALUES ($1,$2,$3,'SALE',-1,'sale',$4)`, uuid.New(), itemID, productID, saleID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES ($1,$2,2,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, uuid.New(), productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customers (id,current_balance,updated_at) VALUES ($1,120,CURRENT_TIMESTAMP)`, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_ledger (id,customer_id,type,amount,reference_id,reference_type) VALUES ($1,$2,'debit',120,$3,'sale')`, uuid.New(), customerID, saleID); err != nil {
		t.Fatal(err)
	}
	result, err := sales.NewSmartDeleteService(db).SmartDelete(context.Background(), saleID, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deleted" || !result.CanProceed {
		t.Fatalf("PostgreSQL sale delete result=%+v; want deleted", result)
	}
	var salesLeft, remainingSales, itemsLeft, movementsLeft, ledgerLeft, deleteAudits int
	var stock int
	var balance float64
	for query, destination := range map[string]*int{
		`SELECT COUNT(*) FROM sales WHERE id=$1`:                                 &salesLeft,
		`SELECT COUNT(*) FROM sale_items WHERE sale_id=$1`:                       &itemsLeft,
		`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=$1`:         &movementsLeft,
		`SELECT COUNT(*) FROM customer_ledger WHERE reference_id=$1`:             &ledgerLeft,
		`SELECT COUNT(*) FROM audit_logs WHERE entity_id=$1 AND action='DELETE'`: &deleteAudits,
	} {
		if err := db.Get(destination, query, saleID); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Get(&remainingSales, `SELECT COUNT(*) FROM sales`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&stock, `SELECT quantity FROM inventory WHERE product_id=$1`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&balance, `SELECT current_balance FROM customers WHERE id=$1`, customerID); err != nil {
		t.Fatal(err)
	}
	var itemStatus string
	if err := db.Get(&itemStatus, `SELECT status FROM inventory_items WHERE id=$1`, itemID); err != nil {
		t.Fatal(err)
	}
	var shiftTotal, unrelatedShiftTotal float64
	var shiftCount, unrelatedShiftCount int
	if err := db.Get(&shiftTotal, `SELECT sales_total FROM pos_shifts WHERE id=$1`, shiftID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&shiftCount, `SELECT sale_count FROM pos_shifts WHERE id=$1`, shiftID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&unrelatedShiftTotal, `SELECT sales_total FROM pos_shifts WHERE id=$1`, unrelatedShiftID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&unrelatedShiftCount, `SELECT sale_count FROM pos_shifts WHERE id=$1`, unrelatedShiftID); err != nil {
		t.Fatal(err)
	}
	if salesLeft != 0 || remainingSales != 1 || itemsLeft != 0 || movementsLeft != 0 || ledgerLeft != 0 || deleteAudits != 1 || stock != 3 || balance != 0 || itemStatus != "AVAILABLE" || shiftTotal != 45 || shiftCount != 1 || unrelatedShiftTotal != 999 || unrelatedShiftCount != 9 {
		t.Fatalf("PostgreSQL sale reversal: target=%d remaining_sales=%d items=%d movements=%d ledger=%d audit=%d stock=%d balance=%v item=%s shift=%v/%d unrelated_shift=%v/%d", salesLeft, remainingSales, itemsLeft, movementsLeft, ledgerLeft, deleteAudits, stock, balance, itemStatus, shiftTotal, shiftCount, unrelatedShiftTotal, unrelatedShiftCount)
	}
}

func TestPostgresSmartDeletePurchaseReversesStockAndCompletedPayment(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE purchases (id UUID PRIMARY KEY, invoice_number TEXT, supplier_id UUID, status TEXT, total_amount NUMERIC(12,2), paid_amount NUMERIC(12,2) DEFAULT 0, reversed_at TIMESTAMPTZ);
		CREATE TABLE suppliers (id UUID PRIMARY KEY, current_balance NUMERIC(12,2), updated_at TIMESTAMPTZ);
		CREATE TABLE products (id UUID PRIMARY KEY, name TEXT);
		CREATE TABLE payments (id UUID PRIMARY KEY, purchase_id UUID, customer_id UUID, supplier_id UUID, sale_id UUID, amount NUMERIC(12,2), status TEXT, payment_status TEXT, payment_date TIMESTAMPTZ, created_at TIMESTAMPTZ DEFAULT NOW());
		CREATE TABLE payment_allocation_batches (payment_id UUID PRIMARY KEY, owner_type TEXT NOT NULL, owner_id UUID NOT NULL, sale_id UUID, tracked_at TIMESTAMPTZ DEFAULT NOW());
		CREATE TABLE payment_debt_allocations (id UUID PRIMARY KEY, payment_id UUID NOT NULL, debt_id UUID NOT NULL, amount NUMERIC(12,2) NOT NULL, created_at TIMESTAMPTZ DEFAULT NOW());
		CREATE TABLE supplier_debts (id UUID PRIMARY KEY, supplier_id UUID, reference_id UUID, reference_type TEXT, amount NUMERIC(12,2), paid_amount NUMERIC(12,2), is_paid BOOLEAN, due_date DATE, created_at TIMESTAMPTZ);
		CREATE TABLE supplier_ledger (id UUID PRIMARY KEY, supplier_id UUID, type TEXT, transaction_type TEXT, amount NUMERIC(12,2), balance NUMERIC(12,2), reference_id UUID, reference_type TEXT, created_at TIMESTAMPTZ DEFAULT NOW());
		CREATE TABLE purchase_items (id UUID PRIMARY KEY, purchase_id UUID, product_id UUID, quantity INTEGER);
		CREATE TABLE supplier_returns (id UUID PRIMARY KEY, purchase_id UUID);
		CREATE TABLE supplier_return_items (id UUID PRIMARY KEY, supplier_return_id UUID);
		CREATE TABLE inventory_items (id UUID PRIMARY KEY, item_code TEXT, product_id UUID, barcode TEXT, status TEXT, condition TEXT);
		CREATE TABLE inventory (id UUID PRIMARY KEY, product_id UUID UNIQUE, quantity INTEGER, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
		CREATE TABLE inventory_movements (id UUID PRIMARY KEY, item_id UUID, product_id UUID, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id UUID, reason TEXT, created_by UUID, created_at TIMESTAMPTZ DEFAULT NOW());
		CREATE TABLE item_history (id UUID PRIMARY KEY, inventory_item_id UUID, reference_type TEXT, reference_id UUID);
		CREATE TABLE returns (id UUID PRIMARY KEY, purchase_id UUID);
		CREATE TABLE barcodes (id UUID PRIMARY KEY, inventory_item_id UUID);
		CREATE TABLE reservations (id UUID PRIMARY KEY, item_id UUID);
		CREATE TABLE acquisition_items (id UUID PRIMARY KEY, inventory_item_id UUID, item_status TEXT, updated_at TIMESTAMPTZ);
		CREATE TABLE inspection_items (id UUID PRIMARY KEY, item_id UUID);
		CREATE TABLE item_repair_costs (id UUID PRIMARY KEY, inventory_item_id UUID);
		CREATE TABLE audit_logs (id UUID PRIMARY KEY, user_id UUID, action TEXT, entity_type TEXT, entity_id UUID, new_values JSONB, created_at TIMESTAMPTZ);
	`); err != nil {
		t.Fatalf("create PostgreSQL purchase deletion schema: %v", err)
	}
	purchaseID, paidPurchaseID, supplierID, productID, inventoryItemID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	itemCode := "ITM-" + purchaseID.String()[:8] + "-001"
	if _, err := db.Exec(`INSERT INTO purchases (id,invoice_number,supplier_id,status,total_amount,paid_amount) VALUES ($1,'PUR-PG-DEL',$3,'received',100,0),($2,'PUR-PG-PAID',$3,'pending',25,25)`, purchaseID, paidPurchaseID, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO suppliers (id,current_balance,updated_at) VALUES ($1,100,CURRENT_TIMESTAMP)`, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id,name) VALUES ($1,'PG test product')`, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchase_items (id,purchase_id,product_id,quantity) VALUES ($1,$2,$3,1)`, uuid.New(), purchaseID, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id,item_code,product_id,barcode,status,condition) VALUES ($1,$2,$3,'BC-PG-1','AVAILABLE','NEW')`, inventoryItemID, itemCode, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES ($1,$2,1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, uuid.New(), productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,quantity,reference_type,reference_id) VALUES ($1,$2,$3,'PURCHASE',1,'purchase',$4)`, uuid.New(), inventoryItemID, productID, purchaseID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_ledger (id,supplier_id,type,transaction_type,amount,balance,reference_id,reference_type) VALUES ($1,$2,'debit','PURCHASE',100,100,$3,'purchase'),($4,$2,'debit','PURCHASE',25,125,$5,'purchase')`, uuid.New(), supplierID, purchaseID, uuid.New(), paidPurchaseID); err != nil {
		t.Fatal(err)
	}
	paymentID := uuid.New()
	if _, err := db.Exec(`INSERT INTO payments (id,purchase_id,supplier_id,amount,status,payment_status,payment_date,created_at) VALUES ($1,$2,$3,25,'completed','completed',NOW(),NOW())`, paymentID, paidPurchaseID, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_ledger (id,supplier_id,type,transaction_type,amount,balance,reference_id,reference_type) VALUES ($1,$2,'credit','PAYMENT',25,100,$3,'payment')`, uuid.New(), supplierID, paymentID); err != nil {
		t.Fatal(err)
	}
	service := purchases.NewSmartDeleteService(db)
	deletedPaid, err := service.SmartDelete(context.Background(), paidPurchaseID, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if deletedPaid.Action != "deleted" || !deletedPaid.CanProceed {
		t.Fatalf("PostgreSQL paid purchase deletion=%+v; want deleted after payment reversal", deletedPaid)
	}
	deleted, err := service.SmartDelete(context.Background(), purchaseID, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if deleted.Action != "deleted" || !deleted.CanProceed {
		t.Fatalf("PostgreSQL received purchase deletion=%+v; want deleted", deleted)
	}
	var receivedLeft, paidLeft, paymentLeft, purchaseItemsLeft, stock, auditCount int
	var supplierBalance float64
	for _, check := range []struct {
		query string
		id    uuid.UUID
		into  *int
	}{
		{`SELECT COUNT(*) FROM purchases WHERE id=$1`, purchaseID, &receivedLeft},
		{`SELECT COUNT(*) FROM purchases WHERE id=$1`, paidPurchaseID, &paidLeft},
		{`SELECT COUNT(*) FROM payments WHERE id=$1`, paymentID, &paymentLeft},
		{`SELECT COUNT(*) FROM purchase_items WHERE purchase_id=$1`, purchaseID, &purchaseItemsLeft},
		{`SELECT COUNT(*) FROM audit_logs WHERE entity_id=$1 AND action='DELETE'`, purchaseID, &auditCount},
	} {
		if err := db.Get(check.into, check.query, check.id); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Get(&stock, `SELECT quantity FROM inventory WHERE product_id=$1`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&supplierBalance, `SELECT current_balance FROM suppliers WHERE id=$1`, supplierID); err != nil {
		t.Fatal(err)
	}
	if receivedLeft != 0 || paidLeft != 0 || paymentLeft != 0 || purchaseItemsLeft != 0 || stock != 0 || auditCount != 1 || supplierBalance != 0 {
		t.Fatalf("PostgreSQL purchase deletion: received=%d paid=%d payment=%d purchase_items=%d stock=%d audit=%d supplier_balance=%v", receivedLeft, paidLeft, paymentLeft, purchaseItemsLeft, stock, auditCount, supplierBalance)
	}
}

func TestPostgresDeleteManualDebtReconcilesCustomerLedger(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE customers (id UUID PRIMARY KEY, current_balance NUMERIC(12,2), updated_at TIMESTAMPTZ);
		CREATE TABLE debts (id UUID PRIMARY KEY, customer_id UUID, sale_id UUID, amount NUMERIC(12,2), paid_amount NUMERIC(12,2), remaining_amount NUMERIC(12,2), status TEXT, notes TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
		CREATE TABLE customer_ledger (id UUID PRIMARY KEY, customer_id UUID, debt_id UUID, reference_id UUID, reference_type TEXT, type TEXT, amount NUMERIC(12,2), balance NUMERIC(12,2), created_at TIMESTAMPTZ);
		CREATE TABLE customer_debts (id UUID PRIMARY KEY, customer_id UUID, amount NUMERIC(12,2));
	`); err != nil {
		t.Fatalf("create PostgreSQL debt deletion schema: %v", err)
	}
	debtID, customerID := uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO customers (id,current_balance) VALUES ($1,40)`, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO debts (id,customer_id,amount,paid_amount,remaining_amount,status) VALUES ($1,$2,40,0,40,'pending')`, debtID, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_ledger (id,customer_id,debt_id,reference_id,reference_type,type,amount,balance,created_at) VALUES ($1,$2,$3,$3,'debt','debit',40,40,NOW())`, uuid.New(), customerID, debtID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_debts (id,customer_id,amount) VALUES ($1,$2,40)`, debtID, customerID); err != nil {
		t.Fatal(err)
	}
	if err := debts.NewHandler(db).DeleteDebtByID(context.Background(), debtID); err != nil {
		t.Fatalf("delete PostgreSQL manual debt: %v", err)
	}
	var debtCount, ledgerCount, mirrorCount int
	var balance float64
	for query, into := range map[string]*int{
		`SELECT COUNT(*) FROM debts WHERE id=$1`:                &debtCount,
		`SELECT COUNT(*) FROM customer_ledger WHERE debt_id=$1`: &ledgerCount,
		`SELECT COUNT(*) FROM customer_debts WHERE id=$1`:       &mirrorCount,
	} {
		if err := db.Get(into, query, debtID); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Get(&balance, `SELECT current_balance FROM customers WHERE id=$1`, customerID); err != nil {
		t.Fatal(err)
	}
	if debtCount != 0 || ledgerCount != 0 || mirrorCount != 0 || balance != 0 {
		t.Fatalf("PostgreSQL manual debt deletion: debt=%d ledger=%d mirror=%d customer_balance=%v; want all zero", debtCount, ledgerCount, mirrorCount, balance)
	}
}

func TestPostgresDeleteCompletedReturnReversesAccountingAndStockEffects(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE returns (id UUID PRIMARY KEY, return_number TEXT, reference_number TEXT, sale_id UUID, purchase_id UUID, customer_id UUID, total_refund_amount NUMERIC(12,2), status TEXT, return_date TIMESTAMPTZ, refund_date TIMESTAMPTZ, refund_method TEXT, debt_id UUID, debt_adjustment NUMERIC(12,2), customer_credit NUMERIC(12,2), created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
		CREATE TABLE return_items (id UUID PRIMARY KEY, return_id UUID, sale_item_id UUID, product_id UUID, inventory_item_id UUID, serial_number TEXT, barcode TEXT, quantity_returned INTEGER, original_quantity INTEGER, unit_price NUMERIC(12,2), total_refund_amount NUMERIC(12,2), original_cost NUMERIC(12,2), resolution TEXT, inventory_status TEXT, created_at TIMESTAMPTZ);
		CREATE TABLE return_effects (id UUID PRIMARY KEY, sale_id UUID, purchase_id UUID, customer_id UUID, total_refund_amount NUMERIC(12,2), status TEXT, return_date TIMESTAMPTZ, refund_date TIMESTAMPTZ, refund_method TEXT, debt_id UUID, debt_adjustment NUMERIC(12,2), customer_credit NUMERIC(12,2), is_reversal BOOLEAN, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
		CREATE TABLE return_effect_items (id UUID PRIMARY KEY, return_effect_id UUID, sale_item_id UUID, product_id UUID, inventory_item_id UUID, serial_number TEXT, barcode TEXT, quantity_returned INTEGER, original_quantity INTEGER, unit_price NUMERIC(12,2), total_refund_amount NUMERIC(12,2), original_cost NUMERIC(12,2), resolution TEXT, inventory_status TEXT, created_at TIMESTAMPTZ);
		CREATE TABLE sale_items (id UUID PRIMARY KEY, sale_id UUID, product_id UUID, inventory_item_id UUID, unit_cost NUMERIC(12,2), created_at TIMESTAMPTZ);
		CREATE TABLE products (id UUID PRIMARY KEY, cost_price NUMERIC(12,2));
		CREATE TABLE inventory_items (id UUID PRIMARY KEY, product_id UUID, status TEXT, sold_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
		CREATE TABLE inventory (product_id UUID PRIMARY KEY, quantity INTEGER, updated_at TIMESTAMPTZ);
		CREATE TABLE inventory_movements (id UUID PRIMARY KEY, item_id UUID, product_id UUID, movement_type TEXT, reference_id UUID, reference_type TEXT, quantity INTEGER);
		CREATE TABLE customers (id UUID PRIMARY KEY, current_balance NUMERIC(12,2), updated_at TIMESTAMPTZ);
		CREATE TABLE customer_ledger (id UUID PRIMARY KEY, customer_id UUID, balance NUMERIC(12,2), created_at TIMESTAMPTZ, reference_id UUID, reference_type TEXT, type TEXT, amount NUMERIC(12,2));
		CREATE TABLE ledger_entries (id UUID PRIMARY KEY, reference_id UUID, reference_type TEXT, amount NUMERIC(12,2));
		CREATE TABLE audit_logs (id UUID PRIMARY KEY, user_id UUID, action TEXT, entity_type TEXT, entity_id UUID, new_values JSONB, created_at TIMESTAMPTZ);
	`); err != nil {
		t.Fatalf("create PostgreSQL return deletion schema: %v", err)
	}
	returnID, itemID, saleItemID, productID, customerID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO inventory (product_id,quantity) VALUES ($1,1)`, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customers (id,current_balance) VALUES ($1,-18)`, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO returns (id,return_number,sale_id,customer_id,total_refund_amount,status,return_date,refund_method,created_at,updated_at) VALUES ($1,'RET-PG-DEL',$2,$3,18,'COMPLETED',NOW(),'cash',NOW(),NOW())`, returnID, uuid.New(), customerID); err != nil {
		t.Fatal(err)
	}
	inventoryItemID := uuid.New()
	if _, err := db.Exec(`INSERT INTO inventory_items (id,product_id,status,sold_at) VALUES ($1,$2,'AVAILABLE',NOW())`, inventoryItemID, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id,product_id,inventory_item_id,unit_cost,created_at) VALUES ($1,$2,$3,7,NOW())`, saleItemID, productID, inventoryItemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO return_items (id,return_id,sale_item_id,product_id,inventory_item_id,quantity_returned,original_quantity,unit_price,total_refund_amount,original_cost,created_at) VALUES ($1,$2,$3,$4,$5,1,1,18,18,7,NOW())`, itemID, returnID, saleItemID, productID, inventoryItemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,movement_type,reference_id,reference_type,quantity) VALUES ($1,$2,$3,'RETURN',$4,'customer_return',1)`, uuid.New(), inventoryItemID, productID, returnID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_ledger (id,customer_id,reference_id,reference_type,type,amount,created_at) VALUES ($1,$2,$3,'return','credit',18,NOW())`, uuid.New(), customerID, returnID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO ledger_entries (id,reference_id,reference_type,amount) VALUES ($1,$2,'return',18)`, uuid.New(), returnID); err != nil {
		t.Fatal(err)
	}
	if err := returns.NewRepository(db).DeleteReturn(context.Background(), returnID); err != nil {
		t.Fatalf("delete PostgreSQL completed return: %v", err)
	}
	var returnsLeft, itemsLeft, effectsLeft, effectItemsLeft, movementCount, ledgerCount, entryCount, auditCount int
	for query, into := range map[string]*int{
		`SELECT COUNT(*) FROM returns WHERE id=$1`:                               &returnsLeft,
		`SELECT COUNT(*) FROM return_items WHERE return_id=$1`:                   &itemsLeft,
		`SELECT COUNT(*) FROM return_effects WHERE id=$1`:                        &effectsLeft,
		`SELECT COUNT(*) FROM return_effect_items WHERE return_effect_id=$1`:     &effectItemsLeft,
		`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=$1`:         &movementCount,
		`SELECT COUNT(*) FROM customer_ledger WHERE reference_id=$1`:             &ledgerCount,
		`SELECT COUNT(*) FROM ledger_entries WHERE reference_id=$1`:              &entryCount,
		`SELECT COUNT(*) FROM audit_logs WHERE entity_id=$1 AND action='DELETE'`: &auditCount,
	} {
		if err := db.Get(into, query, returnID); err != nil {
			t.Fatal(err)
		}
	}
	var stock, customerBalance float64
	var itemStatus string
	if err := db.Get(&stock, `SELECT quantity FROM inventory WHERE product_id=$1`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&customerBalance, `SELECT current_balance FROM customers WHERE id=$1`, customerID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&itemStatus, `SELECT status FROM inventory_items WHERE id=$1`, inventoryItemID); err != nil {
		t.Fatal(err)
	}
	if returnsLeft != 0 || itemsLeft != 0 || effectsLeft != 0 || effectItemsLeft != 0 || movementCount != 0 || ledgerCount != 0 || entryCount != 0 || auditCount != 1 || stock != 0 || customerBalance != 0 || itemStatus != "SOLD" {
		t.Fatalf("PostgreSQL completed return cleanup: source=%d items=%d effect=%d effect_items=%d movements=%d customer_ledger=%d ledger_entries=%d audit=%d stock=%v customer_balance=%v item_status=%s", returnsLeft, itemsLeft, effectsLeft, effectItemsLeft, movementCount, ledgerCount, entryCount, auditCount, stock, customerBalance, itemStatus)
	}
}

func TestPostgresDeleteCompletedSupplierReturnReversesPostedEffects(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE supplier_returns (id UUID PRIMARY KEY, customer_return_id UUID, sale_id UUID, purchase_id UUID, supplier_id UUID, return_number TEXT, status TEXT, source_status TEXT, reason TEXT, notes TEXT, created_by UUID, return_reason TEXT, return_date TIMESTAMPTZ, refund_amount NUMERIC(12,2), created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
		CREATE TABLE supplier_return_items (id UUID PRIMARY KEY, supplier_return_id UUID, customer_return_id UUID, sale_id UUID, sale_item_id UUID, inventory_item_id UUID, purchase_item_id UUID, product_id UUID, barcode TEXT, serial_number TEXT, return_reason TEXT, return_date TIMESTAMPTZ, created_at TIMESTAMPTZ, quantity INTEGER, unit_cost NUMERIC(12,2), purchase_cost NUMERIC(12,2));
		CREATE TABLE inventory_items (id UUID PRIMARY KEY, product_id UUID, status TEXT, updated_at TIMESTAMPTZ);
		CREATE TABLE inventory (product_id UUID PRIMARY KEY, quantity INTEGER, updated_at TIMESTAMPTZ);
		CREATE TABLE suppliers (id UUID PRIMARY KEY, current_balance NUMERIC(12,2), updated_at TIMESTAMPTZ);
		CREATE TABLE inventory_movements (id UUID PRIMARY KEY, item_id UUID, product_id UUID, reference_id UUID, reference_type TEXT, movement_type TEXT, quantity INTEGER);
		CREATE TABLE supplier_ledger (id UUID PRIMARY KEY, supplier_id UUID, type TEXT, balance NUMERIC(12,2), reference_id UUID, reference_type TEXT, transaction_type TEXT, amount NUMERIC(12,2), created_at TIMESTAMPTZ DEFAULT NOW());
		CREATE TABLE deleted_operation_snapshots (entity_type TEXT, operation_id UUID, business_number TEXT, status TEXT, snapshot JSONB, deleted_by UUID, PRIMARY KEY(entity_type,operation_id));
		CREATE TABLE audit_logs (id UUID PRIMARY KEY, user_id UUID, action TEXT, entity_type TEXT, entity_id UUID, new_values JSONB, created_at TIMESTAMPTZ);
	`); err != nil {
		t.Fatalf("create PostgreSQL supplier return deletion schema: %v", err)
	}
	returnID, supplierID, productID, purchaseItemID, inventoryItemID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO inventory_items (id,product_id,status) VALUES ($1,$2,'RETURNED')`, inventoryItemID, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (product_id,quantity) VALUES ($1,0)`, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO suppliers (id,current_balance) VALUES ($1,-12)`, supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_returns (id,purchase_id,supplier_id,return_number,status,reason,return_reason,return_date,refund_amount,created_at,updated_at) VALUES ($1,$2,$3,'SUPRET-PG-DEL','COMPLETED','damaged','damaged',NOW(),12,NOW(),NOW())`, returnID, uuid.New(), supplierID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_return_items (id,supplier_return_id,inventory_item_id,purchase_item_id,product_id,quantity,unit_cost,purchase_cost,created_at) VALUES ($1,$2,$3,$4,$5,1,12,12,NOW())`, uuid.New(), returnID, inventoryItemID, purchaseItemID, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id,product_id,reference_id,reference_type,movement_type,quantity) VALUES ($1,$2,$3,$4,'supplier_return','SUPPLIER_RETURN',-1)`, uuid.New(), inventoryItemID, productID, returnID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_ledger (id,supplier_id,type,balance,reference_id,reference_type,transaction_type,amount) VALUES ($1,$2,'credit',-12,$3,'supplier_return','SUPPLIER_RETURN',12)`, uuid.New(), supplierID, returnID); err != nil {
		t.Fatal(err)
	}
	if err := supplierreturns.NewService(db).Delete(context.Background(), returnID, true); err != nil {
		t.Fatalf("delete PostgreSQL completed supplier return: %v", err)
	}
	var sourceCount, itemCount, movementCount, ledgerCount, snapshotCount, auditCount int
	for query, into := range map[string]*int{
		`SELECT COUNT(*) FROM supplier_returns WHERE id=$1`:                                                        &sourceCount,
		`SELECT COUNT(*) FROM supplier_return_items WHERE supplier_return_id=$1`:                                   &itemCount,
		`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=$1`:                                           &movementCount,
		`SELECT COUNT(*) FROM supplier_ledger WHERE reference_id=$1`:                                               &ledgerCount,
		`SELECT COUNT(*) FROM deleted_operation_snapshots WHERE entity_type='supplier_return' AND operation_id=$1`: &snapshotCount,
		`SELECT COUNT(*) FROM audit_logs WHERE entity_type='supplier_return' AND entity_id=$1 AND action='DELETE'`: &auditCount,
	} {
		if err := db.Get(into, query, returnID); err != nil {
			t.Fatal(err)
		}
	}
	var stock, supplierBalance float64
	var itemStatus string
	if err := db.Get(&stock, `SELECT quantity FROM inventory WHERE product_id=$1`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&supplierBalance, `SELECT current_balance FROM suppliers WHERE id=$1`, supplierID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&itemStatus, `SELECT status FROM inventory_items WHERE id=$1`, inventoryItemID); err != nil {
		t.Fatal(err)
	}
	if sourceCount != 0 || itemCount != 0 || movementCount != 0 || ledgerCount != 0 || snapshotCount != 0 || auditCount != 1 || stock != 1 || supplierBalance != 0 || itemStatus != "AVAILABLE" {
		t.Fatalf("PostgreSQL supplier return cleanup: source=%d items=%d movement=%d ledger=%d snapshot=%d audit=%d stock=%v supplier_balance=%v item_status=%s", sourceCount, itemCount, movementCount, ledgerCount, snapshotCount, auditCount, stock, supplierBalance, itemStatus)
	}
}

func TestPostgresHistoricalCleanupHardDeletesApprovedExpenseAndUpdatesReports(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("UTC"); err != nil {
		t.Fatal(err)
	}
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`CREATE TABLE expenses (id UUID PRIMARY KEY, title TEXT NOT NULL, category TEXT, category_id UUID, amount NUMERIC(12,2) NOT NULL, currency TEXT, reference_number TEXT NOT NULL, reference TEXT, description TEXT, receipt_url TEXT, expense_date DATE NOT NULL, payment_method TEXT, is_recurring BOOLEAN, recurring_period TEXT, approved_by UUID, status TEXT, created_by UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)`); err != nil {
		t.Fatalf("create PostgreSQL expense cleanup schema: %v", err)
	}
	expenseID, categoryID, userID := uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO expenses (id,title,category,category_id,reference_number,reference,description,receipt_url,amount,currency,expense_date,payment_method,is_recurring,recurring_period,approved_by,status,created_by,created_at,updated_at) VALUES ($1,'Office supplies','Supplies',$2,'INV-ORIGINAL','INV-REF','private note','storage://receipt',25,'ILS',DATE '2026-09-25','cash',TRUE,'monthly',$3,'approved',$3,NOW(),NOW())`, expenseID, categoryID, userID); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	before, err := accounting.TotalExpensesForPeriod(context.Background(), db, start, end)
	if err != nil || before != 25 {
		t.Fatalf("PostgreSQL expense total before cleanup=%v err=%v; want 25", before, err)
	}
	beforeReport, err := reports.NewRepository(db).GetExpensesData(context.Background(), start, end)
	if err != nil || beforeReport.TotalExpenses != 25 || beforeReport.ByCategory["Supplies"] != 25 || beforeReport.ByPaymentMethod["cash"] != 25 {
		t.Fatalf("PostgreSQL expense report before cleanup=%+v err=%v; want 25 across total/category/method", beforeReport, err)
	}
	preview, err := loadHistoricalCleanupPreview(context.Background(), db, false, "expenses", "2026-09-25", "2026-09-25")
	if err != nil || preview.CandidateCount != 1 || len(preview.CandidateIDs) != 1 || preview.CandidateIDs[0] != expenseID.String() {
		t.Fatalf("PostgreSQL expense preview=%+v err=%v; want the selected expense", preview, err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(fmt.Sprintf(`{"type":"expenses","start_date":"2026-09-25","end_date":"2026-09-25","candidate_ids":[%q]}`, expenseID.String())))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", userID)
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("PostgreSQL expense cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	var rowCount int
	if err := db.Get(&rowCount, `SELECT COUNT(*) FROM expenses WHERE id=$1`, expenseID); err != nil {
		t.Fatal(err)
	}
	after, err := accounting.TotalExpensesForPeriod(context.Background(), db, start, end)
	if err != nil {
		t.Fatal(err)
	}
	afterReport, err := reports.NewRepository(db).GetExpensesData(context.Background(), start, end)
	if err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 1 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 || rowCount != 0 || after != 0 || afterReport.TotalExpenses != 0 || afterReport.ByCategory["Supplies"] != 0 || afterReport.ByPaymentMethod["cash"] != 0 {
		t.Fatalf("PostgreSQL expense cleanup result=%+v rows=%d totals=%v/%v report=%+v; want deleted row and zeroed aggregates", envelope.Data, rowCount, before, after, afterReport)
	}
}
func TestPostgresPaymentDeleteHardDeletesUnpostedPayments(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`CREATE TABLE payments (id UUID PRIMARY KEY, status TEXT NOT NULL, payment_status TEXT, customer_id UUID, supplier_id UUID, sale_id UUID, amount NUMERIC DEFAULT 0, payment_date TIMESTAMPTZ, created_at TIMESTAMPTZ DEFAULT NOW()); CREATE TABLE customer_ledger (id UUID PRIMARY KEY, reference_id UUID); CREATE TABLE supplier_ledger (id UUID PRIMARY KEY, reference_id UUID);`); err != nil {
		t.Fatalf("create PostgreSQL payment deletion schema: %v", err)
	}
	pendingID, completedID := uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO payments (id,status) VALUES ($1,'pending'),($2,'completed')`, pendingID, completedID); err != nil {
		t.Fatal(err)
	}
	repo := payments.NewRepository(db)
	if err := repo.Delete(context.Background(), completedID); err != nil {
		t.Fatalf("delete completed unposted PostgreSQL payment: %v", err)
	}
	if err := repo.Delete(context.Background(), pendingID); err != nil {
		t.Fatalf("delete pending PostgreSQL payment: %v", err)
	}
	var pendingCount, completedCount int
	if err := db.Get(&pendingCount, `SELECT COUNT(*) FROM payments WHERE id=$1`, pendingID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&completedCount, `SELECT COUNT(*) FROM payments WHERE id=$1`, completedID); err != nil {
		t.Fatal(err)
	}
	if pendingCount != 0 || completedCount != 0 {
		t.Fatalf("PostgreSQL payment deletion left pending=%d completed=%d; want 0/0", pendingCount, completedCount)
	}
}

func TestPostgresDeleteCatalogPreservesProductRowsAndDetachesDeletedBrand(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE categories (id UUID PRIMARY KEY);
		CREATE TABLE brands (id UUID PRIMARY KEY);
		CREATE TABLE products (id UUID PRIMARY KEY, category_id UUID, brand_id UUID);
	`); err != nil {
		t.Fatalf("create PostgreSQL catalog cleanup schema: %v", err)
	}
	categoryID, linkedBrandID, unusedBrandID, productID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO categories (id) VALUES ($1)`, categoryID); err != nil {
		t.Fatal(err)
	}
	for _, id := range []uuid.UUID{linkedBrandID, unusedBrandID} {
		if _, err := db.Exec(`INSERT INTO brands (id) VALUES ($1)`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO products (id,category_id,brand_id) VALUES ($1,$2,$3)`, productID, categoryID, linkedBrandID); err != nil {
		t.Fatal(err)
	}
	repo := products.NewRepository(db)
	if err := repo.DeleteBrand(context.Background(), linkedBrandID); err != nil {
		t.Fatalf("delete PostgreSQL brand and detach product: %v", err)
	}
	if err := repo.DeleteBrand(context.Background(), unusedBrandID); err != nil {
		t.Fatalf("delete unused PostgreSQL brand: %v", err)
	}
	if err := repo.DeleteCategory(context.Background(), categoryID); err != nil {
		t.Fatalf("delete PostgreSQL category and detach product: %v", err)
	}
	var productCount, categoryCount, linkedBrandCount, unusedBrandCount int
	var categoryIsNull bool
	if err := db.Get(&productCount, `SELECT COUNT(*) FROM products WHERE id=$1`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&categoryIsNull, `SELECT category_id IS NULL FROM products WHERE id=$1`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&categoryCount, `SELECT COUNT(*) FROM categories WHERE id=$1`, categoryID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&linkedBrandCount, `SELECT COUNT(*) FROM brands WHERE id=$1`, linkedBrandID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&unusedBrandCount, `SELECT COUNT(*) FROM brands WHERE id=$1`, unusedBrandID); err != nil {
		t.Fatal(err)
	}
	var brandIsNull bool
	if err := db.Get(&brandIsNull, `SELECT brand_id IS NULL FROM products WHERE id=$1`, productID); err != nil {
		t.Fatal(err)
	}
	if productCount != 1 || !categoryIsNull || !brandIsNull || categoryCount != 0 || linkedBrandCount != 0 || unusedBrandCount != 0 {
		t.Fatalf("PostgreSQL catalog cleanup: product=%d category_is_null=%t category=%d linked_brand=%d unused_brand=%d", productCount, categoryIsNull, categoryCount, linkedBrandCount, unusedBrandCount)
	}
}

func TestPostgresDeleteUnusedProductAndBlockProductWithStock(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE products (id UUID PRIMARY KEY, deleted_at TIMESTAMPTZ);
		CREATE TABLE inventory (product_id UUID PRIMARY KEY, quantity INTEGER NOT NULL, reserved_quantity INTEGER NOT NULL DEFAULT 0);
	`); err != nil {
		t.Fatalf("create PostgreSQL product deletion schema: %v", err)
	}
	unusedProductID, stockedProductID := uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id) VALUES ($1),($2)`, unusedProductID, stockedProductID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (product_id, quantity, reserved_quantity) VALUES ($1,0,0),($2,3,0)`, unusedProductID, stockedProductID); err != nil {
		t.Fatal(err)
	}
	repo := products.NewRepository(db)
	if err := repo.DeleteProduct(context.Background(), stockedProductID); err != products.ErrProductHasHistory {
		t.Fatalf("delete PostgreSQL product with stock=%v; want ErrProductHasHistory", err)
	}
	if err := repo.DeleteProduct(context.Background(), unusedProductID); err != nil {
		t.Fatalf("delete PostgreSQL unused product: %v", err)
	}
	var unusedCount, stockedCount, stockCount int
	for _, check := range []struct {
		query string
		id    uuid.UUID
		into  *int
	}{
		{`SELECT COUNT(*) FROM products WHERE id=$1`, unusedProductID, &unusedCount},
		{`SELECT COUNT(*) FROM products WHERE id=$1`, stockedProductID, &stockedCount},
		{`SELECT COUNT(*) FROM inventory WHERE product_id=$1`, stockedProductID, &stockCount},
	} {
		if err := db.Get(check.into, check.query, check.id); err != nil {
			t.Fatal(err)
		}
	}
	if unusedCount != 0 || stockedCount != 1 || stockCount != 1 {
		t.Fatalf("PostgreSQL product deletion counts unused=%d stocked=%d stock=%d; want 0/1/1", unusedCount, stockedCount, stockCount)
	}
}

func TestPostgresDeleteProductQuantityAdjustmentReversesAndRebasesLaterMovement(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE inventory (id UUID PRIMARY KEY, product_id UUID UNIQUE NOT NULL, quantity INTEGER NOT NULL, updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
		CREATE TABLE inventory_items (id UUID PRIMARY KEY, product_id UUID NOT NULL);
		CREATE TABLE inventory_movements (
			id UUID PRIMARY KEY, item_id UUID, product_id UUID, movement_type TEXT NOT NULL,
			quantity INTEGER NOT NULL, before_quantity INTEGER NOT NULL, after_quantity INTEGER NOT NULL,
			before_status TEXT, after_status TEXT, reference_type TEXT, reference_id UUID, reason TEXT, created_by UUID, created_at TIMESTAMPTZ NOT NULL
		);
	`); err != nil {
		t.Fatalf("create PostgreSQL adjustment cleanup schema: %v", err)
	}
	safeProduct, blockedProduct := uuid.New(), uuid.New()
	now := time.Now().UTC()
	for _, productID := range []uuid.UUID{safeProduct, blockedProduct} {
		if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity) VALUES ($1, $2, 5)`, uuid.New(), productID); err != nil {
			t.Fatal(err)
		}
	}
	safeID := uuid.New()
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,created_at) VALUES ($1,$2,'ADJUSTMENT',3,2,5,'product_quantity_adjustment',$2,$3)`, safeID, safeProduct, now.Add(-2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	blockedID := uuid.New()
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,created_at) VALUES ($1,$2,'ADJUSTMENT',3,2,5,'product_quantity_adjustment',$2,$3)`, blockedID, blockedProduct, now.Add(-2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,created_at) VALUES ($1,$2,'SALE',-1,5,4,'sale',$2,$3)`, uuid.New(), blockedProduct, now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	service := inventory.NewService(inventory.NewRepository(db), db)
	if err := service.DeleteProductQuantityAdjustment(context.Background(), safeID); err != nil {
		t.Fatalf("delete PostgreSQL safe adjustment: %v", err)
	}
	if err := service.DeleteProductQuantityAdjustment(context.Background(), blockedID); err != nil {
		t.Fatalf("delete PostgreSQL adjustment with later movement and rebase snapshots: %v", err)
	}
	var safeQuantity, safeMovementCount, blockedQuantity, blockedMovementCount int
	if err := db.Get(&safeQuantity, `SELECT quantity FROM inventory WHERE product_id=$1`, safeProduct); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&safeMovementCount, `SELECT COUNT(*) FROM inventory_movements WHERE id=$1`, safeID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&blockedQuantity, `SELECT quantity FROM inventory WHERE product_id=$1`, blockedProduct); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&blockedMovementCount, `SELECT COUNT(*) FROM inventory_movements WHERE product_id=$1`, blockedProduct); err != nil {
		t.Fatal(err)
	}
	var rebasedBefore, rebasedAfter int
	if err := db.QueryRowx(`SELECT before_quantity,after_quantity FROM inventory_movements WHERE product_id=$1`, blockedProduct).Scan(&rebasedBefore, &rebasedAfter); err != nil {
		t.Fatal(err)
	}
	if safeQuantity != 2 || safeMovementCount != 0 || blockedQuantity != 2 || blockedMovementCount != 1 || rebasedBefore != 2 || rebasedAfter != 1 {
		t.Fatalf("PostgreSQL adjustment cleanup state: safe=%d/%d rebased=%d/%d snapshot=%d->%d", safeQuantity, safeMovementCount, blockedQuantity, blockedMovementCount, rebasedBefore, rebasedAfter)
	}
}

func TestPostgresTechnicalCleanupDeletesExpiredRowsAndKeepsLiveRows(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE idempotency_keys (
			id UUID PRIMARY KEY, expires_at TIMESTAMPTZ NOT NULL,
			idempotency_key TEXT NOT NULL, request_hash TEXT NOT NULL, response_body JSONB NOT NULL
		);
		CREATE TABLE refresh_tokens (id UUID PRIMARY KEY, expires_at TIMESTAMPTZ NOT NULL, token TEXT NOT NULL);
	`); err != nil {
		t.Fatalf("create PostgreSQL technical cleanup schema: %v", err)
	}
	expiredIdempotencyID, liveIdempotencyID := uuid.New(), uuid.New()
	expiredRefreshID, liveRefreshID := uuid.New(), uuid.New()
	if _, err := db.Exec(`
		INSERT INTO idempotency_keys (id, expires_at, idempotency_key, request_hash, response_body)
		VALUES ($1, NOW() - INTERVAL '1 day', 'expired-key', 'expired-hash', '{}'::jsonb),
		       ($2, NOW() + INTERVAL '1 day', 'live-key', 'live-hash', '{}'::jsonb);
		INSERT INTO refresh_tokens (id, expires_at, token)
		VALUES ($3, NOW() - INTERVAL '1 day', 'expired-token'),
		       ($4, NOW() + INTERVAL '1 day', 'live-token')`, expiredIdempotencyID, liveIdempotencyID, expiredRefreshID, liveRefreshID); err != nil {
		t.Fatalf("insert PostgreSQL technical cleanup fixtures: %v", err)
	}
	gin.SetMode(gin.TestMode)
	previewRecorder := httptest.NewRecorder()
	previewContext, _ := gin.CreateTestContext(previewRecorder)
	previewContext.Request = httptest.NewRequest("GET", "/settings/cleanup/preview", nil)
	NewDatabaseHandler(db).PreviewCleanup(previewContext)
	if previewRecorder.Code != 200 {
		t.Fatalf("PostgreSQL technical cleanup preview status=%d body=%s", previewRecorder.Code, previewRecorder.Body.String())
	}
	var previewEnvelope struct {
		Data CleanupPreview `json:"data"`
	}
	if err := json.Unmarshal(previewRecorder.Body.Bytes(), &previewEnvelope); err != nil {
		t.Fatal(err)
	}
	if previewEnvelope.Data.TotalCount != 2 || previewEnvelope.Data.Storage.DatabaseBytes <= 0 {
		t.Fatalf("PostgreSQL technical cleanup preview=%+v; want two expired rows and storage size", previewEnvelope.Data)
	}
	cleanupRecorder := httptest.NewRecorder()
	cleanupContext, _ := gin.CreateTestContext(cleanupRecorder)
	cleanupContext.Request = httptest.NewRequest("POST", "/settings/cleanup", nil)
	NewDatabaseHandler(db).RunCleanup(cleanupContext)
	if cleanupRecorder.Code != 200 {
		t.Fatalf("PostgreSQL technical cleanup status=%d body=%s", cleanupRecorder.Code, cleanupRecorder.Body.String())
	}
	var resultEnvelope struct {
		Data CleanupResult `json:"data"`
	}
	if err := json.Unmarshal(cleanupRecorder.Body.Bytes(), &resultEnvelope); err != nil {
		t.Fatal(err)
	}
	var expiredIdempotencyCount, liveIdempotencyCount, expiredRefreshCount, liveRefreshCount int
	for _, check := range []struct {
		query  string
		id     uuid.UUID
		target *int
	}{
		{`SELECT COUNT(*) FROM idempotency_keys WHERE id=$1`, expiredIdempotencyID, &expiredIdempotencyCount},
		{`SELECT COUNT(*) FROM idempotency_keys WHERE id=$1`, liveIdempotencyID, &liveIdempotencyCount},
		{`SELECT COUNT(*) FROM refresh_tokens WHERE id=$1`, expiredRefreshID, &expiredRefreshCount},
		{`SELECT COUNT(*) FROM refresh_tokens WHERE id=$1`, liveRefreshID, &liveRefreshCount},
	} {
		if err := db.Get(check.target, check.query, check.id); err != nil {
			t.Fatal(err)
		}
	}
	if resultEnvelope.Data.DeletedCount != 2 || resultEnvelope.Data.Remaining != 0 || expiredIdempotencyCount != 0 || liveIdempotencyCount != 1 || expiredRefreshCount != 0 || liveRefreshCount != 1 {
		t.Fatalf("PostgreSQL technical cleanup result=%+v expired/live idempotency=%d/%d expired/live refresh=%d/%d; want 2 deleted and live rows preserved", resultEnvelope.Data, expiredIdempotencyCount, liveIdempotencyCount, expiredRefreshCount, liveRefreshCount)
	}
}
