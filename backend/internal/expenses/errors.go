package expenses

import "errors"

var (
	// ErrExpenseNotFound is returned when expense is not found
	ErrExpenseNotFound = errors.New("expense not found")

	// ErrExpenseCategoryNotFound is returned when expense category is not found
	ErrExpenseCategoryNotFound = errors.New("expense category not found")

	// ErrExpenseExists is returned when expense already exists
	ErrExpenseExists = errors.New("expense already exists")

	// ErrExpenseCategoryExists is returned when expense category already exists
	ErrExpenseCategoryExists = errors.New("expense category already exists")

	// ErrInvalidExpenseStatus is returned when expense status is invalid
	ErrInvalidExpenseStatus = errors.New("invalid expense status")

	// ErrInvalidPaymentMethod is returned when payment method is invalid
	ErrInvalidPaymentMethod = errors.New("invalid payment method")

	// ErrInvalidAmount is returned when amount is invalid
	ErrInvalidAmount = errors.New("invalid amount")

	// ErrInvalidCurrency is returned when currency is invalid
	ErrInvalidCurrency = errors.New("invalid currency")

	// ErrInvalidRecurringPeriod is returned when recurring period is invalid
	ErrInvalidRecurringPeriod = errors.New("invalid recurring period")

	// ErrExpenseAlreadyApproved is returned when expense is already approved
	ErrExpenseAlreadyApproved = errors.New("expense already approved")

	// ErrExpenseAlreadyRejected is returned when expense is already rejected
	ErrExpenseAlreadyRejected = errors.New("expense already rejected")

	// ErrExpenseCategoryHasExpenses is returned when category has expenses
	ErrExpenseCategoryHasExpenses = errors.New("expense category has expenses")

	// ErrCategoryBudgetExceeded is returned when category budget is exceeded
	ErrCategoryBudgetExceeded = errors.New("category budget exceeded")
)
