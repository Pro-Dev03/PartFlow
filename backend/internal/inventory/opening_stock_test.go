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

func TestListInventoryItemsWithSupplierInfo_DefaultIncludesOpeningStockQuantity(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/default-include-manual.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "MANUAL-OPEN-DEFAULT", "Manual Open Product", 15, 30); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db), db)
	userID := uuid.New()
	if _, err := service.CreateOpeningStock(context.Background(), &OpeningStockRequest{ProductID: &productID, Mode: OpeningStockModeQuantity, Quantity: 4, BusinessDate: "2026-03-01"}, userID); err != nil {
		t.Fatalf("create opening stock: %v", err)
	}

	repo := NewRepository(db)
	items, total, err := repo.ListInventoryItemsWithSupplierInfo(context.Background(), 10, 0, map[string]interface{}{})
	if err != nil {
		t.Fatalf("default listing query returned error: %v", err)
	}
	if total == 0 || len(items) == 0 {
		t.Fatalf("default listing missed opening stock quantity: total=%d len=%d", total, len(items))
	}
	found := false
	for _, item := range items {
		if item != nil && item.ProductID != nil && *item.ProductID == productID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected product %s to appear in default inventory rows: %+v", productID, items)
	}
}

func TestListInventoryItemsWithSupplierInfo_ManualOnlyIncludesOpeningStockQuantity(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/manual-only-filter.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "MANUAL-OPEN-001", "Manual Open Product", 15, 30); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db), db)
	userID := uuid.New()
	if _, err := service.CreateOpeningStock(context.Background(), &OpeningStockRequest{ProductID: &productID, Mode: OpeningStockModeQuantity, Quantity: 4, BusinessDate: "2026-03-01"}, userID); err != nil {
		t.Fatalf("create opening stock: %v", err)
	}

	repo := NewRepository(db)
	items, total, err := repo.ListInventoryItemsWithSupplierInfo(context.Background(), 10, 0, map[string]interface{}{"manual_only": true})
	if err != nil {
		t.Fatalf("manual_only query returned error: %v", err)
	}
	if total == 0 || len(items) == 0 {
		t.Fatalf("manual_only filter missed opening stock quantity: total=%d len=%d", total, len(items))
	}
	found := false
	for _, item := range items {
		if item != nil && item.ProductID != nil && *item.ProductID == productID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected product %s to appear in manual-only inventory rows: %+v", productID, items)
	}
}

func TestListInventoryItemsWithSupplierInfo_ManualOnlySeparatesSupplierLinkedItems(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/manual-only-supplier-separation.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	supplierID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "MIXED-001", "Mixed Stock Product", 50, 100); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO suppliers (id, code, name, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))`, supplierID, "MIXED-SUPPLIER", "Mixed Supplier"); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db), db)
	if _, err := service.CreateOpeningStock(context.Background(), &OpeningStockRequest{ProductID: &productID, Mode: OpeningStockModeQuantity, Quantity: 2, BusinessDate: "2026-03-02"}, uuid.New()); err != nil {
		t.Fatalf("create opening stock: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, purchase_cost, selling_price, status, supplier_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, 'AVAILABLE', ?, datetime('now'), datetime('now'))`, uuid.New(), productID, "MIXED-SUPPLIER-ITEM", "MIXED-SUPPLIER-BARCODE", "NEW", 50, 100, supplierID); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	manualItems, manualTotal, err := repo.ListInventoryItemsWithSupplierInfo(context.Background(), 10, 0, map[string]interface{}{"manual_only": true})
	if err != nil {
		t.Fatalf("manual-only query returned error: %v", err)
	}
	if manualTotal != 1 || len(manualItems) != 1 || manualItems[0].ProductID == nil || *manualItems[0].ProductID != productID || manualItems[0].SupplierID != nil {
		t.Fatalf("manual item missing or supplier-linked item leaked into manual-only results: total=%d items=%+v", manualTotal, manualItems)
	}

	supplierItems, supplierTotal, err := repo.ListInventoryItemsWithSupplierInfo(context.Background(), 10, 0, map[string]interface{}{"supplier_only": true})
	if err != nil {
		t.Fatalf("supplier-only query returned error: %v", err)
	}
	if supplierTotal != 1 || len(supplierItems) != 1 || supplierItems[0].ProductID == nil || *supplierItems[0].ProductID != productID {
		t.Fatalf("supplier-linked product missing from supplier-only results: total=%d items=%+v", supplierTotal, supplierItems)
	}
}

