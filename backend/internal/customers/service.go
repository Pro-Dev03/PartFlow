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

// AddDebt adds a debt entry to customer (when they make a purchase on credit)
func (s *Service) AddDebt(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID, amount float64, referenceID uuid.UUID, description string) error {
	customer, err := s.repo.GetByID(ctx, customerID, organizationID)
	if err != nil {
		return err
	}

	// Check if adding debt would exceed credit limit
	if customer.CurrentBalance+amount > customer.CreditLimit {
		return ErrCreditLimitExceeded
	}

	// Add to ledger
	err = s.repo.AddLedgerEntry(ctx, customerID, "debit", amount, description, referenceID)
	if err != nil {
		return fmt.Errorf("failed to add ledger entry: %w", err)
	}

	// Update customer balance
	if err := s.repo.UpdateBalance(ctx, customerID, organizationID, amount); err != nil {
		return fmt.Errorf("failed to update customer balance: %w", err)
	}

	return nil
}

// GetCustomerDebtSummary retrieves debt summary for a customer
func (s *Service) GetCustomerDebtSummary(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID) (*DebtSummary, error) {
	customer, err := s.repo.GetByID(ctx, customerID, organizationID)
	if err != nil {
		return nil, err
	}

	// Get overdue debt (debts older than 30 days)
	overdueQuery := `
		SELECT COALESCE(SUM(amount), 0)
		FROM customer_ledger
		WHERE customer_id = $1
		AND type = 'debit'
		AND created_at < NOW() - INTERVAL '30 days'
	`
	var overdueAmount float64
	err = s.repo.db.GetContext(ctx, &overdueAmount, overdueQuery, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue debt: %w", err)
	}

	// Calculate available credit
	availableCredit := customer.CreditLimit - customer.CurrentBalance

	// Calculate credit utilization percentage
	creditUtilization := 0.0
	if customer.CreditLimit > 0 {
		creditUtilization = (customer.CurrentBalance / customer.CreditLimit) * 100
	}

	return &DebtSummary{
		CustomerID:          customerID,
		CustomerName:        customer.Name,
		CurrentBalance:      customer.CurrentBalance,
		CreditLimit:         customer.CreditLimit,
		AvailableCredit:     availableCredit,
		CreditUtilization:   creditUtilization,
		OverdueAmount:       overdueAmount,
		IsOverdue:           overdueAmount > 0,
		DaysUntilOverdue:    s.calculateDaysUntilOverdue(ctx, customerID),
	}, nil
}

// UpdateCreditLimit updates customer credit limit
func (s *Service) UpdateCreditLimit(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID, newLimit float64) error {
	customer, err := s.repo.GetByID(ctx, customerID, organizationID)
	if err != nil {
		return err
	}

	// Check if new limit is below current balance
	if newLimit < customer.CurrentBalance {
		return ErrCreditLimitBelowBalance
	}

	customer.CreditLimit = newLimit
	customer.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update credit limit: %w", err)
	}

	return nil
}

// GetOverdueCustomers retrieves customers with overdue payments
func (s *Service) GetOverdueCustomers(ctx context.Context, organizationID uuid.UUID) ([]OverdueCustomer, error) {
	query := `
		SELECT c.id, c.name, c.code, c.current_balance, c.credit_limit, c.email, c.phone,
			COALESCE(SUM(CASE WHEN cl.type = 'debit' AND cl.created_at < NOW() - INTERVAL '30 days' THEN cl.amount ELSE 0 END), 0) as overdue_amount
		FROM customers c
		LEFT JOIN customer_ledger cl ON c.id = cl.customer_id
		WHERE c.organization_id = $1
		AND c.is_active = true
		AND c.current_balance > 0
		GROUP BY c.id, c.name, c.code, c.current_balance, c.credit_limit, c.email, c.phone
		HAVING COALESCE(SUM(CASE WHEN cl.type = 'debit' AND cl.created_at < NOW() - INTERVAL '30 days' THEN cl.amount ELSE 0 END), 0) > 0
		ORDER BY overdue_amount DESC
	`

	var overdueCustomers []OverdueCustomer
	err := s.repo.db.SelectContext(ctx, &overdueCustomers, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue customers: %w", err)
	}

	return overdueCustomers, nil
}

// calculateDaysUntilOverdue calculates days until debt becomes overdue
func (s *Service) calculateDaysUntilOverdue(ctx context.Context, customerID uuid.UUID) int {
	query := `
		SELECT EXTRACT(DAY FROM (MIN(created_at) + INTERVAL '30 days' - NOW())) as days
		FROM customer_ledger
		WHERE customer_id = $1
		AND type = 'debit'
		AND created_at >= NOW() - INTERVAL '30 days'
		HAVING MIN(created_at) IS NOT NULL
	`

	var days int
	err := s.repo.db.GetContext(ctx, &days, query, customerID)
	if err != nil {
		return 0
	}

	return days
}

