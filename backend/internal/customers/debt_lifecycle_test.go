package customers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
	customerreturns "github.com/partflow/smart-store/internal/returns"
)

func TestAddOpeningDebtSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/opening-debt.db")
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()

	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	customerID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO customers (id, code, name, credit_limit, current_balance, created_at, updated_at) VALUES (?, ?, ?, ?, 0, ?, ?)`, customerID, "C-OPENING-DEBT", "Opening Debt Customer", 1000, now, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db))
	if err := service.AddOpeningDebt(ctx, customerID, 1500); err != nil {
		t.Fatalf("add opening debt: %v", err)
	}

	var debt struct {
		Amount          float64 `db:"amount"`
		RemainingAmount float64 `db:"remaining_amount"`
		Notes           string  `db:"notes"`
	}
	if err := db.Get(&debt, `SELECT amount, remaining_amount, notes FROM debts WHERE customer_id = ?`, customerID); err != nil {
		t.Fatalf("read opening debt: %v", err)
	}
	if debt.Amount != 1500 || debt.RemainingAmount != 1500 || debt.Notes != "opening_debt" {
		t.Fatalf("opening debt = %+v, want amount and remaining 1500 with opening_debt note", debt)
	}

	var ledgerCount int
	if err := db.Get(&ledgerCount, `SELECT COUNT(*) FROM customer_ledger WHERE customer_id = ? AND type = 'debit' AND amount = 1500`, customerID); err != nil {
		t.Fatalf("read opening debt ledger: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("opening debt ledger count = %d, want 1", ledgerCount)
	}

	var balance float64
	if err := db.Get(&balance, `SELECT current_balance FROM customers WHERE id = ?`, customerID); err != nil {
		t.Fatalf("read opening debt balance: %v", err)
	}
	if balance != 1500 {
		t.Fatalf("customer balance = %v, want 1500", balance)
	}
}

func TestCustomerDebtLifecyclePaymentAndReturnCreditSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/customer-debt-lifecycle.db")
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()

	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	customerID, saleID, productID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := db.Exec(`INSERT INTO customers (id, code, name, credit_limit, current_balance, created_at, updated_at) VALUES (?, ?, ?, ?, 0, ?, ?)`, customerID, "C-DEBT-LIFE", "Debt Lifecycle Customer", 1000, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, productID, "DEBT-LIFE-001", "Debt Lifecycle Product", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id, sale_number, invoice_number, customer_id, sale_date, total_amount, paid_amount, remaining_amount, payment_method, payment_status, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 0, ?, 'credit', 'debt', 'completed', ?, ?)`, saleID, "SALE-DEBT-LIFE", "INV-DEBT-LIFE", customerID, now, 200, 200, now, now); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewRepository(db))
	if err := service.CreateDebtEntry(ctx, customerID, 200, saleID, "sale", time.Now().AddDate(0, 0, 30)); err != nil {
		t.Fatalf("create credit-sale debt: %v", err)
	}
	var debtID uuid.UUID
	if err := db.Get(&debtID, `SELECT id FROM debts WHERE customer_id = ? AND sale_id = ?`, customerID, saleID); err != nil {
		t.Fatalf("read created debt: %v", err)
	}
	assertCustomerDebtBalance(t, db, customerID, 200)

	partialReference := "CUSTOMER-PAY-001"
	if err := service.ProcessDebtPaymentWithReference(ctx, customerID, 75, "cash", &partialReference); err != nil {
		t.Fatalf("partial payment: %v", err)
	}
	assertCustomerDebtBalance(t, db, customerID, 125)
	assertDebtRemaining(t, db, debtID, 125)

	if err := service.ProcessDebtPaymentWithReference(ctx, customerID, 75, "cash", &partialReference); err != ErrPaymentDuplicate {
		t.Fatalf("duplicate payment error = %v, want ErrPaymentDuplicate", err)
	}
	assertCustomerDebtBalance(t, db, customerID, 125)
	assertDebtRemaining(t, db, debtID, 125)

	overpaymentReference := "CUSTOMER-PAY-OVER"
	if err := service.ProcessDebtPaymentWithReference(ctx, customerID, 126, "cash", &overpaymentReference); err == nil {
		t.Fatal("expected overpayment rejection")
	}
	assertCustomerDebtBalance(t, db, customerID, 125)
	assertDebtRemaining(t, db, debtID, 125)

	fullReference := "CUSTOMER-PAY-002"
	if err := service.ProcessDebtPaymentWithReference(ctx, customerID, 125, "cash", &fullReference); err != nil {
		t.Fatalf("full payment: %v", err)
	}
	assertCustomerDebtBalance(t, db, customerID, 0)
	assertDebtRemaining(t, db, debtID, 0)

	returnService := customerreturns.NewService(customerreturns.NewRepository(db))
	returnRecord, err := returnService.CreateReturn(ctx, uuid.New(), &customerreturns.ReturnRequest{
		SaleID:                   &saleID,
		CustomerID:               &customerID,
		ReturnDate:               time.Now().UTC(),
		ReturnType:               "PARTIAL",
		Reason:                   "CUSTOMER_CHANGED_MIND",
		ItemConditionAfterReturn: "READY_FOR_SALE",
		RefundMethod:             "DEBT_ADJUSTMENT",
		DebtID:                   &debtID,
		Items: []customerreturns.ReturnItemRequest{{
			ProductID:         &productID,
			QuantityReturned:  1,
			UnitPrice:         150,
			TotalRefundAmount: 150,
			ReturnedCondition: "DAMAGED",
			Resolution:        "REPAIR",
		}},
	})
	if err != nil {
		t.Fatalf("create customer return: %v", err)
	}
	if _, err := returnService.ApproveReturn(ctx, returnRecord.Return.ID); err != nil {
		t.Fatalf("approve customer return: %v", err)
	}
	completed, err := returnService.CompleteReturn(ctx, returnRecord.Return.ID, uuid.New())
	if err != nil {
		t.Fatalf("complete customer return: %v", err)
	}
	if completed.Return.DebtAdjustment != 0 || completed.Return.CustomerCredit != 150 {
		t.Fatalf("return adjustment = %v, customer credit = %v; want 0 and 150", completed.Return.DebtAdjustment, completed.Return.CustomerCredit)
	}
	assertCustomerDebtBalance(t, db, customerID, -150)
	assertDebtRemaining(t, db, debtID, 0)

	if _, err := returnService.CompleteReturn(ctx, returnRecord.Return.ID, uuid.New()); err == nil {
		t.Fatal("expected duplicate debt adjustment/return completion rejection")
	}
	if err := customerreturns.NewRepository(db).AddDebtAdjustmentLedgerEntry(ctx, &completed.Return); err != nil {
		t.Fatalf("repeat debt adjustment ledger application: %v", err)
	}

	var paymentCount int
	if err := db.Get(&paymentCount, `SELECT COUNT(*) FROM payments WHERE customer_id = ?`, customerID); err != nil {
		t.Fatal(err)
	}
	if paymentCount != 2 {
		t.Fatalf("customer payment rows = %d, want 2", paymentCount)
	}
	var returnCreditPayments int
	if err := db.Get(&returnCreditPayments, `SELECT COUNT(*) FROM payments WHERE customer_id = ? AND amount = 150`, customerID); err != nil {
		t.Fatal(err)
	}
	if returnCreditPayments != 0 {
		t.Fatalf("customer credit was recorded as %d payment(s), want 0", returnCreditPayments)
	}

	var ledgerCount int
	if err := db.Get(&ledgerCount, `SELECT COUNT(*) FROM customer_ledger WHERE customer_id = ?`, customerID); err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 4 {
		t.Fatalf("customer ledger entries = %d, want debit + two payments + return credit", ledgerCount)
	}

	var ledgerBalance float64
	if err := db.Get(&ledgerBalance, `SELECT balance FROM customer_ledger WHERE customer_id = ? AND reference_id = ? AND type = 'credit'`, customerID, returnRecord.Return.ID); err != nil {
		t.Fatal(err)
	}
	if ledgerBalance != -150 {
		t.Fatalf("customer return ledger balance = %v, want -150", ledgerBalance)
	}
}

