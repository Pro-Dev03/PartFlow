package payments

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Repository handles payment data operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new payment repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// parsePaymentMap converts a map result from SQLite into a Payment struct
func parsePaymentMap(record map[string]any) (Payment, error) {
	var payment Payment
	if raw, ok := record["id"]; ok && raw != nil {
		parsed, err := uuid.Parse(raw.(string))
		if err != nil {
			return payment, fmt.Errorf("parse payment id: %w", err)
		}
		payment.ID = parsed
	}
	if raw, ok := record["type"]; ok && raw != nil {
		payment.Type = fmt.Sprint(raw)
	}
	if raw, ok := record["reference_id"]; ok && raw != nil && raw != "" {
		parsed, err := uuid.Parse(raw.(string))
		if err != nil {
			return payment, fmt.Errorf("parse reference_id: %w", err)
		}
		payment.ReferenceID = parsed
	}
	if raw, ok := record["amount"]; ok && raw != nil {
		switch v := raw.(type) {
		case float64:
			payment.Amount = v
		case int64:
			payment.Amount = float64(v)
		}
	}
	if raw, ok := record["payment_date"]; ok && raw != nil && raw != "" {
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return payment, fmt.Errorf("parse payment_date: %w", err)
		}
		payment.PaymentDate = parsed
	}
	if raw, ok := record["method"]; ok && raw != nil {
		payment.Method = fmt.Sprint(raw)
	}
	if raw, ok := record["reference"]; ok && raw != nil && raw != "" {
		value := fmt.Sprint(raw)
		payment.Reference = &value
	}
	if raw, ok := record["notes"]; ok && raw != nil && raw != "" {
		value := fmt.Sprint(raw)
		payment.Notes = &value
	}
	if raw, ok := record["status"]; ok && raw != nil {
		payment.Status = fmt.Sprint(raw)
	}
	if raw, ok := record["created_by"]; ok && raw != nil && raw != "" {
		parsed, err := uuid.Parse(raw.(string))
		if err != nil {
			return payment, fmt.Errorf("parse created_by: %w", err)
		}
		payment.CreatedBy = parsed
	}
	if raw, ok := record["created_at"]; ok && raw != nil && raw != "" {
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return payment, fmt.Errorf("parse created_at: %w", err)
		}
		payment.CreatedAt = parsed
	}
	if raw, ok := record["updated_at"]; ok && raw != nil && raw != "" {
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return payment, fmt.Errorf("parse updated_at: %w", err)
		}
		payment.UpdatedAt = parsed
	}
	if raw, ok := record["is_reversed"]; ok && raw != nil {
		switch v := raw.(type) {
		case bool:
			payment.IsReversed = v
		case int:
			payment.IsReversed = v != 0
		case int64:
			payment.IsReversed = v != 0
		}
	}
	if raw, ok := record["reversed_at"]; ok && raw != nil && raw != "" {
		parsed, err := dbutil.ParseTimestamp(raw)
		if err != nil {
			return payment, fmt.Errorf("parse reversed_at: %w", err)
		}
		payment.ReversedAt = &parsed
	}
	if raw, ok := record["reversed_by"]; ok && raw != nil && raw != "" {
		parsed, err := uuid.Parse(raw.(string))
		if err != nil {
			return payment, fmt.Errorf("parse reversed_by: %w", err)
		}
		payment.ReversedBy = &parsed
	}
	if raw, ok := record["reversal_reason"]; ok && raw != nil && raw != "" {
		value := fmt.Sprint(raw)
		payment.ReversalReason = &value
	}
	if raw, ok := record["reversal_payment_id"]; ok && raw != nil && raw != "" {
		parsed, err := uuid.Parse(raw.(string))
		if err != nil {
			return payment, fmt.Errorf("parse reversal_payment_id: %w", err)
		}
		payment.ReversalPaymentID = &parsed
	}
	return payment, nil
}

