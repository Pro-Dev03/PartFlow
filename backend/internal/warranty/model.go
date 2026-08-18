package warranty

import (
	"time"

	"github.com/google/uuid"
)

// WarrantyClaim represents a warranty claim
type WarrantyClaim struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	ClaimNumber    string     `json:"claim_number" db:"claim_number"`
	ProductID      uuid.UUID  `json:"product_id" db:"product_id"`
	SerialNumber   string     `json:"serial_number" db:"serial_number"`
	CustomerID     *uuid.UUID `json:"customer_id,omitempty" db:"customer_id"`
	SaleID         *uuid.UUID `json:"sale_id,omitempty" db:"sale_id"`
	ClaimDate      time.Time  `json:"claim_date" db:"claim_date"`
	ClaimType      string     `json:"claim_type" db:"claim_type"` // repair, replacement, refund
	Reason         string     `json:"reason" db:"reason"`
	Description    string     `json:"description" db:"description"`
	Status         string     `json:"status" db:"status"` // pending, approved, rejected, in_progress, completed
	Priority       string     `json:"priority" db:"priority"` // low, medium, high, urgent
	EstimatedCost  float64    `json:"estimated_cost" db:"estimated_cost"`
	ActualCost     float64    `json:"actual_cost" db:"actual_cost"`
	Resolution     *string    `json:"resolution,omitempty" db:"resolution"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	AssignedTo     *uuid.UUID `json:"assigned_to,omitempty" db:"assigned_to"`
	Notes          string     `json:"notes" db:"notes"`
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// WarrantyClaimItem represents an item in a warranty claim
type WarrantyClaimItem struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ClaimID         uuid.UUID  `json:"claim_id" db:"claim_id"`
	ProductID       uuid.UUID  `json:"product_id" db:"product_id"`
	SerialNumber    string     `json:"serial_number" db:"serial_number"`
	Quantity        int        `json:"quantity" db:"quantity"`
	Condition       string     `json:"condition" db:"condition"` // new, used, damaged
	DefectType      string     `json:"defect_type" db:"defect_type"` // manufacturing, user_damage, other
	DefectDescription string   `json:"defect_description" db:"defect_description"`
	IsRepaired      bool       `json:"is_repaired" db:"is_repaired"`
	RepairedAt      *time.Time `json:"repaired_at,omitempty" db:"repaired_at"`
	RepairCost      float64    `json:"repair_cost" db:"repair_cost"`
	Notes           string     `json:"notes" db:"notes"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// Warranty represents product warranty information
type Warranty struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	ProductID      uuid.UUID  `json:"product_id" db:"product_id"`
	WarrantyType   string     `json:"warranty_type" db:"warranty_type"` // manufacturer, extended, store
	DurationDays   int        `json:"duration_days" db:"duration_days"`
	StartDate      time.Time  `json:"start_date" db:"start_date"`
	EndDate        time.Time  `json:"end_date" db:"end_date"`
	Terms          string     `json:"terms" db:"terms"`
	Coverage       string     `json:"coverage" db:"coverage"` // parts, labor, both
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for the WarrantyClaim model
func (WarrantyClaim) TableName() string {
	return "warranty_claims"
}

// TableName returns the table name for the WarrantyClaimItem model
func (WarrantyClaimItem) TableName() string {
	return "warranty_claim_items"
}

// TableName returns the table name for the Warranty model
func (Warranty) TableName() string {
	return "warranties"
}

