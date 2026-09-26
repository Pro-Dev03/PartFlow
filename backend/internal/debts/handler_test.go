package debts

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/localdb"
	_ "modernc.org/sqlite"
)

func TestDebtSummaryTracksPartialAndClosedDebt(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE debts (
			id TEXT PRIMARY KEY,
			customer_id TEXT NOT NULL,
			amount REAL NOT NULL,
			remaining_amount REAL NOT NULL,
			status TEXT NOT NULL
		);
		INSERT INTO debts (id, customer_id, amount, remaining_amount, status)
		VALUES ('debt-1', 'customer-1', 250, 250, 'pending');
	`)
	if err != nil {
		t.Fatal(err)
	}

	handler := NewHandler(sqlx.NewDb(db, "sqlite"))
	summary, err := handler.loadDebtSummary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalDebt != 250 || summary.PaidAmount != 0 || summary.RemainingAmount != 250 || summary.CustomerCount != 1 {
		t.Fatalf("initial summary = %+v, want total 250, paid 0, remaining 250, customers 1", summary)
	}

	_, err = db.Exec(`UPDATE debts SET remaining_amount = 150, status = 'partial' WHERE id = 'debt-1'`)
	if err != nil {
		t.Fatal(err)
	}
	summary, err = handler.loadDebtSummary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalDebt != 250 || summary.PaidAmount != 100 || summary.RemainingAmount != 150 || summary.CustomerCount != 1 {
		t.Fatalf("partial summary = %+v, want total 250, paid 100, remaining 150, customers 1", summary)
	}

	_, err = db.Exec(`UPDATE debts SET remaining_amount = 0, status = 'paid' WHERE id = 'debt-1'`)
	if err != nil {
		t.Fatal(err)
	}
	summary, err = handler.loadDebtSummary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalDebt != 250 || summary.PaidAmount != 250 || summary.RemainingAmount != 0 || summary.CustomerCount != 0 {
		t.Fatalf("closed summary = %+v, want total 250, paid 250, remaining 0, customers 0", summary)
	}
}

func TestGetCustomerDebtsReturnsInvoiceAndProductHistorySQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	for _, statement := range []string{
		`CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, sale_id TEXT, amount REAL NOT NULL, paid_amount REAL NOT NULL DEFAULT 0, remaining_amount REAL NOT NULL, due_date TEXT, status TEXT, notes TEXT, created_at TEXT)`,
		`CREATE TABLE sales (id TEXT PRIMARY KEY, invoice_number TEXT)`,
		`CREATE TABLE products (id TEXT PRIMARY KEY, name TEXT)`,
		`CREATE TABLE sale_items (id TEXT PRIMARY KEY, sale_id TEXT, product_id TEXT, quantity INTEGER, unit_price REAL, total_amount REAL, created_at TEXT)`,
		`CREATE TABLE customer_debts (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, reference_id TEXT, reference_type TEXT, due_date TEXT, is_paid INTEGER, paid_amount REAL, created_at TEXT)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}

	customerID, saleID, debtID, manualDebtID, productID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO sales (id, invoice_number) VALUES (?, ?)`, saleID, "INV-2026-001"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO products (id, name) VALUES (?, ?)`, productID, "بطارية اختبار"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_price, total_amount, created_at) VALUES (?, ?, ?, 2, 35, 70, '2026-09-25')`, uuid.New(), saleID, productID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO debts (id, customer_id, sale_id, amount, paid_amount, remaining_amount, due_date, status, notes, created_at) VALUES (?, ?, ?, 70, 20, 50, '2026-10-25', 'partial', 'Sale: INV-2026-001', '2026-09-25')`, debtID, customerID, saleID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO debts (id, customer_id, amount, paid_amount, remaining_amount, due_date, status, notes, created_at) VALUES (?, ?, 15, 0, 15, '2026-10-25', 'pending', 'رصيد يدوي', '2026-09-24')`, manualDebtID, customerID); err != nil {
		t.Fatal(err)
	}
	openingDebtID, duplicateSaleDebtID := uuid.New(), uuid.New()
	if _, err := db.Exec(`INSERT INTO customer_debts (id, customer_id, amount, reference_type, due_date, is_paid, paid_amount, created_at) VALUES (?, ?, 40, 'opening_debt', '2026-10-25', 0, 0, '2026-09-23')`, openingDebtID, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_debts (id, customer_id, amount, reference_id, reference_type, due_date, is_paid, paid_amount, created_at) VALUES (?, ?, 70, ?, 'sale', '2026-10-25', 0, 0, '2026-09-25')`, duplicateSaleDebtID, customerID, saleID); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/debts/customer/:customer_id", NewHandler(sqlx.NewDb(db, "sqlite")).GetCustomerDebts)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/debts/customer/"+customerID.String(), nil)
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("history status = %d, body = %s", response.Code, response.Body.String())
	}

	var payload struct {
		Data []struct {
			ID              string  `json:"id"`
			SaleID          string  `json:"sale_id"`
			ReferenceType   string  `json:"reference_type"`
			InvoiceNumber   string  `json:"invoice_number"`
			Amount          float64 `json:"amount"`
			PaidAmount      float64 `json:"paid_amount"`
			RemainingAmount float64 `json:"remaining_amount"`
			Items           []struct {
				ProductName string  `json:"product_name"`
				Quantity    int     `json:"quantity"`
				UnitPrice   float64 `json:"unit_price"`
				TotalAmount float64 `json:"total_amount"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Data) != 3 {
		t.Fatalf("history entries = %d, want 3", len(payload.Data))
	}
	var linked, manual, opening bool
	for _, entry := range payload.Data {
		if entry.ID == debtID.String() {
			linked = true
			if entry.SaleID != saleID.String() || entry.InvoiceNumber != "INV-2026-001" || entry.Amount != 70 || entry.PaidAmount != 20 || entry.RemainingAmount != 50 {
				t.Fatalf("linked debt summary = %+v", entry)
			}
			if len(entry.Items) != 1 || entry.Items[0].ProductName != "بطارية اختبار" || entry.Items[0].Quantity != 2 || entry.Items[0].UnitPrice != 35 || entry.Items[0].TotalAmount != 70 {
				t.Fatalf("linked debt items = %+v", entry.Items)
			}
		}
		if entry.ID == manualDebtID.String() {
			manual = true
			if len(entry.Items) != 0 {
				t.Fatalf("manual debt unexpectedly has product items: %+v", entry.Items)
			}
		}
		if entry.ID == openingDebtID.String() {
			opening = entry.ReferenceType == "opening_debt" && entry.SaleID == "" && len(entry.Items) == 0
		}
	}
	if !linked || !manual || !opening {
		t.Fatalf("linked debt found=%v, manual debt found=%v, opening debt found=%v", linked, manual, opening)
	}
}

func TestAddDebtPaymentSynchronizesBalanceAndRejectsOverpaymentSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	xdb := sqlx.NewDb(db, "sqlite")
	for _, statement := range []string{
		`CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL NOT NULL DEFAULT 0, updated_at TEXT)`,
		`CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, amount REAL NOT NULL, paid_amount REAL NOT NULL DEFAULT 0, remaining_amount REAL NOT NULL, status TEXT NOT NULL, updated_at TEXT)`,
		`CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT NOT NULL, customer_id TEXT, amount REAL NOT NULL, payment_method TEXT, reference TEXT, notes TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE payment_allocation_batches (payment_id TEXT PRIMARY KEY, owner_type TEXT NOT NULL, owner_id TEXT NOT NULL, sale_id TEXT, tracked_at TEXT NOT NULL)`,
		`CREATE TABLE payment_debt_allocations (id TEXT PRIMARY KEY, payment_id TEXT NOT NULL, debt_id TEXT NOT NULL, amount REAL NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, debt_id TEXT, type TEXT, transaction_type TEXT, amount REAL NOT NULL, balance REAL NOT NULL, description TEXT, reference_id TEXT, created_at TEXT NOT NULL)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	debtID := uuid.New()
	customerID := uuid.New()
	if _, err := db.Exec(`INSERT INTO customers (id,current_balance) VALUES (?,100)`, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO debts (id, customer_id, amount, paid_amount, remaining_amount, status) VALUES (?, ?, 100, 0, 100, 'pending')`, debtID, customerID); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.POST("/debts/:id/payments", NewHandler(xdb).AddPayment)
	request := func(body string) int {
		t.Helper()
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/debts/"+debtID.String()+"/payments", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, req)
		return response.Code
	}
	if status := request(`{"amount":40}`); status != http.StatusOK {
		t.Fatalf("partial payment status = %d, want 200", status)
	}
	var paid, remaining float64
	if err := db.QueryRow(`SELECT paid_amount, remaining_amount FROM debts WHERE id = ?`, debtID.String()).Scan(&paid, &remaining); err != nil {
		t.Fatal(err)
	}
	if paid != 40 || remaining != 60 {
		t.Fatalf("partial debt balance = paid %.2f remaining %.2f, want 40/60", paid, remaining)
	}
	if status := request(`{"amount":60.01}`); status != http.StatusBadRequest {
		t.Fatalf("overpayment status = %d, want 400", status)
	}
	var paymentCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payments`).Scan(&paymentCount); err != nil {
		t.Fatal(err)
	}
	if paymentCount != 1 {
		t.Fatalf("payment rows after rejected overpayment = %d, want 1", paymentCount)
	}
	if status := request(`{"amount":60}`); status != http.StatusOK {
		t.Fatalf("final payment status = %d, want 200", status)
	}
	var finalPaid, finalRemaining float64
	var finalStatus string
	if err := db.QueryRow(`SELECT paid_amount, remaining_amount, status FROM debts WHERE id = ?`, debtID.String()).Scan(&finalPaid, &finalRemaining, &finalStatus); err != nil {
		t.Fatal(err)
	}
	if finalPaid != 100 || finalRemaining != 0 || finalStatus != "paid" {
		t.Fatalf("final debt state = paid %.2f remaining %.2f status %q, want 100/0/paid", finalPaid, finalRemaining, finalStatus)
	}
}

