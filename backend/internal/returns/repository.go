package returns

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Repository handles return data operations
type Repository struct {
	db *sqlx.DB
}

type returnTimestamp struct{ time.Time }

func (t *returnTimestamp) Scan(value any) error {
	parsed, err := dbutil.ParseTimestamp(value)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// localReturnRow keeps SQLite's TEXT UUID/timestamp representation out of the
// public model. SQLite does not have native UUID/timestamp types, so scanning
// directly into uuid.UUID/time.Time is driver-dependent and fails at runtime.
type localReturnRow struct {
	ID                       sql.NullString `db:"id"`
	ReturnNumber             sql.NullString `db:"return_number"`
	ReferenceNumber          sql.NullString `db:"reference_number"`
	SaleID                   sql.NullString `db:"sale_id"`
	PurchaseID               sql.NullString `db:"purchase_id"`
	CustomerID               sql.NullString `db:"customer_id"`
	CustomerName             sql.NullString `db:"customer_name"`
	ReturnDate               sql.NullString `db:"return_date"`
	ReturnType               sql.NullString `db:"return_type"`
	Status                   sql.NullString `db:"status"`
	TotalRefundAmount        float64        `db:"total_refund_amount"`
	RefundMethod             sql.NullString `db:"refund_method"`
	RefundDate               sql.NullString `db:"refund_date"`
	RefundReference          sql.NullString `db:"refund_reference"`
	DebtID                   sql.NullString `db:"debt_id"`
	DebtAdjustment           float64        `db:"debt_adjustment"`
	CustomerCredit           float64        `db:"customer_credit"`
	Reason                   sql.NullString `db:"reason"`
	ReasonDetail             sql.NullString `db:"reason_detail"`
	ItemConditionAfterReturn sql.NullString `db:"item_condition_after_return"`
	IsWarrantyClaim          int            `db:"is_warranty_claim"`
	WarrantyID               sql.NullString `db:"warranty_id"`
	WarrantyValidUntil       sql.NullString `db:"warranty_valid_until"`
	CreatedBy                sql.NullString `db:"created_by"`
	ProcessedBy              sql.NullString `db:"processed_by"`
	ApprovedBy               sql.NullString `db:"approved_by"`
	ApprovedAt               sql.NullString `db:"approved_at"`
	Notes                    sql.NullString `db:"notes"`
	InternalNotes            sql.NullString `db:"internal_notes"`
	CreatedAt                sql.NullString `db:"created_at"`
	UpdatedAt                sql.NullString `db:"updated_at"`
}

func localUUID(value sql.NullString) uuid.UUID {
	if !value.Valid || value.String == "" {
		return uuid.Nil
	}
	id, _ := uuid.Parse(value.String)
	return id
}
func localUUIDPtr(value sql.NullString) *uuid.UUID {
	id := localUUID(value)
	if id == uuid.Nil {
		return nil
	}
	return &id
}
func localTime(value sql.NullString) time.Time {
	if !value.Valid || value.String == "" {
		return time.Time{}
	}
	parsed, _ := dbutil.ParseTimestamp(value.String)
	return parsed
}
func localTimePtr(value sql.NullString) *time.Time {
	t := localTime(value)
	if t.IsZero() {
		return nil
	}
	return &t
}

func (row localReturnRow) model() Return {
	return Return{ID: localUUID(row.ID), ReturnNumber: row.ReturnNumber.String, ReferenceNumber: row.ReferenceNumber.String,
		SaleID: localUUID(row.SaleID), PurchaseID: localUUID(row.PurchaseID), CustomerID: localUUID(row.CustomerID), CustomerName: row.CustomerName.String,
		ReturnDate: localTime(row.ReturnDate), ReturnType: row.ReturnType.String, Status: row.Status.String,
		TotalRefundAmount: row.TotalRefundAmount, RefundMethod: row.RefundMethod.String, RefundDate: localTimePtr(row.RefundDate), RefundReference: row.RefundReference.String,
		DebtID: localUUIDPtr(row.DebtID), DebtAdjustment: row.DebtAdjustment, CustomerCredit: row.CustomerCredit,
		Reason: row.Reason.String, ReasonDetail: row.ReasonDetail.String, ItemConditionAfterReturn: row.ItemConditionAfterReturn.String,
		IsWarrantyClaim: row.IsWarrantyClaim != 0, WarrantyID: localUUIDPtr(row.WarrantyID), WarrantyValidUntil: localTimePtr(row.WarrantyValidUntil),
		CreatedBy: localUUIDPtr(row.CreatedBy), ProcessedBy: localUUIDPtr(row.ProcessedBy), ApprovedBy: localUUIDPtr(row.ApprovedBy), ApprovedAt: localTimePtr(row.ApprovedAt),
		Notes: row.Notes.String, InternalNotes: row.InternalNotes.String, CreatedAt: localTime(row.CreatedAt), UpdatedAt: localTime(row.UpdatedAt)}
}

const localReturnColumns = `id, return_number, COALESCE(reference_number,'') AS reference_number, COALESCE(sale_id,'') AS sale_id, COALESCE(purchase_id,'') AS purchase_id, COALESCE(customer_id, (SELECT customer_id FROM sales WHERE sales.id = returns.sale_id),'') AS customer_id, COALESCE((SELECT name FROM customers WHERE customers.id = COALESCE(returns.customer_id, (SELECT customer_id FROM sales WHERE sales.id = returns.sale_id))),'') AS customer_name, return_date, return_type, status, total_refund_amount, COALESCE(refund_method,'') AS refund_method, refund_date, COALESCE(refund_reference,'') AS refund_reference, debt_id, debt_adjustment, customer_credit, COALESCE(reason,'') AS reason, COALESCE(reason_detail,'') AS reason_detail, COALESCE(item_condition_after_return,'') AS item_condition_after_return, is_warranty_claim, warranty_id, warranty_valid_until, created_by, processed_by, approved_by, approved_at, COALESCE(notes,'') AS notes, COALESCE(internal_notes,'') AS internal_notes, created_at, updated_at`

type localReturnItemRow struct {
	ID                 sql.NullString  `db:"id"`
	ReturnID           sql.NullString  `db:"return_id"`
	SaleItemID         sql.NullString  `db:"sale_item_id"`
	ProductID          sql.NullString  `db:"product_id"`
	InventoryItemID    sql.NullString  `db:"inventory_item_id"`
	SerialNumber       sql.NullString  `db:"serial_number"`
	Barcode            sql.NullString  `db:"barcode"`
	QuantityReturned   int             `db:"quantity_returned"`
	OriginalQuantity   sql.NullInt64   `db:"original_quantity"`
	UnitPrice          float64         `db:"unit_price"`
	TotalRefundAmount  float64         `db:"total_refund_amount"`
	OriginalCondition  sql.NullString  `db:"original_condition"`
	ReturnedCondition  sql.NullString  `db:"returned_condition"`
	ConditionNotes     sql.NullString  `db:"condition_notes"`
	Resolution         sql.NullString  `db:"resolution"`
	InventoryStatus    sql.NullString  `db:"inventory_status"`
	InspectionRequired int             `db:"inspection_required"`
	InspectionDate     sql.NullString  `db:"inspection_date"`
	InspectionResult   sql.NullString  `db:"inspection_result"`
	InspectionNotes    sql.NullString  `db:"inspection_notes"`
	OriginalCost       sql.NullFloat64 `db:"original_cost"`
	RepairCost         float64         `db:"repair_cost"`
	CreatedAt          sql.NullString  `db:"created_at"`
	UpdatedAt          sql.NullString  `db:"updated_at"`
}

func (row localReturnItemRow) model() ReturnItem {
	item := ReturnItem{ID: localUUID(row.ID), ReturnID: localUUID(row.ReturnID), SaleItemID: localUUIDPtr(row.SaleItemID), ProductID: localUUIDPtr(row.ProductID), InventoryItemID: localUUIDPtr(row.InventoryItemID), SerialNumber: row.SerialNumber.String, Barcode: row.Barcode.String, QuantityReturned: row.QuantityReturned, UnitPrice: row.UnitPrice, TotalRefundAmount: row.TotalRefundAmount, OriginalCondition: row.OriginalCondition.String, ReturnedCondition: row.ReturnedCondition.String, ConditionNotes: row.ConditionNotes.String, Resolution: row.Resolution.String, InventoryStatus: row.InventoryStatus.String, InspectionRequired: row.InspectionRequired != 0, InspectionDate: localTimePtr(row.InspectionDate), InspectionResult: row.InspectionResult.String, InspectionNotes: row.InspectionNotes.String, RepairCost: row.RepairCost, CreatedAt: localTime(row.CreatedAt), UpdatedAt: localTime(row.UpdatedAt)}
	if row.OriginalQuantity.Valid {
		v := int(row.OriginalQuantity.Int64)
		item.OriginalQuantity = &v
	}
	if row.OriginalCost.Valid {
		v := row.OriginalCost.Float64
		item.OriginalCost = &v
	}
	return item
}

// NewRepository creates a new return repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateReturn creates a new return
func (r *Repository) CreateReturn(ctx context.Context, returnRecord *Return) error {
	if dbutil.IsSQLite(r.db) {
		if returnRecord.ID == uuid.Nil {
			returnRecord.ID = uuid.New()
		}
		if returnRecord.CreatedAt.IsZero() {
			returnRecord.CreatedAt = time.Now().UTC()
		}
		if returnRecord.UpdatedAt.IsZero() {
			returnRecord.UpdatedAt = returnRecord.CreatedAt
		}
		idArg := func(id uuid.UUID) interface{} {
			if id == uuid.Nil {
				return nil
			}
			return id.String()
		}
		var createdBy interface{}
		if returnRecord.CreatedBy != nil {
			createdBy = returnRecord.CreatedBy.String()
		}
		_, err := r.db.ExecContext(ctx, `INSERT INTO returns (id,return_number,reference_number,sale_id,purchase_id,customer_id,return_date,return_type,status,total_refund_amount,refund_method,refund_date,refund_reference,debt_id,debt_adjustment,customer_credit,reason,reason_detail,item_condition_after_return,is_warranty_claim,warranty_id,warranty_valid_until,created_by,processed_by,approved_by,approved_at,notes,internal_notes,created_at,updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, returnRecord.ID.String(), returnRecord.ReturnNumber, returnRecord.ReferenceNumber, idArg(returnRecord.SaleID), idArg(returnRecord.PurchaseID), idArg(returnRecord.CustomerID), returnRecord.ReturnDate.Format(time.RFC3339Nano), returnRecord.ReturnType, returnRecord.Status, returnRecord.TotalRefundAmount, returnRecord.RefundMethod, returnRecord.RefundDate, returnRecord.RefundReference, idArgPtr(returnRecord.DebtID), returnRecord.DebtAdjustment, returnRecord.CustomerCredit, returnRecord.Reason, returnRecord.ReasonDetail, returnRecord.ItemConditionAfterReturn, returnRecord.IsWarrantyClaim, idArgPtr(returnRecord.WarrantyID), returnRecord.WarrantyValidUntil, createdBy, idArgPtr(returnRecord.ProcessedBy), idArgPtr(returnRecord.ApprovedBy), returnRecord.ApprovedAt, returnRecord.Notes, returnRecord.InternalNotes, returnRecord.CreatedAt.Format(time.RFC3339Nano), returnRecord.UpdatedAt.Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("failed to create return: %w", err)
		}
		return nil
	}
	var purchaseID interface{} = returnRecord.PurchaseID
	if returnRecord.PurchaseID == uuid.Nil {
		purchaseID = nil
	}
	var saleID interface{} = returnRecord.SaleID
	if returnRecord.SaleID == uuid.Nil {
		saleID = nil
	}
	var customerID interface{} = returnRecord.CustomerID
	if returnRecord.CustomerID == uuid.Nil {
		customerID = nil
	}
	createdBy := returnRecord.CreatedBy
	if createdBy != nil {
		var exists bool
		if err := r.db.GetContext(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, *createdBy); err != nil || !exists {
			createdBy = nil
		}
	}
	query := `
		INSERT INTO returns (return_number, reference_number, sale_id, purchase_id, customer_id, 
			return_date, return_type, status, total_refund_amount, refund_method, refund_date,
			refund_reference,
			debt_id, debt_adjustment, customer_credit, 			reason, reason_detail, item_condition_after_return,
			is_warranty_claim, warranty_id, warranty_valid_until, created_by, processed_by, approved_by, approved_at,
			notes, internal_notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		returnRecord.ReturnNumber, returnRecord.ReferenceNumber, saleID, purchaseID, customerID,
		returnRecord.ReturnDate, returnRecord.ReturnType, returnRecord.Status, returnRecord.TotalRefundAmount, returnRecord.RefundMethod,
		returnRecord.RefundDate, returnRecord.RefundReference, returnRecord.DebtID, returnRecord.DebtAdjustment, returnRecord.CustomerCredit,
		returnRecord.Reason, returnRecord.ReasonDetail, returnRecord.ItemConditionAfterReturn, returnRecord.IsWarrantyClaim,
		returnRecord.WarrantyID, returnRecord.WarrantyValidUntil, createdBy, returnRecord.ProcessedBy,
		returnRecord.ApprovedBy, returnRecord.ApprovedAt, returnRecord.Notes, returnRecord.InternalNotes,
		returnRecord.CreatedAt, returnRecord.UpdatedAt,
	).Scan(&returnRecord.ID, &returnRecord.CreatedAt, &returnRecord.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create return: %w", err)
	}
	return nil
}

func idArgPtr(id *uuid.UUID) interface{} {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return id.String()
}

// GetReturnByID retrieves a return by ID
func (r *Repository) GetReturnByID(ctx context.Context, id uuid.UUID) (*Return, error) {
	if dbutil.IsSQLite(r.db) {
		var row localReturnRow
		if err := r.db.GetContext(ctx, &row, `SELECT `+localReturnColumns+` FROM returns WHERE id = ?`, id.String()); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrReturnNotFound
			}
			return nil, fmt.Errorf("failed to get return: %w", err)
		}
		model := row.model()
		return &model, nil
	}
	var returnRecord Return
	query := `
		SELECT r.id, r.return_number, r.reference_number, r.sale_id, r.purchase_id, COALESCE(r.customer_id, s.customer_id) AS customer_id, COALESCE(c.name, '') AS customer_name,
			return_date, return_type, status, total_refund_amount, COALESCE(refund_method, '') AS refund_method, refund_date, COALESCE(refund_reference, '') AS refund_reference,
			debt_id, debt_adjustment, customer_credit, COALESCE(reason, '') AS reason, COALESCE(reason_detail, '') AS reason_detail, COALESCE(item_condition_after_return, '') AS item_condition_after_return,
			is_warranty_claim, warranty_id, warranty_valid_until, created_by, processed_by, approved_by, approved_at,
			COALESCE(notes, '') AS notes, COALESCE(internal_notes, '') AS internal_notes, created_at, updated_at
		FROM returns r
		LEFT JOIN sales s ON s.id = r.sale_id
		LEFT JOIN customers c ON c.id = COALESCE(r.customer_id, s.customer_id)
		WHERE r.id = $1
	`

	err := r.db.GetContext(ctx, &returnRecord, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReturnNotFound
		}
		return nil, fmt.Errorf("failed to get return: %w", err)
	}
	return &returnRecord, nil
}

