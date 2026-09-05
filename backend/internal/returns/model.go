package returns

import (
	"time"

	"github.com/google/uuid"
)

// Return represents a product return
type Return struct {
	ID              uuid.UUID `json:"id" db:"id"`
	ReturnNumber    string    `json:"return_number" db:"return_number"`
	ReferenceNumber string    `json:"reference_number" db:"reference_number"`

	// Source information
	SaleID       uuid.UUID `json:"sale_id" db:"sale_id"`
	PurchaseID   uuid.UUID `json:"purchase_id" db:"purchase_id"`
	CustomerID   uuid.UUID `json:"customer_id" db:"customer_id"`
	CustomerName string    `json:"customer_name,omitempty" db:"customer_name"`

	// Return details
	ReturnDate time.Time `json:"return_date" db:"return_date"`
	ReturnType string    `json:"return_type" db:"return_type"` // FULL, PARTIAL, QUANTITY_PARTIAL
	Status     string    `json:"status" db:"status"`           // PENDING, APPROVED, PROCESSING, COMPLETED, REJECTED, CANCELLED

	// Financial details
	TotalRefundAmount float64    `json:"total_refund_amount" db:"total_refund_amount"`
	RefundMethod      string     `json:"refund_method" db:"refund_method"` // CASH, CREDIT, DEBT_ADJUSTMENT, EXCHANGE, BANK_TRANSFER, STORE_CREDIT
	RefundDate        *time.Time `json:"refund_date" db:"refund_date"`
	RefundReference   string     `json:"refund_reference" db:"refund_reference"`

	// Debt integration
	DebtID         *uuid.UUID `json:"debt_id" db:"debt_id"`
	DebtAdjustment float64    `json:"debt_adjustment" db:"debt_adjustment"`
	CustomerCredit float64    `json:"customer_credit" db:"customer_credit"`

	// Return reason and condition
	Reason                   string `json:"reason" db:"reason"` // DEFECTIVE, WRONG_ITEM, COMPATIBILITY_ISSUE, CUSTOMER_CHANGED_MIND, DAMAGED, WARRANTY, INCORRECT_SPECIFICATION, OTHER
	ReasonDetail             string `json:"reason_detail" db:"reason_detail"`
	ItemConditionAfterReturn string `json:"item_condition_after_return" db:"item_condition_after_return"` // SELLABLE, NEEDS_INSPECTION, NEEDS_REPAIR, DAMAGED, USED, REFURBISHED, SUPPLIER_RETURN, WRITE_OFF, PARTS

	// Warranty information
	IsWarrantyClaim    bool       `json:"is_warranty_claim" db:"is_warranty_claim"`
	WarrantyID         *uuid.UUID `json:"warranty_id" db:"warranty_id"`
	WarrantyValidUntil *time.Time `json:"warranty_valid_until" db:"warranty_valid_until"`

	// Approval workflow
	CreatedBy   *uuid.UUID `json:"created_by" db:"created_by"`
	ProcessedBy *uuid.UUID `json:"processed_by" db:"processed_by"`
	ApprovedBy  *uuid.UUID `json:"approved_by" db:"approved_by"`
	ApprovedAt  *time.Time `json:"approved_at" db:"approved_at"`

	// Notes and audit
	Notes         string    `json:"notes" db:"notes"`
	InternalNotes string    `json:"internal_notes" db:"internal_notes"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// ReturnItem represents an item in a return
type ReturnItem struct {
	ID       uuid.UUID `json:"id" db:"id"`
	ReturnID uuid.UUID `json:"return_id" db:"return_id"`

	// Item identification
	SaleItemID      *uuid.UUID `json:"sale_item_id" db:"sale_item_id"`
	ProductID       *uuid.UUID `json:"product_id" db:"product_id"`
	InventoryItemID *uuid.UUID `json:"inventory_item_id" db:"inventory_item_id"`
	SerialNumber    string     `json:"serial_number" db:"serial_number"`
	Barcode         string     `json:"barcode" db:"barcode"`

	// Quantity and pricing
	QuantityReturned  int     `json:"quantity_returned" db:"quantity_returned"`
	OriginalQuantity  *int    `json:"original_quantity" db:"original_quantity"`
	UnitPrice         float64 `json:"unit_price" db:"unit_price"`
	TotalRefundAmount float64 `json:"total_refund_amount" db:"total_refund_amount"`

	// Item condition
	OriginalCondition string `json:"original_condition" db:"original_condition"`
	ReturnedCondition string `json:"returned_condition" db:"returned_condition"` // NEW, USED, DAMAGED, DEFECTIVE, OPEN_BOX, REFURBISHED
	ConditionNotes    string `json:"condition_notes" db:"condition_notes"`

	// Resolution
	Resolution      string `json:"resolution" db:"resolution"`             // RESTOCK, REPAIR, SUPPLIER_RETURN, WRITE_OFF, PARTS, REPLACEMENT
	InventoryStatus string `json:"inventory_status" db:"inventory_status"` // RETURNED, INSPECTION, REPAIRING, RESTOCKED, SUPPLIER_RETURNED, WRITTEN_OFF, DISMANTLED

	// Inspection details
	InspectionRequired bool       `json:"inspection_required" db:"inspection_required"`
	InspectionDate     *time.Time `json:"inspection_date" db:"inspection_date"`
	InspectionResult   string     `json:"inspection_result" db:"inspection_result"` // PASSED, FAILED, PENDING
	InspectionNotes    string     `json:"inspection_notes" db:"inspection_notes"`

	// Cost tracking
	OriginalCost *float64 `json:"original_cost" db:"original_cost"`
	RepairCost   float64  `json:"repair_cost" db:"repair_cost"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ReturnRequest represents return creation request
type ReturnRequest struct {
	SaleID                   *uuid.UUID          `json:"sale_id"`
	PurchaseID               *uuid.UUID          `json:"purchase_id"`
	CustomerID               *uuid.UUID          `json:"customer_id"`
	ReturnDate               time.Time           `json:"return_date" binding:"required"`
	ReturnType               string              `json:"return_type" binding:"required,oneof=FULL PARTIAL QUANTITY_PARTIAL"`
	Reason                   string              `json:"reason" binding:"required,oneof=DEFECTIVE WRONG_ITEM COMPATIBILITY_ISSUE CUSTOMER_CHANGED_MIND DAMAGED WARRANTY INCORRECT_SPECIFICATION OTHER"`
	ReasonDetail             string              `json:"reason_detail"`
	ItemConditionAfterReturn string              `json:"item_condition_after_return" binding:"required,oneof=SELLABLE NEEDS_INSPECTION NEEDS_REPAIR DAMAGED USED REFURBISHED SUPPLIER_RETURN WRITE_OFF PARTS"`
	Items                    []ReturnItemRequest `json:"items" binding:"required,min=1"`
	RefundMethod             string              `json:"refund_method" binding:"required,oneof=CASH CREDIT DEBT_ADJUSTMENT EXCHANGE BANK_TRANSFER STORE_CREDIT"`
	DebtID                   *uuid.UUID          `json:"debt_id"`
	Notes                    string              `json:"notes"`
	InternalNotes            string              `json:"internal_notes"`
	IsWarrantyClaim          bool                `json:"is_warranty_claim"`
	WarrantyID               *uuid.UUID          `json:"warranty_id"`
}

// ReturnItemRequest represents return item creation request
type ReturnItemRequest struct {
	SaleItemID         *uuid.UUID `json:"sale_item_id" binding:"required"`
	ProductID          *uuid.UUID `json:"product_id"`
	InventoryItemID    *uuid.UUID `json:"inventory_item_id"`
	SerialNumber       string     `json:"serial_number"`
	Barcode            string     `json:"barcode"`
	QuantityReturned   int        `json:"quantity_returned" binding:"required,min=1"`
	UnitPrice          float64    `json:"unit_price" binding:"required,min=0"`
	TotalRefundAmount  float64    `json:"total_refund_amount" binding:"required,min=0"`
	OriginalCondition  string     `json:"original_condition"`
	ReturnedCondition  string     `json:"returned_condition" binding:"required,oneof=NEW USED DAMAGED DEFECTIVE OPEN_BOX REFURBISHED"`
	ConditionNotes     string     `json:"condition_notes"`
	Resolution         string     `json:"resolution" binding:"omitempty,oneof=RESTOCK REPAIR SUPPLIER_RETURN WRITE_OFF PARTS REPLACEMENT"`
	InspectionRequired bool       `json:"inspection_required"`
	OriginalCost       *float64   `json:"original_cost"`
	RepairCost         float64    `json:"repair_cost"`
}

// ReturnUpdateRequest represents return update request
type ReturnUpdateRequest struct {
	Status                   string     `json:"status" binding:"omitempty,oneof=PENDING APPROVED PROCESSING COMPLETED REJECTED CANCELLED"`
	Reason                   string     `json:"reason" binding:"omitempty,oneof=DEFECTIVE WRONG_ITEM COMPATIBILITY_ISSUE CUSTOMER_CHANGED_MIND DAMAGED WARRANTY INCORRECT_SPECIFICATION OTHER"`
	TotalRefundAmount        float64    `json:"total_refund_amount" binding:"omitempty,min=0"`
	RefundMethod             string     `json:"refund_method" binding:"omitempty,oneof=CASH CREDIT DEBT_ADJUSTMENT EXCHANGE BANK_TRANSFER STORE_CREDIT"`
	RefundDate               *time.Time `json:"refund_date"`
	RefundReference          string     `json:"refund_reference"`
	DebtID                   *uuid.UUID `json:"debt_id"`
	DebtAdjustment           float64    `json:"debt_adjustment"`
	ItemConditionAfterReturn string     `json:"item_condition_after_return" binding:"omitempty,oneof=SELLABLE NEEDS_INSPECTION NEEDS_REPAIR DAMAGED USED REFURBISHED SUPPLIER_RETURN WRITE_OFF PARTS"`
	Resolution               string     `json:"resolution" binding:"omitempty,oneof=RESTOCK REPAIR SUPPLIER_RETURN WRITE_OFF PARTS REPLACEMENT"`
	Notes                    string     `json:"notes"`
	InternalNotes            string     `json:"internal_notes"`
	ApprovedBy               *uuid.UUID `json:"approved_by"`
}

// ReturnResponse represents return response with related data
type ReturnResponse struct {
	Return     Return        `json:"return"`
	Items      []ReturnItem  `json:"items"`
	Customer   *CustomerInfo `json:"customer,omitempty"`
	Sale       *SaleInfo     `json:"sale,omitempty"`
	TotalItems int           `json:"total_items"`
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
	ID            uuid.UUID `json:"id" db:"id"`
	InvoiceNumber string    `json:"invoice_number" db:"invoice_number"`
	SaleDate      time.Time `json:"sale_date" db:"sale_date"`
	TotalAmount   float64   `json:"total_amount" db:"total_amount"`
	CustomerID    uuid.UUID `json:"customer_id" db:"customer_id"`
}

// ReturnListRequest represents return list query parameters
type ReturnListRequest struct {
	Page         int        `form:"page" binding:"min=1"`
	PerPage      int        `form:"per_page" binding:"min=1,max=100"`
	CustomerID   *uuid.UUID `form:"customer_id"`
	SaleID       *uuid.UUID `form:"sale_id"`
	Status       string     `form:"status" binding:"omitempty,oneof=PENDING APPROVED PROCESSING COMPLETED REJECTED CANCELLED"`
	ReturnType   string     `form:"return_type" binding:"omitempty,oneof=FULL PARTIAL QUANTITY_PARTIAL"`
	RefundMethod string     `form:"refund_method" binding:"omitempty,oneof=CASH CREDIT DEBT_ADJUSTMENT EXCHANGE BANK_TRANSFER STORE_CREDIT"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	Search       string     `form:"search"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}

// MonthlyReturnsAnalysis represents monthly returns analysis data
type MonthlyReturnsAnalysis struct {
	Month                  time.Time `json:"month" db:"month"`
	TotalReturns           int       `json:"total_returns" db:"total_returns"`
	UniqueCustomers        int       `json:"unique_customers" db:"unique_customers"`
	TotalRefundAmount      float64   `json:"total_refund_amount" db:"total_refund_amount"`
	AvgRefundAmount        float64   `json:"avg_refund_amount" db:"avg_refund_amount"`
	FullReturns            int       `json:"full_returns" db:"full_returns"`
	PartialReturns         int       `json:"partial_returns" db:"partial_returns"`
	QuantityPartialReturns int       `json:"quantity_partial_returns" db:"quantity_partial_returns"`
	DefectiveReturns       int       `json:"defective_returns" db:"defective_returns"`
	WarrantyReturns        int       `json:"warranty_returns" db:"warranty_returns"`
	WarrantyClaims         int       `json:"warranty_claims" db:"warranty_claims"`
	SellableItems          int       `json:"sellable_items" db:"sellable_items"`
	RepairNeeded           int       `json:"repair_needed" db:"repair_needed"`
	WrittenOff             int       `json:"written_off" db:"written_off"`
}

// SalesReturnsAnalysis represents sales vs returns analysis data
type SalesReturnsAnalysis struct {
	Month         time.Time `json:"month" db:"month"`
	TotalSales    int       `json:"total_sales" db:"total_sales"`
	GrossSales    float64   `json:"gross_sales" db:"gross_sales"`
	TotalCost     float64   `json:"total_cost" db:"total_cost"`
	GrossProfit   float64   `json:"gross_profit" db:"gross_profit"`
	ReturnsAmount float64   `json:"returns_amount" db:"returns_amount"`
	ReturnCount   int       `json:"return_count" db:"return_count"`
	NetSales      float64   `json:"net_sales" db:"net_sales"`
}

// ReturnInspectionRequest represents return item inspection request
type ReturnInspectionRequest struct {
	InspectionDate   time.Time `json:"inspection_date" binding:"required"`
	InspectionResult string    `json:"inspection_result" binding:"required,oneof=PASSED FAILED PENDING"`
	InspectionNotes  string    `json:"inspection_notes"`
	Resolution       string    `json:"resolution" binding:"required,oneof=RESTOCK REPAIR SUPPLIER_RETURN WRITE_OFF PARTS REPLACEMENT"`
	RepairCost       float64   `json:"repair_cost" binding:"omitempty,min=0"`
}
