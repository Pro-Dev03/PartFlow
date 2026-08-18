package returns

import (
	"time"

	"github.com/google/uuid"
)

// ToReturnResponse converts Return to ReturnResponse
func (r *Return) ToReturnResponse(items []ReturnItem, customer *CustomerInfo, sale *SaleInfo) *ReturnResponse {
	return &ReturnResponse{
		Return:     *r,
		Items:      items,
		Customer:   customer,
		Sale:       sale,
		TotalItems: len(items),
	}
}

// ToReturnListItem converts Return to list item format
func (r *Return) ToReturnListItem(itemCount int, customerName string, saleInvoiceNumber string) map[string]interface{} {
	return map[string]interface{}{
		"id":                r.ID,
		"return_number":    r.ReturnNumber,
		"return_date":      r.ReturnDate,
		"customer_name":    customerName,
		"sale_invoice":     saleInvoiceNumber,
		"refund_amount":    r.RefundAmount,
		"refund_method":    r.RefundMethod,
		"status":           r.Status,
		"reason":           r.Reason,
		"total_items":      itemCount,
		"created_at":       r.CreatedAt,
	}
}

// CreateReturn creates a Return from request
func CreateReturn(organizationID uuid.UUID, userID uuid.UUID, req *ReturnRequest) *Return {
	return &Return{
		ID:           uuid.New(),
		OrganizationID: organizationID,
		SaleID:       req.SaleID,
		ReturnNumber: generateReturnNumber(),
		ReturnDate:   req.ReturnDate,
		Reason:       req.Reason,
		Condition:    req.Condition,
		Status:       "pending",
		RefundAmount: 0,
		RefundMethod: req.RefundMethod,
		Notes:        req.Notes,
		ProcessedBy:  userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// CreateReturnItem creates a ReturnItem from request
func CreateReturnItem(returnID uuid.UUID, req ReturnItemRequest, unitPrice float64) *ReturnItem {
	totalPrice := float64(req.Quantity) * unitPrice
	return &ReturnItem{
		ID:           uuid.New(),
		ReturnID:     returnID,
		SaleItemID:   req.SaleItemID,
		Quantity:     req.Quantity,
		UnitPrice:    unitPrice,
		TotalPrice:   totalPrice,
		Reason:       req.Reason,
		Condition:    req.Condition,
		IsResellable: req.IsResellable,
		LocationID:   req.LocationID,
		Notes:        req.Notes,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// ValidateReturnRequest validates return request
func ValidateReturnRequest(req *ReturnRequest) error {
	if req.SaleID == uuid.Nil {
		return ErrSaleNotFound
	}
	if req.Reason == "" {
		return ErrReturnNotFound
	}
	if req.Condition != "new" && req.Condition != "used" && req.Condition != "damaged" {
		return ErrInvalidCondition
	}
	if len(req.Items) == 0 {
		return ErrNoItems
	}
	if req.RefundMethod != "cash" && req.RefundMethod != "card" && 
		req.RefundMethod != "bank_transfer" && req.RefundMethod != "store_credit" {
		return ErrInvalidRefundMethod
	}
	for _, item := range req.Items {
		if item.SaleItemID == uuid.Nil {
			return ErrSaleItemNotFound
		}
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if item.Condition != "new" && item.Condition != "used" && item.Condition != "damaged" {
			return ErrInvalidCondition
		}
	}
	return nil
}

// generateReturnNumber generates a unique return number
func generateReturnNumber() string {
	return "RET-" + uuid.New().String()[:8]
}
