package returns

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles return business logic
type Service struct {
	repo *Repository
}

// NewService creates a new return service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateReturn creates a new return with items
func (s *Service) CreateReturn(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req *ReturnRequest) (*ReturnResponse, error) {
	// Validate request
	if err := ValidateReturnRequest(req); err != nil {
		return nil, err
	}

	// Check if sale exists
	sale, err := s.repo.GetSaleInfo(ctx, req.SaleID)
	if err != nil {
		return nil, ErrSaleNotFound
	}

	// Check if return number already exists
	existing, err := s.repo.GetReturnByReturnNumber(ctx, generateReturnNumber(), organizationID)
	if err == nil && existing != nil {
		return nil, ErrReturnExists
	}

	// Create return
	returnRecord := CreateReturn(organizationID, userID, req)
	returnRecord.CustomerID = sale.ID // This should come from the sale

	if err := s.repo.CreateReturn(ctx, returnRecord); err != nil {
		return nil, fmt.Errorf("failed to create return: %w", err)
	}

	// Create return items
	var items []ReturnItem
	var totalRefund float64

	for _, itemReq := range req.Items {
		// Get sale item info
		saleItem, err := s.repo.GetSaleItemInfo(ctx, itemReq.SaleItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get sale item info: %w", err)
		}

		// Validate quantity
		if itemReq.Quantity > saleItem.Quantity {
			return nil, ErrInsufficientStock
		}

		item := CreateReturnItem(returnRecord.ID, itemReq, saleItem.UnitPrice)
		if err := s.repo.CreateReturnItem(ctx, item); err != nil {
			return nil, fmt.Errorf("failed to create return item: %w", err)
		}
		items = append(items, *item)
		totalRefund += item.TotalPrice
	}

	// Update return with total refund amount
	returnRecord.RefundAmount = totalRefund
	returnRecord.UpdatedAt = time.Now()
	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, fmt.Errorf("failed to update return: %w", err)
	}

	// Get customer info
	customer, err := s.repo.GetCustomerInfo(ctx, returnRecord.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer info: %w", err)
	}

	return returnRecord.ToReturnResponse(items, customer, sale), nil
}

// GetReturn retrieves a return by ID
func (s *Service) GetReturn(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetReturnItems(ctx, returnRecord.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get return items: %w", err)
	}

	customer, err := s.repo.GetCustomerInfo(ctx, returnRecord.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer info: %w", err)
	}

	sale, err := s.repo.GetSaleInfo(ctx, returnRecord.SaleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sale info: %w", err)
	}

	return returnRecord.ToReturnResponse(items, customer, sale), nil
}

// ListReturns retrieves returns with pagination and filters
func (s *Service) ListReturns(ctx context.Context, organizationID uuid.UUID, req ReturnListRequest) ([]map[string]interface{}, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	returns, total, err := s.repo.ListReturns(ctx, organizationID, req)
	if err != nil {
		return nil, 0, err
	}

	// Convert to list items
	var result []map[string]interface{}
	for _, returnRecord := range returns {
		items, err := s.repo.GetReturnItems(ctx, returnRecord.ID)
		if err != nil {
			continue
		}

		customer, err := s.repo.GetCustomerInfo(ctx, returnRecord.CustomerID)
		if err != nil {
			continue
		}

		sale, err := s.repo.GetSaleInfo(ctx, returnRecord.SaleID)
		if err != nil {
			continue
		}

		result = append(result, returnRecord.ToReturnListItem(len(items), customer.Name, sale.InvoiceNumber))
	}

	return result, total, nil
}

// UpdateReturn updates a return
func (s *Service) UpdateReturn(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *ReturnUpdateRequest) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Check if return can be updated
	if returnRecord.Status == "completed" {
		return nil, ErrReturnAlreadyCompleted
	}

	// Update fields
	if !req.ReturnDate.IsZero() {
		returnRecord.ReturnDate = req.ReturnDate
	}
	if req.Reason != "" {
		returnRecord.Reason = req.Reason
	}
	if req.Condition != "" {
		returnRecord.Condition = req.Condition
	}
	if req.Status != "" {
		// Validate status transition
		if returnRecord.Status == "approved" && req.Status == "pending" {
			return nil, ErrInvalidReturnStatus
		}
		if returnRecord.Status == "rejected" && req.Status == "pending" {
			return nil, ErrInvalidReturnStatus
		}
		returnRecord.Status = req.Status
	}
	if req.RefundAmount >= 0 {
		returnRecord.RefundAmount = req.RefundAmount
	}
	if req.RefundMethod != "" {
		returnRecord.RefundMethod = req.RefundMethod
	}
	if req.Notes != "" {
		returnRecord.Notes = req.Notes
	}

	returnRecord.UpdatedAt = time.Now()

	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, err
	}

	return s.GetReturn(ctx, id, organizationID)
}

