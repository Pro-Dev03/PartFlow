package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
)

func TestCreateOpeningStockSQLiteSupportsQuantityAndIndividualWithoutPurchase(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/opening-stock.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	customerID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "OPEN-001", "Opening Stock Product", 20, 40); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customers (id, code, name, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))`, customerID, "CUS-OPEN-001", "Opening Stock Customer"); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db), db)
	userID := uuid.New()
	quantityResult, err := service.CreateOpeningStock(context.Background(), &OpeningStockRequest{
		ProductID: &productID, Mode: OpeningStockModeQuantity, Quantity: 3, BusinessDate: "2026-01-05",
	}, userID)
	if err != nil {
		t.Fatalf("quantity opening stock: %v", err)
	}
	if quantityResult.SourceType != OpeningStockSource || quantityResult.Quantity != 3 {
		t.Fatalf("quantity result = %+v", quantityResult)
	}

	barcode := "OPEN-BC-001"
	serial := "OPEN-SN-001"
	individualResult, err := service.CreateOpeningStock(context.Background(), &OpeningStockRequest{
		ProductID: &productID, Mode: OpeningStockModeIndividual, Quantity: 1, BusinessDate: "2026-01-06",
		Barcode: &barcode, SerialNumber: &serial, Condition: ConditionUsed, PurchaseCost: 12, SellingPrice: 30, CustomerID: &customerID,
	}, userID)
	if err != nil {
		t.Fatalf("individual opening stock: %v", err)
	}
	if individualResult.Item == nil || individualResult.Item.SupplierID != nil || individualResult.Item.CustomerID == nil || *individualResult.Item.CustomerID != customerID || individualResult.Item.Barcode == nil || *individualResult.Item.Barcode != barcode || individualResult.Item.SerialNumber == nil || *individualResult.Item.SerialNumber != serial || individualResult.Item.Condition != string(ConditionUsed) {
		t.Fatalf("individual result = %+v", individualResult.Item)
	}

	var quantity int
	if err := db.Get(&quantity, `SELECT quantity FROM inventory WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if quantity != 4 {
		t.Fatalf("inventory quantity = %d, want 4", quantity)
	}
	var source, businessDate string
	var movementQuantity int
	if err := db.QueryRow(`SELECT source_type, business_date, quantity FROM inventory_movements WHERE item_id IS NULL`).Scan(&source, &businessDate, &movementQuantity); err != nil {
		t.Fatal(err)
	}
	if source != OpeningStockSource || businessDate != "2026-01-05" || movementQuantity != 3 {
		t.Fatalf("quantity movement = source=%q date=%q quantity=%d", source, businessDate, movementQuantity)
	}
	var purchaseCount int
	if err := db.Get(&purchaseCount, `SELECT COUNT(*) FROM purchases WHERE id IS NOT NULL`); err != nil {
		t.Fatal(err)
	}
	if purchaseCount != 0 {
		t.Fatalf("opening stock created %d purchases", purchaseCount)
	}
}

func TestCreateOpeningStockAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/opening-stock-api.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "OPEN-API-001", "Opening API Product", 10, 25); err != nil {
		t.Fatal(err)
	}

	body, err := json.Marshal(OpeningStockRequest{ProductID: &productID, Mode: OpeningStockModeQuantity, Quantity: 2, BusinessDate: "2026-02-01"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/inventory/opening-stock", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req
	ctx.Set("user_id", uuid.New())

	NewHandler(NewService(NewRepository(db), db), db).CreateOpeningStock(ctx)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var quantity int
	if err := db.Get(&quantity, `SELECT quantity FROM inventory WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if quantity != 2 {
		t.Fatalf("API inventory quantity = %d, want 2", quantity)
	}
}
