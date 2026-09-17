package paymenttransactions

import (
	"time"

	"github.com/google/uuid"
	"github.com/partflow/smart-store/internal/paymentproviders"
)

type Transaction struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	OrderID               *uuid.UUID `json:"order_id,omitempty" db:"order_id"`
	SaleID                *uuid.UUID `json:"sale_id,omitempty" db:"sale_id"`
	PaymentID             *uuid.UUID `json:"payment_id,omitempty" db:"payment_id"`
	Provider              string     `json:"provider" db:"provider"`
	ProviderPaymentID     *string    `json:"provider_payment_id,omitempty" db:"provider_payment_id"`
	ProviderTransactionID *string    `json:"provider_transaction_id,omitempty" db:"provider_transaction_id"`
	Status                string     `json:"status" db:"status"`
	AmountMinor           int64      `json:"amount_minor" db:"amount_minor"`
	Currency              string     `json:"currency" db:"currency"`
	IdempotencyKey        string     `json:"idempotency_key" db:"idempotency_key"`
	CheckoutURL           *string    `json:"checkout_url,omitempty" db:"checkout_url"`
	FailureCode           *string    `json:"failure_code,omitempty" db:"failure_code"`
	FailureMessage        *string    `json:"failure_message,omitempty" db:"failure_message"`
	Metadata              string     `json:"metadata" db:"metadata"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
	PaidAt                *time.Time `json:"paid_at,omitempty" db:"paid_at"`
	CancelledAt           *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
}

type Refund struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	PaymentTransactionID uuid.UUID  `json:"payment_transaction_id" db:"payment_transaction_id"`
	ProviderRefundID     *string    `json:"provider_refund_id,omitempty" db:"provider_refund_id"`
	AmountMinor          int64      `json:"amount_minor" db:"amount_minor"`
	Currency             string     `json:"currency" db:"currency"`
	Status               string     `json:"status" db:"status"`
	IdempotencyKey       string     `json:"idempotency_key" db:"idempotency_key"`
	Reason               *string    `json:"reason,omitempty" db:"reason"`
	FailureMessage       *string    `json:"failure_message,omitempty" db:"failure_message"`
	CreatedBy            *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

type WebhookEvent struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	Provider             string     `json:"provider" db:"provider"`
	ProviderEventID      string     `json:"provider_event_id" db:"provider_event_id"`
	EventType            *string    `json:"event_type,omitempty" db:"event_type"`
	PaymentTransactionID *uuid.UUID `json:"payment_transaction_id,omitempty" db:"payment_transaction_id"`
	Payload              string     `json:"payload" db:"payload"`
	Status               string     `json:"status" db:"status"`
	ErrorMessage         *string    `json:"error_message,omitempty" db:"error_message"`
	ReceivedAt           time.Time  `json:"received_at" db:"received_at"`
	ProcessedAt          *time.Time `json:"processed_at,omitempty" db:"processed_at"`
}

func ValidateStatusTransition(from, to string) error {
	if !paymentproviders.CanTransition(from, to) {
		return paymentproviders.ErrInvalidTransition
	}
	return nil
}
