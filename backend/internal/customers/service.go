package customers

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/partflow/smart-store/internal/dashboard"
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
func (s *Service) CreateCustomer(ctx context.Context, req *CustomerRequest) (*Customer, error) {
	code := ""
	if req.Code != nil && *req.Code != "" {
		code = *req.Code
	} else {
		code = uuid.New().String()[:8]
	}

	// Check if code already exists
	_, err := s.repo.GetByCode(ctx, code)
	if err == nil {
		return nil, ErrCustomerCodeExists
	}

	// Create customer
	customer := NewCustomer(code, req.Name)
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

	// Invalidate dashboard cache since customers data changed
	dashboard.InvalidateDashboardCacheWithReason("customer_created")

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *Service) GetCustomer(ctx context.Context, id uuid.UUID) (*Customer, error) {
	return s.repo.GetByID(ctx, id)
}

// ListCustomers retrieves customers with pagination and filters
func (s *Service) ListCustomers(ctx context.Context, page, perPage int, search string, isActive *bool) ([]Customer, int, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	return s.repo.List(ctx, page, perPage, search, isActive)
}

// UpdateCustomer updates a customer
func (s *Service) UpdateCustomer(ctx context.Context, id uuid.UUID, req *UpdateCustomerRequest) (*Customer, error) {
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	customer.Name = req.Name
	if req.Email != nil {
		customer.Email = req.Email
	}
	if req.Phone != nil {
		customer.Phone = req.Phone
	}
	if req.Address != nil {
		customer.Address = req.Address
	}
	if req.City != nil {
		customer.City = req.City
	}
	if req.Country != nil {
		customer.Country = req.Country
	}
	if req.TaxID != nil {
		customer.TaxID = req.TaxID
	}
	if req.CreditLimit != nil {
		customer.CreditLimit = *req.CreditLimit
	}
	if req.Notes != nil {
		customer.Notes = req.Notes
	}
	if req.IsActive != nil {
		customer.IsActive = *req.IsActive
	}
	customer.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// Invalidate dashboard cache since customers data changed
	dashboard.InvalidateDashboardCacheWithReason("customer_updated")

	return customer, nil
}

// DeleteCustomer deletes a customer with safety checks
func (s *Service) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	// Safety check: Get customer first
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// CRITICAL: Check if customer has outstanding debt
	if customer.CurrentBalance > 0 {
		return ErrCustomerHasOutstandingDebt
	}

	// Check for active sales/transactions
	hasActiveTransactions, err := s.repo.HasActiveTransactions(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check for active transactions: %w", err)
	}

	if hasActiveTransactions {
		return ErrCustomerHasActiveTransactions
	}

	// Check for active warranties
	hasActiveWarranties, err := s.repo.HasActiveWarranties(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check for active warranties: %w", err)
	}

	if hasActiveWarranties {
		return ErrCustomerHasActiveWarranties
	}

	// Safe to delete
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate dashboard cache since customers data changed
	dashboard.InvalidateDashboardCacheWithReason("customer_deleted")

	return nil
}

// AddPayment adds a payment to customer
func (s *Service) AddPayment(ctx context.Context, customerID uuid.UUID, req *PaymentRequest) (*PaymentResponse, error) {
	customer, err := s.repo.GetByID(ctx, customerID)
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
	if err := s.repo.UpdateBalance(ctx, customerID, -req.Amount); err != nil {
		return nil, fmt.Errorf("failed to update customer balance: %w", err)
	}

	return payment, nil
}