// ListReturns retrieves returns with pagination and filters
func (r *Repository) ListReturns(ctx context.Context, req ReturnListRequest) ([]Return, int, error) {
	if dbutil.IsSQLite(r.db) {
		if req.Page < 1 {
			req.Page = 1
		}
		if req.PerPage < 1 {
			req.PerPage = 20
		}
		where := " WHERE 1=1"
		args := make([]interface{}, 0)
		if req.CustomerID != nil {
			where += " AND customer_id = ?"
			args = append(args, req.CustomerID.String())
		}
		if req.SaleID != nil {
			where += " AND sale_id = ?"
			args = append(args, req.SaleID.String())
		}
		if req.Status != "" {
			where += " AND status = ?"
			args = append(args, req.Status)
		}
		if req.ReturnType != "" {
			where += " AND return_type = ?"
			args = append(args, req.ReturnType)
		}
		if req.RefundMethod != "" {
			where += " AND refund_method = ?"
			args = append(args, req.RefundMethod)
		}
		if req.StartDate != nil {
			where += " AND return_date >= ?"
			args = append(args, req.StartDate.Format(time.RFC3339Nano))
		}
		if req.EndDate != nil {
			where += " AND return_date <= ?"
			args = append(args, req.EndDate.Format(time.RFC3339Nano))
		}
		if req.Search != "" {
			pattern := "%" + req.Search + "%"
			where += " AND (return_number LIKE ? OR reference_number LIKE ? OR reason LIKE ? OR notes LIKE ?)"
			args = append(args, pattern, pattern, pattern, pattern)
		}
		var count int
		if err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM returns"+where, args...); err != nil {
			return nil, 0, fmt.Errorf("failed to count returns: %w", err)
		}
		sortBy := "return_date"
		for _, allowed := range []string{"return_date", "created_at", "status", "total_refund_amount", "return_number"} {
			if req.SortBy == allowed {
				sortBy = allowed
			}
		}
		sortOrder := "DESC"
		if strings.EqualFold(req.SortOrder, "ASC") {
			sortOrder = "ASC"
		}
		query := "SELECT " + localReturnColumns + " FROM returns" + where + " ORDER BY " + sortBy + " " + sortOrder + " LIMIT ? OFFSET ?"
		args = append(args, req.PerPage, (req.Page-1)*req.PerPage)
		var rows []localReturnRow
		if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
			return nil, 0, fmt.Errorf("failed to list returns: %w", err)
		}
		result := make([]Return, 0, len(rows))
		for _, row := range rows {
			result = append(result, row.model())
		}
		return result, count, nil
	}
	var returns []Return
	var count int

	// Build base query
	baseQuery := `
		SELECT r.id, r.return_number, r.reference_number,
			COALESCE(r.sale_id, '00000000-0000-0000-0000-000000000000'::uuid) AS sale_id,
			COALESCE(r.purchase_id, '00000000-0000-0000-0000-000000000000'::uuid) AS purchase_id,
			COALESCE(r.customer_id, s.customer_id, '00000000-0000-0000-0000-000000000000'::uuid) AS customer_id, COALESCE(c.name, '') AS customer_name,
			r.return_date, r.return_type, r.status, r.total_refund_amount, COALESCE(r.refund_method, '') AS refund_method, r.refund_date,
			COALESCE(r.refund_reference, '') AS refund_reference,
			r.debt_id, r.debt_adjustment, r.customer_credit, COALESCE(r.reason, '') AS reason,
			COALESCE(r.reason_detail, '') AS reason_detail, COALESCE(r.item_condition_after_return, '') AS item_condition_after_return,
			r.is_warranty_claim, r.warranty_id, r.warranty_valid_until, r.created_by, r.processed_by, r.approved_by, r.approved_at,
			COALESCE(r.notes, '') AS notes, COALESCE(r.internal_notes, '') AS internal_notes, r.created_at, r.updated_at
		FROM returns r
		LEFT JOIN sales s ON s.id = r.sale_id
		LEFT JOIN customers c ON c.id = COALESCE(r.customer_id, s.customer_id)
		WHERE 1=1
	`

	countQuery := `
		SELECT COUNT(*) FROM returns WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if req.CustomerID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND customer_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND customer_id = $%d", argCount)
		args = append(args, *req.CustomerID)
	}

	if req.SaleID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND sale_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND sale_id = $%d", argCount)
		args = append(args, *req.SaleID)
	}

	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}

	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND return_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND return_date >= $%d", argCount)
		args = append(args, *req.StartDate)
	}

	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND return_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND return_date <= $%d", argCount)
		args = append(args, *req.EndDate)
	}

	if req.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (r.return_number ILIKE $%d OR r.reference_number ILIKE $%d OR r.reason ILIKE $%d OR r.notes ILIKE $%d)", argCount, argCount, argCount, argCount)
		countQuery += fmt.Sprintf(" AND (return_number ILIKE $%d OR reference_number ILIKE $%d OR reason ILIKE $%d OR notes ILIKE $%d)", argCount, argCount, argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if req.ReturnType != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND return_type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND return_type = $%d", argCount)
		args = append(args, req.ReturnType)
	}

	if req.RefundMethod != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND refund_method = $%d", argCount)
		countQuery += fmt.Sprintf(" AND refund_method = $%d", argCount)
		args = append(args, req.RefundMethod)
	}

	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count returns: %w", err)
	}

	// Add sorting
	sortBy := "return_date"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "DESC"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}
	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	offset := (req.Page - 1) * req.PerPage
	argCount++
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.PerPage, offset)

	err = r.db.SelectContext(ctx, &returns, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list returns: %w", err)
	}

	return returns, count, nil
}

// UpdateReturn updates a return
func (r *Repository) UpdateReturn(ctx context.Context, returnRecord *Return) error {
	if dbutil.IsSQLite(r.db) {
		now := time.Now().UTC()
		result, err := r.db.ExecContext(ctx, `UPDATE returns SET return_date=?,return_type=?,status=?,total_refund_amount=?,refund_method=?,refund_date=?,refund_reference=?,debt_id=?,debt_adjustment=?,customer_credit=?,reason=?,reason_detail=?,item_condition_after_return=?,is_warranty_claim=?,warranty_id=?,warranty_valid_until=?,processed_by=?,approved_by=?,approved_at=?,notes=?,internal_notes=?,updated_at=? WHERE id=?`, returnRecord.ReturnDate.Format(time.RFC3339Nano), returnRecord.ReturnType, returnRecord.Status, returnRecord.TotalRefundAmount, returnRecord.RefundMethod, returnRecord.RefundDate, returnRecord.RefundReference, idArgPtr(returnRecord.DebtID), returnRecord.DebtAdjustment, returnRecord.CustomerCredit, returnRecord.Reason, returnRecord.ReasonDetail, returnRecord.ItemConditionAfterReturn, returnRecord.IsWarrantyClaim, idArgPtr(returnRecord.WarrantyID), returnRecord.WarrantyValidUntil, idArgPtr(returnRecord.ProcessedBy), idArgPtr(returnRecord.ApprovedBy), returnRecord.ApprovedAt, returnRecord.Notes, returnRecord.InternalNotes, now.Format(time.RFC3339Nano), returnRecord.ID.String())
		if err != nil {
			return fmt.Errorf("failed to update return: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return ErrReturnNotFound
		}
		returnRecord.UpdatedAt = now
		return nil
	}
	query := `
		UPDATE returns
		SET return_date = $2, return_type = $3, status = $4, total_refund_amount = $5, 
			refund_method = $6, refund_date = $7, refund_reference = $8, debt_id = $9, debt_adjustment = $10,
			customer_credit = $11, reason = $12, reason_detail = $13, item_condition_after_return = $14,
			is_warranty_claim = $15, warranty_id = $16, warranty_valid_until = $17, processed_by = $18,
			approved_by = $19, approved_at = $20, notes = $21, internal_notes = $22, updated_at = $23
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		returnRecord.ID, returnRecord.ReturnDate, returnRecord.ReturnType, returnRecord.Status, returnRecord.TotalRefundAmount,
		returnRecord.RefundMethod, returnRecord.RefundDate, returnRecord.RefundReference, returnRecord.DebtID, returnRecord.DebtAdjustment,
		returnRecord.CustomerCredit, returnRecord.Reason, returnRecord.ReasonDetail, returnRecord.ItemConditionAfterReturn,
		returnRecord.IsWarrantyClaim, returnRecord.WarrantyID, returnRecord.WarrantyValidUntil, returnRecord.ProcessedBy,
		returnRecord.ApprovedBy, returnRecord.ApprovedAt, returnRecord.Notes, returnRecord.InternalNotes, time.Now(),
	).Scan(&returnRecord.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReturnNotFound
		}
		return fmt.Errorf("failed to update return: %w", err)
	}
	return nil
}

