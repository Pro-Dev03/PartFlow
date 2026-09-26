package sales

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestParseSaleTimePostgresTimestampWithSpaceAndUTCOffset(t *testing.T) {
	parsed, err := parseSaleTime("2026-09-25 11:24:32.414013+00")
	if err != nil {
		t.Fatalf("parseSaleTime() error = %v", err)
	}
	if parsed.UTC().Format(time.RFC3339Nano) != "2026-09-25T11:24:32.414013Z" {
		t.Fatalf("parsed timestamp = %s", parsed.UTC().Format(time.RFC3339Nano))
	}
}

func TestCurrentShiftParsesTextTimestamps(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	userID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`
		CREATE TABLE pos_shifts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'open',
			opened_at TEXT NOT NULL,
			opening_cash REAL NOT NULL DEFAULT 0,
			closed_at TEXT,
			closing_cash REAL,
			sales_total REAL NOT NULL DEFAULT 0,
			sale_count INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO pos_shifts (id, user_id, status, opened_at, opening_cash) VALUES (?, ?, 'open', ?, 100)`, uuid.New().String(), userID.String(), now)
	if err != nil {
		t.Fatal(err)
	}

	shift, err := currentShift(context.Background(), db, userID)
	if err != nil {
		t.Fatalf("currentShift() error = %v", err)
	}
	if shift == nil {
		t.Fatal("currentShift() returned nil shift")
	}
	if shift.OpenedAt.IsZero() {
		t.Fatal("currentShift() opened_at is zero")
	}
	if shift.OpeningCash != 100 {
		t.Fatalf("opening_cash = %v, want 100", shift.OpeningCash)
	}
}

func TestGenerateInvoiceNumberIncludesUniqueSuffix(t *testing.T) {
	svc := NewService(nil, nil)
	invoice := svc.generateInvoiceNumber()
	if !strings.HasPrefix(invoice, "INV-") {
		t.Fatalf("invoice = %q, want prefix INV-", invoice)
	}
	parts := strings.Split(invoice, "-")
	if len(parts) != 3 {
		t.Fatalf("invoice = %q, want format INV-YYYYMMDDHHMMSS-XXXX", invoice)
	}
	if len(parts[1]) != 14 {
		t.Fatalf("invoice timestamp segment = %q, want 14 digits", parts[1])
	}
	if len(parts[2]) != 4 {
		t.Fatalf("invoice suffix = %q, want length 4", parts[2])
	}
}

