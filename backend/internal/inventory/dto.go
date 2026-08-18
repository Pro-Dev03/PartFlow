package inventory

import "time"

// InventoryItemResponse represents inventory item response
type InventoryItemResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	ProductID      string    `json:"product_id"`
	ItemCode       string    `json:"item_code"`
	Barcode        string    `json:"barcode"`
	SerialNumber   string    `json:"serial_number"`
	Condition      string    `json:"condition"`
	Grade          string    `json:"grade"`
	PurchaseCost   int64     `json:"purchase_cost"`
	SellingPrice   int64     `json:"selling_price"`
	Status         string    `json:"status"`
	LocationID     *string   `json:"location_id"`
	SupplierID     *string   `json:"supplier_id"`
	PurchaseDate   *time.Time `json:"purchase_date"`
	SoldAt         *time.Time `json:"sold_at"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	
	// Computed fields
	ProductName    string    `json:"product_name,omitempty"`
	LocationName   string    `json:"location_name,omitempty"`
	SupplierName   string    `json:"supplier_name,omitempty"`
}

// LocationResponse represents location response
type LocationResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	ParentID       *string   `json:"parent_id"`
	WarehouseID    *string   `json:"warehouse_id"`
	Description    string    `json:"description"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	
	// Computed fields
	ItemCount      int       `json:"item_count"`
	ParentName     string    `json:"parent_name,omitempty"`
	WarehouseName  string    `json:"warehouse_name,omitempty"`
}

// MovementResponse represents inventory movement response
type MovementResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	ItemID         *string   `json:"item_id"`
	ProductID      *string   `json:"product_id"`
	MovementType   string    `json:"movement_type"`
	Quantity       int       `json:"quantity"`
	BeforeQuantity int       `json:"before_quantity"`
	AfterQuantity  int       `json:"after_quantity"`
	ReferenceType  string    `json:"reference_type"`
	ReferenceID    *string   `json:"reference_id"`
	Reason         string    `json:"reason"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	
	// Computed fields
	ItemName       string    `json:"item_name,omitempty"`
	ProductName    string    `json:"product_name,omitempty"`
	CreatedByName  string    `json:"created_by_name,omitempty"`
}

// ReservationResponse represents reservation response
type ReservationResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	ItemID         string    `json:"item_id"`
	CustomerID     *string   `json:"customer_id"`
	UserID         string    `json:"user_id"`
	ReservedAt     time.Time `json:"reserved_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	Status         string    `json:"status"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	
	// Computed fields
	ItemName       string    `json:"item_name,omitempty"`
	CustomerName   string    `json:"customer_name,omitempty"`
	UserName       string    `json:"user_name,omitempty"`
	IsExpired      bool      `json:"is_expired"`
}

// StockSummary represents stock summary for a product
type StockSummary struct {
	ProductID       string  `json:"product_id"`
	ProductName    string  `json:"product_name"`
	TotalQuantity  int     `json:"total_quantity"`
	Available      int     `json:"available"`
	Reserved       int     `json:"reserved"`
	Sold           int     `json:"sold"`
	Damaged        int     `json:"damaged"`
	InRepair       int     `json:"in_repair"`
	TotalValue     int64   `json:"total_value"` // in minor units
}

// BarcodeLookupResponse represents barcode lookup response
type BarcodeLookupResponse struct {
	Type           string  `json:"type"` // "product" or "item"
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Barcode        string  `json:"barcode"`
	SKU            string  `json:"sku,omitempty"`
	SerialNumber   string  `json:"serial_number,omitempty"`
	Condition      string  `json:"condition,omitempty"`
	Grade          string  `json:"grade,omitempty"`
	Status         string  `json:"status,omitempty"`
	PurchaseCost   int64   `json:"purchase_cost,omitempty"`
	SellingPrice   int64   `json:"selling_price"`
	Location       string  `json:"location,omitempty"`
	Available      bool    `json:"available"`
	WarrantyDays   int     `json:"warranty_days,omitempty"`
}

// InventoryListRequest represents inventory list query parameters
type InventoryListRequest struct {
	Page           int      `form:"page" binding:"min=1"`
	PerPage        int      `form:"per_page" binding:"min=1,max=100"`
	ProductID      *string  `form:"product_id"`
	LocationID     *string  `form:"location_id"`
	Status         *string  `form:"status"`
	Condition      *string  `form:"condition"`
	Grade          *string  `form:"grade"`
	Search         string   `form:"search"`
	SortBy         string   `form:"sort_by"`
	SortOrder      string   `form:"sort_order"`
}
