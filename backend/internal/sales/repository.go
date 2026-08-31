package sales

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

type Repository struct {
	db *sqlx.DB
}

type localSaleRow struct {
	ID             uuid.UUID  `db:"id"`
	SaleDate       string     `db:"sale_date"`
	CustomerID     *uuid.UUID `db:"customer_id"`
	InvoiceNumber  string     `db:"invoice_number"`
	Subtotal       float64    `db:"subtotal"`
	TaxAmount      float64    `db:"tax_amount"`
	DiscountAmount float64    `db:"discount_amount"`
	TotalAmount    float64    `db:"total_amount"`
	CostAmount     float64    `db:"cost_amount"`
	GrossProfit    float64    `db:"gross_profit"`
	NetProfit      float64    `db:"net_profit"`
	PaidAmount     float64    `db:"paid_amount"`
	PaymentMethod  *string    `db:"payment_method"`
	PaymentStatus  string     `db:"payment_status"`
	Status         string     `db:"status"`
	Notes          *string    `db:"notes"`
	CreatedAt      string     `db:"created_at"`
	UpdatedAt      string     `db:"updated_at"`
}

func parseSaleTime(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported local timestamp %q", value)
}

func (r localSaleRow) sale() (Sale, error) {
	sale := Sale{ID: r.ID, InvoiceNumber: r.InvoiceNumber, CustomerID: r.CustomerID, SaleDate: time.Time{}, Subtotal: r.Subtotal, TaxAmount: r.TaxAmount, DiscountAmount: r.DiscountAmount, TotalAmount: r.TotalAmount, CostAmount: r.CostAmount, GrossProfit: r.GrossProfit, NetProfit: r.NetProfit, PaidAmount: r.PaidAmount, PaymentMethod: r.PaymentMethod, PaymentStatus: r.PaymentStatus, Status: r.Status, Notes: r.Notes}
	var err error
	sale.SaleDate, err = parseSaleTime(r.SaleDate)
	if err != nil {
		return Sale{}, err
	}
	sale.CreatedAt, err = parseSaleTime(r.CreatedAt)
	if err != nil {
		return Sale{}, err
	}
	sale.UpdatedAt, err = parseSaleTime(r.UpdatedAt)
	return sale, err
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateSale creates a new sale
func (r *Repository) CreateSale(ctx context.Context, sale *Sale) error {
	query := `
		INSERT INTO sales (id, sale_date, customer_id, invoice_number, 
			subtotal, tax_amount, discount_amount, total_amount, paid_amount, payment_method, 
			payment_status, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.ExecContext(ctx, query,
		sale.ID, sale.SaleDate, sale.CustomerID, sale.InvoiceNumber,
		sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount,
		sale.PaidAmount, sale.PaymentMethod, sale.PaymentStatus, sale.Status, sale.Notes,
		sale.CreatedAt, sale.UpdatedAt)
	return err
}

// GetSaleByID retrieves a sale by ID
func (r *Repository) GetSaleByID(ctx context.Context, id uuid.UUID) (*Sale, error) {
	query := `
		SELECT id, sale_date, customer_id, invoice_number, 
			subtotal, tax_amount, discount_amount, total_amount, cost_amount, gross_profit, net_profit,
			paid_amount, payment_method, payment_status, status, notes, created_at, updated_at
		FROM sales WHERE id = $1
	`
	if dbutil.IsSQLite(r.db) {
		var row localSaleRow
		if err := r.db.GetContext(ctx, &row, query, id); err != nil {
			return nil, err
		}
		sale, err := row.sale()
		if err != nil {
			return nil, err
		}
		return &sale, nil
	}
	var sale Sale
	err := r.db.GetContext(ctx, &sale, query, id)
	if err != nil {
		return nil, err
	}
	return &sale, nil
}

// GetSaleByInvoiceNumber retrieves a sale by invoice number
func (r *Repository) GetSaleByInvoiceNumber(ctx context.Context, invoiceNumber string) (*Sale, error) {
	query := `
		SELECT id, sale_date, customer_id, invoice_number, 
			subtotal, tax_amount, discount_amount, total_amount, cost_amount, gross_profit, net_profit,
			paid_amount, payment_method, payment_status, status, notes, created_at, updated_at
		FROM sales WHERE invoice_number = $1
	`
	if dbutil.IsSQLite(r.db) {
		var row localSaleRow
		if err := r.db.GetContext(ctx, &row, query, invoiceNumber); err != nil {
			return nil, err
		}
		sale, err := row.sale()
		if err != nil {
			return nil, err
		}
		return &sale, nil
	}
	var sale Sale
	err := r.db.GetContext(ctx, &sale, query, invoiceNumber)
	if err != nil {
		return nil, err
	}
	return &sale, nil
}

// ListSales retrieves sales with pagination and filters
func (r *Repository) ListSales(ctx context.Context, page, perPage int, filters map[string]interface{}) ([]Sale, int, error) {
	offset := (page - 1) * perPage

	baseQuery := `
		SELECT id, sale_date, customer_id, invoice_number, 
			subtotal, tax_amount, discount_amount, total_amount, cost_amount, gross_profit, net_profit,
			paid_amount, payment_method, payment_status, status, notes, created_at, updated_at
		FROM sales WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM sales WHERE 1=1`

	args := []interface{}{}
	argCount := 0

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

	baseQuery += fmt.Sprintf(" ORDER BY sale_date DESC, id DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, perPage, offset)

	var rows []localSaleRow
	err := r.db.SelectContext(ctx, &rows, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	sales := make([]Sale, 0, len(rows))
	for _, row := range rows {
		sale, parseErr := row.sale()
		if parseErr != nil {
			return nil, 0, parseErr
		}
		sales = append(sales, sale)
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
	query := fmt.Sprintf(`
		UPDATE sales SET
			customer_id = $2, payment_method = $3, payment_status = $4,
			status = $5, notes = $6, paid_amount = $7, updated_at = %s
		WHERE id = $1
	`, dbutil.NowSQL(r.db))
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
		INSERT INTO sale_items (id, sale_id, product_id, inventory_item_id, quantity, unit_price,
			discount_amount, tax_amount, total_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.SaleID, item.ProductID, item.InventoryItemID, item.Quantity, item.UnitPrice,
		item.DiscountAmount, item.TaxAmount, item.TotalAmount, item.CreatedAt)
	return err
}

// GetSaleItems retrieves items for a sale
func (r *Repository) GetSaleItems(ctx context.Context, saleID uuid.UUID) ([]SaleItem, error) {
	query := `
		SELECT si.id, si.sale_id, si.product_id, si.inventory_item_id, ii.serial_number,
			si.quantity, si.unit_price, si.unit_cost,
			si.discount_amount, si.tax_amount, si.total_amount, si.created_at
		FROM sale_items si
		LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
		WHERE si.sale_id = $1
	`
	if dbutil.IsSQLite(r.db) {
		var rows []struct {
			ID              string  `db:"id"`
			SaleID          string  `db:"sale_id"`
			ProductID       string  `db:"product_id"`
			InventoryItemID string  `db:"inventory_item_id"`
			SerialNumber    string  `db:"serial_number"`
			Quantity        int     `db:"quantity"`
			UnitPrice       float64 `db:"unit_price"`
			UnitCost        float64 `db:"unit_cost"`
			DiscountAmount  float64 `db:"discount_amount"`
			TaxAmount       float64 `db:"tax_amount"`
			TotalAmount     float64 `db:"total_amount"`
			CreatedAt       string  `db:"created_at"`
		}
		if err := r.db.SelectContext(ctx, &rows, query, saleID); err != nil {
			return nil, err
		}
		items := make([]SaleItem, 0, len(rows))
		for _, row := range rows {
			created, err := parseSaleTime(row.CreatedAt)
			if err != nil {
				return nil, err
			}
			item := SaleItem{Quantity: row.Quantity, UnitPrice: row.UnitPrice, UnitCost: row.UnitCost, DiscountAmount: row.DiscountAmount, TaxAmount: row.TaxAmount, TotalAmount: row.TotalAmount, CreatedAt: created}
			if item.ID, err = uuid.Parse(row.ID); err != nil {
				return nil, err
			}
			if item.SaleID, err = uuid.Parse(row.SaleID); err != nil {
				return nil, err
			}
			if item.ProductID, err = uuid.Parse(row.ProductID); err != nil {
				return nil, err
			}
			if row.InventoryItemID != "" {
				v, parseErr := uuid.Parse(row.InventoryItemID)
				if parseErr == nil {
					item.InventoryItemID = &v
				}
			}
			if row.SerialNumber != "" {
				serial := row.SerialNumber
				item.SerialNumber = &serial
			}
			items = append(items, item)
		}
		return items, nil
	}
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
func (r *Repository) GetSalesSummary(ctx context.Context, startDate, endDate string) (*SalesSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_sales,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(paid_amount), 0) as total_paid,
			COALESCE(SUM(discount_amount), 0) as total_discount,
			COALESCE(SUM(tax_amount), 0) as total_tax
		FROM sales 
		WHERE sale_date >= $1 
		AND sale_date <= $2
		AND status = 'completed'
	`
	var summary SalesSummary
	err := r.db.GetContext(ctx, &summary, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// GetTopSellingProducts retrieves top selling products
func (r *Repository) GetTopSellingProducts(ctx context.Context, limit int) ([]TopSellingProduct, error) {
	query := `
		SELECT 
			p.id as product_id,
			p.name as product_name,
			COALESCE(SUM(si.quantity), 0) as total_quantity,
			COALESCE(SUM(si.total_amount), 0) as total_revenue
		FROM products p
		LEFT JOIN sale_items si ON p.id = si.product_id
		LEFT JOIN sales s ON si.sale_id = s.id AND s.status = 'completed'
		GROUP BY p.id, p.name
		ORDER BY total_quantity DESC
		LIMIT $1
	`
	var products []TopSellingProduct
	err := r.db.SelectContext(ctx, &products, query, limit)
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
		INSERT INTO financial_transactions (id, sale_id, type, amount, currency, reference, description,
			debit_account, credit_account, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		tx.ID, tx.SaleID, tx.Type, tx.Amount, tx.Currency, tx.Reference, tx.Description,
		tx.DebitAccount, tx.CreditAccount, tx.Status, tx.CreatedAt, tx.UpdatedAt)
	return err
}

// GetTransactionByID retrieves a transaction by ID
func (r *Repository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	query := `
		SELECT id, sale_id, type, amount, currency, reference, description,
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
func (r *Repository) ListTransactions(ctx context.Context, page, perPage int, filters map[string]interface{}) ([]Transaction, int, error) {
	offset := (page - 1) * perPage

	baseQuery := `
		SELECT id, sale_id, type, amount, currency, reference, description,
			debit_account, credit_account, status, created_at, updated_at
		FROM financial_transactions WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM financial_transactions WHERE 1=1`

	args := []interface{}{}
	argCount := 0

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
		INSERT INTO profit_entries (id, period, start_date, end_date,
			revenue, cost, gross_profit, net_profit, margin, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.Period, entry.StartDate, entry.EndDate,
		entry.Revenue, entry.Cost, entry.GrossProfit, entry.NetProfit, entry.Margin, entry.CreatedAt)
	return err
}

// GetProfitEntries retrieves profit entries for a period
func (r *Repository) GetProfitEntries(ctx context.Context, period string, startDate, endDate time.Time) ([]ProfitEntry, error) {
	query := `
		SELECT id, period, start_date, end_date,
			revenue, cost, gross_profit, net_profit, margin, created_at
		FROM profit_entries 
		WHERE period = $1
		AND start_date >= $2 
		AND end_date <= $3
		ORDER BY start_date DESC
	`
	var entries []ProfitEntry
	err := r.db.SelectContext(ctx, &entries, query, period, startDate, endDate)
	return entries, err
}

// GetAccountBalance retrieves the balance for a specific account
func (r *Repository) GetAccountBalance(ctx context.Context, account string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(CASE WHEN debit_account = $1 THEN amount ELSE -amount END), 0)
		FROM financial_transactions 
		WHERE (debit_account = $1 OR credit_account = $1)
		AND status = 'completed'
	`
	var balance float64
	err := r.db.GetContext(ctx, &balance, query, account)
	return balance, err
}
