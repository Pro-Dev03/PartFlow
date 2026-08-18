package suppliers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles supplier business logic
type Service struct {
	repo *Repository
}

// NewService creates a new supplier service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateSupplier creates a new supplier
func (s *Service) CreateSupplier(ctx context.Context, organizationID uuid.UUID, req *SupplierRequest) (*Supplier, error) {
	// Check if code already exists
	_, err := s.repo.GetByCode(ctx, req.Code, organizationID)
	if err == nil {
		return nil, ErrSupplierCodeExists
	}

	// Create supplier
	supplier := NewSupplier(organizationID, req.Code, req.Name)
	supplier.Email = req.Email
	supplier.Phone = req.Phone
	supplier.Address = req.Address
	supplier.City = req.City
	supplier.Country = req.Country
	supplier.TaxID = req.TaxID
	supplier.PaymentTerms = req.PaymentTerms
	supplier.CreditLimit = req.CreditLimit
	supplier.Notes = req.Notes
	supplier.IsActive = req.IsActive

	if err := s.repo.Create(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	return supplier, nil
}

// GetSupplier retrieves a supplier by ID
func (s *Service) GetSupplier(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Supplier, error) {
	return s.repo.GetByID(ctx, id, organizationID)
}

// ListSuppliers retrieves suppliers with pagination and filters
func (s *Service) ListSuppliers(ctx context.Context, organizationID uuid.UUID, page, perPage int, search string, isActive *bool) ([]Supplier, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	return s.repo.List(ctx, organizationID, page, perPage, search, isActive)
}

// UpdateSupplier updates a supplier
func (s *Service) UpdateSupplier(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *SupplierRequest) (*Supplier, error) {
	supplier, err := s.repo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Check if new code already exists (if changed)
	if req.Code != supplier.Code {
		_, err := s.repo.GetByCode(ctx, req.Code, organizationID)
		if err == nil {
			return nil, ErrSupplierCodeExists
		}
	}

	// Update supplier
	supplier.Code = req.Code
	supplier.Name = req.Name
	supplier.Email = req.Email
	supplier.Phone = req.Phone
	supplier.Address = req.Address
	supplier.City = req.City
	supplier.Country = req.Country
	supplier.TaxID = req.TaxID
	supplier.PaymentTerms = req.PaymentTerms
	supplier.CreditLimit = req.CreditLimit
	supplier.Notes = req.Notes
	supplier.IsActive = req.IsActive
	supplier.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}

	return supplier, nil
}

// DeleteSupplier deletes a supplier
func (s *Service) DeleteSupplier(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	return s.repo.Delete(ctx, id, organizationID)
}

