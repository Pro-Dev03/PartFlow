package payments

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func openPaymentDeleteDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	for _, statement := range []string{
		`CREATE TABLE payments (id TEXT PRIMARY KEY, amount REAL NOT NULL DEFAULT 0, customer_id TEXT, supplier_id TEXT, sale_id TEXT, payment_date TEXT, created_at TEXT, payment_status TEXT)`,
		`CREATE TABLE payment_allocation_batches (payment_id TEXT PRIMARY KEY, owner_type TEXT NOT NULL, owner_id TEXT NOT NULL, sale_id TEXT, tracked_at TEXT)`,
		`CREATE TABLE payment_debt_allocations (id TEXT PRIMARY KEY, payment_id TEXT NOT NULL, debt_id TEXT NOT NULL, amount REAL NOT NULL, created_at TEXT, UNIQUE(payment_id, debt_id))`,
		`CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL, description TEXT, reference_id TEXT, created_at TEXT)`,
		`CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL, description TEXT, reference_id TEXT, created_at TEXT)`,
		`CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL, updated_at TEXT)`,
		`CREATE TABLE suppliers (id TEXT PRIMARY KEY, current_balance REAL, updated_at TEXT)`,
		`CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, paid_amount REAL, remaining_amount REAL, due_date TEXT, status TEXT, sale_id TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE customer_debts (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, paid_amount REAL, is_paid INTEGER)`,
		`CREATE TABLE supplier_debts (id TEXT PRIMARY KEY, supplier_id TEXT, amount REAL, reference_id TEXT, reference_type TEXT, due_date TEXT, paid_amount REAL, is_paid INTEGER, created_at TEXT)`,
		`CREATE TABLE sales (id TEXT PRIMARY KEY, customer_id TEXT, total_amount REAL, paid_amount REAL, remaining_amount REAL, payment_status TEXT, updated_at TEXT)`,
		`CREATE TABLE daily_debt_summary (date TEXT PRIMARY KEY, total_debt REAL DEFAULT 0, new_debt REAL DEFAULT 0, payments_received REAL DEFAULT 0, overdue_debt REAL DEFAULT 0, overdue_count INTEGER DEFAULT 0, paid_debt REAL DEFAULT 0, updated_at TEXT)`,
		`CREATE TABLE monthly_debt_summary (year INTEGER, month INTEGER, total_debt REAL DEFAULT 0, new_debt REAL DEFAULT 0, payments_received REAL DEFAULT 0, overdue_debt REAL DEFAULT 0, overdue_count INTEGER DEFAULT 0, paid_debt REAL DEFAULT 0, updated_at TEXT, PRIMARY KEY(year, month))`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestDeletePaymentDeletesUnpostedPendingAndGenericCompletedRowsSQLite(t *testing.T) {
	db := openPaymentDeleteDB(t)
	pendingID := uuid.New()
	completedID := uuid.New()
	if _, err := db.Exec(`INSERT INTO payments (id, payment_status) VALUES (?, 'pending'), (?, 'completed')`, pendingID, completedID); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)

	if err := repo.Delete(context.Background(), completedID); err != nil {
		t.Fatalf("Delete(completed generic payment): %v", err)
	}
	var completedCount int
	if err := db.Get(&completedCount, `SELECT COUNT(*) FROM payments WHERE id = ?`, completedID); err != nil {
		t.Fatal(err)
	}
	if completedCount != 0 {
		t.Fatalf("completed generic payment rows = %d, want hard delete", completedCount)
	}

	if err := repo.Delete(context.Background(), pendingID); err != nil {
		t.Fatalf("Delete(pending): %v", err)
	}
	var pendingCount int
	if err := db.Get(&pendingCount, `SELECT COUNT(*) FROM payments WHERE id = ?`, pendingID); err != nil {
		t.Fatal(err)
	}
	if pendingCount != 0 {
		t.Fatalf("pending payment rows = %d, want 0", pendingCount)
	}
}

