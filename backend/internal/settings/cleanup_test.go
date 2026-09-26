package settings

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestCleanupPreviewAndRunSQLiteOnlyDeleteExpiredTechnicalRows(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE idempotency_keys (id TEXT PRIMARY KEY, idempotency_key TEXT, request_hash TEXT, response_body TEXT, expires_at TEXT);
		CREATE TABLE refresh_tokens (id TEXT PRIMARY KEY, token TEXT, expires_at TEXT);
		CREATE TABLE password_reset_tokens (id TEXT PRIMARY KEY, token TEXT, expires_at TEXT, used INTEGER);
		CREATE TABLE notifications (id TEXT PRIMARY KEY, title TEXT, message TEXT, data TEXT, expires_at TEXT, status TEXT, created_at TEXT);
		CREATE TABLE reservations (id TEXT PRIMARY KEY, notes TEXT, status TEXT, updated_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL);
	`)
	if err != nil {
		t.Fatalf("create cleanup test schema: %v", err)
	}
	old := time.Now().UTC().AddDate(0, 0, -120).Format("2006-01-02 15:04:05")
	future := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")
	if _, err := db.Exec(`INSERT INTO idempotency_keys VALUES ('idem-old','old','hash','{}',?),('idem-new','new','hash','{}',?)`, old, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO refresh_tokens VALUES ('refresh-old','old',?),('refresh-new','new',?)`, old, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO password_reset_tokens VALUES ('reset-expired','expired',?,0),('reset-used','used',?,1),('reset-valid','valid',?,0)`, old, future, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO notifications VALUES ('notice-old','old','message','{}',NULL,'read',?),('notice-live','live','message','{}',NULL,'unread',?)`, old, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO reservations VALUES ('reservation-old','old','cancelled',?),('reservation-live','live','active',?)`, old, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales VALUES ('sale-keep',100)`); err != nil {
		t.Fatal(err)
	}

	preview, err := buildCleanupPreview(context.Background(), db, true)
	if err != nil {
		t.Fatalf("preview cleanup: %v", err)
	}
	if preview.TotalCount != 6 {
		t.Fatalf("preview candidate count = %d, want 6", preview.TotalCount)
	}
	if preview.EstimatedBytes <= 0 || preview.Storage.DatabaseBytes <= 0 {
		t.Fatalf("preview should contain storage estimates: %+v", preview)
	}

	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/cleanup", nil)
	NewDatabaseHandler(db).RunCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("cleanup status = %d, body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data CleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode cleanup response: %v", err)
	}
	if envelope.Data.DeletedCount != 6 || envelope.Data.Remaining != 0 {
		t.Fatalf("cleanup result = %+v, want 6 deleted and none remaining", envelope.Data)
	}

	for _, check := range []struct {
		table string
		want  int
	}{
		{"idempotency_keys", 1},
		{"refresh_tokens", 1},
		{"password_reset_tokens", 1},
		{"notifications", 1},
		{"reservations", 1},
		{"sales", 1},
	} {
		var count int
		if err := db.Get(&count, `SELECT COUNT(*) FROM `+check.table); err != nil {
			t.Fatal(err)
		}
		if count != check.want {
			t.Errorf("%s rows remaining = %d, want %d", check.table, count, check.want)
		}
	}
}

func TestCleanupPreviewSkipsMissingLegacyTablesSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE notifications (id TEXT, expires_at TEXT, status TEXT, created_at TEXT)`); err != nil {
		t.Fatal(err)
	}
	preview, err := buildCleanupPreview(context.Background(), db, true)
	if err != nil {
		t.Fatalf("preview legacy cleanup: %v", err)
	}
	if preview.TotalCount != 0 || len(preview.Categories) != 0 {
		t.Fatalf("legacy preview = %+v, want incomplete notification schema skipped", preview)
	}
}

func TestLocalCleanupEndpointLabelsAndUsesSQLiteTarget(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL)`); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("GET", "/settings/database/cleanup/preview", nil)
	NewDatabaseHandler(db).PreviewLocalCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("local cleanup preview status = %d, body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data CleanupPreview `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Target != "local" {
		t.Fatalf("cleanup target = %q, want local", envelope.Data.Target)
	}
}

