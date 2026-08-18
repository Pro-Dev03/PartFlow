package warranty

import "errors"

var (
	// ErrWarrantyClaimNotFound is returned when a warranty claim is not found
	ErrWarrantyClaimNotFound = errors.New("warranty claim not found")

	// ErrWarrantyNotFound is returned when a warranty is not found
	ErrWarrantyNotFound = errors.New("warranty not found")

	// ErrWarrantyExpired is returned when warranty has expired
	ErrWarrantyExpired = errors.New("warranty has expired")

	// ErrClaimAlreadyApproved is returned when claim is already approved
	ErrClaimAlreadyApproved = errors.New("claim already approved")

	// ErrClaimAlreadyRejected is returned when claim is already rejected
	ErrClaimAlreadyRejected = errors.New("claim already rejected")

	// ErrClaimAlreadyCompleted is returned when claim is already completed
	ErrClaimAlreadyCompleted = errors.New("claim already completed")

	// ErrInvalidClaimStatus is returned when invalid status transition is attempted
	ErrInvalidClaimStatus = errors.New("invalid claim status transition")

	// ErrInvalidWarrantyType is returned when invalid warranty type is provided
	ErrInvalidWarrantyType = errors.New("invalid warranty type")

	// ErrInvalidClaimType is returned when invalid claim type is provided
	ErrInvalidClaimType = errors.New("invalid claim type")

	// ErrProductNotFound is returned when product is not found
	ErrProductNotFound = errors.New("product not found")

	// ErrCustomerNotFound is returned when customer is not found
	ErrCustomerNotFound = errors.New("customer not found")

	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")
)