func TestDeleteCustomerPaymentReversesDebtLedgerSaleAndBalanceSQLite(t *testing.T) {
	db := openPaymentDeleteDB(t)
	ctx := context.Background()
	customerID, debtID, saleID, paymentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	created := time.Date(2026, 9, 25, 10, 0, 0, 0, time.UTC).Format(time.RFC3339)
	for _, item := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO customers VALUES (?, 60, ?)`, []any{customerID.String(), created}},
		{`INSERT INTO debts VALUES (?, ?, 100, 40, 60, '2026-10-01', 'partial', NULL, ?, ?)`, []any{debtID.String(), customerID.String(), created, created}},
		{`INSERT INTO customer_debts VALUES (?, ?, 100, 40, 0)`, []any{debtID.String(), customerID.String()}},
		{`INSERT INTO sales VALUES (?, ?, 100, 40, 60, 'partial', ?)`, []any{saleID.String(), customerID.String(), created}},
		{`INSERT INTO payments VALUES (?, 40, ?, NULL, ?, ?, ?, 'completed')`, []any{paymentID.String(), customerID.String(), saleID.String(), created, created}},
		{`INSERT INTO payment_allocation_batches VALUES (?, 'customer', ?, ?, ?)`, []any{paymentID.String(), customerID.String(), saleID.String(), created}},
		{`INSERT INTO payment_debt_allocations VALUES (?, ?, ?, 40, ?)`, []any{uuid.New().String(), paymentID.String(), debtID.String(), created}},
		{`INSERT INTO customer_ledger VALUES ('sale-ledger', ?, 'debit', 'SALE', 100, 100, 'Sale', ?, ?)`, []any{customerID.String(), saleID.String(), created}},
		{`INSERT INTO customer_ledger VALUES ('payment-ledger', ?, 'credit', 'PAYMENT', 40, 60, 'Payment', ?, ?)`, []any{customerID.String(), paymentID.String(), created}},
		{`INSERT INTO customer_ledger VALUES ('adjustment-ledger', ?, 'debit', 'ADJUSTMENT', 5, 65, 'Adjustment', 'adjustment', ?)`, []any{customerID.String(), created}},
		{`INSERT INTO daily_debt_summary (date, payments_received, updated_at) VALUES ('2026-09-25', 40, ?)`, []any{created}},
		{`INSERT INTO monthly_debt_summary (year, month, payments_received, updated_at) VALUES (2026, 9, 40, ?)`, []any{created}},
	} {
		if _, err := db.Exec(item.query, item.args...); err != nil {
			t.Fatalf("fixture insert %q: %v", item.query, err)
		}
	}

	if err := NewRepository(db).Delete(ctx, paymentID); err != nil {
		t.Fatalf("Delete(customer payment): %v", err)
	}
	var currentBalance, paid, remaining, salePaid, saleRemaining float64
	if err := db.Get(&currentBalance, `SELECT current_balance FROM customers WHERE id = ?`, customerID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&paid, `SELECT paid_amount FROM debts WHERE id = ?`, debtID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&remaining, `SELECT remaining_amount FROM debts WHERE id = ?`, debtID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&salePaid, `SELECT paid_amount FROM sales WHERE id = ?`, saleID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&saleRemaining, `SELECT remaining_amount FROM sales WHERE id = ?`, saleID.String()); err != nil {
		t.Fatal(err)
	}
	if math.Abs(currentBalance-100) > paymentDeleteTolerance || math.Abs(paid) > paymentDeleteTolerance || math.Abs(remaining-100) > paymentDeleteTolerance || math.Abs(salePaid) > paymentDeleteTolerance || math.Abs(saleRemaining-100) > paymentDeleteTolerance {
		t.Fatalf("reversal mismatch: customer=%v debt=(%v,%v) sale=(%v,%v)", currentBalance, paid, remaining, salePaid, saleRemaining)
	}
	var ledgerCount int
	if err := db.Get(&ledgerCount, `SELECT COUNT(*) FROM customer_ledger`); err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 2 {
		t.Fatalf("ledger entries = %d, want sale and adjustment only", ledgerCount)
	}
	var lastBalance float64
	if err := db.Get(&lastBalance, `SELECT balance FROM customer_ledger ORDER BY created_at DESC, id DESC LIMIT 1`); err != nil {
		t.Fatal(err)
	}
	if math.Abs(lastBalance-105) > paymentDeleteTolerance {
		t.Fatalf("recalculated ledger balance = %v, want 105", lastBalance)
	}
	var summaryPayments float64
	if err := db.Get(&summaryPayments, `SELECT payments_received FROM daily_debt_summary WHERE date = '2026-09-25'`); err != nil {
		t.Fatal(err)
	}
	if math.Abs(summaryPayments) > paymentDeleteTolerance {
		t.Fatalf("daily payments_received = %v, want 0", summaryPayments)
	}
	var paymentCount int
	if err := db.Get(&paymentCount, `SELECT COUNT(*) FROM payments WHERE id = ?`, paymentID.String()); err != nil {
		t.Fatal(err)
	}
	if paymentCount != 0 {
		t.Fatalf("payment rows = %d, want hard-deleted", paymentCount)
	}
}

func TestDeleteSupplierPaymentReversesDebtLedgerAndBalanceSQLite(t *testing.T) {
	db := openPaymentDeleteDB(t)
	ctx := context.Background()
	supplierID, debtID, paymentID := uuid.New(), uuid.New(), uuid.New()
	created := time.Date(2026, 9, 25, 11, 0, 0, 0, time.UTC).Format(time.RFC3339)
	for _, item := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO suppliers VALUES (?, 60, ?)`, []any{supplierID.String(), created}},
		{`INSERT INTO supplier_debts (id, supplier_id, amount, paid_amount, is_paid, due_date, created_at) VALUES (?, ?, 100, 40, 0, '2026-10-01', ?)`, []any{debtID.String(), supplierID.String(), created}},
		{`INSERT INTO payments VALUES (?, 40, NULL, ?, NULL, ?, ?, 'completed')`, []any{paymentID.String(), supplierID.String(), created, created}},
		{`INSERT INTO payment_allocation_batches VALUES (?, 'supplier', ?, NULL, ?)`, []any{paymentID.String(), supplierID.String(), created}},
		{`INSERT INTO payment_debt_allocations VALUES (?, ?, ?, 40, ?)`, []any{uuid.New().String(), paymentID.String(), debtID.String(), created}},
		{`INSERT INTO supplier_ledger VALUES ('purchase-ledger', ?, 'debit', 'PURCHASE', 100, 100, 'Purchase', 'purchase', ?)`, []any{supplierID.String(), created}},
		{`INSERT INTO supplier_ledger VALUES ('payment-ledger', ?, 'credit', 'PAYMENT', 40, 60, 'Payment', ?, ?)`, []any{supplierID.String(), paymentID.String(), created}},
	} {
		if _, err := db.Exec(item.query, item.args...); err != nil {
			t.Fatalf("fixture insert %q: %v", item.query, err)
		}
	}
	if err := NewRepository(db).Delete(ctx, paymentID); err != nil {
		t.Fatalf("Delete(supplier payment): %v", err)
	}
	var currentBalance, paid float64
	var isPaid bool
	if err := db.Get(&currentBalance, `SELECT current_balance FROM suppliers WHERE id = ?`, supplierID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&paid, `SELECT paid_amount FROM supplier_debts WHERE id = ?`, debtID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&isPaid, `SELECT is_paid FROM supplier_debts WHERE id = ?`, debtID.String()); err != nil {
		t.Fatal(err)
	}
	if math.Abs(currentBalance-100) > paymentDeleteTolerance || math.Abs(paid) > paymentDeleteTolerance || isPaid {
		t.Fatalf("supplier reversal mismatch: balance=%v debt paid=%v is_paid=%v", currentBalance, paid, isPaid)
	}
	var ledgerCount int
	if err := db.Get(&ledgerCount, `SELECT COUNT(*) FROM supplier_ledger`); err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 1 {
		t.Fatalf("supplier ledger entries = %d, want only purchase", ledgerCount)
	}
	var paymentCount int
	if err := db.Get(&paymentCount, `SELECT COUNT(*) FROM payments WHERE id = ?`, paymentID.String()); err != nil {
		t.Fatal(err)
	}
	if paymentCount != 0 {
		t.Fatalf("payment rows = %d, want hard-deleted", paymentCount)
	}
}

