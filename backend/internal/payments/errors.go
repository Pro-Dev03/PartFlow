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

	// ErrPaymentAllocationHistoryMissing means a legacy account payment has no
	// recorded debt allocation map, so reversing it by guessing is unsafe.
	ErrPaymentAllocationHistoryMissing = errors.New("payment debt allocation history is missing")

	// ErrPaymentAllocationTrackingUnavailable means the schema needed to record
	// or reverse allocation history has not been migrated yet.
	ErrPaymentAllocationTrackingUnavailable = errors.New("payment allocation tracking schema is unavailable; apply the payment allocation migration")

	// ErrPaymentHistoryInconsistent means linked financial rows do not agree
	// with the payment and the transaction was left unchanged.
	ErrPaymentHistoryInconsistent = errors.New("payment financial history is inconsistent")

	// ErrPaymentRequiresProviderRefund means an external provider still reports
	// money as captured or pending and must be refunded/cancelled first.
	ErrPaymentRequiresProviderRefund = errors.New("payment requires provider refund or cancellation")
)
