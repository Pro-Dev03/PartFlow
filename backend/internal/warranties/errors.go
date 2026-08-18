package warranties

import "errors"

var (
	// ErrWarrantyNotFound is returned when warranty is not found
	ErrWarrantyNotFound = errors.New("warranty not found")

	// ErrWarrantyClaimNotFound is returned when warranty claim is not found
	ErrWarrantyClaimNotFound = errors.New("warranty claim not found")

	// ErrProductNotFound is returned when product is not found
	ErrProductNotFound = errors.New("product not found")

	// ErrCustomerNotFound is returned when customer is not found
	ErrCustomerNotFound = errors.New("customer not found")

	// ErrWarrantyExists is returned when warranty already exists
	ErrWarrantyExists = errors.New("warranty already exists")

	// ErrWarrantyExpired is returned when warranty is expired
	ErrWarrantyExpired = errors.New("warranty is expired")

	// ErrWarrantyClaimed is returned when warranty is already claimed
	ErrWarrantyClaimed = errors.New("warranty is already claimed")

	// ErrInvalidWarrantyType is returned when warranty type is invalid
	ErrInvalidWarrantyType = errors.New("invalid warranty type")

	// ErrInvalidWarrantyStatus is returned when warranty status is invalid
	ErrInvalidWarrantyStatus = errors.New("invalid warranty status")

	// ErrInvalidClaimStatus is returned when claim status is invalid
	ErrInvalidClaimStatus = errors.New("invalid claim status")

	// ErrInvalidWarrantyPeriod is returned when warranty period is invalid
	ErrInvalidWarrantyPeriod = errors.New("invalid warranty period")

	// ErrClaimAlreadyApproved is returned when claim is already approved
	ErrClaimAlreadyApproved = errors.New("claim already approved")

	// ErrClaimAlreadyRejected is returned when claim is already rejected
	ErrClaimAlreadyRejected = errors.New("claim already rejected")

	// ErrClaimAlreadyCompleted is returned when claim is already completed
	ErrClaimAlreadyCompleted = errors.New("claim already completed")

	// ErrCannotClaimExpiredWarranty is returned when trying to claim expired warranty
	ErrCannotClaimExpiredWarranty = errors.New("cannot claim expired warranty")

	// ErrCannotClaimVoidedWarranty is returned when trying to claim voided warranty
	ErrCannotClaimVoidedWarranty = errors.New("cannot claim voided warranty")
)
