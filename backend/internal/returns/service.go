package returns

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/partflow/smart-store/internal/dashboard"
)

// Service handles return business logic
type Service struct {
	repo *Repository
}

// NewService creates a new return service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) findActiveDebtIDForCustomer(ctx context.Context, customerID uuid.UUID, preferred *uuid.UUID) (*uuid.UUID, error) {
	if preferred != nil && *preferred != uuid.Nil {
		return preferred, nil
	}
	if customerID == uuid.Nil {
		return nil, nil
	}

	var debtID uuid.UUID
	err := s.repo.db.GetContext(ctx, &debtID, `SELECT id FROM debts WHERE customer_id = $1 AND remaining_amount > 0 AND status IN ('pending','partial','overdue') ORDER BY due_date ASC, created_at ASC LIMIT 1`, customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find active debt: %w", err)
	}
	return &debtID, nil
}

// CreateReturn creates a new return with items
func (s *Service) CreateReturn(ctx context.Context, userID uuid.UUID, req *ReturnRequest) (*ReturnResponse, error) {
	// Validate request
	if err := ValidateReturnRequest(req); err != nil {
		return nil, err
	}

	// Get sale info
	if req.SaleID == nil || *req.SaleID == uuid.Nil {
		return nil, ErrSaleNotFound
	}

	sale, err := s.repo.GetSaleInfo(ctx, *req.SaleID)
	if err != nil {
		return nil, ErrSaleNotFound
	}

	// Reject duplicate or over-quantity returns before creating the parent
	// record. The database trigger remains a final safety net.
	for _, itemReq := range req.Items {
		if itemReq.SaleItemID == nil || *itemReq.SaleItemID == uuid.Nil {
			continue
		}

		saleItem, err := s.repo.GetSaleItemInfo(ctx, *itemReq.SaleItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get sale item info: %w", err)
		}
		returnedQty, err := s.repo.GetReturnedQuantity(ctx, *itemReq.SaleItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get returned quantity: %w", err)
		}
		if returnedQty+itemReq.QuantityReturned > saleItem.Quantity {
			return nil, ErrInsufficientStock
		}
	}

	// Create return
	returnRecord := CreateReturn(userID, req)

	// Set customer ID from sale if not provided in request
	if req.CustomerID == nil {
		returnRecord.CustomerID = sale.CustomerID
	}

	if returnRecord.RefundMethod == "DEBT_ADJUSTMENT" {
		debtID, err := s.findActiveDebtIDForCustomer(ctx, returnRecord.CustomerID, req.DebtID)
		if err != nil {
			return nil, err
		}
		returnRecord.DebtID = debtID
	}

	if err := s.repo.CreateReturn(ctx, returnRecord); err != nil {
		return nil, fmt.Errorf("failed to create return: %w", err)
	}

	// Create return items
	var items []ReturnItem
	var totalRefund float64

	for _, itemReq := range req.Items {
		// Get sale item info if sale item ID is provided
		var unitPrice float64
		if itemReq.SaleItemID != nil && *itemReq.SaleItemID != uuid.Nil {
			saleItem, err := s.repo.GetSaleItemInfo(ctx, *itemReq.SaleItemID)
			if err != nil {
				return nil, fmt.Errorf("failed to get sale item info: %w", err)
			}

			// Validate quantity
			if itemReq.QuantityReturned > saleItem.Quantity {
				return nil, ErrInsufficientStock
			}
			unitPrice = saleItem.UnitPrice
		} else {
			// Use provided unit price if no sale item
			unitPrice = itemReq.UnitPrice
		}

		item := CreateReturnItem(returnRecord.ID, itemReq, unitPrice)
		if err := s.repo.CreateReturnItem(ctx, item); err != nil {
			return nil, fmt.Errorf("failed to create return item: %w", err)
		}
		items = append(items, *item)
		totalRefund += item.TotalRefundAmount
	}

	// Update return with total refund amount
	returnRecord.TotalRefundAmount = totalRefund
	returnRecord.UpdatedAt = time.Now()
	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, fmt.Errorf("failed to update return: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("return_created")

	// Customer information is optional for walk-in sales.
	var customer *CustomerInfo
	if returnRecord.CustomerID != uuid.Nil {
		customer, err = s.repo.GetCustomerInfo(ctx, returnRecord.CustomerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get customer info: %w", err)
		}
	}

	return returnRecord.ToReturnResponse(items, customer, sale), nil
}

// GetReturn retrieves a return by ID
func (s *Service) GetReturn(ctx context.Context, id uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetReturnItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get return items: %w", err)
	}

	var customer *CustomerInfo
	if returnRecord.CustomerID != uuid.Nil {
		customer, err = s.repo.GetCustomerInfo(ctx, returnRecord.CustomerID)
		if err != nil {
			// Customer not found is not a fatal error
			customer = nil
		}
	}

	var sale *SaleInfo
	if returnRecord.SaleID != uuid.Nil {
		sale, err = s.repo.GetSaleInfo(ctx, returnRecord.SaleID)
		if err != nil {
			// Sale not found is not a fatal error
			sale = nil
		}
	}

	return returnRecord.ToReturnResponse(items, customer, sale), nil
}

