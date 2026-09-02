package parttypes

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Part Types Management
func (s *Service) ListPartTypes(ctx context.Context) ([]PartType, error) {
	return s.repo.ListPartTypes(ctx)
}

func (s *Service) GetPartType(ctx context.Context, id uuid.UUID) (*PartType, error) {
	return s.repo.GetPartType(ctx, id)
}

func (s *Service) GetPartTypeWithSpecs(ctx context.Context, id uuid.UUID) (*PartTypeWithSpecs, error) {
	return s.repo.GetPartTypeWithSpecs(ctx, id)
}

func (s *Service) CreatePartType(ctx context.Context, req *CreatePartTypeRequest) (*PartType, error) {
	partType := &PartType{
		ID:        uuid.New(),
		NameAr:    req.NameAr,
		NameEn:    req.NameEn,
		Icon:      req.Icon,
		Color:     req.Color,
		SortOrder: req.SortOrder,
		IsActive:  true,
	}

	if err := s.repo.CreatePartType(ctx, partType); err != nil {
		return nil, err
	}

	return partType, nil
}

func (s *Service) UpdatePartType(ctx context.Context, id uuid.UUID, req *UpdatePartTypeRequest) (*PartType, error) {
	partType, err := s.repo.GetPartType(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.NameAr != nil {
		partType.NameAr = *req.NameAr
	}
	if req.NameEn != nil {
		partType.NameEn = *req.NameEn
	}
	if req.Icon != nil {
		partType.Icon = *req.Icon
	}
	if req.Color != nil {
		partType.Color = *req.Color
	}
	if req.IsActive != nil {
		partType.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		partType.SortOrder = *req.SortOrder
	}

	if err := s.repo.UpdatePartType(ctx, partType); err != nil {
		return nil, err
	}

	return partType, nil
}

func (s *Service) DeletePartType(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePartType(ctx, id)
}

// Specifications Management
func (s *Service) ListSpecifications(ctx context.Context) ([]PartSpecification, error) {
	return s.repo.ListSpecifications(ctx)
}

func (s *Service) CreateSpecification(ctx context.Context, req *CreateSpecificationRequest) (*PartSpecification, error) {
	spec := &PartSpecification{
		ID:         uuid.New(),
		NameAr:     req.NameAr,
		NameEn:     req.NameEn,
		DataType:   req.DataType,
		Options:    req.Options,
		IsRequired: req.IsRequired,
	}

	if err := s.repo.CreateSpecification(ctx, spec); err != nil {
		return nil, err
	}

	return spec, nil
}

// Type Specifications Linking
func (s *Service) GetTypeSpecifications(ctx context.Context, partTypeID uuid.UUID) ([]PartSpecification, error) {
	return s.repo.GetTypeSpecifications(ctx, partTypeID)
}

func (s *Service) LinkSpecification(ctx context.Context, req *LinkSpecificationRequest) (*TypeSpecification, error) {
	link := &TypeSpecification{
		ID:              uuid.New(),
		PartTypeID:      req.PartTypeID,
		SpecificationID: req.SpecificationID,
		SortOrder:       req.SortOrder,
	}

	if err := s.repo.LinkSpecification(ctx, link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *Service) UnlinkSpecification(ctx context.Context, partTypeID, specificationID uuid.UUID) error {
	return s.repo.UnlinkSpecification(ctx, partTypeID, specificationID)
}

// Item Specification Values
func (s *Service) GetItemSpecifications(ctx context.Context, inventoryItemID uuid.UUID) ([]ItemSpecificationValue, error) {
	return s.repo.GetItemSpecifications(ctx, inventoryItemID)
}

func (s *Service) UpdateItemSpecifications(ctx context.Context, inventoryItemID uuid.UUID, req *UpdateItemSpecsRequest) error {
	// First, delete existing specifications for this item
	if err := s.repo.DeleteItemSpecifications(ctx, inventoryItemID); err != nil {
		return fmt.Errorf("failed to clear existing specifications: %w", err)
	}

	// Then, add new specifications
	for _, specValue := range req.Specifications {
		value := &ItemSpecificationValue{
			ID:              uuid.New(),
			InventoryItemID: inventoryItemID,
			SpecificationID: specValue.SpecificationID,
			ValueText:       specValue.ValueText,
			ValueNumber:     specValue.ValueNumber,
			ValueBoolean:    specValue.ValueBoolean,
		}

		if err := s.repo.SetItemSpecification(ctx, value); err != nil {
			return fmt.Errorf("failed to set specification value: %w", err)
		}
	}

	return nil
}