func TestUpdateDebtOnlyChangesMetadataAndRejectsFinancialFieldsSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE debts (
		id TEXT PRIMARY KEY,
		customer_id TEXT NOT NULL,
		amount REAL NOT NULL,
		paid_amount REAL NOT NULL DEFAULT 0,
		remaining_amount REAL NOT NULL,
		due_date TEXT NOT NULL,
		status TEXT NOT NULL,
		notes TEXT,
		updated_at TEXT
	)`); err != nil {
		t.Fatal(err)
	}
	debtID := uuid.New()
	if _, err := db.Exec(`INSERT INTO debts (id, customer_id, amount, paid_amount, remaining_amount, due_date, status, notes)
		VALUES (?, ?, 100, 25, 75, '2026-09-25', 'partial', 'original')`, debtID, uuid.New()); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.PUT("/debts/:id", NewHandler(sqlx.NewDb(db, "sqlite")).UpdateDebt)
	request := func(body string) int {
		t.Helper()
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/debts/"+debtID.String(), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, req)
		return response.Code
	}
	if status := request(`{"due_date":"2026-10-15","notes":"corrected note"}`); status != http.StatusOK {
		t.Fatalf("metadata update status = %d, want 200", status)
	}
	var amount, paid, remaining float64
	var dueDate, debtStatus, notes string
	if err := db.QueryRow(`SELECT amount, paid_amount, remaining_amount, due_date, status, notes FROM debts WHERE id = ?`, debtID.String()).Scan(&amount, &paid, &remaining, &dueDate, &debtStatus, &notes); err != nil {
		t.Fatal(err)
	}
	if amount != 100 || paid != 25 || remaining != 75 || dueDate != "2026-10-15" || debtStatus != "partial" || notes != "corrected note" {
		t.Fatalf("debt after metadata edit = amount %.2f paid %.2f remaining %.2f due %q status %q notes %q", amount, paid, remaining, dueDate, debtStatus, notes)
	}
	for _, body := range []string{
		`{"amount":1}`,
		`{"remaining_amount":0,"status":"paid"}`,
		`{"due_date":"not-a-date"}`,
		`{}`,
	} {
		if status := request(body); status != http.StatusBadRequest {
			t.Errorf("invalid update %s status = %d, want 400", body, status)
		}
	}
	if status := func() int {
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/debts/"+uuid.NewString(), strings.NewReader(`{"notes":"missing"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, req)
		return response.Code
	}(); status != http.StatusNotFound {
		t.Fatalf("missing debt update status = %d, want 404", status)
	}
}

