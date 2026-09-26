package payments

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/partflow/smart-store/internal/paymentproviders"
	"github.com/partflow/smart-store/internal/paymenttransactions"
)

// Service handles payment business logic
type Service struct {
	repo *Repository
}

// NewService creates a new payment service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreatePayment creates a new payment
func (s *Service) CreatePayment(ctx context.Context, userID uuid.UUID, req *CreatePaymentRequest) (*PaymentResponse, error) {
	// Validate payment type
	if req.Type != "customer" && req.Type != "supplier" && req.Type != "expense" {
		return nil, ErrInvalidPaymentType
	}

	// Validate amount
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Set payment date if not provided
	paymentDate := time.Now()
	if req.PaymentDate != nil {
		paymentDate = *req.PaymentDate
	}

	// Create payment
	payment := NewPayment(req.Type, req.ReferenceID, req.Amount, req.Method, userID)
	payment.PaymentDate = paymentDate
	payment.Reference = req.Reference
	payment.Notes = req.Notes

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Get reference name
	referenceName, _ := s.repo.GetReferenceName(ctx, req.Type, req.ReferenceID)

	return s.toPaymentResponse(payment, referenceName), nil
}

// GetPayment retrieves a payment by ID
func (s *Service) GetPayment(ctx context.Context, id uuid.UUID) (*PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get reference name
	referenceName, _ := s.repo.GetReferenceName(ctx, payment.Type, payment.ReferenceID)

	return s.toPaymentResponse(payment, referenceName), nil
}

// ListPayments retrieves payments with pagination and filters
func (s *Service) ListPayments(ctx context.Context, page, perPage int, filters map[string]interface{}) ([]PaymentResponse, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	payments, total, err := s.repo.List(ctx, page, perPage, filters)
	if err != nil {
		return nil, 0, err
	}

	// Convert to response with reference names
	var responses []PaymentResponse
	for _, payment := range payments {
		referenceName, _ := s.repo.GetReferenceName(ctx, payment.Type, payment.ReferenceID)
		responses = append(responses, *s.toPaymentResponse(&payment, referenceName))
	}

	return responses, total, nil
}

// UpdatePayment updates a payment
func (s *Service) UpdatePayment(ctx context.Context, id uuid.UUID, req *UpdatePaymentRequest) (*PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if payment can be updated
	if payment.Status == "completed" {
		return nil, ErrPaymentAlreadyProcessed
	}

	// Update fields
	if req.Status != "" {
		payment.Status = req.Status
	}
	if req.Method != "" {
		payment.Method = req.Method
	}
	if req.Reference != nil {
		payment.Reference = req.Reference
	}
	if req.Notes != nil {
		payment.Notes = req.Notes
	}
	if req.PaymentDate != nil {
		payment.PaymentDate = *req.PaymentDate
	}

	payment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, err
	}

	// Get reference name
	referenceName, _ := s.repo.GetReferenceName(ctx, payment.Type, payment.ReferenceID)

	return s.toPaymentResponse(payment, referenceName), nil
}

