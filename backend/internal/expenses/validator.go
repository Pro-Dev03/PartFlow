package expenses

import (
	"github.com/google/uuid"
)

// ValidateExpenseAmount validates expense amount
func ValidateExpenseAmount(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// ValidateExpenseStatus validates expense status
func ValidateExpenseStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":  true,
		"approved": true,
		"rejected": true,
	}
	
	if !validStatuses[status] {
		return ErrInvalidExpenseStatus
	}
	return nil
}

// ValidatePaymentMethod validates payment method
func ValidatePaymentMethod(method string) error {
	validMethods := map[string]bool{
		"cash":          true,
		"card":          true,
		"bank_transfer": true,
		"check":         true,
	}
	
	if !validMethods[method] {
		return ErrInvalidPaymentMethod
	}
	return nil
}

// ValidateRecurringPeriod validates recurring period
func ValidateRecurringPeriod(period string) error {
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
		"yearly":  true,
	}
	
	if !validPeriods[period] {
		return ErrInvalidRecurringPeriod
	}
	return nil
}

// ValidateCurrency validates currency code
func ValidateCurrency(currency string) error {
	if currency == "" {
		return ErrInvalidCurrency
	}
	// Add more validation if needed (e.g., check against ISO 4217)
	return nil
}

// ValidateCategoryID validates category ID
func ValidateCategoryID(categoryID uuid.UUID) error {
	if categoryID == uuid.Nil {
		return ErrExpenseCategoryNotFound
	}
	return nil
}
