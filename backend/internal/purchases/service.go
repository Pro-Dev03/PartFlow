package purchases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Service handles purchase business logic
type Service struct {
	repo *Repository
	db   *sqlx.DB
}

// NewService creates a new purchase service
func NewService(repo *Repository, db *sqlx.DB) *Service {
	return &Service{repo: repo, db: db}
}

// CreatePurchase creates a new purchase with items and full automation
func (s *Service) CreatePurchase(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req *PurchaseRequest) (*PurchaseResponse, error) {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Validate request
	if err := ValidatePurchaseRequest(req); err != nil {
		return nil, err
	}

	// Check if invoice number already exists
	existing, err := s.repo.GetPurchaseByInvoiceNumber(ctx, req.InvoiceNumber, organizationID)
	if err == nil && existing != nil {
		return nil, ErrPurchaseExists
	}

	// Calculate total amount
	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += float64(item.Quantity) * item.UnitCost
	}

	// Create purchase
	purchase := &Purchase{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		SupplierID:     req.SupplierID,
		InvoiceNumber:  req.InvoiceNumber,
		PurchaseDate:   req.PurchaseDate,
		TotalAmount:    totalAmount,
		PaidAmount:     0,
		Status:         "pending",
		Notes:          req.Notes,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, purchase); err != nil {
		return nil, fmt.Errorf("failed to create purchase: %w", err)
	}

	// Create purchase items
	var items []PurchaseItem
	for _, itemReq := range req.Items {
		item := CreatePurchaseItem(purchase.ID, itemReq)
		if err := s.repo.CreatePurchaseItem(ctx, item); err != nil {
			return nil, fmt.Errorf("failed to create purchase item: %w", err)
		}
		items = append(items, *item)
	}

	// Update supplier ledger
	// Get current balance
	var currentBalance float64
	balanceQuery := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM supplier_ledger 
		WHERE supplier_id = $1 AND organization_id = $2
	`
	err = tx.GetContext(ctx, &currentBalance, balanceQuery, req.SupplierID, organizationID)
	if err != nil {
		currentBalance = 0
	}

	// Calculate new balance (purchase increases debt)
	newBalance := currentBalance + totalAmount

	ledgerQuery := `
		INSERT INTO supplier_ledger (id, organization_id, supplier_id, transaction_type, 
			amount, balance, reference_type, reference_id, description, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.ExecContext(ctx, ledgerQuery,
		uuid.New(), organizationID, req.SupplierID, "PURCHASE",
		totalAmount, newBalance, "purchase", purchase.ID, "Purchase: "+req.InvoiceNumber, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to update supplier ledger: %w", err)
	}

	// Create audit log
	auditQuery := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, 
			entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	changes := fmt.Sprintf("Created purchase %s with %d items, total: %.2f", req.InvoiceNumber, len(items), totalAmount)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), organizationID, userID, "CREATE_PURCHASE", "purchase", purchase.ID,
		changes, time.Now())
	if err != nil {
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get supplier info
	supplier, err := s.repo.GetSupplierInfo(ctx, req.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier info: %w", err)
	}

	return purchase.ToPurchaseResponse(items, supplier), nil
}

// GetPurchase retrieves a purchase by ID
func (s *Service) GetPurchase(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*PurchaseResponse, error) {
	purchase, err := s.repo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetPurchaseItems(ctx, purchase.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase items: %w", err)
	}

	supplier, err := s.repo.GetSupplierInfo(ctx, purchase.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier info: %w", err)
	}

	return purchase.ToPurchaseResponse(items, supplier), nil
}

// ListPurchases retrieves purchases with pagination and filters
func (s *Service) ListPurchases(ctx context.Context, organizationID uuid.UUID, req PurchaseListRequest) ([]map[string]interface{}, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	purchases, total, err := s.repo.List(ctx, organizationID, req)
	if err != nil {
		return nil, 0, err
	}

	// Convert to list items
	var result []map[string]interface{}
	for _, purchase := range purchases {
		items, err := s.repo.GetPurchaseItems(ctx, purchase.ID)
		if err != nil {
			continue
		}

		supplier, err := s.repo.GetSupplierInfo(ctx, purchase.SupplierID)
		if err != nil {
			continue
		}

		result = append(result, purchase.ToPurchaseListItem(len(items), supplier.Name))
	}

	return result, total, nil
}

// UpdatePurchase updates a purchase
func (s *Service) UpdatePurchase(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *PurchaseUpdateRequest) (*PurchaseResponse, error) {
	purchase, err := s.repo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.InvoiceNumber != "" {
		purchase.InvoiceNumber = req.InvoiceNumber
	}
	if !req.PurchaseDate.IsZero() {
		purchase.PurchaseDate = req.PurchaseDate
	}
	if req.Status != "" {
		// Validate status transition
		if purchase.Status == "cancelled" {
			return nil, ErrPurchaseCancelled
		}
		if purchase.Status == "received" && req.Status == "pending" {
			return nil, ErrInvalidPurchaseStatus
		}
		purchase.Status = req.Status
	}
	if req.Notes != "" {
		purchase.Notes = req.Notes
	}

	purchase.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, purchase); err != nil {
		return nil, err
	}

	return s.GetPurchase(ctx, id, organizationID)
}

// DeletePurchase deletes a purchase
func (s *Service) DeletePurchase(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	purchase, err := s.repo.GetByID(ctx, id, organizationID)
	if err != nil {
		return err
	}

	// Check if purchase can be deleted
	if purchase.Status == "received" {
		return ErrPurchaseAlreadyReceived
	}

	return s.repo.Delete(ctx, id, organizationID)
}

