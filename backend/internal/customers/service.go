package customers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles customer business logic
type Service struct {
	repo *Repository
}

// NewService creates a new customer service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateCustomer creates a new customer
func (s *Service) CreateCustomer(ctx context.Context, organizationID uuid.UUID, req *CustomerRequest) (*Customer, error) {
	// Check if code already exists
	_, err := s.repo.GetByCode(ctx, req.Code, organizationID)
	if err == nil {
		return nil, ErrCustomerCodeExists
	}

	// Create customer
	customer := NewCustomer(organizationID, req.Code, req.Name)
	customer.Email = req.Email
	customer.Phone = req.Phone
	customer.Address = req.Address
	customer.City = req.City
	customer.Country = req.Country
	customer.TaxID = req.TaxID
	customer.CreditLimit = req.CreditLimit
	customer.Notes = req.Notes
	customer.IsActive = req.IsActive

	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *Service) GetCustomer(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Customer, error) {
	return s.repo.GetByID(ctx, id, organizationID)
}

// ListCustomers retrieves customers with pagination and filters
func (s *Service) ListCustomers(ctx context.Context, organizationID uuid.UUID, page, perPage int, search string, isActive *bool) ([]Customer, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	return s.repo.List(ctx, organizationID, page, perPage, search, isActive)
}

// UpdateCustomer updates a customer
func (s *Service) UpdateCustomer(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *CustomerRequest) (*Customer, error) {
	customer, err := s.repo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Check if new code already exists (if changed)
	if req.Code != customer.Code {
		_, err := s.repo.GetByCode(ctx, req.Code, organizationID)
		if err == nil {
			return nil, ErrCustomerCodeExists
		}
	}

	// Update customer
	customer.Code = req.Code
	customer.Name = req.Name
	customer.Email = req.Email
	customer.Phone = req.Phone
	customer.Address = req.Address
	customer.City = req.City
	customer.Country = req.Country
	customer.TaxID = req.TaxID
	customer.CreditLimit = req.CreditLimit
	customer.Notes = req.Notes
	customer.IsActive = req.IsActive
	customer.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}

// DeleteCustomer deletes a customer
func (s *Service) DeleteCustomer(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	return s.repo.Delete(ctx, id, organizationID)
}

// AddPayment adds a payment to customer
func (s *Service) AddPayment(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID, req *PaymentRequest) (*PaymentResponse, error) {
	customer, err := s.repo.GetByID(ctx, customerID, organizationID)
	if err != nil {
		return nil, err
	}

	// Validate payment amount
	if req.Amount <= 0 {
		return nil, ErrPaymentAmountInvalid
	}

	// Validate payment doesn't exceed balance
	if req.Amount > customer.CurrentBalance {
		return nil, ErrPaymentExceedsBalance
	}

	// Set payment date if not provided
	paymentDate := time.Now()
	if req.PaymentDate != nil {
		paymentDate = *req.PaymentDate
	}

	// Create payment
	payment := &PaymentResponse{
		ID:          uuid.New(),
		CustomerID:  customerID,
		Amount:      req.Amount,
		PaymentDate: paymentDate,
		Method:      req.Method,
		Reference:   req.Reference,
		Notes:       req.Notes,
		CreatedAt:   time.Now(),
	}

	// Add payment
	if err := s.repo.AddPayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to add payment: %w", err)
	}

	// Update customer balance
	if err := s.repo.UpdateBalance(ctx, customerID, organizationID, -req.Amount); err != nil {
		return nil, fmt.Errorf("failed to update customer balance: %w", err)
	}

	return payment, nil
}

// GetCustomerLedger retrieves customer ledger
func (s *Service) GetCustomerLedger(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID) (*CustomerLedgerResponse, error) {
	customer, err := s.repo.GetByID(ctx, customerID, organizationID)
	if err != nil {
		return nil, err
	}

	entries, totalPurchases, totalPayments, currentBalance, err := s.repo.GetCustomerLedger(ctx, customerID, organizationID)
	if err != nil {
		return nil, err
	}

	return &CustomerLedgerResponse{
		CustomerID:     customerID,
		CustomerName:   customer.Name,
		TotalPurchases: totalPurchases,
		TotalPayments:  totalPayments,
		CurrentBalance: currentBalance,
		Entries:        entries,
	}, nil
}