// Create creates a new payment
func (r *Repository) Create(ctx context.Context, payment *Payment) error {
	query := `
		INSERT INTO payments (id, type, reference_id, amount, payment_date, method, reference, notes, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.Type, payment.ReferenceID, payment.Amount, payment.PaymentDate,
		payment.Method, payment.Reference, payment.Notes, payment.Status, payment.CreatedBy,
		payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

// GetByID retrieves a payment by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	if dbutil.IsSQLite(r.db) {
		row := r.db.QueryRowxContext(ctx, `SELECT id, CASE WHEN customer_id IS NOT NULL THEN 'customer' WHEN supplier_id IS NOT NULL THEN 'supplier' ELSE '' END AS type, COALESCE(customer_id, supplier_id, '00000000-0000-0000-0000-000000000000') AS reference_id, amount, created_at AS payment_date, COALESCE(payment_method, '') AS method, reference, notes, 'completed' AS status, '00000000-0000-0000-0000-000000000000' AS created_by, created_at, created_at, 0 AS is_reversed, NULL AS reversed_at, NULL AS reversed_by, NULL AS reversal_reason, NULL AS reversal_payment_id FROM payments WHERE id = $1`, id)
		record := map[string]any{}
		if err := row.MapScan(record); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrPaymentNotFound
			}
			return nil, fmt.Errorf("failed to get payment: %w", err)
		}
		payment, err := parsePaymentMap(record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse payment: %w", err)
		}
		return &payment, nil
	}
	query := `
		SELECT id, COALESCE(type, '') AS type, COALESCE(reference_id, '00000000-0000-0000-0000-000000000000') AS reference_id,
			amount, payment_date, COALESCE(method, payment_method) AS method,
			COALESCE(reference, reference_number) AS reference, notes, COALESCE(status, 'completed') AS status,
			COALESCE(created_by, '00000000-0000-0000-0000-000000000000') AS created_by, created_at, updated_at,
			COALESCE(is_reversed, 0) AS is_reversed, reversed_at, reversed_by, reversal_reason, reversal_payment_id
		FROM payments
		WHERE id = $1
	`
	row := r.db.QueryRowxContext(ctx, query, id)
	if row == nil {
		return nil, ErrPaymentNotFound
	}
	record := map[string]any{}
	if err := row.MapScan(record); err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	payment, err := parsePaymentMap(record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payment: %w", err)
	}
	return &payment, nil
}

// List retrieves payments with pagination and filters
func (r *Repository) List(ctx context.Context, page, perPage int, filters map[string]interface{}) ([]Payment, int, error) {
	if dbutil.IsSQLite(r.db) {
		if page <= 0 {
			page = 1
		}
		if perPage <= 0 || perPage > 100 {
			perPage = 20
		}
		query := `SELECT id, CASE WHEN customer_id IS NOT NULL THEN 'customer' WHEN supplier_id IS NOT NULL THEN 'supplier' ELSE '' END AS type, COALESCE(customer_id, supplier_id, '00000000-0000-0000-0000-000000000000') AS reference_id, amount, created_at AS payment_date, COALESCE(payment_method, '') AS method, reference, notes, 'completed' AS status, '00000000-0000-0000-0000-000000000000' AS created_by, created_at, created_at, 0 AS is_reversed, NULL AS reversed_at, NULL AS reversed_by, NULL AS reversal_reason, NULL AS reversal_payment_id FROM payments WHERE 1=1`
		countQuery := `SELECT COUNT(*) FROM payments WHERE 1=1`
		args := make([]interface{}, 0, 5)
		argCount := 0
		add := func(condition string, value interface{}) {
			argCount++
			query += fmt.Sprintf(" AND %s $%d", condition, argCount)
			countQuery += fmt.Sprintf(" AND %s $%d", condition, argCount)
			args = append(args, value)
		}
		if paymentType, ok := filters["type"].(string); ok && paymentType != "" {
			switch paymentType {
			case "customer":
				query += " AND customer_id IS NOT NULL"
				countQuery += " AND customer_id IS NOT NULL"
			case "supplier":
				query += " AND supplier_id IS NOT NULL"
				countQuery += " AND supplier_id IS NOT NULL"
			}
		}
		if referenceID, ok := filters["reference_id"].(uuid.UUID); ok {
			add("(customer_id =", referenceID)
			query += fmt.Sprintf(" OR supplier_id = $%d)", argCount)
			countQuery += fmt.Sprintf(" OR supplier_id = $%d)", argCount)
		}
		if status, ok := filters["status"].(string); ok && status != "" && status != "completed" {
			return []Payment{}, 0, nil
		}
		if method, ok := filters["method"].(string); ok && method != "" {
			add("payment_method =", method)
		}
		var total int
		if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
			return nil, 0, fmt.Errorf("failed to count payments: %w", err)
		}
		query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
		args = append(args, perPage, (page-1)*perPage)
		rows, err := r.db.QueryxContext(ctx, query, args...)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list payments: %w", err)
		}
		defer rows.Close()
		result := make([]Payment, 0)
		for rows.Next() {
			record := map[string]any{}
			if err := rows.MapScan(record); err != nil {
				return nil, 0, fmt.Errorf("failed to scan payment: %w", err)
			}
			payment, err := parsePaymentMap(record)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to parse payment: %w", err)
			}
			result = append(result, payment)
		}
		if err := rows.Err(); err != nil {
			return nil, 0, fmt.Errorf("failed to iterate payments: %w", err)
		}
		return result, total, nil
	}
	offset := (page - 1) * perPage

	query := `
		SELECT id, COALESCE(type, '') AS type, COALESCE(reference_id, '00000000-0000-0000-0000-000000000000') AS reference_id,
			amount, payment_date, COALESCE(method, payment_method) AS method,
			COALESCE(reference, reference_number) AS reference, notes, COALESCE(status, 'completed') AS status,
			COALESCE(created_by, '00000000-0000-0000-0000-000000000000') AS created_by, created_at, updated_at,
			COALESCE(is_reversed, 0) AS is_reversed, reversed_at, reversed_by, reversal_reason, reversal_payment_id
		FROM payments
		WHERE 1=1
	`
	countQuery := `
		SELECT COUNT(*) FROM payments
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	// Add filters
	if paymentType, ok := filters["type"].(string); ok && paymentType != "" {
		argCount++
		query += fmt.Sprintf(" AND type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, paymentType)
	}

	if referenceID, ok := filters["reference_id"].(uuid.UUID); ok {
		argCount++
		query += fmt.Sprintf(" AND reference_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND reference_id = $%d", argCount)
		args = append(args, referenceID)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	if method, ok := filters["method"].(string); ok && method != "" {
		argCount++
		query += fmt.Sprintf(" AND method = $%d", argCount)
		countQuery += fmt.Sprintf(" AND method = $%d", argCount)
		args = append(args, method)
	}

	// Get total count
	countArgs := args

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY payment_date DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, perPage, offset)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		record := map[string]any{}
		if err := rows.MapScan(record); err != nil {
			return nil, 0, fmt.Errorf("failed to scan payment: %w", err)
		}
		payment, err := parsePaymentMap(record)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse payment: %w", err)
		}
		payments = append(payments, payment)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate payments: %w", err)
	}

	return payments, total, nil
}

