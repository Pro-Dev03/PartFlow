package products

import "errors"

var (
	// ErrCategoryNotFound is returned when category is not found
	ErrCategoryNotFound = errors.New("category not found")

	// ErrBrandNotFound is returned when brand is not found
	ErrBrandNotFound = errors.New("brand not found")

	// ErrProductNotFound is returned when product is not found
	ErrProductNotFound = errors.New("product not found")

	// ErrCategoryExists is returned when category already exists
	ErrCategoryExists = errors.New("category already exists")

	// ErrBrandExists is returned when brand already exists
	ErrBrandExists = errors.New("brand already exists")

	// ErrProductExists is returned when product already exists
	ErrProductExists = errors.New("product already exists")

	// ErrInvalidCategory is returned when category is invalid
	ErrInvalidCategory = errors.New("invalid category")

	// ErrInvalidBrand is returned when brand is invalid
	ErrInvalidBrand = errors.New("invalid brand")

	// ErrCategoryHasProducts is returned when category has products
	ErrCategoryHasProducts = errors.New("category has products")

	// ErrBrandHasProducts is returned when brand has products
	ErrBrandHasProducts = errors.New("brand has products")
)
