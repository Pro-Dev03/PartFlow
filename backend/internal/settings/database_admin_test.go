package settings

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestPostgresFullResetPreservesAllAccountsAndClearsStoreData(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`
		CREATE TABLE users (id UUID PRIMARY KEY, email TEXT NOT NULL UNIQUE);
		CREATE TABLE settings (id TEXT PRIMARY KEY, key TEXT NOT NULL UNIQUE, value TEXT, value_type TEXT, category TEXT, description TEXT, is_public BOOLEAN, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
		CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT);
		CREATE TABLE refresh_tokens (id INTEGER PRIMARY KEY, user_id UUID NOT NULL REFERENCES users(id));
		CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY);
	`); err != nil {
		t.Fatal(err)
	}
	ownerID, subscriberID := uuid.New(), uuid.New()
	if _, err := db.Exec(`
		INSERT INTO users VALUES ($1, 'owner@partflow.com'), ($2, 'subscriber@example.test');
		INSERT INTO settings VALUES ('setting-1', 'reset_test_setting', 'store', 'string', 'general', '', TRUE, NOW(), NOW());
		INSERT INTO products VALUES (1, 'test product');
		INSERT INTO refresh_tokens VALUES (1, $2);
		INSERT INTO schema_migrations VALUES (1);
	`, ownerID, subscriberID); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	wrongTargetRecorder := httptest.NewRecorder()
	wrongTargetContext, _ := gin.CreateTestContext(wrongTargetRecorder)
	wrongTargetContext.Request = httptest.NewRequest(http.MethodDelete, "/settings/database?target=offline&confirmation_token=DELETE%20ALL%20DATA", nil)
	NewDatabaseHandler(db).DeleteAllData(wrongTargetContext)
	if wrongTargetRecorder.Code != http.StatusBadRequest {
		t.Fatalf("cloud database accepted local target: status=%d body=%s", wrongTargetRecorder.Code, wrongTargetRecorder.Body.String())
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/settings/database?confirmation_token=DELETE%20ALL%20DATA", nil)
	NewDatabaseHandler(db).resetPostgreSQL(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("database reset status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, check := range []struct {
		table string
		want  int
	}{
		{table: "users", want: 2},
		{table: "products", want: 0},
		{table: "refresh_tokens", want: 0},
		{table: "schema_migrations", want: 1},
	} {
		var count int
		if err := db.Get(&count, `SELECT COUNT(*) FROM `+quoteSQLIdentifier(check.table)); err != nil {
			t.Fatal(err)
		}
		if count != check.want {
			t.Fatalf("%s count=%d; want %d", check.table, count, check.want)
		}
	}
	var customSettingCount, defaultSettingCount int
	if err := db.Get(&customSettingCount, `SELECT COUNT(*) FROM settings WHERE key='reset_test_setting'`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&defaultSettingCount, `SELECT COUNT(*) FROM settings WHERE key='store_name'`); err != nil {
		t.Fatal(err)
	}
	if customSettingCount != 0 || defaultSettingCount != 1 {
		t.Fatalf("settings after reset custom=%d default=%d; want custom data removed and defaults restored", customSettingCount, defaultSettingCount)
	}
}

func TestPostgresCloudBackupDownloadsCompleteJSONSnapshot(t *testing.T) {
	db := openIsolatedPostgresCleanupTestDB(t)
	if _, err := db.Exec(`CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT); INSERT INTO products VALUES (1, 'sample');`); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/settings/database/backup", nil)
	NewDatabaseHandler(db).DownloadCloudBackup(ctx)
	if recorder.Code != http.StatusOK || recorder.Header().Get("Content-Type") != "application/gzip" {
		t.Fatalf("backup status=%d headers=%v body=%s", recorder.Code, recorder.Header(), recorder.Body.String())
	}
	reader, err := gzip.NewReader(strings.NewReader(recorder.Body.String()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		Metadata map[string]any              `json:"metadata"`
		Tables   map[string][]map[string]any `json:"tables"`
	}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("decode backup: %v; data=%s", err, data)
	}
	if snapshot.Metadata["format_version"] != float64(1) || len(snapshot.Tables["products"]) != 1 || snapshot.Tables["products"][0]["name"] != "sample" {
		t.Fatalf("backup snapshot=%+v; want version 1 and product data", snapshot)
	}
}

func TestSQLiteFullResetPreservesSubscriberAccountsAndClearsData(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "reset.sqlite"))
	db, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	subscriberID := uuid.NewString()
	if _, err := db.DB.Exec(`INSERT INTO users (id,email,password_hash,first_name,last_name,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`, subscriberID, "subscriber@example.test", "test-hash", "Test", "Subscriber", "2026-09-27", "2026-09-27"); err != nil {
		db.DB.Close()
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`CREATE TABLE reset_test_data (id INTEGER PRIMARY KEY AUTOINCREMENT, value TEXT); INSERT INTO reset_test_data(value) VALUES ('business'); INSERT INTO settings (id,key,value,created_at,updated_at) VALUES ('test-setting','reset_test_setting','Test','2026-09-27','2026-09-27'); INSERT INTO local_sessions (id,user_id,email,access_token,refresh_token,expires_at) VALUES ('session','subscriber@example.test','subscriber@example.test','access','refresh','2099-01-01');`); err != nil {
		db.DB.Close()
		t.Fatal(err)
	}
	if err := db.DB.Close(); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/settings/database?confirmation_token=DELETE%20ALL%20DATA", nil)
	NewDatabaseHandler(nil).resetSQLite(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("SQLite reset status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	resetDB, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer resetDB.DB.Close()
	for _, check := range []struct {
		query string
		want  int
	}{
		{query: `SELECT COUNT(*) FROM users WHERE id=?`, want: 1},
		{query: `SELECT COUNT(*) FROM settings WHERE key='reset_test_setting'`, want: 0},
		{query: `SELECT COUNT(*) FROM local_sessions`, want: 0},
		{query: `SELECT COUNT(*) FROM reset_test_data`, want: 0},
	} {
		var count int
		if err := resetDB.DB.QueryRow(check.query, subscriberID).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != check.want {
			t.Fatalf("query %q count=%d; want %d", check.query, count, check.want)
		}
	}
}
