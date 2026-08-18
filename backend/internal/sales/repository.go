package sales

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateSale creates a new sale
func (r *Repository) CreateSale(ctx context.Context, sale *Sale) error {
	query := `
		INSERT INTO sales (id, organization_id, invoice_number, customer_id, user_id, sale_date, 
			subtotal, tax_amount, discount_amount, total_amount, paid_amount, payment_method, 
			payment_status, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	_, err := r.db.ExecContext(ctx, query,
		sale.ID, sale.OrganizationID, sale.InvoiceNumber, sale.CustomerID, sale.UserID,
		sale.SaleDate, sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount,
		sale.PaidAmount, sale.PaymentMethod, sale.PaymentStatus, sale.Status, sale.Notes,
		sale.CreatedAt, sale.UpdatedAt)
	return err
}

// GetSaleByID retrieves a sale by ID
func (r *Repository) GetSaleByID(ctx context.Context, id uuid.UUID) (*Sale, error) {
	query := `
		SELECT id, organization_id, invoice_number, customer_id, user_id, sale_date,
			subtotal, tax_amount, discount_amount, total_amount, paid_amount, payment_method,
			payment_status, status, notes, created_at, updated_at
		FROM sales WHERE id = $1
	`
	var sale Sale
	err := r.db.GetContext(ctx, &sale, query, id)
	if err != nil {
		return nil, err
	}
	return &sale, nil
}

// GetSaleByInvoiceNumber retrieves a sale by invoice number
func (r *Repository) GetSaleByInvoiceNumber(ctx context.Context, organizationID uuid.UUID, invoiceNumber string) (*Sale, error) {
	query := `
		SELECT id, organization_id, invoice_number, customer_id, user_id, sale_date,
			subtotal, tax_amount, discount_amount, total_amount, paid_amount, payment_method,
			payment_status, status, notes, created_at, updated_at
		FROM sales WHERE organization_id = $1 AND invoice_number = $2
	`
	var sale Sale
	err := r.db.GetContext(ctx, &sale, query, organizationID, invoiceNumber)
	if err != nil {
		return nil, err
	}
	return &sale, nil
}

// ListSales retrieves sales with pagination and filters
func (r *Repository) ListSales(ctx context.Context, organizationID uuid.UUID, page, perPage int, filters map[string]interface{}) ([]Sale, int, error) {
	offset := (page - 1) * perPage
	
	baseQuery := `
		SELECT id, organization_id, invoice_number, customer_id, user_id, sale_date,
			subtotal, tax_amount, discount_amount, total_amount, paid_amount, payment_method,
			payment_status, status, notes, created_at, updated_at
		FROM sales WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM sales WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
	// Add filters
	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}
	
	if customerID, ok := filters["customer_id"].(uuid.UUID); ok {
		argCount++
		baseQuery += fmt.Sprintf(" AND customer_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND customer_id = $%d", argCount)
		args = append(args, customerID)
	}
	
	if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND sale_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND sale_date >= $%d", argCount)
		args = append(args, startDate)
	}
	
	if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND sale_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND sale_date <= $%d", argCount)
		args = append(args, endDate)
	}
	
	baseQuery += fmt.Sprintf(" ORDER BY sale_date DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, perPage, offset)
	
	var sales []Sale
	err := r.db.SelectContext(ctx, &sales, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, args[:argCount]...)
	if err != nil {
		return nil, 0, err
	}
	
	return sales, total, nil
}

// UpdateSale updates a sale
func (r *Repository) UpdateSale(ctx context.Context, sale *Sale) error {
	query := `
		UPDATE sales SET
			customer_id = $2, payment_method = $3, payment_status = $4,
			status = $5, notes = $6, paid_amount = $7, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		sale.ID, sale.CustomerID, sale.PaymentMethod, sale.PaymentStatus,
		sale.Status, sale.Notes, sale.PaidAmount)
	return err
}

// DeleteSale deletes a sale
func (r *Repository) DeleteSale(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sales WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// CreateSaleItem creates a new sale item
func (r *Repository) CreateSaleItem(ctx context.Context, item *SaleItem) error {
	query := `
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_price, 
			discount_amount, tax_amount, total_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.SaleID, item.ProductID, item.Quantity, item.UnitPrice,
		item.DiscountAmount, item.TaxAmount, item.TotalAmount, item.CreatedAt)
	return err
}