func TestListInventoryItemsWithSupplierInfo_AdvancedDateAndPriceFilters(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/advanced-filter.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productA := uuid.New()
	productB := uuid.New()
	supplierID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productA, "ADV-001", "Advanced Product A", 10, 25); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productB, "ADV-002", "Advanced Product B", 12, 30); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO suppliers (id, code, name, phone, email, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, supplierID, "SUP-ADV", "Supplier A", "0500000000", "a@example.com"); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, serial_number, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'AVAILABLE', ?, ?, datetime('now'), datetime('now'))`, uuid.New(), productA, "ADV-ITEM-001", "BC-ADV-001", "SN-ADV-001", "NEW", 120.0, 25.0, supplierID, "2026-01-05"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, serial_number, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'AVAILABLE', ?, ?, datetime('now'), datetime('now'))`, uuid.New(), productB, "ADV-ITEM-002", "BC-ADV-002", "SN-ADV-002", "NEW", 220.0, 30.0, supplierID, "2026-02-10"); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	items, total, err := repo.ListInventoryItemsWithSupplierInfo(context.Background(), 10, 0, map[string]interface{}{
		"purchase_date_from": "2026-01-01",
		"purchase_date_to":   "2026-02-28",
		"min_purchase_cost":  150.0,
		"max_purchase_cost":  250.0,
	})
	if err != nil {
		t.Fatalf("advanced filters query returned error: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0] == nil || items[0].ProductID == nil || *items[0].ProductID != productB {
		t.Fatalf("expected only product B to match advanced date and cost filters, got total=%d len=%d items=%+v", total, len(items), items)
	}
}

func TestListInventoryItemsWithSupplierInfo_OpeningStockFiltersMatchBusinessDateAndCost(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/opening-stock-advanced-filter.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "ADV-OPEN-001", "Opening Stock Product", 180, 320); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db), db)
	if _, err := service.CreateOpeningStock(context.Background(), &OpeningStockRequest{ProductID: &productID, Mode: OpeningStockModeQuantity, Quantity: 4, BusinessDate: "2026-04-05"}, userID); err != nil {
		t.Fatalf("create opening stock: %v", err)
	}

	repo := NewRepository(db)
	items, total, err := repo.ListInventoryItemsWithSupplierInfo(context.Background(), 10, 0, map[string]interface{}{
		"manual_only":        true,
		"purchase_date_from": "2026-04-01",
		"purchase_date_to":   "2026-04-30",
		"min_purchase_cost":  150.0,
		"max_purchase_cost":  250.0,
	})
	if err != nil {
		t.Fatalf("manual opening stock advanced filters query returned error: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0] == nil || items[0].ProductID == nil || *items[0].ProductID != productID {
		t.Fatalf("expected opening stock row to match business-date and cost filters, got total=%d len=%d items=%+v", total, len(items), items)
	}
}

func TestListInventoryItemsWithSupplierInfo_DeletedProductsAreHidden(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/deleted-product-hidden.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "DEL-001", "Deleted Product", 50, 100); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db), db)
	userID := uuid.New()
	if _, err := service.CreateOpeningStock(context.Background(), &OpeningStockRequest{ProductID: &productID, Mode: OpeningStockModeQuantity, Quantity: 2, BusinessDate: "2026-09-17"}, userID); err != nil {
		t.Fatalf("create opening stock: %v", err)
	}
	if _, err := db.Exec(`UPDATE products SET deleted_at = datetime('now') WHERE id = ?`, productID); err != nil {
		t.Fatalf("soft delete product: %v", err)
	}

	repo := NewRepository(db)
	items, total, err := repo.ListInventoryItemsWithSupplierInfo(context.Background(), 10, 0, map[string]interface{}{})
	if err != nil {
		t.Fatalf("default listing query returned error: %v", err)
	}
	if total != 0 || len(items) != 0 {
		t.Fatalf("deleted product should be hidden from default inventory results: total=%d len=%d items=%+v", total, len(items), items)
	}

	manualItems, manualTotal, err := repo.ListInventoryItemsWithSupplierInfo(context.Background(), 10, 0, map[string]interface{}{"manual_only": true})
	if err != nil {
		t.Fatalf("manual-only listing query returned error: %v", err)
	}
	if manualTotal != 0 || len(manualItems) != 0 {
		t.Fatalf("deleted product should be hidden from manual-only inventory results: total=%d len=%d items=%+v", manualTotal, len(manualItems), manualItems)
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
