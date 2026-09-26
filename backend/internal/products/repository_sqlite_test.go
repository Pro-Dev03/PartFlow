package products

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func boolPtr(v bool) *bool { return &v }

func TestValidateProductRequestRejectsNegativeValues(t *testing.T) {
	tests := []struct {
		name string
		req  ProductRequest
	}{
		{name: "negative cost", req: ProductRequest{Name: "Part", CostPrice: -1}},
		{name: "negative selling price", req: ProductRequest{Name: "Part", SellingPrice: -1}},
		{name: "negative stock threshold", req: ProductRequest{Name: "Part", MinStockLevel: -1}},
		{name: "negative warranty", req: ProductRequest{Name: "Part", WarrantyDays: -1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateProductRequest(&test.req); err == nil {
				t.Fatal("expected negative product value to be rejected")
			}
		})
	}
}

func TestListProductsRejectsUntrustedSortExpressions(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE products (
			id TEXT PRIMARY KEY, category_id TEXT, brand_id TEXT, preferred_supplier_id TEXT,
			name TEXT NOT NULL, description TEXT, model TEXT, sku TEXT, barcode TEXT,
			cost_price REAL NOT NULL DEFAULT 0, selling_price REAL NOT NULL DEFAULT 0,
			track_serial INTEGER NOT NULL DEFAULT 0, track_individual INTEGER NOT NULL DEFAULT 0,
			min_stock_level INTEGER NOT NULL DEFAULT 0, warranty_days INTEGER NOT NULL DEFAULT 0,
			is_active INTEGER NOT NULL DEFAULT 1, deleted_at TEXT,
			created_at TEXT NOT NULL DEFAULT '2026-01-01T00:00:00Z',
			updated_at TEXT NOT NULL DEFAULT '2026-01-01T00:00:00Z'
		);
		INSERT INTO products (id, name, sku, is_active) VALUES
			('00000000-0000-4000-8000-000000000001', 'Zulu', 'Z-1', 1),
			('00000000-0000-4000-8000-000000000002', 'Alpha', 'A-1', 1);
	`)
	if err != nil {
		t.Fatalf("create product sort fixture: %v", err)
	}

	repo := NewRepository(db)
	products, total, err := repo.ListProducts(context.Background(), &ProductListRequest{
		Page: 1, PerPage: 1_000_000, SortBy: "name DESC, id",
	})
	if err != nil {
		t.Fatalf("list with untrusted sort expression: %v", err)
	}
	if total != 2 || len(products) != 2 || products[0].Name != "Alpha" || products[1].Name != "Zulu" {
		t.Fatalf("untrusted sort was not ignored: total=%d products=%#v", total, products)
	}
	var productTableCount int
	if err := db.Get(&productTableCount, `SELECT COUNT(*) FROM products`); err != nil || productTableCount != 2 {
		t.Fatalf("product table changed after invalid sort input: count=%d err=%v", productTableCount, err)
	}
	products, total, err = repo.ListProducts(context.Background(), &ProductListRequest{
		Page: int(^uint(0) >> 1), PerPage: 1_000_000,
	})
	if err != nil || total != 2 || len(products) != 0 {
		t.Fatalf("out-of-range page should return no rows safely: total=%d products=%#v err=%v", total, products, err)
	}
}

func TestListProducts_LowStockFilterUsesInventoryItems(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE products (
			id TEXT PRIMARY KEY,
			sku TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			category_id TEXT,
			brand_id TEXT,
			preferred_supplier_id TEXT,
			model TEXT,
			barcode TEXT,
			cost_price REAL DEFAULT 0,
			selling_price REAL DEFAULT 0,
			track_serial INTEGER NOT NULL DEFAULT 0,
			track_individual INTEGER NOT NULL DEFAULT 0,
			min_stock_level INTEGER DEFAULT 0,
			warranty_days INTEGER DEFAULT 0,
			is_active INTEGER NOT NULL DEFAULT 1,
			deleted_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE inventory_items (
			id TEXT PRIMARY KEY,
			product_id TEXT NOT NULL,
			item_code TEXT NOT NULL UNIQUE,
			barcode TEXT,
			serial_number TEXT,
			condition TEXT,
			purchase_cost REAL DEFAULT 0,
			selling_price REAL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'AVAILABLE',
			supplier_id TEXT,
			purchase_date TEXT,
			sold_at TEXT,
			notes TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE inventory (
			id TEXT PRIMARY KEY,
			product_id TEXT NOT NULL,
			quantity INTEGER NOT NULL DEFAULT 0
		);
	`)
	if err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	productID := uuid.NewString()
	_, err = db.Exec(`
		INSERT INTO products (
			id, sku, name, description, category_id, brand_id, preferred_supplier_id,
			model, barcode, cost_price, selling_price, track_serial, track_individual,
			min_stock_level, warranty_days, is_active, deleted_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?)
	`, productID, "SKU-LOW-1", "Low Stock Product", "Test", nil, nil, nil, "Model A", "BAR-LOW-1", 10.0, 25.0, 0, 0, 3, 0, nil, now, now)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}

	_, err = db.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, uuid.NewString(), productID, "ITEM-LOW-1", "BAR-LOW-1", "available", now, now)
	if err != nil {
		t.Fatalf("insert inventory item: %v", err)
	}

	repo := NewRepository(db)
	products, total, err := repo.ListProducts(context.Background(), &ProductListRequest{Page: 1, PerPage: 20, Search: "BAR-LOW-1", LowStockOnly: boolPtr(true)})
	if err != nil {
		t.Fatalf("ListProducts returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}
	if products[0].Name != "Low Stock Product" {
		t.Fatalf("unexpected product name: %s", products[0].Name)
	}

	products, total, err = repo.ListProducts(context.Background(), &ProductListRequest{
		Page: 1, PerPage: 20, Search: "BAR-LOW-1", InStockOnly: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("ListProducts with in-stock filter returned error: %v", err)
	}
	if total != 1 || len(products) != 1 {
		t.Fatalf("expected available product in filtered results, total=%d items=%d", total, len(products))
	}
	if _, err := db.Exec(`UPDATE inventory_items SET condition = 'USED' WHERE product_id = ?`, productID); err != nil {
		t.Fatalf("mark item used: %v", err)
	}
	products, total, err = repo.ListProducts(context.Background(), &ProductListRequest{
		Page: 1, PerPage: 20, Search: "BAR-LOW-1", InStockOnly: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("ListProducts with used-only stock returned error: %v", err)
	}
	if total != 0 || len(products) != 0 {
		t.Fatalf("expected used-only product to be excluded, total=%d items=%d", total, len(products))
	}
}

func TestGetProductStockCountExcludesUsedItems(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, quantity INTEGER NOT NULL DEFAULT 0);
		CREATE TABLE inventory_items (
			id TEXT PRIMARY KEY,
			product_id TEXT NOT NULL,
			status TEXT NOT NULL,
			condition TEXT
		);
	`)
	if err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	productID := uuid.NewString()
	usedOnlyProductID := uuid.NewString()
	multiLocationProductID := uuid.NewString()
	for _, item := range []struct {
		productID string
		condition string
	}{
		{productID, "NEW"},
		{productID, "NEW"},
		{productID, "USED"},
		{usedOnlyProductID, "USED"},
	} {
		if _, err := db.Exec(
			`INSERT INTO inventory_items (id, product_id, status, condition) VALUES (?, ?, 'AVAILABLE', ?)`,
			uuid.NewString(), item.productID, item.condition,
		); err != nil {
			t.Fatalf("insert inventory item: %v", err)
		}
	}
	for _, quantity := range []int{2, 3} {
		if _, err := db.Exec(
			`INSERT INTO inventory (id, product_id, quantity) VALUES (?, ?, ?)`,
			uuid.NewString(), multiLocationProductID, quantity,
		); err != nil {
			t.Fatalf("insert location stock: %v", err)
		}
	}

	repo := NewRepository(db)
	stock, err := repo.GetProductStockCount(context.Background(), uuid.MustParse(productID))
	if err != nil {
		t.Fatalf("GetProductStockCount returned error: %v", err)
	}
	if stock != 2 {
		t.Fatalf("expected two new available items, got %d", stock)
	}

	stock, err = repo.GetProductStockCount(context.Background(), uuid.MustParse(usedOnlyProductID))
	if err != nil {
		t.Fatalf("GetProductStockCount returned error: %v", err)
	}
	if stock != 0 {
		t.Fatalf("expected used-only stock to be excluded, got %d", stock)
	}

	stock, err = repo.GetProductStockCount(context.Background(), uuid.MustParse(multiLocationProductID))
	if err != nil {
		t.Fatalf("GetProductStockCount for multiple locations returned error: %v", err)
	}
	if stock != 5 {
		t.Fatalf("expected stock from both locations to be summed to five, got %d", stock)
	}
}

func TestDeleteProductWithHistoryIsBlockedWithoutDeletingLinkedTransactionsSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE products (id TEXT PRIMARY KEY, deleted_at TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT NOT NULL REFERENCES products(id));
		CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL, status TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT REFERENCES sales(id), product_id TEXT REFERENCES products(id));
		CREATE TABLE purchases (id TEXT PRIMARY KEY, total_amount REAL, status TEXT);
		CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT REFERENCES purchases(id), product_id TEXT REFERENCES products(id));
		CREATE TABLE returns (id TEXT PRIMARY KEY, total_amount REAL, status TEXT);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, total_amount REAL, status TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT NOT NULL REFERENCES products(id));
		CREATE TABLE return_items (id TEXT PRIMARY KEY, return_id TEXT REFERENCES returns(id), product_id TEXT REFERENCES products(id));
		CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT REFERENCES supplier_returns(id), product_id TEXT REFERENCES products(id), purchase_item_id TEXT, sale_item_id TEXT);
	`)
	if err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	productID := uuid.New()
	if _, err := db.Exec(`INSERT INTO products (id) VALUES (?)`, productID); err != nil {
		t.Fatalf("insert product: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id) VALUES (?, ?)`, uuid.New(), productID); err != nil {
		t.Fatalf("insert inventory summary: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id) VALUES (?, ?)`, uuid.New(), productID); err != nil {
		t.Fatalf("insert inventory item: %v", err)
	}
	saleID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO sales (id, total_amount, status) VALUES (?, ?, ?)`, saleID, 100.0, "completed"); err != nil {
		t.Fatalf("insert sale: %v", err)
	}
	saleItemID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO sale_items (id, sale_id, product_id) VALUES (?, ?, ?)`, saleItemID, saleID, productID); err != nil {
		t.Fatalf("insert sale item: %v", err)
	}
	purchaseID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO purchases (id, total_amount, status) VALUES (?, ?, ?)`, purchaseID, 80.0, "received"); err != nil {
		t.Fatalf("insert purchase: %v", err)
	}
	purchaseItemID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id) VALUES (?, ?, ?)`, purchaseItemID, purchaseID, productID); err != nil {
		t.Fatalf("insert purchase item: %v", err)
	}
	returnID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO returns (id, total_amount, status) VALUES (?, ?, ?)`, returnID, 15.0, "completed"); err != nil {
		t.Fatalf("insert return: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO return_items (id, return_id, product_id) VALUES (?, ?, ?)`, uuid.NewString(), returnID, productID); err != nil {
		t.Fatalf("insert return item: %v", err)
	}
	supplierReturnID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO supplier_returns (id, total_amount, status) VALUES (?, ?, ?)`, supplierReturnID, 10.0, "completed"); err != nil {
		t.Fatalf("insert supplier return: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_return_items (id, supplier_return_id, product_id, purchase_item_id, sale_item_id) VALUES (?, ?, ?, ?, ?)`, uuid.NewString(), supplierReturnID, productID, purchaseItemID, saleItemID); err != nil {
		t.Fatalf("insert supplier return item: %v", err)
	}

	if err := NewRepository(db).DeleteProduct(context.Background(), productID); err != ErrProductHasHistory {
		t.Fatalf("delete product error = %v, want ErrProductHasHistory", err)
	}

	var productCount, inventoryCount, itemCount, saleCount, saleItemCount, purchaseCount, purchaseItemCount, returnCount, returnItemCount, supplierReturnCount, supplierReturnItemCount int
	if err := db.Get(&productCount, `SELECT COUNT(*) FROM products WHERE id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&inventoryCount, `SELECT COUNT(*) FROM inventory WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&itemCount, `SELECT COUNT(*) FROM inventory_items WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&saleItemCount, `SELECT COUNT(*) FROM sale_items WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&purchaseItemCount, `SELECT COUNT(*) FROM purchase_items WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&returnItemCount, `SELECT COUNT(*) FROM return_items WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&supplierReturnItemCount, `SELECT COUNT(*) FROM supplier_return_items WHERE product_id = ?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&saleCount, `SELECT COUNT(*) FROM sales WHERE id = ?`, saleID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&purchaseCount, `SELECT COUNT(*) FROM purchases WHERE id = ?`, purchaseID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&returnCount, `SELECT COUNT(*) FROM returns WHERE id = ?`, returnID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&supplierReturnCount, `SELECT COUNT(*) FROM supplier_returns WHERE id = ?`, supplierReturnID); err != nil {
		t.Fatal(err)
	}
	if productCount != 1 || inventoryCount != 1 || itemCount != 1 || saleCount != 1 || saleItemCount != 1 || purchaseCount != 1 || purchaseItemCount != 1 || returnCount != 1 || returnItemCount != 1 || supplierReturnCount != 1 || supplierReturnItemCount != 1 {
		t.Fatalf("blocked delete changed transaction data: products=%d inventory=%d inventory_items=%d sales=%d sale_items=%d purchases=%d purchase_items=%d returns=%d return_items=%d supplier_returns=%d supplier_return_items=%d", productCount, inventoryCount, itemCount, saleCount, saleItemCount, purchaseCount, purchaseItemCount, returnCount, returnItemCount, supplierReturnCount, supplierReturnItemCount)
	}
}

func TestDeleteUnusedProductAndInventorySQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE products (id TEXT PRIMARY KEY); CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER DEFAULT 0, reserved_quantity INTEGER DEFAULT 0); CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT);`); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	productID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO products (id) VALUES (?)`, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, reserved_quantity) VALUES (?, ?, 0, 0)`, uuid.NewString(), productID); err != nil {
		t.Fatal(err)
	}
	if err := NewRepository(db).DeleteProduct(context.Background(), uuid.MustParse(productID)); err != nil {
		t.Fatalf("delete unused product: %v", err)
	}
	var remaining int
	for _, table := range []string{"products", "inventory", "inventory_items"} {
		if err := db.Get(&remaining, `SELECT COUNT(*) FROM `+table+` WHERE `+map[string]string{"products": "id", "inventory": "product_id", "inventory_items": "product_id"}[table]+`=?`, productID); err != nil {
			t.Fatal(err)
		}
		if remaining != 0 {
			t.Fatalf("%s rows remaining after unused product deletion = %d", table, remaining)
		}
	}
}

