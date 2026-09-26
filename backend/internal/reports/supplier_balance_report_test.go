package reports

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestSuppliersReportIncludesLedgerBalancesWithoutPurchases(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE suppliers (id TEXT PRIMARY KEY, name TEXT, is_active INTEGER, current_balance REAL);
		CREATE TABLE purchases (id TEXT PRIMARY KEY, supplier_id TEXT, total_amount REAL, paid_amount REAL, status TEXT);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, supplier_id TEXT, refund_amount REAL, status TEXT);
		CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, reference_id TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}
	supplierID := uuid.New()
	if _, err := db.Exec(`INSERT INTO suppliers (id, name, is_active, current_balance) VALUES (?, ?, 1, 75)`, supplierID, "Opening Balance Supplier"); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/reports/suppliers", NewHandler(NewService(NewRepository(db))).GenerateSuppliersReport)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/reports/suppliers", nil))
	if recorder.Code != 200 {
		t.Fatalf("suppliers report status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Data struct {
			Outstanding float64 `json:"total_outstanding"`
			BySupplier  []struct {
				Name        string  `json:"supplier_name"`
				Outstanding float64 `json:"outstanding"`
			} `json:"by_supplier"`
			WithBalance []struct {
				ID      uuid.UUID `json:"id"`
				Balance float64   `json:"balance"`
			} `json:"suppliers_with_balance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.Outstanding != 75 {
		t.Fatalf("total outstanding = %.2f, want ledger balance 75", payload.Data.Outstanding)
	}
	if len(payload.Data.BySupplier) != 1 || payload.Data.BySupplier[0].Outstanding != 75 {
		t.Fatalf("supplier breakdown = %#v, want opening balance 75", payload.Data.BySupplier)
	}
	if len(payload.Data.WithBalance) != 1 || payload.Data.WithBalance[0].ID != supplierID || payload.Data.WithBalance[0].Balance != 75 {
		t.Fatalf("suppliers with balance = %#v, want supplier with balance 75", payload.Data.WithBalance)
	}
}