// DeleteReturn deletes a return
func (r *Repository) DeleteReturn(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM returns WHERE id = $1`
	arg := interface{}(id)
	if dbutil.IsSQLite(r.db) {
		query = `DELETE FROM returns WHERE id = ?`
		arg = id.String()
	}

	result, err := r.db.ExecContext(ctx, query, arg)
	if err != nil {
		return fmt.Errorf("failed to delete return: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrReturnNotFound
	}

	return nil
}

// CreateReturnItem creates a new return item
func (r *Repository) CreateReturnItem(ctx context.Context, item *ReturnItem) error {
	if dbutil.IsSQLite(r.db) {
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		now := time.Now().UTC()
		if item.CreatedAt.IsZero() {
			item.CreatedAt = now
		}
		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = item.CreatedAt
		}
		_, err := r.db.ExecContext(ctx, `INSERT INTO return_items (id,return_id,sale_item_id,product_id,inventory_item_id,serial_number,barcode,quantity_returned,original_quantity,unit_price,total_refund_amount,original_condition,returned_condition,condition_notes,resolution,inventory_status,inspection_required,inspection_date,inspection_result,inspection_notes,original_cost,repair_cost,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, item.ID.String(), item.ReturnID.String(), idArgPtr(item.SaleItemID), idArgPtr(item.ProductID), idArgPtr(item.InventoryItemID), item.SerialNumber, item.Barcode, item.QuantityReturned, item.OriginalQuantity, item.UnitPrice, item.TotalRefundAmount, item.OriginalCondition, item.ReturnedCondition, item.ConditionNotes, item.Resolution, item.InventoryStatus, item.InspectionRequired, item.InspectionDate, item.InspectionResult, item.InspectionNotes, item.OriginalCost, item.RepairCost, item.CreatedAt.Format(time.RFC3339Nano), item.UpdatedAt.Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("failed to create return item: %w", err)
		}
		return nil
	}
	var inventoryStatus interface{} = item.InventoryStatus
	if item.InventoryStatus == "" {
		inventoryStatus = nil
	}
	var inspectionResult interface{} = item.InspectionResult
	if item.InspectionResult == "" {
		inspectionResult = nil
	}
	query := `
		INSERT INTO return_items (return_id, sale_item_id, product_id, inventory_item_id, serial_number, barcode,
			quantity_returned, original_quantity, unit_price, total_refund_amount,
			original_condition, returned_condition, condition_notes, resolution, inventory_status,
			inspection_required, inspection_date, inspection_result, inspection_notes,
			original_cost, repair_cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		item.ReturnID, item.SaleItemID, item.ProductID, item.InventoryItemID, item.SerialNumber, item.Barcode,
		item.QuantityReturned, item.OriginalQuantity, item.UnitPrice, item.TotalRefundAmount,
		item.OriginalCondition, item.ReturnedCondition, item.ConditionNotes, item.Resolution, inventoryStatus,
		item.InspectionRequired, item.InspectionDate, inspectionResult, item.InspectionNotes,
		item.OriginalCost, item.RepairCost, time.Now(), time.Now(),
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create return item: %w", err)
	}
	return nil
}

// GetReturnItems retrieves items for a return
func (r *Repository) GetReturnItems(ctx context.Context, returnID uuid.UUID) ([]ReturnItem, error) {
	if dbutil.IsSQLite(r.db) {
		var rows []localReturnItemRow
		if err := r.db.SelectContext(ctx, &rows, `SELECT id,return_id,sale_item_id,product_id,inventory_item_id,serial_number,barcode,quantity_returned,original_quantity,unit_price,total_refund_amount,original_condition,returned_condition,condition_notes,resolution,inventory_status,inspection_required,inspection_date,inspection_result,inspection_notes,original_cost,repair_cost,created_at,updated_at FROM return_items WHERE return_id = ? ORDER BY created_at`, returnID.String()); err != nil {
			return nil, fmt.Errorf("failed to get return items: %w", err)
		}
		items := make([]ReturnItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, row.model())
		}
		return items, nil
	}
	var items []ReturnItem
	query := `
		SELECT id, return_id, sale_item_id, product_id, inventory_item_id, serial_number, barcode,
			quantity_returned, original_quantity, unit_price, total_refund_amount,
			COALESCE(original_condition, '') AS original_condition, COALESCE(returned_condition, '') AS returned_condition,
			COALESCE(condition_notes, '') AS condition_notes, COALESCE(resolution, '') AS resolution,
			COALESCE(inventory_status, '') AS inventory_status,
			inspection_required, inspection_date, COALESCE(inspection_result, '') AS inspection_result,
			COALESCE(inspection_notes, '') AS inspection_notes,
			original_cost, repair_cost, created_at, updated_at
		FROM return_items
		WHERE return_id = $1
		ORDER BY created_at
	`

	err := r.db.SelectContext(ctx, &items, query, returnID)
	if err != nil {
		return nil, fmt.Errorf("failed to get return items: %w", err)
	}
	return items, nil
}

// UpdateReturnItem updates a return item
func (r *Repository) UpdateReturnItem(ctx context.Context, item *ReturnItem) error {
	if dbutil.IsSQLite(r.db) {
		now := time.Now().UTC()
		result, err := r.db.ExecContext(ctx, `UPDATE return_items SET quantity_returned=?,original_quantity=?,unit_price=?,total_refund_amount=?,original_condition=?,returned_condition=?,condition_notes=?,resolution=?,inventory_status=?,inspection_required=?,inspection_date=?,inspection_result=?,inspection_notes=?,original_cost=?,repair_cost=?,updated_at=? WHERE id=?`, item.QuantityReturned, item.OriginalQuantity, item.UnitPrice, item.TotalRefundAmount, item.OriginalCondition, item.ReturnedCondition, item.ConditionNotes, item.Resolution, item.InventoryStatus, item.InspectionRequired, item.InspectionDate, item.InspectionResult, item.InspectionNotes, item.OriginalCost, item.RepairCost, now.Format(time.RFC3339Nano), item.ID.String())
		if err != nil {
			return fmt.Errorf("failed to update return item: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return ErrReturnItemNotFound
		}
		item.UpdatedAt = now
		return nil
	}
	query := `
		UPDATE return_items
		SET quantity_returned = $2, original_quantity = $3, unit_price = $4, total_refund_amount = $5,
			original_condition = $6, returned_condition = $7, condition_notes = $8, resolution = $9, inventory_status = $10,
			inspection_required = $11, inspection_date = $12, inspection_result = $13, inspection_notes = $14,
			original_cost = $15, repair_cost = $16, updated_at = $17
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.QuantityReturned, item.OriginalQuantity, item.UnitPrice, item.TotalRefundAmount,
		item.OriginalCondition, item.ReturnedCondition, item.ConditionNotes, item.Resolution, item.InventoryStatus,
		item.InspectionRequired, item.InspectionDate, item.InspectionResult, item.InspectionNotes,
		item.OriginalCost, item.RepairCost, time.Now(),
	).Scan(&item.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReturnItemNotFound
		}
		return fmt.Errorf("failed to update return item: %w", err)
	}
	return nil
}