// ListReturns retrieves returns with pagination and filters
func (s *Service) ListReturns(ctx context.Context, req ReturnListRequest) ([]Return, int, error) {
	return s.repo.ListReturns(ctx, req)
}

// UpdateReturn updates a return
func (s *Service) UpdateReturn(ctx context.Context, id uuid.UUID, req *ReturnUpdateRequest) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if return can be updated
	if returnRecord.Status == "COMPLETED" {
		return nil, ErrReturnAlreadyCompleted
	}

	// Update fields
	if req.Reason != "" {
		returnRecord.Reason = req.Reason
	}
	if req.Status != "" {
		// Validate status transition
		if returnRecord.Status == "APPROVED" && req.Status == "PENDING" {
			return nil, ErrInvalidReturnStatus
		}
		if returnRecord.Status == "REJECTED" && req.Status == "PENDING" {
			return nil, ErrInvalidReturnStatus
		}
		returnRecord.Status = req.Status
	}
	if req.TotalRefundAmount >= 0 {
		returnRecord.TotalRefundAmount = req.TotalRefundAmount
	}
	if req.RefundMethod != "" {
		returnRecord.RefundMethod = req.RefundMethod
		if returnRecord.RefundMethod == "DEBT_ADJUSTMENT" && returnRecord.DebtID == nil {
			debtID, err := s.findActiveDebtIDForCustomer(ctx, returnRecord.CustomerID, req.DebtID)
			if err != nil {
				return nil, err
			}
			returnRecord.DebtID = debtID
		}
	}
	if req.RefundDate != nil {
		returnRecord.RefundDate = req.RefundDate
	}
	if req.RefundReference != "" {
		returnRecord.RefundReference = req.RefundReference
	}
	if req.DebtID != nil {
		returnRecord.DebtID = req.DebtID
	}
	if req.DebtAdjustment != 0 {
		returnRecord.DebtAdjustment = req.DebtAdjustment
	}
	if req.ItemConditionAfterReturn != "" {
		returnRecord.ItemConditionAfterReturn = req.ItemConditionAfterReturn
	}
	if req.Notes != "" {
		returnRecord.Notes = req.Notes
	}
	if req.InternalNotes != "" {
		returnRecord.InternalNotes = req.InternalNotes
	}
	if req.ApprovedBy != nil {
		returnRecord.ApprovedBy = req.ApprovedBy
		now := time.Now()
		returnRecord.ApprovedAt = &now
	}

	returnRecord.UpdatedAt = time.Now()

	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, err
	}
	dashboard.InvalidateDashboardCacheWithReason("return_updated")

	return s.GetReturn(ctx, id)
}

// DeleteReturn deletes a return
func (s *Service) DeleteReturn(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteReturn(ctx, id); err != nil {
		return err
	}
	dashboard.InvalidateDashboardCacheWithReason("return_deleted")
	return nil
}