// CreateDebtEntry creates a new debt entry for a customer
func (s *Service) CreateDebtEntry(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID, amount float64, referenceID uuid.UUID, referenceType string, dueDate time.Time) error {
	customer, err := s.repo.GetByID(ctx, customerID, organizationID)
	if err != nil {
		return err
	}

	// Check if adding debt would exceed credit limit
	if customer.CurrentBalance+amount > customer.CreditLimit {
		return ErrCreditLimitExceeded
	}

	debt := &DebtEntry{
		ID:            uuid.New(),
		CustomerID:    customerID,
		Amount:        amount,
		ReferenceID:   referenceID,
		ReferenceType: referenceType,
		DueDate:       dueDate,
		IsPaid:        false,
		PaidAmount:    0,
		CreatedAt:     time.Now(),
	}

	if err := s.repo.CreateDebtEntry(ctx, debt); err != nil {
		return fmt.Errorf("failed to create debt entry: %w", err)
	}

	// Add to ledger
	if err := s.repo.AddLedgerEntry(ctx, customerID, "debit", amount, fmt.Sprintf("%s - %s", referenceType, referenceID.String()), referenceID); err != nil {
		return fmt.Errorf("failed to add ledger entry: %w", err)
	}

	// Update customer balance
	if err := s.repo.UpdateBalance(ctx, customerID, organizationID, amount); err != nil {
		return fmt.Errorf("failed to update customer balance: %w", err)
	}

	return nil
}

// GetDebtEntries retrieves debt entries for a customer
func (s *Service) GetDebtEntries(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID) ([]DebtEntry, error) {
	// Verify customer exists
	_, err := s.repo.GetByID(ctx, customerID, organizationID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetDebtEntries(ctx, customerID)
}

// CreateDebtCollection creates a new debt collection action
func (s *Service) CreateDebtCollection(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID, collectionType string, scheduledDate time.Time, notes *string) error {
	// Verify customer exists
	_, err := s.repo.GetByID(ctx, customerID, organizationID)
	if err != nil {
		return err
	}

	collection := &DebtCollection{
		ID:            uuid.New(),
		CustomerID:    customerID,
		Type:          collectionType,
		Status:        "pending",
		Notes:         notes,
		ScheduledDate: scheduledDate,
		CreatedAt:     time.Now(),
	}

	return s.repo.CreateDebtCollection(ctx, collection)
}

// GetDebtCollections retrieves debt collection actions for a customer
func (s *Service) GetDebtCollections(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID) ([]DebtCollection, error) {
	// Verify customer exists
	_, err := s.repo.GetByID(ctx, customerID, organizationID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetDebtCollections(ctx, customerID)
}

// GetPendingDebtCollections retrieves pending debt collection actions for the organization
func (s *Service) GetPendingDebtCollections(ctx context.Context, organizationID uuid.UUID) ([]DebtCollection, error) {
	return s.repo.GetPendingDebtCollections(ctx, organizationID)
}

// ProcessDebtPayment processes a payment for specific debts
func (s *Service) ProcessDebtPayment(ctx context.Context, customerID uuid.UUID, organizationID uuid.UUID, paymentAmount float64, method string) error {
	_, err := s.repo.GetByID(ctx, customerID, organizationID)
	if err != nil {
		return err
	}

	// Get unpaid debts
	debts, err := s.repo.GetDebtEntries(ctx, customerID)
	if err != nil {
		return err
	}

	remainingAmount := paymentAmount
	for _, debt := range debts {
		if debt.IsPaid || remainingAmount <= 0 {
			continue
		}

		amountToPay := debt.Amount - debt.PaidAmount
		if amountToPay > remainingAmount {
			amountToPay = remainingAmount
		}

		// Update debt payment
		if err := s.repo.UpdateDebtPayment(ctx, debt.ID, amountToPay); err != nil {
			return fmt.Errorf("failed to update debt payment: %w", err)
		}

		remainingAmount -= amountToPay
	}

	// Add payment record
	payment := &PaymentResponse{
		ID:          uuid.New(),
		CustomerID:  customerID,
		Amount:      paymentAmount,
		PaymentDate: time.Now(),
		Method:      method,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.AddPayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to add payment: %w", err)
	}

	// Update customer balance
	if err := s.repo.UpdateBalance(ctx, customerID, organizationID, -paymentAmount); err != nil {
		return fmt.Errorf("failed to update customer balance: %w", err)
	}

	return nil
}
