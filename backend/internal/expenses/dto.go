package expenses

import (
	"math"
	"strings"
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
		"id":               e.ID,
		"title":            e.Title,
		"description":      e.Description,
		"amount":           e.Amount,
		"currency":         e.Currency,
		"expense_date":     e.ExpenseDate,
		"category_id":      e.CategoryID,
		"category_name":    categoryName,
		"payment_method":   e.PaymentMethod,
		"status":           e.Status,
		"is_recurring":     e.IsRecurring,
		"recurring_period": e.RecurringPeriod,
		"created_at":       e.CreatedAt,
	}
}

// CreateExpense creates an Expense from request
func CreateExpense(userID uuid.UUID, req *ExpenseRequest) *Expense {
	return &Expense{
		ID:              uuid.New(),
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
func CreateExpenseCategory(req *ExpenseCategoryRequest) *ExpenseCategory {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &ExpenseCategory{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		Budget:      req.Budget,
		IsActive:    isActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ValidateExpenseRequest validates expense request
func ValidateExpenseRequest(req *ExpenseRequest) error {
	if req == nil {
		return ErrInvalidExpenseStatus
	}
	if req.CategoryID == uuid.Nil {
		return ErrExpenseCategoryNotFound
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return ErrInvalidExpenseTitle
	}
	if err := ValidateExpenseAmount(req.Amount); err != nil {
		return err
	}
	if err := ValidateCurrency(req.Currency); err != nil {
		return err
	}
	if err := ValidatePaymentMethod(req.PaymentMethod); err != nil {
		return err
	}
	if req.IsRecurring {
		if err := ValidateRecurringPeriod(req.RecurringPeriod); err != nil {
			return err
		}
	}
	return nil
}

// ValidateExpenseCategoryRequest validates expense category request
func ValidateExpenseCategoryRequest(req *ExpenseCategoryRequest) error {
	if req == nil {
		return ErrInvalidExpenseStatus
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ErrInvalidExpenseCategoryName
	}
	if req.Budget < 0 || math.IsNaN(req.Budget) || math.IsInf(req.Budget, 0) {
		return ErrInvalidAmount
	}
	return nil
}
