package purchases

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

// Repository handles purchase data operations
type Repository struct {
	db *sqlx.DB
}

type localPurchaseSummaryRow struct {
	ID                   string         `db:"id"`
	InvoiceNumber        string         `db:"invoice_number"`
	PurchaseDate         string         `db:"purchase_date"`
	ExpectedDeliveryDate sql.NullString `db:"expected_delivery_date"`
	TotalAmount          float64        `db:"total_amount"`
	PaidAmount           float64        `db:"paid_amount"`
	Remaining            float64        `db:"remaining"`
	Status               string         `db:"status"`
	SupplierName         string         `db:"supplier_name"`
	TotalItems           int            `db:"total_items"`
	CreatedAt            string         `db:"created_at"`
	TotalCount           int            `db:"total_count"`
}

type localPurchaseRow struct {
	ID                   string         `db:"id"`
	SupplierID           string         `db:"supplier_id"`
	InvoiceNumber        string         `db:"invoice_number"`
	PurchaseDate         string         `db:"purchase_date"`
	ExpectedDeliveryDate sql.NullString `db:"expected_delivery_date"`
	TotalAmount          float64        `db:"total_amount"`
	PaidAmount           float64        `db:"paid_amount"`
	Status               string         `db:"status"`
	Notes                sql.NullString `db:"notes"`
	UserID               sql.NullString `db:"user_id"`
	CreatedAt            string         `db:"created_at"`
	UpdatedAt            string         `db:"updated_at"`
}

type localPurchaseItemRow struct {
	ID           string         `db:"id"`
	PurchaseID   string         `db:"purchase_id"`
	ProductID    string         `db:"product_id"`
	Quantity     int            `db:"quantity"`
	UnitCost     float64        `db:"unit_cost"`
	TotalCost    float64        `db:"total_cost"`
	SerialNumber sql.NullString `db:"serial_number"`
	Condition    sql.NullString `db:"condition"`
	LocationID   sql.NullString `db:"location_id"`
	Notes        sql.NullString `db:"notes"`
	CreatedAt    string         `db:"created_at"`
}

func parsePurchaseTime(value string) (time.Time, error) {
	return dbutil.ParseTimestamp(value)
}

// NewRepository creates a new purchase repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new purchase
func (r *Repository) Create(ctx context.Context, purchase *Purchase) error {
	if dbutil.IsSQLite(r.db) {
		return createPurchaseSQLite(ctx, r.db, purchase)
	}
	return createPurchase(ctx, r.db, purchase)
}

// CreateTx creates a purchase using the caller's transaction. Keeping this
// operation on the same executor as its child rows makes purchase creation
// atomic when a ledger or audit insert fails.
func (r *Repository) CreateTx(ctx context.Context, tx *sqlx.Tx, purchase *Purchase) error {
	if dbutil.IsSQLite(r.db) {
		return createPurchaseSQLite(ctx, tx, purchase)
	}
	return createPurchase(ctx, tx, purchase)
}

func createPurchase(ctx context.Context, executor sqlx.ExtContext, purchase *Purchase) error {
	query := `
		INSERT INTO purchases (id, supplier_id, invoice_number, purchase_date, expected_delivery_date,
			total_amount, paid_amount, status, notes, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err := executor.QueryRowxContext(ctx, query,
		purchase.ID, purchase.SupplierID, purchase.InvoiceNumber, purchase.PurchaseDate,
		purchase.ExpectedDeliveryDate, purchase.TotalAmount, purchase.PaidAmount, purchase.Status,
		purchase.Notes, purchase.UserID, purchase.CreatedAt, purchase.UpdatedAt,
	).Scan(&purchase.ID, &purchase.CreatedAt, &purchase.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create purchase: %w", err)
	}
	return nil
}

func createPurchaseSQLite(ctx context.Context, executor sqlx.ExtContext, purchase *Purchase) error {
	_, err := executor.ExecContext(ctx, `
		INSERT INTO purchases (id, purchase_number, supplier_id, total_amount, paid_amount,
			remaining_amount, status, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, purchase.ID, purchase.InvoiceNumber, purchase.SupplierID, purchase.TotalAmount,
		purchase.PaidAmount, purchase.TotalAmount-purchase.PaidAmount, purchase.Status,
		purchase.Notes, purchase.CreatedAt, purchase.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create local purchase: %w", err)
	}
	return nil
}

