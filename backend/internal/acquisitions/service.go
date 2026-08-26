package acquisitions

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Service handles acquisition business logic (USED-PARTS-ACQUISITION.md)
type Service struct {
	db *sqlx.DB
}

// NewService creates a new acquisition service
func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// CreateAcquisition creates a new acquisition (from supplier or customer)
func (s *Service) CreateAcquisition(ctx context.Context, req *AcquisitionRequest, userID uuid.UUID) (*Acquisition, error) {
	// Validate that seller is set based on type
	if req.Type == TypeSupplier && req.SupplierID == nil {
		return nil, fmt.Errorf("supplier_id is required for supplier acquisitions")
	}
	if req.Type == TypeCustomer && req.CustomerID == nil {
		return nil, fmt.Errorf("customer_id is required for customer acquisitions")
	}

	// Start transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Create acquisition
	acquisitionID := uuid.New()
	acquisition := &Acquisition{
		ID:              acquisitionID,
		Type:            req.Type,
		AcquisitionDate: req.AcquisitionDate,
		SupplierID:      req.SupplierID,
		CustomerID:      req.CustomerID,
		TotalCost:       0, // Will be calculated from items
		PaidAmount:       0,
		PaymentStatus:    req.PaymentStatus,
		Status:          StatusDraft,
		Notes:           &req.Notes,
		UserID:          &userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	query := `
		INSERT INTO acquisitions (id, type, acquisition_date, supplier_id, customer_id, 
			total_cost, paid_amount, payment_status, status, notes, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING *
	`

	err = tx.Get(acquisition, query,
		acquisition.ID, acquisition.Type, acquisition.AcquisitionDate,
		acquisition.SupplierID, acquisition.CustomerID, acquisition.TotalCost,
		acquisition.PaidAmount, acquisition.PaymentStatus, acquisition.Status,
		acquisition.Notes, acquisition.UserID, acquisition.CreatedAt, acquisition.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create acquisition: %w", err)
	}

	// Create acquisition items
	for _, itemReq := range req.Items {
		item := &AcquisitionItem{
			ID:              uuid.New(),
			AcquisitionID:   acquisitionID,
			ProductID:       itemReq.ProductID,
			SerialNumber:    itemReq.SerialNumber,
			Condition:       itemReq.Condition,
			Grade:           itemReq.Grade,
			UnitCost:        itemReq.UnitCost,
			TotalCost:       itemReq.UnitCost, // Assuming quantity 1 for individual items
			InspectionStatus: "pending",
			ItemStatus:      "acquired",
			Notes:           itemReq.Notes,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		itemQuery := `
			INSERT INTO acquisition_items (id, acquisition_id, product_id, serial_number, 
				condition, grade, unit_cost, total_cost, inspection_status, item_status, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`

		_, err = tx.Exec(itemQuery,
			item.ID, item.AcquisitionID, item.ProductID, item.SerialNumber,
			item.Condition, item.Grade, item.UnitCost, item.TotalCost,
			item.InspectionStatus, item.ItemStatus, item.Notes, item.CreatedAt, item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create acquisition item: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload acquisition to get calculated total cost
	return s.GetAcquisition(ctx, acquisitionID)
}

// GetAcquisition retrieves an acquisition by ID
func (s *Service) GetAcquisition(ctx context.Context, id uuid.UUID) (*Acquisition, error) {
	var acquisition Acquisition
	query := `SELECT * FROM acquisitions WHERE id = $1`
	err := s.db.Get(&acquisition, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("acquisition not found")
		}
		return nil, fmt.Errorf("failed to get acquisition: %w", err)
	}
	return &acquisition, nil
}

// GetAcquisitionWithItems retrieves an acquisition with its items
func (s *Service) GetAcquisitionWithItems(ctx context.Context, id uuid.UUID) (*AcquisitionResponse, error) {
	// Get acquisition
	acquisition, err := s.GetAcquisition(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get items
	var items []AcquisitionItem
	query := `SELECT * FROM acquisition_items WHERE acquisition_id = $1 ORDER BY created_at`
	err = s.db.Select(&items, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get acquisition items: %w", err)
	}

	// Get seller info
	var seller *SellerInfo
	if acquisition.Type == TypeSupplier && acquisition.SupplierID != nil {
		seller, err = s.getSupplierInfo(ctx, *acquisition.SupplierID)
	} else if acquisition.Type == TypeCustomer && acquisition.CustomerID != nil {
		seller, err = s.getCustomerInfo(ctx, *acquisition.CustomerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get seller info: %w", err)
	}

	// Calculate remaining
	remaining := acquisition.TotalCost - acquisition.PaidAmount

	return &AcquisitionResponse{
		Acquisition: *acquisition,
		Items:       items,
		Seller:      seller,
		TotalItems:  len(items),
		Remaining:   remaining,
	}, nil
}

// getSupplierInfo retrieves supplier information
func (s *Service) getSupplierInfo(ctx context.Context, supplierID uuid.UUID) (*SellerInfo, error) {
	var supplier SellerInfo
	query := `
		SELECT id, 'SUPPLIER' as type, name, phone, email 
		FROM suppliers WHERE id = $1
	`
	err := s.db.Get(&supplier, query, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier info: %w", err)
	}
	return &supplier, nil
}

// getCustomerInfo retrieves customer information
func (s *Service) getCustomerInfo(ctx context.Context, customerID uuid.UUID) (*SellerInfo, error) {
	var customer SellerInfo
	query := `
		SELECT id, 'CUSTOMER' as type, name, phone, email 
		FROM customers WHERE id = $1
	`
	err := s.db.Get(&customer, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer info: %w", err)
	}
	return &customer, nil
}

// ListAcquisitions retrieves acquisitions with filtering
func (s *Service) ListAcquisitions(ctx context.Context, req *AcquisitionListRequest) ([]Acquisition, int, error) {
	// Build query
	query := `SELECT * FROM acquisitions WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM acquisitions WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Apply filters
	if req.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, req.Type)
		argCount++
	}

	if req.SupplierID != nil {
		query += fmt.Sprintf(" AND supplier_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND supplier_id = $%d", argCount)
		args = append(args, *req.SupplierID)
		argCount++
	}

	if req.CustomerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND customer_id = $%d", argCount)
		args = append(args, *req.CustomerID)
		argCount++
	}

	if req.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
		argCount++
	}

	if req.PaymentStatus != "" {
		query += fmt.Sprintf(" AND payment_status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND payment_status = $%d", argCount)
		args = append(args, req.PaymentStatus)
		argCount++
	}

	// Get total count
	var total int
	err := s.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count acquisitions: %w", err)
	}

	// Apply sorting
	sortBy := "acquisition_date"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "DESC"
	if req.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Apply pagination
	offset := (req.Page - 1) * req.PerPage
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.PerPage, offset)

	// Execute query
	var acquisitions []Acquisition
	err = s.db.Select(&acquisitions, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list acquisitions: %w", err)
	}

	return acquisitions, total, nil
}

// UpdateAcquisitionStatus updates acquisition status
func (s *Service) UpdateAcquisitionStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE acquisitions 
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := s.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update acquisition status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("acquisition not found")
	}

	return nil
}

// LinkItemToInventory links an acquisition item to an inventory item
func (s *Service) LinkItemToInventory(ctx context.Context, acquisitionItemID uuid.UUID, inventoryItemID uuid.UUID) error {
	query := `
		UPDATE acquisition_items 
		SET inventory_item_id = $1, item_status = 'available', updated_at = NOW()
		WHERE id = $2
	`
	result, err := s.db.Exec(query, inventoryItemID, acquisitionItemID)
	if err != nil {
		return fmt.Errorf("failed to link item to inventory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("acquisition item not found")
	}

	return nil
}

// CreateSellerPayment creates a payment to a seller (customer)
func (s *Service) CreateSellerPayment(ctx context.Context, req *SellerPaymentRequest, userID uuid.UUID) (*SellerPayment, error) {
	payment := &SellerPayment{
		ID:            uuid.New(),
		AcquisitionID: req.AcquisitionID,
		CustomerID:    req.CustomerID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		PaymentDate:   req.PaymentDate,
		Notes:         &req.Notes,
		UserID:        userID,
		CreatedAt:     time.Now(),
	}

	query := `
		INSERT INTO seller_payments (id, acquisition_id, customer_id, amount, payment_method, payment_date, notes, user_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING *
	`

	err := s.db.Get(payment, query,
		payment.ID, payment.AcquisitionID, payment.CustomerID,
		payment.Amount, payment.PaymentMethod, payment.PaymentDate,
		payment.Notes, payment.UserID, payment.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create seller payment: %w", err)
	}

	// Update acquisition paid amount
	updateQuery := `
		UPDATE acquisitions 
		SET paid_amount = paid_amount + $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err = s.db.Exec(updateQuery, req.Amount, req.AcquisitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update acquisition paid amount: %w", err)
	}

	// Update payment status if fully paid
	var totalCost float64
	err = s.db.Get(&totalCost, "SELECT total_cost FROM acquisitions WHERE id = $1", req.AcquisitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get acquisition total cost: %w", err)
	}

	var totalPaid float64
	err = s.db.Get(&totalPaid, "SELECT paid_amount FROM acquisitions WHERE id = $1", req.AcquisitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get acquisition paid amount: %w", err)
	}

	if totalPaid >= totalCost {
		_, err = s.db.Exec("UPDATE acquisitions SET payment_status = 'paid' WHERE id = $1", req.AcquisitionID)
		if err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}
	} else {
		_, err = s.db.Exec("UPDATE acquisitions SET payment_status = 'partial' WHERE id = $1", req.AcquisitionID)
		if err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}
	}

	return payment, nil
}

// AddRepairCost adds a repair cost to an item
func (s *Service) AddRepairCost(ctx context.Context, inventoryItemID uuid.UUID, acquisitionItemID uuid.UUID, repairType string, cost float64, description string, userID uuid.UUID) error {
	query := `
		INSERT INTO item_repair_costs (inventory_item_id, acquisition_item_id, repair_date, repair_type, cost, description, performed_by, created_at)
		VALUES ($1, $2, CURRENT_DATE, $3, $4, $5, $6, NOW())
	`

	_, err := s.db.Exec(query, inventoryItemID, acquisitionItemID, repairType, cost, description, userID)
	if err != nil {
		return fmt.Errorf("failed to add repair cost: %w", err)
	}

	// Update inventory item cost to include repair cost
	updateQuery := `
		UPDATE inventory_items 
		SET purchase_cost = purchase_cost + $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err = s.db.Exec(updateQuery, cost, inventoryItemID)
	if err != nil {
		return fmt.Errorf("failed to update inventory item cost: %w", err)
	}

	// Add to item history
	historyQuery := `
		INSERT INTO item_history (inventory_item_id, event_type, event_date, reference_type, reference_id, description, metadata, created_by, created_at)
		VALUES ($1, 'repair', NOW(), 'repair_cost', gen_random_uuid(), $2, $3, $4, NOW())
	`
	metadata := fmt.Sprintf(`{"repair_type": "%s", "cost": %.2f}`, repairType, cost)
	_, err = s.db.Exec(historyQuery, inventoryItemID, description, metadata, userID)
	if err != nil {
		return fmt.Errorf("failed to add item history: %w", err)
	}

	return nil
}

// GetItemHistory retrieves the complete history of an item
func (s *Service) GetItemHistory(ctx context.Context, inventoryItemID uuid.UUID) ([]map[string]interface{}, error) {
	query := `
		SELECT * FROM item_history 
		WHERE inventory_item_id = $1 
		ORDER BY event_date DESC
	`

	var history []map[string]interface{}
	err := s.db.Select(&history, query, inventoryItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item history: %w", err)
	}

	return history, nil
}

// GetUsedPartsAging retrieves aging information for used parts
func (s *Service) GetUsedPartsAging(ctx context.Context, alertLevel string) ([]ItemAging, error) {
	query := `SELECT * FROM used_parts_aging`
	args := []interface{}{}

	if alertLevel != "" {
		query += " WHERE alert_level = $1"
		args = append(args, alertLevel)
	}

	query += " ORDER BY days_in_stock DESC"

	var aging []ItemAging
	err := s.db.Select(&aging, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get used parts aging: %w", err)
	}

	return aging, nil
}

// GetSellerBalances retrieves balances for sellers (customers who sold items)
func (s *Service) GetSellerBalances(ctx context.Context) ([]SellerBalance, error) {
	query := `
		SELECT 
			a.customer_id,
			c.name as customer_name,
			COUNT(a.id) as total_acquisitions,
			SUM(a.total_cost) as total_acquired,
			SUM(a.paid_amount) as total_paid,
			SUM(a.total_cost - a.paid_amount) as balance,
			COUNT(DISTINCT a.id) as transaction_count,
			MAX(a.created_at) as last_transaction
		FROM acquisitions a
		JOIN customers c ON a.customer_id = c.id
		WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		GROUP BY a.customer_id, c.name
		ORDER BY balance DESC
	`

	var balances []SellerBalance
	err := s.db.Select(&balances, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get seller balances: %w", err)
	}

	return balances, nil
}