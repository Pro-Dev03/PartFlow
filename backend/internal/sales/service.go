package sales

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"smart-store/internal/ledgers"
)

type Service struct {
	repo *Repository
	db   *sqlx.DB
}

func NewService(repo *Repository, db *sqlx.DB) *Service {
	return &Service{repo: repo, db: db}
}

// CreateSale creates a new sale with complete business logic automation
// This is an atomic transaction that ensures data consistency
func (s *Service) CreateSale(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, req *CreateSaleRequest) (*Sale, error) {
	// Start database transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Generate invoice number
	invoiceNumber := s.generateInvoiceNumber(organizationID)
	
	// Calculate totals and validate stock
	subtotal := 0.0
	totalTax := 0.0
	totalCost := 0.0
	var items []SaleItem
	itemStockMap := make(map[uuid.UUID][]struct {
		ID     uuid.UUID `db:"id"`
		Amount float64   `db:"selling_price"`
	}) // Store available items for each product
	
	for _, itemReq := range req.Items {
		// Check stock availability with row lock (using inventory_items for individual tracking)
		var availableItems []struct {
			ID     uuid.UUID `db:"id"`
			Amount float64   `db:"selling_price"`
		}
		stockQuery := `
			SELECT id, selling_price 
			FROM inventory_items 
			WHERE product_id = $1 AND organization_id = $2 AND status = 'AVAILABLE' 
			ORDER BY created_at ASC 
			FOR UPDATE
		`
		err := tx.SelectContext(ctx, &availableItems, stockQuery, itemReq.ProductID, organizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to check stock: %w", err)
		}
		
		if len(availableItems) < itemReq.Quantity {
			return nil, ErrInsufficientStock
		}
		
		// Store available items for this product
		itemStockMap[itemReq.ProductID] = availableItems
		
		// Get product cost for profit calculation
		var productCost float64
		costQuery := `SELECT COALESCE(cost_price, 0) FROM products WHERE id = $1 AND organization_id = $2`
		err = tx.GetContext(ctx, &productCost, costQuery, itemReq.ProductID, organizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product cost: %w", err)
		}
		
		// Calculate item totals using actual item costs
		itemTotal := float64(itemReq.Quantity) * itemReq.UnitPrice
		itemTax := itemTotal * req.TaxRate / 100
		itemTotalWithTax := itemTotal + itemTax
		
		// Calculate actual cost from available items
		itemCost := 0.0
		for i := 0; i < itemReq.Quantity && i < len(availableItems); i++ {
			itemCost += availableItems[i].Amount // Use actual selling price as cost basis
		}
		
		subtotal += itemTotal
		totalTax += itemTax
		totalCost += itemCost
		
		item := SaleItem{
			ID:            uuid.New(),
			ProductID:     itemReq.ProductID,
			Quantity:      itemReq.Quantity,
			UnitPrice:     itemReq.UnitPrice,
			UnitCost:      productCost, // Store base product cost
			TaxAmount:     itemTax,
			TotalAmount:   itemTotalWithTax,
			CreatedAt:     time.Now(),
		}
		items = append(items, item)
	}
	
	// Calculate discount
	var discountAmount float64
	if req.DiscountType == "percentage" {
		discountAmount = subtotal * req.DiscountValue / 100
	} else if req.DiscountType == "fixed" {
		discountAmount = req.DiscountValue
	}
	
	// Calculate total
	totalAmount := subtotal + totalTax - discountAmount
	grossProfit := subtotal - totalCost
	netProfit := grossProfit - totalTax - discountAmount
	
	// Create sale
	sale := &Sale{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		InvoiceNumber:  invoiceNumber,
		CustomerID:     req.CustomerID,
		UserID:         userID,
		SaleDate:       time.Now(),
		Subtotal:       subtotal,
		TaxAmount:      totalTax,
		DiscountAmount: discountAmount,
		TotalAmount:    totalAmount,
		CostAmount:     totalCost,
		GrossProfit:    grossProfit,
		NetProfit:      netProfit,
		PaidAmount:     0,
		PaymentMethod:  req.PaymentMethod,
		PaymentStatus:  "pending",
		Status:         "completed",
		Notes:          req.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Create sale in database
	saleQuery := `
		INSERT INTO sales (id, organization_id, invoice_number, customer_id, user_id, sale_date,
			subtotal, tax_amount, discount_amount, total_amount, cost_amount, gross_profit, net_profit,
			paid_amount, payment_method, payment_status, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`
	_, err = tx.ExecContext(ctx, saleQuery,
		sale.ID, sale.OrganizationID, sale.InvoiceNumber, sale.CustomerID, sale.UserID, sale.SaleDate,
		sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount, sale.CostAmount, 
		sale.GrossProfit, sale.NetProfit, sale.PaidAmount, sale.PaymentMethod, sale.PaymentStatus,
		sale.Status, sale.Notes, sale.CreatedAt, sale.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create sale: %w", err)
	}
	
	// Create sale items and update inventory items
	for i := range items {
		items[i].SaleID = sale.ID
		
		// Create sale item
		itemQuery := `
			INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_price, unit_cost, 
				tax_amount, total_amount, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = tx.ExecContext(ctx, itemQuery,
			items[i].ID, items[i].SaleID, items[i].ProductID, items[i].Quantity,
			items[i].UnitPrice, items[i].UnitCost, items[i].TaxAmount, items[i].TotalAmount, items[i].CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to create sale item: %w", err)
		}
		
		// Update inventory items (mark specific items as SOLD)
		availableItems := itemStockMap[items[i].ProductID]
		for j := 0; j < items[i].Quantity && j < len(availableItems); j++ {
			itemID := availableItems[j].ID
			
			// Update inventory item status to SOLD
			updateItemQuery := `
				UPDATE inventory_items 
				SET status = 'SOLD', sold_at = NOW(), updated_at = NOW()
				WHERE id = $1 AND organization_id = $2
			`
			_, err = tx.ExecContext(ctx, updateItemQuery, itemID, organizationID)
			if err != nil {
				return nil, fmt.Errorf("failed to update inventory item: %w", err)
			}
			
			// Create inventory movement record for each item
			movementQuery := `
				INSERT INTO inventory_movements (id, organization_id, item_id, product_id, movement_type, 
					quantity, before_quantity, after_quantity, reference_type, reference_id, 
					reason, created_by, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			`
			_, err = tx.ExecContext(ctx, movementQuery,
				uuid.New(), organizationID, itemID, items[i].ProductID, "SALE",
				-1, 1, 0, "sale", sale.ID, "Sale: "+invoiceNumber, userID, time.Now())
			if err != nil {
				return nil, fmt.Errorf("failed to create inventory movement: %w", err)
			}
		}
		
		// Also update the aggregate inventory table for backward compatibility
		inventoryUpdateQuery := `
			UPDATE inventory 
			SET quantity = quantity - $1, updated_at = NOW()
			WHERE product_id = $2 AND organization_id = $3
		`
		_, err = tx.ExecContext(ctx, inventoryUpdateQuery, items[i].Quantity, items[i].ProductID, organizationID)
		if err != nil {
			// Log but don't fail if inventory table doesn't exist or has no record
			fmt.Printf("Warning: failed to update aggregate inventory: %v\n", err)
		}
	}
	
	// Update customer ledger if customer exists
	if req.CustomerID != nil {
		// Get current balance
		var currentBalance float64
		balanceQuery := `
			SELECT COALESCE(SUM(amount), 0) 
			FROM customer_ledger 
			WHERE customer_id = $1 AND organization_id = $2
		`
		err = tx.GetContext(ctx, &currentBalance, balanceQuery, *req.CustomerID, organizationID)
		if err != nil {
			currentBalance = 0
		}
		
		// Calculate new balance
		newBalance := currentBalance + totalAmount
		
		ledgerQuery := `
			INSERT INTO customer_ledger (id, organization_id, customer_id, transaction_type, 
				amount, balance, reference_type, reference_id, description, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err = tx.ExecContext(ctx, ledgerQuery,
			uuid.New(), organizationID, *req.CustomerID, "SALE",
			totalAmount, newBalance, "sale", sale.ID, "Sale: "+invoiceNumber, userID, time.Now())
		if err != nil {
			return nil, fmt.Errorf("failed to update customer ledger: %w", err)
		}
		
		// Update customer current balance
		updateCustomerQuery := `
			UPDATE customers 
			SET current_balance = $1, updated_at = NOW()
			WHERE id = $2 AND organization_id = $3
		`
		_, err = tx.ExecContext(ctx, updateCustomerQuery, newBalance, *req.CustomerID, organizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to update customer balance: %w", err)
		}
	}
	
	// Create payment record if payment is provided
	if req.PaymentAmount > 0 {
		paymentQuery := `
			INSERT INTO payments (id, organization_id, sale_id, customer_id, amount, 
				payment_method, payment_status, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = tx.ExecContext(ctx, paymentQuery,
			uuid.New(), organizationID, sale.ID, req.CustomerID, req.PaymentAmount,
			req.PaymentMethod, "completed", userID, time.Now())
		if err != nil {
			return nil, fmt.Errorf("failed to create payment: %w", err)
		}
		
		// Update sale paid amount
		sale.PaidAmount = req.PaymentAmount
		if sale.PaidAmount >= sale.TotalAmount {
			sale.PaymentStatus = "paid"
		} else {
			sale.PaymentStatus = "partial"
		}
		
		updateSaleQuery := `UPDATE sales SET paid_amount = $1, payment_status = $2 WHERE id = $3`
		_, err = tx.ExecContext(ctx, updateSaleQuery, sale.PaidAmount, sale.PaymentStatus, sale.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update sale payment: %w", err)
		}
	}
	
	// Create warranty records for items if applicable
	if req.CustomerID != nil {
		for _, item := range items {
			warrantyQuery := `
				INSERT INTO warranties (id, organization_id, sale_id, product_id, 
					warranty_period, expires_at, terms, is_active, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`
			warrantyEnd := time.Now().AddDate(1, 0, 0) // 1 year warranty
			_, err = tx.ExecContext(ctx, warrantyQuery,
				uuid.New(), organizationID, sale.ID, item.ProductID,
				12, warrantyEnd, "Standard 1-year warranty", true, time.Now(), time.Now())
			if err != nil {
				// Log but don't fail the sale for warranty creation
				fmt.Printf("Warning: failed to create warranty: %v\n", err)
			}
		}
	}
	
	// Create audit log
	auditQuery := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, 
			entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	changes := fmt.Sprintf("Created sale %s with %d items, total: %.2f", invoiceNumber, len(items), totalAmount)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), organizationID, userID, "CREATE_SALE", "sale", sale.ID,
		changes, time.Now())
	if err != nil {
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return sale, nil
}

// calculateProfit calculates the profit for sale items
func (s *Service) calculateProfit(ctx context.Context, items []SaleItem) (float64, error) {
	totalRevenue := 0.0
	totalCost := 0.0
	
	for _, item := range items {
		// Get product cost
		cost, err := s.repo.GetProductCost(ctx, item.ProductID)
		if err != nil {
			return 0, err
		}
		
		itemRevenue := item.TotalAmount
		itemCost := float64(item.Quantity) * cost
		
		totalRevenue += itemRevenue
		totalCost += itemCost
	}
	
	profit := totalRevenue - totalCost
	return profit, nil
}

// GetSale retrieves a sale by ID with its items
func (s *Service) GetSale(ctx context.Context, id uuid.UUID) (*SaleWithItems, error) {
	sale, err := s.repo.GetSaleByID(ctx, id)
	if err != nil {
		return nil, ErrSaleNotFound
	}
	
	items, err := s.repo.GetSaleItems(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Calculate profit
	profit, err := s.calculateProfit(ctx, items)
	if err != nil {
		profit = 0
	}
	
	return &SaleWithItems{
		Sale:   sale,
		Items:  items,
		Profit: profit,
	}, nil
}

// ListSales retrieves sales with pagination and filters
func (s *Service) ListSales(ctx context.Context, organizationID uuid.UUID, page, perPage int, filters map[string]interface{}) ([]Sale, int, error) {
	return s.repo.ListSales(ctx, organizationID, page, perPage, filters)
}

// UpdateSalePayment updates the payment information for a sale with full automation
func (s *Service) UpdateSalePayment(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, id uuid.UUID, amount float64, paymentMethod string) error {
	// Start database transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	
	// Get sale with row lock
	var sale Sale
	saleQuery := `SELECT * FROM sales WHERE id = $1 FOR UPDATE`
	err = tx.GetContext(ctx, &sale, saleQuery, id)
	if err != nil {
		return ErrSaleNotFound
	}
	
	if amount <= 0 {
		return ErrInvalidPayment
	}
	
	newPaidAmount := sale.PaidAmount + amount
	if newPaidAmount > sale.TotalAmount {
		return ErrInvalidPayment
	}
	
	sale.PaidAmount = newPaidAmount
	sale.PaymentMethod = &paymentMethod
	
	// Update payment status
	if newPaidAmount >= sale.TotalAmount {
		sale.PaymentStatus = "paid"
	} else if newPaidAmount > 0 {
		sale.PaymentStatus = "partial"
	}
	
	// Update sale
	updateSaleQuery := `
		UPDATE sales SET paid_amount = $1, payment_method = $2, payment_status = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err = tx.ExecContext(ctx, updateSaleQuery, sale.PaidAmount, sale.PaymentMethod, sale.PaymentStatus, sale.ID)
	if err != nil {
		return fmt.Errorf("failed to update sale: %w", err)
	}
	
	// Create payment record
	paymentQuery := `
		INSERT INTO payments (id, organization_id, sale_id, customer_id, amount, 
			payment_method, payment_status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.ExecContext(ctx, paymentQuery,
		uuid.New(), organizationID, sale.ID, sale.CustomerID, amount,
		paymentMethod, "completed", userID, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	
	// Update customer ledger if customer exists
	if sale.CustomerID != nil {
		// Get current balance
		var currentBalance float64
		balanceQuery := `
			SELECT COALESCE(SUM(amount), 0) 
			FROM customer_ledger 
			WHERE customer_id = $1 AND organization_id = $2
		`
		err = tx.GetContext(ctx, &currentBalance, balanceQuery, *sale.CustomerID, organizationID)
		if err != nil {
			currentBalance = 0
		}
		
		// Calculate new balance (payment reduces debt)
		newBalance := currentBalance - amount
		
		ledgerQuery := `
			INSERT INTO customer_ledger (id, organization_id, customer_id, transaction_type, 
				amount, balance, reference_type, reference_id, description, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err = tx.ExecContext(ctx, ledgerQuery,
			uuid.New(), organizationID, *sale.CustomerID, "PAYMENT",
			-amount, newBalance, "payment", sale.ID, "Payment for sale "+sale.InvoiceNumber, userID, time.Now())
		if err != nil {
			return fmt.Errorf("failed to update customer ledger: %w", err)
		}
		
		// Update customer current balance
		updateCustomerQuery := `
			UPDATE customers 
			SET current_balance = $1, updated_at = NOW()
			WHERE id = $2 AND organization_id = $3
		`
		_, err = tx.ExecContext(ctx, updateCustomerQuery, newBalance, *sale.CustomerID, organizationID)
		if err != nil {
			return fmt.Errorf("failed to update customer balance: %w", err)
		}
	}
	
	// Create audit log
	auditQuery := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity_type, 
			entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	changes := fmt.Sprintf("Payment of %.2f for sale %s", amount, sale.InvoiceNumber)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), organizationID, userID, "ADD_PAYMENT", "sale", sale.ID,
		changes, time.Now())
	if err != nil {
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// CancelSale cancels a sale
func (s *Service) CancelSale(ctx context.Context, id uuid.UUID) error {
	sale, err := s.repo.GetSaleByID(ctx, id)
	if err != nil {
		return ErrSaleNotFound
	}
	
	if sale.Status == "cancelled" {
		return ErrInvalidSaleStatus
	}
	
	sale.Status = "cancelled"
	return s.repo.UpdateSale(ctx, sale)
}

// GetSalesSummary retrieves sales summary for a period
func (s *Service) GetSalesSummary(ctx context.Context, organizationID uuid.UUID, startDate, endDate string) (*SalesSummary, error) {
	return s.repo.GetSalesSummary(ctx, organizationID, startDate, endDate)
}

// GetTopSellingProducts retrieves top selling products
func (s *Service) GetTopSellingProducts(ctx context.Context, organizationID uuid.UUID, limit int) ([]TopSellingProduct, error) {
	return s.repo.GetTopSellingProducts(ctx, organizationID, limit)
}

// generateInvoiceNumber generates a unique invoice number
func (s *Service) generateInvoiceNumber(organizationID uuid.UUID) string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("INV-%s-%s", organizationID.String()[:8], timestamp)
}

// SaleWithItems represents a sale with its items and profit
type SaleWithItems struct {
	Sale   *Sale      `json:"sale"`
	Items  []SaleItem `json:"items"`
	Profit float64    `json:"profit"`
}

// CreateTransaction creates a new financial transaction
func (s *Service) CreateTransaction(ctx context.Context, organizationID uuid.UUID, tx *Transaction) error {
	tx.ID = uuid.New()
	tx.OrganizationID = organizationID
	tx.Status = "completed"
	tx.CreatedAt = time.Now()
	tx.UpdatedAt = time.Now()
	
	return s.repo.CreateTransaction(ctx, tx)
}

// CreateSaleTransaction creates a transaction for a sale
func (s *Service) CreateSaleTransaction(ctx context.Context, sale *Sale, profit float64) error {
	// Create revenue transaction
	revenueTx := &Transaction{
		ID:            uuid.New(),
		OrganizationID: sale.OrganizationID,
		SaleID:        &sale.ID,
		Type:          "sale",
		Amount:        sale.TotalAmount,
		Currency:      "USD",
		Reference:     sale.InvoiceNumber,
		Description:   stringPtr(fmt.Sprintf("Sale - %s", sale.InvoiceNumber)),
		DebitAccount:  "accounts_receivable",
		CreditAccount: "sales_revenue",
		Status:        "completed",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	if err := s.repo.CreateTransaction(ctx, revenueTx); err != nil {
		return err
	}
	
	// Create cost transaction if profit is calculated
	if profit > 0 {
		cost := sale.TotalAmount - profit
		costTx := &Transaction{
			ID:            uuid.New(),
			OrganizationID: sale.OrganizationID,
			SaleID:        &sale.ID,
			Type:          "sale",
			Amount:        cost,
			Currency:      "USD",
			Reference:     sale.InvoiceNumber,
			Description:   stringPtr(fmt.Sprintf("Cost of goods sold - %s", sale.InvoiceNumber)),
			DebitAccount:  "cost_of_goods_sold",
			CreditAccount: "inventory",
			Status:        "completed",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		
		if err := s.repo.CreateTransaction(ctx, costTx); err != nil {
			return err
		}
	}
	
	return nil
}

// GetTransaction retrieves a transaction by ID
func (s *Service) GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	return s.repo.GetTransactionByID(ctx, id)
}

// ListTransactions retrieves transactions with pagination and filters
func (s *Service) ListTransactions(ctx context.Context, organizationID uuid.UUID, page, perPage int, filters map[string]interface{}) ([]Transaction, int, error) {
	return s.repo.ListTransactions(ctx, organizationID, page, perPage, filters)
}

// CalculateProfitForPeriod calculates profit for a specific period
func (s *Service) CalculateProfitForPeriod(ctx context.Context, organizationID uuid.UUID, period string, startDate, endDate time.Time) (*ProfitEntry, error) {
	// Get sales summary
	summary, err := s.repo.GetSalesSummary(ctx, organizationID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	
	// Calculate total cost (this would need to be calculated from inventory movements)
	// For now, we'll estimate it as 70% of revenue
	totalCost := summary.TotalRevenue * 0.7
	
	grossProfit := summary.TotalRevenue - totalCost
	netProfit := grossProfit // In a real system, you'd subtract expenses, taxes, etc.
	margin := 0.0
	if summary.TotalRevenue > 0 {
		margin = (grossProfit / summary.TotalRevenue) * 100
	}
	
	entry := &ProfitEntry{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Period:         period,
		StartDate:      startDate,
		EndDate:        endDate,
		Revenue:        summary.TotalRevenue,
		Cost:           totalCost,
		GrossProfit:    grossProfit,
		NetProfit:      netProfit,
		Margin:         margin,
		CreatedAt:      time.Now(),
	}
	
	// Store the profit entry
	if err := s.repo.CreateProfitEntry(ctx, entry); err != nil {
		return nil, err
	}
	
	return entry, nil
}

// GetProfitEntries retrieves profit entries for a period
func (s *Service) GetProfitEntries(ctx context.Context, organizationID uuid.UUID, period string, startDate, endDate time.Time) ([]ProfitEntry, error) {
	return s.repo.GetProfitEntries(ctx, organizationID, period, startDate, endDate)
}

// GetAccountBalance retrieves the balance for a specific account
func (s *Service) GetAccountBalance(ctx context.Context, organizationID uuid.UUID, account string) (float64, error) {
	return s.repo.GetAccountBalance(ctx, organizationID, account)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}