package acquisitions

import (
	"time"

	"github.com/google/uuid"
)

// Acquisition Type Constants
const (
	TypeSupplier = "SUPPLIER" // شراء من مورد
	TypeCustomer = "CUSTOMER" // شراء من عميل (قطع مستعملة)
)

// Acquisition Status Constants
const (
	StatusDraft      = "draft"
	StatusPending    = "pending"
	StatusAcquired   = "acquired"
	StatusInspection = "inspection"
	StatusApproved   = "approved"
	StatusRejected   = "rejected"
	StatusCancelled  = "cancelled"
	StatusReversed   = "reversed"
)

// Payment Status Constants
const (
	PaymentStatusPaid    = "paid"
	PaymentStatusPayable = "payable" // المتجر يدين للبائع
	PaymentStatusPartial = "partial"
	PaymentStatusOverdue = "overdue"
)

// Acquisition represents an acquisition (from supplier or customer)
// Based on USED-PARTS-ACQUISITION.md
type Acquisition struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Type            string    `json:"type" db:"type"` // SUPPLIER, CUSTOMER
	AcquisitionDate time.Time `json:"acquisition_date" db:"acquisition_date"`

	// Seller Information (can be supplier or customer)
	SupplierID *uuid.UUID `json:"supplier_id" db:"supplier_id"` // للشراء من مورد
	CustomerID *uuid.UUID `json:"customer_id" db:"customer_id"` // للشراء من عميل

	// Financial Information
	TotalCost     float64 `json:"total_cost" db:"total_cost"`
	PaidAmount    float64 `json:"paid_amount" db:"paid_amount"`
	PaymentStatus string  `json:"payment_status" db:"payment_status"` // paid, payable, partial, overdue

	// Status & Workflow
	Status string `json:"status" db:"status"` // draft, pending, acquired, inspection, approved, rejected, cancelled, reversed

	// Notes & Metadata
	Notes     *string    `json:"notes" db:"notes"`
	UserID    *uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`

	// Reversal tracking (ARCHITECTURE-PRINCIPLES.md - Reverse instead of Delete)
	ReversedAt     *time.Time        `json:"reversed_at" db:"reversed_at"`
	ReversedBy     *uuid.UUID        `json:"reversed_by" db:"reversed_by"`
	ReversalReason *string           `json:"reversal_reason" db:"reversal_reason"`
	Items          []AcquisitionItem `json:"items,omitempty" db:"-"`
}

// AcquisitionItem represents an item in an acquisition
type AcquisitionItem struct {
	ID            uuid.UUID `json:"id" db:"id"`
	AcquisitionID uuid.UUID `json:"acquisition_id" db:"acquisition_id"`
	ProductID     uuid.UUID `json:"product_id" db:"product_id"`
	SerialNumber  string    `json:"serial_number" db:"serial_number"`
	Condition     string    `json:"condition" db:"condition"` // new, used, refurbished
	Grade         string    `json:"grade" db:"grade"`         // excellent, very_good, good, fair, poor
	UnitCost      float64   `json:"unit_cost" db:"unit_cost"`
	TotalCost     float64   `json:"total_cost" db:"total_cost"`

	// Inspection Status
	InspectionID     *uuid.UUID `json:"inspection_id" db:"inspection_id"`
	InspectionStatus string     `json:"inspection_status" db:"inspection_status"` // pending, passed, failed, needs_repair

	// Inventory Status
	InventoryItemID *uuid.UUID `json:"inventory_item_id" db:"inventory_item_id"`
	ItemStatus      string     `json:"item_status" db:"item_status"` // acquired, inspection, available, sold, etc.

	Notes     string    `json:"notes" db:"notes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AcquisitionRequest represents acquisition creation request
type AcquisitionRequest struct {
	Type            string                   `json:"type" binding:"required,oneof=SUPPLIER CUSTOMER"`
	AcquisitionDate time.Time                `json:"acquisition_date" binding:"required"`
	SupplierID      *uuid.UUID               `json:"supplier_id"` // مطلوب إذا كان Type = SUPPLIER
	CustomerID      *uuid.UUID               `json:"customer_id"` // مطلوب إذا كان Type = CUSTOMER
	Items           []AcquisitionItemRequest `json:"items" binding:"required,min=1"`
	Notes           string                   `json:"notes"`
	PaymentStatus   string                   `json:"payment_status" binding:"omitempty,oneof=paid payable partial"`
}

// AcquisitionItemRequest represents acquisition item creation request
type AcquisitionItemRequest struct {
	ProductID    uuid.UUID  `json:"product_id" binding:"required"`
	PartTypeID   *uuid.UUID `json:"part_type_id"`
	SerialNumber string     `json:"serial_number"`
	Condition    string     `json:"condition" binding:"required,oneof=new used refurbished"`
	Grade        string     `json:"grade" binding:"omitempty,oneof=excellent very_good good fair poor"`
	UnitCost     float64    `json:"unit_cost" binding:"required,min=0"`
	SellingPrice float64    `json:"selling_price" binding:"required,min=0"`
	Notes        string     `json:"notes"`
}

// AcquisitionUpdateRequest represents acquisition update request
type AcquisitionUpdateRequest struct {
	AcquisitionDate time.Time `json:"acquisition_date"`
	Status          string    `json:"status" binding:"omitempty,oneof=draft pending acquired inspection approved rejected cancelled reversed"`
	Notes           string    `json:"notes"`
	PaymentStatus   string    `json:"payment_status" binding:"omitempty,oneof=paid payable partial overdue"`
}

