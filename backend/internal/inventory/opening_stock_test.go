package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestHandleError_RecognizesWrappedDuplicateBarcode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := fmt.Errorf("failed to create opening stock item: %w", ErrDuplicateBarcode)
	handleError(c, err)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusConflict)
	}
	if !strings.Contains(w.Body.String(), ErrDuplicateBarcode.Error()) {
		t.Fatalf("response = %q, want to contain %q", w.Body.String(), ErrDuplicateBarcode.Error())
	}
}

func TestDeleteInventoryItemPermanentBlocksAcquiredItemWithoutChangingHistory(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/permanent-delete-linked-item.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")

	productID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "PERM-DEL-001", "Permanent Delete Product", 30, 80); err != nil {
		t.Fatal(err)
	}
	// Sold items are already excluded from the aggregate available quantity.
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))`, uuid.New(), productID, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, status, purchase_cost, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, itemID, productID, "IT-DEL-001", "BAR-DEL-001", "USED", "SOLD", 30, 80); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO acquisition_items (id, acquisition_id, product_id, inventory_item_id, item_code, condition, grade, unit_cost, total_cost, item_status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, uuid.New(), uuid.New(), productID, itemID, "IT-DEL-001", "USED", "GOOD", 30, 30, "sold"); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db), db)
	if err := service.DeleteInventoryItem(context.Background(), itemID, userID); !errors.Is(err, ErrCannotDeleteItemWithHistory) {
		t.Fatalf("DeleteInventoryItem error = %v, want protected-history error", err)
	}

	var itemCount int
	if err := db.Get(&itemCount, `SELECT COUNT(*) FROM inventory_items WHERE id = ?`, itemID); err != nil {
		t.Fatal(err)
	}
	if itemCount != 1 {
		t.Fatalf("inventory item count after blocked delete = %d, want 1", itemCount)
	}

	var qty int
	if err := db.Get(&qty, `SELECT COALESCE(quantity, 0) FROM inventory WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if qty != 0 {
		t.Fatalf("inventory quantity still present after permanent delete: qty=%d", qty)
	}

	var acquisitionCount int
	if err := db.Get(&acquisitionCount, `SELECT COUNT(*) FROM acquisition_items WHERE inventory_item_id = ?`, itemID); err != nil {
		t.Fatal(err)
	}
	if acquisitionCount != 1 {
		t.Fatalf("acquisition link count after blocked delete = %d, want 1", acquisitionCount)
	}
}