// DeletePayment reverses external settlement first, then runs the atomic local
// reversal and hard delete.
func (s *Service) DeletePayment(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.PrepareDelete(ctx, id, userID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// PrepareDelete reconciles any external payment settlement before a caller
// starts an atomic transaction that deletes this payment together with a
// parent sale or purchase. Repeating it is safe because provider refunds use
// a stable idempotency key and the remaining refundable amount is recalculated.
func (s *Service) PrepareDelete(ctx context.Context, id, userID uuid.UUID) error {
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("check payment before deletion: %w", err)
	}
	if !exists {
		return ErrPaymentNotFound
	}
	if err := s.reverseProviderEffects(ctx, id, userID); err != nil {
		return fmt.Errorf("%w: %v", ErrPaymentRequiresProviderRefund, err)
	}
	return nil
}

func (s *Service) reverseProviderEffects(ctx context.Context, paymentID, userID uuid.UUID) error {
	links, err := s.repo.linkedProviderPayments(ctx, paymentID)
	if err != nil || len(links) == 0 {
		return err
	}
	service := paymenttransactions.NewConfiguredService(s.repo.db)
	var createdBy *uuid.UUID
	if userID != uuid.Nil {
		createdBy = &userID
	}
	for _, link := range links {
		switch link.Status {
		case paymentproviders.StatusPending, paymentproviders.StatusProcessing:
			if _, err := service.CancelPayment(ctx, link.ID); err != nil {
				return fmt.Errorf("cancel provider payment %s: %w", link.ID, err)
			}
		case paymentproviders.StatusPaid, paymentproviders.StatusPartiallyRefunded:
			var refunded int64
			if err := s.repo.db.GetContext(ctx, &refunded, `SELECT COALESCE(SUM(amount_minor),0) FROM payment_refunds WHERE payment_transaction_id = $1 AND status IN ('refunded','partially_refunded')`, link.ID); err != nil {
				return fmt.Errorf("read provider refunds for %s: %w", link.ID, err)
			}
			remaining := link.AmountMinor - refunded
			if remaining <= 0 {
				continue
			}
			request := paymentproviders.RefundRequest{
				PaymentID: link.ID, Amount: remaining, Currency: link.Currency,
				Reason:         "Payment record deletion requested",
				IdempotencyKey: "payment-delete-" + paymentID.String() + "-" + link.ID.String(),
			}
			if _, _, err := service.RefundPayment(ctx, link.ID, request, createdBy); err != nil {
				return fmt.Errorf("refund provider payment %s: %w", link.ID, err)
			}
		case paymentproviders.StatusRefunded, paymentproviders.StatusFailed, paymentproviders.StatusCancelled:
			// No outstanding provider balance remains.
		default:
			return fmt.Errorf("provider transaction %s has unsupported status %q", link.ID, link.Status)
		}
	}
	return nil
}

// CompletePayment marks a payment as completed
func (s *Service) CompletePayment(ctx context.Context, id uuid.UUID) (*PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if payment.Status == "completed" {
		return nil, ErrPaymentAlreadyProcessed
	}

	payment.Status = "completed"
	payment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, err
	}

	// Get reference name
	referenceName, _ := s.repo.GetReferenceName(ctx, payment.Type, payment.ReferenceID)

	return s.toPaymentResponse(payment, referenceName), nil
}

// CancelPayment cancels a payment
func (s *Service) CancelPayment(ctx context.Context, id uuid.UUID) (*PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if payment.Status == "completed" {
		return nil, ErrPaymentCannotBeCancelled
	}

	if payment.Status == "cancelled" {
		return nil, ErrPaymentAlreadyProcessed
	}

	payment.Status = "cancelled"
	payment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, err
	}

	// Get reference name
	referenceName, _ := s.repo.GetReferenceName(ctx, payment.Type, payment.ReferenceID)

	return s.toPaymentResponse(payment, referenceName), nil
}

// GetPaymentSummary retrieves payment summary statistics
func (s *Service) GetPaymentSummary(ctx context.Context) (*PaymentSummary, error) {
	return s.repo.GetPaymentSummary(ctx)
}

// toPaymentResponse converts a Payment to PaymentResponse
func (s *Service) toPaymentResponse(payment *Payment, referenceName string) *PaymentResponse {
	var refName *string
	if referenceName != "" {
		refName = &referenceName
	}

	return &PaymentResponse{
		ID:            payment.ID,
		Type:          payment.Type,
		ReferenceID:   payment.ReferenceID,
		ReferenceName: refName,
		Amount:        payment.Amount,
		PaymentDate:   payment.PaymentDate,
		Method:        payment.Method,
		Reference:     payment.Reference,
		Notes:         payment.Notes,
		Status:        payment.Status,
		CreatedBy:     payment.CreatedBy,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}
}
