package suppliers

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/supplierreturns"
	_ "modernc.org/sqlite"
)

func TestSupplierLedgerPaymentLifecycleSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "supplier-ledger.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlite")
	for _, statement := range []string{
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, code TEXT, name TEXT, email TEXT, phone TEXT, address TEXT, city TEXT, country TEXT, tax_id TEXT, payment_terms TEXT, credit_limit REAL DEFAULT 0, current_balance REAL DEFAULT 0, notes TEXT, is_active INTEGER DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE products (id TEXT PRIMARY KEY, sku TEXT, name TEXT, cost_price REAL DEFAULT 0, selling_price REAL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE purchases (id TEXT PRIMARY KEY, purchase_number TEXT, supplier_id TEXT, purchase_date TEXT, total_amount REAL, paid_amount REAL, status TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE purchase_items (id TEXT PRIMARY KEY, purchase_id TEXT, product_id TEXT, quantity INTEGER, unit_price REAL, created_at TEXT)`,
		`CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, reserved_quantity INTEGER DEFAULT 0, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, item_code TEXT, barcode TEXT, condition TEXT, purchase_cost REAL, selling_price REAL, status TEXT, supplier_id TEXT, purchase_date TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT)`,
		`CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT, supplier_id TEXT, return_number TEXT UNIQUE, reason TEXT, notes TEXT, refund_amount REAL DEFAULT 0, status TEXT, created_by TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT, purchase_item_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_cost REAL, created_at TEXT)`,
		`CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL, description TEXT, reference_id TEXT, created_at TEXT)`,
		`CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT, supplier_id TEXT, amount REAL, payment_method TEXT, reference TEXT, notes TEXT, payment_date TEXT, payment_status TEXT, created_at TEXT, updated_at TEXT)`,
	} {
		if _, err := xdb.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}

	ctx := context.Background()
	supplierID := uuid.New()
	productID := uuid.New()
	purchaseID := uuid.New()
	purchaseItemID := uuid.New()
	inventoryItemID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := xdb.Exec(`INSERT INTO suppliers (id, code, name, current_balance, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, supplierID, "SUP-LIFE", "Lifecycle Supplier", 100, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, productID, "SUP-LIFE-001", "Lifecycle Product", 100, 150, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, purchase_date, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, purchaseID, "PUR-SUP-LIFE", supplierID, now, 100, 0, "received", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, created_at) VALUES (?, ?, ?, ?, ?, ?)`, purchaseItemID, purchaseID, productID, 1, 100, now); err != nil {
		t.Fatal(err)
	}
	itemCode := "ITM-" + purchaseID.String()[:8] + "-001"
	if _, err := xdb.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, inventoryItemID, productID, itemCode, "SUP-LIFE-BC", "NEW", 100, 150, "AVAILABLE", supplierID, now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, uuid.New(), productID, 1, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO supplier_ledger (id, supplier_id, type, transaction_type, amount, balance, description, reference_id, created_at) VALUES (?, ?, 'debit', 'PURCHASE', ?, ?, 'Purchase', ?, ?)`, uuid.New(), supplierID, 100, 100, purchaseID, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(xdb), xdb)
	reference := "PAY-LIFE-001"
	partial, err := service.AddPayment(ctx, supplierID, &PaymentRequest{Amount: 40, Method: "cash", Reference: &reference})
	if err != nil {
		t.Fatalf("partial payment: %v", err)
	}
	if partial.Amount != 40 {
		t.Fatalf("partial payment amount = %v", partial.Amount)
	}
	assertSupplierBalance(t, xdb, supplierID, 60)

	if _, err := service.AddPayment(ctx, supplierID, &PaymentRequest{Amount: 40, Method: "cash", Reference: &reference}); err != ErrPaymentDuplicate {
		t.Fatalf("duplicate payment error = %v, want ErrPaymentDuplicate", err)
	}
	assertSupplierBalance(t, xdb, supplierID, 60)

	fullReference := "PAY-LIFE-002"
	if _, err := service.AddPayment(ctx, supplierID, &PaymentRequest{Amount: 60, Method: "cash", Reference: &fullReference}); err != nil {
		t.Fatalf("full payment: %v", err)
	}
	assertSupplierBalance(t, xdb, supplierID, 0)

	suppliers, total, err := service.ListSuppliers(ctx, 1, 20, "", func() *bool { value := true; return &value }())
	if err != nil {
		t.Fatalf("list suppliers after full payment: %v", err)
	}
	if total != 1 || len(suppliers) != 1 {
		t.Fatalf("supplier list = %d items, total %d; want one supplier", len(suppliers), total)
	}
	if suppliers[0].PaidAmount != 100 || suppliers[0].Outstanding != 0 {
		t.Fatalf("supplier summary = paid %.2f outstanding %.2f; want paid 100 outstanding 0", suppliers[0].PaidAmount, suppliers[0].Outstanding)
	}

	if _, err := service.AddPayment(ctx, supplierID, &PaymentRequest{Amount: 1, Method: "cash", Reference: func() *string { value := "PAY-LIFE-003"; return &value }()}); err != ErrPaymentExceedsBalance {
		t.Fatalf("overpayment error = %v, want ErrPaymentExceedsBalance", err)
	}

	returnService := supplierreturns.NewService(xdb)
	createdReturn, err := returnService.Create(ctx, uuid.New(), supplierreturns.CreateRequest{PurchaseID: purchaseID, Reason: "defective"})
	if err != nil {
		t.Fatalf("create supplier return: %v", err)
	}
	if err := returnService.AddItem(ctx, createdReturn.ID, supplierreturns.AddItemRequest{PurchaseItemID: purchaseItemID, Quantity: 1}); err != nil {
		t.Fatalf("add supplier return item: %v", err)
	}
	if err := returnService.Complete(ctx, createdReturn.ID, uuid.Nil); err != nil {
		t.Fatalf("complete supplier return: %v", err)
	}
	assertSupplierBalance(t, xdb, supplierID, -100)

	var inventoryStatus string
	if err := xdb.Get(&inventoryStatus, `SELECT status FROM inventory_items WHERE id = ?`, inventoryItemID); err != nil {
		t.Fatal(err)
	}
	if inventoryStatus != "RETURNED" {
		t.Fatalf("inventory status = %q, want RETURNED", inventoryStatus)
	}

	var ledgerCount int
	if err := xdb.Get(&ledgerCount, `SELECT COUNT(*) FROM supplier_ledger WHERE supplier_id = ?`, supplierID); err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 4 {
		t.Fatalf("ledger entries = %d, want purchase + two payments + supplier return", ledgerCount)
	}

	var ledgerBalance float64
	if err := xdb.Get(&ledgerBalance, `SELECT balance FROM supplier_ledger WHERE supplier_id = ? AND transaction_type = 'SUPPLIER_RETURN' LIMIT 1`, supplierID); err != nil {
		t.Fatal(err)
	}
	if ledgerBalance != -100 {
		t.Fatalf("final ledger balance = %v, want -100", ledgerBalance)
	}
	ledger, err := service.GetSupplierLedger(ctx, supplierID)
	if err != nil {
		t.Fatalf("get supplier ledger: %v", err)
	}
	if ledger.CurrentBalance != 0 || ledger.CreditBalance != 100 {
		t.Fatalf("computed ledger balance/credit = %v/%v, want 0/100", ledger.CurrentBalance, ledger.CreditBalance)
	}
}