func TestDeleteProductWithAggregateStockIsBlockedAndPreservedSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE products (id TEXT PRIMARY KEY); CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER DEFAULT 0, reserved_quantity INTEGER DEFAULT 0); CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT);`); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	productID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO products (id) VALUES (?)`, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory (id, product_id, quantity, reserved_quantity) VALUES (?, ?, 7, 2)`, uuid.NewString(), productID); err != nil {
		t.Fatal(err)
	}
	if err := NewRepository(db).DeleteProduct(context.Background(), uuid.MustParse(productID)); err != ErrProductHasHistory {
		t.Fatalf("DeleteProduct error = %v, want ErrProductHasHistory", err)
	}
	var productCount, quantity, reserved int
	if err := db.Get(&productCount, `SELECT COUNT(*) FROM products WHERE id=?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT quantity, reserved_quantity FROM inventory WHERE product_id=?`, productID).Scan(&quantity, &reserved); err != nil {
		t.Fatal(err)
	}
	if productCount != 1 || quantity != 7 || reserved != 2 {
		t.Fatalf("blocked delete changed stock: products=%d quantity=%d reserved=%d", productCount, quantity, reserved)
	}
}

func TestDeleteProductWithIndividualInventoryItemIsBlockedSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE products (id TEXT PRIMARY KEY); CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT);`); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	productID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO products (id) VALUES (?)`, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id, product_id, status) VALUES (?, ?, 'AVAILABLE')`, uuid.NewString(), productID); err != nil {
		t.Fatal(err)
	}
	if err := NewRepository(db).DeleteProduct(context.Background(), uuid.MustParse(productID)); err != ErrProductHasHistory {
		t.Fatalf("DeleteProduct error = %v, want ErrProductHasHistory", err)
	}
	var productCount, itemCount int
	if err := db.Get(&productCount, `SELECT COUNT(*) FROM products WHERE id=?`, productID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&itemCount, `SELECT COUNT(*) FROM inventory_items WHERE product_id=?`, productID); err != nil {
		t.Fatal(err)
	}
	if productCount != 1 || itemCount != 1 {
		t.Fatalf("blocked delete changed product/item rows: products=%d items=%d", productCount, itemCount)
	}
}

