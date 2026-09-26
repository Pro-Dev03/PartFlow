package returns

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Repository handles return data operations
type Repository struct {
	db *sqlx.DB
}

type returnTimestamp struct{ time.Time }

func (t *returnTimestamp) Scan(value any) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}
	var dateKey string
	switch v := value.(type) {
	case time.Time:
		dateKey = v.Format("2006-01-02")
	case string:
		dateKey = strings.TrimSpace(v)
	case []byte:
		dateKey = strings.TrimSpace(string(v))
	default:
		parsed, err := dbutil.ParseTimestamp(value)
		if err != nil {
			return err
		}
		dateKey = parsed.Format("2006-01-02")
	}
	if len(dateKey) > len("2006-01-02") {
		dateKey = dateKey[:len("2006-01-02")]
	}
	if _, err := time.Parse("2006-01-02", dateKey); err != nil {
		return fmt.Errorf("parse return analysis store date %q: %w", dateKey, err)
	}
	start, _, err := accounting.StoreDateBounds(dateKey)
	if err != nil {
		return err
	}
	location, err := accounting.StoreLocation()
	if err != nil {
		return err
	}
	t.Time = start.In(location)
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
	CreatedByName            sql.NullString `db:"created_by_name"`
	ProcessedByName          sql.NullString `db:"processed_by_name"`
	ApprovedByName           sql.NullString `db:"approved_by_name"`
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
		CreatedBy: localUUIDPtr(row.CreatedBy), ProcessedBy: localUUIDPtr(row.ProcessedBy), ApprovedBy: localUUIDPtr(row.ApprovedBy), ApprovedAt: localTimePtr(row.ApprovedAt), CreatedByName: row.CreatedByName.String, ProcessedByName: row.ProcessedByName.String, ApprovedByName: row.ApprovedByName.String,
		Notes: row.Notes.String, InternalNotes: row.InternalNotes.String, CreatedAt: localTime(row.CreatedAt), UpdatedAt: localTime(row.UpdatedAt)}
}

const localReturnColumns = `id, return_number, COALESCE(reference_number,'') AS reference_number, COALESCE(sale_id,'') AS sale_id, COALESCE(purchase_id,'') AS purchase_id, COALESCE(customer_id, (SELECT customer_id FROM sales WHERE sales.id = returns.sale_id),'') AS customer_id, COALESCE((SELECT name FROM customers WHERE customers.id = COALESCE(returns.customer_id, (SELECT customer_id FROM sales WHERE sales.id = returns.sale_id))),'') AS customer_name, return_date, return_type, status, total_refund_amount, COALESCE(refund_method,'') AS refund_method, refund_date, COALESCE(refund_reference,'') AS refund_reference, debt_id, debt_adjustment, customer_credit, COALESCE(reason,'') AS reason, COALESCE(reason_detail,'') AS reason_detail, COALESCE(item_condition_after_return,'') AS item_condition_after_return, is_warranty_claim, warranty_id, warranty_valid_until, created_by, processed_by, approved_by, approved_at, COALESCE((SELECT first_name || ' ' || last_name FROM users WHERE users.id = returns.created_by), '') AS created_by_name, COALESCE((SELECT first_name || ' ' || last_name FROM users WHERE users.id = returns.processed_by), '') AS processed_by_name, COALESCE((SELECT first_name || ' ' || last_name FROM users WHERE users.id = returns.approved_by), '') AS approved_by_name, COALESCE(notes,'') AS notes, COALESCE(internal_notes,'') AS internal_notes, created_at, updated_at`

func sqliteHasColumns(db *sqlx.DB, table string, required ...string) bool {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false
	}
	defer rows.Close()

	found := make(map[string]bool, len(required))
	for rows.Next() {
		var cid, notNull, pk int
		var name, dataType string
		var defaultValue interface{}
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false
		}
		found[strings.ToLower(name)] = true
	}
	if err := rows.Err(); err != nil {
		return false
	}

	for _, column := range required {
		if !found[strings.ToLower(column)] {
			return false
		}
	}
	return true
}

type localReturnItemRow struct {
	ID                 sql.NullString  `db:"id"`
	ReturnID           sql.NullString  `db:"return_id"`
	SaleItemID         sql.NullString  `db:"sale_item_id"`
	ProductID          sql.NullString  `db:"product_id"`
	ProductName        sql.NullString  `db:"product_name"`
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
	item := ReturnItem{ID: localUUID(row.ID), ReturnID: localUUID(row.ReturnID), SaleItemID: localUUIDPtr(row.SaleItemID), ProductID: localUUIDPtr(row.ProductID), ProductName: row.ProductName.String, InventoryItemID: localUUIDPtr(row.InventoryItemID), SerialNumber: row.SerialNumber.String, Barcode: row.Barcode.String, QuantityReturned: row.QuantityReturned, UnitPrice: row.UnitPrice, TotalRefundAmount: row.TotalRefundAmount, OriginalCondition: row.OriginalCondition.String, ReturnedCondition: row.ReturnedCondition.String, ConditionNotes: row.ConditionNotes.String, Resolution: row.Resolution.String, InventoryStatus: row.InventoryStatus.String, InspectionRequired: row.InspectionRequired != 0, InspectionDate: localTimePtr(row.InspectionDate), InspectionResult: row.InspectionResult.String, InspectionNotes: row.InspectionNotes.String, RepairCost: row.RepairCost, CreatedAt: localTime(row.CreatedAt), UpdatedAt: localTime(row.UpdatedAt)}
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
	return r.createReturn(ctx, r.db, returnRecord)
}

// CreateReturnTx creates the return through the caller's transaction so the
// parent and every item can be committed or rolled back as one operation.
func (r *Repository) CreateReturnTx(ctx context.Context, tx *sqlx.Tx, returnRecord *Return) error {
	return r.createReturn(ctx, tx, returnRecord)
}

