package returns

import (
	"github.com/google/uuid"
)

// ValidateReturnItem validates a return item
func ValidateReturnItem(item *ReturnItemRequest) error {
	if item.SaleItemID != nil && *item.SaleItemID == uuid.Nil {
		return ErrSaleItemNotFound
	}
	if item.QuantityReturned <= 0 {
		return ErrInvalidQuantity
	}
	if item.ReturnedCondition != "NEW" && item.ReturnedCondition != "USED" && item.ReturnedCondition != "DAMAGED" &&
		item.ReturnedCondition != "DEFECTIVE" && item.ReturnedCondition != "OPEN_BOX" && item.ReturnedCondition != "REFURBISHED" {
		return ErrInvalidCondition
	}
	return nil
}

// ValidateReturnStatus validates return status
func ValidateReturnStatus(status string) error {
	validStatuses := map[string]bool{
		"PENDING":    true,
		"APPROVED":   true,
		"PROCESSING": true,
		"COMPLETED":  true,
		"REJECTED":   true,
		"CANCELLED":  true,
	}

	if !validStatuses[status] {
		return ErrInvalidReturnStatus
	}
	return nil
}

// ValidateRefundMethod validates refund method
func ValidateRefundMethod(method string) error {
	validMethods := map[string]bool{
		"CASH":            true,
		"CREDIT":          true,
		"DEBT_ADJUSTMENT": true,
		"EXCHANGE":        true,
		"BANK_TRANSFER":   true,
		"STORE_CREDIT":    true,
	}

	if !validMethods[method] {
		return ErrInvalidRefundMethod
	}
	return nil
}

// ValidateCondition validates condition
func ValidateCondition(condition string) error {
	validConditions := map[string]bool{
		"SELLABLE":         true,
		"NEEDS_INSPECTION": true,
		"NEEDS_REPAIR":     true,
		"DAMAGED":          true,
		"USED":             true,
		"REFURBISHED":      true,
		"SUPPLIER_RETURN":  true,
		"WRITE_OFF":        true,
		"PARTS":            true,
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
