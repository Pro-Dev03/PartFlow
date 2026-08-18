package purchases

import (
	"time"

	"github.com/google/uuid"
)

// Purchase represents a purchase from a supplier
type Purchase struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	SupplierID     uuid.UUID  `json:"supplier_id" db:"supplier_id"`
	InvoiceNumber  string     `json:"invoice_number" db:"invoice_number"`
	PurchaseDate   time.Time  `json:"purchase_date" db:"purchase_date"`
	TotalAmount    float64    `json:"total_amount" db:"total_amount"`
	PaidAmount     float64    `json:"paid_amount" db:"paid_amount"`
	Status         string     `json:"status" db:"status"` // pending, received, cancelled
	Notes          string     `json:"notes" db:"notes"`
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// PurchaseItem represents an item in a purchase
type PurchaseItem struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	PurchaseID     uuid.UUID  `json:"purchase_id" db:"purchase_id"`
	ProductID      uuid.UUID  `json:"product_id" db:"product_id"`
	Quantity       int        `json:"quantity" db:"quantity"`
	UnitCost       float64    `json:"unit_cost" db:"unit_cost"`
	TotalCost      float64    `json:"total_cost" db:"total_cost"`
	SerialNumber   string     `json:"serial_number" db:"serial_number"`
	Condition      string     `json:"condition" db:"condition"` // new, used, refurbished
	LocationID     *uuid.UUID `json:"location_id" db:"location_id"`
	Notes          string     `json:"notes" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// PurchaseRequest represents purchase creation request
type PurchaseRequest struct {
	SupplierID    uuid.UUID              `json:"supplier_id" binding:"required"`
	InvoiceNumber string                 `json:"invoice_number" binding:"required"`
	PurchaseDate  time.Time              `json:"purchase_date" binding:"required"`
	Items         []PurchaseItemRequest  `json:"items" binding:"required,min=1"`
	Notes         string                 `json:"notes"`
}

// PurchaseItemRequest represents purchase item creation request
type PurchaseItemRequest struct {
	ProductID    uuid.UUID  `json:"product_id" binding:"required"`
	Quantity     int        `json:"quantity" binding:"required,min=1"`
	UnitCost     float64    `json:"unit_cost" binding:"required,min=0"`
	SerialNumber string     `json:"serial_number"`
	Condition    string     `json:"condition" binding:"required,oneof=new used refurbished"`
	LocationID   *uuid.UUID `json:"location_id"`
	Notes        string     `json:"notes"`
}

// PurchaseUpdateRequest represents purchase update request
type PurchaseUpdateRequest struct {
	InvoiceNumber string    `json:"invoice_number"`
	PurchaseDate  time.Time `json:"purchase_date"`
	Status        string    `json:"status" binding:"oneof=pending received cancelled"`
	Notes         string    `json:"notes"`
}

// PurchaseResponse represents purchase response with related data
type PurchaseResponse struct {
	Purchase      Purchase       `json:"purchase"`
	Items         []PurchaseItem `json:"items"`
	Supplier      *SupplierInfo  `json:"supplier,omitempty"`
	TotalItems    int            `json:"total_items"`
	Remaining     float64        `json:"remaining"`
}

// SupplierInfo represents supplier information
type SupplierInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Phone string   `json:"phone"`
}

// PurchaseListRequest represents purchase list query parameters
type PurchaseListRequest struct {
	Page         int        `form:"page" binding:"min=1"`
	PerPage      int        `form:"per_page" binding:"min=1,max=100"`
	SupplierID   *uuid.UUID `form:"supplier_id"`
	Status       string     `form:"status" binding:"omitempty,oneof=pending received cancelled"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	Search       string     `form:"search"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}
