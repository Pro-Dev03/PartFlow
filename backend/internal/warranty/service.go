package warranty

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles warranty business logic
type Service struct {
	repo *Repository
}

// NewService creates a new warranty service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateWarrantyClaim creates a new warranty claim with items
func (s *Service) CreateWarrantyClaim(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req *WarrantyClaimRequest) (*WarrantyClaimResponse, error) {
	// Validate warranty coverage
	warranty, err := s.repo.GetActiveWarranty(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate warranty: %w", err)
	}

	// Check if warranty is still valid
	if time.Now().After(warranty.EndDate) {
		return nil, ErrWarrantyExpired
	}

	// Create claim
	claim := NewWarrantyClaim(organizationID, req.ProductID, req.SerialNumber, req.ClaimType, req.Reason, userID)
	claim.CustomerID = req.CustomerID
	claim.SaleID = req.SaleID
	claim.Description = req.Description
	if req.Priority != "" {
		claim.Priority = req.Priority
	}
	if req.EstimatedCost > 0 {
		claim.EstimatedCost = req.EstimatedCost
	}
	claim.AssignedTo = req.AssignedTo
	claim.Notes = req.Notes

	if err := s.repo.CreateWarrantyClaim(ctx, claim); err != nil {
		return nil, fmt.Errorf("failed to create warranty claim: %w", err)
	}

	// Create claim items
	var items []WarrantyClaimItem
	for _, itemReq := range req.Items {
		item := &WarrantyClaimItem{
			ID:               uuid.New(),
			ClaimID:          claim.ID,
			ProductID:        itemReq.ProductID,
			SerialNumber:     itemReq.SerialNumber,
			Quantity:         itemReq.Quantity,
			Condition:        itemReq.Condition,
			DefectType:       itemReq.DefectType,
			DefectDescription: itemReq.DefectDescription,
			Notes:            itemReq.Notes,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := s.repo.CreateWarrantyClaimItem(ctx, item); err != nil {
			return nil, fmt.Errorf("failed to create warranty claim item: %w", err)
		}
		items = append(items, *item)
	}

	// Get related information
	product, _ := s.repo.GetProductInfo(ctx, req.ProductID)
	var customer *CustomerInfo
	if req.CustomerID != nil {
		customer, _ = s.repo.GetCustomerInfo(ctx, *req.CustomerID)
	}
	var assignedTo *UserInfo
	if req.AssignedTo != nil {
		assignedTo, _ = s.repo.GetUserInfo(ctx, *req.AssignedTo)
	}
	createdBy, _ := s.repo.GetUserInfo(ctx, userID)

	return s.toWarrantyClaimResponse(claim, items, product, customer, assignedTo, createdBy), nil
}

// GetWarrantyClaim retrieves a warranty claim by ID
func (s *Service) GetWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*WarrantyClaimResponse, error) {
	claim, err := s.repo.GetWarrantyClaimByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetWarrantyClaimItems(ctx, claim.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claim items: %w", err)
	}

	product, _ := s.repo.GetProductInfo(ctx, claim.ProductID)
	var customer *CustomerInfo
	if claim.CustomerID != nil {
		customer, _ = s.repo.GetCustomerInfo(ctx, *claim.CustomerID)
	}
	var assignedTo *UserInfo
	if claim.AssignedTo != nil {
		assignedTo, _ = s.repo.GetUserInfo(ctx, *claim.AssignedTo)
	}
	createdBy, _ := s.repo.GetUserInfo(ctx, claim.CreatedBy)

	return s.toWarrantyClaimResponse(claim, items, product, customer, assignedTo, createdBy), nil
}

// ListWarrantyClaims retrieves warranty claims with pagination and filters
func (s *Service) ListWarrantyClaims(ctx context.Context, organizationID uuid.UUID, req WarrantyClaimListRequest) ([]map[string]interface{}, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	claims, total, err := s.repo.ListWarrantyClaims(ctx, organizationID, req)
	if err != nil {
		return nil, 0, err
	}

	// Convert to list items
	var result []map[string]interface{}
	for _, claim := range claims {
		items, err := s.repo.GetWarrantyClaimItems(ctx, claim.ID)
		if err != nil {
			continue
		}

		product, _ := s.repo.GetProductInfo(ctx, claim.ProductID)
		productName := ""
		if product != nil {
			productName = product.Name
		}

		result = append(result, map[string]interface{}{
			"id":           claim.ID,
			"claim_number": claim.ClaimNumber,
			"product_name": productName,
			"serial_number": claim.SerialNumber,
			"claim_type":   claim.ClaimType,
			"status":       claim.Status,
			"priority":     claim.Priority,
			"claim_date":   claim.ClaimDate,
			"total_items":  len(items),
		})
	}

	return result, total, nil
}

