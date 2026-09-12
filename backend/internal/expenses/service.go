package expenses

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/partflow/smart-store/internal/dashboard"
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
func (s *Service) CreateExpense(ctx context.Context, userID uuid.UUID, req *ExpenseRequest) (*ExpenseResponse, error) {
	// Validate request
	if err := ValidateExpenseRequest(req); err != nil {
		return nil, err
	}

	// Check if category exists
	category, err := s.repo.GetExpenseCategoryByID(ctx, req.CategoryID)
	if err != nil {
		return nil, ErrExpenseCategoryNotFound
	}

	// Check budget if category has budget
	if category.Budget > 0 {
		// This would require getting current expenses for this category
		// For simplicity, we'll skip budget check for now
	}

	// Create expense
	expense := CreateExpense(userID, req)

	if err := s.repo.CreateExpense(ctx, expense); err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}
	dashboard.InvalidateDashboardCacheWithReason("expense_created")

	return expense.ToExpenseResponse(category), nil
}

// GetExpense retrieves an expense by ID
func (s *Service) GetExpense(ctx context.Context, id uuid.UUID) (*ExpenseResponse, error) {
	expense, err := s.repo.GetExpenseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	category, err := s.repo.GetExpenseCategoryByID(ctx, expense.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense category: %w", err)
	}

	return expense.ToExpenseResponse(category), nil
}

// ListExpenses retrieves expenses with pagination and filters
func (s *Service) ListExpenses(ctx context.Context, req ExpenseListRequest) ([]map[string]interface{}, int, error) {
	if err := s.EnsureRecurringExpenses(ctx, time.Now().UTC()); err != nil {
		return nil, 0, fmt.Errorf("failed to generate recurring expenses: %w", err)
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	expenses, total, err := s.repo.ListExpenses(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	// Convert to list items with category names
	var result []map[string]interface{}
	for _, expense := range expenses {
		var categoryName string
		if expense.CategoryID != uuid.Nil {
			category, err := s.repo.GetExpenseCategoryByID(ctx, expense.CategoryID)
			if err != nil {
				// If category not found, use empty string
				categoryName = ""
			} else {
				categoryName = category.Name
			}
		} else {
			categoryName = ""
		}

		result = append(result, expense.ToExpenseListItem(categoryName))
	}

	return result, total, nil
}

// EnsureRecurringExpenses materializes every recurring template through today.
// Generated rows are ordinary expenses and never act as templates themselves.
func (s *Service) EnsureRecurringExpenses(ctx context.Context, now time.Time) error {
	templates, err := s.repo.ListRecurringTemplates(ctx)
	if err != nil {
		return err
	}
	now = dateAtNoon(now.UTC())
	for _, template := range templates {
		period := strings.ToLower(strings.TrimSpace(template.RecurringPeriod))
		if period == "" {
			continue
		}
		nextDate := dateAtNoon(template.ExpenseDate.UTC())
		for {
			nextDate = nextRecurringDate(nextDate, period)
			if nextDate.After(now) {
				break
			}

			reference := fmt.Sprintf("recurring:%s:%s", template.ID.String(), nextDate.Format("2006-01-02"))
			exists, err := s.repo.RecurringReferenceExists(ctx, reference)
			if err != nil {
				return err
			}
			if !exists {
				generated := template
				generated.ID = uuid.New()
				generated.ExpenseDate = nextDate
				generated.Reference = reference
				generated.IsRecurring = false
				generated.RecurringPeriod = ""
				generated.CreatedAt = now
				generated.UpdatedAt = now
				if err := s.repo.CreateExpense(ctx, &generated); err != nil {
					return fmt.Errorf("create recurring expense for %s: %w", template.ID, err)
				}
			}
		}
	}
	return nil
}

func dateAtNoon(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 12, 0, 0, 0, time.UTC)
}

func nextRecurringDate(value time.Time, period string) time.Time {
	switch period {
	case "daily":
		return value.AddDate(0, 0, 1)
	case "weekly":
		return value.AddDate(0, 0, 7)
	case "yearly":
		return value.AddDate(1, 0, 0)
	default:
		return value.AddDate(0, 1, 0)
	}
}

// UpdateExpense updates an expense
func (s *Service) UpdateExpense(ctx context.Context, id uuid.UUID, req *ExpenseUpdateRequest) (*ExpenseResponse, error) {
	expense, err := s.repo.GetExpenseByID(ctx, id)
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
		_, err := s.repo.GetExpenseCategoryByID(ctx, req.CategoryID)
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
	if req.Amount > 0 && math.Trunc(req.Amount) != req.Amount {
		return nil, ErrInvalidAmount
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
	dashboard.InvalidateDashboardCacheWithReason("expense_updated")

	return s.GetExpense(ctx, id)
}

// DeleteExpense deletes an expense
func (s *Service) DeleteExpense(ctx context.Context, id uuid.UUID) error {
	expense, err := s.repo.GetExpenseByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if expense can be deleted (not approved)
	if expense.Status == "approved" {
		return ErrExpenseAlreadyApproved
	}

	if err := s.repo.DeleteExpense(ctx, id); err != nil {
		return err
	}
	dashboard.InvalidateDashboardCacheWithReason("expense_deleted")
	return nil
}

// ApproveExpense approves an expense
func (s *Service) ApproveExpense(ctx context.Context, id uuid.UUID, approverID uuid.UUID) (*ExpenseResponse, error) {
	expense, err := s.repo.GetExpenseByID(ctx, id)
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
	dashboard.InvalidateDashboardCacheWithReason("expense_approved")

	return s.GetExpense(ctx, id)
}

// RejectExpense rejects an expense
func (s *Service) RejectExpense(ctx context.Context, id uuid.UUID) (*ExpenseResponse, error) {
	expense, err := s.repo.GetExpenseByID(ctx, id)
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
	dashboard.InvalidateDashboardCacheWithReason("expense_rejected")

	return s.GetExpense(ctx, id)
}

// CreateExpenseCategory creates a new expense category
func (s *Service) CreateExpenseCategory(ctx context.Context, req *ExpenseCategoryRequest) (*ExpenseCategory, error) {
	// Validate request
	if err := ValidateExpenseCategoryRequest(req); err != nil {
		return nil, err
	}

	// Check if category name already exists
	existing, err := s.repo.GetExpenseCategoryByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, ErrExpenseCategoryExists
	}

	// Create category
	category := CreateExpenseCategory(req)

	if err := s.repo.CreateExpenseCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create expense category: %w", err)
	}

	return category, nil
}

