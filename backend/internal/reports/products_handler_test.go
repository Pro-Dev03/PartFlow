package reports

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestProductsReportLowStockMatchesAggregateAndSerializedInventory(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE categories (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE products (
			id TEXT PRIMARY KEY, name TEXT, category_id TEXT, is_active INTEGER,
			deleted_at TEXT, min_stock_level INTEGER
		);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER);
		CREATE TABLE inventory_items (
			id TEXT PRIMARY KEY, product_id TEXT, status TEXT, condition TEXT
		);
		INSERT INTO categories VALUES ('cat-1', 'Parts');
		INSERT INTO products VALUES
			('bulk', 'Bulk stock', 'cat-1', 1, NULL, 2),
			('serialized', 'Serialized stock', 'cat-1', 1, NULL, 2),
			('empty', 'Empty stock', 'cat-1', 1, NULL, 1);
		INSERT INTO inventory VALUES ('inventory-bulk', 'bulk', 3);
		INSERT INTO inventory VALUES ('inventory-bulk-location-2', 'bulk', 1);
		INSERT INTO inventory VALUES ('inventory-serialized', 'serialized', 2);
		INSERT INTO inventory_items VALUES
			('item-new-1', 'serialized', 'AVAILABLE', 'NEW'),
			('item-new-2', 'serialized', 'AVAILABLE', 'NEW'),
			('item-used', 'serialized', 'AVAILABLE', 'USED');
	`)
	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/reports/products", NewHandler(NewService(NewRepository(db))).GenerateProductsReport)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/reports/products", nil))
	if recorder.Code != 200 {
		t.Fatalf("products report status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			TotalProducts int            `json:"total_products"`
			LowStockCount int            `json:"low_stock_count"`
			ByCategory    map[string]int `json:"by_category"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.TotalProducts != 3 || payload.Data.LowStockCount != 2 {
		t.Fatalf("products report totals = products %d, low stock %d; want 3 and 2", payload.Data.TotalProducts, payload.Data.LowStockCount)
	}
	if payload.Data.ByCategory["Parts"] != 3 {
		t.Fatalf("products by category = %#v; want 3 products in Parts", payload.Data.ByCategory)
	}
}

func TestProductsReportReturns500InsteadOfInventingEmptyResults(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/reports/products", NewHandler(NewService(NewRepository(db))).GenerateProductsReport)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/reports/products", nil))
	if recorder.Code != 500 {
		t.Fatalf("products report status = %d, want 500 for missing schema", recorder.Code)
	}
}