// UpdateWarrantyClaim updates a warranty claim
func (s *Service) UpdateWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *WarrantyClaimUpdateRequest) (*WarrantyClaimResponse, error) {
	claim, err := s.repo.GetWarrantyClaimByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Check if claim can be updated
	if claim.Status == "completed" {
		return nil, ErrClaimAlreadyCompleted
	}

	// Update fields
	if req.Status != "" {
		// Validate status transition
		if claim.Status == "rejected" && req.Status != "pending" {
			return nil, ErrInvalidClaimStatus
		}
		if claim.Status == "approved" && req.Status == "pending" {
			return nil, ErrInvalidClaimStatus
		}
		claim.Status = req.Status
		
		// Set resolved date if completed
		if req.Status == "completed" {
			now := time.Now()
			claim.ResolvedAt = &now
		}
	}
	if req.Priority != "" {
		claim.Priority = req.Priority
	}
	if req.EstimatedCost >= 0 {
		claim.EstimatedCost = req.EstimatedCost
	}
	if req.ActualCost >= 0 {
		claim.ActualCost = req.ActualCost
	}
	if req.Resolution != nil {
		claim.Resolution = req.Resolution
	}
	if req.AssignedTo != nil {
		claim.AssignedTo = req.AssignedTo
	}
	if req.Notes != "" {
		claim.Notes = req.Notes
	}

	claim.UpdatedAt = time.Now()

	if err := s.repo.UpdateWarrantyClaim(ctx, claim); err != nil {
		return nil, err
	}

	return s.GetWarrantyClaim(ctx, id, organizationID)
}

// ApproveWarrantyClaim approves a warranty claim
func (s *Service) ApproveWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*WarrantyClaimResponse, error) {
	claim, err := s.repo.GetWarrantyClaimByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if claim.Status == "approved" {
		return nil, ErrClaimAlreadyApproved
	}

	if claim.Status == "rejected" {
		return nil, ErrClaimAlreadyRejected
	}

	if claim.Status == "completed" {
		return nil, ErrClaimAlreadyCompleted
	}

	claim.Status = "approved"
	claim.UpdatedAt = time.Now()

	if err := s.repo.UpdateWarrantyClaim(ctx, claim); err != nil {
		return nil, err
	}

	return s.GetWarrantyClaim(ctx, id, organizationID)
}

// RejectWarrantyClaim rejects a warranty claim
func (s *Service) RejectWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, reason string) (*WarrantyClaimResponse, error) {
	claim, err := s.repo.GetWarrantyClaimByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if claim.Status == "approved" {
		return nil, ErrClaimAlreadyApproved
	}

	if claim.Status == "rejected" {
		return nil, ErrClaimAlreadyRejected
	}

	if claim.Status == "completed" {
		return nil, ErrClaimAlreadyCompleted
	}

	claim.Status = "rejected"
	claim.Resolution = &reason
	claim.UpdatedAt = time.Now()

	if err := s.repo.UpdateWarrantyClaim(ctx, claim); err != nil {
		return nil, err
	}

	return s.GetWarrantyClaim(ctx, id, organizationID)
}

// CreateWarranty creates a new warranty for a product
func (s *Service) CreateWarranty(ctx context.Context, organizationID uuid.UUID, req *WarrantyRequest) (*Warranty, error) {
	warranty := NewWarranty(organizationID, req.ProductID, req.WarrantyType, req.DurationDays, req.Coverage)
	warranty.Terms = req.Terms

	if err := s.repo.CreateWarranty(ctx, warranty); err != nil {
		return nil, fmt.Errorf("failed to create warranty: %w", err)
	}

	return warranty, nil
}

// GetWarranty retrieves a warranty by ID
func (s *Service) GetWarranty(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Warranty, error) {
	return s.repo.GetWarrantyByID(ctx, id, organizationID)
}

// GetActiveWarranty retrieves the active warranty for a product
func (s *Service) GetActiveWarranty(ctx context.Context, productID uuid.UUID) (*Warranty, error) {
	return s.repo.GetActiveWarranty(ctx, productID)
}

// GetWarrantyClaimsSummary retrieves warranty claims summary statistics
func (s *Service) GetWarrantyClaimsSummary(ctx context.Context, organizationID uuid.UUID) (*WarrantyClaimsSummary, error) {
	return s.repo.GetWarrantyClaimsSummary(ctx, organizationID)
}

// toWarrantyClaimResponse converts a WarrantyClaim to WarrantyClaimResponse
func (s *Service) toWarrantyClaimResponse(claim *WarrantyClaim, items []WarrantyClaimItem, product *ProductInfo, customer *CustomerInfo, assignedTo *UserInfo, createdBy *UserInfo) *WarrantyClaimResponse {
	return &WarrantyClaimResponse{
		Claim:      *claim,
		Items:      items,
		Product:    product,
		Customer:   customer,
		AssignedTo: assignedTo,
		CreatedBy:  createdBy,
		TotalItems: len(items),
	}
}

// WarrantyClaimsSummary represents warranty claims summary statistics
type WarrantyClaimsSummary struct {
	TotalClaims    int     `json:"total_claims"`
	PendingClaims  int     `json:"pending_claims"`
	ApprovedClaims int     `json:"approved_claims"`
	RejectedClaims int     `json:"rejected_claims"`
	CompletedClaims int    `json:"completed_claims"`
	TotalCost      float64 `json:"total_cost"`
	AverageResolutionTime float64 `json:"average_resolution_time_days"`
}