// DeleteReturnItem deletes a return item
func (r *Repository) DeleteReturnItem(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM return_items WHERE id = $1`
	arg := interface{}(id)
	if dbutil.IsSQLite(r.db) {
		query = `DELETE FROM return_items WHERE id = ?`
		arg = id.String()
	}

	result, err := r.db.ExecContext(ctx, query, arg)
	if err != nil {
		return fmt.Errorf("failed to delete return item: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrReturnItemNotFound
	}

	return nil
}

// GetCustomerInfo retrieves customer information
func (r *Repository) GetCustomerInfo(ctx context.Context, customerID uuid.UUID) (*CustomerInfo, error) {
	if dbutil.IsSQLite(r.db) {
		var row struct {
			ID    string `db:"id"`
			Name  string `db:"name"`
			Phone string `db:"phone"`
			Email string `db:"email"`
		}
		if err := r.db.GetContext(ctx, &row, `SELECT id,name,COALESCE(phone,''),COALESCE(email,'') FROM customers WHERE id = ?`, customerID.String()); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrCustomerNotFound
			}
			return nil, fmt.Errorf("failed to get customer info: %w", err)
		}
		id, _ := uuid.Parse(row.ID)
		return &CustomerInfo{ID: id, Name: row.Name, Phone: row.Phone, Email: row.Email}, nil
	}
	var customer CustomerInfo
	query := `SELECT id, name, COALESCE(phone, '') AS phone, COALESCE(email, '') AS email FROM customers WHERE id = $1`

	err := r.db.GetContext(ctx, &customer, query, customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer info: %w", err)
	}
	return &customer, nil
}

// GetSaleInfo retrieves sale information
func (r *Repository) GetSaleInfo(ctx context.Context, saleID uuid.UUID) (*SaleInfo, error) {
	if dbutil.IsSQLite(r.db) {
		var row struct {
			ID            string  `db:"id"`
			InvoiceNumber string  `db:"invoice_number"`
			SaleDate      string  `db:"sale_date"`
			TotalAmount   float64 `db:"total_amount"`
			CustomerID    string  `db:"customer_id"`
		}
		if err := r.db.GetContext(ctx, &row, `SELECT id,invoice_number,sale_date,total_amount,customer_id FROM sales WHERE id = ?`, saleID.String()); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrSaleNotFound
			}
			return nil, fmt.Errorf("failed to get sale info: %w", err)
		}
		id, _ := uuid.Parse(row.ID)
		cid, _ := uuid.Parse(row.CustomerID)
		return &SaleInfo{ID: id, InvoiceNumber: row.InvoiceNumber, SaleDate: localTime(sql.NullString{String: row.SaleDate, Valid: row.SaleDate != ""}), TotalAmount: row.TotalAmount, CustomerID: cid}, nil
	}
	var sale SaleInfo
	query := `SELECT id, invoice_number, sale_date, total_amount, COALESCE(customer_id, '00000000-0000-0000-0000-000000000000'::uuid) AS customer_id FROM sales WHERE id = $1`

	err := r.db.GetContext(ctx, &sale, query, saleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSaleNotFound
		}
		return nil, fmt.Errorf("failed to get sale info: %w", err)
	}
	return &sale, nil
}

type SaleItemInfo struct {
	ID        uuid.UUID `db:"id"`
	ProductID uuid.UUID `db:"product_id"`
	Quantity  int       `db:"quantity"`
	UnitPrice float64   `db:"unit_price"`
}

// GetSaleItemInfo retrieves sale item information
func (r *Repository) GetSaleItemInfo(ctx context.Context, saleItemID uuid.UUID) (SaleItemInfo, error) {
	if dbutil.IsSQLite(r.db) {
		var row struct {
			ID        string  `db:"id"`
			ProductID string  `db:"product_id"`
			Quantity  int     `db:"quantity"`
			UnitPrice float64 `db:"unit_price"`
		}
		if err := r.db.GetContext(ctx, &row, `SELECT id,product_id,quantity,unit_price FROM sale_items WHERE id = ?`, saleItemID.String()); err != nil {
			if err == sql.ErrNoRows {
				return SaleItemInfo{}, ErrSaleItemNotFound
			}
			return SaleItemInfo{}, fmt.Errorf("failed to get sale item info: %w", err)
		}
		id, _ := uuid.Parse(row.ID)
		pid, _ := uuid.Parse(row.ProductID)
		return SaleItemInfo{ID: id, ProductID: pid, Quantity: row.Quantity, UnitPrice: row.UnitPrice}, nil
	}
	var item SaleItemInfo

	query := `SELECT id, product_id, quantity, unit_price FROM sale_items WHERE id = $1`

	err := r.db.GetContext(ctx, &item, query, saleItemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return item, ErrSaleItemNotFound
		}
		return item, fmt.Errorf("failed to get sale item info: %w", err)
	}
	return item, nil
}

// GetReturnByReturnNumber retrieves a return by return number
func (r *Repository) GetReturnByReturnNumber(ctx context.Context, returnNumber string) (*Return, error) {
	if dbutil.IsSQLite(r.db) {
		var row localReturnRow
		if err := r.db.GetContext(ctx, &row, `SELECT `+localReturnColumns+` FROM returns WHERE return_number = ?`, returnNumber); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrReturnNotFound
			}
			return nil, fmt.Errorf("failed to get return by return number: %w", err)
		}
		model := row.model()
		return &model, nil
	}
	var returnRecord Return
	query := `
		SELECT r.id, r.return_number, r.reference_number, r.sale_id, r.purchase_id, r.customer_id, COALESCE(c.name, '') AS customer_name,
			return_date, return_type, status, total_refund_amount, COALESCE(refund_method, '') AS refund_method, refund_date, refund_reference,
			debt_id, debt_adjustment, customer_credit, reason, reason_detail, item_condition_after_return,
			is_warranty_claim, warranty_id, warranty_valid_until, created_by, processed_by, approved_by, approved_at,
			COALESCE(notes, '') AS notes, COALESCE(internal_notes, '') AS internal_notes, created_at, updated_at
		FROM returns r
		LEFT JOIN customers c ON c.id = r.customer_id
		WHERE r.return_number = $1
	`

	err := r.db.GetContext(ctx, &returnRecord, query, returnNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReturnNotFound
		}
		return nil, fmt.Errorf("failed to get return by return number: %w", err)
	}
	return &returnRecord, nil
}

// GetReturnsBySaleID retrieves returns for a specific sale
func (r *Repository) GetReturnsBySaleID(ctx context.Context, saleID uuid.UUID) ([]Return, error) {
	if dbutil.IsSQLite(r.db) {
		var rows []localReturnRow
		if err := r.db.SelectContext(ctx, &rows, `SELECT `+localReturnColumns+` FROM returns WHERE sale_id = ? ORDER BY return_date DESC`, saleID.String()); err != nil {
			return nil, fmt.Errorf("failed to get returns by sale: %w", err)
		}
		result := make([]Return, 0, len(rows))
		for _, row := range rows {
			result = append(result, row.model())
		}
		return result, nil
	}
	var returns []Return
	query := `
		SELECT r.id, r.return_number, r.reference_number, r.sale_id, r.purchase_id, r.customer_id, COALESCE(c.name, '') AS customer_name,
			return_date, return_type, status, total_refund_amount, COALESCE(refund_method, '') AS refund_method, refund_date, refund_reference,
			debt_id, debt_adjustment, customer_credit, reason, reason_detail, item_condition_after_return,
			is_warranty_claim, warranty_id, warranty_valid_until, created_by, processed_by, approved_by, approved_at,
			notes, internal_notes, created_at, updated_at
		FROM returns r
		LEFT JOIN customers c ON c.id = r.customer_id
		WHERE r.sale_id = $1
		ORDER BY return_date DESC
	`

	err := r.db.SelectContext(ctx, &returns, query, saleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get returns by sale: %w", err)
	}
	return returns, nil
}

// GetReturnItemByID retrieves a return item by ID
func (r *Repository) GetReturnItemByID(ctx context.Context, itemID uuid.UUID) (*ReturnItem, error) {
	if dbutil.IsSQLite(r.db) {
		var row localReturnItemRow
		if err := r.db.GetContext(ctx, &row, `SELECT id,return_id,sale_item_id,product_id,inventory_item_id,serial_number,barcode,quantity_returned,original_quantity,unit_price,total_refund_amount,original_condition,returned_condition,condition_notes,resolution,inventory_status,inspection_required,inspection_date,inspection_result,inspection_notes,original_cost,repair_cost,created_at,updated_at FROM return_items WHERE id = ?`, itemID.String()); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrReturnItemNotFound
			}
			return nil, fmt.Errorf("failed to get return item: %w", err)
		}
		model := row.model()
		return &model, nil
	}
	var item ReturnItem
	query := `
		SELECT id, return_id, sale_item_id, product_id, inventory_item_id, serial_number, barcode,
			quantity_returned, original_quantity, unit_price, total_refund_amount,
			original_condition, returned_condition, condition_notes, resolution, inventory_status,
			inspection_required, inspection_date, inspection_result, inspection_notes,
			original_cost, repair_cost, created_at, updated_at
		FROM return_items
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &item, query, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReturnItemNotFound
		}
		return nil, fmt.Errorf("failed to get return item: %w", err)
	}
	return &item, nil
}