// Update updates a payment
func (r *Repository) Update(ctx context.Context, payment *Payment) error {
	query := `
		UPDATE payments
		SET status = $2, method = $3, reference = $4, notes = $5, payment_date = $6, updated_at = $7
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.Status, payment.Method, payment.Reference,
		payment.Notes, payment.PaymentDate, payment.UpdatedAt,
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
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM payments WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
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
func (r *Repository) GetPaymentSummary(ctx context.Context) (*PaymentSummary, error) {
	if dbutil.IsSQLite(r.db) {
		var summary PaymentSummary
		if err := r.db.GetContext(ctx, &summary, `SELECT COALESCE(SUM(amount), 0) AS total_payments, COALESCE(SUM(amount), 0) AS completed_payments, 0 AS pending_payments, 0 AS cancelled_payments, 0 AS failed_payments, COUNT(*) AS total_count FROM payments`); err != nil {
			return nil, fmt.Errorf("failed to get payment summary: %w", err)
		}
		return &summary, nil
	}
	query := `
		SELECT 
			COALESCE(SUM(amount), 0) as total_payments,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END), 0) as completed_payments,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN amount ELSE 0 END), 0) as pending_payments,
			COALESCE(SUM(CASE WHEN status = 'cancelled' THEN amount ELSE 0 END), 0) as cancelled_payments,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN amount ELSE 0 END), 0) as failed_payments,
			COUNT(*) as total_count
		FROM payments
	`
	var summary PaymentSummary
	err := r.db.GetContext(ctx, &summary, query)
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