func TestListSalesSearchFiltersBeforePagination(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE sales (
			id TEXT PRIMARY KEY, sale_date TEXT, customer_id TEXT, invoice_number TEXT,
			subtotal REAL DEFAULT 0, tax_amount REAL DEFAULT 0, discount_amount REAL DEFAULT 0, total_amount REAL DEFAULT 0,
			cost_amount REAL DEFAULT 0, gross_profit REAL DEFAULT 0, net_profit REAL DEFAULT 0, paid_amount REAL DEFAULT 0,
			payment_method TEXT, payment_status TEXT DEFAULT '', status TEXT DEFAULT '', notes TEXT,
			created_at TEXT, updated_at TEXT
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	customerID := uuid.New()
	otherCustomerID := uuid.New()
	if _, err = db.Exec(`INSERT INTO customers (id, name) VALUES ($1, 'Target'), ($2, 'Other')`, customerID, otherCustomerID); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, sale := range []struct {
		invoice  string
		customer uuid.UUID
	}{
		{invoice: "INV-AB-001", customer: customerID},
		{invoice: "INV-AB-002", customer: customerID},
		{invoice: "INV-XYZ-003", customer: customerID},
		{invoice: "INV-AB-004", customer: otherCustomerID},
	} {
		if _, err := db.Exec(`INSERT INTO sales (id, sale_date, customer_id, invoice_number, created_at, updated_at) VALUES ($1, $2, $3, $4, $2, $2)`, uuid.New(), now, sale.customer, sale.invoice); err != nil {
			t.Fatal(err)
		}
	}

	repo := NewRepository(db)
	rows, total, err := repo.ListSales(context.Background(), 1, 1, map[string]interface{}{
		"customer_id": customerID,
		"search":      "ab-00",
	})
	if err != nil {
		t.Fatalf("ListSales() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("total = %d, want 2 matching invoices for selected customer", total)
	}
	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want one paginated result", len(rows))
	}
	if !strings.Contains(strings.ToLower(rows[0].InvoiceNumber), "ab-00") {
		t.Fatalf("invoice = %q, want search match", rows[0].InvoiceNumber)
	}
}

func TestGetSaleItemsIncludesProductName(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT NOT NULL);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT, purchase_cost REAL, serial_number TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, invoice_number TEXT, sale_date TEXT, customer_id TEXT, subtotal REAL, tax_amount REAL, discount_amount REAL, total_amount REAL, cost_amount REAL, gross_profit REAL, net_profit REAL, paid_amount REAL, payment_method TEXT, payment_status TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_price REAL, unit_cost REAL, discount_amount REAL, tax_amount REAL, total_amount REAL, created_at TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}

	saleID := uuid.New()
	productID := uuid.New()
	_, err = db.Exec(`INSERT INTO products (id, name) VALUES (?, ?)`, productID.String(), "Widget")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO sales (id, invoice_number, sale_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, saleID.String(), "INV-TEST-1", time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_price, unit_cost, total_amount, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, uuid.New().String(), saleID.String(), productID.String(), 1, 25.0, 10.0, 25.0, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	items, err := repo.GetSaleItems(context.Background(), saleID)
	if err != nil {
		t.Fatalf("GetSaleItems failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].ProductName == nil || *items[0].ProductName != "Widget" {
		t.Fatalf("product name = %#v, want Widget", items[0].ProductName)
	}
}

func TestCreateSaleSQLiteUsesLocalSchema(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT NOT NULL, barcode TEXT, purchase_price REAL DEFAULT 0);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, barcode TEXT, purchase_cost REAL DEFAULT 0, condition TEXT, status TEXT, supplier_id TEXT, serial_number TEXT, sold_at TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER DEFAULT 0, updated_at TEXT, created_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, sale_number TEXT, invoice_number TEXT, sale_date TEXT, customer_id TEXT, user_id TEXT, subtotal REAL, tax_amount REAL, discount_amount REAL, total_amount REAL, cost_amount REAL, gross_profit REAL, net_profit REAL, paid_amount REAL, remaining_amount REAL, payment_method TEXT, payment_status TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, barcode TEXT, quantity INTEGER, unit_price REAL, unit_cost REAL, discount_amount REAL, tax_amount REAL, total_amount REAL, supplier_id TEXT, created_at TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT NOT NULL UNIQUE, sale_id TEXT, customer_id TEXT, amount REAL, payment_method TEXT, payment_status TEXT, created_by TEXT, payment_date TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_payment_allocations (id TEXT PRIMARY KEY, sale_id TEXT NOT NULL, amount REAL NOT NULL, payment_method TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending', check_number TEXT, bank_name TEXT, check_date TEXT, created_at TEXT NOT NULL);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, amount REAL, balance REAL, reference_id TEXT, description TEXT, created_at TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, sale_id TEXT UNIQUE, amount REAL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, due_date TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT, current_balance REAL DEFAULT 0, credit_limit REAL DEFAULT 0, updated_at TEXT);
		CREATE TABLE acquisition_items (id TEXT PRIMARY KEY, inventory_item_id TEXT, item_status TEXT, updated_at TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, event_type TEXT, event_date TEXT, reference_type TEXT, reference_id TEXT, description TEXT, metadata TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
		CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO settings (key, value) VALUES ('tax_rate', '0'), ('max_discount_rate', '0')`); err != nil {
		t.Fatal(err)
	}
	productID, itemID, customerID := uuid.New(), uuid.New(), uuid.New()
	_, err = db.Exec(`INSERT INTO products (id,name,purchase_price) VALUES ($1,'Widget',10); INSERT INTO inventory_items (id,product_id,purchase_cost,status,created_at,updated_at) VALUES ($2,$1,10,'AVAILABLE',$3,$3); INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES ($4,$1,1,$3,$3)`, productID, itemID, now, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO customers (id, name) VALUES ($1, $2)`, customerID, "Rana Test"); err != nil {
		t.Fatal(err)
	}
	svc := NewService(NewRepository(db), db)
	price := 25.0
	sale, err := svc.CreateSale(context.Background(), uuid.Nil, &CreateSaleRequest{
		CustomerID:    &customerID,
		Items:         []SaleItemRequest{{ProductID: productID, Barcode: "ALT-ITEM-001", Quantity: 1, UnitPrice: price}},
		PaymentMethod: stringPtr("cash"),
		PaymentAmount: price,
		CashReceived:  price,
	})
	if err != nil {
		t.Fatalf("CreateSale failed: %v", err)
	}
	if sale.TotalAmount != price {
		t.Fatalf("total = %v, want %v", sale.TotalAmount, price)
	}
	var storedSaleDate, storedCreatedAt, storedUpdatedAt string
	if err := db.QueryRow(`SELECT sale_date, created_at, updated_at FROM sales WHERE id = ?`, sale.ID).Scan(&storedSaleDate, &storedCreatedAt, &storedUpdatedAt); err != nil {
		t.Fatal(err)
	}
	parsedSaleDate, err := time.Parse(time.RFC3339Nano, storedSaleDate)
	if err != nil {
		t.Fatalf("stored sale_date is not an explicit UTC timestamp: %q: %v", storedSaleDate, err)
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, storedCreatedAt)
	if err != nil {
		t.Fatalf("stored created_at is not an explicit UTC timestamp: %q: %v", storedCreatedAt, err)
	}
	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, storedUpdatedAt)
	if err != nil {
		t.Fatalf("stored updated_at is not an explicit UTC timestamp: %q: %v", storedUpdatedAt, err)
	}
	if !sale.SaleDate.Equal(parsedSaleDate) || !sale.CreatedAt.Equal(parsedCreatedAt) || !sale.UpdatedAt.Equal(parsedUpdatedAt) {
		t.Fatalf("sale timestamps differ from SQLite: response=(%s, %s, %s) storage=(%s, %s, %s)", sale.SaleDate, sale.CreatedAt, sale.UpdatedAt, parsedSaleDate, parsedCreatedAt, parsedUpdatedAt)
	}
	var status string
	if err := db.Get(&status, `SELECT status FROM inventory_items WHERE id = $1`, itemID); err != nil {
		t.Fatal(err)
	}
	if status != "SOLD" {
		t.Fatalf("inventory status = %s, want SOLD", status)
	}
	var linkedItemID string
	if err := db.Get(&linkedItemID, `SELECT inventory_item_id FROM sale_items`); err != nil {
		t.Fatal(err)
	}
	if linkedItemID != itemID.String() {
		t.Fatalf("sale item inventory_item_id = %s, want %s", linkedItemID, itemID)
	}
	var persistedBarcode string
	if err := db.Get(&persistedBarcode, `SELECT barcode FROM sale_items`); err != nil {
		t.Fatal(err)
	}
	if persistedBarcode != "ALT-ITEM-001" {
		t.Fatalf("sale item barcode = %q, want actual scanned barcode", persistedBarcode)
	}
	details, err := svc.GetSale(context.Background(), sale.ID)
	if err != nil {
		t.Fatalf("GetSale after CreateSale failed: %v", err)
	}
	if details.Sale.CustomerID == nil || *details.Sale.CustomerID != customerID {
		t.Fatalf("persisted customer_id = %v, want %s", details.Sale.CustomerID, customerID)
	}
	if details.Sale.CustomerName == nil || *details.Sale.CustomerName != "Rana Test" {
		t.Fatalf("loaded customer name = %v, want Rana Test", details.Sale.CustomerName)
	}
	if !details.Sale.CreatedAt.Equal(parsedCreatedAt) || !details.Sale.UpdatedAt.Equal(parsedUpdatedAt) {
		t.Fatalf("sale details timestamps differ from SQLite: created=%s updated=%s", details.Sale.CreatedAt, details.Sale.UpdatedAt)
	}
	sales, _, err := svc.ListSales(context.Background(), 1, 10, nil)
	if err != nil {
		t.Fatalf("ListSales failed: %v", err)
	}
	if len(sales) != 1 || sales[0].CustomerName == nil || *sales[0].CustomerName != "Rana Test" {
		t.Fatalf("listed customer name = %#v, want Rana Test", sales)
	}
	if !sales[0].CreatedAt.Equal(parsedCreatedAt) {
		t.Fatalf("sales list created_at = %s, want %s", sales[0].CreatedAt, parsedCreatedAt)
	}
}

