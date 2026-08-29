package purchases

import (
	"time"

	"github.com/google/uuid"
)

// Purchase Status Constants
const (
	StatusDraft             = "draft"
	StatusPending           = "pending"
	StatusReceived          = "received"
	StatusCancelled         = "cancelled"
	StatusReversed          = "reversed"
	StatusPartiallyReceived = "partially_received"
)

// Legacy status aliases for backward compatibility
const (
	StatusPendingLegacy   = "pending"   // Alias for StatusPending
	StatusReceivedLegacy  = "received"  // Alias for StatusReceived
	StatusCancelledLegacy = "cancelled" // Alias for StatusCancelled
)

// Purchase represents a purchase from a supplier
type Purchase struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	SupplierID           uuid.UUID  `json:"supplier_id" db:"supplier_id"`
	InvoiceNumber        string     `json:"invoice_number" db:"invoice_number"`
	PurchaseDate         time.Time  `json:"purchase_date" db:"purchase_date"`
	ExpectedDeliveryDate *time.Time `json:"expected_delivery_date" db:"expected_delivery_date"` // Expected delivery date
	TotalAmount          float64    `json:"total_amount" db:"total_amount"`
	PaidAmount           float64    `json:"paid_amount" db:"paid_amount"`
	Status               string     `json:"status" db:"status"` // pending, received, cancelled, reversed, partially_received
	Notes                *string    `json:"notes" db:"notes"`
	UserID               *uuid.UUID `json:"user_id" db:"user_id"` // Nullable - Maps to created_by in API
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`

	// Reversal tracking (ARCHITECTURE-PRINCIPLES.md - Reverse instead of Delete)
	ReversedAt     *time.Time `json:"reversed_at" db:"reversed_at"`         // متى تم عكس الشراء
	ReversedBy     *uuid.UUID `json:"reversed_by" db:"reversed_by"`         // من قام بالعكس
	ReversalReason *string    `json:"reversal_reason" db:"reversal_reason"` // سبب العكس
}

// PurchaseItem represents an item in a purchase
type PurchaseItem struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	PurchaseID   uuid.UUID  `json:"purchase_id" db:"purchase_id"`
	ProductID    uuid.UUID  `json:"product_id" db:"product_id"`
	Quantity     int        `json:"quantity" db:"quantity"`
	UnitCost     float64    `json:"unit_cost" db:"unit_cost"`
	TotalCost    float64    `json:"total_cost" db:"total_cost"`
	SerialNumber string     `json:"serial_number" db:"serial_number"`
	Condition    string     `json:"condition" db:"condition"` // new, used, refurbished
	Grade        string     `json:"grade" db:"grade"`         // excellent, very_good, good, fair, poor
	LocationID   *uuid.UUID `json:"location_id" db:"location_id"`
	Notes        string     `json:"notes" db:"notes"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// PurchaseRequest represents purchase creation request
type PurchaseRequest struct {
	SupplierID           uuid.UUID             `json:"supplier_id" binding:"required"`
	InvoiceNumber        string                `json:"invoice_number" binding:"required"`
	PurchaseDate         time.Time             `json:"purchase_date" binding:"required"`
	ExpectedDeliveryDate *time.Time            `json:"expected_delivery_date"` // Expected delivery date
	Items                []PurchaseItemRequest `json:"items" binding:"required,min=1"`
	Notes                string                `json:"notes"`
}

// PurchaseItemRequest represents purchase item creation request
type PurchaseItemRequest struct {
	ProductID    uuid.UUID  `json:"product_id" binding:"required"`
	Quantity     int        `json:"quantity" binding:"required,min=1"`
	UnitCost     float64    `json:"unit_cost" binding:"required,min=0"`
	SerialNumber string     `json:"serial_number"`
	Condition    string     `json:"condition" binding:"required,oneof=new used refurbished"`
	Grade        string     `json:"grade" binding:"omitempty,oneof=excellent very_good good fair poor"`
	LocationID   *uuid.UUID `json:"location_id"`
	Notes        string     `json:"notes"`
}

// PurchaseUpdateRequest represents purchase update request
type PurchaseUpdateRequest struct {
	InvoiceNumber string    `json:"invoice_number"`
	PurchaseDate  time.Time `json:"purchase_date"`
	Status        string    `json:"status" binding:"omitempty,oneof=draft pending received cancelled reversed partially_received"`
	Notes         string    `json:"notes"`
}

// PurchaseResponse represents purchase response with related data
type PurchaseResponse struct {
	Purchase   Purchase       `json:"purchase"`
	Items      []PurchaseItem `json:"items"`
	Supplier   *SupplierInfo  `json:"supplier,omitempty"`
	TotalItems int            `json:"total_items"`
	Remaining  float64        `json:"remaining"`
}

// SupplierInfo represents supplier information
type SupplierInfo struct {
	ID    uuid.UUID `json:"id" db:"id"`
	Name  string    `json:"name" db:"name"`
	Phone *string   `json:"phone,omitempty" db:"phone"`
}

// PurchaseListRequest represents purchase list query parameters
type PurchaseListRequest struct {
	Page               int        `form:"page" binding:"min=1"`
	PerPage            int        `form:"per_page" binding:"min=1,max=100"`
	SupplierID         *uuid.UUID `form:"supplier_id"`
	Status             string     `form:"status" binding:"omitempty,oneof=draft pending received cancelled reversed partially_received"`
	StartDate          *time.Time `form:"start_date"`
	EndDate            *time.Time `form:"end_date"`
	Search             string     `form:"search"`
	SortBy             string     `form:"sort_by"`
	SortOrder          string     `form:"sort_order"`
	AvailableForReturn bool       `form:"available_for_return"`
}

// PurchaseReversal represents a reversal of a purchase (ARCHITECTURE-PRINCIPLES.md)
type PurchaseReversal struct {
	ID                     uuid.UUID   `json:"id" db:"id"`
	PurchaseID             uuid.UUID   `json:"purchase_id" db:"purchase_id"`
	Reason                 string      `json:"reason" db:"reason"`                                     // سبب العكس
	ReversedBy             uuid.UUID   `json:"reversed_by" db:"reversed_by"`                           // من قام بالعكس
	ReversedAt             time.Time   `json:"reversed_at" db:"reversed_at"`                           // متى تم العكس
	OriginalTotal          float64     `json:"original_total" db:"original_total"`                     // المبلغ الأصلي
	InventoryAdjustmentIDs []uuid.UUID `json:"inventory_adjustment_ids" db:"inventory_adjustment_ids"` // تعديلات المخزون المرتبطة
	CreatedAt              time.Time   `json:"created_at" db:"created_at"`
}

// PurchaseReversalRequest represents a request to reverse a purchase
type PurchaseReversalRequest struct {
	Reason string `json:"reason" binding:"required"` // سبب العكس (مثلاً: Duplicate invoice, Wrong items)
}
