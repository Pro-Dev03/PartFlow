package returns

import (
	"github.com/google/uuid"
)

// ValidateReturnItem validates a return item
func ValidateReturnItem(item *ReturnItemRequest) error {
	if item.SaleItemID == uuid.Nil {
		return ErrSaleItemNotFound
	}
	if item.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if item.Condition != "new" && item.Condition != "used" && item.Condition != "damaged" {
		return ErrInvalidCondition
	}
	return nil
}

// ValidateReturnStatus validates return status
func ValidateReturnStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":   true,
		"approved":  true,
		"rejected":  true,
		"completed": true,
	}
	
	if !validStatuses[status] {
		return ErrInvalidReturnStatus
	}
	return nil
}

// ValidateRefundMethod validates refund method
func ValidateRefundMethod(method string) error {
	validMethods := map[string]bool{
		"cash":           true,
		"card":           true,
		"bank_transfer":  true,
		"store_credit":   true,
	}
	
	if !validMethods[method] {
		return ErrInvalidRefundMethod
	}
	return nil
}

// ValidateCondition validates condition
func ValidateCondition(condition string) error {
	validConditions := map[string]bool{
		"new":      true,
		"used":     true,
		"damaged":  true,
	}
	
	if !validConditions[condition] {
		return ErrInvalidCondition
	}
	return nil
}

// ValidateRefundAmount validates refund amount
func ValidateRefundAmount(amount float64) error {
	if amount < 0 {
		return ErrInvalidQuantity
	}
	return nil
}