// ApproveReturn approves a return
func (s *Service) ApproveReturn(ctx context.Context, id uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if returnRecord.Status != "PENDING" {
		return nil, ErrInvalidReturnStatus
	}

	returnRecord.Status = "APPROVED"
	returnRecord.UpdatedAt = time.Now()

	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, err
	}
	dashboard.InvalidateDashboardCacheWithReason("return_approved")

	return s.GetReturn(ctx, id)
}

// RejectReturn rejects a return
func (s *Service) RejectReturn(ctx context.Context, id uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if returnRecord.Status != "PENDING" {
		return nil, ErrInvalidReturnStatus
	}

	returnRecord.Status = "REJECTED"
	returnRecord.UpdatedAt = time.Now()

	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, err
	}
	dashboard.InvalidateDashboardCacheWithReason("return_rejected")

	return s.GetReturn(ctx, id)
}

// ProcessRefund processes refund for a return
func (s *Service) ProcessRefund(ctx context.Context, id uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if returnRecord.Status != "APPROVED" {
		return nil, ErrInvalidReturnStatus
	}

	now := time.Now()
	returnRecord.Status = "PROCESSING"
	returnRecord.RefundDate = &now
	returnRecord.UpdatedAt = now

	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, err
	}
	dashboard.InvalidateDashboardCacheWithReason("return_refund_processing")

	return s.GetReturn(ctx, id)
}

// AddReturnItem adds an item to a return
func (s *Service) AddReturnItem(ctx context.Context, returnID uuid.UUID, req ReturnItemRequest) (*ReturnItem, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, returnID)
	if err != nil {
		return nil, err
	}

	// Get sale item info if provided
	var unitPrice float64
	if req.SaleItemID != nil {
		saleItem, err := s.repo.GetSaleItemInfo(ctx, *req.SaleItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get sale item info: %w", err)
		}

		// Validate quantity
		if req.QuantityReturned > saleItem.Quantity {
			return nil, ErrInsufficientStock
		}

		unitPrice = saleItem.UnitPrice
	} else if req.UnitPrice > 0 {
		unitPrice = req.UnitPrice
	} else {
		return nil, fmt.Errorf("must provide either sale_item_id or unit_price")
	}

	item := CreateReturnItem(returnID, req, unitPrice)
	if err := s.repo.CreateReturnItem(ctx, item); err != nil {
		return nil, err
	}

	// Update return total refund amount
	returnRecord.TotalRefundAmount += item.TotalRefundAmount
	returnRecord.UpdatedAt = time.Now()
	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, err
	}
	dashboard.InvalidateDashboardCacheWithReason("return_item_added")

	return item, nil
}

// UpdateReturnItem updates a return item
func (s *Service) UpdateReturnItem(ctx context.Context, itemID uuid.UUID, req ReturnItemRequest) (*ReturnItem, error) {
	// Get the item first
	item, err := s.repo.GetReturnItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.QuantityReturned > 0 {
		item.QuantityReturned = req.QuantityReturned
	}
	if req.UnitPrice > 0 {
		item.UnitPrice = req.UnitPrice
	}
	if req.TotalRefundAmount > 0 {
		item.TotalRefundAmount = req.TotalRefundAmount
	}
	if req.ReturnedCondition != "" {
		item.ReturnedCondition = req.ReturnedCondition
	}
	if req.ConditionNotes != "" {
		item.ConditionNotes = req.ConditionNotes
	}
	if req.Resolution != "" {
		item.Resolution = req.Resolution
	}
	if req.OriginalCost != nil {
		item.OriginalCost = req.OriginalCost
	}
	if req.RepairCost >= 0 {
		item.RepairCost = req.RepairCost
	}
	item.UpdatedAt = time.Now()

	if err := s.repo.UpdateReturnItem(ctx, item); err != nil {
		return nil, err
	}
	dashboard.InvalidateDashboardCacheWithReason("return_item_updated")

	return item, nil
}

// DeleteReturnItem deletes a return item
func (s *Service) DeleteReturnItem(ctx context.Context, itemID uuid.UUID) error {
	if err := s.repo.DeleteReturnItem(ctx, itemID); err != nil {
		return err
	}
	dashboard.InvalidateDashboardCacheWithReason("return_item_deleted")
	return nil
}

