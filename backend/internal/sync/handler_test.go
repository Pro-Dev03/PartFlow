package sync

import (
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
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
