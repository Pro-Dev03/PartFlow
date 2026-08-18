package warranties

import (
	"time"

	"github.com/google/uuid"
)

// Warranty represents a product warranty
type Warranty struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProductID        uuid.UUID  `json:"product_id" db:"product_id"`
	SerialNumber     string     `json:"serial_number" db:"serial_number"`
	WarrantyNumber   string     `json:"warranty_number" db:"warranty_number"`
	WarrantyType     string     `json:"warranty_type" db:"warranty_type"` // manufacturer, seller, extended
	WarrantyPeriod   int        `json:"warranty_period" db:"warranty_period"` // in days
	StartDate        time.Time  `json:"start_date" db:"start_date"`
	EndDate          time.Time  `json:"end_date" db:"end_date"`
	Status           string     `json:"status" db:"status"` // active, expired, claimed, voided
	CustomerID       *uuid.UUID `json:"customer_id" db:"customer_id"`
	SaleID           *uuid.UUID `json:"sale_id" db:"sale_id"`
	Terms            string     `json:"terms" db:"terms"`
	Notes            string     `json:"notes" db:"notes"`
	CreatedBy        uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// WarrantyClaim represents a warranty claim
type WarrantyClaim struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id" db:"organization_id"`
	WarrantyID       uuid.UUID  `json:"warranty_id" db:"warranty_id"`
	ClaimNumber      string     `json:"claim_number" db:"claim_number"`
	ClaimDate        time.Time  `json:"claim_date" db:"claim_date"`
	CustomerID       uuid.UUID  `json:"customer_id" db:"customer_id"`
	IssueDescription string     `json:"issue_description" db:"issue_description"`
	Status           string     `json:"status" db:"status"` // pending, approved, rejected, in_progress, completed
	Resolution       string     `json:"resolution" db:"resolution"`
	ResolutionDate   *time.Time `json:"resolution_date" db:"resolution_date"`
	ApprovedBy       *uuid.UUID `json:"approved_by" db:"approved_by"`
	ApprovedDate     *time.Time `json:"approved_date" db:"approved_date"`
	RejectedBy       *uuid.UUID `json:"rejected_by" db:"rejected_by"`
	RejectedDate     *time.Time `json:"rejected_date" db:"rejected_date"`
	CompletedBy      *uuid.UUID `json:"completed_by" db:"completed_by"`
	CompletedDate    *time.Time `json:"completed_date" db:"completed_date"`
	Notes            string     `json:"notes" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// WarrantyRequest represents warranty creation request
type WarrantyRequest struct {
	ProductID      uuid.UUID  `json:"product_id" binding:"required"`
	SerialNumber   string     `json:"serial_number"`
	WarrantyType   string     `json:"warranty_type" binding:"required,oneof=manufacturer seller extended"`
	WarrantyPeriod int        `json:"warranty_period" binding:"required,min=1"`
	StartDate      time.Time  `json:"start_date" binding:"required"`
	CustomerID     *uuid.UUID `json:"customer_id"`
	SaleID         *uuid.UUID `json:"sale_id"`
	Terms          string     `json:"terms"`
	Notes          string     `json:"notes"`
}

// WarrantyUpdateRequest represents warranty update request
type WarrantyUpdateRequest struct {
	SerialNumber   string     `json:"serial_number"`
	WarrantyType   string     `json:"warranty_type" binding:"omitempty,oneof=manufacturer seller extended"`
	WarrantyPeriod int        `json:"warranty_period" binding:"omitempty,min=1"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	Status         string     `json:"status" binding:"omitempty,oneof=active expired claimed voided"`
	CustomerID     *uuid.UUID `json:"customer_id"`
	Terms          string     `json:"terms"`
	Notes          string     `json:"notes"`
}

// WarrantyClaimRequest represents warranty claim creation request
type WarrantyClaimRequest struct {
	WarrantyID       uuid.UUID `json:"warranty_id" binding:"required"`
	CustomerID       uuid.UUID `json:"customer_id" binding:"required"`
	IssueDescription string   `json:"issue_description" binding:"required"`
	Notes            string   `json:"notes"`
}

// WarrantyClaimUpdateRequest represents warranty claim update request
type WarrantyClaimUpdateRequest struct {
	IssueDescription string    `json:"issue_description"`
	Status           string    `json:"status" binding:"omitempty,oneof=pending approved rejected in_progress completed"`
	Resolution       string    `json:"resolution"`
	ResolutionDate   time.Time `json:"resolution_date"`
	Notes            string    `json:"notes"`
}

// WarrantyResponse represents warranty response with related data
type WarrantyResponse struct {
	Warranty Warranty            `json:"warranty"`
	Product  *ProductInfo        `json:"product,omitempty"`
	Customer *CustomerInfo       `json:"customer,omitempty"`
	Claims   []WarrantyClaim     `json:"claims,omitempty"`
}

// WarrantyClaimResponse represents warranty claim response with related data
type WarrantyClaimResponse struct {
	Claim    WarrantyClaim  `json:"claim"`
	Warranty *Warranty      `json:"warranty,omitempty"`
	Customer *CustomerInfo  `json:"customer,omitempty"`
}

// ProductInfo represents product information
type ProductInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Model       string    `json:"model"`
	SKU         string    `json:"sku"`
	Barcode     string    `json:"barcode"`
}

// CustomerInfo represents customer information
type CustomerInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Phone string    `json:"phone"`
	Email string    `json:"email"`
}

// WarrantyListRequest represents warranty list query parameters
type WarrantyListRequest struct {
	Page         int        `form:"page" binding:"min=1"`
	PerPage      int        `form:"per_page" binding:"min=1,max=100"`
	ProductID    *uuid.UUID `form:"product_id"`
	CustomerID   *uuid.UUID `form:"customer_id"`
	SaleID       *uuid.UUID `form:"sale_id"`
	Status       string     `form:"status" binding:"omitempty,oneof=active expired claimed voided"`
	WarrantyType string     `form:"warranty_type" binding:"omitempty,oneof=manufacturer seller extended"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	Search       string     `form:"search"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}

// WarrantyClaimListRequest represents warranty claim list query parameters
type WarrantyClaimListRequest struct {
	Page          int        `form:"page" binding:"min=1"`
	PerPage       int        `form:"per_page" binding:"min=1,max=100"`
	WarrantyID   *uuid.UUID `form:"warranty_id"`
	CustomerID   *uuid.UUID `form:"customer_id"`
	Status       string     `form:"status" binding:"omitempty,oneof=pending approved rejected in_progress completed"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	Search       string     `form:"search"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}

// WarrantyExpiringSoon represents warranties that will expire soon
type WarrantyExpiringSoon struct {
	WarrantyID     uuid.UUID `json:"warranty_id"`
	ProductID      uuid.UUID `json:"product_id"`
	ProductName    string    `json:"product_name"`
	SerialNumber   string    `json:"serial_number"`
	CustomerName   string    `json:"customer_name"`
	EndDate        time.Time `json:"end_date"`
	DaysRemaining  int       `json:"days_remaining"`
}