func TestCreateSaleSQLiteRecordsPaymentAmount(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT NOT NULL, purchase_price REAL DEFAULT 0);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, purchase_cost REAL DEFAULT 0, condition TEXT, status TEXT, supplier_id TEXT, serial_number TEXT, sold_at TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER DEFAULT 0, updated_at TEXT, created_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, sale_number TEXT, invoice_number TEXT, sale_date TEXT, customer_id TEXT, user_id TEXT, subtotal REAL, tax_amount REAL, discount_amount REAL, total_amount REAL, cost_amount REAL, gross_profit REAL, net_profit REAL, paid_amount REAL, remaining_amount REAL, payment_method TEXT, payment_status TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_price REAL, unit_cost REAL, discount_amount REAL, tax_amount REAL, total_amount REAL, supplier_id TEXT, created_at TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT NOT NULL UNIQUE, sale_id TEXT, customer_id TEXT, amount REAL, payment_method TEXT, payment_status TEXT, created_by TEXT, payment_date TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_payment_allocations (id TEXT PRIMARY KEY, sale_id TEXT NOT NULL, amount REAL NOT NULL, payment_method TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending', check_number TEXT, bank_name TEXT, check_date TEXT, created_at TEXT NOT NULL);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, amount REAL, balance REAL, reference_id TEXT, description TEXT, created_at TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, sale_id TEXT UNIQUE, amount REAL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, due_date TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT, current_balance REAL DEFAULT 0, credit_limit REAL DEFAULT 0, updated_at TEXT);
		CREATE TABLE acquisition_items (id TEXT PRIMARY KEY, inventory_item_id TEXT, item_status TEXT, updated_at TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, event_type TEXT, event_date TEXT, reference_type TEXT, reference_id TEXT, description TEXT, metadata TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}
	productID, itemID := uuid.New(), uuid.New()
	_, err = db.Exec(`INSERT INTO products (id,name,purchase_price) VALUES ($1,'Widget',10); INSERT INTO inventory_items (id,product_id,purchase_cost,status,created_at,updated_at) VALUES ($2,$1,10,'AVAILABLE',$3,$3); INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES ($4,$1,1,$3,$3)`, productID, itemID, now, uuid.New())
	if err != nil {
		t.Fatal(err)
	}

	svc := NewService(NewRepository(db), db)
	sale, err := svc.CreateSale(context.Background(), uuid.Nil, &CreateSaleRequest{
		Items:         []SaleItemRequest{{ProductID: productID, Quantity: 1, UnitPrice: 25}},
		PaymentMethod: stringPtr("cash"),
		PaymentAmount: 25,
		PaymentAllocations: []PaymentAllocationRequest{
			{Amount: 10, Method: "cash"},
			{Amount: 15, Method: "checks", CheckNumber: "CHK-100", BankName: "Test Bank", CheckDate: "2026-09-17"},
		},
	})
	if err != nil {
		t.Fatalf("CreateSale failed: %v", err)
	}

	var paidAmount float64
	if err := db.Get(&paidAmount, `SELECT paid_amount FROM sales`); err != nil {
		t.Fatal(err)
	}
	if paidAmount != 25 {
		t.Fatalf("paid amount = %v, want 25", paidAmount)
	}
	var paymentAmount float64
	if err := db.Get(&paymentAmount, `SELECT amount FROM payments`); err != nil {
		t.Fatal(err)
	}
	if paymentAmount != 25 {
		t.Fatalf("payment amount = %v, want 25", paymentAmount)
	}
	var allocationCount int
	if err := db.Get(&allocationCount, `SELECT COUNT(*) FROM sale_payment_allocations`); err != nil {
		t.Fatal(err)
	}
	if allocationCount != 2 {
		t.Fatalf("allocation count = %d, want 2", allocationCount)
	}
	var checkAmount float64
	if err := db.Get(&checkAmount, `SELECT amount FROM sale_payment_allocations WHERE payment_method = 'checks'`); err != nil {
		t.Fatal(err)
	}
	if checkAmount != 15 {
		t.Fatalf("check allocation = %v, want 15", checkAmount)
	}
	details, err := svc.GetSale(context.Background(), sale.ID)
	if err != nil {
		t.Fatalf("GetSale with allocations failed: %v", err)
	}
	if len(details.PaymentAllocations) != 2 {
		t.Fatalf("sale detail allocation count = %d, want 2", len(details.PaymentAllocations))
	}
}

