package reports

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestSuppliersReportPreservesSupplierCreditSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/reports-reconciliation.db")
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()

	db := sqlx.NewDb(database.DB, "sqlite")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	supplierID, purchaseID := uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO suppliers (id, code, name, current_balance, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, 1, ?, ?)`, supplierID, "REC-SUPPLIER", "Credit Supplier", -100.25, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, 0, 'received', ?, ?)`, purchaseID, "REC-PURCHASE", supplierID, 100, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_returns (id, purchase_id, supplier_id, return_number, status, reason, refund_amount, created_at, updated_at) VALUES (?, ?, ?, ?, 'COMPLETED', 'credit reconciliation', ?, ?, ?)`, uuid.New(), purchaseID, supplierID, "REC-RETURN", 200.25, now, now); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/suppliers-report", NewHandler(NewService(NewRepository(db))).GenerateSuppliersReport)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/suppliers-report", nil)
	router.ServeHTTP(recorder, request)
	if recorder.Code != 200 {
		t.Fatalf("supplier report status = %d", recorder.Code)
	}

	var payload struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	outstanding, ok := payload.Data["total_outstanding"].(float64)
	if !ok {
		t.Fatalf("supplier report total_outstanding missing: %#v", payload.Data)
	}
	if outstanding != 0 {
		t.Fatalf("supplier report outstanding = %v, want 0 when supplier has credit", outstanding)
	}
	credit, ok := payload.Data["supplier_credit_balance"].(float64)
	if !ok || credit != 100.25 {
		t.Fatalf("supplier report credit_balance = %v, want 100.25", payload.Data["supplier_credit_balance"])
	}
}
