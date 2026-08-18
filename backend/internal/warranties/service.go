package warranties

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

// CreateWarranty creates a new warranty
func (s *Service) CreateWarranty(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req *WarrantyRequest) (*WarrantyResponse, error) {
	// Validate request
	if err := ValidateWarrantyRequest(req); err != nil {
		return nil, err
	}

	// Check if product exists
	product, err := s.repo.GetProductInfo(ctx, req.ProductID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	// Create warranty
	warranty := CreateWarranty(organizationID, userID, req)

	if err := s.repo.CreateWarranty(ctx, warranty); err != nil {
		return nil, fmt.Errorf("failed to create warranty: %w", err)
	}

	var customer *CustomerInfo
	if req.CustomerID != nil {
		customer, err = s.repo.GetCustomerInfo(ctx, *req.CustomerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get customer info: %w", err)
		}
	}

	return warranty.ToWarrantyResponse(product, customer, nil), nil
}

// GetWarranty retrieves a warranty by ID
func (s *Service) GetWarranty(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*WarrantyResponse, error) {
	warranty, err := s.repo.GetWarrantyByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	product, err := s.repo.GetProductInfo(ctx, warranty.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product info: %w", err)
	}

	var customer *CustomerInfo
	if warranty.CustomerID != nil {
		customer, err = s.repo.GetCustomerInfo(ctx, *warranty.CustomerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get customer info: %w", err)
		}
	}

	claims, err := s.repo.GetWarrantyClaimsByWarrantyID(ctx, warranty.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claims: %w", err)
	}

	return warranty.ToWarrantyResponse(product, customer, claims), nil
}

// ListWarranties retrieves warranties with pagination and filters
func (s *Service) ListWarranties(ctx context.Context, organizationID uuid.UUID, req WarrantyListRequest) ([]map[string]interface{}, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	warranties, total, err := s.repo.ListWarranties(ctx, organizationID, req)
	if err != nil {
		return nil, 0, err
	}

	// Convert to list items
	var result []map[string]interface{}
	for _, warranty := range warranties {
		product, err := s.repo.GetProductInfo(ctx, warranty.ProductID)
		if err != nil {
			continue
		}

		var customerName string
		if warranty.CustomerID != nil {
			customer, err := s.repo.GetCustomerInfo(ctx, *warranty.CustomerID)
			if err == nil {
				customerName = customer.Name
			}
		}

		result = append(result, warranty.ToWarrantyListItem(product.Name, customerName))
	}

	return result, total, nil
}

// UpdateWarranty updates a warranty
func (s *Service) UpdateWarranty(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *WarrantyUpdateRequest) (*WarrantyResponse, error) {
	warranty, err := s.repo.GetWarrantyByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.SerialNumber != "" {
		warranty.SerialNumber = req.SerialNumber
	}
	if req.WarrantyType != "" {
		warranty.WarrantyType = req.WarrantyType
	}
	if req.WarrantyPeriod > 0 {
		warranty.WarrantyPeriod = req.WarrantyPeriod
		// Recalculate end date if start date is provided
		if !req.StartDate.IsZero() {
			warranty.StartDate = req.StartDate
			warranty.EndDate = req.StartDate.AddDate(0, 0, req.WarrantyPeriod)
		}
	}
	if !req.StartDate.IsZero() {
		warranty.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		warranty.EndDate = req.EndDate
	}
	if req.Status != "" {
		warranty.Status = req.Status
	}
	if req.CustomerID != nil {
		warranty.CustomerID = req.CustomerID
	}
	if req.Terms != "" {
		warranty.Terms = req.Terms
	}
	if req.Notes != "" {
		warranty.Notes = req.Notes
	}

	warranty.UpdatedAt = time.Now()

	if err := s.repo.UpdateWarranty(ctx, warranty); err != nil {
		return nil, err
	}

	return s.GetWarranty(ctx, id, organizationID)
}

// DeleteWarranty deletes a warranty
func (s *Service) DeleteWarranty(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	warranty, err := s.repo.GetWarrantyByID(ctx, id, organizationID)
	if err != nil {
		return err
	}

	// Check if warranty has claims
	claims, err := s.repo.GetWarrantyClaimsByWarrantyID(ctx, warranty.ID)
	if err == nil && len(claims) > 0 {
		return ErrWarrantyClaimed
	}

	return s.repo.DeleteWarranty(ctx, id, organizationID)
}

// CreateWarrantyClaim creates a new warranty claim
func (s *Service) CreateWarrantyClaim(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req *WarrantyClaimRequest) (*WarrantyClaimResponse, error) {
	// Validate request
	if err := ValidateWarrantyClaimRequest(req); err != nil {
		return nil, err
	}

	// Check if warranty exists and is valid
	warranty, err := s.repo.GetWarrantyByID(ctx, req.WarrantyID, organizationID)
	if err != nil {
		return nil, ErrWarrantyNotFound
	}

	// Check if warranty is active
	if warranty.Status != "active" {
		return nil, ErrCannotClaimVoidedWarranty
	}

	// Check if warranty is expired
	if warranty.IsExpired() {
		return nil, ErrCannotClaimExpiredWarranty
	}

	// Check if customer exists
	customer, err := s.repo.GetCustomerInfo(ctx, req.CustomerID)
	if err != nil {
		return nil, ErrCustomerNotFound
	}

	// Create claim
	claim := CreateWarrantyClaim(organizationID, userID, req)

	if err := s.repo.CreateWarrantyClaim(ctx, claim); err != nil {
		return nil, fmt.Errorf("failed to create warranty claim: %w", err)
	}

	// Update warranty status to claimed
	if err := s.repo.UpdateWarrantyStatus(ctx, warranty.ID, "claimed"); err != nil {
		return nil, fmt.Errorf("failed to update warranty status: %w", err)
	}

	// Refresh warranty data
	warranty, _ = s.repo.GetWarrantyByID(ctx, req.WarrantyID, organizationID)

	return claim.ToWarrantyClaimResponse(warranty, customer), nil
}

// GetWarrantyClaim retrieves a warranty claim by ID
func (s *Service) GetWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*WarrantyClaimResponse, error) {
	claim, err := s.repo.GetWarrantyClaimByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	warranty, err := s.repo.GetWarrantyByID(ctx, claim.WarrantyID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty: %w", err)
	}

	customer, err := s.repo.GetCustomerInfo(ctx, claim.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer info: %w", err)
	}

	return claim.ToWarrantyClaimResponse(warranty, customer), nil
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
		warranty, err := s.repo.GetWarrantyByID(ctx, claim.WarrantyID, organizationID)
		if err != nil {
			continue
		}

		customer, err := s.repo.GetCustomerInfo(ctx, claim.CustomerID)
		if err != nil {
			continue
		}

		product, err := s.repo.GetProductInfo(ctx, warranty.ProductID)
		if err != nil {
			continue
		}

		result = append(result, claim.ToWarrantyClaimListItem(warranty.WarrantyNumber, customer.Name, product.Name))
	}

	return result, total, nil
}