func TestCreateCategory_SqliteReturningTimestamps(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE categories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			parent_id TEXT,
			icon TEXT,
			color TEXT,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("create categories table: %v", err)
	}

	repo := NewRepository(db)
	category := &Category{
		Name:        "Electronics",
		Description: "Test category",
		IsActive:    true,
	}

	if err := repo.CreateCategory(context.Background(), category); err != nil {
		t.Fatalf("CreateCategory returned error for SQLite: %v", err)
	}
	if category.ID == uuid.Nil {
		t.Fatal("expected category id to be populated")
	}
	if category.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be populated")
	}
	if category.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at to be populated")
	}
}

func TestCreateProduct_AllowsNilCategory(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE products (
			id TEXT PRIMARY KEY,
			sku TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			category_id TEXT,
			brand_id TEXT,
			preferred_supplier_id TEXT,
			model TEXT,
			barcode TEXT,
			cost_price REAL DEFAULT 0,
			selling_price REAL DEFAULT 0,
			track_serial INTEGER NOT NULL DEFAULT 0,
			track_individual INTEGER NOT NULL DEFAULT 0,
			min_stock_level INTEGER DEFAULT 0,
			warranty_days INTEGER DEFAULT 0,
			is_active INTEGER NOT NULL DEFAULT 1,
			deleted_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("create products table: %v", err)
	}

	service := NewService(NewRepository(db))
	product, err := service.CreateProduct(context.Background(), &ProductRequest{
		Name:         "Used Brake Pad",
		SKU:          "USED-BRAKE-001",
		CostPrice:    30,
		SellingPrice: 60,
	})
	if err != nil {
		t.Fatalf("CreateProduct returned unexpected error for nil category: %v", err)
	}
	if product == nil {
		t.Fatal("expected product to be created")
	}
	if product.CategoryID != nil {
		t.Fatalf("expected category_id to remain nil, got %v", product.CategoryID)
	}
	if product.Name != "Used Brake Pad" {
		t.Fatalf("unexpected product name: %s", product.Name)
	}
}

