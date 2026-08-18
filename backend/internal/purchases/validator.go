package purchases

import (
	"github.com/google/uuid"
)

// ValidatePurchaseItem validates a purchase item
func ValidatePurchaseItem(item *PurchaseItemRequest) error {
	if item.ProductID == uuid.Nil {
		return ErrProductNotFound
	}
	if item.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if item.UnitCost < 0 {
		return ErrInvalidCost
	}
	if item.Condition != "new" && item.Condition != "used" && item.Condition != "refurbished" {
		return ErrInvalidCondition
	}
	return nil
}

// ValidatePurchaseStatus validates purchase status
func ValidatePurchaseStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":  true,
		"received": true,
		"cancelled": true,
	}
	
	if !validStatuses[status] {
		return ErrInvalidPurchaseStatus
	}
	return nil
}

// ValidatePaymentAmount validates payment amount
func ValidatePaymentAmount(amount float64) error {
	if amount <= 0 {
		return ErrInvalidCost
	}
	return nil
}
