package sync

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	authmw "github.com/partflow/smart-store/pkg/middleware"
	_ "modernc.org/sqlite"
)

func TestGetInitialDataReturnsSingleTenantTables(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	// These tables intentionally have no user_id column. This mirrors the
	// deployed single-tenant cloud schema and catches accidental user filters.
	for _, table := range []string{
		"categories", "suppliers", "customers", "products", "inventory_items",
		"sales", "sale_items", "purchases", "purchase_items", "payments", "debts", "expenses",
	} {
		if _, err := db.Exec("CREATE TABLE " + table + " (id TEXT PRIMARY KEY, created_at TEXT)"); err != nil {
			t.Fatalf("create %s: %v", table, err)
		}
		if _, err := db.Exec("INSERT INTO "+table+" (id, created_at) VALUES (?, ?)", table+"-1", "2026-08-30T00:00:00Z"); err != nil {
			t.Fatalf("insert %s: %v", table, err)
		}
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("user_id", "subscriber-1")
	NewHandler(sqlx.NewDb(db, "sqlite")).GetInitialData(ctx)

	if recorder.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Success bool             `json:"success"`
		Data    map[string][]any `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.Success {
		t.Fatal("response success=false")
	}
	for _, key := range []string{"customers", "products", "sales", "sale_items", "purchases", "purchase_items", "payments", "debts", "expenses", "inventory_items"} {
		if len(response.Data[key]) != 1 {
			t.Fatalf("%s rows = %d, want 1", key, len(response.Data[key]))
		}
	}
}

func TestPushDataRejectsEmptyBatch(t *testing.T) {
	// An empty push must never reach the database or be reported as a sync success.
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/sync/push", strings.NewReader(`{"operations":[]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	NewHandler(nil).PushData(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestPushDataRejectsOversizedBatch(t *testing.T) {
	// The batch limit prevents an owner action from becoming an unbounded cloud write.
	operations := make([]PushOperation, 51)
	for index := range operations {
		operations[index] = PushOperation{ID: fmt.Sprintf("operation-%d", index)}
	}
	body, err := json.Marshal(PushRequest{Operations: operations})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/sync/push", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	NewHandler(nil).PushData(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestSubscriberCannotPushLegacySyncOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/validate" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"data":{"valid":true,"user":{"id":%q,"email":"subscriber@example.test"}}}`, userID.String())
	}))
	defer cloud.Close()

	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	t.Setenv("PARTFLOW_CLOUD_API_URL", cloud.URL)
	t.Setenv("PARTFLOW_ADMIN_EMAILS", "")
	authmw.SetDisableAuth(false)
	authmw.SetJWTSecret("sync-subscriber-route-test-secret")
	authmw.SetDatabase(nil)

	localToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("sync-subscriber-route-test-secret"))
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(authmw.Auth())
	router.POST("/sync/push", authmw.Admin(), NewHandler(nil).PushData)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/sync/push", strings.NewReader(`{"operations":[{"id":"legacy-1","entity_type":"customers","entity_id":"customer-1","operation":"upsert","payload":"{}"}]}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+localToken)
	request.Header.Set("X-PartFlow-Cloud-Token", "live-subscriber-cloud-token")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), "ADMIN_REQUIRED") {
		t.Fatalf("subscriber legacy sync push status=%d body=%s; want ADMIN_REQUIRED", recorder.Code, recorder.Body.String())
	}
}
