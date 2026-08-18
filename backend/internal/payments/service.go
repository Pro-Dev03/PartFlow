package payments

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
func (s *Service) CreatePayment(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req *CreatePaymentRequest) (*PaymentResponse, error) {
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
	payment := NewPayment(organizationID, req.Type, req.ReferenceID, req.Amount, req.Method, userID)
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
func (s *Service) GetPayment(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Get reference name
	referenceName, _ := s.repo.GetReferenceName(ctx, payment.Type, payment.ReferenceID)

	return s.toPaymentResponse(payment, referenceName), nil
}

// ListPayments retrieves payments with pagination and filters
func (s *Service) ListPayments(ctx context.Context, organizationID uuid.UUID, page, perPage int, filters map[string]interface{}) ([]PaymentResponse, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	payments, total, err := s.repo.List(ctx, organizationID, page, perPage, filters)
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
func (s *Service) UpdatePayment(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *UpdatePaymentRequest) (*PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id, organizationID)
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

// DeletePayment deletes a payment
func (s *Service) DeletePayment(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	payment, err := s.repo.GetByID(ctx, id, organizationID)
	if err != nil {
		return err
	}

	// Check if payment can be deleted
	if payment.Status == "completed" {
		return ErrPaymentCannotBeCancelled
	}

	return s.repo.Delete(ctx, id, organizationID)
}

// CompletePayment marks a payment as completed
func (s *Service) CompletePayment(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id, organizationID)
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
func (s *Service) CancelPayment(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*PaymentResponse, error) {
	payment, err := s.repo.GetByID(ctx, id, organizationID)
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
func (s *Service) GetPaymentSummary(ctx context.Context, organizationID uuid.UUID) (*PaymentSummary, error) {
	return s.repo.GetPaymentSummary(ctx, organizationID)
}

// toPaymentResponse converts a Payment to PaymentResponse
func (s *Service) toPaymentResponse(payment *Payment, referenceName string) *PaymentResponse {
	var refName *string
	if referenceName != "" {
		refName = &referenceName
	}

	return &PaymentResponse{
		ID:             payment.ID,
		OrganizationID: payment.OrganizationID,
		Type:           payment.Type,
		ReferenceID:    payment.ReferenceID,
		ReferenceName:  refName,
		Amount:         payment.Amount,
		PaymentDate:    payment.PaymentDate,
		Method:         payment.Method,
		Reference:      payment.Reference,
		Notes:          payment.Notes,
		Status:         payment.Status,
		CreatedBy:      payment.CreatedBy,
		CreatedAt:      payment.CreatedAt,
		UpdatedAt:      payment.UpdatedAt,
	}
}