func (r *Repository) createReturn(ctx context.Context, executor sqlx.ExtContext, returnRecord *Return) error {
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
		_, err := executor.ExecContext(ctx, `INSERT INTO returns (id,return_number,reference_number,sale_id,purchase_id,customer_id,return_date,return_type,status,total_refund_amount,refund_method,refund_date,refund_reference,debt_id,debt_adjustment,customer_credit,reason,reason_detail,item_condition_after_return,is_warranty_claim,warranty_id,warranty_valid_until,created_by,processed_by,approved_by,approved_at,notes,internal_notes,created_at,updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, returnRecord.ID.String(), returnRecord.ReturnNumber, returnRecord.ReferenceNumber, idArg(returnRecord.SaleID), idArg(returnRecord.PurchaseID), idArg(returnRecord.CustomerID), returnRecord.ReturnDate.Format(time.RFC3339Nano), returnRecord.ReturnType, returnRecord.Status, returnRecord.TotalRefundAmount, returnRecord.RefundMethod, returnRecord.RefundDate, returnRecord.RefundReference, idArgPtr(returnRecord.DebtID), returnRecord.DebtAdjustment, returnRecord.CustomerCredit, returnRecord.Reason, returnRecord.ReasonDetail, returnRecord.ItemConditionAfterReturn, returnRecord.IsWarrantyClaim, idArgPtr(returnRecord.WarrantyID), returnRecord.WarrantyValidUntil, createdBy, idArgPtr(returnRecord.ProcessedBy), idArgPtr(returnRecord.ApprovedBy), returnRecord.ApprovedAt, returnRecord.Notes, returnRecord.InternalNotes, returnRecord.CreatedAt.Format(time.RFC3339Nano), returnRecord.UpdatedAt.Format(time.RFC3339Nano))
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
		if err := sqlx.GetContext(ctx, executor, &exists, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, *createdBy); err != nil || !exists {
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

	err := executor.QueryRowxContext(ctx, query,
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
		SELECT r.id, r.return_number, r.reference_number,
			COALESCE(r.sale_id, '00000000-0000-0000-0000-000000000000'::uuid) AS sale_id,
			COALESCE(NULLIF(s.invoice_number, ''), NULLIF(s.sale_number, ''), '') AS sale_invoice,
			COALESCE(r.purchase_id, '00000000-0000-0000-0000-000000000000'::uuid) AS purchase_id,
			COALESCE(r.customer_id, s.customer_id, '00000000-0000-0000-0000-000000000000'::uuid) AS customer_id,
			COALESCE(c.name, '') AS customer_name,
			r.return_date, r.return_type, r.status, r.total_refund_amount, COALESCE(r.refund_method, '') AS refund_method, r.refund_date, COALESCE(r.refund_reference, '') AS refund_reference,
			r.debt_id, r.debt_adjustment, r.customer_credit, COALESCE(r.reason, '') AS reason, COALESCE(r.reason_detail, '') AS reason_detail, COALESCE(r.item_condition_after_return, '') AS item_condition_after_return,
			r.is_warranty_claim, r.warranty_id, r.warranty_valid_until, r.created_by, r.processed_by, r.approved_by, r.approved_at,
			COALESCE(r.notes, '') AS notes, COALESCE(r.internal_notes, '') AS internal_notes, r.created_at, r.updated_at
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
		where := " WHERE 1=1 AND COALESCE(reference_number, '') NOT LIKE 'REV-%'"
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
			COALESCE((SELECT first_name || ' ' || last_name FROM users WHERE users.id = r.created_by), '') AS created_by_name,
			COALESCE((SELECT first_name || ' ' || last_name FROM users WHERE users.id = r.processed_by), '') AS processed_by_name,
			COALESCE((SELECT first_name || ' ' || last_name FROM users WHERE users.id = r.approved_by), '') AS approved_by_name,
			COALESCE(r.notes, '') AS notes, COALESCE(r.internal_notes, '') AS internal_notes, r.created_at, r.updated_at
		FROM returns r
		LEFT JOIN sales s ON s.id = r.sale_id
		LEFT JOIN customers c ON c.id = COALESCE(r.customer_id, s.customer_id)
		WHERE 1=1
			AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'
	`

	countQuery := `
		SELECT COUNT(*) FROM returns
		WHERE 1=1 AND COALESCE(reference_number, '') NOT LIKE 'REV-%'
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
		result, err := r.db.ExecContext(ctx, `UPDATE returns SET return_date=?,return_type=?,status=?,total_refund_amount=?,refund_method=?,refund_date=?,refund_reference=?,debt_id=?,debt_adjustment=?,customer_credit=?,reason=?,reason_detail=?,item_condition_after_return=?,is_warranty_claim=?,warranty_id=?,warranty_valid_until=?,processed_by=?,approved_by=?,approved_at=?,notes=?,internal_notes=?,updated_at=? WHERE id=? AND UPPER(COALESCE(status,'')) IN ('PENDING','APPROVED','PROCESSING')`, returnRecord.ReturnDate.Format(time.RFC3339Nano), returnRecord.ReturnType, returnRecord.Status, returnRecord.TotalRefundAmount, returnRecord.RefundMethod, returnRecord.RefundDate, returnRecord.RefundReference, idArgPtr(returnRecord.DebtID), returnRecord.DebtAdjustment, returnRecord.CustomerCredit, returnRecord.Reason, returnRecord.ReasonDetail, returnRecord.ItemConditionAfterReturn, returnRecord.IsWarrantyClaim, idArgPtr(returnRecord.WarrantyID), returnRecord.WarrantyValidUntil, idArgPtr(returnRecord.ProcessedBy), idArgPtr(returnRecord.ApprovedBy), returnRecord.ApprovedAt, returnRecord.Notes, returnRecord.InternalNotes, now.Format(time.RFC3339Nano), returnRecord.ID.String())
		if err != nil {
			return fmt.Errorf("failed to update return: %w", err)
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			var exists bool
			if err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM returns WHERE id=?)`, returnRecord.ID.String()); err != nil {
				return fmt.Errorf("check return after update conflict: %w", err)
			}
			if !exists {
				return ErrReturnNotFound
			}
			return ErrInvalidReturnStatus
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
		WHERE id = $1 AND UPPER(COALESCE(status,'')) IN ('PENDING','APPROVED','PROCESSING')
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
			var exists bool
			if checkErr := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM returns WHERE id=$1)`, returnRecord.ID); checkErr != nil {
				return fmt.Errorf("check return after update conflict: %w", checkErr)
			}
			if !exists {
				return ErrReturnNotFound
			}
			return ErrInvalidReturnStatus
		}
		return fmt.Errorf("failed to update return: %w", err)
	}
	return nil
}

// DeleteReturn removes the operational return record. Posted financial and
// stock effects of completed returns are first detached into the accounting
// effect ledger and remain unchanged.
func (r *Repository) DeleteReturn(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin return delete: %w", err)
	}
	defer tx.Rollback()

	returnID := interface{}(id)
	isSQLite := dbutil.IsSQLite(r.db)
	if isSQLite {
		returnID = id.String()
		// SQLite's deferred transaction does not lock a row on SELECT. Acquire
		// its write lock before inspecting status so completion cannot race delete.
		result, err := tx.ExecContext(ctx, `UPDATE returns SET updated_at = updated_at WHERE id = ?`, returnID)
		if err != nil {
			return fmt.Errorf("lock return before delete: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return ErrReturnNotFound
		}
	}
	var record struct {
		ReturnNumber   string  `db:"return_number"`
		Status         string  `db:"status"`
		RefundMethod   string  `db:"refund_method"`
		CustomerID     string  `db:"customer_id"`
		DebtID         *string `db:"debt_id"`
		DebtAdjustment float64 `db:"debt_adjustment"`
	}
	loadQuery := `SELECT return_number, status, COALESCE(refund_method,'') AS refund_method, COALESCE(customer_id,'') AS customer_id, debt_id, COALESCE(debt_adjustment,0) AS debt_adjustment FROM returns WHERE id = ?`
	if !isSQLite {
		loadQuery += ` FOR UPDATE`
	}
	if err := tx.GetContext(ctx, &record, tx.Rebind(loadQuery), returnID); err != nil {
		if err == sql.ErrNoRows {
			return ErrReturnNotFound
		}
		return fmt.Errorf("load return before delete: %w", err)
	}
	if strings.EqualFold(strings.TrimSpace(record.Status), "COMPLETING") {
		return ErrInvalidReturnStatus
	}
	isCompleted := strings.EqualFold(strings.TrimSpace(record.Status), "COMPLETED")
	if isCompleted {
		if err := preserveCompletedReturnEffectsTx(ctx, tx, dbutil.IsSQLite(r.db), returnID); err != nil {
			return fmt.Errorf("preserve completed return effects: %w", err)
		}
		if err := detachCompletedReturnReferencesTx(ctx, tx, dbutil.IsSQLite(r.db), returnID); err != nil {
			return fmt.Errorf("detach completed return references: %w", err)
		}
	} else if err := reverseReturnEffectsTx(ctx, tx, dbutil.IsSQLite(r.db), returnID, record.CustomerID, record.DebtID, record.DebtAdjustment); err != nil {
		return fmt.Errorf("reverse uncompleted return effects: %w", err)
	}
	var refundTransactions []string
	refundLinksExist, err := returnTableExists(tx, dbutil.IsSQLite(r.db), "return_payment_refunds")
	if err != nil {
		return fmt.Errorf("check return refund links: %w", err)
	}
	if refundLinksExist && !isCompleted {
		if err := tx.SelectContext(ctx, &refundTransactions, tx.Rebind(`SELECT DISTINCT payment_transaction_id FROM return_payment_refunds WHERE return_id = ?`), returnID); err != nil {
			return fmt.Errorf("load return payment transactions: %w", err)
		}
	}

	for _, table := range []string{"payment_refunds", "return_payment_refunds", "return_refunds", "return_inspection", "return_audit_log", "supplier_return_items", "supplier_returns", "inventory_movements", "item_history", "ledger_entries", "customer_ledger", "audit_logs"} {
		exists, err := returnTableExists(tx, dbutil.IsSQLite(r.db), table)
		if err != nil {
			return fmt.Errorf("check %s before return delete: %w", table, err)
		}
		if !exists {
			continue
		}
		if isCompleted {
			switch table {
			case "payment_refunds", "inventory_movements", "item_history", "ledger_entries", "customer_ledger":
				// These records are the posted payment, inventory, and balance
				// effects. Keep them; their reference id is now the effect-ledger id.
				continue
			case "supplier_return_items", "supplier_returns":
				// Supplier returns are independent operations. Remove only the
				// pointer back to the deleted customer-return workflow.
				if table == "supplier_returns" {
					query := `UPDATE supplier_returns SET customer_return_id = NULL WHERE customer_return_id = ?`
					if _, err := tx.ExecContext(ctx, tx.Rebind(query), returnID); err != nil {
						return fmt.Errorf("detach supplier returns: %w", err)
					}
				} else {
					query := `UPDATE supplier_return_items SET customer_return_id = NULL WHERE customer_return_id = ?`
					if _, err := tx.ExecContext(ctx, tx.Rebind(query), returnID); err != nil {
						return fmt.Errorf("detach supplier return items: %w", err)
					}
				}
				continue
			}
		}
		if table == "payment_refunds" {
			refundLinksExist, err := returnTableExists(tx, dbutil.IsSQLite(r.db), "return_payment_refunds")
			if err != nil {
				return fmt.Errorf("check return refund links: %w", err)
			}
			if !refundLinksExist {
				continue
			}
		}
		query := `DELETE FROM ` + table + ` WHERE return_id = ?`
		if table == "return_inspection" {
			query = `DELETE FROM return_inspection WHERE return_item_id IN (SELECT id FROM return_items WHERE return_id = ?)`
		}
		if table == "payment_refunds" {
			query = `DELETE FROM payment_refunds WHERE id IN (SELECT payment_refund_id FROM return_payment_refunds WHERE return_id = ? AND payment_refund_id IS NOT NULL)`
		}
		if table == "supplier_return_items" {
			query = `DELETE FROM supplier_return_items WHERE supplier_return_id IN (SELECT id FROM supplier_returns WHERE customer_return_id = ?) OR customer_return_id = ?`
			if _, err := tx.ExecContext(ctx, tx.Rebind(query), returnID, returnID); err != nil {
				return fmt.Errorf("clean supplier return items: %w", err)
			}
			continue
		}
		if table == "supplier_returns" {
			query = `DELETE FROM supplier_returns WHERE customer_return_id = ?`
		}
		if table == "inventory_movements" {
			query = `DELETE FROM inventory_movements WHERE reference_id = ? AND LOWER(COALESCE(reference_type,'')) LIKE '%return%'`
		}
		if table == "item_history" || table == "ledger_entries" {
			query = `DELETE FROM ` + table + ` WHERE reference_id = ?`
		}
		if table == "customer_ledger" {
			query = `DELETE FROM customer_ledger WHERE reference_id = ?`
		}
		if table == "audit_logs" {
			query = `DELETE FROM audit_logs WHERE entity_id = ? AND LOWER(COALESCE(entity_type, '')) IN ('return','returns','customer_return')`
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(query), returnID); err != nil {
			return fmt.Errorf("clean %s for return delete: %w", table, err)
		}
	}
	if !isCompleted {
		if err := recalculateReturnPaymentTransactionsTx(ctx, tx, dbutil.IsSQLite(r.db), refundTransactions); err != nil {
			return fmt.Errorf("recalculate payment state after return delete: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM return_items WHERE return_id = ?`), returnID); err != nil {
		return fmt.Errorf("delete return items: %w", err)
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM returns WHERE id = ?`), returnID); err != nil {
		return fmt.Errorf("delete return: %w", err)
	}

	if !isCompleted && record.CustomerID != "" && record.CustomerID != uuid.Nil.String() {
		balanceQuery := `UPDATE customers SET current_balance = COALESCE((SELECT SUM(CASE WHEN LOWER(COALESCE(type,''))='debit' THEN amount WHEN LOWER(COALESCE(type,''))='credit' THEN -amount ELSE 0 END) FROM customer_ledger WHERE customer_id = ?),0), updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		if !dbutil.IsSQLite(r.db) {
			balanceQuery = `UPDATE customers SET current_balance = COALESCE((SELECT SUM(CASE WHEN LOWER(COALESCE(type,''))='debit' THEN amount WHEN LOWER(COALESCE(type,''))='credit' THEN -amount ELSE 0 END) FROM customer_ledger WHERE customer_id = $1),0), updated_at = NOW() WHERE id = $2`
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(balanceQuery), record.CustomerID, record.CustomerID); err != nil {
			return fmt.Errorf("recalculate customer balance after return delete: %w", err)
		}
	}

	deletionID := uuid.New()
	auditQuery := `INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,new_values,created_at) VALUES ($1,NULL,'DELETE','return',$2,$3::jsonb,NOW())`
	if dbutil.IsSQLite(r.db) {
		auditQuery = `INSERT INTO audit_logs (id,user_id,action,entity_type,entity_id,new_values,created_at) VALUES (?,NULL,'DELETE','return',?,?,CURRENT_TIMESTAMP)`
	}
	snapshot := fmt.Sprintf(`{"return_number":%q,"status":%q,"financial_effect_preserved":%t}`, record.ReturnNumber, record.Status, isCompleted)
	if _, err := tx.ExecContext(ctx, tx.Rebind(auditQuery), deletionID, returnID, snapshot); err != nil {
		return fmt.Errorf("write return deletion audit: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit return delete: %w", err)
	}
	return nil
}