// GetByID retrieves a purchase by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Purchase, error) {
	if dbutil.IsSQLite(r.db) {
		var row localPurchaseRow
		err := r.db.GetContext(ctx, &row, `SELECT id, supplier_id, purchase_number AS invoice_number, created_at AS purchase_date, NULL AS expected_delivery_date, total_amount, paid_amount, status, notes, NULL AS user_id, created_at, updated_at FROM purchases WHERE id = $1`, id)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrPurchaseNotFound
			}
			return nil, fmt.Errorf("failed to get purchase: %w", err)
		}
		purchaseID, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse purchase id: %w", err)
		}
		supplierID, err := uuid.Parse(row.SupplierID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse supplier id: %w", err)
		}
		purchaseDate, err := parsePurchaseTime(row.PurchaseDate)
		if err != nil {
			return nil, err
		}
		createdAt, err := parsePurchaseTime(row.CreatedAt)
		if err != nil {
			return nil, err
		}
		updatedAt, err := parsePurchaseTime(row.UpdatedAt)
		if err != nil {
			return nil, err
		}
		purchase := &Purchase{ID: purchaseID, SupplierID: supplierID, InvoiceNumber: row.InvoiceNumber, PurchaseDate: purchaseDate, TotalAmount: row.TotalAmount, PaidAmount: row.PaidAmount, Status: row.Status, CreatedAt: createdAt, UpdatedAt: updatedAt}
		if row.ExpectedDeliveryDate.Valid && row.ExpectedDeliveryDate.String != "" {
			t, err := parsePurchaseTime(row.ExpectedDeliveryDate.String)
			if err != nil {
				return nil, err
			}
			purchase.ExpectedDeliveryDate = &t
		}
		if row.Notes.Valid {
			purchase.Notes = &row.Notes.String
		}
		if row.UserID.Valid && row.UserID.String != "" {
			u, err := uuid.Parse(row.UserID.String)
			if err == nil {
				purchase.UserID = &u
			}
		}
		return purchase, nil
	}
	var purchase Purchase
	query := `
		SELECT id, supplier_id, invoice_number, purchase_date, expected_delivery_date,
			total_amount, paid_amount, status, notes, user_id, created_at, updated_at
		FROM purchases
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &purchase, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPurchaseNotFound
		}
		return nil, fmt.Errorf("failed to get purchase: %w", err)
	}
	return &purchase, nil
}

// List retrieves purchases with pagination and filters
func (r *Repository) List(ctx context.Context, req PurchaseListRequest) ([]Purchase, int, error) {
	var purchases []Purchase
	var count int

	// Build base query
	baseQuery := `
		SELECT id, supplier_id, invoice_number, purchase_date, expected_delivery_date,
			total_amount, paid_amount, status, notes, user_id, created_at, updated_at
		FROM purchases
		WHERE 1=1
	`

	countQuery := `
		SELECT COUNT(*) FROM purchases WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if req.SupplierID != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND supplier_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND supplier_id = $%d", argCount)
		args = append(args, *req.SupplierID)
	}

	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}

	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND purchase_date >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND purchase_date >= $%d", argCount)
		args = append(args, *req.StartDate)
	}

	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND purchase_date <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND purchase_date <= $%d", argCount)
		args = append(args, *req.EndDate)
	}

	if req.Search != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND (LOWER(invoice_number) LIKE LOWER($%d) OR LOWER(notes) LIKE LOWER($%d))", argCount, argCount)
		countQuery += fmt.Sprintf(" AND (LOWER(invoice_number) LIKE LOWER($%d) OR LOWER(notes) LIKE LOWER($%d))", argCount, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern)
	}

	// Get total count
	countArgs := append([]interface{}{}, args...)
	err := r.db.GetContext(ctx, &count, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count purchases: %w", err)
	}

	// Add sorting
	sortBy := "purchase_date"
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

	err = r.db.SelectContext(ctx, &purchases, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list purchases: %w", err)
	}

	return purchases, count, nil
}

