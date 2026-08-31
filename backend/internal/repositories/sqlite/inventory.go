package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// InventoryItem represents an inventory item entity
type InventoryItem struct {
	ID           string    `db:"id"`
	ProductID    string    `db:"product_id"`
	ItemCode     string    `db:"item_code"`
	Barcode      *string   `db:"barcode"`
	SerialNumber *string   `db:"serial_number"`
	Condition    *string   `db:"condition"`
	PurchaseCost float64   `db:"purchase_cost"`
	SellingPrice float64   `db:"selling_price"`
	Status       string    `db:"status"`
	SupplierID   *string   `db:"supplier_id"`
	PurchaseDate *string   `db:"purchase_date"`
	SoldAt       *string   `db:"sold_at"`
	Notes        *string   `db:"notes"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// InventoryRepository implements repositories.InventoryRepository for SQLite
type InventoryRepository struct {
	db *sql.DB
}

func scanInventoryItem(row interface{ Scan(...any) error }) (InventoryItem, error) {
	var item InventoryItem
	var barcode, serialNumber, condition, supplierID, purchaseDate, soldAt, notes sql.NullString
	var createdAt, updatedAt string
	if err := row.Scan(&item.ID, &item.ProductID, &item.ItemCode, &barcode, &serialNumber, &condition, &item.PurchaseCost, &item.SellingPrice, &item.Status, &supplierID, &purchaseDate, &soldAt, &notes, &createdAt, &updatedAt); err != nil {
		return InventoryItem{}, err
	}
	if barcode.Valid {
		item.Barcode = &barcode.String
	}
	if serialNumber.Valid {
		item.SerialNumber = &serialNumber.String
	}
	if condition.Valid {
		item.Condition = &condition.String
	}
	if supplierID.Valid {
		item.SupplierID = &supplierID.String
	}
	if purchaseDate.Valid {
		item.PurchaseDate = &purchaseDate.String
	}
	if soldAt.Valid {
		item.SoldAt = &soldAt.String
	}
	if notes.Valid {
		item.Notes = &notes.String
	}
	var err error
	item.CreatedAt, err = parseLocalTime(createdAt)
	if err != nil {
		return InventoryItem{}, err
	}
	item.UpdatedAt, err = parseLocalTime(updatedAt)
	if err != nil {
		return InventoryItem{}, err
	}
	return item, nil
}

// NewInventoryRepository creates a new SQLite inventory repository
func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// Create inserts a new inventory item
func (r *InventoryRepository) Create(ctx context.Context, item interface{}) (string, error) {
	i, ok := item.(*InventoryItem)
	if !ok {
		return "", fmt.Errorf("invalid inventory item type")
	}

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO inventory_items (id, product_id, item_code, barcode, serial_number, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, sold_at, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		id, i.ProductID, i.ItemCode, i.Barcode, i.SerialNumber, i.Condition, i.PurchaseCost, i.SellingPrice, i.Status, i.SupplierID, i.PurchaseDate, i.SoldAt, i.Notes, now, now)

	if err != nil {
		return "", fmt.Errorf("insert inventory item: %w", err)
	}

	return id, nil
}

// GetByID retrieves an inventory item by ID
func (r *InventoryRepository) GetByID(ctx context.Context, id string) (interface{}, error) {
	query := `
		SELECT id, product_id, item_code, barcode, serial_number, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, sold_at, notes, created_at, updated_at
		FROM inventory_items
		WHERE id = ?
	`

	item, err := scanInventoryItem(r.db.QueryRowContext(ctx, query, id))

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query inventory item: %w", err)
	}

	return &item, nil
}

// GetByBarcode retrieves an inventory item by barcode
func (r *InventoryRepository) GetByBarcode(ctx context.Context, barcode string) (interface{}, error) {
	query := `
		SELECT id, product_id, item_code, barcode, serial_number, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, sold_at, notes, created_at, updated_at
		FROM inventory_items
		WHERE barcode = ?
	`

	var item InventoryItem
	err := r.db.QueryRowContext(ctx, query, barcode).Scan(
		&item.ID, &item.ProductID, &item.ItemCode, &item.Barcode, &item.SerialNumber, &item.Condition,
		&item.PurchaseCost, &item.SellingPrice, &item.Status, &item.SupplierID, &item.PurchaseDate, &item.SoldAt, &item.Notes, &item.CreatedAt, &item.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query inventory item by barcode: %w", err)
	}

	return &item, nil
}

// List retrieves all inventory items
func (r *InventoryRepository) List(ctx context.Context, limit, offset int) ([]interface{}, error) {
	query := `
		SELECT id, product_id, item_code, barcode, serial_number, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, sold_at, notes, created_at, updated_at
		FROM inventory_items
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query inventory items: %w", err)
	}
	defer rows.Close()

	var items []interface{}
	for rows.Next() {
		item, err := scanInventoryItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan inventory item: %w", err)
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

// ListByProduct retrieves inventory items for a specific product
func (r *InventoryRepository) ListByProduct(ctx context.Context, productID string, limit, offset int) ([]interface{}, error) {
	query := `
		SELECT id, product_id, item_code, barcode, serial_number, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, sold_at, notes, created_at, updated_at
		FROM inventory_items
		WHERE product_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, productID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query inventory items by product: %w", err)
	}
	defer rows.Close()

	var items []interface{}
	for rows.Next() {
		item, err := scanInventoryItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan inventory item: %w", err)
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

// Update modifies an existing inventory item
func (r *InventoryRepository) Update(ctx context.Context, id string, item interface{}) error {
	i, ok := item.(*InventoryItem)
	if !ok {
		return fmt.Errorf("invalid inventory item type")
	}

	query := `
		UPDATE inventory_items
		SET product_id = ?, item_code = ?, barcode = ?, serial_number = ?, condition = ?, purchase_cost = ?, selling_price = ?, status = ?, supplier_id = ?, purchase_date = ?, sold_at = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		i.ProductID, i.ItemCode, i.Barcode, i.SerialNumber, i.Condition, i.PurchaseCost, i.SellingPrice, i.Status, i.SupplierID, i.PurchaseDate, i.SoldAt, i.Notes, time.Now(), id)

	if err != nil {
		return fmt.Errorf("update inventory item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("inventory item not found")
	}

	return nil
}

// Delete removes an inventory item
func (r *InventoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM inventory_items WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete inventory item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("inventory item not found")
	}

	return nil
}

// Search searches inventory items by item code, barcode, or serial number
func (r *InventoryRepository) Search(ctx context.Context, query string, limit, offset int) ([]interface{}, error) {
	sql := `
		SELECT id, product_id, item_code, barcode, serial_number, condition, purchase_cost, selling_price, status, supplier_id, purchase_date, sold_at, notes, created_at, updated_at
		FROM inventory_items
		WHERE item_code LIKE ? OR barcode LIKE ? OR serial_number LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sql, searchTerm, searchTerm, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search inventory items: %w", err)
	}
	defer rows.Close()

	var items []interface{}
	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.ItemCode, &item.Barcode, &item.SerialNumber, &item.Condition,
			&item.PurchaseCost, &item.SellingPrice, &item.Status, &item.SupplierID, &item.PurchaseDate, &item.SoldAt, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan inventory item: %w", err)
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

// Count returns total number of inventory items
func (r *InventoryRepository) Count(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM inventory_items"
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count inventory items: %w", err)
	}
	return count, nil
}