// GetExpenseCategory retrieves an expense category by ID
func (s *Service) GetExpenseCategory(ctx context.Context, id uuid.UUID) (*ExpenseCategory, error) {
	return s.repo.GetExpenseCategoryByID(ctx, id)
}

// ListExpenseCategories retrieves expense categories with pagination and filters
func (s *Service) ListExpenseCategories(ctx context.Context, req ExpenseCategoryListRequest) ([]ExpenseCategory, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}

	return s.repo.ListExpenseCategories(ctx, req)
}

// UpdateExpenseCategory updates an expense category
func (s *Service) UpdateExpenseCategory(ctx context.Context, id uuid.UUID, req *ExpenseCategoryUpdateRequest) (*ExpenseCategory, error) {
	category, err := s.repo.GetExpenseCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		// Check if name already exists for another category
		existing, err := s.repo.GetExpenseCategoryByName(ctx, req.Name)
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
func (s *Service) DeleteExpenseCategory(ctx context.Context, id uuid.UUID) error {
	// Check if category has expenses
	// This would require a count query
	// For simplicity, we'll allow deletion for now

	return s.repo.DeleteExpenseCategory(ctx, id)
}

// GetExpenseSummary retrieves expense summary statistics
func (s *Service) GetExpenseSummary(ctx context.Context) (*ExpenseSummary, error) {
	return s.repo.GetExpenseSummary(ctx)
}
