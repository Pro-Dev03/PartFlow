package barcodes

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// LookupBarcode looks up a barcode and returns associated information
func (s *Service) LookupBarcode(ctx context.Context, code string, organizationID uuid.UUID) (*Barcode, error) {
	barcode, err := s.repo.GetBarcodeByCode(ctx, code, organizationID)
	if err != nil {
		return nil, fmt.Errorf("barcode not found: %w", err)
	}

	return barcode, nil
}

// GenerateBarcode generates a new barcode
func (s *Service) GenerateBarcode(ctx context.Context, req *BarcodeGenerationRequest, organizationID uuid.UUID) (*Barcode, error) {
	// Generate barcode code based on type
	code := generateBarcodeCode(req.Type, req.ProductID, req.InventoryItemID)

	now := time.Now()
	barcode := &Barcode{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Code:           code,
		Type:           req.Type,
		ProductID:      req.ProductID,
		InventoryItemID: req.InventoryItemID,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateBarcode(ctx, barcode); err != nil {
		return nil, fmt.Errorf("failed to create barcode: %w", err)
	}

	return barcode, nil
}

// ListBarcodes lists all barcodes for an organization
func (s *Service) ListBarcodes(ctx context.Context, organizationID uuid.UUID, page, perPage int) ([]*Barcode, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.ListBarcodes(ctx, organizationID, perPage, offset)
}

// DeleteBarcode deletes a barcode
func (s *Service) DeleteBarcode(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	return s.repo.DeleteBarcode(ctx, id, organizationID)
}

// GenerateLabels generates printable labels for barcodes
func (s *Service) GenerateLabels(ctx context.Context, req *LabelGenerationRequest, organizationID uuid.UUID) ([]*Label, error) {
	var labels []*Label

	for _, barcodeID := range req.BarcodeIDs {
		barcode, err := s.repo.GetBarcodeByID(ctx, barcodeID, organizationID)
		if err != nil {
			continue // Skip invalid barcodes
		}

		label := &Label{
			ID:             uuid.New(),
			OrganizationID: organizationID,
			BarcodeID:      barcodeID,
			BarcodeCode:    barcode.Code,
			ProductName:    req.ProductName,
			Price:          req.Price,
			Quantity:       req.Quantity,
			LabelFormat:    req.LabelFormat,
			PrintCount:     req.PrintCount,
			CreatedAt:      time.Now(),
		}

		labels = append(labels, label)
	}

	return labels, nil
}

// Helper function to generate barcode codes
func generateBarcodeCode(barcodeType BarcodeType, productID *uuid.UUID, inventoryItemID *uuid.UUID) string {
	switch barcodeType {
	case BarcodeTypeSKU:
		if productID != nil {
			return fmt.Sprintf("SKU-%s", productID.String()[:8])
		}
		return fmt.Sprintf("SKU-%d", time.Now().UnixNano())
	case BarcodeTypeSerial:
		if inventoryItemID != nil {
			return fmt.Sprintf("SN-%s", inventoryItemID.String()[:8])
		}
		return fmt.Sprintf("SN-%d", time.Now().UnixNano())
	case BarcodeTypeItemCode:
		if inventoryItemID != nil {
			return fmt.Sprintf("IT-%s", inventoryItemID.String()[:8])
		}
		return fmt.Sprintf("IT-%d", time.Now().UnixNano())
	case BarcodeTypeInternal:
		return fmt.Sprintf("INT-%d", time.Now().UnixNano())
	default:
		return fmt.Sprintf("EXT-%d", time.Now().UnixNano())
	}
}