package paymenttransactions

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/paymentproviders"
	_ "modernc.org/sqlite"
)

func TestFinalizePaidWritesExistingFinancialTables(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE sales (id TEXT PRIMARY KEY, total_amount REAL NOT NULL, paid_amount REAL DEFAULT 0, remaining_amount REAL DEFAULT 0, payment_method TEXT, payment_status TEXT, updated_at TEXT);
		CREATE TABLE payments (id TEXT PRIMARY KEY, transaction_number TEXT UNIQUE, sale_id TEXT, amount REAL, payment_method TEXT, payment_status TEXT, payment_date TEXT, created_at TEXT, updated_at TEXT);
		CREATE TABLE payment_transactions (id TEXT PRIMARY KEY, order_id TEXT, sale_id TEXT, payment_id TEXT, provider TEXT, provider_payment_id TEXT, provider_transaction_id TEXT, status TEXT, amount_minor INTEGER, currency TEXT, idempotency_key TEXT, checkout_url TEXT, failure_code TEXT, failure_message TEXT, metadata TEXT, created_at TEXT, updated_at TEXT, paid_at TEXT, cancelled_at TEXT, UNIQUE(provider, idempotency_key));
	`)
	if err != nil {
		t.Fatal(err)
	}
	saleID := uuid.New()
	if _, err := db.Exec(`INSERT INTO sales (id,total_amount,paid_amount,remaining_amount,payment_status) VALUES (?,?,?,?,?)`, saleID, 25, 0, 25, "pending"); err != nil {
		t.Fatal(err)
	}
	transaction := &Transaction{
		ID: uuid.New(), SaleID: &saleID, Provider: "cardcom", ProviderPaymentID: stringPtr("lp-1"), ProviderTransactionID: stringPtr("9876"), Status: paymentproviders.StatusProcessing, AmountMinor: 2500, Currency: "ILS", IdempotencyKey: "sale-1", Metadata: "{}", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	repo := NewRepository(db)
	if _, err := repo.Create(context.Background(), transaction); err != nil {
		t.Fatal(err)
	}
	if err := repo.FinalizePaid(context.Background(), transaction); err != nil {
		t.Fatalf("FinalizePaid failed: %v", err)
	}
	var paid, remaining float64
	var status string
	if err := db.QueryRow(`SELECT paid_amount, remaining_amount, payment_status FROM sales WHERE id = ?`, saleID).Scan(&paid, &remaining, &status); err != nil {
		t.Fatal(err)
	}
	if paid != 25 || remaining != 0 || status != "paid" {
		t.Fatalf("sale projection = paid %v remaining %v status %q", paid, remaining, status)
	}
	var paymentCount int
	if err := db.Get(&paymentCount, `SELECT COUNT(*) FROM payments WHERE sale_id = ?`, saleID); err != nil {
		t.Fatal(err)
	}
	if paymentCount != 1 {
		t.Fatalf("payment rows = %d, want 1", paymentCount)
	}
}

func TestRecordWebhookEventRejectsDuplicateBeforeProcessing(t *testing.T) {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE payment_webhook_events (id TEXT PRIMARY KEY, provider TEXT NOT NULL, provider_event_id TEXT NOT NULL, event_type TEXT, payment_transaction_id TEXT, payload TEXT NOT NULL, status TEXT NOT NULL, error_message TEXT, received_at TEXT NOT NULL, processed_at TEXT, UNIQUE(provider, provider_event_id));`); err != nil {
		t.Fatal(err)
	}
	repository := NewRepository(db)
	event := &WebhookEvent{Provider: "cardcom", ProviderEventID: "event-1", Payload: "{}"}
	if err := repository.RecordWebhookEvent(context.Background(), event); err != nil {
		t.Fatalf("first webhook insert failed: %v", err)
	}
	if err := repository.RecordWebhookEvent(context.Background(), &WebhookEvent{Provider: "cardcom", ProviderEventID: "event-1", Payload: "{}"}); err != ErrDuplicateWebhook {
		t.Fatalf("duplicate webhook error = %v, want %v", err, ErrDuplicateWebhook)
	}
}

func stringPtr(value string) *string { return &value }
