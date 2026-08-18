package inventory

import "errors"

var (
	// ErrItemNotFound is returned when an inventory item is not found
	ErrItemNotFound = errors.New("inventory item not found")

	// ErrLocationNotFound is returned when a location is not found
	ErrLocationNotFound = errors.New("location not found")

	// ErrInvalidStatus is returned when trying to perform an operation on an item with invalid status
	ErrInvalidStatus = errors.New("invalid item status for this operation")

	// ErrInsufficientStock is returned when there's not enough stock
	ErrInsufficientStock = errors.New("insufficient stock")

	// ErrItemAlreadyReserved is returned when trying to reserve an already reserved item
	ErrItemAlreadyReserved = errors.New("item is already reserved")

	// ErrReservationExpired is returned when trying to use an expired reservation
	ErrReservationExpired = errors.New("reservation has expired")

	// ErrInvalidCondition is returned when condition is invalid
	ErrInvalidCondition = errors.New("invalid condition")

	// ErrInvalidGrade is returned when grade is invalid
	ErrInvalidGrade = errors.New("invalid grade")

	// ErrDuplicateBarcode is returned when barcode already exists
	ErrDuplicateBarcode = errors.New("barcode already exists")

	// ErrDuplicateSerialNumber is returned when serial number already exists
	ErrDuplicateSerialNumber = errors.New("serial number already exists")

	// ErrCannotDeleteSoldItem is returned when trying to delete a sold item
	ErrCannotDeleteSoldItem = errors.New("cannot delete sold item")

	// ErrTransferFailed is returned when transfer fails
	ErrTransferFailed = errors.New("inventory transfer failed")
)
