package paymentproviders

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending           = "pending"
	StatusProcessing        = "processing"
	StatusPaid              = "paid"
	StatusFailed            = "failed"
	StatusCancelled         = "cancelled"
	StatusRefunded          = "refunded"
	StatusPartiallyRefunded = "partially_refunded"
)

var (
	ErrNotConfigured       = errors.New("payment provider is not configured")
	ErrProviderUnavailable = errors.New("payment provider is unavailable")
	ErrInvalidProviderData = errors.New("invalid payment provider response")
	ErrInvalidTransition   = errors.New("invalid payment status transition")
)

type ProviderName string

const (
	ProviderCardcom ProviderName = "cardcom"
	ProviderPayMe   ProviderName = "payme"
	ProviderGrow    ProviderName = "grow"
	ProviderPalPay  ProviderName = "palpay"
	ProviderStripe  ProviderName = "stripe"
	ProviderPayPal  ProviderName = "paypal"
)

type Config struct {
	Provider      ProviderName
	Environment   string
	Currency      string
	MerchantID    string
	TerminalID    string
	APIKey        string
	APISecret     string
	WebhookURL    string
	WebhookSecret string
	BaseURL       string
	HTTPTimeout   time.Duration
}

type PaymentRequest struct {
	PaymentID      uuid.UUID
	OrderID        uuid.UUID
	SaleID         *uuid.UUID
	Amount         int64
	Currency       string
	Description    string
	IdempotencyKey string
	SuccessURL     string
	FailureURL     string
	CancelURL      string
	Metadata       map[string]string
}

type PaymentResult struct {
	ProviderPaymentID     string
	ProviderTransactionID string
	Status                string
	Amount                int64
	Currency              string
	CheckoutURL           string
	Raw                   []byte
}

type StatusRequest struct {
	ProviderPaymentID     string
	ProviderTransactionID string
	PaymentID             uuid.UUID
}

type RefundRequest struct {
	ProviderPaymentID     string
	ProviderTransactionID string
	PaymentID             uuid.UUID
	Amount                int64
	Currency              string
	IdempotencyKey        string
	Reason                string
}

type RefundResult struct {
	ProviderRefundID string
	Status           string
	Amount           int64
	Raw              []byte
}

type WebhookEvent struct {
	EventID           string
	ProviderPaymentID string
	Status            string
	Amount            int64
	Currency          string
	Raw               []byte
}

type PaymentProvider interface {
	Name() ProviderName
	TestConnection(ctx context.Context) error
	CreatePayment(ctx context.Context, request PaymentRequest) (PaymentResult, error)
	GetPaymentStatus(ctx context.Context, request StatusRequest) (PaymentResult, error)
	VerifyPayment(ctx context.Context, request StatusRequest) (PaymentResult, error)
	CancelPayment(ctx context.Context, request StatusRequest) (PaymentResult, error)
	RefundPayment(ctx context.Context, request RefundRequest) (RefundResult, error)
	HandleWebhook(ctx context.Context, headers map[string]string, body []byte) (WebhookEvent, error)
}

func IsTerminalStatus(status string) bool {
	switch status {
	case StatusPaid, StatusFailed, StatusCancelled, StatusRefunded, StatusPartiallyRefunded:
		return true
	default:
		return false
	}
}

func CanTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "":
		return to == StatusPending
	case StatusPending:
		return to == StatusProcessing || to == StatusPaid || to == StatusFailed || to == StatusCancelled
	case StatusProcessing:
		return to == StatusPaid || to == StatusFailed || to == StatusCancelled
	case StatusPaid:
		return to == StatusPartiallyRefunded || to == StatusRefunded
	case StatusPartiallyRefunded:
		return to == StatusPartiallyRefunded || to == StatusRefunded
	default:
		return false
	}
}
