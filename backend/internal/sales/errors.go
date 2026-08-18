package sales

import "errors"

var (
	ErrSaleNotFound       = errors.New("sale not found")
	ErrInvalidSaleStatus  = errors.New("invalid sale status")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrInvalidPayment     = errors.New("invalid payment amount")
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrInvalidCustomer    = errors.New("invalid customer")
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidQuantity    = errors.New("invalid quantity")
	ErrDuplicateInvoice   = errors.New("invoice number already exists")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
)