package suppliers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Service handles supplier business logic
type Service struct {
	repo *Repository
	db   *sqlx.DB
}

// NewService creates a new supplier service
func NewService(repo *Repository, db *sqlx.DB) *Service {
	return &Service{repo: repo, db: db}
}

// CreateSupplier creates a new supplier
func (s *Service) CreateSupplier(ctx context.Context, req *SupplierRequest) (*Supplier, error) {
	code := strings.TrimSpace(req.Code)
	if code == "" {
		var generatedCode string
		for i := 0; i < 10; i++ {
			candidate := generateSupplierCode()
			_, err := s.repo.GetByCode(ctx, candidate)
			if err == ErrSupplierNotFound {
				generatedCode = candidate
				break
			}
			if err != nil && err != ErrSupplierNotFound {
				return nil, fmt.Errorf("failed to generate unique supplier code: %w", err)
			}
		}
		if generatedCode == "" {
			return nil, fmt.Errorf("failed to generate unique supplier code")
		}
		code = generatedCode
		req.Code = generatedCode
	} else {
		_, err := s.repo.GetByCode(ctx, code)
		if err == nil {
			return nil, ErrSupplierCodeExists
		}
	}

	// Create supplier
	supplier := NewSupplier(code, req.Name)
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
func (s *Service) GetSupplier(ctx context.Context, id uuid.UUID) (*Supplier, error) {
	return s.repo.GetByID(ctx, id)
}

// ListSuppliers retrieves suppliers with pagination and filters
func (s *Service) ListSuppliers(ctx context.Context, page, perPage int, search string, isActive *bool) ([]Supplier, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	return s.repo.List(ctx, page, perPage, search, isActive)
}

// UpdateSupplier updates a supplier
func (s *Service) UpdateSupplier(ctx context.Context, id uuid.UUID, req *SupplierRequest) (*Supplier, error) {
	supplier, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if new code already exists (if changed)
	if req.Code != supplier.Code {
		_, err := s.repo.GetByCode(ctx, req.Code)
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
func generateSupplierCode() string {
	return "SUP-" + strings.ToUpper(uuid.NewString()[:8])
}

func (s *Service) DeleteSupplier(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// AddPayment adds a payment to supplier
func (s *Service) AddPayment(ctx context.Context, supplierID uuid.UUID, req *PaymentRequest) (*PaymentResponse, error) {
	_, err := s.repo.GetByID(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	// Validate payment amount
	if req.Amount <= 0 {
		return nil, ErrPaymentAmountInvalid
	}
	_, _, _, currentBalance, err := s.repo.GetSupplierLedger(ctx, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to read supplier balance before payment: %w", err)
	}
	if req.Amount > currentBalance {
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
	if err := s.repo.UpdateBalance(ctx, supplierID, -req.Amount); err != nil {
		return nil, fmt.Errorf("failed to update supplier balance: %w", err)
	}

	return payment, nil
}

// GetSupplierLedger retrieves supplier ledger
func (s *Service) GetSupplierLedger(ctx context.Context, supplierID uuid.UUID) (*SupplierLedgerResponse, error) {
	supplier, err := s.repo.GetByID(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	entries, totalPurchases, totalPayments, currentBalance, err := s.repo.GetSupplierLedger(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	var supplierPayments, supplierReturnCredits float64
	for _, entry := range entries {
		if entry.Type != "credit" {
			continue
		}
		upperDescription := strings.ToUpper(entry.Description)
		if strings.Contains(upperDescription, "SUPPLIER_RETURN") || strings.Contains(upperDescription, "SUPPLIER RETURN") || strings.Contains(entry.Description, "مرتجع مورد") {
			supplierReturnCredits += entry.Amount
		} else if !strings.HasPrefix(upperDescription, "PAYMENT FOR PURCHASE") {
			supplierPayments += entry.Amount
		}
	}

	return &SupplierLedgerResponse{
		SupplierID:            supplierID,
		SupplierName:          supplier.Name,
		TotalPurchases:        totalPurchases,
		TotalPayments:         totalPayments,
		SupplierPayments:      supplierPayments,
		SupplierReturnCredits: supplierReturnCredits,
		CurrentBalance:        currentBalance,
		Entries:               entries,
	}, nil
}

// AddDebt adds a debt entry to supplier (when we make a purchase on credit)
func (s *Service) AddDebt(ctx context.Context, supplierID uuid.UUID, amount float64, referenceID uuid.UUID, description string) error {
	supplier, err := s.repo.GetByID(ctx, supplierID)
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
	if err := s.repo.UpdateBalance(ctx, supplierID, amount); err != nil {
		return fmt.Errorf("failed to update supplier balance: %w", err)
	}

	return nil
}

// GetSupplierDebtSummary retrieves debt summary for a supplier
func (s *Service) GetSupplierDebtSummary(ctx context.Context, supplierID uuid.UUID) (*DebtSummary, error) {
	supplier, err := s.repo.GetByID(ctx, supplierID)
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
	if dbutil.IsSQLite(s.repo.db) {
		overdueQuery = `SELECT COALESCE(SUM(amount), 0) FROM supplier_ledger WHERE supplier_id = $1 AND (type = 'debit' OR transaction_type = 'PURCHASE') AND created_at < datetime('now','-30 days')`
	}
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
func (s *Service) UpdateCreditLimit(ctx context.Context, supplierID uuid.UUID, newLimit float64) error {
	supplier, err := s.repo.GetByID(ctx, supplierID)
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
func (s *Service) GetOverdueSuppliers(ctx context.Context) ([]OverdueSupplier, error) {
	query := `
		SELECT s.id, s.name, s.code, s.current_balance, s.credit_limit, s.email, s.phone,
			COALESCE(SUM(CASE WHEN sl.type = 'debit' AND sl.created_at < NOW() - INTERVAL '30 days' THEN sl.amount ELSE 0 END), 0) as overdue_amount
		FROM suppliers s
		LEFT JOIN supplier_ledger sl ON s.id = sl.supplier_id
		WHERE s.is_active = true
		AND s.current_balance > 0
		GROUP BY s.id, s.name, s.code, s.current_balance, s.credit_limit, s.email, s.phone
		HAVING COALESCE(SUM(CASE WHEN sl.type = 'debit' AND sl.created_at < NOW() - INTERVAL '30 days' THEN sl.amount ELSE 0 END), 0) > 0
		ORDER BY overdue_amount DESC
	`
	if dbutil.IsSQLite(s.repo.db) {
		query = `SELECT s.id, s.name, s.code, s.current_balance, s.credit_limit, s.email, s.phone, COALESCE(SUM(CASE WHEN (sl.type = 'debit' OR sl.transaction_type = 'PURCHASE') AND sl.created_at < datetime('now','-30 days') THEN sl.amount ELSE 0 END), 0) AS overdue_amount FROM suppliers s LEFT JOIN supplier_ledger sl ON s.id = sl.supplier_id WHERE s.is_active = 1 AND s.current_balance > 0 GROUP BY s.id, s.name, s.code, s.current_balance, s.credit_limit, s.email, s.phone HAVING COALESCE(SUM(CASE WHEN (sl.type = 'debit' OR sl.transaction_type = 'PURCHASE') AND sl.created_at < datetime('now','-30 days') THEN sl.amount ELSE 0 END), 0) > 0 ORDER BY overdue_amount DESC`
	}

	var overdueSuppliers []OverdueSupplier
	err := s.repo.db.SelectContext(ctx, &overdueSuppliers, query)
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
	if dbutil.IsSQLite(s.repo.db) {
		query = `SELECT CAST(julianday(datetime('now')) - julianday(MIN(created_at)) AS INTEGER) FROM supplier_ledger WHERE supplier_id = $1 AND (type = 'debit' OR transaction_type = 'PURCHASE') AND created_at >= datetime('now','-30 days') HAVING MIN(created_at) IS NOT NULL`
	}

	var days int
	err := s.repo.db.GetContext(ctx, &days, query, supplierID)
	if err != nil {
		return 0
	}

	return days
}

// CreateDebtEntry creates a new debt entry for a supplier
func (s *Service) CreateDebtEntry(ctx context.Context, supplierID uuid.UUID, amount float64, referenceID uuid.UUID, referenceType string, dueDate time.Time) error {
	supplier, err := s.repo.GetByID(ctx, supplierID)
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
	if err := s.repo.UpdateBalance(ctx, supplierID, amount); err != nil {
		return fmt.Errorf("failed to update supplier balance: %w", err)
	}

	return nil
}

// GetDebtEntries retrieves debt entries for a supplier
func (s *Service) GetDebtEntries(ctx context.Context, supplierID uuid.UUID) ([]DebtEntry, error) {
	// Verify supplier exists
	_, err := s.repo.GetByID(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetDebtEntries(ctx, supplierID)
}

// CreateDebtCollection creates a new debt collection action
func (s *Service) CreateDebtCollection(ctx context.Context, supplierID uuid.UUID, collectionType string, scheduledDate time.Time, notes *string) error {
	// Verify supplier exists
	_, err := s.repo.GetByID(ctx, supplierID)
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
func (s *Service) GetDebtCollections(ctx context.Context, supplierID uuid.UUID) ([]DebtCollection, error) {
	// Verify supplier exists
	_, err := s.repo.GetByID(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetDebtCollections(ctx, supplierID)
}

// GetPendingDebtCollections retrieves pending debt collection actions
func (s *Service) GetPendingDebtCollections(ctx context.Context) ([]DebtCollection, error) {
	return s.repo.GetPendingDebtCollections(ctx)
}

// ProcessDebtPayment processes a payment for specific debts
func (s *Service) ProcessDebtPayment(ctx context.Context, supplierID uuid.UUID, paymentAmount float64, method string) error {
	_, err := s.repo.GetByID(ctx, supplierID)
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
	if err := s.repo.UpdateBalance(ctx, supplierID, -paymentAmount); err != nil {
		return fmt.Errorf("failed to update supplier balance: %w", err)
	}

	return nil
}

// GetSupplierInventory retrieves inventory items from a specific supplier
func (s *Service) GetSupplierInventory(ctx context.Context, supplierID uuid.UUID) ([]SupplierInventoryItem, error) {
	// Verify supplier exists
	_, err := s.repo.GetByID(ctx, supplierID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			p.id as product_id,
			p.name as product_name,
			p.sku,
			COUNT(ii.id) as total_received,
			COUNT(CASE WHEN ii.status = 'AVAILABLE' THEN 1 END) as available,
			COUNT(CASE WHEN ii.status = 'SOLD' THEN 1 END) as sold,
			COUNT(CASE WHEN ii.status = 'RESERVED' THEN 1 END) as reserved,
			COUNT(CASE WHEN ii.status = 'DAMAGED' THEN 1 END) as damaged,
			COALESCE(AVG(ii.purchase_cost), 0) as avg_cost,
			COALESCE(AVG(ii.selling_price), 0) as avg_price,
			COALESCE(MIN(ii.purchase_date), NOW()) as first_purchase_date,
			COALESCE(MAX(ii.sold_at), NULL) as last_sale_date
		FROM inventory_items ii
		JOIN products p ON ii.product_id = p.id
		WHERE ii.supplier_id = $1
		GROUP BY p.id, p.name, p.sku
		ORDER BY p.name
	`
	if dbutil.IsSQLite(s.db) {
		query = `SELECT p.id AS product_id, p.name AS product_name, p.sku, COUNT(ii.id) AS total_received, COUNT(CASE WHEN ii.status = 'AVAILABLE' THEN 1 END) AS available, COUNT(CASE WHEN ii.status = 'SOLD' THEN 1 END) AS sold, COUNT(CASE WHEN ii.status = 'RESERVED' THEN 1 END) AS reserved, COUNT(CASE WHEN ii.status = 'DAMAGED' THEN 1 END) AS damaged, COALESCE(AVG(ii.purchase_cost), 0) AS avg_cost, COALESCE(AVG(ii.selling_price), 0) AS avg_price, MIN(COALESCE(ii.purchase_date, ii.created_at)) AS first_purchase_date, MAX(ii.sold_at) AS last_sale_date FROM inventory_items ii JOIN products p ON ii.product_id = p.id WHERE ii.supplier_id = $1 GROUP BY p.id, p.name, p.sku ORDER BY p.name`
	}

	var items []SupplierInventoryItem
	if dbutil.IsSQLite(s.db) {
		rows, queryErr := s.db.QueryxContext(ctx, query, supplierID)
		if queryErr != nil {
			return nil, fmt.Errorf("failed to get supplier inventory: %w", queryErr)
		}
		defer rows.Close()
		for rows.Next() {
			var row struct {
				ProductID, ProductName, SKU                       string
				TotalReceived, Available, Sold, Reserved, Damaged int
				AvgCost, AvgPrice                                 float64
				FirstPurchaseDate, LastSaleDate                   sql.NullString
			}
			if err := rows.StructScan(&row); err != nil {
				return nil, err
			}
			productID, err := uuid.Parse(row.ProductID)
			if err != nil {
				return nil, err
			}
			first, err := dbutil.ParseTimestamp(row.FirstPurchaseDate.String)
			if err != nil {
				return nil, err
			}
			item := SupplierInventoryItem{ProductID: productID, ProductName: row.ProductName, SKU: row.SKU, TotalReceived: row.TotalReceived, Available: row.Available, Sold: row.Sold, Reserved: row.Reserved, Damaged: row.Damaged, AvgCost: row.AvgCost, AvgPrice: row.AvgPrice, FirstPurchaseDate: first}
			if row.LastSaleDate.Valid && row.LastSaleDate.String != "" {
				parsed, e := dbutil.ParseTimestamp(row.LastSaleDate.String)
				if e == nil {
					item.LastSaleDate = &parsed
				}
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	} else {
		err = s.db.SelectContext(ctx, &items, query, supplierID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier inventory: %w", err)
	}

	return items, nil
}

// SupplierInventoryItem represents inventory item summary for a supplier
type SupplierInventoryItem struct {
	ProductID         uuid.UUID  `json:"product_id" db:"product_id"`
	ProductName       string     `json:"product_name" db:"product_name"`
	SKU               string     `json:"sku" db:"sku"`
	TotalReceived     int        `json:"total_received" db:"total_received"`
	Available         int        `json:"available" db:"available"`
	Sold              int        `json:"sold" db:"sold"`
	Reserved          int        `json:"reserved" db:"reserved"`
	Damaged           int        `json:"damaged" db:"damaged"`
	AvgCost           float64    `json:"avg_cost" db:"avg_cost"`
	AvgPrice          float64    `json:"avg_price" db:"avg_price"`
	FirstPurchaseDate time.Time  `json:"first_purchase_date" db:"first_purchase_date"`
	LastSaleDate      *time.Time `json:"last_sale_date" db:"last_sale_date"`
}
