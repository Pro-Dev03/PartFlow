package inventory

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestCreateTradeInMaterializesCustomerAndProduct(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/tradein.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	h := NewHandler(NewService(NewRepository(db), db), db)
	gin.SetMode(gin.TestMode)
	body, _ := json.Marshal(map[string]any{"customer_name": "New Seller", "product_name": "Used GPU", "purchase_cost": 100, "selling_price": 150})
	req := httptest.NewRequest(http.MethodPost, "/inventory/trade-ins", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req
	ctx.Set("user_id", uuid.New())
	h.CreateTradeIn(ctx)
	if ctx.Writer.Status() != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", ctx.Writer.Status(), http.StatusCreated, recorder.Body.String())
	}
	var count int
	if err := local.DB.QueryRow(`SELECT COUNT(*) FROM trade_ins`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("trade-in count = %d, want 1", count)
	}
	if err := local.DB.QueryRow(`SELECT COUNT(*) FROM customers WHERE name = 'New Seller'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("customer count = %d, want 1", count)
	}
	if err := local.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE name = 'Used GPU'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("product count = %d, want 1", count)
	}
}