func TestCreateProductsBulk_ValidatesCategoryAndDuplicateSKU(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE categories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			parent_id TEXT,
			icon TEXT,
			color TEXT,
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE products (
			id TEXT PRIMARY KEY,
			sku TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			category_id TEXT,
			brand_id TEXT,
			preferred_supplier_id TEXT,
			model TEXT,
			barcode TEXT,
			cost_price REAL DEFAULT 0,
			selling_price REAL DEFAULT 0,
			track_serial INTEGER NOT NULL DEFAULT 0,
			track_individual INTEGER NOT NULL DEFAULT 0,
			min_stock_level INTEGER DEFAULT 0,
			warranty_days INTEGER DEFAULT 0,
			is_active INTEGER NOT NULL DEFAULT 1,
			deleted_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	categoryID := uuid.New()
	_, err = db.Exec(`INSERT INTO categories (id, name, description, is_active, created_at, updated_at) VALUES (?, ?, ?, 1, ?, ?)`, categoryID.String(), "Electronics", "Test category", time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}

	service := NewService(NewRepository(db))
	items := []ProductRequest{
		{Name: "Keyboard", SKU: "KEY-001", CategoryID: &categoryID, CostPrice: 20, SellingPrice: 40},
		{Name: "Mouse", SKU: "KEY-001", CategoryID: &categoryID, CostPrice: 15, SellingPrice: 30},
		{Name: "", SKU: "KEY-003", CategoryID: &categoryID, CostPrice: 10, SellingPrice: 20},
	}

	created, failed, err := service.CreateProductsBulk(context.Background(), items)
	if err != nil {
		t.Fatalf("CreateProductsBulk returned unexpected error: %v", err)
	}
	if len(created) != 1 {
		t.Fatalf("expected 1 created product, got %d", len(created))
	}
	if len(failed) != 2 {
		t.Fatalf("expected 2 failed items, got %d", len(failed))
	}
	if failed[0].Error == "" {
		t.Fatalf("expected first failed item to include an error")
	}
}