func TestDeleteInventoryItemPermanentBlocksSaleAndAcquisitionLinksAtomically(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/permanent-delete-all-links.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")

	productID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, productID, "PERM-DEL-ALL", "Permanent Delete All Links", 20, 60); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))`, uuid.New(), productID, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, status, purchase_cost, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`, itemID, productID, "IT-DEL-ALL", "BAR-DEL-ALL", "USED", "AVAILABLE", 20, 60); err != nil {
		t.Fatal(err)
	}

	for _, statement := range []string{
		`CREATE TABLE IF NOT EXISTS item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT NOT NULL, event_type TEXT NOT NULL, event_date TEXT NOT NULL, reference_type TEXT, reference_id TEXT, description TEXT, metadata TEXT DEFAULT '{}', created_by TEXT, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS item_repair_costs (id TEXT PRIMARY KEY, inventory_item_id TEXT NOT NULL, acquisition_item_id TEXT, repair_date TEXT NOT NULL, repair_type TEXT NOT NULL, cost REAL NOT NULL DEFAULT 0, description TEXT, performed_by TEXT, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS trade_ins (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, inventory_item_id TEXT, purchase_price REAL NOT NULL DEFAULT 0, purchase_date TEXT NOT NULL, notes TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS barcodes (id TEXT PRIMARY KEY, code TEXT NOT NULL UNIQUE, product_id TEXT, inventory_item_id TEXT, type TEXT NOT NULL, is_active INTEGER DEFAULT 1, generated_at TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS inspections (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, inventory_item_id TEXT, inspector_id TEXT NOT NULL, inspection_date TEXT NOT NULL, result TEXT NOT NULL, condition TEXT, grade TEXT, notes TEXT, images TEXT DEFAULT '[]', test_results TEXT DEFAULT '{}', acquisition_item_id TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS item_specification_values (id TEXT PRIMARY KEY, inventory_item_id TEXT NOT NULL, specification_id TEXT NOT NULL, value_text TEXT, value_number REAL, value_boolean INTEGER, created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(inventory_item_id, specification_id))`,
		`CREATE TABLE IF NOT EXISTS sale_items (id TEXT PRIMARY KEY, sale_id TEXT NOT NULL, inventory_item_id TEXT, product_id TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price REAL NOT NULL, item_total REAL NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS return_items (id TEXT PRIMARY KEY, return_id TEXT NOT NULL, product_id TEXT, inventory_item_id TEXT, quantity_returned INTEGER NOT NULL DEFAULT 0, original_quantity INTEGER NOT NULL DEFAULT 0, unit_price REAL NOT NULL DEFAULT 0, total_refund_amount REAL NOT NULL DEFAULT 0, reason TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT NOT NULL, purchase_item_id TEXT NOT NULL, product_id TEXT NOT NULL, inventory_item_id TEXT, quantity INTEGER NOT NULL CHECK (quantity > 0), unit_cost REAL NOT NULL, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("create test table for permanent delete regression: %v", err)
		}
	}

	saleID := uuid.New()
	acquisitionID := uuid.New()
	customerID := uuid.New()
	if _, err := db.Exec(`INSERT INTO sales (id, sale_number, total_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`, saleID, "SALE-DEL-ALL-1", 0, "completed"); err != nil {
		t.Fatalf("create sale parent row: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO customers (id, code, name, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))`, customerID, "CUST-DEL-ALL", "Delete All Customer"); err != nil {
		t.Fatalf("create customer parent row: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO acquisitions (id, type, acquisition_date, total_cost, status, created_at, updated_at) VALUES (?, ?, datetime('now'), ?, ?, datetime('now'), datetime('now'))`, acquisitionID, "purchase", 0, "draft"); err != nil {
		t.Fatalf("create acquisition parent row: %v", err)
	}

	linkRows := []struct {
		tableName string
		insertSQL string
	}{
		{tableName: "item_history", insertSQL: `INSERT INTO item_history (id, inventory_item_id, event_type, event_date, description, created_at) VALUES (?, ?, ?, datetime('now'), ?, datetime('now'))`},
		{tableName: "item_repair_costs", insertSQL: `INSERT INTO item_repair_costs (id, inventory_item_id, repair_date, repair_type, cost, description, created_at) VALUES (?, ?, datetime('now'), ?, ?, ?, datetime('now'))`},
		{tableName: "acquisition_items", insertSQL: `INSERT INTO acquisition_items (id, acquisition_id, product_id, inventory_item_id, item_code, condition, grade, unit_cost, total_cost, item_status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`},
		{tableName: "trade_ins", insertSQL: `INSERT INTO trade_ins (id, customer_id, inventory_item_id, purchase_price, purchase_date, notes, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), ?, datetime('now'), datetime('now'))`},
		{tableName: "barcodes", insertSQL: `INSERT INTO barcodes (id, code, product_id, inventory_item_id, type, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, datetime('now'), datetime('now'))`},
		{tableName: "inspections", insertSQL: `INSERT INTO inspections (id, product_id, inventory_item_id, inspector_id, inspection_date, result, condition, grade, notes, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), ?, ?, ?, ?, datetime('now'), datetime('now'))`},
		{tableName: "sale_items", insertSQL: `INSERT INTO sale_items (id, sale_id, product_id, inventory_item_id, quantity, unit_price, item_total, created_at) VALUES (?, ?, ?, ?, 1, 60, 60, datetime('now'))`},
		{tableName: "item_specification_values", insertSQL: `INSERT INTO item_specification_values (id, inventory_item_id, specification_id, value_text, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))`},
	}

	for _, row := range linkRows {
		var err error
		switch row.tableName {
		case "item_history":
			_, err = db.Exec(row.insertSQL, uuid.New(), itemID, "TEST_EVENT", "test detail")
		case "item_repair_costs":
			_, err = db.Exec(row.insertSQL, uuid.New(), itemID, "repair", 12.5, "test repair")
		case "acquisition_items":
			_, err = db.Exec(row.insertSQL, uuid.New(), acquisitionID, productID, itemID, "IT-DEL-ALL", "USED", "GOOD", 20, 20, "sold")
		case "trade_ins":
			_, err = db.Exec(row.insertSQL, uuid.New(), customerID, itemID, 25, "trade-in note")
		case "barcodes":
			_, err = db.Exec(row.insertSQL, uuid.New(), "BAR-DEL-ALL-2", productID, itemID, "ITEM")
		case "inspections":
			_, err = db.Exec(row.insertSQL, uuid.New(), productID, itemID, uuid.New(), "PASS", "GOOD", "A", "works")
		case "sale_items":
			_, err = db.Exec(row.insertSQL, uuid.New(), saleID, productID, itemID)
		case "item_specification_values":
			_, err = db.Exec(row.insertSQL, uuid.New(), itemID, uuid.New(), "ok")
		}
		if err != nil {
			t.Fatalf("insert %s link row: %v", row.tableName, err)
		}
	}

	service := NewService(NewRepository(db), db)
	if err := service.DeleteInventoryItem(context.Background(), itemID, userID); !errors.Is(err, ErrCannotDeleteItemWithHistory) {
		t.Fatalf("DeleteInventoryItem error = %v, want protected-history error", err)
	}

	for _, row := range linkRows {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE inventory_item_id = ?", row.tableName)
		if err := db.Get(&count, query, itemID); err != nil {
			t.Fatalf("count linked rows in %s: %v", row.tableName, err)
		}
		if count != 1 {
			t.Fatalf("linked rows in %s changed after blocked delete: count=%d, want 1", row.tableName, count)
		}
	}
}

func TestDeleteUnusedInventoryItemPermanentlyRemovesItemAndStock(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/permanent-delete-unused-item.db")
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	productID := uuid.New()
	itemID := uuid.New()
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, created_at, updated_at) VALUES (?, ?, ?, datetime('now'), datetime('now'))`, productID, "UNUSED-DEL-001", "Unused delete product"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES (?, ?, 1, datetime('now'), datetime('now'))`, uuid.New(), productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, status, created_at, updated_at) VALUES (?, ?, ?, 'AVAILABLE', datetime('now'), datetime('now'))`, itemID, productID, "UNUSED-ITEM-001"); err != nil {
		t.Fatal(err)
	}

	if err := NewService(NewRepository(db), db).DeleteInventoryItem(context.Background(), itemID, userID); err != nil {
		t.Fatalf("delete unused item: %v", err)
	}
	var remaining int
	if err := db.Get(&remaining, `SELECT COUNT(*) FROM inventory_items WHERE id = ?`, itemID); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("inventory item count = %d, want 0", remaining)
	}
	if err := db.Get(&remaining, `SELECT quantity FROM inventory WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("inventory quantity = %d, want 0 after physical delete", remaining)
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
