package settings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/expenses"
	_ "modernc.org/sqlite"
)

func TestCleanupPreviewAndRunSQLiteOnlyDeleteExpiredTechnicalRows(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE idempotency_keys (id TEXT PRIMARY KEY, idempotency_key TEXT, request_hash TEXT, response_body TEXT, expires_at TEXT);
		CREATE TABLE refresh_tokens (id TEXT PRIMARY KEY, token TEXT, expires_at TEXT);
		CREATE TABLE password_reset_tokens (id TEXT PRIMARY KEY, token TEXT, expires_at TEXT, used INTEGER);
		CREATE TABLE notifications (id TEXT PRIMARY KEY, title TEXT, message TEXT, data TEXT, expires_at TEXT, status TEXT, created_at TEXT);
		CREATE TABLE reservations (id TEXT PRIMARY KEY, notes TEXT, status TEXT, updated_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL);
	`)
	if err != nil {
		t.Fatalf("create cleanup test schema: %v", err)
	}
	old := time.Now().UTC().AddDate(0, 0, -120).Format("2006-01-02 15:04:05")
	future := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")
	if _, err := db.Exec(`INSERT INTO idempotency_keys VALUES ('idem-old','old','hash','{}',?),('idem-new','new','hash','{}',?)`, old, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO refresh_tokens VALUES ('refresh-old','old',?),('refresh-new','new',?)`, old, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO password_reset_tokens VALUES ('reset-expired','expired',?,0),('reset-used','used',?,1),('reset-valid','valid',?,0)`, old, future, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO notifications VALUES ('notice-old','old','message','{}',NULL,'read',?),('notice-live','live','message','{}',NULL,'unread',?)`, old, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO reservations VALUES ('reservation-old','old','cancelled',?),('reservation-live','live','active',?)`, old, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sales VALUES ('sale-keep',100)`); err != nil {
		t.Fatal(err)
	}

	preview, err := buildCleanupPreview(context.Background(), db, true)
	if err != nil {
		t.Fatalf("preview cleanup: %v", err)
	}
	if preview.TotalCount != 6 {
		t.Fatalf("preview candidate count = %d, want 6", preview.TotalCount)
	}
	if preview.EstimatedBytes <= 0 || preview.Storage.DatabaseBytes <= 0 {
		t.Fatalf("preview should contain storage estimates: %+v", preview)
	}

	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/cleanup", nil)
	NewDatabaseHandler(db).RunCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("cleanup status = %d, body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data CleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode cleanup response: %v", err)
	}
	if envelope.Data.DeletedCount != 6 || envelope.Data.Remaining != 0 {
		t.Fatalf("cleanup result = %+v, want 6 deleted and none remaining", envelope.Data)
	}

	for _, check := range []struct {
		table string
		want  int
	}{
		{"idempotency_keys", 1},
		{"refresh_tokens", 1},
		{"password_reset_tokens", 1},
		{"notifications", 1},
		{"reservations", 1},
		{"sales", 1},
	} {
		var count int
		if err := db.Get(&count, `SELECT COUNT(*) FROM `+check.table); err != nil {
			t.Fatal(err)
		}
		if count != check.want {
			t.Errorf("%s rows remaining = %d, want %d", check.table, count, check.want)
		}
	}
}

