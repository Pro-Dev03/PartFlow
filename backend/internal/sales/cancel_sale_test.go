package sales

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func newSaleCancellationTestService(t *testing.T) (*Service, string, string, string) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE sales (id TEXT PRIMARY KEY, invoice_number TEXT, customer_id TEXT, user_id TEXT, created_at TEXT, status TEXT, total_amount REAL, paid_amount REAL DEFAULT 0, payment_status TEXT, updated_at TEXT);
		CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, quantity INTEGER);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT, status TEXT, sold_at TEXT, updated_at TEXT);
		CREATE TABLE inventory_movements (id TEXT PRIMARY KEY, item_id TEXT, product_id TEXT, movement_type TEXT, quantity INTEGER, before_quantity INTEGER, after_quantity INTEGER, reference_type TEXT, reference_id TEXT, reason TEXT, created_by TEXT, created_at TEXT);
		CREATE TABLE inventory (id TEXT PRIMARY KEY, product_id TEXT UNIQUE, quantity INTEGER, created_at TEXT, updated_at TEXT);
		CREATE TABLE acquisition_items (id TEXT PRIMARY KEY, inventory_item_id TEXT, item_status TEXT, updated_at TEXT);
		CREATE TABLE returns (id TEXT PRIMARY KEY, sale_id TEXT);
		CREATE TABLE payment_transactions (id TEXT PRIMARY KEY, sale_id TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, sale_id TEXT);
		CREATE TABLE sale_payment_allocations (id TEXT PRIMARY KEY, sale_id TEXT);
		CREATE TABLE debts (id TEXT PRIMARY KEY, sale_id TEXT, paid_amount REAL, remaining_amount REAL, status TEXT, updated_at TEXT);
		CREATE TABLE customer_debts (id TEXT PRIMARY KEY, customer_id TEXT, reference_id TEXT, reference_type TEXT, amount REAL, paid_amount REAL DEFAULT 0, is_paid INTEGER DEFAULT 0);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, type TEXT, amount REAL, balance REAL, reference_id TEXT, description TEXT, created_at TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL, updated_at TEXT);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, old_values TEXT, new_values TEXT, created_at TEXT);
		CREATE TABLE pos_shifts (id TEXT PRIMARY KEY, user_id TEXT, opened_at TEXT, closed_at TEXT, sales_total REAL, sale_count INTEGER);
		CREATE TABLE item_history (id TEXT PRIMARY KEY, inventory_item_id TEXT, event_type TEXT, event_date TEXT, reference_type TEXT, reference_id TEXT, description TEXT, metadata TEXT, created_by TEXT, created_at TEXT);
	`)
	if err != nil {
		t.Fatal(err)
	}

	saleID, productID, itemID, customerID, cashierID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	_, err = db.Exec(`INSERT INTO sales (id,invoice_number,customer_id,user_id,created_at,status,total_amount,paid_amount,payment_status,updated_at) VALUES (?, 'INV-CANCEL-01', ?, ?, CURRENT_TIMESTAMP, 'completed', 120, 0, 'debt', CURRENT_TIMESTAMP)`, saleID, customerID, cashierID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO sale_items VALUES (?, ?, ?, 1)`, uuid.NewString(), saleID, productID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory_items VALUES (?, ?, 'SOLD', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, itemID, productID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory_movements VALUES (?, ?, NULL, 'SALE', -1, 2, 1, 'sale', ?, 'Sale: INV-CANCEL-01', NULL, CURRENT_TIMESTAMP)`, uuid.NewString(), itemID, saleID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO inventory VALUES (?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, uuid.NewString(), productID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO acquisition_items VALUES (?, ?, 'sold', CURRENT_TIMESTAMP)`, uuid.NewString(), itemID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO customers VALUES (?, 120, CURRENT_TIMESTAMP)`, customerID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO customer_ledger VALUES (?, ?, 'debit', 120, 120, ?, 'Sale: INV-CANCEL-01', CURRENT_TIMESTAMP)`, uuid.NewString(), customerID, saleID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO debts VALUES (?, ?, 0, 120, 'pending', CURRENT_TIMESTAMP)`, uuid.NewString(), saleID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO customer_debts VALUES (?, ?, ?, 'sale', 120, 0, 0)`, uuid.NewString(), customerID, saleID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO pos_shifts VALUES (?, ?, '2000-01-01 00:00:00', NULL, 120, 1)`, uuid.NewString(), cashierID)
	if err != nil {
		t.Fatal(err)
	}

	return &Service{db: sqlx.NewDb(db, "sqlite")}, saleID, itemID, customerID
}

func TestCancelUnpaidSaleReversesStockDebtAndRecordsHistoryAtomically(t *testing.T) {
	service, saleID, itemID, customerID := newSaleCancellationTestService(t)
	userID := uuid.New()
	if err := service.CancelSale(context.Background(), userID, uuid.MustParse(saleID)); err != nil {
		t.Fatal(err)
	}

	var status, paymentStatus, itemStatus string
	var stock int
	var balance, reversalAmount float64
	var debtStatus string
	var debtRemaining float64
	var customerDebtCount, cancellationMovementCount, auditCount, itemHistoryCount int
	var shiftTotal float64
	var shiftSalesCount int
	db := service.db.DB
	if err := db.QueryRow(`SELECT status,payment_status FROM sales WHERE id=?`, saleID).Scan(&status, &paymentStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT status FROM inventory_items WHERE id=?`, itemID).Scan(&itemStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT quantity FROM inventory`).Scan(&stock); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT current_balance FROM customers WHERE id=?`, customerID).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT amount FROM customer_ledger WHERE type='credit' AND reference_id=?`, saleID).Scan(&reversalAmount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT status,remaining_amount FROM debts WHERE sale_id=?`, saleID).Scan(&debtStatus, &debtRemaining); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM customer_debts WHERE reference_id=?`, saleID).Scan(&customerDebtCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM inventory_movements WHERE reference_id=? AND movement_type='SALE_CANCELLATION'`, saleID).Scan(&cancellationMovementCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE entity_id=? AND action='CANCEL'`, saleID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM item_history WHERE inventory_item_id=? AND event_type='sale_cancelled'`, itemID).Scan(&itemHistoryCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT sales_total,sale_count FROM pos_shifts`).Scan(&shiftTotal, &shiftSalesCount); err != nil {
		t.Fatal(err)
	}
	if status != "cancelled" || paymentStatus != "cancelled" || itemStatus != "AVAILABLE" || stock != 2 || balance != 0 || reversalAmount != 120 || debtStatus != "cancelled" || debtRemaining != 0 || customerDebtCount != 0 || cancellationMovementCount != 1 || auditCount != 1 || shiftTotal != 0 || shiftSalesCount != 0 || itemHistoryCount != 1 {
		t.Fatalf("incomplete cancellation: status=%s payment=%s item=%s stock=%d balance=%v reversal=%v debt=%s/%v customerDebts=%d movements=%d audit=%d shift=%v/%d history=%d", status, paymentStatus, itemStatus, stock, balance, reversalAmount, debtStatus, debtRemaining, customerDebtCount, cancellationMovementCount, auditCount, shiftTotal, shiftSalesCount, itemHistoryCount)
	}
}

func TestCancelSaleWithCollectionOrReturnIsBlockedWithoutChanges(t *testing.T) {
	for _, scenario := range []string{"payment", "return"} {
		t.Run(scenario, func(t *testing.T) {
			service, saleID, itemID, _ := newSaleCancellationTestService(t)
			db := service.db.DB
			switch scenario {
			case "payment":
				if _, err := db.Exec(`UPDATE sales SET paid_amount=20 WHERE id=?`, saleID); err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`INSERT INTO payments (id,sale_id) VALUES (?,?)`, uuid.NewString(), saleID); err != nil {
					t.Fatal(err)
				}
			case "return":
				if _, err := db.Exec(`INSERT INTO returns (id,sale_id) VALUES (?,?)`, uuid.NewString(), saleID); err != nil {
					t.Fatal(err)
				}
			}
			if err := service.CancelSale(context.Background(), uuid.New(), uuid.MustParse(saleID)); err != ErrInvalidSaleStatus {
				t.Fatalf("CancelSale() error = %v, want ErrInvalidSaleStatus", err)
			}
			var status, itemStatus string
			var stock int
			if err := db.QueryRow(`SELECT status FROM sales WHERE id=?`, saleID).Scan(&status); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(`SELECT status FROM inventory_items WHERE id=?`, itemID).Scan(&itemStatus); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRow(`SELECT quantity FROM inventory`).Scan(&stock); err != nil {
				t.Fatal(err)
			}
			if status != "completed" || itemStatus != "SOLD" || stock != 1 {
				t.Fatalf("blocked cancellation changed sale or stock: status=%s item=%s stock=%d", status, itemStatus, stock)
			}
		})
	}
}

func TestUpdateSalePaymentRejectsCancelledSale(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`CREATE TABLE sales (
		id TEXT PRIMARY KEY, sale_date TEXT, customer_id TEXT, invoice_number TEXT,
		subtotal REAL, tax_amount REAL, discount_amount REAL, total_amount REAL,
		cost_amount REAL, gross_profit REAL, net_profit REAL, paid_amount REAL,
		payment_method TEXT, payment_status TEXT, status TEXT, notes TEXT,
		created_at TEXT, updated_at TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}
	saleID := uuid.New()
	stamp := "2026-09-25T10:00:00Z"
	_, err = db.Exec(`INSERT INTO sales (id,sale_date,invoice_number,subtotal,tax_amount,discount_amount,total_amount,cost_amount,gross_profit,net_profit,paid_amount,payment_status,status,created_at,updated_at) VALUES (?,?,?,120,0,0,120,0,120,120,0,'cancelled','cancelled',?,?)`, saleID.String(), stamp, "INV-CANCELLED", stamp, stamp)
	if err != nil {
		t.Fatal(err)
	}

	service := &Service{db: db}
	if err := service.UpdateSalePayment(context.Background(), uuid.New(), saleID, 10, "cash"); err != ErrInvalidSaleStatus {
		t.Fatalf("UpdateSalePayment() error = %v, want ErrInvalidSaleStatus", err)
	}
	var paidAmount float64
	if err := db.Get(&paidAmount, `SELECT paid_amount FROM sales WHERE id=?`, saleID.String()); err != nil {
		t.Fatal(err)
	}
	if paidAmount != 0 {
		t.Fatalf("cancelled sale paid_amount = %v, want 0", paidAmount)
	}
}
