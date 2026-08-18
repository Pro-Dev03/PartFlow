package suppliers

import "errors"

var (
	// ErrSupplierNotFound is returned when a supplier is not found
	ErrSupplierNotFound = errors.New("supplier not found")

	// ErrSupplierCodeExists is returned when a supplier code already exists
	ErrSupplierCodeExists = errors.New("supplier code already exists")

	// ErrInvalidSupplierData is returned when supplier data is invalid
	ErrInvalidSupplierData = errors.New("invalid supplier data")

	// ErrPaymentAmountInvalid is returned when payment amount is invalid
	ErrPaymentAmountInvalid = errors.New("payment amount must be greater than zero")

	// ErrCreditLimitExceeded is returned when credit limit is exceeded
	ErrCreditLimitExceeded = errors.New("credit limit exceeded")

	// ErrCreditLimitBelowBalance is returned when credit limit is set below current balance
	ErrCreditLimitBelowBalance = errors.New("credit limit cannot be set below current balance")
)