func TestCleanupSupportsCloudNotificationSchemaWithIsRead(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE notifications (id TEXT PRIMARY KEY, title TEXT, message TEXT, data TEXT, is_read INTEGER, created_at TEXT)`); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().AddDate(0, 0, -120).Format("2006-01-02 15:04:05")
	future := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")
	if _, err := db.Exec(`INSERT INTO notifications VALUES ('read-old','old','message','{}',1,?),('unread-old','old','message','{}',0,?),('read-new','new','message','{}',1,?)`, old, old, future); err != nil {
		t.Fatal(err)
	}

	preview, err := buildCleanupPreview(context.Background(), db, true)
	if err != nil {
		t.Fatalf("preview cloud-compatible notification schema: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Fatalf("notification preview count = %d, want only the old read row", preview.TotalCount)
	}
	specs, err := cleanupAvailableSpecs(context.Background(), db, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range specs {
		if spec.key == "old_notifications" && !strings.Contains(spec.wherePG, "is_read IS TRUE") {
			t.Fatalf("PostgreSQL notification predicate = %q, want is_read boolean support", spec.wherePG)
		}
	}
}

func TestCleanupDeletesOnlyExpiredReservationsAndOrphanSpecificationsSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE reservations (id TEXT PRIMARY KEY, notes TEXT, status TEXT, updated_at TEXT, expires_at TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY);
		CREATE TABLE item_specification_values (id TEXT PRIMARY KEY, inventory_item_id TEXT, value_text TEXT, value_number REAL, value_boolean INTEGER);
	`); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().AddDate(0, 0, -120).Format("2006-01-02 15:04:05")
	future := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")
	if _, err := db.Exec(`INSERT INTO reservations VALUES ('stale-active','expired','active',?,?),('live-active','live','active',?,?)`, old, old, future, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items VALUES ('item-present')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO item_specification_values VALUES ('orphan','item-missing','value',NULL,NULL),('linked','item-present','value',NULL,NULL)`); err != nil {
		t.Fatal(err)
	}

	preview, err := buildCleanupPreview(context.Background(), db, true)
	if err != nil {
		t.Fatalf("preview cleanup: %v", err)
	}
	if preview.TotalCount != 2 {
		t.Fatalf("preview candidates = %d, want expired reservation and orphan specification", preview.TotalCount)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/cleanup", nil)
	NewDatabaseHandler(db).RunCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("cleanup status = %d, body=%s", response.Code, response.Body.String())
	}
	var staleReservationCount, liveReservationCount, orphanCount, linkedCount int
	for query, destination := range map[string]*int{
		`SELECT COUNT(*) FROM reservations WHERE id='stale-active'`:        &staleReservationCount,
		`SELECT COUNT(*) FROM reservations WHERE id='live-active'`:         &liveReservationCount,
		`SELECT COUNT(*) FROM item_specification_values WHERE id='orphan'`: &orphanCount,
		`SELECT COUNT(*) FROM item_specification_values WHERE id='linked'`: &linkedCount,
	} {
		if err := db.Get(destination, query); err != nil {
			t.Fatal(err)
		}
	}
	if staleReservationCount != 0 || orphanCount != 0 || liveReservationCount != 1 || linkedCount != 1 {
		t.Fatalf("remaining stale=%d live=%d orphan=%d linked=%d", staleReservationCount, liveReservationCount, orphanCount, linkedCount)
	}
}

func TestCleanupRemovesExpiredLocalSessionsAndOldSyncedQueueOnly(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE local_sessions (id TEXT PRIMARY KEY, expires_at TEXT, access_token TEXT, refresh_token TEXT);
		CREATE TABLE sync_queue (id TEXT PRIMARY KEY, payload TEXT, synced_at TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().AddDate(0, 0, -120).Format("2006-01-02 15:04:05")
	future := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")
	if _, err := db.Exec(`INSERT INTO local_sessions VALUES ('expired',?,'access','refresh'),('active',?,'access','refresh')`, old, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sync_queue VALUES ('synced','{}',?),('pending','{}',NULL)`, old); err != nil {
		t.Fatal(err)
	}
	preview, err := buildCleanupPreview(context.Background(), db, true)
	if err != nil {
		t.Fatalf("preview cleanup: %v", err)
	}
	if preview.TotalCount != 2 {
		t.Fatalf("preview count = %d, want one expired session and one old synced operation", preview.TotalCount)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/cleanup", nil)
	NewDatabaseHandler(db).RunCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("cleanup status = %d, body=%s", response.Code, response.Body.String())
	}
	var expired, active, synced, pending int
	for query, destination := range map[string]*int{
		`SELECT COUNT(*) FROM local_sessions WHERE id='expired'`: &expired,
		`SELECT COUNT(*) FROM local_sessions WHERE id='active'`:  &active,
		`SELECT COUNT(*) FROM sync_queue WHERE id='synced'`:      &synced,
		`SELECT COUNT(*) FROM sync_queue WHERE id='pending'`:     &pending,
	} {
		if err := db.Get(destination, query); err != nil {
			t.Fatal(err)
		}
	}
	if expired != 0 || active != 1 || synced != 0 || pending != 1 {
		t.Fatalf("expired=%d active=%d synced=%d pending=%d", expired, active, synced, pending)
	}
}