// GetCustomerLedger retrieves customer ledger
func (s *Service) GetCustomerLedger(ctx context.Context, customerID uuid.UUID) (*CustomerLedgerResponse, error) {
	customer, err := s.repo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// Get ledger entries
	entries, totalPurchases, totalPayments, currentBalance, err := s.repo.GetCustomerLedger(ctx, customerID)
	if err != nil {
		// If ledger is empty, return empty response with customer info
		return &CustomerLedgerResponse{
			CustomerID:     customerID,
			CustomerName:   customer.Name,
			TotalPurchases: 0,
			TotalPayments:  0,
			CurrentBalance: customer.CurrentBalance,
			Entries:        []LedgerEntry{},
		}, nil
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
func (s *Service) AddDebt(ctx context.Context, customerID uuid.UUID, amount float64, referenceID uuid.UUID, description string) error {
	customer, err := s.repo.GetByID(ctx, customerID)
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
	if err := s.repo.UpdateBalance(ctx, customerID, amount); err != nil {
		return fmt.Errorf("failed to update customer balance: %w", err)
	}

	return nil
}

// GetCustomerDebtSummary retrieves debt summary for a customer
func (s *Service) GetCustomerDebtSummary(ctx context.Context, customerID uuid.UUID) (*DebtSummary, error) {
	customer, err := s.repo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
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
		OverdueAmount:       customer.CurrentBalance, // Use current balance as overdue for now
		IsOverdue:           customer.CurrentBalance > 0,
		DaysUntilOverdue:    0,
	}, nil
}

// UpdateCreditLimit updates customer credit limit
func (s *Service) UpdateCreditLimit(ctx context.Context, customerID uuid.UUID, newLimit float64) error {
	customer, err := s.repo.GetByID(ctx, customerID)
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
func (s *Service) GetOverdueCustomers(ctx context.Context) ([]OverdueCustomer, error) {
	// Query to get customers with debts and their debt entries
	query := `
		SELECT c.id, c.name, c.code, c.current_balance, c.credit_limit,
			c.current_balance as overdue_amount
		FROM customers c
		WHERE c.is_active = true
		AND c.current_balance > 0
		ORDER BY c.current_balance DESC
	`

	var overdueCustomers []OverdueCustomer
	err := s.repo.db.SelectContext(ctx, &overdueCustomers, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue customers: %w", err)
	}

	// For each customer, get their debts
	for i := range overdueCustomers {
		debtQuery := `
			SELECT id, amount, remaining_amount, due_date, status
			FROM debts
			WHERE customer_id = $1
			ORDER BY due_date ASC
		`
		
		type DebtInfo struct {
			ID              string  `db:"id"`
			Amount          float64 `db:"amount"`
			RemainingAmount float64 `db:"remaining_amount"`
			DueDate         string  `db:"due_date"`
			Status          string  `db:"status"`
		}
		
		var debts []DebtInfo
		err := s.repo.db.SelectContext(ctx, &debts, debtQuery, overdueCustomers[i].ID)
		if err != nil || len(debts) == 0 {
			// If no debts found, set empty array
			overdueCustomers[i].Debts = []map[string]interface{}{}
		} else {
			// Convert to map[string]interface{}
			debtMaps := make([]map[string]interface{}, len(debts))
			for j, debt := range debts {
				debtMaps[j] = map[string]interface{}{
					"id":               debt.ID,
					"amount":           debt.Amount,
					"remaining_amount": debt.RemainingAmount,
					"due_date":         debt.DueDate,
					"status":           debt.Status,
				}
			}
			overdueCustomers[i].Debts = debtMaps
		}
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
func (s *Service) CreateDebtEntry(ctx context.Context, customerID uuid.UUID, amount float64, referenceID uuid.UUID, referenceType string, dueDate time.Time) error {
	customer, err := s.repo.GetByID(ctx, customerID)
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
	if err := s.repo.UpdateBalance(ctx, customerID, amount); err != nil {
		return fmt.Errorf("failed to update customer balance: %w", err)
	}

	return nil
}

// GetDebtEntries retrieves debt entries for a customer
func (s *Service) GetDebtEntries(ctx context.Context, customerID uuid.UUID) ([]DebtEntry, error) {
	// Verify customer exists
	_, err := s.repo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetDebtEntries(ctx, customerID)
}

// CreateDebtCollection creates a new debt collection action
func (s *Service) CreateDebtCollection(ctx context.Context, customerID uuid.UUID, collectionType string, scheduledDate time.Time, notes *string) error {
	// Verify customer exists
	_, err := s.repo.GetByID(ctx, customerID)
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
func (s *Service) GetDebtCollections(ctx context.Context, customerID uuid.UUID) ([]DebtCollection, error) {
	// Verify customer exists
	_, err := s.repo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetDebtCollections(ctx, customerID)
}

// GetPendingDebtCollections retrieves pending debt collection actions
func (s *Service) GetPendingDebtCollections(ctx context.Context) ([]DebtCollection, error) {
	return s.repo.GetPendingDebtCollections(ctx)
}

// ProcessDebtPayment processes a payment for specific debts
func (s *Service) ProcessDebtPayment(ctx context.Context, customerID uuid.UUID, paymentAmount float64, method string) error {
	_, err := s.repo.GetByID(ctx, customerID)
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
	if err := s.repo.UpdateBalance(ctx, customerID, -paymentAmount); err != nil {
		return fmt.Errorf("failed to update customer balance: %w", err)
	}

	return nil
}

// GeneratePaymentReceipt generates a PDF receipt for payment
func (s *Service) GeneratePaymentReceipt(ctx context.Context, customerID uuid.UUID, req *PaymentReceiptRequest) ([]byte, error) {
	// Verify customer exists
	_, err := s.repo.GetByID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// Determine language (default to Arabic)
	language := req.Language
	if language == "" {
		language = "ar"
	}
	isRTL := language == "ar"

	// Create PDF with professional design
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Color scheme
	const (
		primaryColor = 34  // Green
		secondaryColor = 197
		accentColor = 94
		textDark = 0
		textMedium = 80
		textLight = 150
		borderColor = 200
	)

	// Set font based on language
	if isRTL {
		pdf.SetFont("Arial", "", 12)
	} else {
		pdf.SetFont("Arial", "", 12)
	}

	// Helper function to write text with RTL support
	writeText := func(text string, x, y float64, fontSize float64, isBold bool, colorR, colorG, colorB int) {
		pdf.SetFont("Arial", "", fontSize)
		if isBold {
			pdf.SetFont("Arial", "B", fontSize)
		}
		pdf.SetTextColor(colorR, colorG, colorB)
		pdf.SetXY(x, y)
		pdf.Cell(0, 8, text)
	}

	// Helper function to draw box
	drawBox := func(x, y, w, h float64, colorR, colorG, colorB int) {
		pdf.SetDrawColor(colorR, colorG, colorB)
		pdf.SetLineWidth(0.5)
		pdf.Rect(x, y, w, h, "D")
	}

	// ==================== HEADER SECTION ====================
	// Draw header background
	pdf.SetFillColor(primaryColor, secondaryColor, accentColor)
	pdf.Rect(20, 20, 170, 40, "F")

	// Company logo/title
	writeText("PartFlow", 95, 28, 24.0, true, 255, 255, 255)
	writeText("Store Management System", 95, 38, 10.0, false, 200, 200, 200)

	// Receipt number and date in header
	receiptNumber := time.Now().Format("20060102150405")[:8]
	if isRTL {
		writeText(fmt.Sprintf("رقم الإيصال: %s", receiptNumber), 25, 28, 10.0, false, 255, 255, 255)
		writeText(fmt.Sprintf("التاريخ: %s", req.Date), 25, 38, 10.0, false, 200, 200, 200)
	} else {
		writeText(fmt.Sprintf("Receipt #: %s", receiptNumber), 25, 28, 10.0, false, 255, 255, 255)
		writeText(fmt.Sprintf("Date: %s", req.Date), 25, 38, 10.0, false, 200, 200, 200)
	}

	// ==================== CUSTOMER SECTION ====================
	yPos := 70.0
	drawBox(20, yPos, 170, 30, borderColor, borderColor, borderColor)
	
	if isRTL {
		writeText("معلومات العميل", 25, yPos+5, 14.0, true, textDark, textDark, textDark)
		writeText(fmt.Sprintf("الاسم: %s", req.CustomerName), 25, yPos+15, 12.0, false, textMedium, textMedium, textMedium)
	} else {
		writeText("Customer Information", 25, yPos+5, 14.0, true, textDark, textDark, textDark)
		writeText(fmt.Sprintf("Name: %s", req.CustomerName), 25, yPos+15, 12.0, false, textMedium, textMedium, textMedium)
	}

	// ==================== PAYMENT DETAILS TABLE ====================
	yPos += 40
	drawBox(20, yPos, 170, 80, borderColor, borderColor, borderColor)

	// Table header
	pdf.SetFillColor(primaryColor, secondaryColor, accentColor)
	pdf.Rect(20, yPos, 170, 15, "F")
	
	if isRTL {
		writeText("تفاصيل الدفعة", 95, yPos+5, 14.0, true, 255, 255, 255)
	} else {
		writeText("Payment Details", 95, yPos+5, 14.0, true, 255, 255, 255)
	}

	// Payment method
	methodText := ""
	if isRTL {
		switch req.Method {
		case "cash":
			methodText = "نقدي"
		case "credit":
			methodText = "بطاقة ائتمان"
		case "bank_transfer":
			methodText = "تحويل بنكي"
		case "check":
			methodText = "شيك"
		default:
			methodText = "نقدي"
		}
		writeText(fmt.Sprintf("طريقة الدفع: %s", methodText), 25, yPos+25, 12.0, false, textDark, textDark, textDark)
		writeText(fmt.Sprintf("المبلغ: ₪%.2f", req.Amount), 25, yPos+40, 12.0, false, textDark, textDark, textDark)
	} else {
		switch req.Method {
		case "cash":
			methodText = "Cash"
		case "credit":
			methodText = "Credit Card"
		case "bank_transfer":
			methodText = "Bank Transfer"
		case "check":
			methodText = "Check"
		default:
			methodText = "Cash"
		}
		writeText(fmt.Sprintf("Payment Method: %s", methodText), 25, yPos+25, 12.0, false, textDark, textDark, textDark)
		writeText(fmt.Sprintf("Amount: ₪%.2f", req.Amount), 25, yPos+40, 12.0, false, textDark, textDark, textDark)
	}

	// Total amount with larger font
	pdf.SetFont("Arial", "B", 20.0)
	pdf.SetTextColor(primaryColor, secondaryColor, accentColor)
	pdf.SetXY(25, yPos+60)
	pdf.Cell(0, 12, fmt.Sprintf("₪%.2f", req.Amount))

	if isRTL {
		writeText("المجموع", 120, yPos+62, 14.0, true, textMedium, textMedium, textMedium)
	} else {
		writeText("Total", 120, yPos+62, 14.0, true, textMedium, textMedium, textMedium)
	}

	// ==================== FOOTER SECTION ====================
	yPos += 90
	drawBox(20, yPos, 170, 25, borderColor, borderColor, borderColor)
	
	if isRTL {
		writeText("شكراً لتعاملكم معنا", 95, yPos+8, 12.0, true, textMedium, textMedium, textMedium)
		writeText("PartFlow - نظام إدارة المتاجر", 95, yPos+18, 10.0, false, textLight, textLight, textLight)
	} else {
		writeText("Thank you for your business", 95, yPos+8, 12.0, true, textMedium, textMedium, textMedium)
		writeText("PartFlow - Store Management System", 95, yPos+18, 10.0, false, textLight, textLight, textLight)
	}

	// ==================== TERMS AND CONDITIONS ====================
	yPos += 30
	writeText("Terms & Conditions / الشروط والأحكام:", 20, yPos, 9.0, false, textLight, textLight, textLight)
	
	if isRTL {
		writeText("• هذا الإيصال إثبات للدفعة المسجلة", 20, yPos+8, 8.0, false, textLight, textLight, textLight)
		writeText("• يرجى الاحتفاظ به للمراجعة", 20, yPos+16, 8.0, false, textLight, textLight, textLight)
		writeText("• لأي استفسار، يرجى التواصل مع الإدارة", 20, yPos+24, 8.0, false, textLight, textLight, textLight)
	} else {
		writeText("• This receipt serves as proof of payment", 20, yPos+8, 8.0, false, textLight, textLight, textLight)
		writeText("• Please keep for your records", 20, yPos+16, 8.0, false, textLight, textLight, textLight)
		writeText("• For inquiries, please contact management", 20, yPos+24, 8.0, false, textLight, textLight, textLight)
	}

	// Generate PDF bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}