func TestManualDebtCreateAndDeleteReconcileCustomerLedgerSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	xdb := sqlx.NewDb(db, "sqlite")
	for _, statement := range []string{
		`CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL NOT NULL DEFAULT 0, updated_at TEXT)`,
		`CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, sale_id TEXT, amount REAL NOT NULL, paid_amount REAL NOT NULL DEFAULT 0, remaining_amount REAL NOT NULL, due_date TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE customer_debts (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, reference_id TEXT, reference_type TEXT, due_date TEXT, is_paid INTEGER, paid_amount REAL, created_at TEXT)`,
		`CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, debt_id TEXT, type TEXT, amount REAL, balance REAL, description TEXT, reference_id TEXT, created_at TEXT)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	customerID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO customers (id, current_balance) VALUES (?, 0)`, customerID); err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	handler := NewHandler(xdb)
	router.POST("/debts", handler.CreateDebt)
	router.DELETE("/debts/:id", handler.DeleteDebt)
	created := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/debts", strings.NewReader(`{"customer_id":"`+customerID+`","amount":125.5,"due_date":"2026-10-26","notes":"manual test debt"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(created, request)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", created.Code, created.Body.String())
	}
	var createPayload struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &createPayload); err != nil {
		t.Fatal(err)
	}
	debtID := createPayload.Data.ID
	if debtID == "" {
		t.Fatalf("create response omitted debt id: %s", created.Body.String())
	}
	var balance, ledgerBalance float64
	if err := db.QueryRow(`SELECT current_balance FROM customers WHERE id = ?`, customerID).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT balance FROM customer_ledger WHERE reference_id = ?`, debtID).Scan(&ledgerBalance); err != nil {
		t.Fatal(err)
	}
	if balance != 125.5 || ledgerBalance != 125.5 {
		t.Fatalf("created customer balance %.2f, ledger balance %.2f; want 125.50 each", balance, ledgerBalance)
	}

	deleted := httptest.NewRecorder()
	router.ServeHTTP(deleted, httptest.NewRequest(http.MethodDelete, "/debts/"+debtID, nil))
	if deleted.Code != http.StatusOK {
		t.Fatalf("delete status = %d, body = %s", deleted.Code, deleted.Body.String())
	}
	var debtCount, historyCount, ledgerCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM debts WHERE id = ?`, debtID).Scan(&debtCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM customer_debts WHERE id = ?`, debtID).Scan(&historyCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM customer_ledger WHERE reference_id = ?`, debtID).Scan(&ledgerCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT current_balance FROM customers WHERE id = ?`, customerID).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if debtCount != 0 || historyCount != 0 || ledgerCount != 0 || balance != 0 {
		t.Fatalf("after delete debt/history/ledger=%d/%d/%d balance=%.2f; want 0/0/0/0", debtCount, historyCount, ledgerCount, balance)
	}
}

func TestDeleteInvoiceDebtReversesAndHardDeletesSaleSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "invoice-debt-delete.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	customerID, saleID, debtID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO customers (id,code,name,current_balance,created_at,updated_at) VALUES (?,?,?,50,?,?)`, []any{customerID, "DEBT-CASCADE-CUSTOMER", "Debt cascade customer", now, now}},
		{`INSERT INTO sales (id,sale_number,customer_id,total_amount,paid_amount,remaining_amount,payment_method,status,created_at,updated_at) VALUES (?,?,?,?,0,50,'credit','completed',?,?)`, []any{saleID, "DEBT-CASCADE-SALE", customerID, 50, now, now}},
		{`INSERT INTO debts (id,customer_id,sale_id,amount,paid_amount,remaining_amount,due_date,status,notes,created_at,updated_at) VALUES (?,?,?,50,0,50,?,'pending','invoice debt',?,?)`, []any{debtID, customerID, saleID, now, now, now}},
		{`INSERT INTO customer_ledger (id,customer_id,debt_id,type,amount,balance,description,reference_id,created_at) VALUES (?,?,?,'debit',50,50,'sale debt',?,?)`, []any{uuid.New(), customerID, debtID, saleID, now}},
	} {
		if _, err := db.ExecContext(context.Background(), statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	router := gin.New()
	router.DELETE("/debts/:id", NewHandler(db).DeleteDebt)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/debts/"+debtID.String(), nil))
	if response.Code != http.StatusOK {
		t.Fatalf("delete invoice debt status = %d, body=%s", response.Code, response.Body.String())
	}
	var saleCount, debtCount, ledgerCount int
	if err := db.Get(&saleCount, `SELECT COUNT(*) FROM sales WHERE id=?`, saleID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&debtCount, `SELECT COUNT(*) FROM debts WHERE id=?`, debtID); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&ledgerCount, `SELECT COUNT(*) FROM customer_ledger WHERE customer_id=?`, customerID); err != nil {
		t.Fatal(err)
	}
	var balance float64
	if err := db.Get(&balance, `SELECT current_balance FROM customers WHERE id=?`, customerID); err != nil {
		t.Fatal(err)
	}
	if saleCount != 0 || debtCount != 0 || ledgerCount != 0 || balance != 0 {
		t.Fatalf("invoice debt cascade left sale/debt/ledger/balance=%d/%d/%d/%.2f; want all zero", saleCount, debtCount, ledgerCount, balance)
	}
}