// GetReturnedQuantity gets the total returned quantity for a sale item
func (r *Repository) GetReturnedQuantity(ctx context.Context, saleItemID uuid.UUID) (int, error) {
	var returnedQty int
	query := `
		SELECT COALESCE(SUM(quantity_returned), 0)
		FROM return_items
		WHERE sale_item_id = $1
	`

	arg := interface{}(saleItemID)
	if dbutil.IsSQLite(r.db) {
		query = `SELECT COALESCE(SUM(quantity_returned),0) FROM return_items WHERE sale_item_id = ?`
		arg = saleItemID.String()
	}
	err := r.db.GetContext(ctx, &returnedQty, query, arg)
	if err != nil {
		return 0, fmt.Errorf("failed to get returned quantity: %w", err)
	}
	return returnedQty, nil
}

// GetReturnSummary gets summary of returns with related data
func (r *Repository) GetReturnSummary(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT * FROM returns_summary
		ORDER BY return_date DESC
		LIMIT 100
	`

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get return summary: %w", err)
	}
	defer rows.Close()
	var summary []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			return nil, fmt.Errorf("failed to scan return summary: %w", err)
		}
		summary = append(summary, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read return summary: %w", err)
	}
	return summary, nil
}

// ReverseReturn reverses a return (instead of deleting it)
func (r *Repository) ReverseReturn(ctx context.Context, id uuid.UUID, reversedBy uuid.UUID) error {
	if dbutil.IsSQLite(r.db) {
		result, err := r.db.ExecContext(ctx, `UPDATE returns SET status='REVERSED',processed_by=?,updated_at=? WHERE id=?`, idArgPtr(&reversedBy), time.Now().UTC().Format(time.RFC3339Nano), id.String())
		if err != nil {
			return fmt.Errorf("failed to reverse return: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return ErrReturnNotFound
		}
		return nil
	}
	query := `
		UPDATE returns 
		SET status = 'REVERSED', 
		    processed_by = $2, 
		    updated_at = NOW() 
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query, id, reversedBy).Scan(new(time.Time))
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReturnNotFound
		}
		return fmt.Errorf("failed to reverse return: %w", err)
	}
	return nil
}

