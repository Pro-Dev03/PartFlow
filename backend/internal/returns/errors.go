package returns

import "errors"

var (
	// ErrReturnNotFound is returned when return is not found
	ErrReturnNotFound = errors.New("return not found")

	// ErrReturnItemNotFound is returned when return item is not found
	ErrReturnItemNotFound = errors.New("return item not found")

	// ErrSaleNotFound is returned when sale is not found
	ErrSaleNotFound = errors.New("sale not found")

	// ErrSaleItemNotFound is returned when sale item is not found
	ErrSaleItemNotFound = errors.New("sale item not found")

	// ErrCustomerNotFound is returned when customer is not found
	ErrCustomerNotFound = errors.New("customer not found")

	// ErrReturnExists is returned when return already exists
	ErrReturnExists = errors.New("return already exists")

	// ErrInvalidReturnStatus is returned when return status is invalid
	ErrInvalidReturnStatus = errors.New("invalid return status")

	// ErrInvalidRefundMethod is returned when refund method is invalid
	ErrInvalidRefundMethod = errors.New("invalid refund method")

	// ErrInvalidCondition is returned when condition is invalid
	ErrInvalidCondition = errors.New("invalid condition")

	// ErrInvalidQuantity is returned when quantity is invalid
	ErrInvalidQuantity = errors.New("invalid quantity")

	// ErrReturnAlreadyApproved is returned when return is already approved
	ErrReturnAlreadyApproved = errors.New("return already approved")

	// ErrReturnAlreadyRejected is returned when return is already rejected
	ErrReturnAlreadyRejected = errors.New("return already rejected")

	// ErrReturnAlreadyCompleted is returned when return is already completed
	ErrReturnAlreadyCompleted = errors.New("return already completed")

	// ErrNoItems is returned when return has no items
	ErrNoItems = errors.New("return must have at least one item")

	// ErrInsufficientStock is returned when trying to return more than sold
	ErrInsufficientStock = errors.New("insufficient stock for return")

	// ErrRefundAlreadyProcessed is returned when refund is already processed
	ErrRefundAlreadyProcessed = errors.New("refund already processed")
)