// ListSummaries returns the complete list projection in one database query.
// Keeping related data in this query avoids one supplier and one item query per purchase.
func (r *Repository) ListSummaries(ctx context.Context, req PurchaseListRequest) ([]PurchaseListItem, int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 || req.PerPage > 100 {
		req.PerPage = 20
	}
	if dbutil.IsSQLite(r.db) {
		query := `
			SELECT p.id AS id, p.purchase_number AS invoice_number, p.created_at AS purchase_date, NULL AS expected_delivery_date,
			       p.total_amount, p.paid_amount, p.total_amount - p.paid_amount AS remaining,
			       p.status, COALESCE(s.name, '') AS supplier_name, COUNT(pi.id) AS total_items,
			       p.created_at, COUNT(*) OVER() AS total_count
			FROM purchases p
			LEFT JOIN suppliers s ON s.id = p.supplier_id
			LEFT JOIN purchase_items pi ON pi.purchase_id = p.id
			WHERE 1=1`
		args := make([]interface{}, 0, 7)
		argCount := 0
		addFilter := func(condition string, value interface{}) {
			argCount++
			query += fmt.Sprintf(" AND %s $%d", condition, argCount)
			args = append(args, value)
		}
		if req.SupplierID != nil {
			addFilter("p.supplier_id =", *req.SupplierID)
		}
		if req.Status != "" {
			addFilter("p.status =", req.Status)
		}
		if req.StartDate != nil {
			addFilter("date(p.created_at) >=", req.StartDate.Format("2006-01-02"))
		}
		if req.EndDate != nil {
			addFilter("date(p.created_at) <=", req.EndDate.Format("2006-01-02"))
		}
		if req.Search != "" {
			argCount++
			query += fmt.Sprintf(" AND (LOWER(COALESCE(p.purchase_number, '')) LIKE LOWER($%d) OR LOWER(COALESCE(p.notes, '')) LIKE LOWER($%d))", argCount, argCount)
			args = append(args, "%"+req.Search+"%")
		}
		if req.AvailableForReturn {
			query += ` AND EXISTS (SELECT 1 FROM purchase_items return_pi JOIN inventory_items return_ii ON return_ii.product_id = return_pi.product_id WHERE return_pi.purchase_id = p.id AND return_ii.status = 'AVAILABLE')`
		}
		query += ` GROUP BY p.id, p.purchase_number, p.created_at, p.total_amount, p.paid_amount, p.status, s.name`
		sortColumns := map[string]string{"purchase_date": "p.created_at", "created_at": "p.created_at", "total_amount": "p.total_amount", "status": "p.status"}
		sortColumn := sortColumns[req.SortBy]
		if sortColumn == "" {
			sortColumn = "p.purchase_date"
		}
		sortOrder := "DESC"
		if strings.EqualFold(req.SortOrder, "asc") {
			sortOrder = "ASC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s, p.id DESC LIMIT $%d OFFSET $%d", sortColumn, sortOrder, argCount+1, argCount+2)
		args = append(args, req.PerPage, (req.Page-1)*req.PerPage)
		var rows []localPurchaseSummaryRow
		if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
			return nil, 0, fmt.Errorf("failed to list purchase summaries: %w", err)
		}
		summaries := make([]PurchaseListItem, 0, len(rows))
		total := 0
		for _, row := range rows {
			purchaseDate, err := parsePurchaseTime(row.PurchaseDate)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to parse purchase date: %w", err)
			}
			createdAt, err := parsePurchaseTime(row.CreatedAt)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to parse purchase created_at: %w", err)
			}
			item := PurchaseListItem{ID: row.ID, InvoiceNumber: row.InvoiceNumber, PurchaseDate: purchaseDate, TotalAmount: row.TotalAmount, PaidAmount: row.PaidAmount, Remaining: row.Remaining, Status: row.Status, SupplierName: row.SupplierName, TotalItems: row.TotalItems, CreatedAt: createdAt, TotalCount: row.TotalCount}
			if row.ExpectedDeliveryDate.Valid && row.ExpectedDeliveryDate.String != "" {
				t, err := parsePurchaseTime(row.ExpectedDeliveryDate.String)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to parse expected delivery date: %w", err)
				}
				item.ExpectedDeliveryDate = &t
			}
			summaries = append(summaries, item)
			if total == 0 {
				total = row.TotalCount
			}
		}
		return summaries, total, nil
	}

	query := `
		SELECT
			p.id AS id,
			p.invoice_number,
			p.purchase_date,
			p.expected_delivery_date,
			p.total_amount,
			p.paid_amount,
			p.total_amount - p.paid_amount AS remaining,
			p.status,
			COALESCE(s.name, '') AS supplier_name,
			COUNT(pi.id) AS total_items,
			p.created_at,
			COUNT(*) OVER() AS total_count
		FROM purchases p
		LEFT JOIN suppliers s ON s.id = p.supplier_id
		LEFT JOIN purchase_items pi ON pi.purchase_id = p.id
		WHERE 1=1
	`
	args := make([]interface{}, 0, 7)
	argCount := 0

	addFilter := func(condition string, value interface{}) {
		argCount++
		query += fmt.Sprintf(" AND %s $%d", condition, argCount)
		args = append(args, value)
	}
	if req.SupplierID != nil {
		addFilter("p.supplier_id =", *req.SupplierID)
	}
	if req.Status != "" {
		addFilter("p.status =", req.Status)
	}
	if req.StartDate != nil {
		addFilter("p.purchase_date >=", *req.StartDate)
	}
	if req.EndDate != nil {
		addFilter("p.purchase_date <=", *req.EndDate)
	}
	if req.Search != "" {
		argCount++
		query += fmt.Sprintf(" AND (LOWER(p.invoice_number) LIKE LOWER($%d) OR LOWER(p.notes) LIKE LOWER($%d))", argCount, argCount)
		args = append(args, "%"+req.Search+"%")
	}
	if req.AvailableForReturn {
		query += ` AND EXISTS (
			SELECT 1 FROM purchase_items return_pi
			JOIN inventory_items return_ii ON return_ii.product_id = return_pi.product_id
			WHERE return_pi.purchase_id = p.id AND return_ii.status = 'AVAILABLE'
		)`
	}

	query += `
		GROUP BY p.id, p.invoice_number, p.purchase_date, p.total_amount,
			p.paid_amount, p.expected_delivery_date, p.status, p.created_at, s.name
	`

	sortColumns := map[string]string{
		"purchase_date": "p.purchase_date",
		"created_at":    "p.created_at",
		"total_amount":  "p.total_amount",
		"status":        "p.status",
	}
	sortColumn := sortColumns[req.SortBy]
	if sortColumn == "" {
		sortColumn = "p.purchase_date"
	}
	sortOrder := "DESC"
	if strings.EqualFold(req.SortOrder, "asc") {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s, p.id DESC", sortColumn, sortOrder)

	argCount++
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.PerPage, (req.Page-1)*req.PerPage)

	var summaries []PurchaseListItem
	if err := r.db.SelectContext(ctx, &summaries, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list purchase summaries: %w", err)
	}
	total := 0
	if len(summaries) > 0 {
		total = summaries[0].TotalCount
	}
	return summaries, total, nil
}