func TestHistoricalCleanupPreviewUsesStoreCalendarDatesSQLite(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE sales (id TEXT PRIMARY KEY, created_at TEXT); INSERT INTO sales VALUES ('before-midnight','2026-05-01T02:30:00Z'),('after-midnight','2026-05-01T04:30:00Z')`); err != nil {
		t.Fatal(err)
	}
	preview, err := loadHistoricalCleanupPreview(context.Background(), db, true, "sales", "2026-04-30", "2026-04-30")
	if err != nil {
		t.Fatalf("preview historical cleanup: %v", err)
	}
	if preview.CandidateCount != 1 || preview.Timezone != "America/New_York" {
		t.Fatalf("preview = %+v; want only the timestamp before store midnight", preview)
	}
	if _, _, err := historicalCleanupRange("2026-04-31", "2026-05-01"); err == nil {
		t.Fatal("invalid calendar date should be rejected")
	}
}

func TestRunHistoricalExpenseCleanupDeletesOnlySelectedStoreDateSQLite(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TABLE expenses (
		id TEXT PRIMARY KEY, category_id TEXT, title TEXT NOT NULL, description TEXT, amount REAL NOT NULL,
		currency TEXT, expense_date TEXT NOT NULL, payment_method TEXT, reference TEXT, receipt_url TEXT,
		is_recurring INTEGER, recurring_period TEXT, approved_by TEXT, status TEXT, created_by TEXT,
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{uuid.NewString(), uuid.NewString()}
	if _, err := db.Exec(`INSERT INTO expenses (id,title,amount,currency,expense_date,status,is_recurring,created_at,updated_at) VALUES (?, 'in range', 12, 'ILS', '2026-05-01T02:30:00Z', 'pending', 0, '2026-05-01T02:30:00Z', '2026-05-01T02:30:00Z'), (?, 'out of range', 20, 'ILS', '2026-05-01T04:30:00Z', 'pending', 0, '2026-05-01T04:30:00Z', '2026-05-01T04:30:00Z')`, ids[0], ids[1]); err != nil {
		t.Fatal(err)
	}
	if _, err := expenses.NewRepository(db).GetExpenseByID(context.Background(), uuid.MustParse(ids[0])); err != nil {
		t.Fatalf("load expense through cleanup service: %v", err)
	}
	newInRangeID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO expenses (id,title,amount,currency,expense_date,status,is_recurring,created_at,updated_at) VALUES (?, 'added after preview', 7, 'ILS', '2026-05-01T03:00:00Z', 'pending', 0, '2026-05-01T03:00:00Z', '2026-05-01T03:00:00Z')`, newInRangeID); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(`{"type":"expenses","start_date":"2026-04-30","end_date":"2026-04-30","candidate_ids":["`+ids[0]+`"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("historical cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Total != 1 || envelope.Data.Deleted != 1 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 {
		t.Fatalf("cleanup result = %+v", envelope.Data)
	}
	var remaining int
	if err := db.Get(&remaining, `SELECT COUNT(*) FROM expenses`); err != nil {
		t.Fatal(err)
	}
	if remaining != 2 {
		t.Fatalf("remaining expenses=%d; want the out-of-range and post-preview records preserved", remaining)
	}
}

func TestRunHistoricalApprovedExpenseCleanupHardDeletesAndUpdatesFinancialTotalSQLite(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("UTC"); err != nil {
		t.Fatal(err)
	}
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE expenses (
		id TEXT PRIMARY KEY, title TEXT NOT NULL, category_id TEXT, amount REAL NOT NULL, currency TEXT,
		description TEXT, expense_date TEXT NOT NULL, payment_method TEXT, reference TEXT, receipt_url TEXT,
		is_recurring INTEGER, recurring_period TEXT, approved_by TEXT, status TEXT, created_by TEXT,
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	)`); err != nil {
		t.Fatal(err)
	}
	expenseID := uuid.New()
	if _, err := db.Exec(`INSERT INTO expenses (id,title,amount,currency,description,expense_date,payment_method,reference,receipt_url,is_recurring,status,created_at,updated_at) VALUES (?, 'Office supplies',25,'ILS','private description','2026-04-30','cash','REF-123','receipt://old',1,'approved','2026-04-30','2026-04-30')`, expenseID.String()); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	before, err := accounting.TotalExpensesForPeriod(context.Background(), db, start, end)
	if err != nil || before != 25 {
		t.Fatalf("SQLite expenses before cleanup=%v err=%v; want 25", before, err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/settings/history-cleanup", strings.NewReader(`{"type":"expenses","start_date":"2026-04-30","end_date":"2026-04-30","candidate_ids":["`+expenseID.String()+`"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != http.StatusOK {
		t.Fatalf("approved expense cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM expenses WHERE id=?`, expenseID.String()).Scan(&count); err != nil {
		t.Fatal(err)
	}
	after, err := accounting.TotalExpensesForPeriod(context.Background(), db, start, end)
	if err != nil || envelope.Data.Deleted != 1 || envelope.Data.Failed != 0 || count != 0 || after != 0 {
		t.Fatalf("approved expense cleanup: result=%+v count=%d total_before=%v total_after=%v err=%v; want deleted=1, absent row, zero total", envelope.Data, count, before, after, err)
	}
}
func TestRunHistoricalDebtCleanupDeletesManualDebtAndReversesLedgerSQLite(t *testing.T) {
	previousTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(previousTimezone) })
	if err := accounting.ConfigureStoreTimezone("America/New_York"); err != nil {
		t.Fatal(err)
	}
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
		CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT NOT NULL, sale_id TEXT, amount REAL NOT NULL, paid_amount REAL, remaining_amount REAL, created_at TEXT);
		CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL NOT NULL DEFAULT 0, updated_at TEXT);
		CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, debt_id TEXT, type TEXT, amount REAL, balance REAL, reference_id TEXT, created_at TEXT);
		CREATE TABLE customer_debts (id TEXT PRIMARY KEY);
	`)
	if err != nil {
		t.Fatal(err)
	}
	manualDebtID, customerID := uuid.NewString(), uuid.NewString()
	if _, err := db.Exec(`INSERT INTO customers (id,current_balance,updated_at) VALUES (?,?,?)`, customerID, 50, "2026-05-01T02:30:00Z"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO debts (id,customer_id,amount,paid_amount,remaining_amount,created_at) VALUES (?,?,50,0,50,?)`, manualDebtID, customerID, "2026-05-01T02:30:00Z"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_debts (id) VALUES (?)`, manualDebtID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customer_ledger (id,customer_id,debt_id,type,amount,balance,reference_id,created_at) VALUES (?,?,?,'debit',50,50,?,?)`, uuid.NewString(), customerID, manualDebtID, manualDebtID, "2026-05-01T02:30:00Z"); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(`{"type":"debts","start_date":"2026-04-30","end_date":"2026-04-30","candidate_ids":["`+manualDebtID+`"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != http.StatusOK {
		t.Fatalf("historical debt cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Total != 1 || envelope.Data.Deleted != 1 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 {
		t.Fatalf("cleanup result = %+v; want selected manual debt reversed and deleted", envelope.Data)
	}
	var remainingDebtCount, remainingLedgerCount, remainingMirrorCount, clearedCustomerBalance int
	if err := db.Get(&remainingDebtCount, `SELECT COUNT(*) FROM debts`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&remainingLedgerCount, `SELECT COUNT(*) FROM customer_ledger`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&remainingMirrorCount, `SELECT COUNT(*) FROM customer_debts`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&clearedCustomerBalance, `SELECT COUNT(*) FROM customers WHERE id=? AND current_balance=0`, customerID); err != nil {
		t.Fatal(err)
	}
	if remainingDebtCount != 0 || remainingLedgerCount != 0 || remainingMirrorCount != 0 || clearedCustomerBalance != 1 {
		t.Fatalf("remaining debts=%d ledgers=%d mirrors=%d cleared_customer_balance=%d", remainingDebtCount, remainingLedgerCount, remainingMirrorCount, clearedCustomerBalance)
	}
}

func TestHistoricalIncomeCleanupDeletesOnlyStandaloneIncomeSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE financial_transactions (id TEXT PRIMARY KEY, sale_id TEXT, type TEXT NOT NULL, amount REAL NOT NULL, created_at TEXT NOT NULL);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, entity_type TEXT, entity_id TEXT, new_values TEXT, created_at TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	incomeID, linkedSaleTransactionID, expenseID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	for _, row := range []struct {
		id, saleID, kind string
	}{{incomeID, "", "income"}, {linkedSaleTransactionID, uuid.NewString(), "sale"}, {expenseID, "", "expense"}} {
		if _, err := db.Exec(`INSERT INTO financial_transactions (id,sale_id,type,amount,created_at) VALUES (?,?,?,?,?)`, row.id, nullableString(row.saleID), row.kind, 25, "2026-05-01T12:00:00Z"); err != nil {
			t.Fatal(err)
		}
	}
	preview, err := loadHistoricalCleanupPreview(context.Background(), db, true, "income", "2026-05-01", "2026-05-01")
	if err != nil || preview.CandidateCount != 1 || len(preview.CandidateIDs) != 1 || preview.CandidateIDs[0] != incomeID {
		t.Fatalf("income preview=%+v err=%v; want only standalone income", preview, err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(`{"type":"income","start_date":"2026-05-01","end_date":"2026-05-01","candidate_ids":["`+incomeID+`"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != http.StatusOK {
		t.Fatalf("income cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 1 || envelope.Data.Failed != 0 || envelope.Data.Blocked != 0 {
		t.Fatalf("income cleanup result=%+v; want one deleted income", envelope.Data)
	}
	var remaining int
	if err := db.Get(&remaining, `SELECT COUNT(*) FROM financial_transactions`); err != nil {
		t.Fatal(err)
	}
	if remaining != 2 {
		t.Fatalf("remaining financial transactions=%d; want linked sale and non-income transaction retained", remaining)
	}
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func TestRunHistoricalPaymentCleanupReversesCompletedCustomerPaymentSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE payment_allocation_batches (payment_id TEXT PRIMARY KEY, owner_type TEXT, owner_id TEXT, sale_id TEXT, tracked_at TEXT); CREATE TABLE payment_debt_allocations (id TEXT PRIMARY KEY, payment_id TEXT, debt_id TEXT, amount REAL, created_at TEXT); CREATE TABLE customer_ledger (id TEXT PRIMARY KEY, customer_id TEXT, reference_id TEXT, type TEXT, transaction_type TEXT, amount REAL, balance REAL, description TEXT, created_at TEXT); CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, reference_id TEXT, type TEXT, transaction_type TEXT); CREATE TABLE customers (id TEXT PRIMARY KEY, current_balance REAL, updated_at TEXT); CREATE TABLE debts (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, paid_amount REAL, remaining_amount REAL, due_date TEXT, status TEXT, sale_id TEXT, created_at TEXT, updated_at TEXT); CREATE TABLE customer_debts (id TEXT PRIMARY KEY, customer_id TEXT, amount REAL, paid_amount REAL, is_paid INTEGER); CREATE TABLE daily_debt_summary (date TEXT PRIMARY KEY, total_debt REAL, new_debt REAL, payments_received REAL, overdue_debt REAL, overdue_count INTEGER, paid_debt REAL, updated_at TEXT); CREATE TABLE monthly_debt_summary (year INTEGER, month INTEGER, total_debt REAL, new_debt REAL, payments_received REAL, overdue_debt REAL, overdue_count INTEGER, paid_debt REAL, updated_at TEXT, PRIMARY KEY(year,month));`); err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TABLE payments (
		id TEXT PRIMARY KEY, customer_id TEXT, supplier_id TEXT, sale_id TEXT, amount REAL, payment_date TEXT,
		payment_method TEXT, reference TEXT, notes TEXT, payment_status TEXT, created_by TEXT,
		created_at TEXT, updated_at TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}
	pendingID, completedID, changedID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	customerID := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO payments (id,customer_id,amount,payment_date,payment_method,payment_status,created_at,updated_at) VALUES (?, ?, 10, '2026-04-30', 'cash', 'pending', '2026-04-30T10:00:00Z', '2026-04-30T10:00:00Z'), (?, ?, 20, '2026-04-30', 'cash', 'completed', '2026-04-30T11:00:00Z', '2026-04-30T11:00:00Z'), (?, ?, 30, '2026-04-30', 'cash', 'pending', '2026-04-30T12:00:00Z', '2026-04-30T12:00:00Z')`, pendingID, customerID, completedID, customerID, changedID, customerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO customers VALUES (?, 30, '2026-04-30T12:00:00Z'); INSERT INTO customer_ledger VALUES ('legacy-payment-ledger', ?, ?, 'credit', 'PAYMENT', 30, -30, 'Payment', '2026-04-30T12:00:00Z')`, customerID, customerID, changedID); err != nil {
		t.Fatal(err)
	}
	preview, err := loadHistoricalCleanupPreview(context.Background(), db, true, "payments", "2026-04-30", "2026-04-30")
	if err != nil {
		t.Fatal(err)
	}
	if preview.CandidateCount != 3 || len(preview.CandidateIDs) != 3 {
		t.Fatalf("payment preview=%+v; want pending and completed payments", preview)
	}
	if _, err := db.Exec(`UPDATE payments SET payment_status='completed' WHERE id=?`, changedID); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(`{"type":"payments","start_date":"2026-04-30","end_date":"2026-04-30","candidate_ids":["`+pendingID+`","`+changedID+`"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != http.StatusOK {
		t.Fatalf("historical payment cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 2 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 {
		t.Fatalf("payment cleanup result=%+v", envelope.Data)
	}
	var remaining []string
	if err := db.Select(&remaining, `SELECT id FROM payments ORDER BY id`); err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 || remaining[0] != completedID {
		t.Fatalf("remaining payment IDs=%v; want only unselected completed payment", remaining)
	}
	var restoredBalance float64
	if err := db.Get(&restoredBalance, `SELECT current_balance FROM customers WHERE id = ?`, customerID); err != nil {
		t.Fatal(err)
	}
	if restoredBalance != 60 {
		t.Fatalf("customer balance after payment cleanup=%v; want original 60 after removing 30 payment from 30 balance", restoredBalance)
	}
}

func TestRunHistoricalProductCleanupReversesInventoryAndDeletesProductsSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE products (id TEXT PRIMARY KEY, created_at TEXT); CREATE TABLE inventory_items (id TEXT PRIMARY KEY, product_id TEXT);`); err != nil {
		t.Fatal(err)
	}
	unusedID, usedID := uuid.NewString(), uuid.NewString()
	if _, err := db.Exec(`INSERT INTO products (id,created_at) VALUES (?, '2020-01-01'), (?, '2020-01-02')`, unusedID, usedID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items (id,product_id) VALUES (?,?)`, uuid.NewString(), usedID); err != nil {
		t.Fatal(err)
	}
	preview, err := loadHistoricalCleanupPreview(context.Background(), db, true, "products", "2020-01-01", "2020-01-02")
	if err != nil {
		t.Fatal(err)
	}
	if preview.CandidateCount != 2 {
		t.Fatalf("product preview count=%d; want both old products", preview.CandidateCount)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(`{"type":"products","start_date":"2020-01-01","end_date":"2020-01-02","candidate_ids":["`+unusedID+`","`+usedID+`"]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())
	NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
	if response.Code != http.StatusOK {
		t.Fatalf("historical product cleanup status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data HistoricalCleanupResult `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Deleted != 2 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 {
		t.Fatalf("product cleanup result=%+v", envelope.Data)
	}
	var productCount, itemCount int
	if err := db.Get(&productCount, `SELECT COUNT(*) FROM products`); err != nil {
		t.Fatal(err)
	}
	if err := db.Get(&itemCount, `SELECT COUNT(*) FROM inventory_items WHERE product_id=?`, usedID); err != nil {
		t.Fatal(err)
	}
	if productCount != 0 || itemCount != 0 {
		t.Fatalf("remaining products=%d linked inventory items=%d; want all selected products and linked stock deleted", productCount, itemCount)
	}
}

func TestHistoricalDirectoryCleanupHardDeletesSelectedCustomersAndSuppliersSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE customers (id TEXT PRIMARY KEY, code TEXT, name TEXT, email TEXT, phone TEXT, address TEXT, city TEXT, country TEXT, tax_id TEXT, credit_limit REAL NOT NULL DEFAULT 0, current_balance REAL NOT NULL DEFAULT 0, notes TEXT, is_active INTEGER NOT NULL DEFAULT 1, created_at TEXT, updated_at TEXT);
		CREATE TABLE suppliers (id TEXT PRIMARY KEY, code TEXT, name TEXT, email TEXT, phone TEXT, address TEXT, city TEXT, country TEXT, tax_id TEXT, payment_terms TEXT, credit_limit REAL NOT NULL DEFAULT 0, current_balance REAL NOT NULL DEFAULT 0, notes TEXT, is_active INTEGER NOT NULL DEFAULT 1, created_at TEXT, updated_at TEXT);
		CREATE TABLE sales (id TEXT PRIMARY KEY, customer_id TEXT, created_at TEXT);
		CREATE TABLE purchases (id TEXT PRIMARY KEY, supplier_id TEXT, total_amount REAL DEFAULT 0, paid_amount REAL DEFAULT 0, status TEXT DEFAULT 'pending', purchase_date TEXT, created_at TEXT);
		CREATE TABLE supplier_returns (id TEXT PRIMARY KEY, supplier_id TEXT, status TEXT, refund_amount REAL DEFAULT 0);
		CREATE TABLE supplier_ledger (id TEXT PRIMARY KEY, supplier_id TEXT, type TEXT, transaction_type TEXT, amount REAL, reference_id TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	freeCustomerID, linkedCustomerID := uuid.NewString(), uuid.NewString()
	freeSupplierID, linkedSupplierID := uuid.NewString(), uuid.NewString()
	for _, id := range []string{freeCustomerID, linkedCustomerID} {
		if _, err := db.Exec(`INSERT INTO customers (id,code,name,created_at) VALUES (?,?,?, '2020-01-01')`, id, id, "Test customer"); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{freeSupplierID, linkedSupplierID} {
		if _, err := db.Exec(`INSERT INTO suppliers (id,code,name,created_at,updated_at) VALUES (?,?,?, '2020-01-01','2020-01-01')`, id, id, "Test supplier"); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		entity     string
		freeID     string
		linkedID   string
		countQuery string
		wantCount  int
	}{
		{entity: "customers", freeID: freeCustomerID, linkedID: linkedCustomerID, countQuery: `SELECT COUNT(*) FROM customers`, wantCount: 0},
		{entity: "suppliers", freeID: freeSupplierID, linkedID: linkedSupplierID, countQuery: `SELECT COUNT(*) FROM suppliers`, wantCount: 0},
	} {
		gin.SetMode(gin.TestMode)
		response := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(response)
		ctx.Request = httptest.NewRequest("POST", "/settings/history-cleanup", strings.NewReader(`{"type":"`+tc.entity+`","start_date":"2020-01-01","end_date":"2020-01-01","candidate_ids":["`+tc.freeID+`","`+tc.linkedID+`"]}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Set("user_id", uuid.New())
		NewDatabaseHandler(db).RunHistoricalCleanup(ctx)
		if response.Code != http.StatusOK {
			t.Fatalf("%s cleanup status=%d body=%s", tc.entity, response.Code, response.Body.String())
		}
		var envelope struct {
			Data HistoricalCleanupResult `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Data.Deleted != 2 || envelope.Data.Blocked != 0 || envelope.Data.Failed != 0 {
			t.Fatalf("%s cleanup result=%+v", tc.entity, envelope.Data)
		}
		var count int
		if err := db.Get(&count, tc.countQuery); err != nil {
			t.Fatal(err)
		}
		if count != tc.wantCount {
			t.Fatalf("%s count=%d; want %d", tc.entity, count, tc.wantCount)
		}
	}
}

func TestCleanupPreviewSkipsMissingLegacyTablesSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE notifications (id TEXT, expires_at TEXT, status TEXT, created_at TEXT)`); err != nil {
		t.Fatal(err)
	}
	preview, err := buildCleanupPreview(context.Background(), db, true)
	if err != nil {
		t.Fatalf("preview legacy cleanup: %v", err)
	}
	if preview.TotalCount != 0 || len(preview.Categories) != 0 {
		t.Fatalf("legacy preview = %+v, want incomplete notification schema skipped", preview)
	}
}

func TestLocalCleanupEndpointLabelsAndUsesSQLiteTarget(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL)`); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("GET", "/settings/database/cleanup/preview", nil)
	NewDatabaseHandler(db).PreviewLocalCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("local cleanup preview status = %d, body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data CleanupPreview `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Target != "local" {
		t.Fatalf("cleanup target = %q, want local", envelope.Data.Target)
	}
}

func TestCleanupSupportsCloudNotificationSchemaWithIsRead(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE notifications (id TEXT PRIMARY KEY, title TEXT, message TEXT, data TEXT, is_read INTEGER, created_at TEXT)`); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().AddDate(0, 0, -120).Format("2006-01-02 15:04:05")
	future := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")
	if _, err := db.Exec(`INSERT INTO notifications VALUES ('read-old','old','message','{}',1,?),('unread-old','old','message','{}',0,?),('read-new','new','message','{}',1,?)`, old, old, future); err != nil {
		t.Fatal(err)
	}

	preview, err := buildCleanupPreview(context.Background(), db, true)
	if err != nil {
		t.Fatalf("preview cloud-compatible notification schema: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Fatalf("notification preview count = %d, want only the old read row", preview.TotalCount)
	}
	specs, err := cleanupAvailableSpecs(context.Background(), db, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range specs {
		if spec.key == "old_notifications" && !strings.Contains(spec.wherePG, "is_read IS TRUE") {
			t.Fatalf("PostgreSQL notification predicate = %q, want is_read boolean support", spec.wherePG)
		}
	}
}

func TestCleanupDeletesOnlyExpiredReservationsAndOrphanSpecificationsSQLite(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE reservations (id TEXT PRIMARY KEY, notes TEXT, status TEXT, updated_at TEXT, expires_at TEXT);
		CREATE TABLE inventory_items (id TEXT PRIMARY KEY);
		CREATE TABLE item_specification_values (id TEXT PRIMARY KEY, inventory_item_id TEXT, value_text TEXT, value_number REAL, value_boolean INTEGER);
	`); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().AddDate(0, 0, -120).Format("2006-01-02 15:04:05")
	future := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")
	if _, err := db.Exec(`INSERT INTO reservations VALUES ('stale-active','expired','active',?,?),('live-active','live','active',?,?)`, old, old, future, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO inventory_items VALUES ('item-present')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO item_specification_values VALUES ('orphan','item-missing','value',NULL,NULL),('linked','item-present','value',NULL,NULL)`); err != nil {
		t.Fatal(err)
	}

	preview, err := buildCleanupPreview(context.Background(), db, true)
	if err != nil {
		t.Fatalf("preview cleanup: %v", err)
	}
	if preview.TotalCount != 2 {
		t.Fatalf("preview candidates = %d, want expired reservation and orphan specification", preview.TotalCount)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/cleanup", nil)
	NewDatabaseHandler(db).RunCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("cleanup status = %d, body=%s", response.Code, response.Body.String())
	}
	var staleReservationCount, liveReservationCount, orphanCount, linkedCount int
	for query, destination := range map[string]*int{
		`SELECT COUNT(*) FROM reservations WHERE id='stale-active'`:        &staleReservationCount,
		`SELECT COUNT(*) FROM reservations WHERE id='live-active'`:         &liveReservationCount,
		`SELECT COUNT(*) FROM item_specification_values WHERE id='orphan'`: &orphanCount,
		`SELECT COUNT(*) FROM item_specification_values WHERE id='linked'`: &linkedCount,
	} {
		if err := db.Get(destination, query); err != nil {
			t.Fatal(err)
		}
	}
	if staleReservationCount != 0 || orphanCount != 0 || liveReservationCount != 1 || linkedCount != 1 {
		t.Fatalf("remaining stale=%d live=%d orphan=%d linked=%d", staleReservationCount, liveReservationCount, orphanCount, linkedCount)
	}
}

func TestCleanupRemovesExpiredLocalSessionsAndOldSyncedQueueOnly(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE local_sessions (id TEXT PRIMARY KEY, expires_at TEXT, access_token TEXT, refresh_token TEXT);
		CREATE TABLE sync_queue (id TEXT PRIMARY KEY, payload TEXT, synced_at TEXT);
	`); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().AddDate(0, 0, -120).Format("2006-01-02 15:04:05")
	future := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")
	if _, err := db.Exec(`INSERT INTO local_sessions VALUES ('expired',?,'access','refresh'),('active',?,'access','refresh')`, old, future); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO sync_queue VALUES ('synced','{}',?),('pending','{}',NULL)`, old); err != nil {
		t.Fatal(err)
	}
	preview, err := buildCleanupPreview(context.Background(), db, true)
	if err != nil {
		t.Fatalf("preview cleanup: %v", err)
	}
	if preview.TotalCount != 2 {
		t.Fatalf("preview count = %d, want one expired session and one old synced operation", preview.TotalCount)
	}
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/settings/cleanup", nil)
	NewDatabaseHandler(db).RunCleanup(ctx)
	if response.Code != 200 {
		t.Fatalf("cleanup status = %d, body=%s", response.Code, response.Body.String())
	}
	var expired, active, synced, pending int
	for query, destination := range map[string]*int{
		`SELECT COUNT(*) FROM local_sessions WHERE id='expired'`: &expired,
		`SELECT COUNT(*) FROM local_sessions WHERE id='active'`:  &active,
		`SELECT COUNT(*) FROM sync_queue WHERE id='synced'`:      &synced,
		`SELECT COUNT(*) FROM sync_queue WHERE id='pending'`:     &pending,
	} {
		if err := db.Get(destination, query); err != nil {
			t.Fatal(err)
		}
	}
	if expired != 0 || active != 1 || synced != 0 || pending != 1 {
		t.Fatalf("expired=%d active=%d synced=%d pending=%d", expired, active, synced, pending)
	}
}