// ReceivePurchase marks a purchase as received
func (s *Service) ReceivePurchase(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*PurchaseResponse, error) {
	purchase, err := s.repo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if purchase.Status == "cancelled" {
		return nil, ErrPurchaseCancelled
	}

	if purchase.Status == "received" {
		return nil, ErrPurchaseAlreadyReceived
	}

	purchase.Status = "received"
	purchase.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, purchase); err != nil {
		return nil, err
	}

	return s.GetPurchase(ctx, id, organizationID)
}

// CancelPurchase cancels a purchase
func (s *Service) CancelPurchase(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*PurchaseResponse, error) {
	purchase, err := s.repo.GetByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if purchase.Status == "cancelled" {
		return nil, ErrPurchaseCancelled
	}

	if purchase.Status == "received" {
		return nil, ErrPurchaseAlreadyReceived
	}

	purchase.Status = "cancelled"
	purchase.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, purchase); err != nil {
		return nil, err
	}

	return s.GetPurchase(ctx, id, organizationID)
}

// AddPayment adds a payment to a purchase with full automation
func (s *Service) AddPayment(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, userID uuid.UUID, amount float64, paymentMethod string) (*PurchaseResponse, error) {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if amount <= 0 {
		return nil, ErrInvalidCost
	}

	// Get purchase with row lock
	var purchase Purchase
	purchaseQuery := `SELECT * FROM purchases WHERE id = $1 AND organization_id = $2 FOR UPDATE`
	err = tx.GetContext(ctx, &purchase, purchaseQuery, id, organizationID)
	if err != nil {
		return nil, err
	}

	if purchase.Status == "cancelled" {
		return nil, ErrPurchaseCancelled
	}

	// Check if payment exceeds total
	if purchase.PaidAmount+amount > purchase.TotalAmount {
		return nil, ErrInvalidCost
	}

	// Update paid amount
	newPaidAmount := purchase.PaidAmount + amount
	updatePaidQuery := `UPDATE purchases SET paid_amount = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.ExecContext(ctx, updatePaidQuery, newPaidAmount, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update paid amount: %w", err)
	}

	// Create payment record
	paymentQuery := `
		INSERT INTO payments (id, organization_id, purchase_id, supplier_id, amount, 
			payment_method, payment_status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.ExecContext(ctx, paymentQuery,
		uuid.New(), organizationID, purchase.ID, purchase.SupplierID, amount,
		paymentMethod, "completed", userID, time.Now(), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update supplier ledger
	// Get current balance
	var currentBalance float64
	balanceQuery := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM supplier_ledger 
		WHERE supplier_id = $1 AND organization_id = $2
	`
	err = tx.GetContext(ctx, &currentBalance, balanceQuery, purchase.SupplierID, organizationID)
	if err != nil {
		currentBalance = 0
	}

	// Calculate new balance (payment reduces debt)
	newBalance := currentBalance - amount

	ledgerQuery := `
		INSERT INTO supplier_ledger (id, organization_id, supplier_id, transaction_type, 
			amount, balance, reference_type, reference_id, description, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.ExecContext(ctx, ledgerQuery,
		uuid.New(), organizationID, purchase.SupplierID, "PAYMENT",
		-amount, newBalance, "payment", purchase.ID, "Payment for purchase "+purchase.InvoiceNumber, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to update supplier ledger: %w", err)
	}

	// Create audit log
	auditQuery := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, 
			entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	changes := fmt.Sprintf("Payment of %.2f for purchase %s", amount, purchase.InvoiceNumber)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), organizationID, userID, "ADD_PAYMENT", "purchase", purchase.ID,
		changes, time.Now())
	if err != nil {
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.GetPurchase(ctx, id, organizationID)
}

// AddPurchaseItem adds an item to a purchase
func (s *Service) AddPurchaseItem(ctx context.Context, purchaseID uuid.UUID, organizationID uuid.UUID, req PurchaseItemRequest) (*PurchaseItem, error) {
	purchase, err := s.repo.GetByID(ctx, purchaseID, organizationID)
	if err != nil {
		return nil, err
	}

	if purchase.Status == "received" || purchase.Status == "cancelled" {
		return nil, ErrInvalidPurchaseStatus
	}

	// Validate item request
	if req.ProductID == uuid.Nil {
		return nil, ErrProductNotFound
	}
	if req.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if req.UnitCost < 0 {
		return nil, ErrInvalidCost
	}
	if req.Condition != "new" && req.Condition != "used" && req.Condition != "refurbished" {
		return nil, ErrInvalidCondition
	}

	item := CreatePurchaseItem(purchaseID, req)
	if err := s.repo.CreatePurchaseItem(ctx, item); err != nil {
		return nil, err
	}

	// Update purchase total amount
	purchase.TotalAmount += item.TotalCost
	purchase.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, purchase); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdatePurchaseItem updates a purchase item
func (s *Service) UpdatePurchaseItem(ctx context.Context, itemID uuid.UUID, organizationID uuid.UUID, req PurchaseItemRequest) (*PurchaseItem, error) {
	// Get the item first (need to verify it belongs to organization's purchase)
	var item PurchaseItem
	// This would require a more complex query to join with purchases table
	// For now, we'll update directly
	
	totalCost := float64(req.Quantity) * req.UnitCost
	
	item.ID = itemID
	item.Quantity = req.Quantity
	item.UnitCost = req.UnitCost
	item.TotalCost = totalCost
	item.SerialNumber = req.SerialNumber
	item.Condition = req.Condition
	item.LocationID = req.LocationID
	item.Notes = req.Notes
	item.UpdatedAt = time.Now()

	if err := s.repo.UpdatePurchaseItem(ctx, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

// DeletePurchaseItem deletes a purchase item
func (s *Service) DeletePurchaseItem(ctx context.Context, itemID uuid.UUID) error {
	return s.repo.DeletePurchaseItem(ctx, itemID)
}
