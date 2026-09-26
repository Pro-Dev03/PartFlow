package returns

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestGetReturnWithItemsReturnsNotFoundOnlyForMissingReturn(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("database error is not masked as not found", func(t *testing.T) {
		db, err := sqlx.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		if _, err := db.Exec(`CREATE TABLE returns (id TEXT PRIMARY KEY)`); err != nil {
			t.Fatal(err)
		}
		if got := callGetReturnWithItems(t, db, uuid.New()); got != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d for a database/schema failure", got, http.StatusInternalServerError)
		}
	})

	t.Run("missing return is still not found", func(t *testing.T) {
		db, err := sqlx.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		_, err = db.Exec(`
			CREATE TABLE returns (
				id TEXT PRIMARY KEY, return_number TEXT, reference_number TEXT, sale_id TEXT, purchase_id TEXT,
				customer_id TEXT, return_date TEXT, return_type TEXT, status TEXT, total_refund_amount REAL,
				refund_method TEXT, refund_date TEXT, refund_reference TEXT, debt_id TEXT, debt_adjustment REAL,
				customer_credit REAL, reason TEXT, reason_detail TEXT, item_condition_after_return TEXT,
				is_warranty_claim INTEGER, warranty_id TEXT, warranty_valid_until TEXT, created_by TEXT,
				processed_by TEXT, approved_by TEXT, approved_at TEXT, notes TEXT, internal_notes TEXT,
				created_at TEXT, updated_at TEXT
			);
			CREATE TABLE sales (id TEXT PRIMARY KEY, customer_id TEXT);
			CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT);
			CREATE TABLE users (id TEXT PRIMARY KEY, first_name TEXT, last_name TEXT);
		`)
		if err != nil {
			t.Fatal(err)
		}
		if got := callGetReturnWithItems(t, db, uuid.New()); got != http.StatusNotFound {
			t.Fatalf("status = %d, want %d for a missing return", got, http.StatusNotFound)
		}
	})
}

func callGetReturnWithItems(t *testing.T, db *sqlx.DB, id uuid.UUID) int {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/returns/"+id.String()+"/with-items", nil)
	ctx.Params = gin.Params{{Key: "id", Value: id.String()}}
	NewHandler(NewService(NewRepository(db))).GetReturnWithItems(ctx)
	return recorder.Code
}
