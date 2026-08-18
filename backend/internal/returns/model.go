package returns

import (
	"time"

	"github.com/google/uuid"
)

// Return represents a product return
type Return struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	SaleID         uuid.UUID  `json:"sale_id" db:"sale_id"`
	CustomerID     uuid.UUID  `json:"customer_id" db:"customer_id"`
	ReturnNumber   string     `json:"return_number" db:"return_number"`
	ReturnDate     time.Time  `json:"return_date" db:"return_date"`
	Reason         string     `json:"reason" db:"reason"`
	Condition      string     `json:"condition" db:"condition"` // new, used, damaged
	Status         string     `json:"status" db:"status"` // pending, approved, rejected, completed
	RefundAmount   float64    `json:"refund_amount" db:"refund_amount"`
	RefundMethod   string     `json:"refund_method" db:"refund_method"` // cash, card, bank_transfer, store_credit
	RefundDate     *time.Time `json:"refund_date" db:"refund_date"`
	Notes          string     `json:"notes" db:"notes"`
	ProcessedBy    uuid.UUID  `json:"processed_by" db:"processed_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// ReturnItem represents an item in a return
type ReturnItem struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ReturnID       uuid.UUID  `json:"return_id" db:"return_id"`
	SaleItemID     uuid.UUID  `json:"sale_item_id" db:"sale_item_id"`
	ProductID      uuid.UUID  `json:"product_id" db:"product_id"`
	Quantity       int        `json:"quantity" db:"quantity"`
	UnitPrice      float64    `json:"unit_price" db:"unit_price"`
	TotalPrice     float64    `json:"total_price" db:"total_price"`
	Reason         string     `json:"reason" db:"reason"`
	Condition      string     `json:"condition" db:"condition"` // new, used, damaged
	IsResellable   bool       `json:"is_resellable" db:"is_resellable"`
	LocationID     *uuid.UUID `json:"location_id" db:"location_id"`
	Notes          string     `json:"notes" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// ReturnRequest represents return creation request
type ReturnRequest struct {
	SaleID       uuid.UUID           `json:"sale_id" binding:"required"`
	ReturnDate   time.Time           `json:"return_date" binding:"required"`
	Reason       string              `json:"reason" binding:"required"`
	Condition    string              `json:"condition" binding:"required,oneof=new used damaged"`
	Items        []ReturnItemRequest `json:"items" binding:"required,min=1"`
	RefundMethod string              `json:"refund_method" binding:"required,oneof=cash card bank_transfer store_credit"`
	Notes        string              `json:"notes"`
}

// ReturnItemRequest represents return item creation request
type ReturnItemRequest struct {
	SaleItemID   uuid.UUID  `json:"sale_item_id" binding:"required"`
	Quantity     int        `json:"quantity" binding:"required,min=1"`
	Reason       string     `json:"reason" binding:"required"`
	Condition    string     `json:"condition" binding:"required,oneof=new used damaged"`
	IsResellable bool       `json:"is_resellable"`
	LocationID   *uuid.UUID `json:"location_id"`
	Notes        string     `json:"notes"`
}

// ReturnUpdateRequest represents return update request
type ReturnUpdateRequest struct {
	ReturnDate   time.Time  `json:"return_date"`
	Reason       string     `json:"reason"`
	Condition    string     `json:"condition" binding:"omitempty,oneof=new used damaged"`
	Status       string     `json:"status" binding:"omitempty,oneof=pending approved rejected completed"`
	RefundAmount float64    `json:"refund_amount" binding:"omitempty,min=0"`
	RefundMethod string     `json:"refund_method" binding:"omitempty,oneof=cash card bank_transfer store_credit"`
	Notes        string     `json:"notes"`
}

// ReturnResponse represents return response with related data
type ReturnResponse struct {
	Return     Return       `json:"return"`
	Items      []ReturnItem `json:"items"`
	Customer   *CustomerInfo `json:"customer,omitempty"`
	Sale       *SaleInfo    `json:"sale,omitempty"`
	TotalItems int          `json:"total_items"`
}

// CustomerInfo represents customer information
type CustomerInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Phone string    `json:"phone"`
	Email string    `json:"email"`
}

// SaleInfo represents sale information
type SaleInfo struct {
	ID           uuid.UUID `json:"id"`
	InvoiceNumber string   `json:"invoice_number"`
	SaleDate     time.Time `json:"sale_date"`
	TotalAmount  float64   `json:"total_amount"`
}

// ReturnListRequest represents return list query parameters
type ReturnListRequest struct {
	Page         int        `form:"page" binding:"min=1"`
	PerPage      int        `form:"per_page" binding:"min=1,max=100"`
	CustomerID   *uuid.UUID `form:"customer_id"`
	SaleID       *uuid.UUID `form:"sale_id"`
	Status       string     `form:"status" binding:"omitempty,oneof=pending approved rejected completed"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	Search       string     `form:"search"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}
