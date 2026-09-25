package debts

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
		`CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, amount REAL NOT NULL, paid_amount REAL NOT NULL DEFAULT 0, remaining_amount REAL NOT NULL, status TEXT NOT NULL, updated_at TEXT)`,
		`CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT NOT NULL, customer_id TEXT, amount REAL NOT NULL, payment_method TEXT, reference TEXT, notes TEXT, created_at TEXT NOT NULL)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	debtID := uuid.New()
	if _, err := db.Exec(`INSERT INTO debts (id, customer_id, amount, paid_amount, remaining_amount, status) VALUES (?, ?, 100, 0, 100, 'pending')`, debtID, uuid.New()); err != nil {
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