func TestDeletePartiallyPaidManualDebtReversesPaymentLedgerAndBalanceSQLite(t *testing.T) {
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "paid-manual-debt-delete.sqlite"))
	local, err := localdb.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer local.DB.Close()
	db := sqlx.NewDb(local.DB, "sqlite")
	customerID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.Exec(`INSERT INTO customers (id,code,name,current_balance,created_at,updated_at) VALUES (?,?,?,0,?,?)`, customerID, "DEBT-PAY-CUSTOMER", "Debt payment customer", now, now); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	handler := NewHandler(db)
	router.POST("/debts", handler.CreateDebt)
	router.POST("/debts/:id/payment", handler.AddPayment)
	router.DELETE("/debts/:id", handler.DeleteDebt)
	created := httptest.NewRecorder()
	router.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/debts", strings.NewReader(`{"customer_id":"`+customerID.String()+`","amount":100,"due_date":"2026-10-26","notes":"manual debt"}`)))
	if created.Code != http.StatusCreated {
		t.Fatalf("create debt status=%d body=%s", created.Code, created.Body.String())
	}
	var payload struct {
		Data struct {
			ID uuid.UUID `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.ID == uuid.Nil {
		t.Fatalf("create debt returned no id: %s", created.Body.String())
	}
	paid := httptest.NewRecorder()
	router.ServeHTTP(paid, httptest.NewRequest(http.MethodPost, "/debts/"+payload.Data.ID.String()+"/payment", strings.NewReader(`{"amount":25,"notes":"partial payment"}`)))
	if paid.Code != http.StatusOK {
		t.Fatalf("add partial payment status=%d body=%s", paid.Code, paid.Body.String())
	}
	deleted := httptest.NewRecorder()
	router.ServeHTTP(deleted, httptest.NewRequest(http.MethodDelete, "/debts/"+payload.Data.ID.String(), nil))
	if deleted.Code != http.StatusOK {
		t.Fatalf("delete paid debt status=%d body=%s", deleted.Code, deleted.Body.String())
	}
	var debtCount, paymentCount, allocationCount, ledgerCount int
	for _, check := range []struct {
		query string
		args  []any
		into  *int
	}{
		{`SELECT COUNT(*) FROM debts WHERE id=?`, []any{payload.Data.ID}, &debtCount},
		{`SELECT COUNT(*) FROM payments WHERE customer_id=?`, []any{customerID}, &paymentCount},
		{`SELECT COUNT(*) FROM payment_debt_allocations WHERE debt_id=?`, []any{payload.Data.ID}, &allocationCount},
		{`SELECT COUNT(*) FROM customer_ledger WHERE customer_id=?`, []any{customerID}, &ledgerCount},
	} {
		if err := db.Get(check.into, check.query, check.args...); err != nil {
			t.Fatal(err)
		}
	}
	var balance float64
	if err := db.Get(&balance, `SELECT current_balance FROM customers WHERE id=?`, customerID); err != nil {
		t.Fatal(err)
	}
	if debtCount != 0 || paymentCount != 0 || allocationCount != 0 || ledgerCount != 0 || balance != 0 {
		t.Fatalf("after cascade debt/payment/allocation/ledger=%d/%d/%d/%d balance=%.2f; want all zero", debtCount, paymentCount, allocationCount, ledgerCount, balance)
	}
}

func TestDeleteDebtUsesDebtLinkAndKeepsProductReferenceSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	for _, statement := range []string{
		`CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL NOT NULL DEFAULT 0, updated_at TEXT)`,
		`CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, sale_id TEXT, amount REAL NOT NULL, paid_amount REAL NOT NULL DEFAULT 0, remaining_amount REAL NOT NULL, due_date TEXT, status TEXT, notes TEXT, created_at TEXT, updated_at TEXT)`,
		`CREATE TABLE customer_debts (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, reference_id TEXT, reference_type TEXT, due_date TEXT, is_paid INTEGER, paid_amount REAL, created_at TEXT)`,
		`CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, debt_id TEXT, type TEXT, amount REAL, balance REAL, description TEXT, reference_id TEXT, created_at TEXT)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	customerID, debtID, productID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	if _, err := db.Exec(`INSERT INTO customers (id, current_balance) VALUES (?, 42)`, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO debts (id, customer_id, amount, paid_amount, remaining_amount, status, notes) VALUES (?, ?, 42, 0, 42, 'pending', 'MANUAL_ADJUSTMENT')`, debtID, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_debts (id, customer_id, amount, reference_type, is_paid, paid_amount) VALUES (?, ?, 42, 'MANUAL_ADJUSTMENT', 0, 0)`, debtID, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_ledger (id, customer_id, debt_id, type, amount, balance, description, reference_id, created_at) VALUES (?, ?, ?, 'debit', 42, 42, 'manual debt with product', ?, CURRENT_TIMESTAMP)`, uuid.NewString(), customerID, debtID, productID); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.DELETE("/debts/:id", NewHandler(sqlx.NewDb(db, "sqlite")).DeleteDebt)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/debts/"+debtID, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("delete linked debt status = %d, body=%s", response.Code, response.Body.String())
	}
	var customerBalance float64
	var ledgerCount, debtCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM customer_ledger WHERE reference_id = ? OR debt_id = ?`, productID, debtID).Scan(&ledgerCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT current_balance FROM customers WHERE id = ?`, customerID).Scan(&customerBalance); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM debts WHERE id = ?`, debtID).Scan(&debtCount); err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 0 || debtCount != 0 || customerBalance != 0 {
		t.Fatalf("debt/ledger rows=%d/%d balance=%.2f; want 0/0/0", debtCount, ledgerCount, customerBalance)
	}
}
