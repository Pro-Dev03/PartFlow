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

func TestCreateInventoryItemQuantityValidationSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/quantity-validation.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()

	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "QTY-001", "Quantity Validation", 10, 20); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db), db)
	for _, quantity := range []int{-1, 0} {
		_, err := service.CreateInventoryItem(context.Background(), &InventoryItemRequest{
			ProductID:    &productID,
			Quantity:     quantity,
			Condition:    ConditionNew,
			PurchaseCost: 10,
			SellingPrice: 20,
		}, uuid.New())
		if err != ErrInvalidQuantity {
			t.Fatalf("quantity %d: error = %v, want ErrInvalidQuantity", quantity, err)
		}
	}
	if _, err := service.CreateInventoryItem(context.Background(), &InventoryItemRequest{
		ProductID:    &productID,
		Quantity:     10001,
		Condition:    ConditionNew,
		PurchaseCost: 10,
		SellingPrice: 20,
	}, uuid.New()); err == nil {
		t.Fatal("quantity 10001 was accepted")
	}

	item, err := service.CreateInventoryItem(context.Background(), &InventoryItemRequest{
		ProductID:    &productID,
		Quantity:     1,
		Condition:    ConditionNew,
		PurchaseCost: 10,
		SellingPrice: 20,
	}, uuid.New())
	if err != nil {
		t.Fatalf("quantity 1: %v", err)
	}
	if item == nil || item.ID == uuid.Nil {
		t.Fatal("quantity 1 did not create an inventory item")
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM inventory_items WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("inventory rows = %d, want 1", count)
	}
}

func TestCreateInventoryItemRejectsInvalidQuantityAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, quantity := range []int{-1, 0, 10001} {
		t.Run(string(rune('0'+quantity+2)), func(t *testing.T) {
			h := NewHandler(NewService(nil, nil), nil)
			body, err := json.Marshal(InventoryItemRequest{
				ProductID: &[]uuid.UUID{uuid.New()}[0],
				Quantity:  quantity,
				Condition: ConditionNew,
			})
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodPost, "/inventory/items", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = req
			ctx.Set("user_id", uuid.New())

			h.CreateInventoryItem(ctx)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("quantity %d: status = %d, body = %s", quantity, recorder.Code, recorder.Body.String())
			}
		})
	}
}
