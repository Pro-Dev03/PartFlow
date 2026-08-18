package expenses

import (
	"time"

	"github.com/google/uuid"
)

// Expense represents an expense in the system
type Expense struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	CategoryID     uuid.UUID  `json:"category_id" db:"category_id"`
	Title          string     `json:"title" db:"title"`
	Description    string     `json:"description" db:"description"`
	Amount         float64    `json:"amount" db:"amount"`
	Currency       string     `json:"currency" db:"currency"`
	ExpenseDate    time.Time  `json:"expense_date" db:"expense_date"`
	PaymentMethod  string     `json:"payment_method" db:"payment_method"` // cash, card, bank_transfer, check
	Reference      string     `json:"reference" db:"reference"`
	ReceiptURL     string     `json:"receipt_url" db:"receipt_url"`
	IsRecurring    bool       `json:"is_recurring" db:"is_recurring"`
	RecurringPeriod string    `json:"recurring_period" db:"recurring_period"` // daily, weekly, monthly, yearly
	ApprovedBy     *uuid.UUID `json:"approved_by" db:"approved_by"`
	Status         string     `json:"status" db:"status"` // pending, approved, rejected
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// ExpenseCategory represents an expense category
type ExpenseCategory struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Description    string     `json:"description" db:"description"`
	Color          string     `json:"color" db:"color"`
	Icon           string     `json:"icon" db:"icon"`
	Budget         float64    `json:"budget" db:"budget"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// ExpenseRequest represents expense creation request
type ExpenseRequest struct {
	CategoryID      uuid.UUID  `json:"category_id" binding:"required"`
	Title           string     `json:"title" binding:"required"`
	Description     string     `json:"description"`
	Amount          float64    `json:"amount" binding:"required,min=0"`
	Currency        string     `json:"currency" binding:"required"`
	ExpenseDate     time.Time  `json:"expense_date" binding:"required"`
	PaymentMethod   string     `json:"payment_method" binding:"required,oneof=cash card bank_transfer check"`
	Reference       string     `json:"reference"`
	ReceiptURL      string     `json:"receipt_url"`
	IsRecurring     bool       `json:"is_recurring"`
	RecurringPeriod string     `json:"recurring_period" binding:"omitempty,oneof=daily weekly monthly yearly"`
}

// ExpenseUpdateRequest represents expense update request
type ExpenseUpdateRequest struct {
	CategoryID      uuid.UUID  `json:"category_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Amount          float64    `json:"amount" binding:"omitempty,min=0"`
	Currency        string     `json:"currency"`
	ExpenseDate     time.Time  `json:"expense_date"`
	PaymentMethod   string     `json:"payment_method" binding:"omitempty,oneof=cash card bank_transfer check"`
	Reference       string     `json:"reference"`
	ReceiptURL      string     `json:"receipt_url"`
	IsRecurring     bool       `json:"is_recurring"`
	RecurringPeriod string     `json:"recurring_period" binding:"omitempty,oneof=daily weekly monthly yearly"`
	Status          string     `json:"status" binding:"omitempty,oneof=pending approved rejected"`
}

// ExpenseCategoryRequest represents expense category creation request
type ExpenseCategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Icon        string  `json:"icon"`
	Budget      float64 `json:"budget" binding:"omitempty,min=0"`
	IsActive    bool    `json:"is_active"`
}

// ExpenseCategoryUpdateRequest represents expense category update request
type ExpenseCategoryUpdateRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Icon        string  `json:"icon"`
	Budget      float64 `json:"budget" binding:"omitempty,min=0"`
	IsActive    bool    `json:"is_active"`
}

// ExpenseResponse represents expense response with related data
type ExpenseResponse struct {
	Expense  Expense           `json:"expense"`
	Category *ExpenseCategory  `json:"category,omitempty"`
}

// ExpenseListRequest represents expense list query parameters
type ExpenseListRequest struct {
	Page            int        `form:"page" binding:"min=1"`
	PerPage         int        `form:"per_page" binding:"min=1,max=100"`
	CategoryID      *uuid.UUID `form:"category_id"`
	Status          string     `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	PaymentMethod   string     `form:"payment_method" binding:"omitempty,oneof=cash card bank_transfer check"`
	StartDate       *time.Time `form:"start_date"`
	EndDate         *time.Time `form:"end_date"`
	IsRecurring     *bool      `form:"is_recurring"`
	Search          string     `form:"search"`
	SortBy          string     `form:"sort_by"`
	SortOrder       string     `form:"sort_order"`
}

// ExpenseCategoryListRequest represents expense category list query parameters
type ExpenseCategoryListRequest struct {
	Page       int    `form:"page" binding:"min=1"`
	PerPage    int    `form:"per_page" binding:"min=1,max=100"`
	IsActive   *bool  `form:"is_active"`
	Search     string `form:"search"`
	SortBy     string `form:"sort_by"`
	SortOrder  string `form:"sort_order"`
}

// ExpenseSummary represents expense summary
type ExpenseSummary struct {
	TotalExpenses    float64            `json:"total_expenses"`
	PendingExpenses  float64            `json:"pending_expenses"`
	ApprovedExpenses float64            `json:"approved_expenses"`
	ByCategory       map[string]float64 `json:"by_category"`
	ByPaymentMethod  map[string]float64 `json:"by_payment_method"`
	ThisMonth        float64            `json:"this_month"`
	LastMonth        float64            `json:"last_month"`
}