// DeleteReturn deletes a return
func (s *Service) DeleteReturn(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	returnRecord, err := s.repo.GetReturnByID(ctx, id, organizationID)
	if err != nil {
		return err
	}

	// Check if return can be deleted
	if returnRecord.Status == "completed" {
		return ErrReturnAlreadyCompleted
	}

	return s.repo.DeleteReturn(ctx, id, organizationID)
}

// ApproveReturn approves a return
func (s *Service) ApproveReturn(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if returnRecord.Status == "approved" {
		return nil, ErrReturnAlreadyApproved
	}

	if returnRecord.Status == "rejected" {
		return nil, ErrReturnAlreadyRejected
	}

	if returnRecord.Status == "completed" {
		return nil, ErrReturnAlreadyCompleted
	}

	returnRecord.Status = "approved"
	returnRecord.UpdatedAt = time.Now()

	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, err
	}

	return s.GetReturn(ctx, id, organizationID)
}

// RejectReturn rejects a return
func (s *Service) RejectReturn(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if returnRecord.Status == "approved" {
		return nil, ErrReturnAlreadyApproved
	}

	if returnRecord.Status == "rejected" {
		return nil, ErrReturnAlreadyRejected
	}

	if returnRecord.Status == "completed" {
		return nil, ErrReturnAlreadyCompleted
	}

	returnRecord.Status = "rejected"
	returnRecord.UpdatedAt = time.Now()

	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, err
	}

	return s.GetReturn(ctx, id, organizationID)
}

// ProcessRefund processes refund for a return
func (s *Service) ProcessRefund(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*ReturnResponse, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if returnRecord.Status != "approved" {
		return nil, ErrInvalidReturnStatus
	}

	if returnRecord.RefundDate != nil {
		return nil, ErrRefundAlreadyProcessed
	}

	now := time.Now()
	returnRecord.RefundDate = &now
	returnRecord.Status = "completed"
	returnRecord.UpdatedAt = now

	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, err
	}

	return s.GetReturn(ctx, id, organizationID)
}

// AddReturnItem adds an item to a return
func (s *Service) AddReturnItem(ctx context.Context, returnID uuid.UUID, organizationID uuid.UUID, req ReturnItemRequest) (*ReturnItem, error) {
	returnRecord, err := s.repo.GetReturnByID(ctx, returnID, organizationID)
	if err != nil {
		return nil, err
	}

	if returnRecord.Status == "completed" {
		return nil, ErrReturnAlreadyCompleted
	}

	// Get sale item info
	saleItem, err := s.repo.GetSaleItemInfo(ctx, req.SaleItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sale item info: %w", err)
	}

	// Validate quantity
	if req.Quantity > saleItem.Quantity {
		return nil, ErrInsufficientStock
	}

	item := CreateReturnItem(returnID, req, saleItem.UnitPrice)
	if err := s.repo.CreateReturnItem(ctx, item); err != nil {
		return nil, err
	}

	// Update return total refund amount
	returnRecord.RefundAmount += item.TotalPrice
	returnRecord.UpdatedAt = time.Now()
	if err := s.repo.UpdateReturn(ctx, returnRecord); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateReturnItem updates a return item
func (s *Service) UpdateReturnItem(ctx context.Context, itemID uuid.UUID, organizationID uuid.UUID, req ReturnItemRequest) (*ReturnItem, error) {
	// Get the item first
	items, err := s.repo.GetReturnItems(ctx, uuid.Nil) // This would need proper implementation
	if err != nil {
		return nil, err
	}

	// Find the item by ID
	var foundItem *ReturnItem
	for i := range items {
		if items[i].ID == itemID {
			foundItem = &items[i]
			break
		}
	}

	if foundItem == nil {
		return nil, ErrReturnItemNotFound
	}

	// Update fields
	foundItem.Quantity = req.Quantity
	foundItem.Reason = req.Reason
	foundItem.Condition = req.Condition
	foundItem.IsResellable = req.IsResellable
	foundItem.LocationID = req.LocationID
	foundItem.Notes = req.Notes
	foundItem.UpdatedAt = time.Now()

	if err := s.repo.UpdateReturnItem(ctx, foundItem); err != nil {
		return nil, err
	}

	return foundItem, nil
}

// DeleteReturnItem deletes a return item
func (s *Service) DeleteReturnItem(ctx context.Context, itemID uuid.UUID) error {
	return s.repo.DeleteReturnItem(ctx, itemID)
}
