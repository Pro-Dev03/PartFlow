package sales

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
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
func (s *Service) CreateSale(ctx context.Context, userID uuid.UUID, req *CreateSaleRequest) (*Sale, error) {
	// Start database transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	sqlNow := dbutil.NowSQL(s.db)
	metadataValue := "$4::jsonb"
	if dbutil.IsSQLite(s.db) {
		metadataValue = "$4"
	}

	// Generate invoice number
	invoiceNumber := s.generateInvoiceNumber()

	// Calculate totals and validate stock
	subtotal := 0.0
	totalTax := 0.0
	totalCost := 0.0

	// Tax and percentage discount limits are system settings, not sale inputs.
	var taxRate, maxDiscountRate float64
	if err := tx.GetContext(ctx, &taxRate, `SELECT COALESCE(CAST(value AS DOUBLE PRECISION), 0) FROM settings WHERE key = 'tax_rate'`); err != nil || taxRate < 0 || taxRate > 100 {
		taxRate = 0
	}
	if req.TaxExempt {
		taxRate = 0
	}
	if err := tx.GetContext(ctx, &maxDiscountRate, `SELECT COALESCE(CAST(value AS DOUBLE PRECISION), 15) FROM settings WHERE key = 'max_discount_rate'`); err != nil || maxDiscountRate < 0 || maxDiscountRate > 100 {
		maxDiscountRate = 15
	}
	discountsEnabled := true
	var discountsEnabledValue string
	if err := tx.GetContext(ctx, &discountsEnabledValue, `SELECT COALESCE(value, 'true') FROM settings WHERE key = 'discounts_enabled'`); err == nil {
		discountsEnabled = strings.EqualFold(strings.TrimSpace(discountsEnabledValue), "true") || strings.TrimSpace(discountsEnabledValue) == "1"
	}
	if !discountsEnabled && (req.DiscountValue > 0 || strings.TrimSpace(req.DiscountType) != "") {
		return nil, fmt.Errorf("discounts are disabled by store settings")
	}

	var items []SaleItem
	productNames := make(map[uuid.UUID]string)
	aggregateStockMap := make(map[uuid.UUID]int)
	itemStockMap := make(map[uuid.UUID][]struct {
		ID   uuid.UUID `db:"id"`
		Cost float64   `db:"purchase_cost"`
	}) // Store available items for each product

	for _, itemReq := range req.Items {
		// Check stock availability with row lock (using inventory_items for individual tracking)
		var availableItems []struct {
			ID   uuid.UUID `db:"id"`
			Cost float64   `db:"purchase_cost"`
		}
		stockQuery := `
			SELECT id, purchase_cost
			FROM inventory_items
			WHERE product_id = $1 AND UPPER(TRIM(COALESCE(status, ''))) = 'AVAILABLE'
			ORDER BY created_at ASC
		`
		var stockArgs []interface{}
		if itemReq.InventoryItemID != nil {
			stockQuery = `
				SELECT id, purchase_cost
				FROM inventory_items
				WHERE id = $1 AND product_id = $2 AND UPPER(TRIM(COALESCE(status, ''))) = 'AVAILABLE'
			`
			stockArgs = []interface{}{*itemReq.InventoryItemID, itemReq.ProductID}
		} else {
			stockArgs = []interface{}{itemReq.ProductID}
		}
		if !dbutil.IsSQLite(s.db) {
			stockQuery += " FOR UPDATE"
		}
		err := tx.SelectContext(ctx, &availableItems, stockQuery, stockArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to check stock: %w", err)
		}

		var productName string
		if err := tx.GetContext(ctx, &productName, `SELECT name FROM products WHERE id = $1`, itemReq.ProductID); err != nil {
			return nil, fmt.Errorf("failed to get product name: %w", err)
		}
		productNames[itemReq.ProductID] = productName

		if len(availableItems) < itemReq.Quantity {
			// Quantity-based products do not have one inventory_items row per unit.
			// Fall back to the aggregate inventory balance for those products.
			if len(availableItems) == 0 {
				var aggregateQuantity int
				aggregateErr := tx.GetContext(ctx, &aggregateQuantity, `SELECT COALESCE(quantity, 0) FROM inventory WHERE product_id = $1`, itemReq.ProductID)
				if aggregateErr == nil && aggregateQuantity >= itemReq.Quantity {
					aggregateStockMap[itemReq.ProductID] = aggregateQuantity
				} else {
					return nil, &InsufficientStockError{
						ProductID:   itemReq.ProductID,
						ProductName: productName,
						Requested:   itemReq.Quantity,
						Available:   aggregateQuantity,
					}
				}
			} else {
				return nil, &InsufficientStockError{
					ProductID:   itemReq.ProductID,
					ProductName: productName,
					Requested:   itemReq.Quantity,
					Available:   len(availableItems),
				}
			}
		}
		if len(availableItems) < itemReq.Quantity && aggregateStockMap[itemReq.ProductID] == 0 {
			return nil, &InsufficientStockError{
				ProductID:   itemReq.ProductID,
				ProductName: productName,
				Requested:   itemReq.Quantity,
				Available:   len(availableItems),
			}
		}

		// Store available items for this product
		itemStockMap[itemReq.ProductID] = availableItems

		// Calculate item totals using actual item costs
		itemTotal := float64(itemReq.Quantity) * itemReq.UnitPrice

		// Calculate actual cost from available items
		itemCost := 0.0
		for i := 0; i < itemReq.Quantity && i < len(availableItems); i++ {
			itemCost += availableItems[i].Cost
		}
		inventoryItemID := itemReq.InventoryItemID
		if inventoryItemID == nil && len(availableItems) > 0 {
			inventoryItemID = &availableItems[0].ID
		}

		subtotal += itemTotal
		totalCost += itemCost

		item := SaleItem{
			ID:              uuid.New(),
			ProductID:       itemReq.ProductID,
			InventoryItemID: inventoryItemID,
			Quantity:        itemReq.Quantity,
			UnitPrice:       itemReq.UnitPrice,
			UnitCost:        itemCost / float64(itemReq.Quantity),
			TaxAmount:       0,
			TotalAmount:     itemTotal,
			CreatedAt:       time.Now(),
		}
		items = append(items, item)
	}

	// Calculate discount
	discountAmount, totalTax, totalAmount, grossProfit, netProfit := calculateSaleAmounts(
		subtotal, totalCost, taxRate, maxDiscountRate, req.DiscountType, req.DiscountValue,
	)
	if subtotal > 0 {
		for i := range items {
			itemSubtotal := items[i].TotalAmount
			itemDiscount := discountAmount * itemSubtotal / subtotal
			itemTax := addedTax(itemSubtotal-itemDiscount, taxRate)
			items[i].DiscountAmount = itemDiscount
			items[i].TaxAmount = itemTax
			items[i].TotalAmount = itemSubtotal - itemDiscount + itemTax
		}
	}

	paymentMethod := ""
	if req.PaymentMethod != nil {
		paymentMethod = *req.PaymentMethod
	}
	paymentAmount := req.PaymentAmount
	if req.PaymentTransactionID != nil {
		var externalStatus string
		var externalAmountMinor int64
		if err := tx.QueryRowContext(ctx, `SELECT status, amount_minor FROM payment_transactions WHERE id = $1`, *req.PaymentTransactionID).Scan(&externalStatus, &externalAmountMinor); err != nil {
			return nil, fmt.Errorf("payment transaction not found: %w", err)
		}
		if externalStatus != "paid" {
			return nil, fmt.Errorf("payment transaction is not verified: %s", externalStatus)
		}
		if math.Abs(float64(externalAmountMinor)/100-totalAmount) > 0.01 {
			return nil, fmt.Errorf("verified payment amount does not match sale total")
		}
		paymentMethod = "card"
		paymentAmount = totalAmount
	}
	cashReceived := req.CashReceived
	if strings.EqualFold(paymentMethod, "cash") {
		if cashReceived <= 0 {
			cashReceived = paymentAmount
		}
		if cashReceived < paymentAmount {
			return nil, fmt.Errorf("cash received cannot be less than applied payment")
		}
		if cashReceived > paymentAmount {
			paymentAmount = math.Min(cashReceived, totalAmount)
		}
	}
	if paymentAmount < 0 || paymentAmount > totalAmount {
		return nil, fmt.Errorf("payment amount must be between 0 and the sale total")
	}
	changeAmount := 0.0
	if strings.EqualFold(paymentMethod, "cash") && cashReceived > totalAmount {
		changeAmount = cashReceived - totalAmount
	}
	isDebtSale := paymentMethod == "debt"
	if isDebtSale && req.CustomerID == nil {
		return nil, fmt.Errorf("credit sales require a customer")
	}
	paymentStatus := "pending"
	if totalAmount > 0 && paymentAmount >= totalAmount {
		paymentStatus = "paid"
	} else if isDebtSale && totalAmount > paymentAmount {
		paymentStatus = "debt"
	} else if paymentAmount > 0 {
		paymentStatus = "partial"
	}
	allocations := req.PaymentAllocations
	if len(allocations) == 0 && paymentAmount > 0 {
		method := paymentMethod
		if strings.EqualFold(method, "transfer") {
			method = "checks"
		}
		allocations = []PaymentAllocationRequest{{Amount: paymentAmount, Method: method}}
	}
	allocationTotal := 0.0
	for _, allocation := range allocations {
		if allocation.Amount <= 0 {
			return nil, fmt.Errorf("payment allocation amount must be greater than zero")
		}
		if strings.EqualFold(allocation.Method, "checks") && strings.TrimSpace(allocation.CheckNumber) == "" {
			return nil, fmt.Errorf("check number is required for cheque payments")
		}
		allocationTotal += allocation.Amount
	}
	if len(allocations) > 0 && math.Abs(allocationTotal-paymentAmount) > 0.01 {
		return nil, fmt.Errorf("payment allocations must equal the applied payment amount")
	}

	// Create sale - allow nil user_id for testing
	var userIDPtr *uuid.UUID
	if userID != uuid.Nil {
		userIDPtr = &userID
	}

	eventTime := time.Now()
	storeDate, err := accounting.StoreDate(eventTime)
	if err != nil {
		return nil, fmt.Errorf("calculate sale store date: %w", err)
	}
	saleDate, err := time.Parse("2006-01-02", storeDate)
	if err != nil {
		return nil, fmt.Errorf("parse sale store date: %w", err)
	}

	sale := &Sale{
		ID:             uuid.New(),
		InvoiceNumber:  invoiceNumber,
		CustomerID:     req.CustomerID,
		UserID:         userIDPtr,
		SaleDate:       saleDate,
		Subtotal:       subtotal,
		TaxAmount:      totalTax,
		DiscountAmount: discountAmount,
		TotalAmount:    totalAmount,
		CostAmount:     totalCost,
		GrossProfit:    grossProfit,
		NetProfit:      netProfit,
		PaidAmount:     paymentAmount,
		CashReceived:   cashReceived,
		ChangeAmount:   changeAmount,
		PaymentMethod:  req.PaymentMethod,
		PaymentStatus:  paymentStatus,
		Status:         "completed",
		Notes:          req.Notes,
		CreatedAt:      eventTime,
		UpdatedAt:      eventTime,
	}

	// Create sale in database
	saleQuery := `
		INSERT INTO sales (id, sale_date, customer_id, invoice_number, user_id,
			subtotal, tax_amount, discount_amount, total_amount, cost_amount, gross_profit, net_profit,
			paid_amount, payment_method, payment_status, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`
	hasCashColumns := true
	if dbutil.IsSQLite(s.db) {
		var cashColumnCount int
		if err := tx.GetContext(ctx, &cashColumnCount, `SELECT COUNT(*) FROM pragma_table_info('sales') WHERE name IN ('cash_received', 'change_amount')`); err != nil {
			return nil, fmt.Errorf("failed to inspect cash change schema: %w", err)
		}
		hasCashColumns = cashColumnCount == 2
	}
	if dbutil.IsSQLite(s.db) && hasCashColumns {
		_, err = tx.ExecContext(ctx, `INSERT INTO sales (id, sale_number, invoice_number, sale_date, customer_id, user_id, subtotal, tax_amount, discount_amount, total_amount, cost_amount, gross_profit, net_profit, paid_amount, cash_received, change_amount, remaining_amount, payment_method, payment_status, status, notes, created_at, updated_at) VALUES ($1, $2, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`,
			sale.ID, sale.InvoiceNumber, sale.SaleDate.UTC().Format(time.RFC3339Nano), sale.CustomerID, sale.UserID,
			sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount, sale.CostAmount,
			sale.GrossProfit, sale.NetProfit, sale.PaidAmount, sale.CashReceived, sale.ChangeAmount,
			sale.TotalAmount-sale.PaidAmount, sale.PaymentMethod, sale.PaymentStatus, sale.Status, sale.Notes,
			sale.CreatedAt.UTC().Format(time.RFC3339Nano), sale.UpdatedAt.UTC().Format(time.RFC3339Nano))
	} else if dbutil.IsSQLite(s.db) {
		// SQLite stores timestamps as text. Format them explicitly so the driver
		// does not persist Go's internal monotonic-clock suffix (m=...).
		_, err = tx.ExecContext(ctx, `INSERT INTO sales (id, sale_number, invoice_number, sale_date, customer_id, user_id, subtotal, tax_amount, discount_amount, total_amount, cost_amount, gross_profit, net_profit, paid_amount, remaining_amount, payment_method, payment_status, status, notes, created_at, updated_at) VALUES ($1, $2, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`,
			sale.ID, sale.InvoiceNumber, sale.SaleDate.UTC().Format(time.RFC3339Nano), sale.CustomerID, sale.UserID,
			sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount, sale.CostAmount,
			sale.GrossProfit, sale.NetProfit, sale.PaidAmount, sale.TotalAmount-sale.PaidAmount,
			sale.PaymentMethod, sale.PaymentStatus, sale.Status, sale.Notes,
			sale.CreatedAt.UTC().Format(time.RFC3339Nano), sale.UpdatedAt.UTC().Format(time.RFC3339Nano))
	} else {
		if hasCashColumns {
			saleQuery = `
				INSERT INTO sales (id, sale_date, customer_id, invoice_number, user_id,
					subtotal, tax_amount, discount_amount, total_amount, cost_amount, gross_profit, net_profit,
					paid_amount, cash_received, change_amount, payment_method, payment_status, status, notes, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`
			_, err = tx.ExecContext(ctx, saleQuery,
				sale.ID, sale.SaleDate, sale.CustomerID, sale.InvoiceNumber, sale.UserID,
				sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount, sale.CostAmount,
				sale.GrossProfit, sale.NetProfit, sale.PaidAmount, sale.CashReceived, sale.ChangeAmount,
				sale.PaymentMethod, sale.PaymentStatus, sale.Status, sale.Notes, sale.CreatedAt, sale.UpdatedAt)
		} else {
			_, err = tx.ExecContext(ctx, saleQuery,
				sale.ID, sale.SaleDate, sale.CustomerID, sale.InvoiceNumber, sale.UserID,
				sale.Subtotal, sale.TaxAmount, sale.DiscountAmount, sale.TotalAmount, sale.CostAmount,
				sale.GrossProfit, sale.NetProfit, sale.PaidAmount, sale.PaymentMethod, sale.PaymentStatus,
				sale.Status, sale.Notes, sale.CreatedAt, sale.UpdatedAt)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create sale: %w", err)
	}

	// Create sale items and update inventory items
	for i := range items {
		items[i].SaleID = sale.ID

		// Get supplier_id from the first inventory item being sold
		var supplierID *uuid.UUID
		availableItems := itemStockMap[items[i].ProductID]
		if items[i].InventoryItemID != nil {
			availableItems = []struct {
				ID   uuid.UUID `db:"id"`
				Cost float64   `db:"purchase_cost"`
			}{{ID: *items[i].InventoryItemID, Cost: items[i].UnitCost}}
		}
		if len(availableItems) > 0 {
			// Query supplier_id from inventory_items
			var supplierIDFromDB *uuid.UUID
			supplierQuery := `SELECT supplier_id FROM inventory_items WHERE id = $1`
			err := tx.GetContext(ctx, &supplierIDFromDB, supplierQuery, availableItems[0].ID)
			if err == nil && supplierIDFromDB != nil {
				supplierID = supplierIDFromDB
			}
		}

		// Local snapshots use item_total, while the cloud schema uses total_amount.
		itemQuery := `
			INSERT INTO sale_items (id, sale_id, product_id, inventory_item_id, quantity, unit_price, unit_cost,
				tax_amount, total_amount, supplier_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		itemArgs := []interface{}{
			items[i].ID, items[i].SaleID, items[i].ProductID, items[i].InventoryItemID, items[i].Quantity,
			items[i].UnitPrice, items[i].UnitCost, items[i].TaxAmount, items[i].TotalAmount,
			supplierID, items[i].CreatedAt,
		}
		if dbutil.IsSQLite(s.db) {
			var itemTotalColumns int
			if err := tx.GetContext(ctx, &itemTotalColumns, `SELECT COUNT(*) FROM pragma_table_info('sale_items') WHERE name = 'item_total'`); err != nil {
				return nil, fmt.Errorf("failed to inspect sale item schema: %w", err)
			}
			if itemTotalColumns > 0 {
				itemQuery = `
					INSERT INTO sale_items (id, sale_id, product_id, inventory_item_id, quantity, unit_price, item_total, unit_cost, discount_amount, tax_amount, total_amount, supplier_id, created_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
				`
				itemArgs = []interface{}{
					items[i].ID, items[i].SaleID, items[i].ProductID, items[i].InventoryItemID, items[i].Quantity,
					items[i].UnitPrice, items[i].TotalAmount, items[i].UnitCost, items[i].DiscountAmount,
					items[i].TaxAmount, items[i].TotalAmount, supplierID, items[i].CreatedAt,
				}
			}
			if dbutil.IsSQLite(s.db) {
				itemArgs[len(itemArgs)-1] = items[i].CreatedAt.UTC().Format(time.RFC3339Nano)
			}
		}
		_, err = tx.ExecContext(ctx, itemQuery, itemArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to create sale item: %w", err)
		}

		// Update inventory items (delete specific items for trade-ins, mark others as SOLD)
		for j := 0; j < items[i].Quantity && j < len(availableItems); j++ {
			itemID := availableItems[j].ID

			// Preserve every inventory item for auditability; status is the source of truth.
			var itemCondition string
			conditionQuery := `SELECT COALESCE(condition, '') FROM inventory_items WHERE id = $1`
			err := tx.GetContext(ctx, &itemCondition, conditionQuery, itemID)
			if err != nil {
				return nil, fmt.Errorf("failed to check item condition: %w", err)
			}

			updateItemQuery := fmt.Sprintf(`
				UPDATE inventory_items
					SET status = 'SOLD', sold_at = %s, updated_at = %s
					WHERE id = $1 AND UPPER(TRIM(COALESCE(status, ''))) = 'AVAILABLE'
			`, sqlNow, sqlNow)
			result, err := tx.ExecContext(ctx, updateItemQuery, itemID)
			if err != nil {
				return nil, fmt.Errorf("failed to update inventory item: %w", err)
			}
			if affected, _ := result.RowsAffected(); affected != 1 {
				return nil, &InsufficientStockError{
					ProductID:   items[i].ProductID,
					ProductName: productNames[items[i].ProductID],
					Requested:   items[i].Quantity,
					Available:   0,
				}
			}

			movementQuery := `
				INSERT INTO inventory_movements (id, item_id, movement_type,
					quantity, before_quantity, after_quantity, reference_type, reference_id,
					reason, created_by, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			`
			reason := "Sale: " + invoiceNumber
			if itemCondition == "USED" {
				reason = "Sold used item: " + invoiceNumber
			}
			var afterQuantity int
			if err = tx.GetContext(ctx, &afterQuantity, `
				SELECT COUNT(*) FROM inventory_items
				WHERE product_id = $1 AND UPPER(TRIM(COALESCE(status, ''))) = 'AVAILABLE'
			`, items[i].ProductID); err != nil {
				return nil, fmt.Errorf("failed to read inventory quantity: %w", err)
			}
			beforeQuantity := afterQuantity + 1
			_, err = tx.ExecContext(ctx, movementQuery,
				uuid.New(), itemID, "SALE",
				-1, beforeQuantity, afterQuantity, "sale", sale.ID, reason, userID, time.Now())
			if err != nil {
				return nil, fmt.Errorf("failed to create inventory movement: %w", err)
			}

			_, err = tx.ExecContext(ctx, fmt.Sprintf(`
				INSERT INTO item_history
					(inventory_item_id, event_type, event_date, reference_type, reference_id,
					 description, metadata, created_by, created_at)
				VALUES ($1, 'sold', %s, 'sale', $2, $3, %s, $5, %s)
			`, sqlNow, metadataValue, sqlNow), itemID, sale.ID, reason,
				fmt.Sprintf(`{"sale_id":"%s","unit_price":%.2f}`, sale.ID, items[i].UnitPrice), userID)
			if err != nil {
				if !(dbutil.IsSQLite(s.db) && strings.Contains(strings.ToLower(err.Error()), "no such table: item_history")) {
					return nil, fmt.Errorf("failed to create item history: %w", err)
				}
				fmt.Printf("Warning: item_history table is unavailable; continuing sale: %v\n", err)
			}

			_, err = tx.ExecContext(ctx, fmt.Sprintf(`
				UPDATE acquisition_items
				SET item_status = 'sold', updated_at = %s
				WHERE inventory_item_id = $1
			`, sqlNow), itemID)
			if err != nil {
				return nil, fmt.Errorf("failed to update acquired item status: %w", err)
			}
		}

		// Also update the aggregate inventory table for backward compatibility
		inventoryUpdateQuery := fmt.Sprintf(`
			UPDATE inventory
			SET quantity = quantity - $1, updated_at = %s
			WHERE product_id = $2
		`, sqlNow)
		_, inventoryErr := tx.ExecContext(ctx, inventoryUpdateQuery, items[i].Quantity, items[i].ProductID)
		if inventoryErr != nil {
			// Log but don't fail if inventory table doesn't exist or has no record
			fmt.Printf("Warning: failed to update aggregate inventory: %v\n", inventoryErr)
		}

		if beforeQuantity, isAggregate := aggregateStockMap[items[i].ProductID]; isAggregate {
			var currentQuantity int
			if err := tx.GetContext(ctx, &currentQuantity, `SELECT COALESCE(quantity, 0) FROM inventory WHERE product_id = $1`, items[i].ProductID); err != nil {
				return nil, fmt.Errorf("failed to read aggregate inventory quantity: %w", err)
			}
			if currentQuantity < items[i].Quantity {
				return nil, &InsufficientStockError{
					ProductID:   items[i].ProductID,
					ProductName: productNames[items[i].ProductID],
					Requested:   items[i].Quantity,
					Available:   currentQuantity,
				}
			}
			afterQuantity := currentQuantity - items[i].Quantity
			_, err := tx.ExecContext(ctx, fmt.Sprintf(`
				UPDATE inventory SET quantity = quantity - $1, updated_at = %s WHERE product_id = $2
			`, dbutil.NowSQL(s.db)), items[i].Quantity, items[i].ProductID)
			if err != nil {
				return nil, fmt.Errorf("failed to decrement aggregate inventory quantity: %w", err)
			}
			_, err = tx.ExecContext(ctx, fmt.Sprintf(`
				INSERT INTO inventory_movements (id, item_id, product_id, movement_type, quantity, before_quantity, after_quantity, reference_type, reference_id, reason, created_by, created_at)
				VALUES ($1, NULL, $2, 'SALE', $3, $4, $5, 'sale', $6, $7, $8, %s)
			`, dbutil.NowSQL(s.db)), uuid.New(), items[i].ProductID, -items[i].Quantity, beforeQuantity, afterQuantity, sale.ID, "Sale: "+invoiceNumber, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to create aggregate inventory movement: %w", err)
			}
		}
	}

	// Update customer ledger if customer exists (SALES-PHILOSOPHY.md - automatic debt calculation)
	if req.CustomerID != nil {
		// Get current balance
		var currentBalance float64
		balanceQuery := `
			SELECT COALESCE(SUM(amount), 0)
			FROM customer_ledger
			WHERE customer_id = $1
		`
		balanceErr := tx.GetContext(ctx, &currentBalance, balanceQuery, *req.CustomerID)
		if balanceErr != nil {
			currentBalance = 0
		}

		debtAmount := totalAmount - paymentAmount
		newBalance := currentBalance + debtAmount
		if debtAmount > 0 {
			ledgerQuery := `
				INSERT INTO customer_ledger (id, customer_id, type,
					amount, balance, reference_id, description, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`
			_, ledgerErr := tx.ExecContext(ctx, ledgerQuery,
				uuid.New(), *req.CustomerID, "debit",
				debtAmount, newBalance, sale.ID, "Sale: "+invoiceNumber, time.Now())
			if ledgerErr != nil {
				return nil, fmt.Errorf("failed to update customer ledger: %w", ledgerErr)
			}

			updateCustomerQuery := fmt.Sprintf(`
				UPDATE customers
				SET current_balance = $1, updated_at = %s
				WHERE id = $2
			`, dbutil.NowSQL(s.db))
			_, customerErr := tx.ExecContext(ctx, updateCustomerQuery, newBalance, *req.CustomerID)
			if customerErr != nil {
				return nil, fmt.Errorf("failed to update customer balance: %w", customerErr)
			}

			if isDebtSale {
				debtStatus := "pending"
				debtQuery := fmt.Sprintf(`
					INSERT INTO debts (id, customer_id, sale_id, amount, paid_amount, remaining_amount, due_date, status, notes, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, %s, %s)
				`, dbutil.NowSQL(s.db), dbutil.NowSQL(s.db))
				if _, debtErr := tx.ExecContext(ctx, debtQuery,
					uuid.New(), *req.CustomerID, sale.ID, debtAmount, 0, debtAmount,
					time.Now().AddDate(0, 0, 30), debtStatus, "Sale: "+invoiceNumber); debtErr != nil {
					return nil, fmt.Errorf("failed to create sale debt: %w", debtErr)
				}
			}
		}

		// Check credit limit (SALES-PHILOSOPHY.md)
		var creditLimit *float64
		limitQuery := `SELECT credit_limit FROM customers WHERE id = $1`
		limitErr := tx.GetContext(ctx, &creditLimit, limitQuery, *req.CustomerID)
		if limitErr == nil && creditLimit != nil && *creditLimit > 0 {
			if newBalance > *creditLimit {
				// Log warning but don't fail the sale (store owner's decision)
				fmt.Printf("Warning: Customer %s will exceed credit limit. Current: %.2f, Limit: %.2f, New: %.2f\n",
					*req.CustomerID, currentBalance, *creditLimit, newBalance)
			}
		}
	}

	// Create payment record if payment is provided
	if paymentAmount > 0 && req.PaymentTransactionID == nil {
		paymentQuery := `
			INSERT INTO payments (id, sale_id, customer_id, amount,
				payment_method, payment_status, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		paymentID := uuid.New()
		paymentTime := time.Now()
		var paymentErr error
		if dbutil.IsSQLite(s.db) {
			_, paymentErr = tx.ExecContext(ctx, `INSERT INTO payments (id, transaction_number, sale_id, customer_id, amount, payment_method, payment_status, created_by, payment_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9, $9)`, paymentID, paymentID.String(), sale.ID, req.CustomerID, paymentAmount, req.PaymentMethod, "completed", userID, paymentTime)
		} else {
			_, paymentErr = tx.ExecContext(ctx, paymentQuery,
				paymentID, sale.ID, req.CustomerID, paymentAmount,
				req.PaymentMethod, "completed", userID, paymentTime)
		}
		if paymentErr != nil {
			return nil, fmt.Errorf("failed to create payment: %w", paymentErr)
		}
		// Update sale paid amount
		sale.PaidAmount = paymentAmount
		if sale.PaidAmount >= sale.TotalAmount {
			sale.PaymentStatus = "paid"
		} else if !isDebtSale {
			sale.PaymentStatus = "partial"
		}

		updateSaleQuery := `UPDATE sales SET paid_amount = $1, payment_status = $2 WHERE id = $3`
		_, updateSaleErr := tx.ExecContext(ctx, updateSaleQuery, sale.PaidAmount, sale.PaymentStatus, sale.ID)
		if updateSaleErr != nil {
			return nil, fmt.Errorf("failed to update sale payment: %w", updateSaleErr)
		}
		for _, allocation := range allocations {
			allocationMethod := strings.ToLower(strings.TrimSpace(allocation.Method))
			if allocationMethod == "transfer" {
				allocationMethod = "checks"
			}
			var checkDate interface{}
			if strings.TrimSpace(allocation.CheckDate) != "" {
				checkDate = allocation.CheckDate
			}
			if dbutil.IsSQLite(s.db) {
				if _, err := tx.ExecContext(ctx, `INSERT INTO sale_payment_allocations (id, sale_id, amount, payment_method, status, check_number, bank_name, check_date, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, uuid.New(), sale.ID, allocation.Amount, allocationMethod, "pending", allocation.CheckNumber, allocation.BankName, checkDate, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
					return nil, fmt.Errorf("failed to save payment allocation: %w", err)
				}
			} else if _, err := tx.ExecContext(ctx, `INSERT INTO sale_payment_allocations (id, sale_id, amount, payment_method, status, check_number, bank_name, check_date) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, uuid.New(), sale.ID, allocation.Amount, allocationMethod, "pending", allocation.CheckNumber, allocation.BankName, checkDate); err != nil {
				return nil, fmt.Errorf("failed to save payment allocation: %w", err)
			}
		}
	}
	if req.PaymentTransactionID != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE payment_transactions SET sale_id = $1, updated_at = $2 WHERE id = $3 AND status = 'paid'`, sale.ID, time.Now(), *req.PaymentTransactionID); err != nil {
			return nil, fmt.Errorf("link payment transaction to sale: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE payments SET sale_id = $1, updated_at = $2 WHERE id = $3`, sale.ID, time.Now(), *req.PaymentTransactionID); err != nil {
			return nil, fmt.Errorf("link legacy payment to sale: %w", err)
		}
	}

	// Create warranty records for items if applicable
	if req.CustomerID != nil {
		var warrantiesTable *string
		if tableErr := tx.GetContext(ctx, &warrantiesTable, `SELECT to_regclass('public.warranties')`); tableErr == nil && warrantiesTable != nil {
			for _, item := range items {
				warrantyQuery := `
					INSERT INTO warranties (id, sale_id, product_id,
						warranty_period, expires_at, terms, is_active, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				`
				warrantyEnd := time.Now().AddDate(1, 0, 0)
				_, warrantyErr := tx.ExecContext(ctx, warrantyQuery,
					uuid.New(), sale.ID, item.ProductID,
					12, warrantyEnd, "Standard 1-year warranty", true, time.Now(), time.Now())
				if warrantyErr != nil {
					return nil, fmt.Errorf("failed to create warranty: %w", warrantyErr)
				}
			}
		}
	}

	// Create audit log
	var auditLogsTable *string
	var userExists bool
	if userID != uuid.Nil {
		_ = tx.GetContext(ctx, &userExists, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, userID)
	}
	if userExists {
		if tableErr := tx.GetContext(ctx, &auditLogsTable, `SELECT to_regclass('public.audit_logs')`); tableErr == nil && auditLogsTable != nil {
			auditQuery := `
			INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at)
			VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
		`
			changes, marshalErr := json.Marshal(map[string]interface{}{
				"invoice_number": invoiceNumber,
				"item_count":     len(items),
				"total":          totalAmount,
			})
			if marshalErr != nil {
				return nil, fmt.Errorf("failed to encode audit log: %w", marshalErr)
			}
			if _, auditErr := tx.ExecContext(ctx, auditQuery,
				uuid.New(), userID, "CREATE_SALE", "sale", sale.ID,
				string(changes), time.Now()); auditErr != nil {
				return nil, fmt.Errorf("failed to create audit log: %w", auditErr)
			}
		}
	}

	// Commit transaction
	if commitErr := tx.Commit(); commitErr != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", commitErr)
	}
	committed = true

	// Invalidate dashboard cache since sales data changed
	dashboard.InvalidateDashboardCacheWithReason("sale_created")

	return sale, nil
}

func calculateSaleAmounts(subtotal, totalCost, taxRate, maxDiscountRate float64, discountType string, discountValue float64) (discountAmount, totalTax, totalAmount, grossProfit, netProfit float64) {
	if discountType == "percentage" {
		discountRate := discountValue
		if discountRate < 0 {
			discountRate = 0
		}
		if discountRate > maxDiscountRate {
			discountRate = maxDiscountRate
		}
		discountAmount = subtotal * discountRate / 100
	} else if discountType == "fixed" {
		discountAmount = discountValue
		if discountAmount < 0 {
			discountAmount = 0
		}
		if discountAmount > subtotal {
			discountAmount = subtotal
		}
	}

	taxableSubtotal := subtotal - discountAmount
	totalTax = addedTax(taxableSubtotal, taxRate)
	totalAmount = taxableSubtotal + totalTax
	netRevenue := taxableSubtotal
	grossProfit = netRevenue - totalCost
	netProfit = grossProfit
	return
}

func addedTax(netAmount, taxRate float64) float64 {
	if netAmount <= 0 || taxRate <= 0 {
		return 0
	}
	return netAmount * taxRate / 100
}

// calculateProfit calculates the profit for sale items
func (s *Service) calculateProfit(ctx context.Context, items []SaleItem) (float64, error) {
	totalRevenue := 0.0
	totalCost := 0.0

	for _, item := range items {
		itemRevenue := item.TotalAmount
		itemCost := float64(item.Quantity) * item.UnitCost

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
		return nil, fmt.Errorf("failed to read sale %s: %w", id, err)
	}

	items, err := s.repo.GetSaleItems(ctx, id)
	if err != nil {
		return nil, err
	}
	allocations, err := s.repo.GetPaymentAllocations(ctx, id)
	if err != nil {
		return nil, err
	}

	// Calculate profit
	profit, err := s.calculateProfit(ctx, items)
	if err != nil {
		profit = 0
	}

	return &SaleWithItems{
		Sale:               sale,
		Items:              items,
		PaymentAllocations: allocations,
		Profit:             profit,
	}, nil
}

// ListSales retrieves sales with pagination and filters
func (s *Service) ListSales(ctx context.Context, page, perPage int, filters map[string]interface{}) ([]Sale, int, error) {
	return s.repo.ListSales(ctx, page, perPage, filters)
}

// UpdateSalePayment updates the payment information for a sale with full automation
func (s *Service) UpdateSalePayment(ctx context.Context, userID uuid.UUID, id uuid.UUID, amount float64, paymentMethod string) error {
	// Start database transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// Get sale with row lock
	saleQuery := `SELECT id, sale_date, customer_id, invoice_number, subtotal, tax_amount, discount_amount, total_amount, cost_amount, gross_profit, net_profit, paid_amount, payment_method, payment_status, status, notes, created_at, updated_at FROM sales WHERE id = $1`
	if !dbutil.IsSQLite(s.db) {
		saleQuery += " FOR UPDATE"
	}
	var sale Sale
	if dbutil.IsSQLite(s.db) {
		var localRow localSaleRow
		err = tx.GetContext(ctx, &localRow, saleQuery, id)
		if err == nil {
			sale, err = localRow.sale()
		}
	} else {
		err = tx.GetContext(ctx, &sale, saleQuery, id)
	}
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
	updateSaleQuery := fmt.Sprintf(`
		UPDATE sales SET paid_amount = $1, payment_method = $2, payment_status = $3, updated_at = %s
		WHERE id = $4
	`, dbutil.NowSQL(s.db))
	_, err = tx.ExecContext(ctx, updateSaleQuery, sale.PaidAmount, sale.PaymentMethod, sale.PaymentStatus, sale.ID)
	if err != nil {
		return fmt.Errorf("failed to update sale: %w", err)
	}

	// Create payment record
	paymentQuery := `
		INSERT INTO payments (id, sale_id, customer_id, amount,
			payment_method, payment_status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	paymentID := uuid.New()
	paymentTime := time.Now()
	if dbutil.IsSQLite(s.db) {
		_, err = tx.ExecContext(ctx, `INSERT INTO payments (id, transaction_number, sale_id, customer_id, amount, payment_method, payment_status, created_by, payment_date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9, $9)`, paymentID, paymentID.String(), sale.ID, sale.CustomerID, amount, paymentMethod, "completed", userID, paymentTime)
	} else {
		_, err = tx.ExecContext(ctx, paymentQuery,
			paymentID, sale.ID, sale.CustomerID, amount,
			paymentMethod, "completed", userID, paymentTime, paymentTime)
	}
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
			WHERE customer_id = $1
		`
		err = tx.GetContext(ctx, &currentBalance, balanceQuery, *sale.CustomerID)
		if err != nil {
			currentBalance = 0
		}

		// Calculate new balance (payment reduces debt)
		newBalance := currentBalance - amount

		ledgerQuery := `
			INSERT INTO customer_ledger (id, customer_id, transaction_type,
				amount, balance, reference_type, reference_id, description, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		_, err = tx.ExecContext(ctx, ledgerQuery,
			uuid.New(), *sale.CustomerID, "PAYMENT",
			-amount, newBalance, "payment", sale.ID, "Payment for sale "+sale.InvoiceNumber, userID, time.Now())
		if err != nil {
			return fmt.Errorf("failed to update customer ledger: %w", err)
		}

		// Update customer current balance
		updateCustomerQuery := fmt.Sprintf(`
			UPDATE customers 
			SET current_balance = $1, updated_at = %s
			WHERE id = $2
		`, dbutil.NowSQL(s.db))
		_, err = tx.ExecContext(ctx, updateCustomerQuery, newBalance, *sale.CustomerID)
		if err != nil {
			return fmt.Errorf("failed to update customer balance: %w", err)
		}
	}

	// Create audit log
	auditQuery := `
		INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, new_values, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	changes := fmt.Sprintf("Payment of %.2f for sale %s", amount, sale.InvoiceNumber)
	_, err = tx.ExecContext(ctx, auditQuery,
		uuid.New(), userID, "ADD_PAYMENT", "sale", sale.ID,
		changes, time.Now())
	if err != nil {
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	// Invalidate dashboard cache since payment data changed
	dashboard.InvalidateDashboardCacheWithReason("payment_updated")

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
func (s *Service) GetSalesSummary(ctx context.Context, startDate, endDate string) (*SalesSummary, error) {
	return s.repo.GetSalesSummary(ctx, startDate, endDate)
}

// GetTopSellingProducts retrieves top selling products
func (s *Service) GetTopSellingProducts(ctx context.Context, limit int) ([]TopSellingProduct, error) {
	return s.repo.GetTopSellingProducts(ctx, limit)
}

// generateInvoiceNumber generates a unique invoice number with a short suffix.
func (s *Service) generateInvoiceNumber() string {
	timestamp := time.Now().Format("20060102150405")
	suffix := 0
	var buf [2]byte
	if _, err := rand.Read(buf[:]); err == nil {
		suffix = int(binary.BigEndian.Uint16(buf[:])) % 10000
	}
	if suffix == 0 {
		suffix = int(time.Now().UnixNano() % 10000)
	}
	return fmt.Sprintf("INV-%s-%04d", timestamp, suffix)
}

// SaleWithItems represents a sale with its items and profit
type SaleWithItems struct {
	Sale               *Sale               `json:"sale"`
	Items              []SaleItem          `json:"items"`
	PaymentAllocations []PaymentAllocation `json:"payment_allocations"`
	Profit             float64             `json:"profit"`
}

// CreateTransaction creates a new financial transaction
func (s *Service) CreateTransaction(ctx context.Context, tx *Transaction) error {
	tx.ID = uuid.New()
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
func (s *Service) ListTransactions(ctx context.Context, page, perPage int, filters map[string]interface{}) ([]Transaction, int, error) {
	return s.repo.ListTransactions(ctx, page, perPage, filters)
}

// CalculateProfitForPeriod calculates profit for a specific period
func (s *Service) CalculateProfitForPeriod(ctx context.Context, period string, startDate, endDate time.Time) (*ProfitEntry, error) {
	// Get sales summary
	summary, err := s.repo.GetSalesSummary(ctx, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
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
		ID:          uuid.New(),
		Period:      period,
		StartDate:   startDate,
		EndDate:     endDate,
		Revenue:     summary.TotalRevenue,
		Cost:        totalCost,
		GrossProfit: grossProfit,
		NetProfit:   netProfit,
		Margin:      margin,
		CreatedAt:   time.Now(),
	}

	// Store the profit entry
	if err := s.repo.CreateProfitEntry(ctx, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// GetProfitEntries retrieves profit entries for a period
func (s *Service) GetProfitEntries(ctx context.Context, period string, startDate, endDate time.Time) ([]ProfitEntry, error) {
	return s.repo.GetProfitEntries(ctx, period, startDate, endDate)
}

// GetAccountBalance retrieves the balance for a specific account
func (s *Service) GetAccountBalance(ctx context.Context, account string) (float64, error) {
	return s.repo.GetAccountBalance(ctx, account)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
