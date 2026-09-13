package acquisitions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

func nullableUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

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
	if req == nil {
		return nil, fmt.Errorf("acquisition request is required")
	}
	if req.AcquisitionDate.IsZero() {
		return nil, fmt.Errorf("acquisition_date is required")
	}
	if req.AcquisitionDate.After(time.Now().Add(24 * time.Hour)) {
		return nil, fmt.Errorf("acquisition_date cannot be in the future")
	}
	if req.Type != TypeSupplier && req.Type != TypeCustomer {
		return nil, fmt.Errorf("type must be SUPPLIER or CUSTOMER")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("at least one acquisition item is required")
	}
	if req.PaymentStatus == "" {
		req.PaymentStatus = PaymentStatusPayable
	}
	if req.PaymentStatus != PaymentStatusPaid && req.PaymentStatus != PaymentStatusPayable && req.PaymentStatus != PaymentStatusPartial && req.PaymentStatus != PaymentStatusOverdue {
		return nil, fmt.Errorf("invalid payment_status")
	}
	for _, item := range req.Items {
		if item.ProductID == uuid.Nil {
			return nil, fmt.Errorf("product_id is required for every acquisition item")
		}
		if item.UnitCost < 0 {
			return nil, fmt.Errorf("unit_cost cannot be negative")
		}
		if item.Condition != "new" && item.Condition != "used" && item.Condition != "refurbished" {
			return nil, fmt.Errorf("invalid acquisition item condition")
		}
	}
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
		PaidAmount:      0,
		PaymentStatus:   req.PaymentStatus,
		Status:          StatusDraft,
		Notes:           &req.Notes,
		UserID:          nullableUUID(userID),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	query := `
		INSERT INTO acquisitions (id, type, acquisition_date, supplier_id, customer_id, 
			total_cost, paid_amount, payment_status, status, notes, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = tx.Exec(query,
		acquisition.ID, acquisition.Type, acquisition.AcquisitionDate,
		acquisition.SupplierID, acquisition.CustomerID, acquisition.TotalCost,
		acquisition.PaidAmount, acquisition.PaymentStatus, acquisition.Status,
		acquisition.Notes, acquisition.UserID, acquisition.CreatedAt, acquisition.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create acquisition: %w", err)
	}

	// Create acquisition items
	var calculatedTotal float64
	for _, itemReq := range req.Items {
		condition := strings.ToLower(strings.TrimSpace(itemReq.Condition))
		grade := strings.ToLower(strings.TrimSpace(itemReq.Grade))
		if grade == "" {
			grade = "good"
		}
		item := &AcquisitionItem{
			ID:            uuid.New(),
			AcquisitionID: acquisitionID,
			ProductID:     itemReq.ProductID,
			SerialNumber:  itemReq.SerialNumber,
			Condition:     condition,
			Grade:         grade,
			UnitCost:      itemReq.UnitCost,
			TotalCost:     itemReq.UnitCost, // Assuming quantity 1 for individual items
			ItemStatus:    "available",
			Notes:         itemReq.Notes,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		itemQuery := `
			INSERT INTO acquisition_items (id, acquisition_id, product_id, serial_number,
				condition, grade, unit_cost, total_cost, item_status, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`

		_, err = tx.Exec(itemQuery,
			item.ID, item.AcquisitionID, item.ProductID, item.SerialNumber,
			item.Condition, item.Grade, item.UnitCost, item.TotalCost,
			item.ItemStatus, item.Notes, item.CreatedAt, item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create acquisition item: %w", err)
		}

		// Every acquired unit is immediately available for sale.
		inventoryID := uuid.New()
		itemCode := fmt.Sprintf("ACQ-%s", strings.ToUpper(strings.ReplaceAll(inventoryID.String()[:13], "-", "")))
		barcode := fmt.Sprintf("ACQ-%s", strings.ToUpper(strings.ReplaceAll(inventoryID.String(), "-", "")))
		inventoryQuery := fmt.Sprintf(`
			INSERT INTO inventory_items (id, product_id, part_type_id, item_code, barcode, serial_number,
				condition, grade, purchase_cost, selling_price, status, supplier_id,
				purchase_date, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, %s, %s)
		`, dbutil.NowSQL(s.db), dbutil.NowSQL(s.db))
		if _, err = tx.ExecContext(ctx, inventoryQuery, inventoryID, itemReq.ProductID, itemReq.PartTypeID, itemCode, barcode,
			nullableString(itemReq.SerialNumber), strings.ToUpper(condition), strings.ToUpper(grade),
			itemReq.UnitCost, itemReq.SellingPrice, "AVAILABLE", req.SupplierID, req.AcquisitionDate, nullableString(itemReq.Notes)); err != nil {
			return nil, fmt.Errorf("failed to create inventory item for acquisition: %w", err)
		}
		if _, err = tx.ExecContext(ctx, fmt.Sprintf(`UPDATE acquisition_items SET inventory_item_id = $1, item_status = 'available', updated_at = %s WHERE id = $2`, dbutil.NowSQL(s.db)), inventoryID, item.ID); err != nil {
			return nil, fmt.Errorf("failed to link acquisition item to inventory: %w", err)
		}
		updateAggregate := fmt.Sprintf(`UPDATE inventory SET quantity = quantity + 1, updated_at = %s WHERE product_id = $1`, dbutil.NowSQL(s.db))
		aggregateResult, aggregateErr := tx.ExecContext(ctx, updateAggregate, itemReq.ProductID)
		if aggregateErr != nil {
			return nil, fmt.Errorf("failed to update inventory aggregate: %w", aggregateErr)
		}
		if rows, rowsErr := aggregateResult.RowsAffected(); rowsErr != nil {
			return nil, fmt.Errorf("failed to inspect inventory aggregate: %w", rowsErr)
		} else if rows == 0 {
			if _, aggregateErr = tx.ExecContext(ctx, fmt.Sprintf(`
				INSERT INTO inventory (id, product_id, quantity, reserved_quantity, created_at, updated_at)
				VALUES ($1, $2, 1, 0, %s, %s)
			`, dbutil.NowSQL(s.db), dbutil.NowSQL(s.db)), uuid.New(), itemReq.ProductID); aggregateErr != nil {
				return nil, fmt.Errorf("failed to create inventory aggregate: %w", aggregateErr)
			}
		}
		calculatedTotal += itemReq.UnitCost
	}
	if _, err = tx.ExecContext(ctx, fmt.Sprintf(`UPDATE acquisitions SET total_cost = $1, updated_at = %s WHERE id = $2`, dbutil.NowSQL(s.db)), calculatedTotal, acquisitionID); err != nil {
		return nil, fmt.Errorf("failed to update acquisition total: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload acquisition to get calculated total cost
	return s.GetAcquisition(ctx, acquisitionID)
}

func nullableString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

// GetAcquisition retrieves an acquisition by ID
func (s *Service) GetAcquisition(ctx context.Context, id uuid.UUID) (*Acquisition, error) {
	row := map[string]any{}
	if err := s.db.QueryRowxContext(ctx, `SELECT * FROM acquisitions WHERE id = $1`, id).MapScan(row); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("acquisition not found")
		}
		return nil, fmt.Errorf("failed to get acquisition: %w", err)
	}
	acquisition := &Acquisition{ID: parseUUIDValue(row["id"]), Type: fmt.Sprint(row["type"]), TotalCost: parseFloatValue(row["total_cost"]), PaidAmount: parseFloatValue(row["paid_amount"]), PaymentStatus: fmt.Sprint(row["payment_status"]), Status: fmt.Sprint(row["status"])}
	if acquisition.ID == uuid.Nil {
		acquisition.ID = id
	}
	acquisition.SupplierID = parseNullableUUIDValue(row["supplier_id"])
	acquisition.CustomerID = parseNullableUUIDValue(row["customer_id"])
	acquisition.UserID = parseNullableUUIDValue(row["user_id"])
	acquisition.ReversedBy = parseNullableUUIDValue(row["reversed_by"])
	acquisition.Notes = parseNullableStringValue(row["notes"])
	if value, err := dbutil.ParseTimestamp(row["acquisition_date"]); err == nil {
		acquisition.AcquisitionDate = value
	} else {
		return nil, fmt.Errorf("parse acquisition date: %w", err)
	}
	if value, err := dbutil.ParseTimestamp(row["created_at"]); err == nil {
		acquisition.CreatedAt = value
	}
	if value, err := dbutil.ParseTimestamp(row["updated_at"]); err == nil {
		acquisition.UpdatedAt = value
	}
	if value, err := dbutil.ParseTimestamp(row["reversed_at"]); err == nil && !value.IsZero() {
		acquisition.ReversedAt = &value
	}
	return acquisition, nil
}

func parseUUIDValue(value any) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	if parsed, ok := value.(uuid.UUID); ok {
		return parsed
	}
	if bytes, ok := value.([]byte); ok && len(bytes) == 16 {
		var parsed uuid.UUID
		copy(parsed[:], bytes)
		return parsed
	}
	parsed, _ := uuid.Parse(strings.TrimSpace(fmt.Sprint(value)))
	return parsed
}

func parseNullableUUIDValue(value any) *uuid.UUID {
	parsed := parseUUIDValue(value)
	if parsed == uuid.Nil {
		return nil
	}
	return &parsed
}

func parseNullableStringValue(value any) *string {
	if value == nil {
		return nil
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" {
		return nil
	}
	return &text
}

func parseFloatValue(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case []byte:
		var out float64
		_, _ = fmt.Sscan(string(v), &out)
		return out
	default:
		var out float64
		_, _ = fmt.Sscan(fmt.Sprint(value), &out)
		return out
	}
}

// GetAcquisitionWithItems retrieves an acquisition with its items
func (s *Service) GetAcquisitionWithItems(ctx context.Context, id uuid.UUID) (*AcquisitionResponse, error) {
	// Get acquisition
	acquisition, err := s.GetAcquisition(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get items
	items, err := s.loadAcquisitionItems(ctx, id)
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

// loadAcquisitionItems parses timestamps and UUIDs explicitly. PostgreSQL's
// driver returns native time/UUID values, while SQLite returns text; scanning
// directly into time.Time/uuid.UUID makes the local application fail after a
// perfectly valid write.
func (s *Service) loadAcquisitionItems(ctx context.Context, acquisitionID uuid.UUID) ([]AcquisitionItem, error) {
	rows, err := s.db.QueryxContext(ctx, `SELECT id, acquisition_id, product_id, inventory_item_id, serial_number, condition, grade, unit_cost, total_cost, item_status, notes, created_at, updated_at FROM acquisition_items WHERE acquisition_id = $1 ORDER BY created_at`, acquisitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AcquisitionItem, 0)
	for rows.Next() {
		record := map[string]any{}
		if err := rows.MapScan(record); err != nil {
			return nil, err
		}
		item := AcquisitionItem{
			ID:            parseUUIDValue(record["id"]),
			AcquisitionID: parseUUIDValue(record["acquisition_id"]),
			ProductID:     parseUUIDValue(record["product_id"]),
			SerialNumber:  strings.TrimSpace(fmt.Sprint(record["serial_number"])),
			Condition:     strings.TrimSpace(fmt.Sprint(record["condition"])),
			Grade:         strings.TrimSpace(fmt.Sprint(record["grade"])),
			UnitCost:      parseFloatValue(record["unit_cost"]),
			TotalCost:     parseFloatValue(record["total_cost"]),
			ItemStatus:    strings.TrimSpace(fmt.Sprint(record["item_status"])),
			Notes:         strings.TrimSpace(fmt.Sprint(record["notes"])),
		}
		item.InventoryItemID = parseNullableUUIDValue(record["inventory_item_id"])
		if value, parseErr := dbutil.ParseTimestamp(record["created_at"]); parseErr == nil {
			item.CreatedAt = value
		}
		if value, parseErr := dbutil.ParseTimestamp(record["updated_at"]); parseErr == nil {
			item.UpdatedAt = value
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
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
	if req.StartDate != nil {
		query += fmt.Sprintf(" AND acquisition_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND acquisition_date >= $%d", argCount)
		args = append(args, *req.StartDate)
		argCount++
	}
	if req.EndDate != nil {
		query += fmt.Sprintf(" AND acquisition_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND acquisition_date <= $%d", argCount)
		args = append(args, *req.EndDate)
		argCount++
	}
	if search := strings.TrimSpace(req.Search); search != "" {
		query += fmt.Sprintf(" AND (LOWER(COALESCE(notes, '')) LIKE LOWER($%d) OR CAST(id AS TEXT) LIKE $%d)", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (LOWER(COALESCE(notes, '')) LIKE LOWER($%d) OR CAST(id AS TEXT) LIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	// Get total count
	var total int
	err := s.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count acquisitions: %w", err)
	}

	// Apply sorting
	sortColumns := map[string]string{
		"acquisition_date": "acquisition_date",
		"created_at":       "created_at",
		"total_cost":       "total_cost",
		"status":           "status",
		"payment_status":   "payment_status",
	}
	sortBy := sortColumns[req.SortBy]
	if sortBy == "" {
		sortBy = "acquisition_date"
	}
	sortOrder := "DESC"
	if strings.EqualFold(req.SortOrder, "asc") {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Apply pagination
	offset := (req.Page - 1) * req.PerPage
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.PerPage, offset)

	// Execute query
	rows, err := s.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list acquisitions: %w", err)
	}
	defer rows.Close()
	acquisitions := make([]Acquisition, 0)
	for rows.Next() {
		record := map[string]any{}
		if err := rows.MapScan(record); err != nil {
			return nil, 0, fmt.Errorf("failed to scan acquisition: %w", err)
		}
		acquisition := Acquisition{
			ID:            parseUUIDValue(record["id"]),
			Type:          fmt.Sprint(record["type"]),
			TotalCost:     parseFloatValue(record["total_cost"]),
			PaidAmount:    parseFloatValue(record["paid_amount"]),
			PaymentStatus: fmt.Sprint(record["payment_status"]),
			Status:        fmt.Sprint(record["status"]),
			Notes:         parseNullableStringValue(record["notes"]),
			SupplierID:    parseNullableUUIDValue(record["supplier_id"]),
			CustomerID:    parseNullableUUIDValue(record["customer_id"]),
			UserID:        parseNullableUUIDValue(record["user_id"]),
			ReversedBy:    parseNullableUUIDValue(record["reversed_by"]),
		}
		if value, parseErr := dbutil.ParseTimestamp(record["acquisition_date"]); parseErr == nil {
			acquisition.AcquisitionDate = value
		}
		if value, parseErr := dbutil.ParseTimestamp(record["created_at"]); parseErr == nil {
			acquisition.CreatedAt = value
		}
		if value, parseErr := dbutil.ParseTimestamp(record["updated_at"]); parseErr == nil {
			acquisition.UpdatedAt = value
		}
		if value, parseErr := dbutil.ParseTimestamp(record["reversed_at"]); parseErr == nil && !value.IsZero() {
			acquisition.ReversedAt = &value
		}
		items, itemErr := s.loadAcquisitionItems(ctx, acquisition.ID)
		if itemErr != nil {
			return nil, 0, fmt.Errorf("failed to load acquisition items: %w", itemErr)
		}
		acquisition.Items = items
		acquisitions = append(acquisitions, acquisition)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to list acquisitions: %w", err)
	}

	return acquisitions, total, nil
}

// UpdateAcquisitionStatus updates acquisition status
func (s *Service) UpdateAcquisitionStatus(ctx context.Context, id uuid.UUID, status string) error {
	status = strings.ToLower(strings.TrimSpace(status))
	updateAcquisition := func() error {
		result, err := s.db.ExecContext(ctx, fmt.Sprintf(`
			UPDATE acquisitions
			SET status = $1,
			updated_at = %s
			WHERE id = $2
		`, dbutil.NowSQL(s.db)), status, id)
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

	switch status {
	case StatusDraft, StatusAcquired, StatusCancelled, StatusReversed:
		return updateAcquisition()
	default:
		return fmt.Errorf("invalid acquisition status: %s", status)
	}
}

// LinkItemToInventory links an acquisition item to an inventory item
func (s *Service) LinkItemToInventory(ctx context.Context, acquisitionItemID uuid.UUID, inventoryItemID uuid.UUID) error {
	query := fmt.Sprintf(`
		UPDATE acquisition_items 
		SET inventory_item_id = $1, item_status = 'available', updated_at = %s
		WHERE id = $2
	`, dbutil.NowSQL(s.db))
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

	var err error
	if dbutil.IsSQLite(s.db) {
		_, err = s.db.ExecContext(ctx, `INSERT INTO seller_payments (id, acquisition_id, customer_id, amount, payment_method, payment_date, notes, user_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, payment.ID, payment.AcquisitionID, payment.CustomerID, payment.Amount, payment.PaymentMethod, payment.PaymentDate, payment.Notes, payment.UserID, payment.CreatedAt)
	} else {
		err = s.db.Get(payment, query,
			payment.ID, payment.AcquisitionID, payment.CustomerID,
			payment.Amount, payment.PaymentMethod, payment.PaymentDate,
			payment.Notes, payment.UserID, payment.CreatedAt,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create seller payment: %w", err)
	}

	// Update acquisition paid amount
	updateQuery := fmt.Sprintf(`
		UPDATE acquisitions 
		SET paid_amount = paid_amount + $1, updated_at = %s
		WHERE id = $2
	`, dbutil.NowSQL(s.db))
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
	nowSQL := dbutil.NowSQL(s.db)
	query := fmt.Sprintf(`
		INSERT INTO item_repair_costs (inventory_item_id, acquisition_item_id, repair_date, repair_type, cost, description, performed_by, created_at)
		VALUES ($1, $2, CURRENT_DATE, $3, $4, $5, $6, %s)
	`, nowSQL)

	_, err := s.db.Exec(query, inventoryItemID, acquisitionItemID, repairType, cost, description, userID)
	if err != nil {
		return fmt.Errorf("failed to add repair cost: %w", err)
	}

	// Update inventory item cost to include repair cost
	updateQuery := fmt.Sprintf(`
		UPDATE inventory_items 
		SET purchase_cost = purchase_cost + $1, updated_at = %s
		WHERE id = $2
	`, nowSQL)
	_, err = s.db.Exec(updateQuery, cost, inventoryItemID)
	if err != nil {
		return fmt.Errorf("failed to update inventory item cost: %w", err)
	}

	// Add to item history
	historyQuery := fmt.Sprintf(`
		INSERT INTO item_history (inventory_item_id, event_type, event_date, reference_type, reference_id, description, metadata, created_by, created_at)
		VALUES ($1, 'repair', %s, 'repair_cost', $5, $2, $3, $4, %s)
	`, nowSQL, nowSQL)
	metadata := fmt.Sprintf(`{"repair_type": "%s", "cost": %.2f}`, repairType, cost)
	_, err = s.db.Exec(historyQuery, inventoryItemID, description, metadata, userID, uuid.New())
	if err != nil {
		return fmt.Errorf("failed to add item history: %w", err)
	}

	return nil
}

