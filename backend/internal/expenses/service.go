package expenses

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles expense business logic
type Service struct {
	repo *Repository
}

// NewService creates a new expense service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateExpense creates a new expense
func (s *Service) CreateExpense(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req *ExpenseRequest) (*ExpenseResponse, error) {
	// Validate request
	if err := ValidateExpenseRequest(req); err != nil {
		return nil, err
	}

	// Check if category exists
	category, err := s.repo.GetExpenseCategoryByID(ctx, req.CategoryID, organizationID)
	if err != nil {
		return nil, ErrExpenseCategoryNotFound
	}

	// Check budget if category has budget
	if category.Budget > 0 {
		// This would require getting current expenses for this category
		// For simplicity, we'll skip budget check for now
	}

	// Create expense
	expense := CreateExpense(organizationID, userID, req)

	if err := s.repo.CreateExpense(ctx, expense); err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	return expense.ToExpenseResponse(category), nil
}

// GetExpense retrieves an expense by ID
func (s *Service) GetExpense(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*ExpenseResponse, error) {
	expense, err := s.repo.GetExpenseByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	category, err := s.repo.GetExpenseCategoryByID(ctx, expense.CategoryID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense category: %w", err)
	}

	return expense.ToExpenseResponse(category), nil
}

// ListExpenses retrieves expenses with pagination and filters
func (s *Service) ListExpenses(ctx context.Context, organizationID uuid.UUID, req ExpenseListRequest) ([]map[string]interface{}, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	expenses, total, err := s.repo.ListExpenses(ctx, organizationID, req)
	if err != nil {
		return nil, 0, err
	}

	// Convert to list items with category names
	var result []map[string]interface{}
	for _, expense := range expenses {
		category, err := s.repo.GetExpenseCategoryByID(ctx, expense.CategoryID, organizationID)
		if err != nil {
			continue
		}

		result = append(result, expense.ToExpenseListItem(category.Name))
	}

	return result, total, nil
}

// UpdateExpense updates an expense
func (s *Service) UpdateExpense(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *ExpenseUpdateRequest) (*ExpenseResponse, error) {
	expense, err := s.repo.GetExpenseByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Check if expense can be updated (not approved/rejected)
	if expense.Status == "approved" || expense.Status == "rejected" {
		return nil, ErrExpenseAlreadyApproved
	}

	// Update fields
	if req.CategoryID != uuid.Nil {
		// Verify category exists
		_, err := s.repo.GetExpenseCategoryByID(ctx, req.CategoryID, organizationID)
		if err != nil {
			return nil, ErrExpenseCategoryNotFound
		}
		expense.CategoryID = req.CategoryID
	}
	if req.Title != "" {
		expense.Title = req.Title
	}
	if req.Description != "" {
		expense.Description = req.Description
	}
	if req.Amount > 0 {
		expense.Amount = req.Amount
	}
	if req.Currency != "" {
		expense.Currency = req.Currency
	}
	if !req.ExpenseDate.IsZero() {
		expense.ExpenseDate = req.ExpenseDate
	}
	if req.PaymentMethod != "" {
		expense.PaymentMethod = req.PaymentMethod
	}
	if req.Reference != "" {
		expense.Reference = req.Reference
	}
	if req.ReceiptURL != "" {
		expense.ReceiptURL = req.ReceiptURL
	}
	expense.IsRecurring = req.IsRecurring
	if req.RecurringPeriod != "" {
		expense.RecurringPeriod = req.RecurringPeriod
	}
	if req.Status != "" {
		expense.Status = req.Status
	}

	expense.UpdatedAt = time.Now()

	if err := s.repo.UpdateExpense(ctx, expense); err != nil {
		return nil, err
	}

	return s.GetExpense(ctx, id, organizationID)
}

// DeleteExpense deletes an expense
func (s *Service) DeleteExpense(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	expense, err := s.repo.GetExpenseByID(ctx, id, organizationID)
	if err != nil {
		return err
	}

	// Check if expense can be deleted (not approved)
	if expense.Status == "approved" {
		return ErrExpenseAlreadyApproved
	}

	return s.repo.DeleteExpense(ctx, id, organizationID)
}