func preserveCompletedReturnEffectsTx(ctx context.Context, tx *sqlx.Tx, isSQLite bool, returnID interface{}) error {
	var exists int
	if err := tx.GetContext(ctx, &exists, tx.Rebind(`SELECT COUNT(*) FROM return_effects WHERE id = ?`), returnID); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	parentInsert := `INSERT INTO return_effects (id, sale_id, purchase_id, customer_id, total_refund_amount, status, return_date, refund_date, refund_method, debt_id, debt_adjustment, customer_credit, is_reversal, created_at, updated_at)
		SELECT id, sale_id, purchase_id, customer_id, total_refund_amount, 'COMPLETED', return_date, refund_date, refund_method, debt_id, COALESCE(debt_adjustment,0), COALESCE(customer_credit,0), CASE WHEN UPPER(COALESCE(reference_number,'')) LIKE 'REV-%' THEN TRUE ELSE FALSE END, COALESCE(created_at,CURRENT_TIMESTAMP), COALESCE(updated_at,created_at,CURRENT_TIMESTAMP)
		FROM returns WHERE id = ? AND UPPER(COALESCE(status,'')) = 'COMPLETED'`
	result, err := tx.ExecContext(ctx, tx.Rebind(parentInsert), returnID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("completed return snapshot missing or status changed")
	}
	itemsInsert := `INSERT INTO return_effect_items (id, return_effect_id, sale_item_id, product_id, inventory_item_id, serial_number, barcode, quantity_returned, original_quantity, unit_price, total_refund_amount, original_cost, resolution, inventory_status, created_at)
		SELECT ri.id, ri.return_id, ri.sale_item_id, COALESCE(ri.product_id,si.product_id), ri.inventory_item_id, ri.serial_number, ri.barcode, COALESCE(ri.quantity_returned,0), ri.original_quantity, COALESCE(ri.unit_price,0), COALESCE(ri.total_refund_amount,0), COALESCE(ri.original_cost,si.unit_cost,p.cost_price,0), ri.resolution, ri.inventory_status, COALESCE(ri.created_at,CURRENT_TIMESTAMP)
		FROM return_items ri
		LEFT JOIN sale_items si ON si.id = ri.sale_item_id
		LEFT JOIN products p ON p.id = COALESCE(ri.product_id,si.product_id)
		WHERE ri.return_id = ?`
	if _, err := tx.ExecContext(ctx, tx.Rebind(itemsInsert), returnID); err != nil {
		return err
	}
	refundTable, err := returnTableExists(tx, isSQLite, "return_refunds")
	if err != nil {
		return err
	}
	if refundTable {
		refundsInsert := `INSERT INTO return_effect_refunds (id, return_effect_id, refund_type, amount, refund_date, payment_method, transaction_reference, debt_id, debt_reduction_amount, created_at)
			SELECT id, return_id, refund_type, amount, refund_date, payment_method, transaction_reference, debt_id, debt_reduction_amount, created_at
			FROM return_refunds WHERE return_id = ?`
		if _, err := tx.ExecContext(ctx, tx.Rebind(refundsInsert), returnID); err != nil {
			return err
		}
	}
	return nil
}

func detachCompletedReturnReferencesTx(ctx context.Context, tx *sqlx.Tx, isSQLite bool, returnID interface{}) error {
	for _, item := range []struct{ table, query string }{
		{"inventory_movements", `UPDATE inventory_movements SET reference_type='return_effect' WHERE reference_id=? AND LOWER(COALESCE(reference_type,'')) LIKE '%return%'`},
		{"item_history", `UPDATE item_history SET reference_type='return_effect' WHERE reference_id=? AND LOWER(COALESCE(reference_type,'')) LIKE '%return%'`},
		{"ledger_entries", `UPDATE ledger_entries SET reference_type='return_effect' WHERE reference_id=? AND LOWER(COALESCE(reference_type,'')) LIKE '%return%'`},
		{"customer_ledger", `UPDATE customer_ledger SET reference_type='return_effect' WHERE reference_id=? AND LOWER(COALESCE(reference_type,'')) LIKE '%return%'`},
	} {
		exists, err := returnTableExists(tx, isSQLite, item.table)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(item.query), returnID); err != nil {
			// Legacy schemas may not have a reference_type on every ledger table;
			// its reference_id still resolves to the effect ledger id.
			if !strings.Contains(strings.ToLower(err.Error()), "no such column") && !strings.Contains(strings.ToLower(err.Error()), "does not exist") {
				return err
			}
		}
	}
	return nil
}

func recalculateReturnPaymentTransactionsTx(ctx context.Context, tx *sqlx.Tx, isSQLite bool, transactionIDs []string) error {
	if len(transactionIDs) == 0 {
		return nil
	}
	for _, transactionID := range transactionIDs {
		var state struct {
			AmountMinor   int64 `db:"amount_minor"`
			RefundedMinor int64 `db:"refunded_minor"`
		}
		query := `SELECT pt.amount_minor, COALESCE(SUM(CASE WHEN pr.status IN ('refunded','succeeded') THEN pr.amount_minor ELSE 0 END),0) AS refunded_minor FROM payment_transactions pt LEFT JOIN payment_refunds pr ON pr.payment_transaction_id=pt.id WHERE pt.id=? GROUP BY pt.id,pt.amount_minor`
		if !isSQLite {
			query = `SELECT pt.amount_minor, COALESCE(SUM(CASE WHEN pr.status IN ('refunded','succeeded') THEN pr.amount_minor ELSE 0 END),0) AS refunded_minor FROM payment_transactions pt LEFT JOIN payment_refunds pr ON pr.payment_transaction_id=pt.id WHERE pt.id=$1 GROUP BY pt.id,pt.amount_minor`
		}
		if err := tx.GetContext(ctx, &state, tx.Rebind(query), transactionID); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return err
		}
		status := "paid"
		if state.RefundedMinor > 0 {
			status = "partially_refunded"
			if state.RefundedMinor >= state.AmountMinor {
				status = "refunded"
			}
		}
		updatedAt := `CURRENT_TIMESTAMP`
		if !isSQLite {
			updatedAt = `NOW()`
		}
		update := `UPDATE payment_transactions SET status=?,updated_at=` + updatedAt + ` WHERE id=?`
		if !isSQLite {
			update = `UPDATE payment_transactions SET status=$1,updated_at=` + updatedAt + ` WHERE id=$2`
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(update), status, transactionID); err != nil {
			return err
		}
	}
	return nil
}

