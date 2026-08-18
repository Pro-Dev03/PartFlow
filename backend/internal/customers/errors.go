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

	// ErrInvalidPaymentMethod is returned when payment method is invalid
	ErrInvalidPaymentMethod = errors.New("invalid payment method")

	// ErrCreditLimitExceeded is returned when credit limit is exceeded
	ErrCreditLimitExceeded = errors.New("credit limit exceeded")

	// ErrCreditLimitBelowBalance is returned when credit limit is set below current balance
	ErrCreditLimitBelowBalance = errors.New("credit limit cannot be set below current balance")
)