func TestCreateSaleSQLiteUsesAggregateInventoryWhenSingleRepresentativeRowExists(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT NOT NULL, cost_price REAL DEFAULT 0, purchase_price REAL DEFAULT 0);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, purchase_cost REAL DEFAULT 0, condition TEXT, status TEXT, supplier_id TEXT, serial_number TEXT, sold_at TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER DEFAULT 0, updated_at TEXT, created_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, sale_number TEXT, invoice_number TEXT, sale_date TEXT, customer_id TEXT, user_id TEXT, subtotal REAL, tax_amount REAL, discount_amount REAL, total_amount REAL, cost_amount REAL, gross_profit REAL, net_profit REAL, paid_amount REAL, remaining_amount REAL, payment_method TEXT, payment_status TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_price REAL, unit_cost REAL, discount_amount REAL, tax_amount REAL, total_amount REAL, supplier_id TEXT, created_at TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT NOT NULL UNIQUE, sale_id TEXT, customer_id TEXT, amount REAL, payment_method TEXT, payment_status TEXT, created_by TEXT, payment_date TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_payment_allocations (id TEXT PRIMARY KEY, sale_id TEXT NOT NULL, amount REAL NOT NULL, payment_method TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending', check_number TEXT, bank_name TEXT, check_date TEXT, created_at TEXT NOT NULL);
		CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT, current_balance REAL DEFAULT 0, credit_limit REAL DEFAULT 0, updated_at TEXT);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, amount REAL, balance REAL, reference_id TEXT, description TEXT, created_at TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, sale_id TEXT UNIQUE, amount REAL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, due_date TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, event_type TEXT, event_date TEXT, reference_type TEXT, reference_id TEXT, description TEXT, metadata TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
		CREATE TABLE acquisition_items (id TEXT PRIMARY KEY, inventory_item_id TEXT, item_status TEXT, updated_at TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO settings (key, value) VALUES ('tax_rate', '0'), ('max_discount_rate', '0')`); err != nil {
		t.Fatal(err)
	}
	productID, itemID := uuid.New(), uuid.New()
	if _, err = db.Exec(`INSERT INTO products (id, name, cost_price, purchase_price) VALUES ($1, 'Widget', 20, 99); INSERT INTO inventory_items (id, product_id, purchase_cost, status, created_at, updated_at) VALUES ($2, $1, 20, 'AVAILABLE', $3, $3); INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES ($4, $1, 5, $3, $3)`, productID, itemID, now, uuid.New()); err != nil {
		t.Fatal(err)
	}

	svc := NewService(NewRepository(db), db)
	sale, err := svc.CreateSale(context.Background(), uuid.Nil, &CreateSaleRequest{
		Items:         []SaleItemRequest{{ProductID: productID, Quantity: 3, UnitPrice: 25, InventoryItemID: &itemID}},
		PaymentMethod: stringPtr("cash"),
		PaymentAmount: 75,
	})
	if err != nil {
		t.Fatalf("CreateSale failed when aggregate stock should cover representative item row: %v", err)
	}
	if sale.TotalAmount != 75 {
		t.Fatalf("total = %v, want 75", sale.TotalAmount)
	}
	if sale.CostAmount != 60 || sale.GrossProfit != 15 || sale.NetProfit != 15 {
		t.Fatalf("representative plus aggregate sale profit = cost %v, gross %v, net %v; want 60, 15, 15", sale.CostAmount, sale.GrossProfit, sale.NetProfit)
	}

	var qty int
	if err := db.Get(&qty, `SELECT quantity FROM inventory WHERE product_id = $1`, productID); err != nil {
		t.Fatal(err)
	}
	if qty != 2 {
		t.Fatalf("aggregate inventory quantity = %d, want 2 after selling 3 of 5", qty)
	}

	var status string
	if err := db.Get(&status, `SELECT status FROM inventory_items WHERE id = $1`, itemID); err != nil {
		t.Fatal(err)
	}
	if status != "SOLD" {
		t.Fatalf("representative inventory item status = %s, want SOLD", status)
	}

	// The representative item is now sold, leaving only aggregate stock. The
	// next sale must fall back to the product's recorded unit cost.
	aggregateOnlySale, err := svc.CreateSale(context.Background(), uuid.Nil, &CreateSaleRequest{
		Items:         []SaleItemRequest{{ProductID: productID, Quantity: 1, UnitPrice: 25}},
		PaymentMethod: stringPtr("cash"),
		PaymentAmount: 25,
	})
	if err != nil {
		t.Fatalf("CreateSale failed for aggregate-only stock: %v", err)
	}
	if aggregateOnlySale.CostAmount != 20 || aggregateOnlySale.GrossProfit != 5 || aggregateOnlySale.NetProfit != 5 {
		t.Fatalf("aggregate-only sale profit = cost %v, gross %v, net %v; want 20, 5, 5", aggregateOnlySale.CostAmount, aggregateOnlySale.GrossProfit, aggregateOnlySale.NetProfit)
	}
	details, err := svc.GetSale(context.Background(), aggregateOnlySale.ID)
	if err != nil {
		t.Fatalf("GetSale for aggregate-only sale: %v", err)
	}
	if details.Profit != 5 || len(details.Items) != 1 || details.Items[0].UnitCost != 20 {
		t.Fatalf("aggregate-only sale detail profit/cost = %v/%+v; want profit 5 and unit cost 20", details.Profit, details.Items)
	}
}

func TestCalculateSaleDetailProfitExcludesTax(t *testing.T) {
	service := &Service{}
	profit, err := service.calculateProfit(context.Background(), []SaleItem{{
		Quantity:    1,
		UnitCost:    20,
		TotalAmount: 27.5,
		TaxAmount:   2.5,
	}})
	if err != nil {
		t.Fatalf("calculateProfit returned error: %v", err)
	}
	if profit != 5 {
		t.Fatalf("sale detail profit = %v, want 5 after excluding 2.50 tax and 20 cost", profit)
	}
}

func TestCreateSaleSQLiteCreatesLinkedCreditDebt(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT NOT NULL, purchase_price REAL DEFAULT 0);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, purchase_cost REAL DEFAULT 0, condition TEXT, status TEXT, supplier_id TEXT, sold_at TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER DEFAULT 0, updated_at TEXT, created_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, sale_number TEXT, invoice_number TEXT, sale_date TEXT, customer_id TEXT, user_id TEXT, subtotal REAL, tax_amount REAL, discount_amount REAL, total_amount REAL, cost_amount REAL, gross_profit REAL, net_profit REAL, paid_amount REAL, remaining_amount REAL, payment_method TEXT, payment_status TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_price REAL, unit_cost REAL, discount_amount REAL, tax_amount REAL, total_amount REAL, supplier_id TEXT, created_at TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT NOT NULL UNIQUE, sale_id TEXT, customer_id TEXT, amount REAL, payment_method TEXT, payment_status TEXT, created_by TEXT, payment_date TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_payment_allocations (id TEXT PRIMARY KEY, sale_id TEXT NOT NULL, amount REAL NOT NULL, payment_method TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending', check_number TEXT, bank_name TEXT, check_date TEXT, created_at TEXT NOT NULL);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, amount REAL, balance REAL, reference_id TEXT, description TEXT, created_at TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, name TEXT, current_balance REAL DEFAULT 0, credit_limit REAL DEFAULT 0, updated_at TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, sale_id TEXT UNIQUE, amount REAL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, due_date TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE acquisition_items (id TEXT PRIMARY KEY, inventory_item_id TEXT, item_status TEXT, updated_at TEXT);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, event_type TEXT, event_date TEXT, reference_type TEXT, reference_id TEXT, description TEXT, metadata TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO settings (key, value) VALUES ('tax_rate', '0'), ('max_discount_rate', '15')`); err != nil {
		t.Fatal(err)
	}
	productID, itemID, customerID := uuid.New(), uuid.New(), uuid.New()
	_, err = db.Exec(`INSERT INTO products (id,name,purchase_price) VALUES ($1,'Widget',10); INSERT INTO inventory_items (id,product_id,purchase_cost,status,created_at,updated_at) VALUES ($2,$1,10,'AVAILABLE',$3,$3); INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES ($4,$1,1,$3,$3); INSERT INTO customers (id,current_balance,credit_limit,updated_at) VALUES ($5,0,1000,$3)`, productID, itemID, now, uuid.New(), customerID)
	if err != nil {
		t.Fatal(err)
	}

	svc := NewService(NewRepository(db), db)
	sale, err := svc.CreateSale(context.Background(), uuid.Nil, &CreateSaleRequest{
		CustomerID:    &customerID,
		Items:         []SaleItemRequest{{ProductID: productID, Quantity: 1, UnitPrice: 25}},
		PaymentMethod: stringPtr("debt"),
		PaymentAmount: 5,
	})
	if err != nil {
		t.Fatalf("CreateSale failed: %v", err)
	}
	if sale.PaymentStatus != "debt" {
		t.Fatalf("payment status = %s, want debt", sale.PaymentStatus)
	}

	var amount, paid, remaining, balance float64
	var debtSaleID string
	if err := db.QueryRow(`SELECT amount, paid_amount, remaining_amount, sale_id FROM debts`).Scan(&amount, &paid, &remaining, &debtSaleID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT current_balance FROM customers WHERE id = $1`, customerID).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if amount != 20 || paid != 0 || remaining != 20 || balance != 20 || debtSaleID != sale.ID.String() {
		t.Fatalf("debt values amount=%v paid=%v remaining=%v balance=%v sale_id=%s", amount, paid, remaining, balance, debtSaleID)
	}
}