// GetReturnsByCustomer retrieves returns for a specific customer
func (r *Repository) GetReturnsByCustomer(ctx context.Context, customerID uuid.UUID) ([]Return, error) {
	if dbutil.IsSQLite(r.db) {
		var rows []localReturnRow
		if err := r.db.SelectContext(ctx, &rows, `SELECT `+localReturnColumns+` FROM returns WHERE customer_id = ? ORDER BY return_date DESC`, customerID.String()); err != nil {
			return nil, fmt.Errorf("failed to get returns by customer: %w", err)
		}
		result := make([]Return, 0, len(rows))
		for _, row := range rows {
			result = append(result, row.model())
		}
		return result, nil
	}
	var returns []Return
	query := `
		SELECT r.id, r.return_number, r.reference_number, r.sale_id, r.purchase_id, r.customer_id, COALESCE(c.name, '') AS customer_name,
			return_date, return_type, status, total_refund_amount, COALESCE(refund_method, '') AS refund_method, refund_date, COALESCE(refund_reference, '') AS refund_reference,
			debt_id, debt_adjustment, customer_credit, reason, reason_detail, item_condition_after_return,
			is_warranty_claim, warranty_id, warranty_valid_until, created_by, processed_by, approved_by, approved_at,
			notes, internal_notes, created_at, updated_at
		FROM returns r
		LEFT JOIN customers c ON c.id = r.customer_id
		WHERE r.customer_id = $1
		ORDER BY return_date DESC
	`

	err := r.db.SelectContext(ctx, &returns, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get returns by customer: %w", err)
	}
	return returns, nil
}