// AddPayment adds a payment to supplier
func (s *Service) AddPayment(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID, req *PaymentRequest) (*PaymentResponse, error) {
	_, err := s.repo.GetByID(ctx, supplierID, organizationID)
	if err != nil {
		return nil, err
	}

	// Validate payment amount
	if req.Amount <= 0 {
		return nil, ErrPaymentAmountInvalid
	}

	// Set payment date if not provided
	paymentDate := time.Now()
	if req.PaymentDate != nil {
		paymentDate = *req.PaymentDate
	}

	// Create payment
	payment := &PaymentResponse{
		ID:          uuid.New(),
		SupplierID:  supplierID,
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

	// Update supplier balance
	if err := s.repo.UpdateBalance(ctx, supplierID, organizationID, -req.Amount); err != nil {
		return nil, fmt.Errorf("failed to update supplier balance: %w", err)
	}

	return payment, nil
}

// GetSupplierLedger retrieves supplier ledger
func (s *Service) GetSupplierLedger(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID) (*SupplierLedgerResponse, error) {
	supplier, err := s.repo.GetByID(ctx, supplierID, organizationID)
	if err != nil {
		return nil, err
	}

	entries, totalPurchases, totalPayments, currentBalance, err := s.repo.GetSupplierLedger(ctx, supplierID, organizationID)
	if err != nil {
		return nil, err
	}

	return &SupplierLedgerResponse{
		SupplierID:     supplierID,
		SupplierName:   supplier.Name,
		TotalPurchases: totalPurchases,
		TotalPayments:  totalPayments,
		CurrentBalance: currentBalance,
		Entries:        entries,
	}, nil
}

// AddDebt adds a debt entry to supplier (when we make a purchase on credit)
func (s *Service) AddDebt(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID, amount float64, referenceID uuid.UUID, description string) error {
	supplier, err := s.repo.GetByID(ctx, supplierID, organizationID)
	if err != nil {
		return err
	}

	// Check if adding debt would exceed credit limit
	if supplier.CurrentBalance+amount > supplier.CreditLimit {
		return ErrCreditLimitExceeded
	}

	// Add to ledger
	err = s.repo.AddLedgerEntry(ctx, supplierID, "debit", amount, description, referenceID)
	if err != nil {
		return fmt.Errorf("failed to add ledger entry: %w", err)
	}

	// Update supplier balance
	if err := s.repo.UpdateBalance(ctx, supplierID, organizationID, amount); err != nil {
		return fmt.Errorf("failed to update supplier balance: %w", err)
	}

	return nil
}

// GetSupplierDebtSummary retrieves debt summary for a supplier
func (s *Service) GetSupplierDebtSummary(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID) (*DebtSummary, error) {
	supplier, err := s.repo.GetByID(ctx, supplierID, organizationID)
	if err != nil {
		return nil, err
	}

	// Get overdue debt (debts older than 30 days)
	overdueQuery := `
		SELECT COALESCE(SUM(amount), 0)
		FROM supplier_ledger
		WHERE supplier_id = $1
		AND type = 'debit'
		AND created_at < NOW() - INTERVAL '30 days'
	`
	var overdueAmount float64
	err = s.repo.db.QueryRowContext(ctx, overdueQuery, supplierID).Scan(&overdueAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue debt: %w", err)
	}

	// Calculate available credit
	availableCredit := supplier.CreditLimit - supplier.CurrentBalance

	// Calculate credit utilization percentage
	creditUtilization := 0.0
	if supplier.CreditLimit > 0 {
		creditUtilization = (supplier.CurrentBalance / supplier.CreditLimit) * 100
	}

	return &DebtSummary{
		SupplierID:        supplierID,
		SupplierName:      supplier.Name,
		CurrentBalance:    supplier.CurrentBalance,
		CreditLimit:       supplier.CreditLimit,
		AvailableCredit:   availableCredit,
		CreditUtilization: creditUtilization,
		OverdueAmount:     overdueAmount,
		IsOverdue:         overdueAmount > 0,
		DaysUntilOverdue:  s.calculateDaysUntilOverdue(ctx, supplierID),
	}, nil
}

// UpdateCreditLimit updates supplier credit limit
func (s *Service) UpdateCreditLimit(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID, newLimit float64) error {
	supplier, err := s.repo.GetByID(ctx, supplierID, organizationID)
	if err != nil {
		return err
	}

	// Check if new limit is below current balance
	if newLimit < supplier.CurrentBalance {
		return ErrCreditLimitBelowBalance
	}

	supplier.CreditLimit = newLimit
	supplier.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, supplier); err != nil {
		return fmt.Errorf("failed to update credit limit: %w", err)
	}

	return nil
}

// GetOverdueSuppliers retrieves suppliers with overdue payments
func (s *Service) GetOverdueSuppliers(ctx context.Context, organizationID uuid.UUID) ([]OverdueSupplier, error) {
	query := `
		SELECT s.id, s.name, s.code, s.current_balance, s.credit_limit, s.email, s.phone,
			COALESCE(SUM(CASE WHEN sl.type = 'debit' AND sl.created_at < NOW() - INTERVAL '30 days' THEN sl.amount ELSE 0 END), 0) as overdue_amount
		FROM suppliers s
		LEFT JOIN supplier_ledger sl ON s.id = sl.supplier_id
		WHERE s.organization_id = $1
		AND s.is_active = true
		AND s.current_balance > 0
		GROUP BY s.id, s.name, s.code, s.current_balance, s.credit_limit, s.email, s.phone
		HAVING COALESCE(SUM(CASE WHEN sl.type = 'debit' AND sl.created_at < NOW() - INTERVAL '30 days' THEN sl.amount ELSE 0 END), 0) > 0
		ORDER BY overdue_amount DESC
	`

	var overdueSuppliers []OverdueSupplier
	err := s.repo.db.SelectContext(ctx, &overdueSuppliers, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue suppliers: %w", err)
	}

	return overdueSuppliers, nil
}

