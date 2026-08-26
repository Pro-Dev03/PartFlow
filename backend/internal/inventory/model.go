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
	StatusForParts     Status = "FOR_PARTS"
	StatusArchived     Status = "ARCHIVED"
)

// InventoryItem represents an individual inventory item
type InventoryItem struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ProductID      *uuid.UUID `json:"product_id" db:"product_id"`
	PartTypeID     *uuid.UUID `json:"part_type_id" db:"part_type_id"`
	ItemCode       *string    `json:"item_code" db:"item_code"`
	Barcode        *string    `json:"barcode" db:"barcode"`
	SerialNumber   *string    `json:"serial_number" db:"serial_number"`
	Condition      string     `json:"condition" db:"condition"`
	Grade          *string    `json:"grade" db:"grade"`
	PurchaseCost   float64    `json:"purchase_cost" db:"purchase_cost"`
	SellingPrice   float64    `json:"selling_price" db:"selling_price"`
	Status         string     `json:"status" db:"status"`
	LocationID     *uuid.UUID `json:"location_id" db:"location_id"`
	SupplierID     *uuid.UUID `json:"supplier_id" db:"supplier_id"`
	PurchaseDate   *time.Time `json:"purchase_date" db:"purchase_date"`
	SoldAt         *time.Time `json:"sold_at" db:"sold_at"`
	Notes          *string    `json:"notes" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`

	// Current State fields (جديدة - للعمليات اليومية - ARCHITECTURE-PRINCIPLES.md)
	// هذه الحالات تُحفظ لتجنب إعادة الحساب من كل التاريخ
	CurrentQuantity    int        `json:"current_quantity" db:"current_quantity"`             // الكمية الحالية (للمنتجات)
	ReservedQuantity   int        `json:"reserved_quantity" db:"reserved_quantity"`           // المحجوز
	AvailableQuantity  int        `json:"available_quantity" db:"available_quantity"`         // المتاح = Current - Reserved
	CurrentCost        float64    `json:"current_cost" db:"current_cost"`                     // التكلفة الحالية
	CurrentValue       float64    `json:"current_value" db:"current_value"`                   // القيمة الحالية = Quantity * Cost
	LastMovementID     *uuid.UUID `json:"last_movement_id" db:"last_movement_id"`             // آخر حركة
}