// GetReturnBySale retrieves a return for a specific sale (for validation)
func (r *Repository) GetReturnBySale(ctx context.Context, saleID uuid.UUID) (*Return, error) {
	if dbutil.IsSQLite(r.db) {
		var row localReturnRow
		if err := r.db.GetContext(ctx, &row, `SELECT `+localReturnColumns+` FROM returns WHERE sale_id = ? ORDER BY created_at DESC LIMIT 1`, saleID.String()); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrReturnNotFound
			}
			return nil, fmt.Errorf("failed to get return by sale: %w", err)
		}
		model := row.model()
		return &model, nil
	}
	var returnRecord Return
	query := `
		SELECT r.id, r.return_number, r.reference_number, r.sale_id, r.purchase_id, r.customer_id, COALESCE(c.name, '') AS customer_name,
			return_date, return_type, status, total_refund_amount, COALESCE(refund_method, '') AS refund_method, refund_date, refund_reference,
			debt_id, debt_adjustment, customer_credit, reason, reason_detail, item_condition_after_return,
			is_warranty_claim, warranty_id, warranty_valid_until, created_by, processed_by, approved_by, approved_at,
			notes, internal_notes, created_at, updated_at
		FROM returns r
		LEFT JOIN customers c ON c.id = r.customer_id
		WHERE r.sale_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &returnRecord, query, saleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReturnNotFound
		}
		return nil, fmt.Errorf("failed to get return by sale: %w", err)
	}
	return &returnRecord, nil
}

// GetReturnStatistics returns statistics about returns
func (r *Repository) GetReturnStatistics(ctx context.Context) (map[string]interface{}, error) {
	var row struct {
		TotalReturns     int     `db:"total_returns"`
		CompletedReturns int     `db:"completed_returns"`
		PendingReturns   int     `db:"pending_returns"`
		FullReturns      int     `db:"full_returns"`
		PartialReturns   int     `db:"partial_returns"`
		TotalRefunded    float64 `db:"total_refunded"`
		DefectiveReturns int     `db:"defective_returns"`
		WarrantyReturns  int     `db:"warranty_returns"`
	}
	query := `
		SELECT 
			COUNT(*) as total_returns,
			COUNT(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN 1 END) as completed_returns,
			COUNT(CASE WHEN UPPER(COALESCE(status, '')) = 'PENDING' THEN 1 END) as pending_returns,
			COUNT(CASE WHEN UPPER(COALESCE(return_type, '')) = 'FULL' THEN 1 END) as full_returns,
			COUNT(CASE WHEN UPPER(COALESCE(return_type, '')) IN ('PARTIAL', 'QUANTITY_PARTIAL') THEN 1 END) as partial_returns,
			COALESCE(SUM(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN total_refund_amount ELSE 0 END), 0) as total_refunded,
			COUNT(CASE WHEN UPPER(COALESCE(reason, '')) = 'DEFECTIVE' THEN 1 END) as defective_returns,
			COUNT(CASE WHEN UPPER(COALESCE(reason, '')) = 'WARRANTY' THEN 1 END) as warranty_returns
		FROM returns
		WHERE return_date >= date('now', '-30 days')
	`
	if !dbutil.IsSQLite(r.db) {
		query = `
			SELECT
				COUNT(*) AS total_returns,
				COUNT(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN 1 END) AS completed_returns,
				COUNT(CASE WHEN UPPER(COALESCE(status, '')) = 'PENDING' THEN 1 END) AS pending_returns,
				COUNT(CASE WHEN UPPER(COALESCE(return_type, '')) = 'FULL' THEN 1 END) AS full_returns,
				COUNT(CASE WHEN UPPER(COALESCE(return_type, '')) IN ('PARTIAL', 'QUANTITY_PARTIAL') THEN 1 END) AS partial_returns,
				COALESCE(SUM(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN total_refund_amount ELSE 0 END), 0) AS total_refunded,
				COUNT(CASE WHEN UPPER(COALESCE(reason, '')) = 'DEFECTIVE' THEN 1 END) AS defective_returns,
				COUNT(CASE WHEN UPPER(COALESCE(reason, '')) = 'WARRANTY' THEN 1 END) AS warranty_returns
			FROM returns
			WHERE return_date >= CURRENT_DATE - INTERVAL '30 days'
		`
	}

	err := r.db.GetContext(ctx, &row, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get return statistics: %w", err)
	}
	return map[string]interface{}{
		"total_returns":     row.TotalReturns,
		"completed_returns": row.CompletedReturns,
		"pending_returns":   row.PendingReturns,
		"full_returns":      row.FullReturns,
		"partial_returns":   row.PartialReturns,
		"total_refunded":    row.TotalRefunded,
		"defective_returns": row.DefectiveReturns,
		"warranty_returns":  row.WarrantyReturns,
	}, nil
}

// GetMonthlyReturnsAnalysis gets monthly returns analysis
func (r *Repository) GetMonthlyReturnsAnalysis(ctx context.Context) ([]MonthlyReturnsAnalysis, error) {
	var analysis []MonthlyReturnsAnalysis
	if strings.EqualFold(r.db.DriverName(), "sqlite") {
		var exists int
		if err := r.db.GetContext(ctx, &exists, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'view' AND name = 'monthly_returns_analysis'`); err != nil || exists == 0 {
			return []MonthlyReturnsAnalysis{}, nil
		}
		var rows []struct {
			Month                  returnTimestamp `db:"month"`
			TotalReturns           int             `db:"total_returns"`
			UniqueCustomers        int             `db:"unique_customers"`
			TotalRefundAmount      float64         `db:"total_refund_amount"`
			AvgRefundAmount        float64         `db:"avg_refund_amount"`
			FullReturns            int             `db:"full_returns"`
			PartialReturns         int             `db:"partial_returns"`
			QuantityPartialReturns int             `db:"quantity_partial_returns"`
			DefectiveReturns       int             `db:"defective_returns"`
			WarrantyReturns        int             `db:"warranty_returns"`
			WarrantyClaims         int             `db:"warranty_claims"`
			SellableItems          int             `db:"sellable_items"`
			RepairNeeded           int             `db:"repair_needed"`
			WrittenOff             int             `db:"written_off"`
		}
		if err := r.db.SelectContext(ctx, &rows, `SELECT month, total_returns, unique_customers, total_refund_amount, avg_refund_amount, full_returns, partial_returns, quantity_partial_returns, defective_returns, warranty_returns, warranty_claims, sellable_items, repair_needed, written_off FROM monthly_returns_analysis ORDER BY month DESC LIMIT 12`); err != nil {
			return nil, fmt.Errorf("failed to get monthly returns analysis: %w", err)
		}
		analysis = make([]MonthlyReturnsAnalysis, 0, len(rows))
		for _, row := range rows {
			analysis = append(analysis, MonthlyReturnsAnalysis{Month: row.Month.Time, TotalReturns: row.TotalReturns, UniqueCustomers: row.UniqueCustomers, TotalRefundAmount: row.TotalRefundAmount, AvgRefundAmount: row.AvgRefundAmount, FullReturns: row.FullReturns, PartialReturns: row.PartialReturns, QuantityPartialReturns: row.QuantityPartialReturns, DefectiveReturns: row.DefectiveReturns, WarrantyReturns: row.WarrantyReturns, WarrantyClaims: row.WarrantyClaims, SellableItems: row.SellableItems, RepairNeeded: row.RepairNeeded, WrittenOff: row.WrittenOff})
		}
		return analysis, nil
	}
	query := `
		SELECT 
			month,
			total_returns,
			unique_customers,
			total_refund_amount,
			avg_refund_amount,
			full_returns,
			partial_returns,
			quantity_partial_returns,
			defective_returns,
			warranty_returns,
			warranty_claims,
			sellable_items,
			repair_needed,
			written_off
		FROM monthly_returns_analysis
		ORDER BY month DESC
		LIMIT 12
	`

	err := r.db.SelectContext(ctx, &analysis, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly returns analysis: %w", err)
	}
	return analysis, nil
}

