package inventory

import (
	"time"

	"github.com/google/uuid"
)

// Condition represents the condition of an inventory item
type Condition string

const (
	ConditionNew        Condition = "NEW"
	ConditionUsed       Condition = "USED"
	ConditionRefurbished Condition = "REFURBISHED"
	ConditionDamaged    Condition = "DAMAGED"
	ConditionForParts   Condition = "FOR_PARTS"
)

// Grade represents the grade of a used item
type Grade string

const (
	GradeExcellent Grade = "EXCELLENT"
	GradeVeryGood  Grade = "VERY_GOOD"
	GradeGood      Grade = "GOOD"
	GradeFair      Grade = "FAIR"
	GradePoor      Grade = "POOR"
)

// Status represents the status of an inventory item
type Status string

const (
	StatusPurchased    Status = "PURCHASED"
	StatusReceived     Status = "RECEIVED"
	StatusInspection   Status = "INSPECTION"
	StatusAvailable    Status = "AVAILABLE"
	StatusReserved     Status = "RESERVED"
	StatusSold         Status = "SOLD"
	StatusDamaged      Status = "DAMAGED"
	StatusInRepair     Status = "IN_REPAIR"
	StatusReturned     Status = "RETURNED"
	StatusWarranty     Status = "WARRANTY"
	StatusForParts     Status = "FOR_PARTS"
	StatusArchived     Status = "ARCHIVED"
)

// InventoryItem represents an individual inventory item
type InventoryItem struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProductID      uuid.UUID  `json:"product_id" db:"product_id"`
	ItemCode       string     `json:"item_code" db:"item_code"`
	Barcode        string     `json:"barcode" db:"barcode"`
	SerialNumber   string     `json:"serial_number" db:"serial_number"`
	Condition      Condition  `json:"condition" db:"condition"`
	Grade          Grade      `json:"grade" db:"grade"`
	PurchaseCost   int64      `json:"purchase_cost" db:"purchase_cost"` // in minor units
	SellingPrice   int64      `json:"selling_price" db:"selling_price"` // in minor units
	Status         Status     `json:"status" db:"status"`
	LocationID     *uuid.UUID `json:"location_id" db:"location_id"`
	SupplierID     *uuid.UUID `json:"supplier_id" db:"supplier_id"`
	PurchaseDate   *time.Time `json:"purchase_date" db:"purchase_date"`
	SoldAt         *time.Time `json:"sold_at" db:"sold_at"`
	Notes          string     `json:"notes" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Location represents a storage location
type Location struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Type           string     `json:"type" db:"type"` // warehouse, shelf, box, display
	ParentID       *uuid.UUID `json:"parent_id" db:"parent_id"`
	WarehouseID    *uuid.UUID `json:"warehouse_id" db:"warehouse_id"`
	Description    string     `json:"description" db:"description"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// MovementType represents the type of inventory movement
type MovementType string

const (
	MovementPurchase    MovementType = "PURCHASE"
	MovementSale        MovementType = "SALE"
	MovementReturn      MovementType = "RETURN"
	MovementAdjustment  MovementType = "ADJUSTMENT"
	MovementTransfer    MovementType = "TRANSFER"
	MovementReservation MovementType = "RESERVATION"
	MovementRelease     MovementType = "RELEASE"
	MovementDamage      MovementType = "DAMAGE"
	MovementRepair      MovementType = "REPAIR"
)

// InventoryMovement represents a movement in inventory
type InventoryMovement struct {
	ID             uuid.UUID    `json:"id" db:"id"`
	OrganizationID uuid.UUID    `json:"organization_id" db:"organization_id"`
	ItemID         *uuid.UUID   `json:"item_id" db:"item_id"`
	ProductID      *uuid.UUID   `json:"product_id" db:"product_id"`
	MovementType   MovementType `json:"movement_type" db:"movement_type"`
	Quantity       int          `json:"quantity" db:"quantity"`
	BeforeQuantity int          `json:"before_quantity" db:"before_quantity"`
	AfterQuantity  int          `json:"after_quantity" db:"after_quantity"`
	ReferenceType  string       `json:"reference_type" db:"reference_type"` // sale, purchase, return, etc.
	ReferenceID    *uuid.UUID   `json:"reference_id" db:"reference_id"`
	Reason         string       `json:"reason" db:"reason"`
	CreatedBy      uuid.UUID    `json:"created_by" db:"created_by"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
}

// Reservation represents an item reservation
type Reservation struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	ItemID         uuid.UUID  `json:"item_id" db:"item_id"`
	CustomerID     *uuid.UUID `json:"customer_id" db:"customer_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	ReservedAt     time.Time  `json:"reserved_at" db:"reserved_at"`
	ExpiresAt      time.Time  `json:"expires_at" db:"expires_at"`
	Status         string     `json:"status" db:"status"` // active, expired, converted, cancelled
	Notes          string     `json:"notes" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// InventoryItemRequest represents inventory item creation/update request
type InventoryItemRequest struct {
	ProductID    uuid.UUID  `json:"product_id" binding:"required"`
	ItemCode     string     `json:"item_code"`
	Barcode      string     `json:"barcode"`
	SerialNumber string     `json:"serial_number"`
	Condition    Condition  `json:"condition" binding:"required"`
	Grade        Grade      `json:"grade"`
	PurchaseCost int64      `json:"purchase_cost" binding:"required"`
	SellingPrice int64      `json:"selling_price" binding:"required"`
	LocationID   *uuid.UUID `json:"location_id"`
	SupplierID   *uuid.UUID `json:"supplier_id"`
	Notes        string     `json:"notes"`
}

// LocationRequest represents location creation/update request
type LocationRequest struct {
	Name        string     `json:"name" binding:"required"`
	Type        string     `json:"type" binding:"required"`
	ParentID    *uuid.UUID `json:"parent_id"`
	WarehouseID *uuid.UUID `json:"warehouse_id"`
	Description string     `json:"description"`
}

// MovementRequest represents inventory movement request
type MovementRequest struct {
	ItemID        *uuid.UUID   `json:"item_id"`
	ProductID     *uuid.UUID   `json:"product_id"`
	MovementType  MovementType `json:"movement_type" binding:"required"`
	Quantity      int          `json:"quantity" binding:"required"`
	ReferenceType string       `json:"reference_type"`
	ReferenceID   *uuid.UUID   `json:"reference_id"`
	Reason        string       `json:"reason"`
}

// AdjustmentRequest represents inventory adjustment request
type AdjustmentRequest struct {
	ItemID        uuid.UUID   `json:"item_id" binding:"required"`
	NewQuantity   int         `json:"new_quantity" binding:"required"`
	Reason        string      `json:"reason" binding:"required"`
}

// TransferRequest represents inventory transfer request
type TransferRequest struct {
	ItemID     uuid.UUID `json:"item_id" binding:"required"`
	FromLocationID uuid.UUID `json:"from_location_id" binding:"required"`
	ToLocationID   uuid.UUID `json:"to_location_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required"`
	Reason    string    `json:"reason"`
}

// ReservationRequest represents reservation request
type ReservationRequest struct {
	ItemID     uuid.UUID  `json:"item_id" binding:"required"`
	CustomerID *uuid.UUID `json:"customer_id"`
	ExpiresIn  int       `json:"expires_in"` // minutes
	Notes      string    `json:"notes"`
}