// Update updates a purchase
func (r *Repository) Update(ctx context.Context, purchase *Purchase) error {
	if dbutil.IsSQLite(r.db) {
		result, err := r.db.ExecContext(ctx, `UPDATE purchases SET purchase_number = ?, status = ?, notes = ?, updated_at = ? WHERE id = ?`, purchase.InvoiceNumber, purchase.Status, purchase.Notes, purchase.UpdatedAt, purchase.ID)
		if err != nil {
			return fmt.Errorf("failed to update local purchase: %w", err)
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			return ErrPurchaseNotFound
		}
		return nil
	}
	query := `
		UPDATE purchases
		SET invoice_number = $2, purchase_date = $3, status = $4, notes = $5, updated_at = $6
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		purchase.ID, purchase.InvoiceNumber, purchase.PurchaseDate, purchase.Status,
		purchase.Notes, purchase.UpdatedAt,
	).Scan(&purchase.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPurchaseNotFound
		}
		return fmt.Errorf("failed to update purchase: %w", err)
	}
	return nil
}

// Delete deletes a purchase
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM purchases WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete purchase: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrPurchaseNotFound
	}

	return nil
}

// CreatePurchaseItem creates a new purchase item
func (r *Repository) CreatePurchaseItem(ctx context.Context, item *PurchaseItem) error {
	if dbutil.IsSQLite(r.db) {
		return createPurchaseItemSQLite(ctx, r.db, item)
	}
	return createPurchaseItem(ctx, r.db, item)
}

// CreatePurchaseItemTx creates a purchase item using the caller's transaction.
func (r *Repository) CreatePurchaseItemTx(ctx context.Context, tx *sqlx.Tx, item *PurchaseItem) error {
	if dbutil.IsSQLite(r.db) {
		return createPurchaseItemSQLite(ctx, tx, item)
	}
	return createPurchaseItem(ctx, tx, item)
}

func createPurchaseItem(ctx context.Context, executor sqlx.ExtContext, item *PurchaseItem) error {
	query := `
		INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price,
			total_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err := executor.QueryRowxContext(ctx, query,
		item.ID, item.PurchaseID, item.ProductID, item.Quantity, item.UnitCost,
		item.TotalCost, item.CreatedAt,
	).Scan(&item.ID, &item.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create purchase item: %w", err)
	}
	return nil
}

