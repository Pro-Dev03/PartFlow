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
		CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT NOT NULL, purchase_price REAL DEFAULT 0);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, purchase_cost REAL DEFAULT 0, condition TEXT, status TEXT, supplier_id TEXT, sold_at TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER DEFAULT 0, updated_at TEXT, created_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, sale_number TEXT, invoice_number TEXT, sale_date TEXT, customer_id TEXT, user_id TEXT, subtotal REAL, tax_amount REAL, discount_amount REAL, total_amount REAL, cost_amount REAL, gross_profit REAL, net_profit REAL, paid_amount REAL, remaining_amount REAL, payment_method TEXT, payment_status TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_price REAL, unit_cost REAL, discount_amount REAL, tax_amount REAL, total_amount REAL, supplier_id TEXT, created_at TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT NOT NULL UNIQUE, sale_id TEXT, customer_id TEXT, amount REAL, payment_method TEXT, payment_status TEXT, created_by TEXT, payment_date TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, amount REAL, balance REAL, reference_id TEXT, description TEXT, created_at TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, sale_id TEXT UNIQUE, amount REAL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, due_date TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL DEFAULT 0, credit_limit REAL DEFAULT 0, updated_at TEXT);
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
	productID, itemID := uuid.New(), uuid.New()
	_, err = db.Exec(`INSERT INTO products (id,name,purchase_price) VALUES ($1,'Widget',10); INSERT INTO inventory_items (id,product_id,purchase_cost,status,created_at,updated_at) VALUES ($2,$1,10,'AVAILABLE',$3,$3); INSERT INTO inventory (id,product_id,quantity,created_at,updated_at) VALUES ($4,$1,1,$3,$3)`, productID, itemID, now, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(NewRepository(db), db)
	price := 25.0
	sale, err := svc.CreateSale(context.Background(), uuid.Nil, &CreateSaleRequest{Items: []SaleItemRequest{{ProductID: productID, Quantity: 1, UnitPrice: price}}})
	if err != nil {
		t.Fatalf("CreateSale failed: %v", err)
	}
	if sale.TotalAmount != price {
		t.Fatalf("total = %v, want %v", sale.TotalAmount, price)
	}
	var status string
	if err := db.Get(&status, `SELECT status FROM inventory_items WHERE id = $1`, itemID); err != nil {
		t.Fatal(err)
	}
	if status != "SOLD" {
		t.Fatalf("inventory status = %s, want SOLD", status)
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
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT NOT NULL, purchase_cost REAL DEFAULT 0, condition TEXT, status TEXT, supplier_id TEXT, sold_at TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER DEFAULT 0, updated_at TEXT, created_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, sale_number TEXT, invoice_number TEXT, sale_date TEXT, customer_id TEXT, user_id TEXT, subtotal REAL, tax_amount REAL, discount_amount REAL, total_amount REAL, cost_amount REAL, gross_profit REAL, net_profit REAL, paid_amount REAL, remaining_amount REAL, payment_method TEXT, payment_status TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_price REAL, unit_cost REAL, discount_amount REAL, tax_amount REAL, total_amount REAL, supplier_id TEXT, created_at TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT NOT NULL UNIQUE, sale_id TEXT, customer_id TEXT, amount REAL, payment_method TEXT, payment_status TEXT, created_by TEXT, payment_date TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, amount REAL, balance REAL, reference_id TEXT, description TEXT, created_at TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, sale_id TEXT UNIQUE, amount REAL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, due_date TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL DEFAULT 0, credit_limit REAL DEFAULT 0, updated_at TEXT);
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
	_, err = svc.CreateSale(context.Background(), uuid.Nil, &CreateSaleRequest{
		Items:         []SaleItemRequest{{ProductID: productID, Quantity: 1, UnitPrice: 25}},
		PaymentMethod: stringPtr("cash"),
		PaymentAmount: 25,
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
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, amount REAL, balance REAL, reference_id TEXT, description TEXT, created_at TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL DEFAULT 0, credit_limit REAL DEFAULT 0, updated_at TEXT);
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
