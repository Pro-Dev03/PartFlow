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