// GetItemHistory retrieves the item snapshot and complete lifecycle history.
func (s *Service) GetItemHistory(ctx context.Context, inventoryItemID uuid.UUID) (*ItemHistoryResponse, error) {
	var item struct {
		ID           uuid.UUID  `db:"id"`
		ProductName  string     `db:"product_name"`
		SerialNumber string     `db:"serial_number"`
		Barcode      string     `db:"barcode"`
		Condition    string     `db:"condition"`
		Grade        string     `db:"grade"`
		Status       string     `db:"status"`
		Cost         float64    `db:"cost"`
		SellingPrice float64    `db:"selling_price"`
		PurchaseDate *time.Time `db:"purchase_date"`
		Location     string     `db:"location"`
		Notes        string     `db:"notes"`
	}
	err := s.db.GetContext(ctx, &item, `
		SELECT ii.id,
		       COALESCE(p.name, '') AS product_name,
		       COALESCE(ii.serial_number, '') AS serial_number,
		       COALESCE(ii.barcode, '') AS barcode,
		       COALESCE(ii.condition, '') AS condition,
		       COALESCE(ii.grade, '') AS grade,
		       COALESCE(ii.status, '') AS status,
		       COALESCE(ii.purchase_cost, 0) AS cost,
		       COALESCE(ii.selling_price, 0) AS selling_price,
		       ii.purchase_date,
		       COALESCE(l.name, '') AS location,
		       COALESCE(ii.notes, '') AS notes
		FROM inventory_items ii
		LEFT JOIN products p ON p.id = ii.product_id
		LEFT JOIN locations l ON l.id = ii.location_id
		WHERE ii.id = $1
	`, inventoryItemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("inventory item not found")
		}
		return nil, fmt.Errorf("failed to get inventory item: %w", err)
	}

	query := `
		SELECT id, event_type, event_date, description, metadata AS details
		FROM item_history
		WHERE inventory_item_id = $1
		ORDER BY event_date DESC
	`
	var rows []struct {
		ID          uuid.UUID `db:"id"`
		EventType   string    `db:"event_type"`
		EventDate   time.Time `db:"event_date"`
		Description *string   `db:"description"`
		Details     []byte    `db:"details"`
	}
	err = s.db.Select(&rows, query, inventoryItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item history: %w", err)
	}
	events := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		details := map[string]interface{}{}
		if len(row.Details) > 0 {
			if err := json.Unmarshal(row.Details, &details); err != nil {
				return nil, fmt.Errorf("failed to decode item history metadata: %w", err)
			}
		}
		events = append(events, map[string]interface{}{
			"id":          row.ID,
			"event_type":  row.EventType,
			"event_date":  row.EventDate,
			"description": row.Description,
			"details":     details,
		})
	}

	var repairs []struct {
		Description string    `db:"description"`
		Date        time.Time `db:"date"`
		Amount      float64   `db:"amount"`
	}
	err = s.db.Select(&repairs, `
		SELECT COALESCE(description, '') AS description,
		       repair_date AS date,
		       COALESCE(cost, 0) AS amount
		FROM item_repair_costs
		WHERE inventory_item_id = $1
		ORDER BY repair_date DESC, created_at DESC
	`, inventoryItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item repair costs: %w", err)
	}

	repairCosts := make([]map[string]interface{}, 0, len(repairs))
	var totalRepairCosts float64
	for _, repair := range repairs {
		totalRepairCosts += repair.Amount
		repairCosts = append(repairCosts, map[string]interface{}{
			"description": repair.Description,
			"date":        repair.Date,
			"amount":      repair.Amount,
		})
	}

	return &ItemHistoryResponse{
		Item: map[string]interface{}{
			"id":            item.ID,
			"product_name":  item.ProductName,
			"serial_number": item.SerialNumber,
			"barcode":       item.Barcode,
			"condition":     item.Condition,
			"grade":         item.Grade,
			"status":        item.Status,
			"cost":          item.Cost,
			"selling_price": item.SellingPrice,
			"purchase_date": item.PurchaseDate,
			"location":      item.Location,
			"notes":         item.Notes,
		},
		Events:           events,
		RepairCosts:      repairCosts,
		TotalRepairCosts: totalRepairCosts,
	}, nil
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
	// SQLite stores timestamps as TEXT. Scanning the view directly into
	// time.Time therefore fails even though the rows are valid. Use the same
	// compatibility scanner used by seller balances and map the rows explicitly.
	if strings.EqualFold(s.db.DriverName(), "sqlite") {
		type sqliteAgingRow struct {
			ItemID          uuid.UUID       `db:"item_id"`
			AcquisitionID   uuid.UUID       `db:"acquisition_id"`
			AcquisitionDate nullableSQLTime `db:"acquisition_date"`
			DaysInStock     int             `db:"days_in_stock"`
			Status          string          `db:"status"`
			Condition       string          `db:"condition"`
			Cost            float64         `db:"cost"`
			CurrentPrice    float64         `db:"current_price"`
			AgingCategory   string          `db:"aging_category"`
			AlertLevel      string          `db:"alert_level"`
		}
		var rows []sqliteAgingRow
		if err := s.db.SelectContext(ctx, &rows, query, args...); err != nil {
			return nil, fmt.Errorf("failed to get used parts aging: %w", err)
		}
		aging := make([]ItemAging, 0, len(rows))
		for _, row := range rows {
			item := ItemAging{
				ItemID: row.ItemID, AcquisitionID: row.AcquisitionID,
				DaysInStock: row.DaysInStock, Status: row.Status,
				Condition: row.Condition, Cost: row.Cost,
				CurrentPrice: row.CurrentPrice, AgingCategory: row.AgingCategory,
				AlertLevel: row.AlertLevel,
			}
			if row.AcquisitionDate.Valid {
				item.AcquisitionDate = row.AcquisitionDate.Time
			}
			aging = append(aging, item)
		}
		return aging, nil
	}

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
			c.code as customer_code,
			COALESCE(c.phone, '') as customer_phone,
			c.email as customer_email,
			COUNT(a.id) as total_acquisitions,
			SUM(a.total_cost) as total_acquired,
			SUM(a.paid_amount) as total_paid,
			SUM(a.total_cost - a.paid_amount) as balance,
			COUNT(DISTINCT a.id) as transaction_count,
			MAX(a.created_at) as last_transaction
		FROM acquisitions a
		JOIN customers c ON a.customer_id = c.id
		WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		GROUP BY a.customer_id, c.name, c.code, c.phone, c.email
		ORDER BY balance DESC
	`

	// SQLite stores legacy timestamps as TEXT while PostgreSQL returns a
	// time.Time. Scan through a small compatibility type so the same endpoint
	// works in both modes (and with NULL for sellers without a transaction).
	type sellerBalanceRow struct {
		CustomerID        uuid.UUID       `db:"customer_id"`
		CustomerName      string          `db:"customer_name"`
		CustomerCode      string          `db:"customer_code"`
		CustomerPhone     string          `db:"customer_phone"`
		CustomerEmail     *string         `db:"customer_email"`
		TotalAcquisitions int             `db:"total_acquisitions"`
		TotalAcquired     float64         `db:"total_acquired"`
		TotalPaid         float64         `db:"total_paid"`
		Balance           float64         `db:"balance"`
		TransactionCount  int             `db:"transaction_count"`
		LastTransaction   nullableSQLTime `db:"last_transaction"`
	}
	var rows []sellerBalanceRow
	err := s.db.SelectContext(ctx, &rows, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get seller balances: %w", err)
	}

	balances := make([]SellerBalance, 0, len(rows))
	for _, row := range rows {
		balance := SellerBalance{
			CustomerID:        row.CustomerID,
			CustomerName:      row.CustomerName,
			CustomerCode:      row.CustomerCode,
			CustomerPhone:     row.CustomerPhone,
			CustomerEmail:     row.CustomerEmail,
			TotalAcquisitions: row.TotalAcquisitions,
			TotalAcquired:     row.TotalAcquired,
			TotalPaid:         row.TotalPaid,
			Balance:           row.Balance,
			TransactionCount:  row.TransactionCount,
		}
		if row.LastTransaction.Valid {
			lastTransaction := row.LastTransaction.Time
			balance.LastTransaction = &lastTransaction
		}
		balances = append(balances, balance)
	}

	return balances, nil
}

// CreateSellerBalancePayment applies a payment to the oldest unpaid acquisition for a seller.
func (s *Service) CreateSellerBalancePayment(ctx context.Context, customerID uuid.UUID, amount float64, userID uuid.UUID) (*SellerPayment, error) {
	var acquisitionID uuid.UUID
	err := s.db.GetContext(ctx, &acquisitionID, `
		SELECT id
		FROM acquisitions
		WHERE customer_id = $1
		  AND type = 'CUSTOMER'
		  AND status NOT IN ('cancelled', 'reversed')
		  AND total_cost > paid_amount
		ORDER BY acquisition_date ASC, created_at ASC
		LIMIT 1
	`, customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no unpaid acquisition found for seller")
		}
		return nil, fmt.Errorf("failed to find unpaid acquisition: %w", err)
	}

	return s.CreateSellerPayment(ctx, &SellerPaymentRequest{
		AcquisitionID: acquisitionID,
		CustomerID:    customerID,
		Amount:        amount,
		PaymentMethod: "cash",
		PaymentDate:   time.Now(),
	}, userID)
}

// nullableSQLTime accepts both database driver representations used by the
// application: time.Time from PostgreSQL and text from SQLite.
type nullableSQLTime struct {
	time.Time
	Valid bool
}

func (t *nullableSQLTime) Scan(value interface{}) error {
	if value == nil {
		t.Valid = false
		t.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.Time, t.Valid = v, true
		return nil
	case []byte:
		return t.parse(string(v))
	case string:
		return t.parse(v)
	default:
		return fmt.Errorf("unsupported timestamp type %T", value)
	}
}

func (t *nullableSQLTime) parse(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		t.Valid = false
		t.Time = time.Time{}
		return nil
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			t.Time, t.Valid = parsed, true
			return nil
		}
	}
	return fmt.Errorf("unsupported timestamp value %q", raw)
}