// calculateDaysUntilOverdue calculates days until debt becomes overdue
func (s *Service) calculateDaysUntilOverdue(ctx context.Context, supplierID uuid.UUID) int {
	query := `
		SELECT EXTRACT(DAY FROM (MIN(created_at) + INTERVAL '30 days' - NOW())) as days
		FROM supplier_ledger
		WHERE supplier_id = $1
		AND type = 'debit'
		AND created_at >= NOW() - INTERVAL '30 days'
		HAVING MIN(created_at) IS NOT NULL
	`

	var days int
	err := s.repo.db.GetContext(ctx, &days, query, supplierID)
	if err != nil {
		return 0
	}

	return days
}

// CreateDebtEntry creates a new debt entry for a supplier
func (s *Service) CreateDebtEntry(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID, amount float64, referenceID uuid.UUID, referenceType string, dueDate time.Time) error {
	supplier, err := s.repo.GetByID(ctx, supplierID, organizationID)
	if err != nil {
		return err
	}

	// Check if adding debt would exceed credit limit
	if supplier.CurrentBalance+amount > supplier.CreditLimit {
		return ErrCreditLimitExceeded
	}

	debt := &DebtEntry{
		ID:            uuid.New(),
		SupplierID:    supplierID,
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
	if err := s.repo.AddLedgerEntry(ctx, supplierID, "debit", amount, fmt.Sprintf("%s - %s", referenceType, referenceID.String()), referenceID); err != nil {
		return fmt.Errorf("failed to add ledger entry: %w", err)
	}

	// Update supplier balance
	if err := s.repo.UpdateBalance(ctx, supplierID, organizationID, amount); err != nil {
		return fmt.Errorf("failed to update supplier balance: %w", err)
	}

	return nil
}

// GetDebtEntries retrieves debt entries for a supplier
func (s *Service) GetDebtEntries(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID) ([]DebtEntry, error) {
	// Verify supplier exists
	_, err := s.repo.GetByID(ctx, supplierID, organizationID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetDebtEntries(ctx, supplierID)
}

// CreateDebtCollection creates a new debt collection action
func (s *Service) CreateDebtCollection(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID, collectionType string, scheduledDate time.Time, notes *string) error {
	// Verify supplier exists
	_, err := s.repo.GetByID(ctx, supplierID, organizationID)
	if err != nil {
		return err
	}

	collection := &DebtCollection{
		ID:            uuid.New(),
		SupplierID:    supplierID,
		Type:          collectionType,
		Status:        "pending",
		Notes:         notes,
		ScheduledDate: scheduledDate,
		CreatedAt:     time.Now(),
	}

	return s.repo.CreateDebtCollection(ctx, collection)
}

// GetDebtCollections retrieves debt collection actions for a supplier
func (s *Service) GetDebtCollections(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID) ([]DebtCollection, error) {
	// Verify supplier exists
	_, err := s.repo.GetByID(ctx, supplierID, organizationID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetDebtCollections(ctx, supplierID)
}

// GetPendingDebtCollections retrieves pending debt collection actions for the organization
func (s *Service) GetPendingDebtCollections(ctx context.Context, organizationID uuid.UUID) ([]DebtCollection, error) {
	return s.repo.GetPendingDebtCollections(ctx, organizationID)
}

// ProcessDebtPayment processes a payment for specific debts
func (s *Service) ProcessDebtPayment(ctx context.Context, supplierID uuid.UUID, organizationID uuid.UUID, paymentAmount float64, method string) error {
	_, err := s.repo.GetByID(ctx, supplierID, organizationID)
	if err != nil {
		return err
	}

	// Get unpaid debts
	debts, err := s.repo.GetDebtEntries(ctx, supplierID)
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
		SupplierID:  supplierID,
		Amount:      paymentAmount,
		PaymentDate: time.Now(),
		Method:      method,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.AddPayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to add payment: %w", err)
	}

	// Update supplier balance
	if err := s.repo.UpdateBalance(ctx, supplierID, organizationID, -paymentAmount); err != nil {
		return fmt.Errorf("failed to update supplier balance: %w", err)
	}

	return nil
}