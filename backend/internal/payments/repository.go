package payments

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles payment data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new payment repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new payment
func (r *Repository) Create(ctx context.Context, payment *Payment) error {
	query := `
		INSERT INTO payments (id, organization_id, type, reference_id, amount, payment_date, method, reference, notes, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.OrganizationID, payment.Type, payment.ReferenceID,
		payment.Amount, payment.PaymentDate, payment.Method, payment.Reference,
		payment.Notes, payment.Status, payment.CreatedBy, payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

// GetByID retrieves a payment by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*Payment, error) {
	query := `
		SELECT id, organization_id, type, reference_id, amount, payment_date, method, reference, notes, status, created_by, created_at, updated_at
		FROM payments
		WHERE id = $1 AND organization_id = $2
	`
	var payment Payment
	err := r.db.GetContext(ctx, &payment, query, id, organizationID)
	if err != nil {
		return nil, ErrPaymentNotFound
	}
	return &payment, nil
}

// List retrieves payments with pagination and filters
func (r *Repository) List(ctx context.Context, organizationID uuid.UUID, page, perPage int, filters map[string]interface{}) ([]Payment, int, error) {
	offset := (page - 1) * perPage

	query := `
		SELECT id, organization_id, type, reference_id, amount, payment_date, method, reference, notes, status, created_by, created_at, updated_at
		FROM payments
		WHERE organization_id = $1
	`
	args := []interface{}{organizationID}
	argCount := 1

	// Add filters
	if paymentType, ok := filters["type"].(string); ok && paymentType != "" {
		argCount++
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, paymentType)
	}

	if referenceID, ok := filters["reference_id"].(uuid.UUID); ok {
		argCount++
		query += fmt.Sprintf(" AND reference_id = $%d", argCount)
		args = append(args, referenceID)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	if method, ok := filters["method"].(string); ok && method != "" {
		argCount++
		query += fmt.Sprintf(" AND method = $%d", argCount)
		args = append(args, method)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM payments WHERE organization_id = $1"
	countArgs := []interface{}{organizationID}
	countArgCount := 1

	if paymentType, ok := filters["type"].(string); ok && paymentType != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND type = $%d", countArgCount)
		countArgs = append(countArgs, paymentType)
	}

	if referenceID, ok := filters["reference_id"].(uuid.UUID); ok {
		countArgCount++
		countQuery += fmt.Sprintf(" AND reference_id = $%d", countArgCount)
		countArgs = append(countArgs, referenceID)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND status = $%d", countArgCount)
		countArgs = append(countArgs, status)
	}

	if method, ok := filters["method"].(string); ok && method != "" {
		countArgCount++
		countQuery += fmt.Sprintf(" AND method = $%d", countArgCount)
		countArgs = append(countArgs, method)
	}

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY payment_date DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, perPage, offset)

	var payments []Payment
	err = r.db.SelectContext(ctx, &payments, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list payments: %w", err)
	}

	return payments, total, nil
}

// Update updates a payment
func (r *Repository) Update(ctx context.Context, payment *Payment) error {
	query := `
		UPDATE payments
		SET status = $2, method = $3, reference = $4, notes = $5, payment_date = $6, updated_at = $7
		WHERE id = $1 AND organization_id = $8
	`
	result, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.Status, payment.Method, payment.Reference,
		payment.Notes, payment.PaymentDate, payment.UpdatedAt, payment.OrganizationID,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrPaymentNotFound
	}

	return nil
}

// Delete deletes a payment
func (r *Repository) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	query := `DELETE FROM payments WHERE id = $1 AND organization_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, organizationID)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrPaymentNotFound
	}

	return nil
}

// GetPaymentSummary retrieves payment summary statistics
func (r *Repository) GetPaymentSummary(ctx context.Context, organizationID uuid.UUID) (*PaymentSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(amount), 0) as total_payments,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END), 0) as completed_payments,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN amount ELSE 0 END), 0) as pending_payments,
			COALESCE(SUM(CASE WHEN status = 'cancelled' THEN amount ELSE 0 END), 0) as cancelled_payments,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN amount ELSE 0 END), 0) as failed_payments,
			COUNT(*) as total_count
		FROM payments
		WHERE organization_id = $1
	`
	var summary PaymentSummary
	err := r.db.GetContext(ctx, &summary, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment summary: %w", err)
	}
	return &summary, nil
}

// GetReferenceName retrieves the name for a reference (customer/supplier name)
func (r *Repository) GetReferenceName(ctx context.Context, paymentType string, referenceID uuid.UUID) (string, error) {
	var tableName string
	var nameField string

	switch paymentType {
	case "customer":
		tableName = "customers"
		nameField = "name"
	case "supplier":
		tableName = "suppliers"
		nameField = "name"
	default:
		return "", fmt.Errorf("invalid payment type")
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", nameField, tableName)
	var name string
	err := r.db.GetContext(ctx, &name, query, referenceID)
	if err != nil {
		return "", err
	}
	return name, nil
}