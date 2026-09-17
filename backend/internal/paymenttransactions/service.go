package paymenttransactions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/paymentproviders"
	"github.com/partflow/smart-store/internal/secrets"
)

type ReturnRefundResult struct {
	RefundID string
	Status   string
}

type Service struct {
	repo      *Repository
	providers map[paymentproviders.ProviderName]paymentproviders.PaymentProvider
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo, providers: make(map[paymentproviders.ProviderName]paymentproviders.PaymentProvider)}
}

func NewConfiguredService(db *sqlx.DB) *Service {
	service := NewService(NewRepository(db))
	var providerName, environment, currency, publicKey, secretKey, merchantID, terminalID, webhookURL, webhookSecret string
	_ = db.Get(&providerName, `SELECT value FROM settings WHERE key = 'payment_provider'`)
	_ = db.Get(&environment, `SELECT value FROM settings WHERE key = 'payment_environment'`)
	_ = db.Get(&currency, `SELECT value FROM settings WHERE key = 'currency'`)
	_ = db.Get(&publicKey, `SELECT value FROM settings WHERE key = 'payment_public_key'`)
	_ = db.Get(&secretKey, `SELECT value FROM settings WHERE key = 'payment_secret_key'`)
	_ = db.Get(&merchantID, `SELECT value FROM settings WHERE key = 'payment_merchant_id'`)
	_ = db.Get(&terminalID, `SELECT value FROM settings WHERE key = 'payment_terminal_id'`)
	_ = db.Get(&webhookURL, `SELECT value FROM settings WHERE key = 'payment_webhook_url'`)
	_ = db.Get(&webhookSecret, `SELECT value FROM settings WHERE key = 'payment_webhook_secret'`)
	if decrypted, err := secrets.Decrypt(secretKey); err == nil {
		secretKey = decrypted
	}
	if decrypted, err := secrets.Decrypt(webhookSecret); err == nil {
		webhookSecret = decrypted
	}
	name := paymentproviders.ProviderName(providerName)
	if name == "" || name == "manual" {
		return service
	}
	adapter, err := paymentproviders.NewRegistry().Build(name, paymentproviders.Config{
		Provider: name, Environment: environment, Currency: currency, APIKey: publicKey, APISecret: secretKey,
		MerchantID: merchantID, TerminalID: terminalID, WebhookURL: webhookURL, WebhookSecret: webhookSecret,
	})
	if err == nil {
		service.RegisterProvider(adapter)
	}
	return service
}

func (s *Service) RefundForReturn(ctx context.Context, saleID, returnID uuid.UUID, amountMinor int64, createdBy *uuid.UUID) (ReturnRefundResult, error) {
	var transactionID uuid.UUID
	if err := s.repo.db.GetContext(ctx, &transactionID, `SELECT id FROM payment_transactions WHERE sale_id = $1 AND status IN ('paid','partially_refunded') ORDER BY created_at DESC LIMIT 1`, saleID); err != nil {
		return ReturnRefundResult{}, fmt.Errorf("electronic payment not found for sale: %w", err)
	}
	refund, transaction, err := s.RefundPayment(ctx, transactionID, RefundRequestForReturn(returnID, amountMinor), createdBy)
	if err != nil {
		return ReturnRefundResult{}, err
	}
	return ReturnRefundResult{RefundID: refund.ID.String(), Status: transaction.Status}, nil
}

func RefundRequestForReturn(returnID uuid.UUID, amountMinor int64) paymentproviders.RefundRequest {
	return paymentproviders.RefundRequest{
		Amount:         amountMinor,
		IdempotencyKey: "return-refund:" + returnID.String(),
		Reason:         "customer_return:" + returnID.String(),
	}
}

func (s *Service) RegisterProvider(provider paymentproviders.PaymentProvider) {
	s.providers[provider.Name()] = provider
}

func (s *Service) provider(name paymentproviders.ProviderName) (paymentproviders.PaymentProvider, error) {
	provider, ok := s.providers[name]
	if !ok {
		return nil, fmt.Errorf("payment provider %q is not registered", name)
	}
	return provider, nil
}