// ApproveExpense approves an expense
func (s *Service) ApproveExpense(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, approverID uuid.UUID) (*ExpenseResponse, error) {
	expense, err := s.repo.GetExpenseByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if expense.Status == "approved" {
		return nil, ErrExpenseAlreadyApproved
	}

	if expense.Status == "rejected" {
		return nil, ErrExpenseAlreadyRejected
	}

	expense.Status = "approved"
	expense.ApprovedBy = &approverID
	expense.UpdatedAt = time.Now()

	if err := s.repo.UpdateExpense(ctx, expense); err != nil {
		return nil, err
	}

	return s.GetExpense(ctx, id, organizationID)
}

// RejectExpense rejects an expense
func (s *Service) RejectExpense(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*ExpenseResponse, error) {
	expense, err := s.repo.GetExpenseByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	if expense.Status == "approved" {
		return nil, ErrExpenseAlreadyApproved
	}

	if expense.Status == "rejected" {
		return nil, ErrExpenseAlreadyRejected
	}

	expense.Status = "rejected"
	expense.UpdatedAt = time.Now()

	if err := s.repo.UpdateExpense(ctx, expense); err != nil {
		return nil, err
	}

	return s.GetExpense(ctx, id, organizationID)
}

// CreateExpenseCategory creates a new expense category
func (s *Service) CreateExpenseCategory(ctx context.Context, organizationID uuid.UUID, req *ExpenseCategoryRequest) (*ExpenseCategory, error) {
	// Validate request
	if err := ValidateExpenseCategoryRequest(req); err != nil {
		return nil, err
	}

	// Check if category name already exists
	existing, err := s.repo.GetExpenseCategoryByName(ctx, req.Name, organizationID)
	if err == nil && existing != nil {
		return nil, ErrExpenseCategoryExists
	}

	// Create category
	category := CreateExpenseCategory(organizationID, req)

	if err := s.repo.CreateExpenseCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create expense category: %w", err)
	}

	return category, nil
}

// GetExpenseCategory retrieves an expense category by ID
func (s *Service) GetExpenseCategory(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*ExpenseCategory, error) {
	return s.repo.GetExpenseCategoryByID(ctx, id, organizationID)
}

// ListExpenseCategories retrieves expense categories with pagination and filters
func (s *Service) ListExpenseCategories(ctx context.Context, organizationID uuid.UUID, req ExpenseCategoryListRequest) ([]ExpenseCategory, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	return s.repo.ListExpenseCategories(ctx, organizationID, req)
}

// UpdateExpenseCategory updates an expense category
func (s *Service) UpdateExpenseCategory(ctx context.Context, id uuid.UUID, organizationID uuid.UUID, req *ExpenseCategoryUpdateRequest) (*ExpenseCategory, error) {
	category, err := s.repo.GetExpenseCategoryByID(ctx, id, organizationID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		// Check if name already exists for another category
		existing, err := s.repo.GetExpenseCategoryByName(ctx, req.Name, organizationID)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrExpenseCategoryExists
		}
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.Color != "" {
		category.Color = req.Color
	}
	if req.Icon != "" {
		category.Icon = req.Icon
	}
	if req.Budget >= 0 {
		category.Budget = req.Budget
	}
	category.IsActive = req.IsActive
	category.UpdatedAt = time.Now()

	if err := s.repo.UpdateExpenseCategory(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteExpenseCategory deletes an expense category
func (s *Service) DeleteExpenseCategory(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	// Check if category has expenses
	// This would require a count query
	// For simplicity, we'll allow deletion for now

	return s.repo.DeleteExpenseCategory(ctx, id, organizationID)
}

// GetExpenseSummary retrieves expense summary statistics
func (s *Service) GetExpenseSummary(ctx context.Context, organizationID uuid.UUID) (*ExpenseSummary, error) {
	return s.repo.GetExpenseSummary(ctx, organizationID)
}