func reverseReturnEffectsTx(ctx context.Context, tx *sqlx.Tx, isSQLite bool, returnID interface{}, customerID string, debtID *string, debtAdjustment float64) error {
	var movementTable bool
	if exists, err := returnTableExists(tx, isSQLite, "inventory_movements"); err != nil {
		return err
	} else {
		movementTable = exists
	}
	if movementTable {
		var movements []struct {
			ItemID    sql.NullString `db:"item_id"`
			ProductID sql.NullString `db:"product_id"`
			Quantity  int            `db:"quantity"`
		}
		if err := tx.SelectContext(ctx, &movements, tx.Rebind(`SELECT item_id, product_id, quantity FROM inventory_movements WHERE reference_id = ? AND LOWER(COALESCE(reference_type,'')) LIKE '%return%'`), returnID); err != nil {
			return err
		}
		for _, movement := range movements {
			if movement.Quantity <= 0 || !movement.ProductID.Valid || movement.ProductID.String == "" {
				continue
			}
			if !movement.ItemID.Valid || movement.ItemID.String == "" {
				query := `UPDATE inventory SET quantity=MAX(0,COALESCE(quantity,0)-?),updated_at=CURRENT_TIMESTAMP WHERE product_id=?`
				if !isSQLite {
					query = `UPDATE inventory SET quantity=GREATEST(0,COALESCE(quantity,0)-$1),updated_at=NOW() WHERE product_id=$2`
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(query), movement.Quantity, movement.ProductID.String); err != nil {
					return fmt.Errorf("reverse aggregate return stock: %w", err)
				}
				continue
			}
			var status string
			if err := tx.GetContext(ctx, &status, tx.Rebind(`SELECT COALESCE(status,'') FROM inventory_items WHERE id = ?`), movement.ItemID.String); err != nil && err != sql.ErrNoRows {
				return fmt.Errorf("read returned item status: %w", err)
			}
			if strings.EqualFold(status, "AVAILABLE") {
				query := `UPDATE inventory SET quantity=MAX(0,COALESCE(quantity,0)-?),updated_at=CURRENT_TIMESTAMP WHERE product_id=?`
				if !isSQLite {
					query = `UPDATE inventory SET quantity=GREATEST(0,COALESCE(quantity,0)-$1),updated_at=NOW() WHERE product_id=$2`
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(query), movement.Quantity, movement.ProductID.String); err != nil {
					return fmt.Errorf("reverse returned unit stock: %w", err)
				}
			}
		}
	}

	itemTable, err := returnTableExists(tx, isSQLite, "return_items")
	if err != nil {
		return err
	}
	if itemTable {
		soldAt := `(SELECT MAX(si.created_at) FROM sale_items si WHERE si.inventory_item_id = inventory_items.id)`
		updatedAt := `CURRENT_TIMESTAMP`
		if !isSQLite {
			updatedAt = `NOW()`
		}
		query := `UPDATE inventory_items SET status='SOLD', sold_at=COALESCE(` + soldAt + `, sold_at), updated_at=` + updatedAt + ` WHERE id IN (SELECT inventory_item_id FROM return_items WHERE return_id = ? AND inventory_item_id IS NOT NULL)`
		if _, err := tx.ExecContext(ctx, tx.Rebind(query), returnID); err != nil {
			return fmt.Errorf("restore returned inventory units to sold state: %w", err)
		}
	}

	if debtAdjustment > 0 && debtID != nil && strings.TrimSpace(*debtID) != "" {
		query := `UPDATE debts SET paid_amount=MAX(0,COALESCE(paid_amount,0)-?),remaining_amount=MIN(amount,COALESCE(remaining_amount,0)+?),status=CASE WHEN COALESCE(remaining_amount,0)+? >= amount THEN 'pending' WHEN COALESCE(remaining_amount,0)+? <= 0 THEN 'paid' ELSE 'partial' END,updated_at=CURRENT_TIMESTAMP WHERE id=?`
		if !isSQLite {
			query = `UPDATE debts SET paid_amount=GREATEST(0,COALESCE(paid_amount,0)-$1),remaining_amount=LEAST(amount,COALESCE(remaining_amount,0)+$2),status=CASE WHEN COALESCE(remaining_amount,0)+$3 >= amount THEN 'pending' WHEN COALESCE(remaining_amount,0)+$4 <= 0 THEN 'paid' ELSE 'partial' END,updated_at=NOW() WHERE id=$5`
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(query), debtAdjustment, debtAdjustment, debtAdjustment, debtAdjustment, *debtID); err != nil {
			return fmt.Errorf("reverse return debt adjustment: %w", err)
		}
	}

	if err := reverseReturnSupplierBridgeTx(ctx, tx, isSQLite, returnID); err != nil {
		return err
	}
	_ = customerID // Customer balance is recalculated after the return's ledger row is removed.
	return nil
}

func reverseReturnSupplierBridgeTx(ctx context.Context, tx *sqlx.Tx, isSQLite bool, returnID interface{}) error {
	exists, err := returnTableExists(tx, isSQLite, "supplier_returns")
	if err != nil || !exists {
		return err
	}
	var rows []struct {
		ID         string `db:"id"`
		SupplierID string `db:"supplier_id"`
	}
	if err := tx.SelectContext(ctx, &rows, tx.Rebind(`SELECT id, COALESCE(supplier_id,'') AS supplier_id FROM supplier_returns WHERE customer_return_id = ?`), returnID); err != nil {
		return err
	}
	for _, row := range rows {
		if row.SupplierID != "" {
			ledgerExists, err := returnTableExists(tx, isSQLite, "supplier_ledger")
			if err != nil {
				return err
			}
			if ledgerExists {
				if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM supplier_ledger WHERE reference_id = ?`), row.ID); err != nil {
					return fmt.Errorf("remove supplier return ledger: %w", err)
				}
				balanceQuery := `UPDATE suppliers SET current_balance=COALESCE((SELECT SUM(CASE WHEN type='debit' OR transaction_type='PURCHASE' THEN amount ELSE -amount END) FROM supplier_ledger WHERE supplier_id=?),0),updated_at=CURRENT_TIMESTAMP WHERE id=?`
				if !isSQLite {
					balanceQuery = `UPDATE suppliers SET current_balance=COALESCE((SELECT SUM(CASE WHEN type='debit' OR transaction_type='PURCHASE' THEN amount ELSE -amount END) FROM supplier_ledger WHERE supplier_id=$1),0),updated_at=NOW() WHERE id=$2`
				}
				if _, err := tx.ExecContext(ctx, tx.Rebind(balanceQuery), row.SupplierID, row.SupplierID); err != nil {
					return fmt.Errorf("recalculate supplier balance after return delete: %w", err)
				}
			}
		}
		movementExists, err := returnTableExists(tx, isSQLite, "inventory_movements")
		if err != nil {
			return err
		}
		if movementExists {
			if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM inventory_movements WHERE reference_id = ? AND LOWER(COALESCE(reference_type,'')) LIKE '%supplier_return%'`), row.ID); err != nil {
				return fmt.Errorf("remove supplier return movement: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM supplier_return_items WHERE supplier_return_id = ?`), row.ID); err != nil {
			return fmt.Errorf("remove supplier return items: %w", err)
		}
	}
	return nil
}

// archiveReturnLegacy is retained for migration compatibility; the API no longer exposes it.
func (r *Repository) archiveReturnLegacy(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin return archive: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	returnID := interface{}(id)
	if dbutil.IsSQLite(r.db) {
		returnID = id.String()
	}
	var returnCount int
	if err := tx.GetContext(ctx, &returnCount, tx.Rebind(`SELECT COUNT(*) FROM returns WHERE id = ?`), returnID); err != nil {
		return fmt.Errorf("failed to find return: %w", err)
	}
	if returnCount == 0 {
		return ErrReturnNotFound
	}

	_, err = tx.ExecContext(ctx, tx.Rebind(`
		UPDATE inventory_items
		SET status = 'ARCHIVED', updated_at = CURRENT_TIMESTAMP
		WHERE id IN (SELECT inventory_item_id FROM return_items WHERE return_id = ? AND inventory_item_id IS NOT NULL)
	`), returnID)
	if err != nil {
		return fmt.Errorf("failed to archive returned inventory items: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit return archive: %w", err)
	}
	committed = true
	return nil
}

// DeleteReturnPermanently removes a return from active operation while preserving linked history.
func (r *Repository) deleteReturnLegacyPermanent(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin permanent return deletion: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	returnID := interface{}(id)
	if dbutil.IsSQLite(r.db) {
		returnID = id.String()
	}
	var returnCount int
	if err := tx.GetContext(ctx, &returnCount, tx.Rebind(`SELECT COUNT(*) FROM returns WHERE id = ?`), returnID); err != nil {
		return fmt.Errorf("failed to find return: %w", err)
	}
	if returnCount == 0 {
		return ErrReturnNotFound
	}

	deleteIfPresent := func(table, query string, args ...interface{}) error {
		exists, err := returnTableExists(tx, dbutil.IsSQLite(r.db), table)
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(query), args...); err != nil {
			return fmt.Errorf("failed to clean %s: %w", table, err)
		}
		return nil
	}

	// Keep the physical item out of active stock before removing the return rows.
	if err := deleteIfPresent("inventory_items", `
		UPDATE inventory_items SET status = 'ARCHIVED', updated_at = CURRENT_TIMESTAMP
		WHERE id IN (SELECT inventory_item_id FROM return_items WHERE return_id = ? AND inventory_item_id IS NOT NULL)
	`, returnID); err != nil {
		return err
	}
	for _, cleanup := range []struct {
		table string
		query string
	}{
		{"return_payment_refunds", `DELETE FROM return_payment_refunds WHERE return_id = ?`},
		{"return_refunds", `DELETE FROM return_refunds WHERE return_id = ?`},
		{"return_inspection", `DELETE FROM return_inspection WHERE return_item_id IN (SELECT id FROM return_items WHERE return_id = ?)`},
		{"return_audit_log", `DELETE FROM return_audit_log WHERE return_id = ?`},
		{"supplier_return_items", `DELETE FROM supplier_return_items WHERE customer_return_id = ?`},
		{"supplier_returns", `DELETE FROM supplier_returns WHERE customer_return_id = ?`},
		{"inventory_movements", `DELETE FROM inventory_movements WHERE reference_id = ? AND LOWER(COALESCE(reference_type, '')) LIKE '%return%'`},
		{"item_history", `DELETE FROM item_history WHERE reference_id = ?`},
		{"ledger_entries", `DELETE FROM ledger_entries WHERE reference_id = ?`},
		{"customer_ledger", `DELETE FROM customer_ledger WHERE reference_id = ?`},
		{"audit_logs", `DELETE FROM audit_logs WHERE entity_id = ? AND LOWER(COALESCE(entity_type, '')) IN ('return', 'returns', 'customer_return')`},
	} {
		if err := deleteIfPresent(cleanup.table, cleanup.query, returnID); err != nil {
			return err
		}
	}
	if err := deleteIfPresent("return_items", `DELETE FROM return_items WHERE return_id = ?`, returnID); err != nil {
		return err
	}
	if err := deleteIfPresent("returns", `DELETE FROM returns WHERE id = ?`, returnID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit permanent return deletion: %w", err)
	}
	committed = true
	return nil
}

