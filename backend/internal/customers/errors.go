package customers

import "errors"

var (
	// ErrCustomerNotFound is returned when a customer is not found
	ErrCustomerNotFound = errors.New("customer not found")

	// ErrCustomerCodeExists is returned when a customer code already exists
	ErrCustomerCodeExists = errors.New("customer code already exists")

	// ErrInvalidCustomerData is returned when customer data is invalid
	ErrInvalidCustomerData = errors.New("invalid customer data")

	// ErrPaymentAmountInvalid is returned when payment amount is invalid
	ErrPaymentAmountInvalid = errors.New("payment amount must be greater than zero")

	// ErrPaymentExceedsBalance is returned when payment exceeds balance
	ErrPaymentExceedsBalance = errors.New("payment amount exceeds customer balance")

	// ErrPaymentDuplicate is returned when a payment reference was already used
	ErrPaymentDuplicate = errors.New("duplicate customer payment reference")

	// ErrInvalidPaymentMethod is returned when payment method is invalid
	ErrInvalidPaymentMethod = errors.New("invalid payment method")

	// ErrInvalidDebtAdjustmentProduct is returned when an optional product
	// attached to a debt adjustment does not exist or has an invalid quantity.
	ErrInvalidDebtAdjustmentProduct = errors.New("invalid debt adjustment product")

	// ErrCreditLimitExceeded is returned when credit limit is exceeded
	ErrCreditLimitExceeded = errors.New("credit limit exceeded")

	// ErrCreditLimitBelowBalance is returned when credit limit is set below current balance
	ErrCreditLimitBelowBalance = errors.New("credit limit cannot be set below current balance")

	// ErrCustomerHasOutstandingDebt is returned when trying to delete a customer with debt
	ErrCustomerHasOutstandingDebt = errors.New("cannot delete customer with outstanding debt")

	// ErrCustomerHasActiveTransactions is returned when trying to delete a customer with active transactions
	ErrCustomerHasActiveTransactions = errors.New("cannot delete customer with active transactions")

	// ErrCustomerHasActiveWarranties is returned when trying to delete a customer with active warranties
	ErrCustomerHasActiveWarranties = errors.New("cannot delete customer with active warranties")
)
