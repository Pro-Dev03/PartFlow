package barcodes

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ProductInfo represents product information from barcode lookup
type ProductInfo struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	SKU          string    `json:"sku"`
	Barcode      string    `json:"barcode"`
	SellingPrice float64   `json:"sellingPrice"`
	CostPrice    float64   `json:"costPrice"`
	Stock        int       `json:"stock"`
	Condition    string    `json:"condition"`
	Category     string    `json:"category"`
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// LookupBarcode looks up a barcode and returns associated information
func (s *Service) LookupBarcode(ctx context.Context, code string) (*Barcode, error) {
	barcode, err := s.repo.GetBarcodeByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("barcode not found: %w", err)
	}

	return barcode, nil
}

// LookupProductByBarcode looks up product information by barcode code
func (s *Service) LookupProductByBarcode(ctx context.Context, code string) (*ProductInfo, error) {
	product, err := s.repo.GetProductByBarcode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	return product, nil
}

// LookupProductBySKU looks up product information by SKU
func (s *Service) LookupProductBySKU(ctx context.Context, sku string) (*ProductInfo, error) {
	product, err := s.repo.GetProductBySKU(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	return product, nil
}

// GenerateBarcode generates a new barcode
func (s *Service) GenerateBarcode(ctx context.Context, req *BarcodeGenerationRequest) (*Barcode, error) {
	// Generate barcode code based on type
	code := generateBarcodeCode(req.Type, req.ProductID, req.InventoryItemID)

	now := time.Now()
	barcode := &Barcode{
		ID:              uuid.New(),
		Code:            code,
		Type:            req.Type,
		ProductID:       req.ProductID,
		InventoryItemID: req.InventoryItemID,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.CreateBarcode(ctx, barcode); err != nil {
		return nil, fmt.Errorf("failed to create barcode: %w", err)
	}

	return barcode, nil
}

// ListBarcodes lists all barcodes
func (s *Service) ListBarcodes(ctx context.Context, page, perPage int) ([]*Barcode, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.ListBarcodes(ctx, perPage, offset)
}

// DeleteBarcode deletes a barcode
func (s *Service) DeleteBarcode(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteBarcode(ctx, id)
}

// GenerateLabels generates printable labels for barcodes
func (s *Service) GenerateLabels(ctx context.Context, req *LabelGenerationRequest) ([]*Label, error) {
	var labels []*Label

	for _, barcodeID := range req.BarcodeIDs {
		barcode, err := s.repo.GetBarcodeByID(ctx, barcodeID)
		if err != nil {
			continue // Skip invalid barcodes
		}

		label := &Label{
			ID:          uuid.New(),
			BarcodeID:   barcodeID,
			BarcodeCode: barcode.Code,
			ProductName: req.ProductName,
			Price:       req.Price,
			Quantity:    req.Quantity,
			LabelFormat: req.LabelFormat,
			PrintCount:  req.PrintCount,
			CreatedAt:   time.Now(),
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
