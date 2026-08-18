package expenses

import (
	"time"

	"github.com/google/uuid"
)

// ToExpenseResponse converts Expense to ExpenseResponse
func (e *Expense) ToExpenseResponse(category *ExpenseCategory) *ExpenseResponse {
	return &ExpenseResponse{
		Expense:  *e,
		Category: category,
	}
}

// ToExpenseListItem converts Expense to list item format
func (e *Expense) ToExpenseListItem(categoryName string) map[string]interface{} {
	return map[string]interface{}{
		"id":              e.ID,
		"title":           e.Title,
		"amount":          e.Amount,
		"currency":        e.Currency,
		"expense_date":    e.ExpenseDate,
		"category_name":   categoryName,
		"payment_method":  e.PaymentMethod,
		"status":          e.Status,
		"is_recurring":    e.IsRecurring,
		"created_at":      e.CreatedAt,
	}
}

// CreateExpense creates an Expense from request
func CreateExpense(organizationID uuid.UUID, userID uuid.UUID, req *ExpenseRequest) *Expense {
	return &Expense{
		ID:              uuid.New(),
		OrganizationID:  organizationID,
		CategoryID:      req.CategoryID,
		Title:           req.Title,
		Description:     req.Description,
		Amount:          req.Amount,
		Currency:        req.Currency,
		ExpenseDate:     req.ExpenseDate,
		PaymentMethod:   req.PaymentMethod,
		Reference:       req.Reference,
		ReceiptURL:      req.ReceiptURL,
		IsRecurring:     req.IsRecurring,
		RecurringPeriod: req.RecurringPeriod,
		Status:          "pending",
		CreatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// CreateExpenseCategory creates an ExpenseCategory from request
func CreateExpenseCategory(organizationID uuid.UUID, req *ExpenseCategoryRequest) *ExpenseCategory {
	return &ExpenseCategory{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Name:           req.Name,
		Description:    req.Description,
		Color:          req.Color,
		Icon:           req.Icon,
		Budget:         req.Budget,
		IsActive:       req.IsActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// ValidateExpenseRequest validates expense request
func ValidateExpenseRequest(req *ExpenseRequest) error {
	if req.CategoryID == uuid.Nil {
		return ErrExpenseCategoryNotFound
	}
	if req.Title == "" {
		return ErrExpenseNotFound
	}
	if req.Amount <= 0 {
		return ErrInvalidAmount
	}
	if req.Currency == "" {
		return ErrInvalidCurrency
	}
	if req.PaymentMethod != "cash" && req.PaymentMethod != "card" && 
		req.PaymentMethod != "bank_transfer" && req.PaymentMethod != "check" {
		return ErrInvalidPaymentMethod
	}
	if req.IsRecurring && (req.RecurringPeriod != "daily" && 
		req.RecurringPeriod != "weekly" && req.RecurringPeriod != "monthly" && 
		req.RecurringPeriod != "yearly") {
		return ErrInvalidRecurringPeriod
	}
	return nil
}

// ValidateExpenseCategoryRequest validates expense category request
func ValidateExpenseCategoryRequest(req *ExpenseCategoryRequest) error {
	if req.Name == "" {
		return ErrExpenseCategoryNotFound
	}
	if req.Budget < 0 {
		return ErrInvalidAmount
	}
	return nil
}