// GetMonthlyReturnsAnalysis gets monthly returns analysis
func (s *Service) GetMonthlyReturnsAnalysis(ctx context.Context) ([]MonthlyReturnsAnalysis, error) {
	return s.repo.GetMonthlyReturnsAnalysis(ctx)
}

// GetSalesReturnsAnalysis gets sales vs returns analysis
func (s *Service) GetSalesReturnsAnalysis(ctx context.Context) ([]SalesReturnsAnalysis, error) {
	return s.repo.GetSalesReturnsAnalysis(ctx)
}

// CompleteReturn completes a return and processes financial effects
func (s *Service) CompleteReturn(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(returnRecord.Status, "COMPLETED") {
		return nil, ErrReturnAlreadyCompleted
	}
	if !strings.EqualFold(returnRecord.Status, "APPROVED") && !strings.EqualFold(returnRecord.Status, "PROCESSING") {
		return nil, ErrInvalidReturnStatus
	}

	items, err := s.repo.GetReturnItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load return items: %w", err)
	}

	now := time.Now()
	returnRecord.Status = "COMPLETED"
	var approvedUserExists bool
	if approvedBy != uuid.Nil {
		if err := s.repo.db.GetContext(ctx, &approvedUserExists,
			`SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, approvedBy); err != nil {
			return nil, fmt.Errorf("failed to validate approving user: %w", err)
		}
	}
	if approvedUserExists {
		returnRecord.ProcessedBy = &approvedBy
		returnRecord.ApprovedBy = &approvedBy
	} else {
		returnRecord.ProcessedBy = nil
		returnRecord.ApprovedBy = nil
	}
	returnRecord.ApprovedAt = &now
	returnRecord.UpdatedAt = now

	// Handle debt adjustment if applicable
	if returnRecord.RefundMethod == "DEBT_ADJUSTMENT" {
		if returnRecord.DebtID == nil {
			debtID, err := s.findActiveDebtIDForCustomer(ctx, returnRecord.CustomerID, nil)
			if err != nil {
				return nil, err
			}
			returnRecord.DebtID = debtID
		}
		if returnRecord.DebtAdjustment == 0 {
			returnRecord.DebtAdjustment = returnRecord.TotalRefundAmount
		}
	}

	if returnRecord.RefundMethod == "CASH" || returnRecord.RefundMethod == "BANK_TRANSFER" {
		returnRecord.RefundDate = &now
	}

	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, fmt.Errorf("failed to complete return: %w", err)
	}
	if err := s.repo.AddDebtAdjustmentLedgerEntry(ctx, returnRecord); err != nil {
		return nil, err
	}
	if err := s.createSupplierReturnBridge(ctx, returnRecord, items); err != nil {
		return nil, err
	}
	dashboard.InvalidateDashboardCacheWithReason("return_completed")

	// The database trigger will handle the actual debt adjustment
	// This ensures consistency and prevents race conditions

	return s.GetReturn(ctx, id)
}

func (s *Service) createSupplierReturnBridge(ctx context.Context, returnRecord *Return, items []ReturnItem) error {
	needsSupplierBridge := strings.EqualFold(returnRecord.ItemConditionAfterReturn, "RETURN_TO_SUPPLIER")
	if !needsSupplierBridge {
		for _, item := range items {
			if strings.EqualFold(item.Resolution, "SUPPLIER_RETURN") {
				needsSupplierBridge = true
				break
			}
		}
	}
	if !needsSupplierBridge {
		return nil
	}

	for _, item := range items {
		if item.ProductID == nil || *item.ProductID == uuid.Nil {
			continue
		}
		if !strings.EqualFold(returnRecord.ItemConditionAfterReturn, "RETURN_TO_SUPPLIER") && !strings.EqualFold(item.Resolution, "SUPPLIER_RETURN") {
			continue
		}

		purchaseItemID, purchaseID, supplierID, err := s.findPurchaseItemForReturnItem(ctx, item)
		if err != nil {
			return err
		}
		if purchaseItemID == uuid.Nil || purchaseID == uuid.Nil || supplierID == uuid.Nil {
			continue
		}

		if err := s.ensureSupplierReturnBridgeEntry(ctx, returnRecord, purchaseID, supplierID, purchaseItemID, item); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) findPurchaseItemForReturnItem(ctx context.Context, item ReturnItem) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	if item.ProductID == nil || *item.ProductID == uuid.Nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, nil
	}

	var candidate struct {
		PurchaseItemID string `db:"purchase_item_id"`
		PurchaseID     string `db:"purchase_id"`
		SupplierID     string `db:"supplier_id"`
	}

	query := `
		SELECT pi.id AS purchase_item_id, pi.purchase_id, p.supplier_id
		FROM purchase_items pi
		JOIN purchases p ON p.id = pi.purchase_id
		WHERE pi.product_id = ?
		ORDER BY pi.created_at DESC
		LIMIT 1
	`
	if err := s.repo.db.GetContext(ctx, &candidate, query, item.ProductID.String()); err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, uuid.Nil, uuid.Nil, nil
		}
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("failed to find purchase item for supplier return bridge: %w", err)
	}

	purchaseItemID, err := uuid.Parse(candidate.PurchaseItemID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("parse purchase_item_id for supplier return bridge: %w", err)
	}
	purchaseID, err := uuid.Parse(candidate.PurchaseID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("parse purchase_id for supplier return bridge: %w", err)
	}
	supplierID, err := uuid.Parse(candidate.SupplierID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("parse supplier_id for supplier return bridge: %w", err)
	}
	return purchaseItemID, purchaseID, supplierID, nil
}

func (s *Service) ensureSupplierReturnBridgeEntry(ctx context.Context, returnRecord *Return, purchaseID, supplierID, purchaseItemID uuid.UUID, item ReturnItem) error {
	var supplierReturnID string
	if err := s.repo.db.GetContext(ctx, &supplierReturnID,
		`SELECT id FROM supplier_returns WHERE purchase_id = ? ORDER BY created_at DESC LIMIT 1`,
		purchaseID.String()); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to load existing supplier return bridge: %w", err)
	}

	if supplierReturnID == "" {
		supplierReturnID = uuid.New().String()
		returnNumber := "SRET-" + strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:10], "-", ""))
		now := time.Now().UTC().Format(time.RFC3339Nano)
		var createdBy interface{}
		if returnRecord.CreatedBy != nil && *returnRecord.CreatedBy != uuid.Nil {
			createdBy = returnRecord.CreatedBy.String()
		} else if returnRecord.ProcessedBy != nil && *returnRecord.ProcessedBy != uuid.Nil {
			createdBy = returnRecord.ProcessedBy.String()
		}
		_, err := s.repo.db.ExecContext(ctx, `
			INSERT INTO supplier_returns (id, purchase_id, supplier_id, return_number, status, reason, refund_amount, notes, created_by, created_at, updated_at)
			VALUES (?, ?, ?, ?, 'PENDING', ?, 0, ?, ?, ?, ?)
		`, supplierReturnID, purchaseID.String(), supplierID.String(), returnNumber, fmt.Sprintf("Customer return %s", returnRecord.ReturnNumber), fmt.Sprintf("Auto-created from customer return %s", returnRecord.ReturnNumber), createdBy, now, now)
		if err != nil {
			return fmt.Errorf("failed to create supplier return bridge: %w", err)
		}
	}

	var existingSupplierItemCount int
	if err := s.repo.db.GetContext(ctx, &existingSupplierItemCount,
		`SELECT COUNT(*) FROM supplier_return_items WHERE supplier_return_id = ? AND purchase_item_id = ?`,
		supplierReturnID, purchaseItemID.String()); err != nil {
		return fmt.Errorf("failed to check supplier return bridge item: %w", err)
	}
	if existingSupplierItemCount > 0 {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := s.repo.db.ExecContext(ctx, `
		INSERT INTO supplier_return_items (id, supplier_return_id, purchase_item_id, product_id, quantity, unit_cost, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, uuid.New().String(), supplierReturnID, purchaseItemID.String(), item.ProductID.String(), item.QuantityReturned, item.UnitPrice, now); err != nil {
		return fmt.Errorf("failed to add supplier return bridge item: %w", err)
	}
	return nil
}

// GetReturnBySale retrieves returns for a specific sale
func (s *Service) GetReturnBySale(ctx context.Context, saleID uuid.UUID) ([]Return, error) {
	return s.repo.GetReturnsBySaleID(ctx, saleID)
}

// ValidateReturnQuantity validates return quantity against original sale
func (s *Service) ValidateReturnQuantity(ctx context.Context, saleItemID uuid.UUID, quantity int) error {
	saleItem, err := s.repo.GetSaleItemInfo(ctx, saleItemID)
	if err != nil {
		return fmt.Errorf("failed to get sale item: %w", err)
	}

	// Get already returned quantity
	returnedQty, err := s.repo.GetReturnedQuantity(ctx, saleItemID)
	if err != nil {
		return fmt.Errorf("failed to get returned quantity: %w", err)
	}

	if returnedQty+quantity > saleItem.Quantity {
		return ErrInsufficientStock
	}

	return nil
}

// GetReturnSummary gets return summary with related data
func (s *Service) GetReturnSummary(ctx context.Context) ([]map[string]interface{}, error) {
	return s.repo.GetReturnSummary(ctx)
}

// ReverseReturn reverses a return instead of deleting it
func (s *Service) ReverseReturn(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if returnRecord.Status == "REVERSED" {
		return nil, ErrInvalidReturnStatus
	}

	// Create reversal record
	reversal := &Return{
		ReturnNumber:      "", // Auto-generated by trigger
		ReferenceNumber:   "REV-" + returnRecord.ReturnNumber,
		SaleID:            returnRecord.SaleID,
		CustomerID:        returnRecord.CustomerID,
		ReturnDate:        time.Now(),
		ReturnType:        returnRecord.ReturnType,
		Status:            "COMPLETED",
		TotalRefundAmount: -returnRecord.TotalRefundAmount, // Negative to reverse
		RefundMethod:      "REVERSAL",
		Reason:            "RETURN_REVERSAL",
		ReasonDetail:      "Reversal of return " + returnRecord.ReturnNumber,
		CreatedBy:         &userID,
		Notes:             "Automatic reversal of return " + returnRecord.ReturnNumber,
	}

	if err := s.repo.CreateReturn(ctx, reversal); err != nil {
		return nil, fmt.Errorf("failed to create reversal: %w", err)
	}

	// Update original return status
	returnRecord.Status = "REVERSED"
	returnRecord.InternalNotes = "Reversed by return " + reversal.ReturnNumber
	returnRecord.UpdatedAt = time.Now()
	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, fmt.Errorf("failed to update original return: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("return_reversed")

	return s.GetReturn(ctx, id)
}

// GetReturnsByCustomer retrieves returns for a specific customer
func (s *Service) GetReturnsByCustomer(ctx context.Context, customerID uuid.UUID) ([]Return, error) {
	return s.repo.GetReturnsByCustomer(ctx, customerID)
}

// GetPendingReturns retrieves returns that are pending approval
func (s *Service) GetPendingReturns(ctx context.Context) ([]Return, error) {
	return s.repo.GetPendingReturns(ctx)
}

// GetReturnStatistics returns statistics about returns
func (s *Service) GetReturnStatistics(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetReturnStatistics(ctx)
}

// GetReturnWithItems retrieves a return with all its items
func (s *Service) GetReturnWithItems(ctx context.Context, returnID uuid.UUID) (*Return, []ReturnItem, error) {
	return s.repo.GetReturnWithItems(ctx, returnID)
}

// ProcessReturnStatusChange handles status changes and their effects
func (s *Service) ProcessReturnStatusChange(ctx context.Context, returnID uuid.UUID, newStatus string, processedBy uuid.UUID) error {
	return s.repo.ProcessReturnStatusChange(ctx, returnID, newStatus, processedBy)
}

// UpdateReturnItemStatus updates the inventory status of a return item
func (s *Service) UpdateReturnItemStatus(ctx context.Context, itemID uuid.UUID, newStatus string) error {
	return s.repo.UpdateReturnItemStatus(ctx, itemID, newStatus)
}