func TestDeleteLegacyCustomerPaymentReconstructsItsRecordedDebtPortionSQLite(t *testing.T) {
	db := openPaymentDeleteDB(t)
	customerID, paymentID, debtID := uuid.New(), uuid.New(), uuid.New()
	created := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO customers VALUES (?, 10, ?)`, customerID.String(), created); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payments VALUES (?, 10, ?, NULL, NULL, ?, ?, 'completed')`, paymentID.String(), customerID.String(), created, created); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO debts VALUES (?, ?, 100, 40, 60, '2026-10-01', 'partial', NULL, ?, ?)`, debtID.String(), customerID.String(), created, created); err != nil {
		t.Fatal(err)
	}
	if err := NewRepository(db).Delete(context.Background(), paymentID); err != nil {
		t.Fatalf("Delete(legacy payment): %v", err)
	}
	var exists int
	if err := db.Get(&exists, `SELECT COUNT(*) FROM payments WHERE id = ?`, paymentID.String()); err != nil {
		t.Fatal(err)
	}
	var paid, remaining, balance float64
	if err := db.Get(&paid, `SELECT paid_amount FROM debts WHERE id=?`, debtID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&remaining, `SELECT remaining_amount FROM debts WHERE id=?`, debtID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&balance, `SELECT current_balance FROM customers WHERE id=?`, customerID.String()); err != nil {
		t.Fatal(err)
	}
	if exists != 0 || math.Abs(paid-30) > paymentDeleteTolerance || math.Abs(remaining-70) > paymentDeleteTolerance || math.Abs(balance-20) > paymentDeleteTolerance {
		t.Fatalf("legacy payment reversal: remaining payment=%d debt=(%v,%v) customer balance=%v; want 0, (30,70), 20", exists, paid, remaining, balance)
	}
}

func TestDeleteLegacySupplierPaymentReconstructsItsRecordedDebtPortionSQLite(t *testing.T) {
	db := openPaymentDeleteDB(t)
	supplierID, paymentID, debtID := uuid.New(), uuid.New(), uuid.New()
	created := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO suppliers VALUES (?,10,?)`, supplierID.String(), created); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_debts (id, supplier_id, amount, paid_amount, is_paid, due_date, created_at) VALUES (?, ?, 100, 40, 0, '2026-10-01', ?)`, debtID.String(), supplierID.String(), created); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payments VALUES (?,10,NULL,?,NULL,?,?,'completed')`, paymentID.String(), supplierID.String(), created, created); err != nil {
		t.Fatal(err)
	}
	if err := NewRepository(db).Delete(context.Background(), paymentID); err != nil {
		t.Fatalf("Delete(legacy supplier payment): %v", err)
	}
	var exists int
	var paid, balance float64
	var isPaid bool
	if err := db.Get(&exists, `SELECT COUNT(*) FROM payments WHERE id=?`, paymentID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&paid, `SELECT paid_amount FROM supplier_debts WHERE id=?`, debtID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&isPaid, `SELECT is_paid FROM supplier_debts WHERE id=?`, debtID.String()); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&balance, `SELECT current_balance FROM suppliers WHERE id=?`, supplierID.String()); err != nil {
		t.Fatal(err)
	}
	if exists != 0 || math.Abs(paid-30) > paymentDeleteTolerance || isPaid || math.Abs(balance-20) > paymentDeleteTolerance {
		t.Fatalf("legacy supplier payment reversal: payment=%d debt paid=%v is_paid=%v balance=%v; want 0, 30, false, 20", exists, paid, isPaid, balance)
	}
}

