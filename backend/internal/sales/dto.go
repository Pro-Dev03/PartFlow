package sales

import "github.com/google/uuid"

// CreateSaleRequest represents the request to create a sale
type CreateSaleRequest struct {
	CustomerID    *uuid.UUID        `json:"customer_id,omitempty"`
	Items         []SaleItemRequest `json:"items" binding:"required"`
	PaymentMethod *string           `json:"payment_method,omitempty"`
	PaymentAmount float64           `json:"payment_amount,omitempty"`
	Notes         *string           `json:"notes,omitempty"`
	TaxRate       float64           `json:"tax_rate" binding:"required"`
	DiscountType  string            `json:"discount_type"` // "percentage" or "fixed"
	DiscountValue float64           `json:"discount_value"`
}

// SaleItemRequest represents an item in a sale request
type SaleItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
	UnitPrice float64   `json:"unit_price" binding:"required,min=0"`
}

// UpdatePaymentRequest represents the request to update payment
type UpdatePaymentRequest struct {
	Amount        float64 `json:"amount" binding:"required,min=0"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
}

// SalesListRequest represents the request to list sales
type SalesListRequest struct {
	Page       int    `form:"page" binding:"min=1"`
	PerPage    int    `form:"per_page" binding:"min=1,max=100"`
	Status     string `form:"status"`
	CustomerID string `form:"customer_id"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
}

// SaleResponse represents the response for a sale
type SaleResponse struct {
	Sale   *Sale      `json:"sale"`
	Items  []SaleItem `json:"items"`
	Profit float64    `json:"profit"`
}