func (s *Service) CreatePayment(ctx context.Context, request paymentproviders.PaymentRequest, providerName paymentproviders.ProviderName) (*Transaction, error) {
	provider, err := s.provider(providerName)
	if err != nil {
		return nil, err
	}
	if request.PaymentID == uuid.Nil {
		request.PaymentID = uuid.New()
	}
	if request.IdempotencyKey == "" {
		request.IdempotencyKey = request.PaymentID.String()
	}
	transaction, err := s.repo.Create(ctx, &Transaction{
		ID:             request.PaymentID,
		OrderID:        &request.OrderID,
		SaleID:         request.SaleID,
		PaymentID:      &request.PaymentID,
		Provider:       string(providerName),
		Status:         paymentproviders.StatusPending,
		AmountMinor:    request.Amount,
		Currency:       request.Currency,
		IdempotencyKey: request.IdempotencyKey,
		Metadata:       "{}",
	})
	if err != nil {
		return nil, err
	}
	if transaction.Status != paymentproviders.StatusPending {
		return transaction, nil
	}

	result, err := provider.CreatePayment(ctx, request)
	if err != nil {
		transaction.Status = paymentproviders.StatusFailed
		message := err.Error()
		transaction.FailureMessage = &message
		_ = s.repo.UpdateStatus(ctx, transaction, paymentproviders.StatusFailed)
		return nil, err
	}
	setProviderIDs(transaction, result)
	transaction.CheckoutURL = stringPointer(result.CheckoutURL)
	if result.Status != paymentproviders.StatusPending {
		if err := s.repo.UpdateStatus(ctx, transaction, result.Status); err != nil {
			return nil, err
		}
	} else if err := s.repo.UpdateStatus(ctx, transaction, paymentproviders.StatusPending); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *Service) VerifyPayment(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	provider, err := s.provider(paymentproviders.ProviderName(transaction.Provider))
	if err != nil {
		return nil, err
	}
	providerID := valueOrEmpty(transaction.ProviderPaymentID)
	result, err := provider.VerifyPayment(ctx, paymentproviders.StatusRequest{ProviderPaymentID: providerID, PaymentID: id})
	if err != nil {
		return nil, err
	}
	setProviderIDs(transaction, result)
	if result.Status == paymentproviders.StatusPaid {
		if err := s.repo.FinalizePaid(ctx, transaction); err != nil {
			return nil, err
		}
	} else if err := s.repo.UpdateStatus(ctx, transaction, result.Status); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *Service) CancelPayment(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if transaction.Status != paymentproviders.StatusPending && transaction.Status != paymentproviders.StatusProcessing {
		return nil, fmt.Errorf("payment cannot be cancelled from status: %s", transaction.Status)
	}
	provider, err := s.provider(paymentproviders.ProviderName(transaction.Provider))
	if err != nil {
		return nil, err
	}
	result, err := provider.CancelPayment(ctx, paymentproviders.StatusRequest{
		ProviderPaymentID:     valueOrEmpty(transaction.ProviderPaymentID),
		ProviderTransactionID: valueOrEmpty(transaction.ProviderTransactionID),
		PaymentID:             transaction.ID,
	})
	if err != nil {
		return nil, err
	}
	setProviderIDs(transaction, result)
	if err := s.repo.UpdateStatus(ctx, transaction, paymentproviders.StatusCancelled); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *Service) RefundPayment(ctx context.Context, id uuid.UUID, request paymentproviders.RefundRequest, createdBy *uuid.UUID) (*Refund, *Transaction, error) {
	transaction, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if transaction.Status != paymentproviders.StatusPaid && transaction.Status != paymentproviders.StatusPartiallyRefunded {
		return nil, nil, fmt.Errorf("payment must be paid before refund: %s", transaction.Status)
	}
	if request.ProviderPaymentID == "" {
		request.ProviderPaymentID = valueOrEmpty(transaction.ProviderPaymentID)
	}
	if request.ProviderTransactionID == "" {
		request.ProviderTransactionID = valueOrEmpty(transaction.ProviderTransactionID)
	}
	if request.Currency == "" {
		request.Currency = transaction.Currency
	}
	if request.IdempotencyKey == "" {
		return nil, nil, fmt.Errorf("refund idempotency key is required")
	}
	if existing, err := s.repo.GetRefundByIdempotencyKey(ctx, request.IdempotencyKey); err == nil {
		return existing, transaction, nil
	}
	var refundedMinor int64
	if err := s.repo.db.GetContext(ctx, &refundedMinor, `SELECT COALESCE(SUM(amount_minor), 0) FROM payment_refunds WHERE payment_transaction_id = $1 AND status IN ('refunded','partially_refunded')`, transaction.ID); err != nil {
		return nil, nil, err
	}
	if request.Amount <= 0 || request.Amount > transaction.AmountMinor-refundedMinor {
		return nil, nil, fmt.Errorf("refund amount must be greater than zero and no more than the payment amount")
	}
	provider, err := s.provider(paymentproviders.ProviderName(transaction.Provider))
	if err != nil {
		return nil, nil, err
	}
	result, err := provider.RefundPayment(ctx, request)
	if err != nil {
		return nil, nil, err
	}
	if result.ProviderRefundID == "" || (result.Status != paymentproviders.StatusRefunded && result.Status != paymentproviders.StatusPartiallyRefunded) || result.Amount != 0 && result.Amount != request.Amount {
		return nil, nil, fmt.Errorf("provider refund response could not be verified")
	}
	refund := &Refund{
		PaymentTransactionID: transaction.ID,
		ProviderRefundID:     stringPointer(result.ProviderRefundID),
		AmountMinor:          request.Amount,
		Currency:             request.Currency,
		Status:               result.Status,
		IdempotencyKey:       request.IdempotencyKey,
		Reason:               stringPointer(request.Reason),
		CreatedBy:            createdBy,
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}
	if err := s.repo.CreateRefund(ctx, refund); err != nil {
		return nil, nil, err
	}
	if err := s.repo.UpdateStatus(ctx, transaction, result.Status); err != nil {
		return nil, nil, err
	}
	return refund, transaction, nil
}

func setProviderIDs(transaction *Transaction, result paymentproviders.PaymentResult) {
	if result.ProviderPaymentID != "" {
		transaction.ProviderPaymentID = stringPointer(result.ProviderPaymentID)
	}
	if result.ProviderTransactionID != "" {
		transaction.ProviderTransactionID = stringPointer(result.ProviderTransactionID)
	}
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