func returnTableExists(tx *sqlx.Tx, isSQLite bool, table string) (bool, error) {
	var count int
	var err error
	if isSQLite {
		err = tx.Get(&count, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, table)
	} else {
		err = tx.Get(&count, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1`, table)
	}
	return count > 0, err
}

// CreateReturnItem creates a new return item
func (r *Repository) CreateReturnItem(ctx context.Context, item *ReturnItem) error {
	return r.createReturnItem(ctx, r.db, item)
}

func (r *Repository) CreateReturnItemTx(ctx context.Context, tx *sqlx.Tx, item *ReturnItem) error {
	return r.createReturnItem(ctx, tx, item)
}

func (r *Repository) createReturnItem(ctx context.Context, executor sqlx.ExtContext, item *ReturnItem) error {
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
		_, err := executor.ExecContext(ctx, `INSERT INTO return_items (id,return_id,sale_item_id,product_id,inventory_item_id,serial_number,barcode,quantity_returned,original_quantity,unit_price,total_refund_amount,original_condition,returned_condition,condition_notes,resolution,inventory_status,inspection_required,inspection_date,inspection_result,inspection_notes,original_cost,repair_cost,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, item.ID.String(), item.ReturnID.String(), idArgPtr(item.SaleItemID), idArgPtr(item.ProductID), idArgPtr(item.InventoryItemID), item.SerialNumber, item.Barcode, item.QuantityReturned, item.OriginalQuantity, item.UnitPrice, item.TotalRefundAmount, item.OriginalCondition, item.ReturnedCondition, item.ConditionNotes, item.Resolution, item.InventoryStatus, item.InspectionRequired, item.InspectionDate, item.InspectionResult, item.InspectionNotes, item.OriginalCost, item.RepairCost, item.CreatedAt.Format(time.RFC3339Nano), item.UpdatedAt.Format(time.RFC3339Nano))
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

	err := executor.QueryRowxContext(ctx, query,
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

// AddDebtAdjustmentLedgerEntry records the full return amount as a customer
// credit movement. The debt trigger separately caps the debt reduction, so
// the ledger preserves any excess as customer credit without creating a
// payment entry.
func (r *Repository) AddDebtAdjustmentLedgerEntry(ctx context.Context, returnRecord *Return) error {
	if returnRecord.CustomerID == uuid.Nil || returnRecord.RefundMethod != "DEBT_ADJUSTMENT" {
		return nil
	}
	if dbutil.IsSQLite(r.db) && returnRecord.DebtID != nil {
		tx, err := r.db.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin debt adjustment: %w", err)
		}
		defer tx.Rollback()

		var existing int
		if err := tx.GetContext(ctx, &existing, `SELECT COUNT(*) FROM customer_ledger WHERE customer_id = ? AND reference_id = ? AND type = 'credit'`, returnRecord.CustomerID, returnRecord.ID); err != nil {
			return fmt.Errorf("check debt adjustment idempotency: %w", err)
		}
		if existing > 0 {
			return tx.Commit()
		}

		var currentDebt float64
		if err := tx.GetContext(ctx, &currentDebt, `SELECT COALESCE(remaining_amount, 0) FROM debts WHERE id = ?`, *returnRecord.DebtID); err != nil {
			return fmt.Errorf("read debt for adjustment: %w", err)
		}
		adjustment := returnRecord.TotalRefundAmount
		if adjustment < 0 {
			adjustment = 0
		}
		applied := adjustment
		if applied > currentDebt {
			applied = currentDebt
		}
		customerCredit := adjustment - applied
		if _, err := tx.ExecContext(ctx, `UPDATE debts SET paid_amount = MIN(amount, COALESCE(paid_amount, 0) + ?), remaining_amount = MAX(0, COALESCE(remaining_amount, 0) - ?), status = CASE WHEN COALESCE(remaining_amount, 0) - ? <= 0 THEN 'paid' ELSE status END, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, applied, applied, applied, *returnRecord.DebtID); err != nil {
			return fmt.Errorf("apply debt adjustment: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE returns SET debt_adjustment = ?, customer_credit = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, applied, customerCredit, returnRecord.ID); err != nil {
			return fmt.Errorf("store debt adjustment result: %w", err)
		}
		var previousBalance float64
		_ = tx.GetContext(ctx, &previousBalance, `SELECT COALESCE(balance, 0) FROM customer_ledger WHERE customer_id = ? ORDER BY created_at DESC LIMIT 1`, returnRecord.CustomerID)
		if _, err := tx.ExecContext(ctx, `INSERT INTO customer_ledger (id, customer_id, type, amount, balance, description, reference_id, created_at) VALUES (?, ?, 'credit', ?, ?, ?, ?, CURRENT_TIMESTAMP)`, uuid.New(), returnRecord.CustomerID, adjustment, previousBalance-adjustment, "Customer return: "+returnRecord.ReturnNumber, returnRecord.ID); err != nil {
			return fmt.Errorf("record debt adjustment ledger: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE customers SET current_balance = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, previousBalance-adjustment, returnRecord.CustomerID); err != nil {
			return fmt.Errorf("update customer credit: %w", err)
		}
		return tx.Commit()
	}

	nowSQL := dbutil.NowSQL(r.db)
	query := fmt.Sprintf(`
		INSERT INTO customer_ledger
			(id, customer_id, type, amount, balance, description, reference_id, created_at)
		SELECT $1, $2, 'credit', $3,
			COALESCE((SELECT balance FROM customer_ledger WHERE customer_id = $2 ORDER BY created_at DESC LIMIT 1), 0) - $3,
			$4, $5, %s
		WHERE NOT EXISTS (
			SELECT 1 FROM customer_ledger
			WHERE customer_id = $2 AND reference_id = $5 AND type = 'credit'
		)
	`, nowSQL)
	if _, err := r.db.ExecContext(ctx, query,
		uuid.New(), returnRecord.CustomerID, returnRecord.TotalRefundAmount,
		"Customer return: "+returnRecord.ReturnNumber, returnRecord.ID,
	); err != nil {
		return fmt.Errorf("failed to add return ledger entry: %w", err)
	}

	updateQuery := fmt.Sprintf(`
		UPDATE customers
		SET current_balance = COALESCE((
			SELECT balance FROM customer_ledger
			WHERE customer_id = $1 AND reference_id = $2 AND type = 'credit'
			ORDER BY created_at DESC LIMIT 1
		), current_balance), updated_at = %s
		WHERE id = $1
	`, nowSQL)
	if _, err := r.db.ExecContext(ctx, updateQuery, returnRecord.CustomerID, returnRecord.ID); err != nil {
		return fmt.Errorf("failed to update customer balance for return: %w", err)
	}

	return nil
}

// GetReturnItems retrieves items for a return
func (r *Repository) GetReturnItems(ctx context.Context, returnID uuid.UUID) ([]ReturnItem, error) {
	if dbutil.IsSQLite(r.db) {
		var rows []localReturnItemRow
		if err := r.db.SelectContext(ctx, &rows, `SELECT ri.id,ri.return_id,ri.sale_item_id,ri.product_id,COALESCE(NULLIF(p.name, ''),(SELECT p0.name FROM products p0 JOIN inventory_items ii0 ON ii0.product_id = p0.id WHERE ii0.id = ri.inventory_item_id LIMIT 1),(SELECT p1.name FROM products p1 JOIN inventory_items ii1 ON ii1.product_id = p1.id WHERE ii1.barcode = COALESCE(NULLIF(ri.barcode, ''), ii.barcode) LIMIT 1),(SELECT p2.name FROM products p2 WHERE p2.barcode = COALESCE(NULLIF(ri.barcode, ''), ii.barcode) LIMIT 1),(SELECT p3.name FROM sale_items si2 LEFT JOIN products p3 ON p3.id = si2.product_id WHERE si2.id = ri.sale_item_id LIMIT 1),'') AS product_name,ri.inventory_item_id,COALESCE(NULLIF(ri.serial_number, ''), ii.serial_number, '') AS serial_number,COALESCE(NULLIF(ri.barcode, ''), ii.barcode, '') AS barcode,ri.quantity_returned,COALESCE(ri.original_quantity,(SELECT si.quantity FROM sale_items si JOIN returns rr ON rr.sale_id = si.sale_id WHERE rr.id = ri.return_id AND (ri.product_id IS NULL OR si.product_id = ri.product_id) ORDER BY si.created_at LIMIT 1)) AS original_quantity,ri.unit_price,ri.total_refund_amount,ri.original_condition,ri.returned_condition,ri.condition_notes,ri.resolution,ri.inventory_status,ri.inspection_required,ri.inspection_date,ri.inspection_result,ri.inspection_notes,ri.original_cost,ri.repair_cost,ri.created_at,ri.updated_at FROM return_items ri LEFT JOIN inventory_items ii ON ii.id = ri.inventory_item_id LEFT JOIN products p ON p.id = ri.product_id WHERE ri.return_id = ? ORDER BY ri.created_at`, returnID.String()); err != nil {
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
		SELECT ri.id, ri.return_id, ri.sale_item_id, ri.product_id,
			COALESCE(NULLIF(p.name, ''), (SELECT p0.name FROM products p0 JOIN inventory_items ii0 ON ii0.product_id = p0.id WHERE ii0.id = ri.inventory_item_id LIMIT 1), (SELECT p1.name FROM products p1 JOIN inventory_items ii1 ON ii1.product_id = p1.id WHERE ii1.barcode = COALESCE(NULLIF(ri.barcode, ''), ii.barcode) LIMIT 1), (SELECT p2.name FROM products p2 WHERE p2.barcode = COALESCE(NULLIF(ri.barcode, ''), ii.barcode) LIMIT 1), (SELECT p3.name FROM sale_items si2 LEFT JOIN products p3 ON p3.id = si2.product_id WHERE si2.id = ri.sale_item_id LIMIT 1), '') AS product_name,
			ri.inventory_item_id,
			COALESCE(NULLIF(ri.serial_number, ''), ii.serial_number, '') AS serial_number,
			COALESCE(NULLIF(ri.barcode, ''), ii.barcode, '') AS barcode,
			ri.quantity_returned, COALESCE(ri.original_quantity, (SELECT si.quantity FROM sale_items si JOIN returns rr ON rr.sale_id = si.sale_id WHERE rr.id = ri.return_id AND (ri.product_id IS NULL OR si.product_id = ri.product_id) ORDER BY si.created_at LIMIT 1)) AS original_quantity, ri.unit_price, ri.total_refund_amount,
			COALESCE(ri.original_condition, '') AS original_condition, COALESCE(ri.returned_condition, '') AS returned_condition,
			COALESCE(ri.condition_notes, '') AS condition_notes, COALESCE(ri.resolution, '') AS resolution,
			COALESCE(ri.inventory_status, '') AS inventory_status,
			ri.inspection_required, ri.inspection_date, COALESCE(ri.inspection_result, '') AS inspection_result,
			COALESCE(ri.inspection_notes, '') AS inspection_notes,
			ri.original_cost, ri.repair_cost, ri.created_at, ri.updated_at
		FROM return_items ri
		LEFT JOIN inventory_items ii ON ii.id = ri.inventory_item_id
		LEFT JOIN products p ON p.id = ri.product_id
		WHERE ri.return_id = $1
		ORDER BY ri.created_at
	`

	err := r.db.SelectContext(ctx, &items, query, returnID)
	if err != nil {
		return nil, fmt.Errorf("failed to get return items: %w", err)
	}
	return items, nil
}

// UpdateReturnItem updates a return item
func (r *Repository) UpdateReturnItem(ctx context.Context, item *ReturnItem) error {
	return r.updateReturnItem(ctx, r.db, item)
}

func (r *Repository) UpdateReturnItemTx(ctx context.Context, tx *sqlx.Tx, item *ReturnItem) error {
	return r.updateReturnItem(ctx, tx, item)
}

func (r *Repository) updateReturnItem(ctx context.Context, executor sqlx.ExtContext, item *ReturnItem) error {
	if dbutil.IsSQLite(r.db) {
		now := time.Now().UTC()
		result, err := executor.ExecContext(ctx, `UPDATE return_items SET quantity_returned=?,original_quantity=?,unit_price=?,total_refund_amount=?,original_condition=?,returned_condition=?,condition_notes=?,resolution=?,inventory_status=?,inspection_required=?,inspection_date=?,inspection_result=?,inspection_notes=?,original_cost=?,repair_cost=?,updated_at=? WHERE id=?`, item.QuantityReturned, item.OriginalQuantity, item.UnitPrice, item.TotalRefundAmount, item.OriginalCondition, item.ReturnedCondition, item.ConditionNotes, item.Resolution, item.InventoryStatus, item.InspectionRequired, item.InspectionDate, item.InspectionResult, item.InspectionNotes, item.OriginalCost, item.RepairCost, now.Format(time.RFC3339Nano), item.ID.String())
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

	err := executor.QueryRowxContext(ctx, query,
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
	return r.deleteReturnItem(ctx, r.db, id)
}

func (r *Repository) DeleteReturnItemTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) error {
	return r.deleteReturnItem(ctx, tx, id)
}

func (r *Repository) deleteReturnItem(ctx context.Context, executor sqlx.ExtContext, id uuid.UUID) error {
	query := `DELETE FROM return_items WHERE id = $1`
	arg := interface{}(id)
	if dbutil.IsSQLite(r.db) {
		query = `DELETE FROM return_items WHERE id = ?`
		arg = id.String()
	}

	result, err := executor.ExecContext(ctx, query, arg)
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
		if err := r.db.GetContext(ctx, &row, `SELECT id, name, COALESCE(phone, '') AS phone, COALESCE(email, '') AS email FROM customers WHERE id = ?`, customerID.String()); err != nil {
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
			ID             string         `db:"id"`
			InvoiceNumber  string         `db:"invoice_number"`
			SaleDate       string         `db:"sale_date"`
			Subtotal       float64        `db:"subtotal"`
			DiscountAmount float64        `db:"discount_amount"`
			TotalAmount    float64        `db:"total_amount"`
			CustomerID     sql.NullString `db:"customer_id"`
		}
		if err := r.db.GetContext(ctx, &row, `SELECT id,invoice_number,sale_date,subtotal,discount_amount,total_amount,customer_id FROM sales WHERE id = ?`, saleID.String()); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrSaleNotFound
			}
			return nil, fmt.Errorf("failed to get sale info: %w", err)
		}
		id, _ := uuid.Parse(row.ID)
		var cid uuid.UUID
		if row.CustomerID.Valid {
			cid, _ = uuid.Parse(row.CustomerID.String)
		}
		return &SaleInfo{ID: id, InvoiceNumber: row.InvoiceNumber, SaleDate: localTime(sql.NullString{String: row.SaleDate, Valid: row.SaleDate != ""}), Subtotal: row.Subtotal, DiscountAmount: row.DiscountAmount, TotalAmount: row.TotalAmount, CustomerID: cid}, nil
	}
	var sale SaleInfo
	query := `SELECT id, invoice_number, sale_date, subtotal, discount_amount, total_amount, COALESCE(customer_id, '00000000-0000-0000-0000-000000000000'::uuid) AS customer_id FROM sales WHERE id = $1`

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
	ID              uuid.UUID  `db:"id"`
	ProductID       uuid.UUID  `db:"product_id"`
	InventoryItemID *uuid.UUID `db:"inventory_item_id"`
	Quantity        int        `db:"quantity"`
	UnitPrice       float64    `db:"unit_price"`
}

// GetSaleItemInfo retrieves sale item information
func (r *Repository) GetSaleItemInfo(ctx context.Context, saleItemID uuid.UUID) (SaleItemInfo, error) {
	if dbutil.IsSQLite(r.db) {
		var row struct {
			ID              string  `db:"id"`
			ProductID       string  `db:"product_id"`
			InventoryItemID string  `db:"inventory_item_id"`
			Quantity        int     `db:"quantity"`
			UnitPrice       float64 `db:"unit_price"`
		}
		if err := r.db.GetContext(ctx, &row, `SELECT id,product_id,COALESCE(inventory_item_id,'') AS inventory_item_id,quantity,unit_price FROM sale_items WHERE id = ?`, saleItemID.String()); err != nil {
			if err == sql.ErrNoRows {
				return SaleItemInfo{}, ErrSaleItemNotFound
			}
			return SaleItemInfo{}, fmt.Errorf("failed to get sale item info: %w", err)
		}
		id, _ := uuid.Parse(row.ID)
		pid, _ := uuid.Parse(row.ProductID)
		var inventoryItemID *uuid.UUID
		if parsed, parseErr := uuid.Parse(row.InventoryItemID); parseErr == nil {
			inventoryItemID = &parsed
		}
		return SaleItemInfo{ID: id, ProductID: pid, InventoryItemID: inventoryItemID, Quantity: row.Quantity, UnitPrice: row.UnitPrice}, nil
	}
	var item SaleItemInfo

	query := `SELECT id, product_id, inventory_item_id, quantity, unit_price FROM sale_items WHERE id = $1`

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
		FROM return_items ri
		JOIN returns r ON r.id = ri.return_id
		WHERE ri.sale_item_id = $1 AND UPPER(COALESCE(r.status, '')) NOT IN ('REJECTED', 'CANCELLED')
	`

	arg := interface{}(saleItemID)
	if dbutil.IsSQLite(r.db) {
		query = `SELECT COALESCE(SUM(ri.quantity_returned),0) FROM return_items ri JOIN returns r ON r.id = ri.return_id WHERE ri.sale_item_id = ? AND UPPER(COALESCE(r.status, '')) NOT IN ('REJECTED', 'CANCELLED')`
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

// ReverseReturn cancels the original return instead of creating a second return record.
func (r *Repository) ReverseReturn(ctx context.Context, id uuid.UUID, reversedBy uuid.UUID) error {
	updatedAt := interface{}(time.Now().UTC())
	returnID := interface{}(id)
	userID := interface{}(reversedBy)
	if dbutil.IsSQLite(r.db) {
		updatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		returnID = id.String()
		userID = idArgPtr(&reversedBy)
	}
	query := `UPDATE returns SET status = ?, processed_by = ?, internal_notes = ?, updated_at = ? WHERE id = ? AND UPPER(COALESCE(status,'')) IN ('PENDING','APPROVED','PROCESSING')`
	result, err := r.db.ExecContext(ctx, r.db.Rebind(query), "CANCELLED", userID, "Cancelled before return completion.", updatedAt, returnID)
	if err != nil {
		return fmt.Errorf("failed to reverse return: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		var exists bool
		if err := r.db.GetContext(ctx, &exists, r.db.Rebind(`SELECT EXISTS(SELECT 1 FROM returns WHERE id = ?)`), returnID); err != nil {
			return fmt.Errorf("check return after failed reverse: %w", err)
		}
		if !exists {
			return ErrReturnNotFound
		}
		return ErrInvalidReturnStatus
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

	isSQLite := dbutil.IsSQLite(r.db)
	periodStart, periodEnd, err := accounting.StoreDateRange(time.Now(), 30)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate return statistics period: %w", err)
	}
	returnDate := r.returnsDateExpression("")
	if isSQLite && !sqliteHasColumns(r.db, "returns", "return_type") {
		conditions := []string{fmt.Sprintf("%s >= date(?)", returnDate), fmt.Sprintf("%s < date(?)", returnDate)}
		if sqliteHasColumns(r.db, "accounting_returns", "reference_number") {
			conditions = append(conditions, "COALESCE(reference_number, '') NOT LIKE 'REV-%'")
		}
		if sqliteHasColumns(r.db, "accounting_returns", "status") {
			conditions = append(conditions, "UPPER(COALESCE(status, '')) <> 'CANCELLED'")
		}
		query := fmt.Sprintf(`
			SELECT
				COUNT(*) AS total_returns,
				COUNT(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN 1 END) AS completed_returns,
				COUNT(CASE WHEN UPPER(COALESCE(status, '')) = 'PENDING' THEN 1 END) AS pending_returns,
				0 AS full_returns,
				0 AS partial_returns,
				COALESCE(SUM(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN total_refund_amount ELSE 0 END), 0) AS total_refunded,
				COUNT(CASE WHEN UPPER(COALESCE(reason, '')) = 'DEFECTIVE' THEN 1 END) AS defective_returns,
				COUNT(CASE WHEN UPPER(COALESCE(reason, '')) = 'WARRANTY' THEN 1 END) AS warranty_returns
			FROM accounting_returns
				WHERE %s
			`, strings.Join(conditions, " AND "))
		err := r.db.GetContext(ctx, &row, query, periodStart, periodEnd)
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

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_returns,
			COUNT(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN 1 END) as completed_returns,
			COUNT(CASE WHEN UPPER(COALESCE(status, '')) = 'PENDING' THEN 1 END) as pending_returns,
			COUNT(CASE WHEN UPPER(COALESCE(return_type, '')) = 'FULL' THEN 1 END) as full_returns,
			COUNT(CASE WHEN UPPER(COALESCE(return_type, '')) IN ('PARTIAL', 'QUANTITY_PARTIAL') THEN 1 END) as partial_returns,
			COALESCE(SUM(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN total_refund_amount ELSE 0 END), 0) as total_refunded,
			COUNT(CASE WHEN UPPER(COALESCE(reason, '')) = 'DEFECTIVE' THEN 1 END) as defective_returns,
			COUNT(CASE WHEN UPPER(COALESCE(reason, '')) = 'WARRANTY' THEN 1 END) as warranty_returns
		FROM accounting_returns
		WHERE %s >= date(?) AND %s < date(?)
		  AND COALESCE(reference_number, '') NOT LIKE 'REV-%%'
		  AND UPPER(COALESCE(status, '')) <> 'CANCELLED'
	`, returnDate, returnDate)
	if !isSQLite {
		query = fmt.Sprintf(`
			SELECT
				COUNT(*) AS total_returns,
				COUNT(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN 1 END) AS completed_returns,
				COUNT(CASE WHEN UPPER(COALESCE(status, '')) = 'PENDING' THEN 1 END) AS pending_returns,
				COUNT(CASE WHEN UPPER(COALESCE(return_type, '')) = 'FULL' THEN 1 END) AS full_returns,
				COUNT(CASE WHEN UPPER(COALESCE(return_type, '')) IN ('PARTIAL', 'QUANTITY_PARTIAL') THEN 1 END) AS partial_returns,
				COALESCE(SUM(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN total_refund_amount ELSE 0 END), 0) AS total_refunded,
				COUNT(CASE WHEN UPPER(COALESCE(reason, '')) = 'DEFECTIVE' THEN 1 END) AS defective_returns,
				COUNT(CASE WHEN UPPER(COALESCE(reason, '')) = 'WARRANTY' THEN 1 END) AS warranty_returns
			FROM accounting_returns
			WHERE %s >= $1::date AND %s < $2::date
			  AND COALESCE(reference_number, '') NOT LIKE 'REV-%%'
			  AND UPPER(COALESCE(status, '')) <> 'CANCELLED'
		`, returnDate, returnDate)
	}

	err = r.db.GetContext(ctx, &row, query, periodStart, periodEnd)
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

func (r *Repository) returnsDateExpression(alias string) string {
	qualified := func(column string) string {
		if alias == "" {
			return column
		}
		return alias + "." + column
	}
	if dbutil.IsSQLite(r.db) {
		columns := make([]string, 0, 3)
		for _, column := range []string{"refund_date", "return_date", "created_at"} {
			if sqliteHasColumns(r.db, "accounting_returns", column) {
				columns = append(columns, qualified(column))
			}
		}
		if len(columns) == 0 {
			for _, column := range []string{"refund_date", "return_date", "created_at"} {
				if sqliteHasColumns(r.db, "returns", column) {
					columns = append(columns, qualified(column))
				}
			}
		}
		if len(columns) == 0 {
			return "store_date(" + qualified("return_date") + ")"
		}
		return "store_date(COALESCE(" + strings.Join(columns, ", ") + "))"
	}
	return "COALESCE(" + qualified("refund_date") + "::date, " + qualified("return_date") + "::date, " + accounting.PostgresStoreDateExpression(qualified("created_at")) + ")"
}

func (r *Repository) salesDateExpression(alias string) string {
	if dbutil.IsSQLite(r.db) {
		columns := make([]string, 0, 2)
		for _, column := range []string{"sale_date", "created_at"} {
			if sqliteHasColumns(r.db, "sales", column) {
				columns = append(columns, alias+"."+column)
			}
		}
		if len(columns) == 0 {
			return "store_date(" + alias + ".sale_date)"
		}
		return "store_date(COALESCE(" + strings.Join(columns, ", ") + "))"
	}
	return "COALESCE(" + alias + ".sale_date::date, " + accounting.PostgresStoreDateExpression(alias+".created_at") + ")"
}

func (r *Repository) monthExpression(dateExpression string) string {
	if dbutil.IsSQLite(r.db) {
		return "strftime('%Y-%m-01', " + dateExpression + ")"
	}
	return "DATE_TRUNC('month', " + dateExpression + ")"
}

// GetMonthlyReturnsAnalysis groups posted returns by their store-local refund
// date so results follow the configured timezone even when a legacy view does not.
func (r *Repository) GetMonthlyReturnsAnalysis(ctx context.Context) ([]MonthlyReturnsAnalysis, error) {
	dateExpression := r.returnsDateExpression("r")
	monthExpression := r.monthExpression(dateExpression)
	returnTable := "accounting_returns"
	if dbutil.IsSQLite(r.db) && !sqliteHasColumns(r.db, returnTable, "id") {
		returnTable = "returns"
	}
	query := fmt.Sprintf(`SELECT %s AS month,
		COUNT(DISTINCT r.id) AS total_returns,
		COUNT(DISTINCT r.customer_id) AS unique_customers,
		COALESCE(SUM(r.total_refund_amount), 0) AS total_refund_amount,
		COALESCE(AVG(r.total_refund_amount), 0) AS avg_refund_amount,
		COUNT(CASE WHEN UPPER(COALESCE(r.return_type, '')) = 'FULL' THEN 1 END) AS full_returns,
		COUNT(CASE WHEN UPPER(COALESCE(r.return_type, '')) IN ('PARTIAL', 'QUANTITY_PARTIAL') THEN 1 END) AS partial_returns,
		COUNT(CASE WHEN UPPER(COALESCE(r.return_type, '')) = 'QUANTITY_PARTIAL' THEN 1 END) AS quantity_partial_returns,
		COUNT(CASE WHEN UPPER(COALESCE(r.reason, '')) = 'DEFECTIVE' THEN 1 END) AS defective_returns,
		COUNT(CASE WHEN UPPER(COALESCE(r.reason, '')) = 'WARRANTY' THEN 1 END) AS warranty_returns,
		COUNT(CASE WHEN UPPER(COALESCE(r.status, '')) = 'COMPLETED' AND UPPER(COALESCE(r.reason, '')) = 'WARRANTY' THEN 1 END) AS warranty_claims,
		SUM(CASE WHEN UPPER(COALESCE(r.item_condition_after_return, '')) = 'SELLABLE' THEN 1 ELSE 0 END) AS sellable_items,
		SUM(CASE WHEN UPPER(COALESCE(r.item_condition_after_return, '')) = 'NEEDS_REPAIR' THEN 1 ELSE 0 END) AS repair_needed,
		SUM(CASE WHEN UPPER(COALESCE(r.item_condition_after_return, '')) IN ('WRITE_OFF', 'DAMAGED') THEN 1 ELSE 0 END) AS written_off
		FROM %s r WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
		GROUP BY %s ORDER BY month DESC LIMIT 12`, monthExpression, returnTable, monthExpression)
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
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("failed to get monthly returns analysis: %w", err)
	}
	analysis := make([]MonthlyReturnsAnalysis, 0, len(rows))
	for _, row := range rows {
		analysis = append(analysis, MonthlyReturnsAnalysis{Month: row.Month.Time, TotalReturns: row.TotalReturns, UniqueCustomers: row.UniqueCustomers, TotalRefundAmount: row.TotalRefundAmount, AvgRefundAmount: row.AvgRefundAmount, FullReturns: row.FullReturns, PartialReturns: row.PartialReturns, QuantityPartialReturns: row.QuantityPartialReturns, DefectiveReturns: row.DefectiveReturns, WarrantyReturns: row.WarrantyReturns, WarrantyClaims: row.WarrantyClaims, SellableItems: row.SellableItems, RepairNeeded: row.RepairNeeded, WrittenOff: row.WrittenOff})
	}
	return analysis, nil
}

// GetSalesReturnsAnalysis gets sales vs returns analysis
func (r *Repository) GetSalesReturnsAnalysis(ctx context.Context) ([]SalesReturnsAnalysis, error) {
	salesMonth := r.monthExpression(r.salesDateExpression("s"))
	returnMonth := r.monthExpression(r.returnsDateExpression("r"))
	returnTable := "accounting_returns"
	if dbutil.IsSQLite(r.db) && !sqliteHasColumns(r.db, returnTable, "id") {
		returnTable = "returns"
	}
	returnFilter := "UPPER(COALESCE(r.status, '')) = 'COMPLETED' AND r.sale_id IS NOT NULL"
	if !dbutil.IsSQLite(r.db) || sqliteHasColumns(r.db, returnTable, "reference_number") {
		returnFilter += " AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'"
	}
	query := fmt.Sprintf(`WITH sales_by_month AS (
		SELECT %s AS month, COUNT(*) AS total_sales,
			COALESCE(SUM(s.total_amount), 0) AS gross_sales,
			COALESCE(SUM(COALESCE(s.cost_amount, 0)), 0) AS total_cost,
			COALESCE(SUM(COALESCE(s.gross_profit, s.total_amount - COALESCE(s.tax_amount, 0) - COALESCE(s.cost_amount, 0), 0)), 0) AS gross_profit
		FROM sales s WHERE UPPER(COALESCE(s.status, '')) = 'COMPLETED'
		GROUP BY %s
	), returns_by_month AS (
		SELECT %s AS month, COALESCE(SUM(r.total_refund_amount), 0) AS returns_amount,
			COUNT(*) AS return_count FROM %s r WHERE %s GROUP BY %s
	), months AS (
		SELECT month FROM sales_by_month UNION SELECT month FROM returns_by_month
	)
	SELECT months.month, COALESCE(s.total_sales, 0) AS total_sales, COALESCE(s.gross_sales, 0) AS gross_sales,
		COALESCE(s.total_cost, 0) AS total_cost, COALESCE(s.gross_profit, 0) AS gross_profit, COALESCE(r.returns_amount, 0) AS returns_amount,
		COALESCE(r.return_count, 0) AS return_count, COALESCE(s.gross_sales, 0) - COALESCE(r.returns_amount, 0) AS net_sales
	FROM months LEFT JOIN sales_by_month s ON s.month = months.month
	LEFT JOIN returns_by_month r ON r.month = months.month
	ORDER BY months.month DESC LIMIT 12`, salesMonth, salesMonth, returnMonth, returnTable, returnFilter, returnMonth)
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
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("failed to get sales returns analysis: %w", err)
	}
	analysis := make([]SalesReturnsAnalysis, 0, len(rows))
	for _, row := range rows {
		analysis = append(analysis, SalesReturnsAnalysis{Month: row.Month.Time, TotalSales: row.TotalSales, GrossSales: row.GrossSales, TotalCost: row.TotalCost, GrossProfit: row.GrossProfit, ReturnsAmount: row.ReturnsAmount, ReturnCount: row.ReturnCount, NetSales: row.NetSales})
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

// UpdateInventoryForReturnItem applies the inventory decision made for a completed customer return.
func (r *Repository) UpdateInventoryForReturnItem(ctx context.Context, itemID uuid.UUID, newStatus string) error {
	query := `UPDATE inventory_items SET status = $2, sold_at = NULL, updated_at = NOW() WHERE id = $1`
	args := []interface{}{itemID, newStatus}
	if dbutil.IsSQLite(r.db) {
		query = `UPDATE inventory_items SET status = ?, sold_at = NULL, updated_at = ? WHERE id = ?`
		args = []interface{}{newStatus, time.Now().UTC().Format(time.RFC3339Nano), itemID.String()}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update returned inventory item: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("inventory item %s was not found", itemID)
	}
	if strings.EqualFold(strings.TrimSpace(newStatus), "AVAILABLE") {
		result, err := r.db.ExecContext(ctx, `UPDATE inventory SET quantity = quantity + 1, updated_at = CURRENT_TIMESTAMP WHERE product_id = (SELECT product_id FROM inventory_items WHERE id = $1)`, itemID)
		if err != nil {
			return fmt.Errorf("failed to restore returned inventory quantity: %w", err)
		}
		quantityRows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to read restored inventory quantity: %w", err)
		}
		if quantityRows == 0 {
			if _, err := r.db.ExecContext(ctx, `INSERT INTO inventory (id, product_id, quantity, created_at, updated_at) SELECT $1, product_id, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP FROM inventory_items WHERE id = $2`, uuid.New(), itemID); err != nil {
				return fmt.Errorf("failed to create restored inventory quantity: %w", err)
			}
		}
	}
	return nil
}
