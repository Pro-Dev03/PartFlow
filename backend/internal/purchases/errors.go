package purchases

import "errors"

var (
	// ErrPurchaseNotFound is returned when purchase is not found
	ErrPurchaseNotFound = errors.New("purchase not found")

	// ErrPurchaseItemNotFound is returned when purchase item is not found
	ErrPurchaseItemNotFound = errors.New("purchase item not found")

	// ErrSupplierNotFound is returned when supplier is not found
	ErrSupplierNotFound = errors.New("supplier not found")

	// ErrProductNotFound is returned when product is not found
	ErrProductNotFound = errors.New("product not found")

	// ErrPurchaseExists is returned when purchase already exists
	ErrPurchaseExists = errors.New("purchase already exists")

	// ErrInvalidPurchaseStatus is returned when purchase status is invalid
	ErrInvalidPurchaseStatus = errors.New("invalid purchase status")

	// ErrPurchaseAlreadyReceived is returned when purchase is already received
	ErrPurchaseAlreadyReceived = errors.New("purchase already received")

	// ErrPurchaseCancelled is returned when purchase is cancelled
	ErrPurchaseCancelled = errors.New("purchase is cancelled")

	// ErrNoItems is returned when purchase has no items
	ErrNoItems = errors.New("purchase must have at least one item")

	// ErrInvalidQuantity is returned when quantity is invalid
	ErrInvalidQuantity = errors.New("invalid quantity")

	// ErrInvalidCost is returned when cost is invalid
	ErrInvalidCost = errors.New("invalid cost")

	// ErrInvalidCondition is returned when condition is invalid
	ErrInvalidCondition = errors.New("invalid condition")
)