// DBInventoryItem is a simplified struct for database scanning
type DBInventoryItem struct {
	ID             string    `db:"id"`
	ProductID      *string   `db:"product_id"`
	PartTypeID     *string   `db:"part_type_id"`
	ItemCode       *string   `db:"item_code"`
	Barcode        *string   `db:"barcode"`
	SerialNumber   *string   `db:"serial_number"`
	Condition      string    `db:"condition"`
	Grade          *string   `db:"grade"`
	PurchaseCost   float64   `db:"purchase_cost"`
	SellingPrice   float64   `db:"selling_price"`
	Status         string    `db:"status"`
	LocationID     *string   `db:"location_id"`
	SupplierID     *string   `db:"supplier_id"`
	PurchaseDate   *time.Time `db:"purchase_date"`
	SoldAt         *time.Time `db:"sold_at"`
	Notes          *string   `db:"notes"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// Location represents a storage location
type Location struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Type           string     `json:"type" db:"type"` // warehouse, shelf, box, display
	ParentID       *uuid.UUID `json:"parent_id" db:"parent_id"`
	WarehouseID    *uuid.UUID `json:"warehouse_id" db:"warehouse_id"`
	Description    *string    `json:"description" db:"description"`
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

	// Reverse movements (للعمليات العكسية - ARCHITECTURE-PRINCIPLES.md)
	MovementReversePurchase MovementType = "REVERSE_PURCHASE"
	MovementReverseSale     MovementType = "REVERSE_SALE"
	MovementReverseReturn   MovementType = "REVERSE_RETURN"
)

// InventoryMovement represents a movement in inventory
type InventoryMovement struct {
	ID             uuid.UUID    `json:"id" db:"id"`
	ItemID         *uuid.UUID   `json:"item_id" db:"item_id"`
	ProductID      *uuid.UUID   `json:"product_id" db:"product_id"`
	MovementType   MovementType `json:"movement_type" db:"movement_type"`
	Quantity       int          `json:"quantity" db:"quantity"`
	BeforeQuantity int          `json:"before_quantity" db:"before_quantity"`
	AfterQuantity  int          `json:"after_quantity" db:"after_quantity"`
	ReferenceType  string       `json:"reference_type" db:"reference_type"` // sale, purchase, return, etc.
	ReferenceID    *uuid.UUID   `json:"reference_id" db:"reference_id"`
	Reason         *string      `json:"reason" db:"reason"`
	CreatedBy      uuid.UUID    `json:"created_by" db:"created_by"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`

	// Enhanced fields for Immutable History (ARCHITECTURE-PRINCIPLES.md)
	CostBefore     float64    `json:"cost_before" db:"cost_before"`               // التكلفة قبل الحركة
	CostAfter      float64    `json:"cost_after" db:"cost_after"`                 // التكلفة بعد الحركة
	ValueBefore    float64    `json:"value_before" db:"value_before"`             // القيمة قبل الحركة
	ValueAfter     float64    `json:"value_after" db:"value_after"`               // القيمة بعد الحركة
	IsReversed     bool       `json:"is_reversed" db:"is_reversed"`             // هل تم عكس هذه الحركة؟
	ReversedBy     *uuid.UUID  `json:"reversed_by" db:"reversed_by"`             // من قام بالعكس
	ReversedAt     *time.Time  `json:"reversed_at" db:"reversed_at"`             // متى تم العكس
	ReversalReason *string     `json:"reversal_reason" db:"reversal_reason"`     // سبب العكس
}

// Reservation represents an item reservation
type Reservation struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ItemID         uuid.UUID  `json:"item_id" db:"item_id"`
	CustomerID     *uuid.UUID `json:"customer_id" db:"customer_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	ReservedAt     time.Time  `json:"reserved_at" db:"reserved_at"`
	ExpiresAt      time.Time  `json:"expires_at" db:"expires_at"`
	Status         string     `json:"status" db:"status"` // active, expired, converted, cancelled
	Notes          *string    `json:"notes" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// InventoryItemRequest represents inventory item creation/update request
type InventoryItemRequest struct {
	ProductID    *uuid.UUID `json:"product_id"`
	PartTypeID   *uuid.UUID `json:"part_type_id"`
	ItemCode     *string    `json:"item_code"`
	Barcode      *string    `json:"barcode"`
	SerialNumber *string    `json:"serial_number"`
	Condition    Condition  `json:"condition" binding:"required"`
	Grade        *Grade     `json:"grade"`
	PurchaseCost float64    `json:"purchase_cost" binding:"required"`
	SellingPrice float64    `json:"selling_price" binding:"required"`
	Status       Status     `json:"status"`
	LocationID   *uuid.UUID `json:"location_id"`
	SupplierID   *uuid.UUID `json:"supplier_id"`
	Notes        *string    `json:"notes"`
}

// LocationRequest represents location creation/update request
type LocationRequest struct {
	Name        string     `json:"name" binding:"required"`
	Type        string     `json:"type" binding:"required"`
	ParentID    *uuid.UUID `json:"parent_id"`
	WarehouseID *uuid.UUID `json:"warehouse_id"`
	Description *string    `json:"description"`
}

// MovementRequest represents inventory movement request
type MovementRequest struct {
	ItemID        *uuid.UUID   `json:"item_id"`
	ProductID     *uuid.UUID   `json:"product_id"`
	MovementType  MovementType `json:"movement_type" binding:"required"`
	Quantity      int          `json:"quantity" binding:"required"`
	ReferenceType string       `json:"reference_type"`
	ReferenceID   *uuid.UUID   `json:"reference_id"`
	Reason        *string      `json:"reason"`
}

// AdjustmentRequest represents inventory adjustment request
type AdjustmentRequest struct {
	ItemID        uuid.UUID   `json:"item_id" binding:"required"`
	NewQuantity   int         `json:"new_quantity" binding:"required"`
	NewStatus     *string     `json:"new_status"`
	Reason        *string     `json:"reason"`
}

// TransferRequest represents inventory transfer request
type TransferRequest struct {
	ItemID     uuid.UUID `json:"item_id" binding:"required"`
	FromLocationID uuid.UUID `json:"from_location_id" binding:"required"`
	ToLocationID   uuid.UUID `json:"to_location_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required"`
	Reason    *string   `json:"reason"`
}

// ReservationRequest represents reservation request
type ReservationRequest struct {
	ItemID     uuid.UUID  `json:"item_id" binding:"required"`
	CustomerID *uuid.UUID `json:"customer_id"`
	ExpiresIn  int       `json:"expires_in"` // minutes
	Notes      *string   `json:"notes"`
}

// InventoryItemWithSupplier represents an inventory item with supplier and product join info
type InventoryItemWithSupplier struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	ProductID    *uuid.UUID `json:"product_id" db:"product_id"`
	PartTypeID   *uuid.UUID `json:"part_type_id" db:"part_type_id"`
	ItemCode     *string    `json:"item_code" db:"item_code"`
	Barcode      *string    `json:"barcode" db:"barcode"`
	SerialNumber *string    `json:"serial_number" db:"serial_number"`
	Condition    string     `json:"condition" db:"condition"`
	Grade        *string    `json:"grade" db:"grade"`
	PurchaseCost float64    `json:"purchase_cost" db:"purchase_cost"`
	SellingPrice float64    `json:"selling_price" db:"selling_price"`
	Status       string     `json:"status" db:"status"`
	LocationID   *uuid.UUID `json:"location_id" db:"location_id"`
	SupplierID   *uuid.UUID `json:"supplier_id" db:"supplier_id"`
	PurchaseDate *time.Time `json:"purchase_date" db:"purchase_date"`
	SoldAt       *time.Time `json:"sold_at" db:"sold_at"`
	Notes        *string    `json:"notes" db:"notes"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	ProductName  *string    `json:"product_name" db:"product_name"`
	SupplierName *string    `json:"supplier_name" db:"supplier_name"`
	SupplierPhone *string   `json:"supplier_phone" db:"supplier_phone"`
}
