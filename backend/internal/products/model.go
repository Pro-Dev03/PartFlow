package products

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a product category
type Category struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Description    string     `json:"description" db:"description"`
	ParentID       *uuid.UUID `json:"parent_id" db:"parent_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Brand represents a product brand
type Brand struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Description    string     `json:"description" db:"description"`
	LogoURL        string     `json:"logo_url" db:"logo_url"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Product represents a product in the catalog
type Product struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	CategoryID     *uuid.UUID `json:"category_id" db:"category_id"`
	BrandID        *uuid.UUID `json:"brand_id" db:"brand_id"`
	Name           string     `json:"name" db:"name"`
	Description    string     `json:"description" db:"description"`
	Model          string     `json:"model" db:"model"`
	SKU            string     `json:"sku" db:"sku"`
	Barcode        string     `json:"barcode" db:"barcode"`
	TrackSerial    bool       `json:"track_serial" db:"track_serial"`
	TrackIndividual bool      `json:"track_individual" db:"track_individual"`
	MinStockLevel  int        `json:"min_stock_level" db:"min_stock_level"`
	WarrantyDays   int        `json:"warranty_days" db:"warranty_days"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// CategoryRequest represents category creation/update request
type CategoryRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
}

// BrandRequest represents brand creation/update request
type BrandRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
}

// ProductRequest represents product creation/update request
type ProductRequest struct {
	CategoryID     *uuid.UUID `json:"category_id"`
	BrandID        *uuid.UUID `json:"brand_id"`
	Name           string     `json:"name" binding:"required"`
	Description    string     `json:"description"`
	Model          string     `json:"model"`
	SKU            string     `json:"sku"`
	Barcode        string     `json:"barcode"`
	TrackSerial    bool       `json:"track_serial"`
	TrackIndividual bool      `json:"track_individual"`
	MinStockLevel  int        `json:"min_stock_level"`
	WarrantyDays   int        `json:"warranty_days"`
}

// ProductResponse represents product response with related data
type ProductResponse struct {
	Product    Product     `json:"product"`
	Category   *Category   `json:"category,omitempty"`
	Brand      *Brand      `json:"brand,omitempty"`
	StockCount int         `json:"stock_count"`
}

// ProductListRequest represents product list query parameters
type ProductListRequest struct {
	Page         int        `form:"page" binding:"min=1"`
	PerPage      int        `form:"per_page" binding:"min=1,max=100"`
	CategoryID   *uuid.UUID `form:"category_id"`
	BrandID      *uuid.UUID `form:"brand_id"`
	Search       string     `form:"search"`
	TrackSerial  *bool      `form:"track_serial"`
	TrackIndividual *bool  `form:"track_individual"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}

// ProductStockInfo represents detailed stock information for a product
type ProductStockInfo struct {
	ProductID       uuid.UUID `json:"product_id"`
	TotalStock      int       `json:"total_stock"`
	Available       int       `json:"available"`
	Reserved        int       `json:"reserved"`
	TrackIndividual bool      `json:"track_individual"`
	MinStockLevel   int       `json:"min_stock_level"`
	IsLowStock      bool      `json:"is_low_stock"`
}
