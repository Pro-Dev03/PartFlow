package inspections

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles inspection business logic
type Service struct {
	repo *Repository
}

// NewService creates a new inspection service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateInspection creates a new inspection
func (s *Service) CreateInspection(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req *InspectionRequest) (*InspectionResponse, error) {
	// Validate request
	if err := ValidateInspectionRequest(req); err != nil {
		return nil, err
	}

	// Check if product exists
	product, err := s.repo.GetProductInfo(ctx, req.ProductID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	// Create inspection
	inspection := CreateInspection(organizationID, userID, req)

	if err := s.repo.CreateInspection(ctx, inspection); err != nil {
		return nil, fmt.Errorf("failed to create inspection: %w", err)
	}

	inspector, err := s.repo.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspector info: %w", err)
	}

	return inspection.ToInspectionResponse(product, inspector), nil
}

// GetInspection retrieves an inspection by ID
func (s *Service) GetInspection(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*InspectionResponse, error) {
	inspection, err := s.repo.GetInspectionByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	product, err := s.repo.GetProductInfo(ctx, inspection.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product info: %w", err)
	}

	inspector, err := s.repo.GetUserInfo(ctx, inspection.InspectedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspector info: %w", err)
	}

	return inspection.ToInspectionResponse(product, inspector), nil
}

// ListInspections retrieves inspections with pagination and filters
func (s *Service) ListInspections(ctx context.Context, organizationID uuid.UUID, req InspectionListRequest) ([]map[string]interface{}, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	inspections, total, err := s.repo.ListInspections(ctx, organizationID, req)
	if err != nil {
		return nil, 0, err
	}

	// Convert to list items
	var result []map[string]interface{}
	for _, inspection := range inspections {
		product, err := s.repo.GetProductInfo(ctx, inspection.ProductID)
		if err != nil {
			continue
		}

		inspector, err := s.repo.GetUserInfo(ctx, inspection.InspectedBy)
		if err != nil {
			continue
		}

		inspectorName := inspector.FirstName + " " + inspector.LastName
		result = append(result, inspection.ToInspectionListItem(product.Name, inspectorName))
	}

	return result, total, nil
}

// UpdateInspection updates an inspection
func (s *Service) UpdateInspection(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *InspectionUpdateRequest) (*InspectionResponse, error) {
	inspection, err := s.repo.GetInspectionByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Check if inspection can be updated
	if inspection.IsCompleted() {
		return nil, ErrCannotUpdateCompletedInspection
	}

	// Update fields
	if !req.InspectionDate.IsZero() {
		inspection.InspectionDate = req.InspectionDate
	}
	if req.Status != "" {
		inspection.Status = req.Status
	}
	if req.Condition != "" {
		inspection.Condition = req.Condition
	}
	if req.Grade != "" {
		inspection.Grade = req.Grade
	}
	if req.Notes != "" {
		inspection.Notes = req.Notes
	}
	if req.Photos != nil {
		inspection.Photos = req.Photos
	}
	if req.TestResults.PowerTest || req.TestResults.TemperatureTest || 
		req.TestResults.PerformanceTest || req.TestResults.PortsTest || 
		req.TestResults.StorageTest || req.TestResults.VisualTest || 
		req.TestResults.SerialTest {
		inspection.TestResults = req.TestResults
	}

	inspection.UpdatedAt = time.Now()

	if err := s.repo.UpdateInspection(ctx, inspection); err != nil {
		return nil, err
	}

	return s.GetInspection(ctx, id, organizationID)
}

// DeleteInspection deletes an inspection
func (s *Service) DeleteInspection(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	inspection, err := s.repo.GetInspectionByID(ctx, id, organizationID)
	if err != nil {
		return err
	}

	// Check if inspection can be deleted
	if inspection.IsCompleted() {
		return ErrCannotUpdateCompletedInspection
	}

	return s.repo.DeleteInspection(ctx, id, organizationID)
}

// PassInspection marks an inspection as passed
func (s *Service) PassInspection(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*InspectionResponse, error) {
	inspection, err := s.repo.GetInspectionByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if inspection.IsCompleted() {
		return nil, ErrInspectionAlreadyPassed
	}

	inspection.Status = "passed"
	inspection.UpdatedAt = time.Now()

	if err := s.repo.UpdateInspection(ctx, inspection); err != nil {
		return nil, err
	}

	return s.GetInspection(ctx, id, organizationID)
}

// FailInspection marks an inspection as failed
func (s *Service) FailInspection(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*InspectionResponse, error) {
	inspection, err := s.repo.GetInspectionByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if inspection.IsCompleted() {
		return nil, ErrInspectionAlreadyFailed
	}

	inspection.Status = "failed"
	inspection.UpdatedAt = time.Now()

	if err := s.repo.UpdateInspection(ctx, inspection); err != nil {
		return nil, err
	}

	return s.GetInspection(ctx, id, organizationID)
}

// GetInspectionSummary retrieves inspection summary statistics
func (s *Service) GetInspectionSummary(ctx context.Context, organizationID uuid.UUID) (*InspectionSummary, error) {
	return s.repo.GetInspectionSummary(ctx, organizationID)
}