// NewWarrantyClaim creates a new WarrantyClaim instance
func NewWarrantyClaim(organizationID uuid.UUID, productID uuid.UUID, serialNumber string, claimType string, reason string, userID uuid.UUID) *WarrantyClaim {
	return &WarrantyClaim{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		ClaimNumber:    generateClaimNumber(),
		ProductID:      productID,
		SerialNumber:   serialNumber,
		ClaimDate:      time.Now(),
		ClaimType:      claimType,
		Reason:         reason,
		Status:         "pending",
		Priority:       "medium",
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// NewWarranty creates a new Warranty instance
func NewWarranty(organizationID uuid.UUID, productID uuid.UUID, warrantyType string, durationDays int, coverage string) *Warranty {
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, durationDays)
	
	return &Warranty{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		ProductID:      productID,
		WarrantyType:   warrantyType,
		DurationDays:   durationDays,
		StartDate:      startDate,
		EndDate:        endDate,
		Coverage:       coverage,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// generateClaimNumber generates a unique claim number
func generateClaimNumber() string {
	timestamp := time.Now().Format("20060102150405")
	return "WC-" + timestamp
}

// WarrantyClaimRequest represents warranty claim creation request
type WarrantyClaimRequest struct {
	ProductID      uuid.UUID  `json:"product_id" binding:"required"`
	SerialNumber   string     `json:"serial_number" binding:"required"`
	CustomerID     *uuid.UUID `json:"customer_id,omitempty"`
	SaleID         *uuid.UUID `json:"sale_id,omitempty"`
	ClaimType      string     `json:"claim_type" binding:"required,oneof=repair replacement refund"`
	Reason         string     `json:"reason" binding:"required"`
	Description    string     `json:"description"`
	Priority       string     `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	EstimatedCost  float64    `json:"estimated_cost" binding:"omitempty,min=0"`
	AssignedTo     *uuid.UUID `json:"assigned_to,omitempty"`
	Notes          string     `json:"notes"`
	Items          []WarrantyClaimItemRequest `json:"items" binding:"required,min=1"`
}

// WarrantyClaimItemRequest represents warranty claim item creation request
type WarrantyClaimItemRequest struct {
	ProductID       uuid.UUID `json:"product_id" binding:"required"`
	SerialNumber    string    `json:"serial_number" binding:"required"`
	Quantity        int       `json:"quantity" binding:"required,min=1"`
	Condition       string    `json:"condition" binding:"required,oneof=new used damaged"`
	DefectType      string    `json:"defect_type" binding:"required,oneof=manufacturing user_damage other"`
	DefectDescription string  `json:"defect_description" binding:"required"`
	Notes           string    `json:"notes"`
}

// WarrantyClaimUpdateRequest represents warranty claim update request
type WarrantyClaimUpdateRequest struct {
	Status         string     `json:"status" binding:"omitempty,oneof=pending approved rejected in_progress completed"`
	Priority       string     `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	EstimatedCost  float64    `json:"estimated_cost" binding:"omitempty,min=0"`
	ActualCost     float64    `json:"actual_cost" binding:"omitempty,min=0"`
	Resolution     *string    `json:"resolution,omitempty"`
	AssignedTo     *uuid.UUID `json:"assigned_to,omitempty"`
	Notes          string     `json:"notes"`
}

// WarrantyRequest represents warranty creation request
type WarrantyRequest struct {
	ProductID    uuid.UUID `json:"product_id" binding:"required"`
	WarrantyType string     `json:"warranty_type" binding:"required,oneof=manufacturer extended store"`
	DurationDays int        `json:"duration_days" binding:"required,min=1"`
	Terms        string     `json:"terms"`
	Coverage     string     `json:"coverage" binding:"required,oneof=parts labor both"`
}

// WarrantyClaimResponse represents warranty claim response with related data
type WarrantyClaimResponse struct {
	Claim      WarrantyClaim     `json:"claim"`
	Items      []WarrantyClaimItem `json:"items"`
	Product    *ProductInfo      `json:"product,omitempty"`
	Customer   *CustomerInfo     `json:"customer,omitempty"`
	AssignedTo *UserInfo         `json:"assigned_to,omitempty"`
	CreatedBy  *UserInfo         `json:"created_by,omitempty"`
	TotalItems int               `json:"total_items"`
}

// ProductInfo represents product information
type ProductInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Model string    `json:"model"`
	SKU   string    `json:"sku"`
}

// CustomerInfo represents customer information
type CustomerInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Phone string    `json:"phone"`
	Email string    `json:"email"`
}

// UserInfo represents user information
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

// WarrantyClaimListRequest represents warranty claim list query parameters
type WarrantyClaimListRequest struct {
	Page         int        `form:"page" binding:"min=1"`
	PerPage      int        `form:"per_page" binding:"min=1,max=100"`
	ProductID    *uuid.UUID `form:"product_id"`
	CustomerID   *uuid.UUID `form:"customer_id"`
	Status       string     `form:"status" binding:"omitempty,oneof=pending approved rejected in_progress completed"`
	Priority     string     `form:"priority" binding:"omitempty,oneof=low medium high urgent"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	Search       string     `form:"search"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}