func TestDeleteFullyRefundedProviderPaymentHardDeletesLinkedProviderHistorySQLite(t *testing.T) {
	db := openPaymentDeleteDB(t)
	paymentID, providerTransactionID := uuid.New(), uuid.New()
	created := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO payments (id, amount, payment_date, created_at, payment_status) VALUES (?, 25, ?, ?, 'completed')`, paymentID.String(), created, created); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE payment_transactions (id TEXT PRIMARY KEY, payment_id TEXT, status TEXT, amount_minor INTEGER, currency TEXT, created_at TEXT)`,
		`CREATE TABLE payment_refunds (id TEXT PRIMARY KEY, payment_transaction_id TEXT, status TEXT)`,
		`CREATE TABLE payment_webhook_events (id TEXT PRIMARY KEY, payment_transaction_id TEXT)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO payment_transactions VALUES (?, ?, 'refunded', 2500, 'ILS', ?)`, providerTransactionID.String(), paymentID.String(), created); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payment_refunds VALUES ('provider-refund', ?, 'refunded')`, providerTransactionID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payment_webhook_events VALUES ('provider-event', ?)`, providerTransactionID.String()); err != nil {
		t.Fatal(err)
	}
	if err := NewService(NewRepository(db)).DeletePayment(context.Background(), paymentID, uuid.Nil); err != nil {
		t.Fatalf("Delete(refunded provider payment): %v", err)
	}
	for _, table := range []string{"payments", "payment_transactions", "payment_refunds", "payment_webhook_events"} {
		var count int
		if err := db.Get(&count, `SELECT COUNT(*) FROM `+table); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("%s rows=%d, want hard-deleted", table, count)
		}
	}
}

func TestDeletePaymentMissingRowReturnsNotFoundSQLite(t *testing.T) {
	db := openPaymentDeleteDB(t)
	if err := NewRepository(db).Delete(context.Background(), uuid.New()); err != ErrPaymentNotFound {
		t.Fatalf("Delete(missing) error = %v, want ErrPaymentNotFound", err)
	}
}