// GetSalesReturnsAnalysis gets sales vs returns analysis
func (r *Repository) GetSalesReturnsAnalysis(ctx context.Context) ([]SalesReturnsAnalysis, error) {
	var analysis []SalesReturnsAnalysis
	if strings.EqualFold(r.db.DriverName(), "sqlite") {
		var exists int
		if err := r.db.GetContext(ctx, &exists, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'view' AND name = 'sales_returns_analysis'`); err != nil || exists == 0 {
			return []SalesReturnsAnalysis{}, nil
		}
		var rows []struct {
			Month         returnTimestamp `db:"month"`
			TotalSales    int             `db:"total_sales"`
			GrossSales    float64         `db:"gross_sales"`
			TotalCost     float64         `db:"total_cost"`
			GrossProfit   float64         `db:"gross_profit"`
			ReturnsAmount float64         `db:"returns_amount"`
			ReturnCount   int             `db:"return_count"`
			NetSales      float64         `db:"net_sales"`
		}
		if err := r.db.SelectContext(ctx, &rows, `SELECT month, total_sales, gross_sales, total_cost, gross_profit, returns_amount, return_count, net_sales FROM sales_returns_analysis ORDER BY month DESC LIMIT 12`); err != nil {
			return nil, fmt.Errorf("failed to get sales returns analysis: %w", err)
		}
		analysis = make([]SalesReturnsAnalysis, 0, len(rows))
		for _, row := range rows {
			analysis = append(analysis, SalesReturnsAnalysis{Month: row.Month.Time, TotalSales: row.TotalSales, GrossSales: row.GrossSales, TotalCost: row.TotalCost, GrossProfit: row.GrossProfit, ReturnsAmount: row.ReturnsAmount, ReturnCount: row.ReturnCount, NetSales: row.NetSales})
		}
		return analysis, nil
	}
	query := `
		SELECT 
			month,
			total_sales,
			gross_sales,
			total_cost,
			gross_profit,
			returns_amount,
			return_count,
			net_sales
		FROM sales_returns_analysis
		ORDER BY month DESC
		LIMIT 12
	`

	err := r.db.SelectContext(ctx, &analysis, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales returns analysis: %w", err)
	}
	return analysis, nil
}

// GetPendingReturns retrieves returns that are pending approval
func (r *Repository) GetPendingReturns(ctx context.Context) ([]Return, error) {
	if dbutil.IsSQLite(r.db) {
		var rows []localReturnRow
		if err := r.db.SelectContext(ctx, &rows, `SELECT `+localReturnColumns+` FROM returns WHERE UPPER(status) = 'PENDING' ORDER BY return_date ASC`); err != nil {
			return nil, fmt.Errorf("failed to get pending returns: %w", err)
		}
		result := make([]Return, 0, len(rows))
		for _, row := range rows {
			result = append(result, row.model())
		}
		return result, nil
	}
	var returns []Return
	query := `
		SELECT r.id, r.return_number, r.reference_number, r.sale_id, r.purchase_id, r.customer_id, COALESCE(c.name, '') AS customer_name,
			return_date, return_type, status, total_refund_amount, COALESCE(refund_method, '') AS refund_method, refund_date, COALESCE(refund_reference, '') AS refund_reference,
			debt_id, debt_adjustment, customer_credit, COALESCE(reason, '') AS reason, COALESCE(reason_detail, '') AS reason_detail, COALESCE(item_condition_after_return, '') AS item_condition_after_return,
			is_warranty_claim, warranty_id, warranty_valid_until, created_by, processed_by, approved_by, approved_at,
			COALESCE(notes, '') AS notes, COALESCE(internal_notes, '') AS internal_notes, created_at, updated_at
		FROM returns r
		LEFT JOIN customers c ON c.id = r.customer_id
		WHERE r.status = 'PENDING'
		ORDER BY return_date ASC
	`

	err := r.db.SelectContext(ctx, &returns, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending returns: %w", err)
	}
	return returns, nil
}

// GetReturnWithItems retrieves a return with all its items
func (r *Repository) GetReturnWithItems(ctx context.Context, returnID uuid.UUID) (*Return, []ReturnItem, error) {
	returnRecord, err := r.GetReturnByID(ctx, returnID)
	if err != nil {
		return nil, nil, err
	}

	items, err := r.GetReturnItems(ctx, returnID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get return items: %w", err)
	}

	return returnRecord, items, nil
}

// ProcessReturnStatusChange handles status changes and their effects
func (r *Repository) ProcessReturnStatusChange(ctx context.Context, returnID uuid.UUID, newStatus string, processedBy uuid.UUID) error {
	if dbutil.IsSQLite(r.db) {
		result, err := r.db.ExecContext(ctx, `UPDATE returns SET status=?,processed_by=?,updated_at=? WHERE id=?`, newStatus, idArgPtr(&processedBy), time.Now().UTC().Format(time.RFC3339Nano), returnID.String())
		if err != nil {
			return fmt.Errorf("failed to process return status change: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return ErrReturnNotFound
		}
		return nil
	}
	query := `
		UPDATE returns 
		SET status = $2, 
		    processed_by = $3, 
		    updated_at = NOW() 
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query, returnID, newStatus, processedBy).Scan(new(time.Time))
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReturnNotFound
		}
		return fmt.Errorf("failed to process return status change: %w", err)
	}
	return nil
}

// UpdateReturnItemStatus updates the inventory status of a return item
func (r *Repository) UpdateReturnItemStatus(ctx context.Context, itemID uuid.UUID, newStatus string) error {
	if dbutil.IsSQLite(r.db) {
		result, err := r.db.ExecContext(ctx, `UPDATE return_items SET inventory_status=?,updated_at=? WHERE id=?`, newStatus, time.Now().UTC().Format(time.RFC3339Nano), itemID.String())
		if err != nil {
			return fmt.Errorf("failed to update return item status: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return ErrReturnItemNotFound
		}
		return nil
	}
	query := `
		UPDATE return_items
		SET inventory_status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query, itemID, newStatus).Scan(new(time.Time))
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReturnItemNotFound
		}
		return fmt.Errorf("failed to update return item status: %w", err)
	}
	return nil
}