// AcquisitionResponse represents acquisition response with related data
type AcquisitionResponse struct {
	Acquisition Acquisition       `json:"acquisition"`
	Items       []AcquisitionItem `json:"items"`
	Seller      *SellerInfo       `json:"seller,omitempty"` // يمكن أن يكون Supplier أو Customer
	TotalItems  int               `json:"total_items"`
	Remaining   float64           `json:"remaining"` // المبلغ المتبقي إذا كان payable
}

// SellerInfo represents seller information (can be supplier or customer)
type SellerInfo struct {
	ID    uuid.UUID `json:"id"`
	Type  string    `json:"type"` // SUPPLIER, CUSTOMER
	Name  string    `json:"name"`
	Phone string    `json:"phone"`
	Email *string   `json:"email,omitempty"`
}

// AcquisitionListRequest represents acquisition list query parameters
type AcquisitionListRequest struct {
	Page          int        `form:"page" binding:"min=1"`
	PerPage       int        `form:"per_page" binding:"min=1,max=100"`
	Type          string     `form:"type" binding:"omitempty,oneof=SUPPLIER CUSTOMER"`
	SupplierID    *uuid.UUID `form:"supplier_id"`
	CustomerID    *uuid.UUID `form:"customer_id"`
	Status        string     `form:"status" binding:"omitempty,oneof=draft pending acquired inspection approved rejected cancelled reversed"`
	PaymentStatus string     `form:"payment_status" binding:"omitempty,oneof=paid payable partial overdue"`
	StartDate     *time.Time `form:"start_date"`
	EndDate       *time.Time `form:"end_date"`
	Search        string     `form:"search"`
	SortBy        string     `form:"sort_by"`
	SortOrder     string     `form:"sort_order"`
}

// AcquisitionReversal represents a reversal of an acquisition (ARCHITECTURE-PRINCIPLES.md)
type AcquisitionReversal struct {
	ID                     uuid.UUID   `json:"id" db:"id"`
	AcquisitionID          uuid.UUID   `json:"acquisition_id" db:"acquisition_id"`
	Reason                 string      `json:"reason" db:"reason"`
	ReversedBy             uuid.UUID   `json:"reversed_by" db:"reversed_by"`
	ReversedAt             time.Time   `json:"reversed_at" db:"reversed_at"`
	OriginalTotal          float64     `json:"original_total" db:"original_total"`
	InventoryAdjustmentIDs []uuid.UUID `json:"inventory_adjustment_ids" db:"inventory_adjustment_ids"`
	CreatedAt              time.Time   `json:"created_at" db:"created_at"`
}

// AcquisitionReversalRequest represents a request to reverse an acquisition
type AcquisitionReversalRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// SellerPayment represents a payment to a seller (customer who sold items to store)
type SellerPayment struct {
	ID            uuid.UUID `json:"id" db:"id"`
	AcquisitionID uuid.UUID `json:"acquisition_id" db:"acquisition_id"`
	CustomerID    uuid.UUID `json:"customer_id" db:"customer_id"`
	Amount        float64   `json:"amount" db:"amount"`
	PaymentMethod string    `json:"payment_method" db:"payment_method"` // cash, transfer, etc.
	PaymentDate   time.Time `json:"payment_date" db:"payment_date"`
	Notes         *string   `json:"notes" db:"notes"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// SellerPaymentRequest represents a payment to seller request
type SellerPaymentRequest struct {
	AcquisitionID uuid.UUID `json:"acquisition_id" binding:"required"`
	CustomerID    uuid.UUID `json:"customer_id" binding:"required"`
	Amount        float64   `json:"amount" binding:"required,min=0"`
	PaymentMethod string    `json:"payment_method" binding:"required"`
	PaymentDate   time.Time `json:"payment_date" binding:"required"`
	Notes         string    `json:"notes"`
}

// ItemAging represents aging information for used items
type ItemAging struct {
	ItemID          uuid.UUID `json:"item_id" db:"item_id"`
	AcquisitionID   uuid.UUID `json:"acquisition_id" db:"acquisition_id"`
	AcquisitionDate time.Time `json:"acquisition_date" db:"acquisition_date"`
	DaysInStock     int       `json:"days_in_stock" db:"days_in_stock"`
	Status          string    `json:"status" db:"status"`
	Condition       string    `json:"condition" db:"condition"`
	Cost            float64   `json:"cost" db:"cost"`
	CurrentPrice    float64   `json:"current_price" db:"current_price"`
	AgingCategory   string    `json:"aging_category" db:"aging_category"` // fresh, normal, aged, long_aged
	AlertLevel      string    `json:"alert_level" db:"alert_level"`       // none, warning, critical
}

// SellerBalance represents balance information for a seller (customer who sold items)
type SellerBalance struct {
	CustomerID        uuid.UUID  `json:"customer_id" db:"customer_id"`
	CustomerName      string     `json:"customer_name" db:"customer_name"`
	CustomerCode      string     `json:"customer_code" db:"customer_code"`
	CustomerPhone     string     `json:"customer_phone" db:"customer_phone"`
	CustomerEmail     *string    `json:"customer_email,omitempty" db:"customer_email"`
	TotalAcquisitions int        `json:"total_acquisitions" db:"total_acquisitions"`
	TotalAcquired     float64    `json:"total_acquired" db:"total_acquired"`
	TotalPaid         float64    `json:"total_paid" db:"total_paid"`
	Balance           float64    `json:"balance" db:"balance"`
	TransactionCount  int        `json:"transaction_count" db:"transaction_count"`
	LastTransaction   *time.Time `json:"last_transaction" db:"last_transaction"`
}