func TestSupplierInventoryCountsSQLiteStatusesCaseInsensitively(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "supplier-inventory-case.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlite")
	for _, statement := range []string{
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, code TEXT, name TEXT, email TEXT, phone TEXT, address TEXT, city TEXT, country TEXT, tax_id TEXT, payment_terms TEXT, credit_limit REAL DEFAULT 0, current_balance REAL DEFAULT 0, notes TEXT, is_active INTEGER DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE products (id TEXT PRIMARY KEY, sku TEXT, name TEXT, cost_price REAL DEFAULT 0, selling_price REAL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE purchases (id TEXT PRIMARY KEY, purchase_number TEXT, supplier_id TEXT, total_amount REAL, paid_amount REAL, status TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, item_code TEXT, barcode TEXT, condition TEXT, purchase_cost REAL, selling_price REAL, status TEXT, supplier_id TEXT, purchase_date TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, purchase_id TEXT, supplier_id TEXT, return_number TEXT UNIQUE, reason TEXT, notes TEXT, refund_amount REAL DEFAULT 0, status TEXT, created_by TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE supplier_return_items (id TEXT PRIMARY KEY, supplier_return_id TEXT, purchase_item_id TEXT, product_id TEXT, inventory_item_id TEXT, quantity INTEGER, unit_cost REAL, created_at TEXT)`,
		`CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL, description TEXT, reference_id TEXT, created_at TEXT)`,
	} {
		if _, err := xdb.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	supplierID := uuid.New()
	productID := uuid.New()
	inventoryItemID := uuid.New()
	returnID := uuid.New()
	if _, err := xdb.Exec(`INSERT INTO suppliers (id, code, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, supplierID, "SUP-CASE", "Case Supplier", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO products (id, sku, name, cost_price, selling_price, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, productID, "CASE-001", "Case Product", 100, 150, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO inventory_items (id, product_id, item_code, barcode, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, inventoryItemID, productID, "ITM-CASE-001", "BAR-CASE-001", "NEW", 100, 150, "available", supplierID, now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO supplier_returns (id, purchase_id, supplier_id, return_number, reason, refund_amount, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, returnID, uuid.New(), supplierID, "SRET-CASE-001", "bad item", 2.0, "completed", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := xdb.Exec(`INSERT INTO supplier_return_items (id, supplier_return_id, purchase_item_id, product_id, inventory_item_id, quantity, unit_cost, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, uuid.New(), returnID, uuid.New(), productID, inventoryItemID, 2, 100, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(xdb), xdb)
	items, err := service.GetSupplierInventory(context.Background(), supplierID)
	if err != nil {
		t.Fatalf("GetSupplierInventory returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].Available != 1 {
		t.Fatalf("available = %d, want 1", items[0].Available)
	}
	if items[0].Returned != 2 {
		t.Fatalf("returned = %d, want 2", items[0].Returned)
	}
}

func assertSupplierBalance(t *testing.T, db *sqlx.DB, supplierID uuid.UUID, want float64) {
	t.Helper()
	var balance float64
	if err := db.Get(&balance, `SELECT current_balance FROM suppliers WHERE id = ?`, supplierID); err != nil {
		t.Fatal(err)
	}
	if balance != want {
		t.Fatalf("supplier balance = %v, want %v", balance, want)
	}
}

func TestSupplierLedgerAdjustmentReconcilesProjectionSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "supplier-adjustment.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	xdb := sqlx.NewDb(db, "sqlite")
	for _, statement := range []string{
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, code TEXT, name TEXT, email TEXT, phone TEXT, address TEXT, city TEXT, country TEXT, tax_id TEXT, payment_terms TEXT, credit_limit REAL DEFAULT 0, current_balance REAL DEFAULT 0, notes TEXT, is_active INTEGER DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE purchases (id TEXT PRIMARY KEY, supplier_id TEXT, total_amount REAL, paid_amount REAL, status TEXT, created_at TEXT)`,
		`CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, supplier_id TEXT, refund_amount REAL DEFAULT 0, status TEXT)`,
		`CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL, description TEXT, reference_id TEXT, created_at TEXT)`,
	} {
		if _, err := xdb.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}

	ctx := context.Background()
	supplierID := uuid.New()
	referenceID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := xdb.Exec(`INSERT INTO suppliers (id, code, name, credit_limit, current_balance, created_at, updated_at) VALUES (?, ?, ?, ?, 0, ?, ?)`, supplierID, "SUP-ADJ", "Adjustment Supplier", 500, now, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(xdb), xdb)
	if err := service.CreateDebtEntry(ctx, supplierID, 50, referenceID, "MANUAL_ADJUSTMENT", time.Now().AddDate(0, 0, 30)); err != nil {
		t.Fatalf("create supplier adjustment: %v", err)
	}
	assertSupplierBalance(t, xdb, supplierID, 50)

	var ledgerCount int
	if err := xdb.Get(&ledgerCount, `SELECT COUNT(*) FROM supplier_ledger WHERE supplier_id = ? AND reference_id = ?`, supplierID, referenceID); err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 1 {
		t.Fatalf("adjustment ledger entries = %d, want 1", ledgerCount)
	}

	var ledgerBalance float64
	if err := xdb.Get(&ledgerBalance, `SELECT balance FROM supplier_ledger WHERE supplier_id = ? AND reference_id = ?`, supplierID, referenceID); err != nil {
		t.Fatal(err)
	}
	if ledgerBalance != 50 {
		t.Fatalf("adjustment ledger balance = %v, want 50", ledgerBalance)
	}
}
