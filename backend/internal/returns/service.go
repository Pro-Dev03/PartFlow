package returns

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
	"github.com/partflow/smart-store/internal/paymenttransactions"
)

// Service handles return business logic
type Service struct {
	repo            *Repository
	refundProcessor ElectronicRefundProcessor
}

type ElectronicRefundProcessor interface {
	RefundForReturn(ctx context.Context, saleID, returnID uuid.UUID, amountMinor int64, createdBy *uuid.UUID) (paymenttransactions.ReturnRefundResult, error)
}

// NewService creates a new return service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetElectronicRefundProcessor(processor ElectronicRefundProcessor) {
	s.refundProcessor = processor
}

func (s *Service) loadRefundState(ctx context.Context, returnRecord *Return) error {
	var state returnRefundState
	err := s.repo.db.GetContext(ctx, &state, `SELECT payment_transaction_id, payment_refund_id, status, error_message FROM return_payment_refunds WHERE return_id = $1`, returnRecord.ID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	returnRecord.RefundStatus = state.Status
	returnRecord.PaymentTransactionID = returnStateUUIDPtr(state.PaymentTransactionID)
	returnRecord.PaymentRefundID = returnStateUUIDPtr(state.PaymentRefundID)
	returnRecord.RefundError = state.ErrorMessage.String
	return nil
}

type returnRefundState struct {
	PaymentTransactionID sql.NullString `db:"payment_transaction_id"`
	PaymentRefundID      sql.NullString `db:"payment_refund_id"`
	Status               string         `db:"status"`
	ErrorMessage         sql.NullString `db:"error_message"`
}

func returnStateUUIDPtr(value sql.NullString) *uuid.UUID {
	if !value.Valid || value.String == "" {
		return nil
	}
	id, err := uuid.Parse(value.String)
	if err != nil {
		return nil
	}
	return &id
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

	requiresSupplierSource := strings.EqualFold(req.ItemConditionAfterReturn, "RETURN_TO_SUPPLIER")

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
		if requiresSupplierSource {
			inventoryItemID := itemReq.InventoryItemID
			if inventoryItemID == nil {
				inventoryItemID = saleItem.InventoryItemID
			}
			productID := itemReq.ProductID
			if productID == nil {
				productID = &saleItem.ProductID
			}
			candidate := ReturnItem{SaleItemID: itemReq.SaleItemID, InventoryItemID: inventoryItemID, ProductID: productID}
			purchaseItemID, purchaseID, supplierID, sourceErr := s.findPurchaseItemForReturnItem(ctx, candidate)
			if sourceErr != nil {
				return nil, sourceErr
			}
			if purchaseItemID == uuid.Nil || purchaseID == uuid.Nil || supplierID == uuid.Nil {
				return nil, ErrSupplierSourceUnavailable
			}
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
		var originalQuantity *int
		if itemReq.SaleItemID != nil && *itemReq.SaleItemID != uuid.Nil {
			saleItem, err := s.repo.GetSaleItemInfo(ctx, *itemReq.SaleItemID)
			if err != nil {
				return nil, fmt.Errorf("failed to get sale item info: %w", err)
			}
			if itemReq.InventoryItemID == nil {
				itemReq.InventoryItemID = saleItem.InventoryItemID
			}

			// Validate quantity
			if itemReq.QuantityReturned > saleItem.Quantity {
				return nil, ErrInsufficientStock
			}
			quantity := saleItem.Quantity
			originalQuantity = &quantity
			unitPrice = saleItem.UnitPrice
			if sale.DiscountAmount > 0 {
				subtotal := sale.Subtotal
				if subtotal <= 0 {
					subtotal = sale.TotalAmount + sale.DiscountAmount
				}
				itemGross := saleItem.UnitPrice * float64(saleItem.Quantity)
				if subtotal > 0 && itemGross > 0 {
					unitPrice = saleItem.UnitPrice - (sale.DiscountAmount * itemGross / subtotal / float64(saleItem.Quantity))
				}
			}
		} else {
			// Use provided unit price if no sale item
			unitPrice = itemReq.UnitPrice
		}

		item := CreateReturnItem(returnRecord.ID, itemReq, unitPrice)
		item.OriginalQuantity = originalQuantity
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
	if err := s.loadRefundState(ctx, returnRecord); err != nil {
		return nil, fmt.Errorf("failed to load return refund state: %w", err)
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
		if sale != nil {
			returnRecord.SaleInvoice = sale.InvoiceNumber
		}
	}

	return returnRecord.ToReturnResponse(items, customer, sale), nil
}

// ListReturns retrieves returns with pagination and filters
func (s *Service) ListReturns(ctx context.Context, req ReturnListRequest) ([]Return, int, error) {
	returns, total, err := s.repo.ListReturns(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	return returns, total, nil
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
	if req.TotalRefundAmount != nil {
		returnRecord.TotalRefundAmount = *req.TotalRefundAmount
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
		returnRecord.ItemConditionAfterReturn = normalizeReturnCondition(req.ItemConditionAfterReturn)
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
	var originalQuantity *int
	if req.SaleItemID != nil {
		saleItem, err := s.repo.GetSaleItemInfo(ctx, *req.SaleItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get sale item info: %w", err)
		}

		// Validate quantity
		if req.QuantityReturned > saleItem.Quantity {
			return nil, ErrInsufficientStock
		}
		quantity := saleItem.Quantity
		originalQuantity = &quantity

		unitPrice = saleItem.UnitPrice
	} else if req.UnitPrice > 0 {
		unitPrice = req.UnitPrice
	} else {
		return nil, fmt.Errorf("must provide either sale_item_id or unit_price")
	}

	item := CreateReturnItem(returnID, req, unitPrice)
	item.OriginalQuantity = originalQuantity
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
	parentReturn, err := s.repo.GetReturnByID(ctx, item.ReturnID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(parentReturn.Status, "COMPLETED") {
		return nil, ErrReturnAlreadyCompleted
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
	item, err := s.repo.GetReturnItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	parentReturn, err := s.repo.GetReturnByID(ctx, item.ReturnID)
	if err != nil {
		return err
	}
	if strings.EqualFold(parentReturn.Status, "COMPLETED") {
		return ErrReturnAlreadyCompleted
	}
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
	if err := s.loadRefundState(ctx, returnRecord); err != nil {
		return nil, fmt.Errorf("failed to load return refund state: %w", err)
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

	if err := s.processElectronicRefund(ctx, returnRecord, approvedBy); err != nil {
		return nil, err
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

	tx, err := s.repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin return completion: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := updateCompletedReturnTx(ctx, tx, s.repo.db.DriverName(), returnRecord); err != nil {
		return nil, err
	}
	for _, item := range items {
		status := returnInventoryStatus(returnRecord, item)
		if item.InventoryItemID != nil && *item.InventoryItemID != uuid.Nil {
			if err := updateReturnedInventoryTx(ctx, tx, s.repo.db.DriverName(), *item.InventoryItemID, status, returnRecord.ID); err != nil {
				return nil, err
			}
		} else if status == "AVAILABLE" && item.ProductID != nil && *item.ProductID != uuid.Nil {
			if err := restoreAggregateInventoryTx(ctx, tx, s.repo.db.DriverName(), *item.ProductID, item.QuantityReturned, returnRecord.ID); err != nil {
				return nil, err
			}
		}
	}
	if err := s.createSupplierReturnBridgeTx(ctx, tx, s.repo.db.DriverName(), returnRecord, items); err != nil {
		return nil, err
	}
	if err := addDebtAdjustmentLedgerEntryTx(ctx, tx, s.repo.db.DriverName(), returnRecord); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit return completion: %w", err)
	}
	committed = true
	dashboard.InvalidateDashboardCacheWithReason("return_completed")

	// The database trigger will handle the actual debt adjustment
	// This ensures consistency and prevents race conditions

	return s.GetReturn(ctx, id)
}

func (s *Service) processElectronicRefund(ctx context.Context, returnRecord *Return, approvedBy uuid.UUID) error {
	var transactionID, provider string
	queryErr := s.repo.db.QueryRowContext(ctx, `SELECT id, provider FROM payment_transactions WHERE sale_id = $1 AND status IN ('paid','partially_refunded') ORDER BY created_at DESC LIMIT 1`, returnRecord.SaleID).Scan(&transactionID, &provider)
	if queryErr == sql.ErrNoRows || strings.EqualFold(provider, "manual") {
		return nil
	}
	if queryErr != nil {
		return fmt.Errorf("failed to find electronic sale payment: %w", queryErr)
	}
	if returnRecord.RefundStatus == "refunded" || returnRecord.RefundStatus == "partially_refunded" {
		return nil
	}
	transactionUUID, err := uuid.Parse(transactionID)
	if err != nil {
		return fmt.Errorf("invalid electronic payment id: %w", err)
	}
	if s.refundProcessor == nil {
		_ = s.saveRefundState(ctx, returnRecord.ID, transactionUUID, nil, "failed", returnRecord.TotalRefundAmount, "electronic refund processor is not configured")
		return fmt.Errorf("electronic refund processor is not configured")
	}
	amountMinor := int64(returnRecord.TotalRefundAmount*100 + 0.5)
	if amountMinor <= 0 {
		return fmt.Errorf("return refund amount must be greater than zero")
	}
	if err := s.saveRefundState(ctx, returnRecord.ID, transactionUUID, nil, "processing", returnRecord.TotalRefundAmount, ""); err != nil {
		return fmt.Errorf("failed to record refund processing state: %w", err)
	}
	var userID *uuid.UUID
	if approvedBy != uuid.Nil {
		userID = &approvedBy
	}
	result, err := s.refundProcessor.RefundForReturn(ctx, returnRecord.SaleID, returnRecord.ID, amountMinor, userID)
	if err != nil {
		_ = s.saveRefundState(ctx, returnRecord.ID, transactionUUID, nil, "failed", returnRecord.TotalRefundAmount, err.Error())
		return fmt.Errorf("electronic refund failed: %w", err)
	}
	refundID, err := uuid.Parse(result.RefundID)
	if err != nil {
		return fmt.Errorf("invalid refund id returned by processor: %w", err)
	}
	if err := s.saveRefundState(ctx, returnRecord.ID, transactionUUID, &refundID, result.Status, returnRecord.TotalRefundAmount, ""); err != nil {
		return fmt.Errorf("failed to finalize refund state: %w", err)
	}
	return nil
}

func (s *Service) saveRefundState(ctx context.Context, returnID, transactionID uuid.UUID, refundID *uuid.UUID, status string, amount float64, errorMessage string) error {
	var refundArg, returnArg, transactionArg interface{} = nil, returnID, transactionID
	if refundID != nil {
		refundArg = *refundID
	}
	amountMinor := int64(amount*100 + 0.5)
	query := `INSERT INTO return_payment_refunds (id, return_id, payment_transaction_id, payment_refund_id, status, amount_minor, idempotency_key, error_message, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9) ON CONFLICT (return_id) DO UPDATE SET payment_refund_id=EXCLUDED.payment_refund_id, status=EXCLUDED.status, amount_minor=EXCLUDED.amount_minor, error_message=EXCLUDED.error_message, updated_at=EXCLUDED.updated_at`
	args := []interface{}{uuid.New(), returnArg, transactionArg, refundArg, status, amountMinor, "return-refund:" + returnID.String(), nullableString(errorMessage), time.Now().UTC()}
	if dbutil.IsSQLite(s.repo.db) {
		query = `INSERT INTO return_payment_refunds (id, return_id, payment_transaction_id, payment_refund_id, status, amount_minor, idempotency_key, error_message, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?) ON CONFLICT(return_id) DO UPDATE SET payment_refund_id=excluded.payment_refund_id, status=excluded.status, amount_minor=excluded.amount_minor, error_message=excluded.error_message, updated_at=excluded.updated_at`
		for index, arg := range args {
			if value, ok := arg.(uuid.UUID); ok {
				args[index] = value.String()
			}
		}
		args[8] = time.Now().UTC().Format(time.RFC3339Nano)
		args = append(args, args[8])
	}
	_, err := s.repo.db.ExecContext(ctx, query, args...)
	return err
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func updateCompletedReturnTx(ctx context.Context, tx *sqlx.Tx, driver string, returnRecord *Return) error {
	if driver == "sqlite" {
		_, err := tx.ExecContext(ctx, `UPDATE returns SET status=?, processed_by=?, approved_by=?, approved_at=?, debt_id=?, debt_adjustment=?, refund_date=?, updated_at=? WHERE id=?`,
			returnRecord.Status, idArgPtr(returnRecord.ProcessedBy), idArgPtr(returnRecord.ApprovedBy), returnRecord.ApprovedAt, idArgPtr(returnRecord.DebtID), returnRecord.DebtAdjustment, returnRecord.RefundDate, returnRecord.UpdatedAt.Format(time.RFC3339Nano), returnRecord.ID.String())
		if err != nil {
			return fmt.Errorf("failed to complete return: %w", err)
		}
		return nil
	}
	_, err := tx.ExecContext(ctx, `UPDATE returns SET status=$1, processed_by=$2, approved_by=$3, approved_at=$4, debt_id=$5, debt_adjustment=$6, refund_date=$7, updated_at=$8 WHERE id=$9`,
		returnRecord.Status, returnRecord.ProcessedBy, returnRecord.ApprovedBy, returnRecord.ApprovedAt, returnRecord.DebtID, returnRecord.DebtAdjustment, returnRecord.RefundDate, returnRecord.UpdatedAt, returnRecord.ID)
	if err != nil {
		return fmt.Errorf("failed to complete return: %w", err)
	}
	return nil
}

func updateReturnedInventoryTx(ctx context.Context, tx *sqlx.Tx, driver string, itemID uuid.UUID, status string, returnID uuid.UUID) error {
	var result sql.Result
	var err error
	if driver == "sqlite" {
		result, err = tx.ExecContext(ctx, `UPDATE inventory_items SET status=?, sold_at=NULL, updated_at=? WHERE id=?`, status, time.Now().UTC().Format(time.RFC3339Nano), itemID.String())
	} else {
		result, err = tx.ExecContext(ctx, `UPDATE inventory_items SET status=$1, sold_at=NULL, updated_at=NOW() WHERE id=$2`, status, itemID)
	}
	if err != nil {
		return fmt.Errorf("failed to update returned inventory item: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("inventory item %s was not found", itemID)
	}
	if strings.EqualFold(status, "AVAILABLE") {
		var afterQuantity int
		if err := tx.GetContext(ctx, &afterQuantity, `SELECT COUNT(*) FROM inventory_items WHERE product_id=(SELECT product_id FROM inventory_items WHERE id=$1) AND UPPER(TRIM(COALESCE(status, ''))) = 'AVAILABLE'`, itemID); err != nil {
			return fmt.Errorf("failed to read returned item availability: %w", err)
		}
		beforeQuantity := afterQuantity - 1
		if driver == "sqlite" {
			_, err = tx.ExecContext(ctx, `INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_at) SELECT ?, ?, product_id, 'RETURN', 1, ?, ?, 'return', ?, 'Customer return restock', CURRENT_TIMESTAMP FROM inventory_items WHERE id=?`, uuid.New().String(), itemID.String(), beforeQuantity, afterQuantity, returnID.String(), itemID.String())
		} else {
			_, err = tx.ExecContext(ctx, `INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_at) SELECT uuid_generate_v4(), $1, product_id, 'RETURN', 1, $2, $3, 'return', $4, 'Customer return restock', NOW() FROM inventory_items WHERE id=$1`, itemID, beforeQuantity, afterQuantity, returnID)
		}
		if err != nil {
			return fmt.Errorf("failed to record returned item movement: %w", err)
		}
		var quantityResult sql.Result
		if driver == "sqlite" {
			quantityResult, err = tx.ExecContext(ctx, `UPDATE inventory SET quantity=quantity+1, updated_at=CURRENT_TIMESTAMP WHERE product_id=(SELECT product_id FROM inventory_items WHERE id=?)`, itemID.String())
		} else {
			quantityResult, err = tx.ExecContext(ctx, `UPDATE inventory SET quantity=quantity+1, updated_at=NOW() WHERE product_id=(SELECT product_id FROM inventory_items WHERE id=$1)`, itemID)
		}
		if err != nil {
			return fmt.Errorf("failed to restore returned inventory quantity: %w", err)
		}
		if affected, _ := quantityResult.RowsAffected(); affected == 0 {
			if driver == "sqlite" {
				_, err = tx.ExecContext(ctx, `INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) SELECT ?, product_id, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP FROM inventory_items WHERE id=?`, uuid.New().String(), itemID.String())
			} else {
				_, err = tx.ExecContext(ctx, `INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) SELECT uuid_generate_v4(), product_id, 1, NOW(), NOW() FROM inventory_items WHERE id=$1`, itemID)
			}
			if err != nil {
				return fmt.Errorf("failed to create restored inventory quantity: %w", err)
			}
		}
	}
	return nil
}

func restoreAggregateInventoryTx(ctx context.Context, tx *sqlx.Tx, driver string, productID uuid.UUID, quantity int, returnID uuid.UUID) error {
	if quantity <= 0 {
		return nil
	}
	var result sql.Result
	var err error
	if driver == "sqlite" {
		result, err = tx.ExecContext(ctx, `UPDATE inventory SET quantity=COALESCE(quantity, 0)+?, updated_at=CURRENT_TIMESTAMP WHERE product_id=?`, quantity, productID.String())
	} else {
		result, err = tx.ExecContext(ctx, `UPDATE inventory SET quantity=COALESCE(quantity, 0)+$1, updated_at=NOW() WHERE product_id=$2`, quantity, productID)
	}
	if err != nil {
		return fmt.Errorf("failed to restore aggregate inventory quantity: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		if driver == "sqlite" {
			_, err = tx.ExecContext(ctx, `INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, uuid.New().String(), productID.String(), quantity)
		} else {
			_, err = tx.ExecContext(ctx, `INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) VALUES (uuid_generate_v4(), $1, $2, NOW(), NOW())`, productID, quantity)
		}
		if err != nil {
			return fmt.Errorf("failed to create restored aggregate inventory: %w", err)
		}
	}
	if driver == "sqlite" {
		_, err = tx.ExecContext(ctx, `INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_at) SELECT ?, NULL, ?, 'RETURN', ?, quantity-?, quantity, 'return', ?, 'Customer return restock', CURRENT_TIMESTAMP FROM inventory WHERE product_id=?`, uuid.New().String(), productID.String(), quantity, quantity, returnID.String(), productID.String())
	} else {
		_, err = tx.ExecContext(ctx, `INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_at) SELECT uuid_generate_v4(), NULL, $1, 'RETURN', $2, quantity-$2, quantity, 'return', $3, 'Customer return restock', NOW() FROM inventory WHERE product_id=$1`, productID, quantity, returnID)
	}
	if err != nil {
		return fmt.Errorf("failed to record aggregate return movement: %w", err)
	}
	return nil
}

func nullableUUIDArg(id uuid.UUID) interface{} {
	if id == uuid.Nil {
		return nil
	}
	return id.String()
}

func nullableUUIDPtrArg(id *uuid.UUID) interface{} {
	if id == nil {
		return nil
	}
	return nullableUUIDArg(*id)
}

func addDebtAdjustmentLedgerEntryTx(ctx context.Context, tx *sqlx.Tx, driver string, returnRecord *Return) error {
	if returnRecord.CustomerID == uuid.Nil || returnRecord.RefundMethod != "DEBT_ADJUSTMENT" {
		return nil
	}
	var existing int
	if err := tx.GetContext(ctx, &existing, `SELECT COUNT(*) FROM customer_ledger WHERE customer_id = $1 AND reference_id = $2 AND type = 'credit'`, returnRecord.CustomerID, returnRecord.ID); err != nil {
		return fmt.Errorf("check debt adjustment idempotency: %w", err)
	}
	if existing > 0 {
		return nil
	}
	adjustment := returnRecord.TotalRefundAmount
	if adjustment < 0 {
		adjustment = 0
	}
	applied := adjustment
	var currentDebt float64
	if returnRecord.DebtID != nil {
		if err := tx.GetContext(ctx, &currentDebt, `SELECT COALESCE(remaining_amount, 0) FROM debts WHERE id = $1`, returnRecord.DebtID); err != nil {
			return fmt.Errorf("read debt for adjustment: %w", err)
		}
		if applied > currentDebt {
			applied = currentDebt
		}
		remaining := currentDebt - applied
		if driver == "sqlite" {
			if _, err := tx.ExecContext(ctx, `UPDATE debts SET paid_amount=MIN(amount, COALESCE(paid_amount,0)+?), remaining_amount=MAX(0, COALESCE(remaining_amount,0)-?), status=CASE WHEN ? <= 0 THEN 'paid' ELSE status END, updated_at=CURRENT_TIMESTAMP WHERE id=?`, applied, applied, remaining, returnRecord.DebtID.String()); err != nil {
				return fmt.Errorf("apply debt adjustment: %w", err)
			}
		} else if _, err := tx.ExecContext(ctx, `UPDATE debts SET paid_amount=LEAST(amount, COALESCE(paid_amount,0)+$1), remaining_amount=GREATEST(0, COALESCE(remaining_amount,0)-$2), status=CASE WHEN $3 <= 0 THEN 'paid' ELSE status END, updated_at=NOW() WHERE id=$4`, applied, applied, remaining, returnRecord.DebtID); err != nil {
			return fmt.Errorf("apply debt adjustment: %w", err)
		}
	}
	returnRecord.DebtAdjustment = applied
	returnRecord.CustomerCredit = adjustment - applied
	if driver == "sqlite" {
		if _, err := tx.ExecContext(ctx, `UPDATE returns SET debt_adjustment=?, customer_credit=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, applied, returnRecord.CustomerCredit, returnRecord.ID.String()); err != nil {
			return fmt.Errorf("store debt adjustment result: %w", err)
		}
	} else if _, err := tx.ExecContext(ctx, `UPDATE returns SET debt_adjustment=$1, customer_credit=$2, updated_at=NOW() WHERE id=$3`, applied, returnRecord.CustomerCredit, returnRecord.ID); err != nil {
		return fmt.Errorf("store debt adjustment result: %w", err)
	}
	var balance float64
	_ = tx.GetContext(ctx, &balance, `SELECT COALESCE(balance, 0) FROM customer_ledger WHERE customer_id = $1 ORDER BY created_at DESC LIMIT 1`, returnRecord.CustomerID)
	newBalance := balance - adjustment
	if driver == "sqlite" {
		if _, err := tx.ExecContext(ctx, `INSERT INTO customer_ledger (id, customer_id, type, amount, balance, description, reference_id, created_at) VALUES (?, ?, 'credit', ?, ?, ?, ?, CURRENT_TIMESTAMP)`, uuid.New().String(), returnRecord.CustomerID.String(), adjustment, newBalance, "Customer return: "+returnRecord.ReturnNumber, returnRecord.ID.String()); err != nil {
			return fmt.Errorf("record debt adjustment ledger: %w", err)
		}
		_, err := tx.ExecContext(ctx, `UPDATE customers SET current_balance=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, newBalance, returnRecord.CustomerID.String())
		if err != nil {
			return fmt.Errorf("update customer credit: %w", err)
		}
		return nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO customer_ledger (id, customer_id, type, amount, balance, description, reference_id, created_at) VALUES ($1,$2,'credit',$3,$4,$5,$6,NOW())`, uuid.New(), returnRecord.CustomerID, adjustment, newBalance, "Customer return: "+returnRecord.ReturnNumber, returnRecord.ID); err != nil {
		return fmt.Errorf("record debt adjustment ledger: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE customers SET current_balance=$1, updated_at=NOW() WHERE id=$2`, newBalance, returnRecord.CustomerID); err != nil {
		return fmt.Errorf("update customer credit: %w", err)
	}
	return nil
}

func (s *Service) createSupplierReturnBridgeTx(ctx context.Context, tx *sqlx.Tx, driver string, returnRecord *Return, items []ReturnItem) error {
	for _, item := range items {
		resolution := strings.ToUpper(strings.TrimSpace(item.Resolution))
		condition := strings.ToUpper(strings.TrimSpace(returnRecord.ItemConditionAfterReturn))
		if condition != "RETURN_TO_SUPPLIER" && condition != "SUPPLIER_RETURN" && resolution != "RETURN_TO_SUPPLIER" && resolution != "SUPPLIER_RETURN" {
			continue
		}

		var purchaseItemID, purchaseID, supplierID string
		var purchaseCost float64
		placeholder := "?"
		if driver != "sqlite" {
			placeholder = "$1"
		}
		query := fmt.Sprintf(`SELECT pi.id, pi.purchase_id, p.supplier_id, COALESCE(pi.unit_price, 0)
			FROM sale_items si
			JOIN inventory_items ii ON ii.id = %s AND ii.id = si.inventory_item_id
			JOIN purchase_items pi ON pi.product_id = si.product_id
			JOIN purchases p ON p.id = pi.purchase_id
			WHERE si.id = %s AND si.product_id = %s
			AND (NULLIF(pi.serial_number, '') = NULLIF(ii.serial_number, '')
			 OR NULLIF(pi.barcode, '') = NULLIF(ii.barcode, '')
			 OR ii.item_code LIKE 'ITM-' || substr(replace(pi.purchase_id::text, '-', ''), 1, 8) || '-%%')
			ORDER BY (NULLIF(pi.serial_number, '') = NULLIF(ii.serial_number, '')) DESC,
			         (NULLIF(pi.barcode, '') = NULLIF(ii.barcode, '')) DESC, pi.created_at DESC LIMIT 1`, placeholder, func() string {
			if driver == "sqlite" {
				return "?"
			}
			return "$2"
		}(), func() string {
			if driver == "sqlite" {
				return "?"
			}
			return "$3"
		}())
		args := []interface{}{item.InventoryItemID, item.SaleItemID, item.ProductID}
		if driver == "sqlite" {
			query = strings.Replace(query, "pi.purchase_id::text", "pi.purchase_id", 1)
			args = []interface{}{nullableUUIDPtrArg(item.InventoryItemID), nullableUUIDPtrArg(item.SaleItemID), item.ProductID.String()}
		}
		err := tx.QueryRowContext(ctx, query, args...).Scan(&purchaseItemID, &purchaseID, &supplierID, &purchaseCost)
		resolved := err == nil
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("resolve supplier return source: %w", err)
		}

		requestID := uuid.New().String()
		returnNumber := "SRET-" + strings.ToUpper(strings.ReplaceAll(requestID[:10], "-", ""))
		status, sourceStatus := "PENDING", "RESOLVED"
		var purchaseArg, supplierArg interface{}
		if resolved {
			purchaseArg, supplierArg = purchaseID, supplierID
		} else {
			status, sourceStatus = "NEEDS_SOURCE_DATA", "NEEDS_SOURCE_DATA"
			purchaseArg, supplierArg = nil, nil
		}
		var existingID string
		selectQuery := `SELECT id FROM supplier_returns WHERE customer_return_id = ? LIMIT 1`
		insertQuery := `INSERT INTO supplier_returns (id, customer_return_id, sale_id, purchase_id, supplier_id, return_number, status, source_status, reason, refund_amount, notes, return_reason, return_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?) ON CONFLICT DO NOTHING`
		if driver != "sqlite" {
			selectQuery = `SELECT id FROM supplier_returns WHERE customer_return_id = $1 LIMIT 1`
			insertQuery = `INSERT INTO supplier_returns (id, customer_return_id, sale_id, purchase_id, supplier_id, return_number, status, source_status, reason, refund_amount, notes, return_reason, return_date, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,0,$10,$11,$12,NOW(),NOW()) ON CONFLICT DO NOTHING`
		}
		if err := tx.QueryRowContext(ctx, selectQuery, returnRecord.ID.String()).Scan(&existingID); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("load supplier return request: %w", err)
		}
		if existingID == "" {
			now := time.Now().UTC().Format(time.RFC3339Nano)
			args := []interface{}{requestID, returnRecord.ID.String(), nullableUUIDArg(returnRecord.SaleID), purchaseArg, supplierArg, returnNumber, status, sourceStatus, returnRecord.Reason, fmt.Sprintf("Auto-created from customer return %s", returnRecord.ReturnNumber), returnRecord.Reason, returnRecord.ReturnDate, now, now}
			if driver != "sqlite" {
				args = []interface{}{requestID, returnRecord.ID, returnRecord.SaleID, purchaseArg, supplierArg, returnNumber, status, sourceStatus, returnRecord.Reason, fmt.Sprintf("Auto-created from customer return %s", returnRecord.ReturnNumber), returnRecord.Reason, returnRecord.ReturnDate}
			}
			if _, err := tx.ExecContext(ctx, insertQuery, args...); err != nil {
				return fmt.Errorf("create supplier return request: %w", err)
			}
			if err := tx.QueryRowContext(ctx, selectQuery, returnRecord.ID.String()).Scan(&existingID); err != nil {
				return fmt.Errorf("reload supplier return request: %w", err)
			}
		}
		if !resolved {
			continue
		}
		var activeItemOwner string
		activeItemQuery := `SELECT COALESCE(sr.customer_return_id, '') FROM supplier_return_items sri JOIN supplier_returns sr ON sr.id = sri.supplier_return_id WHERE sri.inventory_item_id = ? AND sr.status IN ('PENDING','SHIPPED','RECEIVED') LIMIT 1`
		activeItemArg := nullableUUIDPtrArg(item.InventoryItemID)
		if driver != "sqlite" {
			activeItemQuery = `SELECT COALESCE(sr.customer_return_id::text, '') FROM supplier_return_items sri JOIN supplier_returns sr ON sr.id = sri.supplier_return_id WHERE sri.inventory_item_id = $1 AND sr.status IN ('PENDING','SHIPPED','RECEIVED') LIMIT 1`
			activeItemArg = item.InventoryItemID
		}
		if err := tx.QueryRowContext(ctx, activeItemQuery, activeItemArg).Scan(&activeItemOwner); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("check active supplier return item: %w", err)
		}
		if activeItemOwner != "" && activeItemOwner != returnRecord.ID.String() {
			return fmt.Errorf("inventory item %s already has an active supplier return request", item.InventoryItemID)
		}

		itemQuery := `INSERT INTO supplier_return_items (id, customer_return_id, sale_id, sale_item_id, inventory_item_id, supplier_return_id, purchase_item_id, product_id, quantity, unit_cost, barcode, serial_number, purchase_cost, return_reason, return_date, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING`
		itemArgs := []interface{}{uuid.New().String(), returnRecord.ID.String(), nullableUUIDArg(returnRecord.SaleID), nullableUUIDPtrArg(item.SaleItemID), nullableUUIDPtrArg(item.InventoryItemID), existingID, purchaseItemID, item.ProductID.String(), item.QuantityReturned, purchaseCost, item.Barcode, item.SerialNumber, purchaseCost, returnRecord.Reason, returnRecord.ReturnDate}
		if driver != "sqlite" {
			itemQuery = `INSERT INTO supplier_return_items (id, customer_return_id, sale_id, sale_item_id, inventory_item_id, supplier_return_id, purchase_item_id, product_id, quantity, unit_cost, barcode, serial_number, purchase_cost, return_reason, return_date, created_at) VALUES (uuid_generate_v4(), $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW()) ON CONFLICT DO NOTHING`
			itemArgs = []interface{}{returnRecord.ID, returnRecord.SaleID, item.SaleItemID, item.InventoryItemID, existingID, purchaseItemID, item.ProductID, item.QuantityReturned, purchaseCost, item.Barcode, item.SerialNumber, purchaseCost, returnRecord.Reason, returnRecord.ReturnDate}
		}
		if _, err := tx.ExecContext(ctx, itemQuery, itemArgs...); err != nil {
			return fmt.Errorf("create supplier return item: %w", err)
		}
	}
	return nil
}

func returnInventoryStatus(returnRecord *Return, item ReturnItem) string {
	resolution := strings.ToUpper(strings.TrimSpace(item.Resolution))
	if resolution == "" {
		resolution = strings.ToUpper(strings.TrimSpace(returnRecord.ItemConditionAfterReturn))
	}
	switch resolution {
	case "RESTOCK", "READY_FOR_SALE", "SELLABLE":
		return "AVAILABLE"
	case "REPAIR", "NEEDS_REPAIR":
		return "IN_REPAIR"
	case "SUPPLIER_RETURN", "RETURN_TO_SUPPLIER":
		return "RETURNED"
	case "WRITE_OFF", "NOT_FOR_SALE":
		return "ARCHIVED"
	case "PARTS":
		return "FOR_PARTS"
	default:
		return "DAMAGED"
	}
}

func (s *Service) createSupplierReturnBridge(ctx context.Context, returnRecord *Return, items []ReturnItem) error {
	needsSupplierBridge := strings.EqualFold(returnRecord.ItemConditionAfterReturn, "RETURN_TO_SUPPLIER") ||
		strings.EqualFold(returnRecord.ItemConditionAfterReturn, "SUPPLIER_RETURN")
	if !needsSupplierBridge {
		for _, item := range items {
			if strings.EqualFold(item.Resolution, "SUPPLIER_RETURN") || strings.EqualFold(item.Resolution, "RETURN_TO_SUPPLIER") {
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
		isSupplierReturnCondition := strings.EqualFold(returnRecord.ItemConditionAfterReturn, "RETURN_TO_SUPPLIER") ||
			strings.EqualFold(returnRecord.ItemConditionAfterReturn, "SUPPLIER_RETURN") ||
			strings.EqualFold(item.Resolution, "SUPPLIER_RETURN") ||
			strings.EqualFold(item.Resolution, "RETURN_TO_SUPPLIER")
		if !isSupplierReturnCondition {
			continue
		}

		purchaseItemID, purchaseID, supplierID, err := s.findPurchaseItemForReturnItem(ctx, item)
		if err != nil {
			return err
		}
		if purchaseItemID == uuid.Nil || purchaseID == uuid.Nil || supplierID == uuid.Nil {
			if err := s.ensureSupplierReturnBridgeEntry(ctx, returnRecord, uuid.Nil, uuid.Nil, uuid.Nil, item); err != nil {
				return err
			}
			continue
		}

		if err := s.ensureSupplierReturnBridgeEntry(ctx, returnRecord, purchaseID, supplierID, purchaseItemID, item); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) findPurchaseItemForReturnItem(ctx context.Context, item ReturnItem) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	if item.ProductID == nil || *item.ProductID == uuid.Nil || item.InventoryItemID == nil || *item.InventoryItemID == uuid.Nil || item.SaleItemID == nil || *item.SaleItemID == uuid.Nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, nil
	}

	var candidate struct {
		PurchaseItemID string `db:"purchase_item_id"`
		PurchaseID     string `db:"purchase_id"`
		SupplierID     string `db:"supplier_id"`
	}

	query := `
		SELECT pi.id AS purchase_item_id, pi.purchase_id, p.supplier_id
		FROM sale_items si
		JOIN inventory_items ii ON ii.id = ? AND ii.id = si.inventory_item_id
		JOIN purchase_items pi ON pi.product_id = si.product_id
		JOIN purchases p ON p.id = pi.purchase_id
		WHERE si.id = ? AND si.product_id = ?
		  AND (NULLIF(pi.serial_number, '') = NULLIF(ii.serial_number, '')
		       OR NULLIF(pi.barcode, '') = NULLIF(ii.barcode, '')
		       OR ii.item_code LIKE 'ITM-' || substr(replace(pi.purchase_id, '-', ''), 1, 8) || '-%')
		ORDER BY (NULLIF(pi.serial_number, '') = NULLIF(ii.serial_number, '')) DESC,
		         (NULLIF(pi.barcode, '') = NULLIF(ii.barcode, '')) DESC,
		         pi.created_at DESC
		LIMIT 1
	`
	inventoryID, saleItemID := uuid.Nil, uuid.Nil
	if item.InventoryItemID != nil {
		inventoryID = *item.InventoryItemID
	}
	if item.SaleItemID != nil {
		saleItemID = *item.SaleItemID
	}
	if err := s.repo.db.GetContext(ctx, &candidate, query, inventoryID.String(), saleItemID.String(), item.ProductID.String()); err != nil {
		if err == sql.ErrNoRows {
			// Legacy callers may invoke the bridge helper before source identity
			// has been populated. The live completion transaction never uses this
			// fallback; it persists NEEDS_SOURCE_DATA instead.
			return uuid.Nil, uuid.Nil, uuid.Nil, nil
		}
		if err != sql.ErrNoRows {
			return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("failed to find purchase item for supplier return bridge: %w", err)
		}
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
	if returnRecord == nil {
		return fmt.Errorf("customer return is required for supplier return bridge")
	}

	storeSupplierReturnBridgeLink := func(supplierReturnID string) error {
		var returnExists int
		if err := s.repo.db.GetContext(ctx, &returnExists, `SELECT COUNT(*) FROM returns WHERE id = ?`, returnRecord.ID.String()); err != nil {
			return fmt.Errorf("failed to check customer return record before supplier bridge update: %w", err)
		}
		if returnExists == 0 {
			return nil
		}

		var count int
		if err := s.repo.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM supplier_returns WHERE customer_return_id = ? AND status IN ('PENDING', 'SHIPPED', 'RECEIVED')`, returnRecord.ID.String()); err != nil {
			return fmt.Errorf("failed to check active supplier return bridge: %w", err)
		}
		if count > 0 {
			return nil
		}

		updatedAt := time.Now().UTC().Format(time.RFC3339Nano)
		if _, err := s.repo.db.ExecContext(ctx, `
			UPDATE supplier_returns
			SET customer_return_id = ?, sale_id = ?, return_reason = COALESCE(NULLIF(?, ''), reason), return_date = COALESCE(?, created_at), updated_at = ?
			WHERE id = ?
		`, returnRecord.ID.String(), returnRecord.SaleID.String(), returnRecord.Reason, returnRecord.ReturnDate.UTC().Format(time.RFC3339Nano), updatedAt, supplierReturnID); err != nil {
			return fmt.Errorf("failed to update supplier return source links: %w", err)
		}
		return nil
	}

	var supplierReturnID string
	if err := s.repo.db.GetContext(ctx, &supplierReturnID,
		`SELECT id FROM supplier_returns WHERE customer_return_id = ? ORDER BY created_at DESC LIMIT 1`,
		returnRecord.ID.String()); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to load existing supplier return bridge: %w", err)
	}

	if supplierReturnID == "" {
		supplierReturnID = uuid.New().String()
		returnNumber := "SRET-" + strings.ToUpper(strings.ReplaceAll(uuid.New().String()[:10], "-", ""))
		now := time.Now().UTC().Format(time.RFC3339Nano)
		status := "PENDING"
		sourceStatus := "RESOLVED"
		var purchaseArg interface{} = purchaseID.String()
		var supplierArg interface{} = supplierID.String()
		if purchaseID == uuid.Nil || supplierID == uuid.Nil || item.ProductID == nil || *item.ProductID == uuid.Nil {
			status = "NEEDS_SOURCE_DATA"
			sourceStatus = "NEEDS_SOURCE_DATA"
			purchaseArg = nil
			supplierArg = nil
		}
		var createdBy interface{}
		if returnRecord.CreatedBy != nil && *returnRecord.CreatedBy != uuid.Nil {
			createdBy = returnRecord.CreatedBy.String()
		} else if returnRecord.ProcessedBy != nil && *returnRecord.ProcessedBy != uuid.Nil {
			createdBy = returnRecord.ProcessedBy.String()
		}
		customerReturnArg := nullableUUIDArg(returnRecord.ID)
		saleArg := nullableUUIDArg(returnRecord.SaleID)
		_, err := s.repo.db.ExecContext(ctx, `
			INSERT INTO supplier_returns (id, customer_return_id, sale_id, purchase_id, supplier_id, return_number, status, source_status, reason, refund_amount, notes, created_by, return_reason, return_date, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?, ?)
		`, supplierReturnID, customerReturnArg, saleArg, purchaseArg, supplierArg, returnNumber, status, sourceStatus, fmt.Sprintf("Customer return %s", returnRecord.ReturnNumber), fmt.Sprintf("Auto-created from customer return %s", returnRecord.ReturnNumber), createdBy, returnRecord.Reason, returnRecord.ReturnDate.UTC().Format(time.RFC3339Nano), now, now)
		if err != nil {
			return fmt.Errorf("failed to create supplier return bridge: %w", err)
		}
		if err := storeSupplierReturnBridgeLink(supplierReturnID); err != nil {
			return err
		}
		if purchaseID == uuid.Nil || supplierID == uuid.Nil || item.ProductID == nil || *item.ProductID == uuid.Nil {
			return nil
		}
	}

	if err := storeSupplierReturnBridgeLink(supplierReturnID); err != nil {
		return err
	}

	var existingSupplierReturnCount int
	if err := s.repo.db.GetContext(ctx, &existingSupplierReturnCount,
		`SELECT COUNT(*) FROM supplier_returns WHERE customer_return_id = ? AND purchase_id = ? AND status IN ('PENDING','SHIPPED','RECEIVED')`,
		returnRecord.ID.String(), purchaseID.String()); err != nil {
		return fmt.Errorf("failed to validate customer return supplier bridge: %w", err)
	}
	if existingSupplierReturnCount > 1 {
		return fmt.Errorf("duplicate active supplier return request already exists for customer return %s", returnRecord.ReturnNumber)
	}

	if purchaseItemID == uuid.Nil || purchaseID == uuid.Nil || supplierID == uuid.Nil || item.ProductID == nil || *item.ProductID == uuid.Nil {
		return nil
	}

	var existingSupplierItemCount int
	if err := s.repo.db.GetContext(ctx, &existingSupplierItemCount,
		`SELECT COUNT(*) FROM supplier_return_items sri JOIN supplier_returns sr ON sr.id = sri.supplier_return_id WHERE sri.supplier_return_id = ? AND sri.purchase_item_id = ? AND sr.status IN ('PENDING','SHIPPED','RECEIVED','COMPLETED')`,
		supplierReturnID, purchaseItemID.String(),
	); err != nil {
		return fmt.Errorf("failed to check supplier return bridge item: %w", err)
	}
	if existingSupplierItemCount > 0 {
		return nil
	}

	if item.InventoryItemID != nil && *item.InventoryItemID != uuid.Nil {
		var existingItemForCustomerReturn int
		if err := s.repo.db.GetContext(ctx, &existingItemForCustomerReturn,
			`SELECT COUNT(*) FROM supplier_return_items sri JOIN supplier_returns sr ON sr.id = sri.supplier_return_id WHERE sri.customer_return_id = ? AND sri.inventory_item_id = ? AND sr.status IN ('PENDING','SHIPPED','RECEIVED')`,
			returnRecord.ID.String(), item.InventoryItemID.String()); err != nil {
			return fmt.Errorf("failed to check duplicate inventory item supplier return: %w", err)
		}
		if existingItemForCustomerReturn > 0 {
			return nil
		}
	}

	var returnRowExists int
	if err := s.repo.db.GetContext(ctx, &returnRowExists, `SELECT COUNT(*) FROM returns WHERE id = ?`, returnRecord.ID.String()); err != nil {
		return fmt.Errorf("failed to validate customer return exists before bridge item insert: %w", err)
	}
	if returnRowExists == 0 {
		return nil
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	barcode := item.Barcode
	serialNumber := item.SerialNumber
	var purchaseCost float64
	if item.OriginalCost != nil {
		purchaseCost = *item.OriginalCost
	}
	if purchaseCost == 0 && item.InventoryItemID != nil && *item.InventoryItemID != uuid.Nil {
		if err := s.repo.db.GetContext(ctx, &purchaseCost, `SELECT COALESCE(purchase_cost, 0) FROM inventory_items WHERE id = ?`, item.InventoryItemID.String()); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to read inventory item purchase cost: %w", err)
		}
	}
	if purchaseCost == 0 {
		if err := s.repo.db.GetContext(ctx, &purchaseCost, `SELECT COALESCE(unit_price, 0) FROM purchase_items WHERE id = ?`, purchaseItemID.String()); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to read purchase item cost: %w", err)
		}
	}
	if barcode == "" && item.InventoryItemID != nil && *item.InventoryItemID != uuid.Nil {
		if err := s.repo.db.GetContext(ctx, &barcode, `SELECT COALESCE(barcode, '') FROM inventory_items WHERE id = ?`, item.InventoryItemID.String()); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to read inventory item barcode: %w", err)
		}
	}
	if serialNumber == "" && item.InventoryItemID != nil && *item.InventoryItemID != uuid.Nil {
		if err := s.repo.db.GetContext(ctx, &serialNumber, `SELECT COALESCE(serial_number, '') FROM inventory_items WHERE id = ?`, item.InventoryItemID.String()); err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to read inventory item serial number: %w", err)
		}
	}
	if _, err := s.repo.db.ExecContext(ctx, `
		INSERT INTO supplier_return_items (id, customer_return_id, sale_id, sale_item_id, inventory_item_id, supplier_return_id, purchase_item_id, product_id, quantity, unit_cost, barcode, serial_number, purchase_cost, return_reason, return_date, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, uuid.New().String(), returnRecord.ID.String(), returnRecord.SaleID.String(), coalesceUUIDString(item.SaleItemID), coalesceUUIDString(item.InventoryItemID), supplierReturnID, purchaseItemID.String(), item.ProductID.String(), item.QuantityReturned, item.UnitPrice, barcode, serialNumber, purchaseCost, returnRecord.Reason, returnRecord.ReturnDate.UTC().Format(time.RFC3339Nano), now); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique constraint") || strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return nil
		}
		return fmt.Errorf("failed to add supplier return bridge item: %w", err)
	}
	return nil
}

func coalesceUUIDString(value *uuid.UUID) string {
	if value == nil || *value == uuid.Nil {
		return ""
	}
	return value.String()
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

// ReverseReturn cancels the original return instead of creating a new return record.
func (s *Service) ReverseReturn(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if returnRecord.Status == "REVERSED" || returnRecord.Status == "CANCELLED" {
		return nil, ErrInvalidReturnStatus
	}

	returnRecord.Status = "CANCELLED"
	returnRecord.ProcessedBy = &userID
	returnRecord.InternalNotes = "Cancelled without creating a new return record."
	returnRecord.UpdatedAt = time.Now()
	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, fmt.Errorf("failed to cancel original return: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("return_reversed")

	return s.GetReturn(ctx, id)
}

func normalizeReversalReturnType(value string) string {
	switch strings.ToUpper(value) {
	case "FULL", "PARTIAL", "QUANTITY_PARTIAL":
		return strings.ToUpper(value)
	default:
		return "PARTIAL"
	}
}

func normalizeReversalRefundMethod(value string) string {
	switch strings.ToUpper(value) {
	case "CASH", "CREDIT", "DEBT_ADJUSTMENT", "EXCHANGE", "BANK_TRANSFER", "STORE_CREDIT":
		return strings.ToUpper(value)
	default:
		return "CASH"
	}
}

func normalizeReversalReason(value string) string {
	switch strings.ToUpper(value) {
	case "DEFECTIVE", "WRONG_ITEM", "COMPATIBILITY_ISSUE", "CUSTOMER_CHANGED_MIND", "DAMAGED", "WARRANTY", "INCORRECT_SPECIFICATION", "OTHER":
		return strings.ToUpper(value)
	default:
		return "OTHER"
	}
}

func normalizeReversalCondition(value string) string {
	switch strings.ToUpper(value) {
	case "READY_FOR_SALE", "SELLABLE":
		return "SELLABLE"
	case "NOT_FOR_SALE", "WRITE_OFF":
		return "WRITE_OFF"
	case "RETURN_TO_SUPPLIER", "SUPPLIER_RETURN":
		return "SUPPLIER_RETURN"
	case "NEEDS_INSPECTION", "NEEDS_REPAIR", "DAMAGED", "USED", "REFURBISHED", "PARTS":
		return strings.ToUpper(value)
	default:
		return "SELLABLE"
	}
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
	returnRecord, items, err := s.repo.GetReturnWithItems(ctx, returnID)
	if err != nil {
		return nil, nil, err
	}

	if returnRecord.SaleID != uuid.Nil {
		sale, err := s.repo.GetSaleInfo(ctx, returnRecord.SaleID)
		if err == nil && sale != nil {
			returnRecord.SaleInvoice = sale.InvoiceNumber
		}
	}

	return returnRecord, items, nil
}

// ProcessReturnStatusChange handles status changes and their effects
func (s *Service) ProcessReturnStatusChange(ctx context.Context, returnID uuid.UUID, newStatus string, processedBy uuid.UUID) error {
	return s.repo.ProcessReturnStatusChange(ctx, returnID, newStatus, processedBy)
}

// UpdateReturnItemStatus updates the inventory status of a return item
func (s *Service) UpdateReturnItemStatus(ctx context.Context, itemID uuid.UUID, newStatus string) error {
	return s.repo.UpdateReturnItemStatus(ctx, itemID, newStatus)
}