func TestDeleteCustomerArchivesAndPreservesFinancialRowsSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/delete-customer-cascade.db")
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()

	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	customerID := uuid.New()
	saleID := uuid.New()
	productID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := db.Exec(`INSERT INTO customers (id, code, name, credit_limit, current_balance, created_at, updated_at) VALUES (?, ?, ?, ?, 200, ?, ?)`, customerID, "C-DELETE-CASCADE", "Delete Cascade Customer", 1000, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id, sku, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, productID, "DELETE-CASCADE-001", "Delete Cascade Product", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id, sale_number, customer_id, total_amount, tax_amount, discount_amount, paid_amount, cash_received, change_amount, remaining_amount, payment_method, status, notes, created_at, updated_at) VALUES (?, ?, ?, 200, 0, 0, 0, 0, 0, 200, 'cash', 'completed', 'cascade delete test', ?, ?)`, saleID, "SALE-DELETE-CASCADE", customerID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO debts (id, customer_id, sale_id, amount, paid_amount, remaining_amount, due_date, status, notes, created_at, updated_at) VALUES (?, ?, ?, 200, 0, 200, ?, 'pending', 'cascade delete test', ?, ?)`, uuid.New(), customerID, saleID, time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339), now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO payments (id, transaction_number, customer_id, amount, payment_method, reference, notes, created_at) VALUES (?, ?, ?, 50, 'cash', ?, ?, ?)`, uuid.New(), "PAY-DELETE-CASCADE", customerID, "REF-DELETE-CASCADE", "cascade delete test", now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_ledger (id, customer_id, type, transaction_type, amount, balance, description, reference_id, reference_type, created_by, created_at) VALUES (?, ?, 'debit', 'sale', 200, 200, 'cascade delete test', ?, 'sale', 'system', ?)`, uuid.New(), customerID, saleID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_payments (id, customer_id, amount, payment_date, method, reference, notes, created_at) VALUES (?, ?, 50, ?, 'cash', ?, ?, ?)`, uuid.New(), customerID, now, "REF-DELETE-CASCADE", "cascade delete test", now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db))
	if err := service.DeleteCustomer(ctx, customerID); err != nil {
		t.Fatalf("delete customer: %v", err)
	}

	var isActive bool
	if err := db.Get(&isActive, `SELECT is_active FROM customers WHERE id = ?`, customerID); err != nil {
		t.Fatal(err)
	}
	if isActive {
		t.Fatal("customer remains active after archive")
	}

	for _, query := range []string{
		`SELECT COUNT(*) FROM sales WHERE customer_id = ?`,
		`SELECT COUNT(*) FROM debts WHERE customer_id = ?`,
		`SELECT COUNT(*) FROM payments WHERE customer_id = ?`,
		`SELECT COUNT(*) FROM customer_ledger WHERE customer_id = ?`,
		`SELECT COUNT(*) FROM customer_payments WHERE customer_id = ?`,
	} {
		var count int
		if err := db.Get(&count, query, customerID); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("query %s has %d rows after archive, want 1", query, count)
		}
	}
}

func TestDeleteCustomerPreservesSupplierReturnReferencesSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", t.TempDir()+"/delete-customer-supplier-returns.db")
	database, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer database.DB.Close()

	db := sqlx.NewDb(database.DB, "sqlite")
	ctx := context.Background()
	customerID := uuid.New()
	returnID := uuid.New()
	supplierID := uuid.New()
	saleID := uuid.New()
	purchaseID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := db.Exec(`INSERT INTO customers (id, code, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, customerID, "C-SUPPLIER-RET", "Supplier Return Customer", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales (id, sale_number, customer_id, total_amount, remaining_amount, payment_method, status, created_at, updated_at) VALUES (?, ?, ?, 200, 200, 'cash', 'completed', ?, ?)`, saleID, "SALE-SUPPLIER-RET", customerID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO suppliers (id, code, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, supplierID, "SUP-RET-001", "Supplier Return Vendor", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO purchases (id, purchase_number, supplier_id, total_amount, paid_amount, status, created_at, updated_at) VALUES (?, ?, ?, 50, 0, 'received', ?, ?)`, purchaseID, "PUR-SUPPLIER-RET-001", supplierID, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO returns (id, return_number, customer_id, sale_id, total_refund_amount, refund_status, status, reason, return_date, created_at, updated_at) VALUES (?, ?, ?, ?, 50, 'pending', 'pending', 'customer_changed_mind', ?, ?, ?)`, returnID, "RET-SUPPLIER-001", customerID, saleID, now, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO supplier_returns (id, customer_return_id, sale_id, purchase_id, supplier_id, return_number, status, source_status, reason, refund_amount, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 'PENDING', 'RESOLVED', 'customer_return', 50, ?, ?)`, uuid.New(), returnID, saleID, purchaseID, supplierID, "SRET-SUPPLIER-001", now, now); err != nil {
		t.Fatal(err)
	}

	service := NewService(NewRepository(db))
	if err := service.DeleteCustomer(ctx, customerID); err != nil {
		t.Fatalf("delete customer with supplier return references: %v", err)
	}

	var returnCount int
	if err := db.Get(&returnCount, `SELECT COUNT(*) FROM returns WHERE id = ?`, returnID); err != nil {
		t.Fatal(err)
	}
	if returnCount != 1 {
		t.Fatalf("customer return count after archive = %d, want 1", returnCount)
	}
	var supplierReturnCount int
	if err := db.Get(&supplierReturnCount, `SELECT COUNT(*) FROM supplier_returns WHERE customer_return_id = ?`, returnID); err != nil {
		t.Fatal(err)
	}
	if supplierReturnCount != 1 {
		t.Fatalf("supplier return count after archive = %d, want 1", supplierReturnCount)
	}
}

func assertCustomerDebtBalance(t *testing.T, db *sqlx.DB, customerID uuid.UUID, want float64) {
	t.Helper()
	var balance float64
	if err := db.Get(&balance, `SELECT current_balance FROM customers WHERE id = ?`, customerID); err != nil {
		t.Fatal(err)
	}
	if balance != want {
		t.Fatalf("customer balance = %v, want %v", balance, want)
	}
}

func assertDebtRemaining(t *testing.T, db *sqlx.DB, debtID uuid.UUID, want float64) {
	t.Helper()
	var remaining float64
	if err := db.Get(&remaining, `SELECT remaining_amount FROM debts WHERE id = ?`, debtID); err != nil {
		t.Fatal(err)
	}
	if remaining != want {
		t.Fatalf("debt remaining = %v, want %v", remaining, want)
	}
	if remaining < 0 {
		t.Fatalf("debt remaining became negative: %v", remaining)
	}
}
