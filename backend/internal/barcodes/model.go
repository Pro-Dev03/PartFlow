package barcodes

import (
	"time"

	"github.com/google/uuid"
)

// BarcodeType represents the type of barcode
type BarcodeType string

const (
	BarcodeTypeExternal BarcodeType = "EXTERNAL"
	BarcodeTypeInternal BarcodeType = "INTERNAL"
	BarcodeTypeSKU       BarcodeType = "SKU"
	BarcodeTypeSerial    BarcodeType = "SERIAL"
	BarcodeTypeItemCode  BarcodeType = "ITEM_CODE"
)

// Barcode represents a barcode
type Barcode struct {
	ID             uuid.UUID    `json:"id" db:"id"`
	OrganizationID uuid.UUID    `json:"organization_id" db:"organization_id"`
	Code           string       `json:"code" db:"code"`
	Type           BarcodeType  `json:"type" db:"type"`
	ProductID      *uuid.UUID   `json:"product_id" db:"product_id"`
	InventoryItemID *uuid.UUID  `json:"inventory_item_id" db:"inventory_item_id"`
	IsActive       bool         `json:"is_active" db:"is_active"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
}

// BarcodeGenerationRequest represents barcode generation request
type BarcodeGenerationRequest struct {
	Type           BarcodeType `json:"type" binding:"required"`
	ProductID      *uuid.UUID  `json:"product_id"`
	InventoryItemID *uuid.UUID `json:"inventory_item_id"`
	Quantity       int         `json:"quantity"`
}

// BarcodeLabelRequest represents barcode label generation request
type BarcodeLabelRequest struct {
	Barcodes []string `json:"barcodes" binding:"required"`
	Format   string   `json:"format"` // pdf, zpl, etc
}

// LabelGenerationRequest represents label generation request
type LabelGenerationRequest struct {
	BarcodeIDs   []uuid.UUID `json:"barcode_ids" binding:"required"`
	ProductName  string      `json:"product_name"`
	Price        float64     `json:"price"`
	Quantity     int         `json:"quantity"`
	LabelFormat  string      `json:"label_format"` // small, medium, large
	PrintCount   int         `json:"print_count"`
}

// Label represents a printable label
type Label struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	BarcodeID      uuid.UUID `json:"barcode_id" db:"barcode_id"`
	BarcodeCode    string    `json:"barcode_code" db:"barcode_code"`
	ProductName    string    `json:"product_name" db:"product_name"`
	Price          float64   `json:"price" db:"price"`
	Quantity       int       `json:"quantity" db:"quantity"`
	LabelFormat    string    `json:"label_format" db:"label_format"`
	PrintCount     int       `json:"print_count" db:"print_count"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