// GetSaleItems retrieves items for a sale
func (r *Repository) GetSaleItems(ctx context.Context, saleID uuid.UUID) ([]SaleItem, error) {
	query := `
		SELECT id, sale_id, product_id, quantity, unit_price, 
			discount_amount, tax_amount, total_amount, created_at
		FROM sale_items WHERE sale_id = $1
	`
	var items []SaleItem
	err := r.db.SelectContext(ctx, &items, query, saleID)
	return items, err
}

// DeleteSaleItems deletes all items for a sale
func (r *Repository) DeleteSaleItems(ctx context.Context, saleID uuid.UUID) error {
	query := `DELETE FROM sale_items WHERE sale_id = $1`
	_, err := r.db.ExecContext(ctx, query, saleID)
	return err
}

// GetProductCost retrieves the cost price of a product
func (r *Repository) GetProductCost(ctx context.Context, productID uuid.UUID) (float64, error) {
	query := `SELECT cost_price FROM products WHERE id = $1`
	var costPrice float64
	err := r.db.GetContext(ctx, &costPrice, query, productID)
	return costPrice, err
}

// GetProductStock retrieves current stock for a product
func (r *Repository) GetProductStock(ctx context.Context, productID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM inventory_items 
		WHERE product_id = $1 AND status = 'AVAILABLE'
	`
	var count int
	err := r.db.GetContext(ctx, &count, query, productID)
	return count, err
}

// GetSalesSummary retrieves sales summary for a period
func (r *Repository) GetSalesSummary(ctx context.Context, organizationID uuid.UUID, startDate, endDate string) (*SalesSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_sales,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(paid_amount), 0) as total_paid,
			COALESCE(SUM(discount_amount), 0) as total_discount,
			COALESCE(SUM(tax_amount), 0) as total_tax
		FROM sales 
		WHERE organization_id = $1 
		AND sale_date >= $2 
		AND sale_date <= $3
		AND status = 'completed'
	`
	var summary SalesSummary
	err := r.db.GetContext(ctx, &summary, query, organizationID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// GetTopSellingProducts retrieves top selling products
func (r *Repository) GetTopSellingProducts(ctx context.Context, organizationID uuid.UUID, limit int) ([]TopSellingProduct, error) {
	query := `
		SELECT 
			p.id as product_id,
			p.name as product_name,
			COALESCE(SUM(si.quantity), 0) as total_quantity,
			COALESCE(SUM(si.total_amount), 0) as total_revenue
		FROM products p
		LEFT JOIN sale_items si ON p.id = si.product_id
		LEFT JOIN sales s ON si.sale_id = s.id
		WHERE p.organization_id = $1
		AND s.status = 'completed'
		GROUP BY p.id, p.name
		ORDER BY total_quantity DESC
		LIMIT $2
	`
	var products []TopSellingProduct
	err := r.db.SelectContext(ctx, &products, query, organizationID, limit)
	return products, err
}

// SalesSummary represents sales summary statistics
type SalesSummary struct {
	TotalSales    int     `json:"total_sales" db:"total_sales"`
	TotalRevenue  float64 `json:"total_revenue" db:"total_revenue"`
	TotalPaid     float64 `json:"total_paid" db:"total_paid"`
	TotalDiscount float64 `json:"total_discount" db:"total_discount"`
	TotalTax      float64 `json:"total_tax" db:"total_tax"`
}

// TopSellingProduct represents a top selling product
type TopSellingProduct struct {
	ProductID     uuid.UUID `json:"product_id" db:"product_id"`
	ProductName   string    `json:"product_name" db:"product_name"`
	TotalQuantity int       `json:"total_quantity" db:"total_quantity"`
	TotalRevenue  float64   `json:"total_revenue" db:"total_revenue"`
}

// CreateTransaction creates a new financial transaction
func (r *Repository) CreateTransaction(ctx context.Context, tx *Transaction) error {
	query := `
		INSERT INTO financial_transactions (id, organization_id, sale_id, type, amount, currency, 
			reference, description, debit_account, credit_account, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		tx.ID, tx.OrganizationID, tx.SaleID, tx.Type, tx.Amount, tx.Currency,
		tx.Reference, tx.Description, tx.DebitAccount, tx.CreditAccount, tx.Status,
		tx.CreatedAt, tx.UpdatedAt)
	return err
}

// GetTransactionByID retrieves a transaction by ID
func (r *Repository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	query := `
		SELECT id, organization_id, sale_id, type, amount, currency, reference, description,
			debit_account, credit_account, status, created_at, updated_at
		FROM financial_transactions WHERE id = $1
	`
	var tx Transaction
	err := r.db.GetContext(ctx, &tx, query, id)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// ListTransactions retrieves transactions with pagination and filters
func (r *Repository) ListTransactions(ctx context.Context, organizationID uuid.UUID, page, perPage int, filters map[string]interface{}) ([]Transaction, int, error) {
	offset := (page - 1) * perPage
	
	baseQuery := `
		SELECT id, organization_id, sale_id, type, amount, currency, reference, description,
			debit_account, credit_account, status, created_at, updated_at
		FROM financial_transactions WHERE organization_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM financial_transactions WHERE organization_id = $1`
	
	args := []interface{}{organizationID}
	argCount := 1
	
	// Add filters
	if txType, ok := filters["type"].(string); ok && txType != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, txType)
	}
	
	if status, ok := filters["status"].(string); ok && status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}
	
	if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, startDate)
	}
	
	if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, endDate)
	}
	
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, perPage, offset)
	
	var transactions []Transaction
	err := r.db.SelectContext(ctx, &transactions, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, args[:argCount]...)
	if err != nil {
		return nil, 0, err
	}
	
	return transactions, total, nil
}

// CreateProfitEntry creates a new profit entry
func (r *Repository) CreateProfitEntry(ctx context.Context, entry *ProfitEntry) error {
	query := `
		INSERT INTO profit_entries (id, organization_id, sale_id, period, start_date, end_date,
			revenue, cost, gross_profit, net_profit, margin, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.OrganizationID, entry.SaleID, entry.Period, entry.StartDate, entry.EndDate,
		entry.Revenue, entry.Cost, entry.GrossProfit, entry.NetProfit, entry.Margin, entry.CreatedAt)
	return err
}

// GetProfitEntries retrieves profit entries for a period
func (r *Repository) GetProfitEntries(ctx context.Context, organizationID uuid.UUID, period string, startDate, endDate time.Time) ([]ProfitEntry, error) {
	query := `
		SELECT id, organization_id, sale_id, period, start_date, end_date,
			revenue, cost, gross_profit, net_profit, margin, created_at
		FROM profit_entries 
		WHERE organization_id = $1 
		AND period = $2
		AND start_date >= $3 
		AND end_date <= $4
		ORDER BY start_date DESC
	`
	var entries []ProfitEntry
	err := r.db.SelectContext(ctx, &entries, query, organizationID, period, startDate, endDate)
	return entries, err
}

// GetAccountBalance retrieves the balance for a specific account
func (r *Repository) GetAccountBalance(ctx context.Context, organizationID uuid.UUID, account string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(CASE WHEN debit_account = $2 THEN amount ELSE -amount END), 0)
		FROM financial_transactions 
		WHERE organization_id = $1 
		AND (debit_account = $2 OR credit_account = $2)
		AND status = 'completed'
	`
	var balance float64
	err := r.db.GetContext(ctx, &balance, query, organizationID, account)
	return balance, err
}