// UpdateWarrantyClaim updates a warranty claim
func (s *Service) UpdateWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, userID uuid.UUID, req *WarrantyClaimUpdateRequest) (*WarrantyClaimResponse, error) {
	claim, err := s.repo.GetWarrantyClaimByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.IssueDescription != "" {
		claim.IssueDescription = req.IssueDescription
	}
	if req.Status != "" {
		// Validate status transition
		if claim.Status == "approved" && req.Status == "pending" {
			return nil, ErrInvalidClaimStatus
		}
		if claim.Status == "rejected" && req.Status == "pending" {
			return nil, ErrInvalidClaimStatus
		}
		if claim.Status == "completed" && req.Status != "completed" {
			return nil, ErrClaimAlreadyCompleted
		}
		
		claim.Status = req.Status
		
		// Set appropriate fields based on status
		now := time.Now()
		switch req.Status {
		case "approved":
			if claim.ApprovedBy == nil {
				claim.ApprovedBy = &userID
				claim.ApprovedDate = &now
			}
		case "rejected":
			if claim.RejectedBy == nil {
				claim.RejectedBy = &userID
				claim.RejectedDate = &now
			}
		case "completed":
			if claim.CompletedBy == nil {
				claim.CompletedBy = &userID
				claim.CompletedDate = &now
			}
		}
	}
	if req.Resolution != "" {
		claim.Resolution = req.Resolution
	}
	if !req.ResolutionDate.IsZero() {
		claim.ResolutionDate = &req.ResolutionDate
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

// DeleteWarrantyClaim deletes a warranty claim
func (s *Service) DeleteWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	claim, err := s.repo.GetWarrantyClaimByID(ctx, id, organizationID)
	if err != nil {
		return err
	}

	// Check if claim can be deleted
	if claim.Status == "completed" {
		return ErrClaimAlreadyCompleted
	}

	return s.repo.DeleteWarrantyClaim(ctx, id, organizationID)
}

// GetWarrantiesExpiringSoon retrieves warranties that will expire soon
func (s *Service) GetWarrantiesExpiringSoon(ctx context.Context, organizationID uuid.UUID, days int) ([]WarrantyExpiringSoon, error) {
	return s.repo.GetWarrantiesExpiringSoon(ctx, organizationID, days)
}

// ApproveWarrantyClaim approves a warranty claim
func (s *Service) ApproveWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, userID uuid.UUID) (*WarrantyClaimResponse, error) {
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

	now := time.Now()
	claim.Status = "approved"
	claim.ApprovedBy = &userID
	claim.ApprovedDate = &now
	claim.UpdatedAt = now

	if err := s.repo.UpdateWarrantyClaim(ctx, claim); err != nil {
		return nil, err
	}

	return s.GetWarrantyClaim(ctx, id, organizationID)
}

// RejectWarrantyClaim rejects a warranty claim
func (s *Service) RejectWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, userID uuid.UUID) (*WarrantyClaimResponse, error) {
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

	now := time.Now()
	claim.Status = "rejected"
	claim.RejectedBy = &userID
	claim.RejectedDate = &now
	claim.UpdatedAt = now

	if err := s.repo.UpdateWarrantyClaim(ctx, claim); err != nil {
		return nil, err
	}

	// Revert warranty status to active
	if err := s.repo.UpdateWarrantyStatus(ctx, claim.WarrantyID, "active"); err != nil {
		return nil, fmt.Errorf("failed to revert warranty status: %w", err)
	}

	return s.GetWarrantyClaim(ctx, id, organizationID)
}

// CompleteWarrantyClaim completes a warranty claim
func (s *Service) CompleteWarrantyClaim(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, userID uuid.UUID, resolution string) (*WarrantyClaimResponse, error) {
	claim, err := s.repo.GetWarrantyClaimByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if claim.Status != "approved" && claim.Status != "in_progress" {
		return nil, ErrInvalidClaimStatus
	}

	now := time.Now()
	claim.Status = "completed"
	claim.Resolution = resolution
	claim.ResolutionDate = &now
	claim.CompletedBy = &userID
	claim.CompletedDate = &now
	claim.UpdatedAt = now

	if err := s.repo.UpdateWarrantyClaim(ctx, claim); err != nil {
		return nil, err
	}

	return s.GetWarrantyClaim(ctx, id, organizationID)
}