func createPurchaseItemSQLite(ctx context.Context, executor sqlx.ExtContext, item *PurchaseItem) error {
	_, err := executor.ExecContext(ctx, `
		INSERT INTO purchase_items (id, purchase_id, product_id, quantity, unit_price, item_total, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, item.ID, item.PurchaseID, item.ProductID, item.Quantity, item.UnitCost, item.TotalCost, item.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create local purchase item: %w", err)
	}
	return nil
}

// GetPurchaseItems retrieves items for a purchase
func (r *Repository) GetPurchaseItems(ctx context.Context, purchaseID uuid.UUID) ([]PurchaseItem, error) {
	if dbutil.IsSQLite(r.db) {
		var rows []localPurchaseItemRow
		if err := r.db.SelectContext(ctx, &rows, `SELECT id, purchase_id, product_id, quantity, unit_price AS unit_cost, item_total AS total_cost, '' AS serial_number, '' AS condition, NULL AS location_id, '' AS notes, created_at FROM purchase_items WHERE purchase_id = $1 ORDER BY created_at`, purchaseID); err != nil {
			return nil, fmt.Errorf("failed to get purchase items: %w", err)
		}
		items := make([]PurchaseItem, 0, len(rows))
		for _, row := range rows {
			id, err := uuid.Parse(row.ID)
			if err != nil {
				return nil, err
			}
			pID, err := uuid.Parse(row.PurchaseID)
			if err != nil {
				return nil, err
			}
			productID, err := uuid.Parse(row.ProductID)
			if err != nil {
				return nil, err
			}
			createdAt, err := parsePurchaseTime(row.CreatedAt)
			if err != nil {
				return nil, err
			}
			item := PurchaseItem{ID: id, PurchaseID: pID, ProductID: productID, Quantity: row.Quantity, UnitCost: row.UnitCost, TotalCost: row.TotalCost, CreatedAt: createdAt, UpdatedAt: createdAt}
			if row.SerialNumber.Valid {
				item.SerialNumber = row.SerialNumber.String
			}
			if row.Condition.Valid {
				item.Condition = row.Condition.String
			}
			if row.LocationID.Valid && row.LocationID.String != "" {
				locationID, err := uuid.Parse(row.LocationID.String)
				if err == nil {
					item.LocationID = &locationID
				}
			}
			if row.Notes.Valid {
				item.Notes = row.Notes.String
			}
			items = append(items, item)
		}
		return items, nil
	}
	var items []PurchaseItem
	query := `
		SELECT id, purchase_id, product_id, quantity, unit_price as unit_cost, total_amount as total_cost,
			created_at
		FROM purchase_items
		WHERE purchase_id = $1
		ORDER BY created_at
	`

	err := r.db.SelectContext(ctx, &items, query, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase items: %w", err)
	}
	return items, nil
}

// UpdatePurchaseItem updates a purchase item
func (r *Repository) UpdatePurchaseItem(ctx context.Context, item *PurchaseItem) error {
	if dbutil.IsSQLite(r.db) {
		result, err := r.db.ExecContext(ctx, `UPDATE purchase_items SET quantity = ?, unit_price = ?, item_total = ? WHERE id = ?`, item.Quantity, item.UnitCost, item.TotalCost, item.ID.String())
		if err != nil {
			return fmt.Errorf("failed to update local purchase item: %w", err)
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			return ErrPurchaseItemNotFound
		}
		return nil
	}
	query := `
		UPDATE purchase_items
		SET quantity = $2, unit_cost = $3, total_cost = $4, serial_number = $5, 
			condition = $6, location_id = $7, notes = $8, updated_at = $9
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		item.ID, item.Quantity, item.UnitCost, item.TotalCost, item.SerialNumber,
		item.Condition, item.LocationID, item.Notes, item.UpdatedAt,
	).Scan(&item.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPurchaseItemNotFound
		}
		return fmt.Errorf("failed to update purchase item: %w", err)
	}
	return nil
}

