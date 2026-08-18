package payments

import "errors"

var (
	// ErrPaymentNotFound is returned when a payment is not found
	ErrPaymentNotFound = errors.New("payment not found")

	// ErrInvalidPaymentType is returned when payment type is invalid
	ErrInvalidPaymentType = errors.New("invalid payment type")

	// ErrInvalidPaymentMethod is returned when payment method is invalid
	ErrInvalidPaymentMethod = errors.New("invalid payment method")

	// ErrInvalidAmount is returned when payment amount is invalid
	ErrInvalidAmount = errors.New("invalid payment amount")

	// ErrPaymentAlreadyProcessed is returned when payment is already processed
	ErrPaymentAlreadyProcessed = errors.New("payment already processed")

	// ErrPaymentCannotBeCancelled is returned when payment cannot be cancelled
	ErrPaymentCannotBeCancelled = errors.New("payment cannot be cancelled")
)