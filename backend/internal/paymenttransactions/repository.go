package paymenttransactions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/internal/paymentproviders"
)

type Repository struct {
	db *sqlx.DB
}

var ErrDuplicateWebhook = errors.New("duplicate payment webhook event")

func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	var transaction Transaction
	if err := r.db.GetContext(ctx, &transaction, `SELECT id, order_id, sale_id, payment_id, provider, provider_payment_id, provider_transaction_id, status, amount_minor, currency, idempotency_key, checkout_url, failure_code, failure_message, metadata, created_at, updated_at, paid_at, cancelled_at FROM payment_transactions WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *Repository) GetByIdempotencyKey(ctx context.Context, provider, key string) (*Transaction, error) {
	var transaction Transaction
	err := r.db.GetContext(ctx, &transaction, `SELECT id, order_id, sale_id, payment_id, provider, provider_payment_id, provider_transaction_id, status, amount_minor, currency, idempotency_key, checkout_url, failure_code, failure_message, metadata, created_at, updated_at, paid_at, cancelled_at FROM payment_transactions WHERE provider = $1 AND idempotency_key = $2`, provider, key)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *Repository) GetByProviderPaymentID(ctx context.Context, provider, providerPaymentID string) (*Transaction, error) {
	var transaction Transaction
	err := r.db.GetContext(ctx, &transaction, `SELECT id, order_id, sale_id, payment_id, provider, provider_payment_id, provider_transaction_id, status, amount_minor, currency, idempotency_key, checkout_url, failure_code, failure_message, metadata, created_at, updated_at, paid_at, cancelled_at FROM payment_transactions WHERE provider = $1 AND provider_payment_id = $2`, provider, providerPaymentID)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *Repository) Create(ctx context.Context, transaction *Transaction) (*Transaction, error) {
	if existing, err := r.GetByIdempotencyKey(ctx, transaction.Provider, transaction.IdempotencyKey); err == nil {
		return existing, nil
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	if transaction.ID == uuid.Nil {
		transaction.ID = uuid.New()
	}
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now().UTC()
	}
	if transaction.UpdatedAt.IsZero() {
		transaction.UpdatedAt = transaction.CreatedAt
	}
	if transaction.Metadata == "" {
		transaction.Metadata = "{}"
	}
	query := `INSERT INTO payment_transactions (id, order_id, sale_id, payment_id, provider, provider_payment_id, provider_transaction_id, status, amount_minor, currency, idempotency_key, checkout_url, failure_code, failure_message, metadata, created_at, updated_at, paid_at, cancelled_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`
	args := []any{transaction.ID, transaction.OrderID, transaction.SaleID, transaction.PaymentID, transaction.Provider, transaction.ProviderPaymentID, transaction.ProviderTransactionID, transaction.Status, transaction.AmountMinor, transaction.Currency, transaction.IdempotencyKey, transaction.CheckoutURL, transaction.FailureCode, transaction.FailureMessage, transaction.Metadata, transaction.CreatedAt, transaction.UpdatedAt, transaction.PaidAt, transaction.CancelledAt}
	if dbutil.IsSQLite(r.db) {
		for i, arg := range args {
			switch value := arg.(type) {
			case time.Time:
				args[i] = value.UTC().Format(time.RFC3339Nano)
			case *time.Time:
				if value != nil {
					args[i] = value.UTC().Format(time.RFC3339Nano)
				}
			}
		}
	}
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return r.GetByIdempotencyKey(ctx, transaction.Provider, transaction.IdempotencyKey)
		}
		return nil, fmt.Errorf("create payment transaction: %w", err)
	}
	return transaction, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, transaction *Transaction, status string) error {
	if err := ValidateStatusTransition(transaction.Status, status); err != nil {
		return err
	}
	transaction.Status = status
	transaction.UpdatedAt = time.Now().UTC()
	if status == paymentproviders.StatusPaid {
		now := transaction.UpdatedAt
		transaction.PaidAt = &now
	}
	if status == paymentproviders.StatusCancelled {
		now := transaction.UpdatedAt
		transaction.CancelledAt = &now
	}
	_, err := r.db.ExecContext(ctx, `UPDATE payment_transactions SET status = $1, updated_at = $2, paid_at = $3, cancelled_at = $4, provider_payment_id = $5, provider_transaction_id = $6, checkout_url = $7, failure_code = $8, failure_message = $9 WHERE id = $10`, transaction.Status, transaction.UpdatedAt, transaction.PaidAt, transaction.CancelledAt, transaction.ProviderPaymentID, transaction.ProviderTransactionID, transaction.CheckoutURL, transaction.FailureCode, transaction.FailureMessage, transaction.ID)
	return err
}

// FinalizePaid records the legacy payment and sale projection together with
// the external transaction state. This keeps the existing financial tables as
// the source of truth instead of introducing a second ledger.
func (r *Repository) FinalizePaid(ctx context.Context, transaction *Transaction) error {
	if err := ValidateStatusTransition(transaction.Status, paymentproviders.StatusPaid); err != nil {
		return err
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	paymentID := transaction.ID
	transaction.PaymentID = &paymentID
	reference := fmt.Sprintf("%s:%s", transaction.Provider, valueOrEmpty(transaction.ProviderPaymentID))
	amount := float64(transaction.AmountMinor) / 100
	if dbutil.IsSQLite(r.db) {
		_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO payments (id, transaction_number, sale_id, amount, payment_method, payment_status, payment_date, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$7,$7)`, paymentID, reference, transaction.SaleID, amount, "electronic", "completed", now.Format(time.RFC3339Nano))
	} else {
		_, err = tx.ExecContext(ctx, `INSERT INTO payments (id, reference_number, sale_id, amount, payment_method, payment_date, notes, created_at, updated_at) SELECT $1,$2,$3,$4,$5,$6,$7,$8,$8 WHERE NOT EXISTS (SELECT 1 FROM payments WHERE reference_number = $2)`, paymentID, reference, transaction.SaleID, amount, "electronic", now, "provider transaction: "+reference, now)
	}
	if err != nil {
		return fmt.Errorf("record electronic payment: %w", err)
	}
	if transaction.SaleID != nil {
		if _, err = tx.ExecContext(ctx, `UPDATE sales SET paid_amount = COALESCE(paid_amount, 0) + $1, remaining_amount = MAX(0, COALESCE(total_amount, 0) - (COALESCE(paid_amount, 0) + $1)), payment_method = $2, payment_status = CASE WHEN COALESCE(paid_amount, 0) + $1 >= COALESCE(total_amount, 0) THEN 'paid' ELSE 'partial' END, updated_at = $3 WHERE id = $4`, amount, "electronic", now, *transaction.SaleID); err != nil {
			return fmt.Errorf("apply electronic payment to sale: %w", err)
		}
	}
	transaction.Status = paymentproviders.StatusPaid
	transaction.UpdatedAt = now
	transaction.PaidAt = &now
	if _, err = tx.ExecContext(ctx, `UPDATE payment_transactions SET payment_id = $1, status = $2, updated_at = $3, paid_at = $4, provider_payment_id = $5, provider_transaction_id = $6 WHERE id = $7`, paymentID, transaction.Status, now, now, transaction.ProviderPaymentID, transaction.ProviderTransactionID, transaction.ID); err != nil {
		return fmt.Errorf("finalize payment transaction: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *Repository) CreateRefund(ctx context.Context, refund *Refund) error {
	if refund.ID == uuid.Nil {
		refund.ID = uuid.New()
	}
	if refund.CreatedAt.IsZero() {
		refund.CreatedAt = time.Now().UTC()
	}
	if refund.UpdatedAt.IsZero() {
		refund.UpdatedAt = refund.CreatedAt
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO payment_refunds (id, payment_transaction_id, provider_refund_id, amount_minor, currency, status, idempotency_key, reason, failure_message, created_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, refund.ID, refund.PaymentTransactionID, refund.ProviderRefundID, refund.AmountMinor, refund.Currency, refund.Status, refund.IdempotencyKey, refund.Reason, refund.FailureMessage, refund.CreatedBy, refund.CreatedAt, refund.UpdatedAt)
	return err
}

func (r *Repository) GetRefundByIdempotencyKey(ctx context.Context, key string) (*Refund, error) {
	var refund Refund
	err := r.db.GetContext(ctx, &refund, `SELECT id, payment_transaction_id, provider_refund_id, amount_minor, currency, status, idempotency_key, reason, failure_message, created_by, created_at, updated_at FROM payment_refunds WHERE idempotency_key = $1`, key)
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *Repository) RecordWebhookEvent(ctx context.Context, event *WebhookEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.ReceivedAt.IsZero() {
		event.ReceivedAt = time.Now().UTC()
	}
	if event.Status == "" {
		event.Status = "received"
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO payment_webhook_events (id, provider, provider_event_id, event_type, payment_transaction_id, payload, status, error_message, received_at, processed_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, event.ID, event.Provider, event.ProviderEventID, event.EventType, event.PaymentTransactionID, event.Payload, event.Status, event.ErrorMessage, event.ReceivedAt, event.ProcessedAt)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "unique") {
		return ErrDuplicateWebhook
	}
	return err
}
