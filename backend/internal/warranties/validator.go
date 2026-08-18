package warranties

import (
	"github.com/google/uuid"
)

// ValidateWarrantyType validates warranty type
func ValidateWarrantyType(warrantyType string) error {
	validTypes := map[string]bool{
		"manufacturer": true,
		"seller":       true,
		"extended":     true,
	}
	
	if !validTypes[warrantyType] {
		return ErrInvalidWarrantyType
	}
	return nil
}

// ValidateWarrantyStatus validates warranty status
func ValidateWarrantyStatus(status string) error {
	validStatuses := map[string]bool{
		"active":   true,
		"expired":  true,
		"claimed":  true,
		"voided":   true,
	}
	
	if !validStatuses[status] {
		return ErrInvalidWarrantyStatus
	}
	return nil
}

// ValidateClaimStatus validates claim status
func ValidateClaimStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":     true,
		"approved":    true,
		"rejected":    true,
		"in_progress": true,
		"completed":   true,
	}
	
	if !validStatuses[status] {
		return ErrInvalidClaimStatus
	}
	return nil
}

// ValidateWarrantyPeriod validates warranty period
func ValidateWarrantyPeriod(period int) error {
	if period <= 0 {
		return ErrInvalidWarrantyPeriod
	}
	return nil
}

// ValidateProductID validates product ID
func ValidateProductID(productID uuid.UUID) error {
	if productID == uuid.Nil {
		return ErrProductNotFound
	}
	return nil
}

// ValidateCustomerID validates customer ID
func ValidateCustomerID(customerID uuid.UUID) error {
	if customerID == uuid.Nil {
		return ErrCustomerNotFound
	}
	return nil
}

// ValidateWarrantyID validates warranty ID
func ValidateWarrantyID(warrantyID uuid.UUID) error {
	if warrantyID == uuid.Nil {
		return ErrWarrantyNotFound
	}
	return nil
}
