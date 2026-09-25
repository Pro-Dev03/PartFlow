package sales

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestCurrentShiftDoesNotRequireManagerRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sqlDB, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	userID := uuid.New()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	if err := registerShiftRoutes(router.Group("/api/v1"), sqlDB); err != nil {
		t.Fatalf("register shift routes: %v", err)
	}
	if _, err := sqlDB.Exec(`CREATE TABLE sales (
		id TEXT PRIMARY KEY, user_id TEXT NOT NULL, status TEXT,
		created_at TEXT, total_amount REAL NOT NULL DEFAULT 0
	)`); err != nil {
		t.Fatal(err)
	}

	current := httptest.NewRecorder()
	router.ServeHTTP(current, httptest.NewRequest(http.MethodGet, "/api/v1/sales/shifts/current", nil))
	if current.Code != http.StatusOK {
		t.Fatalf("get current shift status=%d body=%s; want ordinary authenticated POS access", current.Code, current.Body.String())
	}

	open := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sales/shifts/open", strings.NewReader(`{"amount":25}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(open, request)
	if open.Code != http.StatusCreated {
		t.Fatalf("open shift as a subscriber status=%d body=%s; want store workflow access", open.Code, open.Body.String())
	}

	close := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/v1/sales/shifts/close", strings.NewReader(`{"amount":25}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(close, request)
	if close.Code != http.StatusOK {
		t.Fatalf("close shift as a subscriber status=%d body=%s; want store workflow access", close.Code, close.Body.String())
	}
}