// DeletePurchaseItem deletes a purchase item
func (r *Repository) DeletePurchaseItem(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM purchase_items WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete purchase item: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrPurchaseItemNotFound
	}

	return nil
}

// GetSupplierInfo retrieves supplier information
func (r *Repository) GetSupplierInfo(ctx context.Context, supplierID uuid.UUID) (*SupplierInfo, error) {
	if dbutil.IsSQLite(r.db) {
		var row struct {
			ID    string         `db:"id"`
			Name  string         `db:"name"`
			Phone sql.NullString `db:"phone"`
		}
		if err := r.db.GetContext(ctx, &row, `SELECT id, name, phone FROM suppliers WHERE id = ?`, supplierID.String()); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrSupplierNotFound
			}
			return nil, fmt.Errorf("failed to get supplier info: %w", err)
		}
		id, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse supplier id: %w", err)
		}
		var phone *string
		if row.Phone.Valid {
			phone = &row.Phone.String
		}
		return &SupplierInfo{ID: id, Name: row.Name, Phone: phone}, nil
	}
	var supplier SupplierInfo
	query := `SELECT id, name, phone FROM suppliers WHERE id = $1`

	err := r.db.GetContext(ctx, &supplier, query, supplierID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier info: %w", err)
	}
	return &supplier, nil
}

// UpdatePaidAmount updates the paid amount for a purchase
func (r *Repository) UpdatePaidAmount(ctx context.Context, purchaseID uuid.UUID, amount float64) error {
	if dbutil.IsSQLite(r.db) {
		result, err := r.db.ExecContext(ctx, `UPDATE purchases SET paid_amount = paid_amount + ?, remaining_amount = CASE WHEN total_amount - (paid_amount + ?) > 0 THEN total_amount - (paid_amount + ?) ELSE 0 END, updated_at = ? WHERE id = ?`, amount, amount, amount, time.Now().UTC().Format(time.RFC3339Nano), purchaseID.String())
		if err != nil {
			return fmt.Errorf("failed to update local paid amount: %w", err)
		}
		count, _ := result.RowsAffected()
		if count == 0 {
			return ErrPurchaseNotFound
		}
		return nil
	}
	query := `
		UPDATE purchases
		SET paid_amount = paid_amount + $2, updated_at = $3
		WHERE id = $1
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query, purchaseID, amount, time.Now()).Scan(&updatedAt)
	if err != nil {
		return fmt.Errorf("failed to update paid amount: %w", err)
	}
	return nil
}

// GetPurchaseByInvoiceNumber retrieves a purchase by invoice number
func (r *Repository) GetPurchaseByInvoiceNumber(ctx context.Context, invoiceNumber string) (*Purchase, error) {
	if dbutil.IsSQLite(r.db) {
		var row localPurchaseRow
		if err := r.db.GetContext(ctx, &row, `SELECT id, supplier_id, purchase_number AS invoice_number, created_at AS purchase_date, NULL AS expected_delivery_date, total_amount, paid_amount, status, notes, NULL AS user_id, created_at, updated_at FROM purchases WHERE purchase_number = $1`, invoiceNumber); err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrPurchaseNotFound
			}
			return nil, fmt.Errorf("failed to get local purchase by invoice number: %w", err)
		}
		purchaseID, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, err
		}
		supplierID, err := uuid.Parse(row.SupplierID)
		if err != nil {
			return nil, err
		}
		purchaseDate, err := parsePurchaseTime(row.PurchaseDate)
		if err != nil {
			return nil, err
		}
		createdAt, err := parsePurchaseTime(row.CreatedAt)
		if err != nil {
			return nil, err
		}
		updatedAt, err := parsePurchaseTime(row.UpdatedAt)
		if err != nil {
			return nil, err
		}
		return &Purchase{ID: purchaseID, SupplierID: supplierID, InvoiceNumber: row.InvoiceNumber, PurchaseDate: purchaseDate, TotalAmount: row.TotalAmount, PaidAmount: row.PaidAmount, Status: row.Status, CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
	}
	var purchase Purchase
	query := `
		SELECT id, supplier_id, invoice_number, purchase_date,
			total_amount, paid_amount, status, notes, user_id, created_at, updated_at
		FROM purchases
		WHERE invoice_number = $1
	`

	err := r.db.GetContext(ctx, &purchase, query, invoiceNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPurchaseNotFound
		}
		return nil, fmt.Errorf("failed to get purchase by invoice number: %w", err)
	}
	return &purchase, nil
}
