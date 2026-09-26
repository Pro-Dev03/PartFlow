package settings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/localdb"
	_ "modernc.org/sqlite"
)

func TestRunHistoricalInventoryCleanupReversesStockAndDeletesAllSelectedItemsSQLite(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("UTC"); err != nil {
		t.Fatal(err)
	}
	db, err := sqlx.Open("sqlite", "file:inventory_cleanup_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(5)
	if _, err := db.Exec(`
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT);
		CREATE TABLE reservations (id TEXT PRIMARY KEY, item_id TEXT);
		CREATE TABLE inventory (product_id TEXT PRIMARY KEY, quantity INTEGER NOT NULL, reserved_quantity INTEGER NOT NULL DEFAULT 0, updated_at TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
	`); err != nil {
		t.Fatalf("create SQLite inventory cleanup schema: %v", err)
	}
	eligibleID, protectedID, productID := uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{eligibleID, protectedID} {
		if _, err := db.Exec(`INSERT INTO inventory_items (id,product_id,status,created_at) VALUES (?,?, 'AVAILABLE','2020-01-01T12:00:00Z')`, id, productID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO inventory_movements (id,item_id) VALUES (?,?)`, uuid.New(), protectedID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (product_id,quantity,reserved_quantity) VALUES (?,2,0)`, productID); err != nil {
		t.Fatal(err)
	}
	preview, err := loadHistoricalCleanupPreview(context.Background(), db, true, "inventory_items", "2020-01-01", "2020-01-01")
	if err != nil || preview.CandidateCount != 2 {
		t.Fatalf("inventory item cleanup preview=%+v err=%v; want both candidates", preview, err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(`{"type":"inventory_items","start_date":"2020-01-01","end_date":"2020-01-01","candidate_ids":["`+eligibleID.String()+`","`+protectedID.String()+`"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != http.StatusOK {
		t.Fatalf("inventory item cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 2 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 {
		t.Fatalf("inventory cleanup result=%+v; want both items deleted after reversing history", envelope.Data)
	}
	var itemCount, quantity, movementCount, auditCount int
	for query, destination := range map[string]*int{
		`SELECT COUNT(*) FROM inventory_items`:                     &itemCount,
		`SELECT COUNT(*) FROM inventory_movements WHERE item_id=?`: &movementCount,
		`SELECT COUNT(*) FROM audit_logs WHERE entity_id=?`:        &auditCount,
	} {
		var args []any
		if strings.Contains(query, "inventory_movements") {
			args = append(args, protectedID.String())
		} else if strings.Contains(query, "audit_logs") {
			args = append(args, eligibleID.String())
		}
		if err := db.Get(destination, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Get(&quantity, `SELECT quantity FROM inventory WHERE product_id=?`, productID.String()); err != nil {
		t.Fatal(err)
	}
	if itemCount != 0 || quantity != 0 || movementCount != 0 || auditCount != 1 {
		t.Fatalf("after inventory cleanup: items=%d stock=%d protected movements=%d deletion audits=%d", itemCount, quantity, movementCount, auditCount)
	}
}

func TestHistoricalInventoryAdjustmentCleanupReversesSafeRowsSQLite(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("UTC"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "inventory-adjustment-cleanup.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatalf("open local database: %v", err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	safeProduct, blockedProduct := uuid.New(), uuid.New()
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	for index, productID := range []uuid.UUID{safeProduct, blockedProduct} {
		if _, err := db.Exec(`INSERT INTO products (id,sku,name,cost_price,selling_price,is_active,created_at,updated_at) VALUES (?,?,?,0,10,1,?,?)`, productID, uuid.NewString(), "Adjustment cleanup", now, now); err != nil {
			t.Fatalf("insert product %d: %v", index, err)
		}
		if _, err := db.Exec(`INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES (?,?,5,?,?)`, uuid.New(), productID, now, now); err != nil {
			t.Fatalf("insert inventory %d: %v", index, err)
		}
	}
	safeID, blockedID := uuid.New(), uuid.New()
	insertMovement := func(id, productID uuid.UUID, movementType, referenceType string, quantity, before, after int, createdAt time.Time) {
		t.Helper()
		if _, err := db.Exec(`INSERT INTO inventory_movements (id,product_id,movement_type,quantity,before_quantity,after_quantity,reference_type,reference_id,reason,created_by,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, id, productID, movementType, quantity, before, after, referenceType, productID, "cleanup test", uuid.New(), createdAt.Format(time.RFC3339Nano)); err != nil {
			t.Fatalf("insert movement: %v", err)
		}
	}
	insertMovement(safeID, safeProduct, "ADJUSTMENT", "product_quantity_adjustment", 3, 2, 5, now)
	insertMovement(blockedID, blockedProduct, "ADJUSTMENT", "product_quantity_adjustment", 3, 2, 5, now)
	insertMovement(uuid.New(), blockedProduct, "SALE", "sale", -1, 5, 4, now.Add(time.Minute))
	if _, err := db.Exec(`UPDATE inventory SET quantity=4 WHERE product_id=?`, blockedProduct); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	previewRecorder := httptest.NewRecorder()
	previewContext, _ := gin.CreateTestContext(previewRecorder)
	previewContext.Request = httptest.NewRequest(http.MethodGet, "/settings/cleanup/historical?type=inventory_adjustments&start_date=2026-06-01&end_date=2026-06-01", nil)
	handler := NewDatabaseHandler(db)
	handler.PreviewHistoricalCleanup(previewContext)
	if previewRecorder.Code != http.StatusOK {
		t.Fatalf("preview adjustment cleanup status=%d body=%s", previewRecorder.Code, previewRecorder.Body.String())
	}
	var previewEnvelope struct {
		Data HistoricalCleanupPreview `json:"data"`
	}
	if err := json.Unmarshal(previewRecorder.Body.Bytes(), &previewEnvelope); err != nil {
		t.Fatal(err)
	}
	if previewEnvelope.Data.CandidateCount != 2 || len(previewEnvelope.Data.CandidateIDs) != 2 {
		t.Fatalf("adjustment cleanup preview=%+v; want two product adjustments", previewEnvelope.Data)
	}
	requestBody, err := json.Marshal(map[string]any{
		"type": "inventory_adjustments", "start_date": "2026-06-01", "end_date": "2026-06-01", "candidate_ids": previewEnvelope.Data.CandidateIDs,
	})
	if err != nil {
		t.Fatal(err)
	}
	runRecorder := httptest.NewRecorder()
	runContext, _ := gin.CreateTestContext(runRecorder)
	runContext.Request = httptest.NewRequest(http.MethodPost, "/settings/cleanup/historical", strings.NewReader(string(requestBody)))
	runContext.Request.Header.Set("Content-Type", "application/json")
	handler.RunHistoricalCleanup(runContext)
	if runRecorder.Code != http.StatusOK {
		t.Fatalf("run adjustment cleanup status=%d body=%s", runRecorder.Code, runRecorder.Body.String())
	}
	var resultEnvelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(runRecorder.Body.Bytes(), &resultEnvelope); err != nil {
		t.Fatal(err)
	}
	if resultEnvelope.Data.Deleted != 2 || resultEnvelope.Data.Blocked != 0 || resultEnvelope.Data.Failed != 0 {
		t.Fatalf("adjustment cleanup result=%+v; want both adjustments reversed and deleted", resultEnvelope.Data)
	}
	var safeQuantity, blockedQuantity int
	if err := db.Get(&safeQuantity, `SELECT quantity FROM inventory WHERE product_id=?`, safeProduct); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&blockedQuantity, `SELECT quantity FROM inventory WHERE product_id=?`, blockedProduct); err != nil {
		t.Fatal(err)
	}
	if safeQuantity != 2 || blockedQuantity != 1 {
		t.Fatalf("adjustment cleanup quantities safe=%d blocked=%d; want 2 and 1", safeQuantity, blockedQuantity)
	}
	var laterBefore, laterAfter int
	if err := db.QueryRow(`SELECT before_quantity,after_quantity FROM inventory_movements WHERE product_id=? AND movement_type='SALE'`, blockedProduct).Scan(&laterBefore, &laterAfter); err != nil {
		t.Fatal(err)
	}
	if laterBefore != 2 || laterAfter != 1 {
		t.Fatalf("later movement snapshots before=%d after=%d; want rebased to 2/1", laterBefore, laterAfter)
	}
}

func TestRunHistoricalHeldSaleCleanupIsScopedToCurrentUserSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE held_sales (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, items TEXT NOT NULL, created_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	userID, otherUserID := uuid.New(), uuid.New()
	ownedID, foreignID := uuid.New(), uuid.New()
	for _, row := range []struct{ id, owner string }{{ownedID.String(), userID.String()}, {foreignID.String(), otherUserID.String()}} {
		if _, err := db.Exec(`INSERT INTO held_sales (id,user_id,items,created_at) VALUES (?,?, '[]','2026-04-30T12:00:00Z')`, row.id, row.owner); err != nil {
			t.Fatal(err)
		}
	}
	preview, err := loadHistoricalCleanupPreviewForUser(context.Background(), db, true, "held_sales", "2026-04-30", "2026-04-30", userID)
	if err != nil || preview.CandidateCount != 1 || len(preview.CandidateIDs) != 1 || preview.CandidateIDs[0] != ownedID.String() {
		t.Fatalf("held sale preview=%+v err=%v; want only current user's record", preview, err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(`{"type":"held_sales","start_date":"2026-04-30","end_date":"2026-04-30","candidate_ids":["`+ownedID.String()+`","`+foreignID.String()+`"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", userID)
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != http.StatusOK {
		t.Fatalf("held sale cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 1 || envelope.Data.Blocked != 1 || envelope.Data.Failed != 0 {
		t.Fatalf("held sales cleanup result=%+v; want one deletion and one blocked foreign row", envelope.Data)
	}
	var ownRemaining, otherRemaining int
	if err := db.Get(&ownRemaining, `SELECT COUNT(*) FROM held_sales WHERE user_id=?`, userID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&otherRemaining, `SELECT COUNT(*) FROM held_sales WHERE user_id=?`, otherUserID.String()); err != nil {
		t.Fatal(err)
	}
	if ownRemaining != 0 || otherRemaining != 1 {
		t.Fatalf("remaining held sales: current user=%d other user=%d; foreign record must remain", ownRemaining, otherRemaining)
	}
}

func TestRunHistoricalInspectionCleanupReversesLinkedAndCompletedInspectionsSQLite(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("UTC"); err != nil {
		t.Fatal(err)
	}
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE inspections (
			id TEXT PRIMARY KEY, product_id TEXT, inventory_item_id TEXT, inspection_date TEXT NOT NULL,
			inspector_id TEXT NOT NULL, result TEXT NOT NULL, condition TEXT, grade TEXT, notes TEXT,
			images TEXT NOT NULL DEFAULT '[]', test_results TEXT NOT NULL DEFAULT '{}', acquisition_item_id TEXT,
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);
		CREATE TABLE inspection_items (id TEXT PRIMARY KEY, inspection_id TEXT, item_id TEXT, checkpoint_name TEXT, status TEXT, notes TEXT, images TEXT, created_at TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, serial_number TEXT, status TEXT, updated_at TEXT);
		CREATE TABLE acquisitions (id TEXT PRIMARY KEY, status TEXT, updated_at TEXT);
		CREATE TABLE acquisition_items (id TEXT PRIMARY KEY, acquisition_id TEXT, inventory_item_id TEXT, inspection_id TEXT, inspection_status TEXT, item_status TEXT, serial_number TEXT, updated_at TEXT);
		CREATE TABLE inspection_workflow_snapshots (inspection_id TEXT PRIMARY KEY, inventory_item_id TEXT, inventory_status_before TEXT, acquisition_item_id TEXT, acquisition_inventory_item_id_before TEXT, acquisition_inspection_id_before TEXT, acquisition_inspection_status_before TEXT, acquisition_item_status_before TEXT, acquisition_id TEXT, acquisition_status_before TEXT, captured_at TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	standaloneID, inventoryLinkedID, completedID := uuid.New(), uuid.New(), uuid.New()
	itemID, acquisitionID, inspectorID := uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO inventory_items (id,serial_number,status) VALUES (?, 'SER-1','INSPECTION')`, itemID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO acquisitions (id,status) VALUES (?, 'inspection')`, acquisitionID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO acquisition_items (id,acquisition_id,inventory_item_id,inspection_id,inspection_status,item_status,serial_number) VALUES (?,?,?,?,'pending','inspection','SER-2')`, acquisitionID, acquisitionID, itemID, inventoryLinkedID); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		id, result  string
		inventory   any
		acquisition any
	}{{standaloneID.String(), "pending", nil, nil}, {inventoryLinkedID.String(), "pending", itemID.String(), acquisitionID.String()}, {completedID.String(), "passed", nil, nil}} {
		if _, err := db.Exec(`INSERT INTO inspections (id,inspection_date,inspector_id,result,inventory_item_id,acquisition_item_id,created_at,updated_at) VALUES (?,'2020-01-01',?,?,?,?, '2020-01-01T12:00:00Z','2020-01-01T12:00:00Z')`, row.id, inspectorID.String(), row.result, row.inventory, row.acquisition); err != nil {
			t.Fatal(err)
		}
	}
	preview, err := loadHistoricalCleanupPreview(context.Background(), db, true, "inspections", "2020-01-01", "2020-01-01")
	if err != nil || preview.CandidateCount != 3 {
		t.Fatalf("inspection cleanup preview=%+v err=%v; want all three historical records", preview, err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(`{"type":"inspections","start_date":"2020-01-01","end_date":"2020-01-01","candidate_ids":["`+standaloneID.String()+`","`+inventoryLinkedID.String()+`","`+completedID.String()+`"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != http.StatusOK {
		t.Fatalf("inspection cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 3 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 {
		t.Fatalf("inspection cleanup result=%+v; want all three inspections reversed and deleted", envelope.Data)
	}
	var remaining int
	if err := db.Get(&remaining, `SELECT COUNT(*) FROM inspections`); err != nil {
		t.Fatal(err)
	}
	var inventoryStatus, acquisitionStatus, acquisitionItemStatus string
	if err := db.Get(&inventoryStatus, `SELECT status FROM inventory_items WHERE id=?`, itemID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&acquisitionStatus, `SELECT status FROM acquisitions WHERE id=?`, acquisitionID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&acquisitionItemStatus, `SELECT item_status FROM acquisition_items WHERE id=?`, acquisitionID); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 || inventoryStatus != "AVAILABLE" || acquisitionStatus != "draft" || acquisitionItemStatus != "available" {
		t.Fatalf("inspection delete result: remaining=%d inventory=%s acquisition=%s item=%s", remaining, inventoryStatus, acquisitionStatus, acquisitionItemStatus)
	}
}
