package barcodes

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

type localBarcodeRow struct {
	ID              string `db:"id"`
	Code            string `db:"code"`
	ProductID       string `db:"product_id"`
	InventoryItemID string `db:"inventory_item_id"`
	Type            string `db:"type"`
	IsActive        bool   `db:"is_active"`
	GeneratedAt     string `db:"generated_at"`
	CreatedAt       string `db:"created_at"`
	UpdatedAt       string `db:"updated_at"`
}

func idArg(id *uuid.UUID) interface{} {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return id.String()
}
func (r localBarcodeRow) model() (*Barcode, error) {
	id, e := uuid.Parse(r.ID)
	if e != nil {
		return nil, e
	}
	b := &Barcode{ID: id, Code: r.Code, Type: BarcodeType(r.Type), IsActive: r.IsActive}
	if r.ProductID != "" {
		if v, e := uuid.Parse(r.ProductID); e == nil {
			b.ProductID = &v
		}
	}
	if r.InventoryItemID != "" {
		if v, e := uuid.Parse(r.InventoryItemID); e == nil {
			b.InventoryItemID = &v
		}
	}
	if r.GeneratedAt != "" {
		b.GeneratedAt, _ = dbutil.ParseTimestamp(r.GeneratedAt)
	}
	b.CreatedAt, _ = dbutil.ParseTimestamp(r.CreatedAt)
	b.UpdatedAt, _ = dbutil.ParseTimestamp(r.UpdatedAt)
	return b, nil
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetBarcodeByCode retrieves a barcode by its code
func (r *Repository) GetBarcodeByCode(ctx context.Context, code string) (*Barcode, error) {
	if dbutil.IsSQLite(r.db) {
		var row localBarcodeRow
		if err := r.db.GetContext(ctx, &row, `SELECT id,code,COALESCE(product_id,'') AS product_id,COALESCE(inventory_item_id,'') AS inventory_item_id,type,is_active,COALESCE(generated_at,'') AS generated_at,created_at,updated_at FROM barcodes WHERE code = ? AND is_active = 1`, code); err != nil {
			return nil, fmt.Errorf("failed to get barcode: %w", err)
		}
		return row.model()
	}
	query := `
		SELECT id, code, product_id, type, generated_at, created_at, updated_at
		FROM barcodes
		WHERE code = $1
	`

	var barcode Barcode
	err := r.db.GetContext(ctx, &barcode, query, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode: %w", err)
	}

	return &barcode, nil
}

// GetProductByBarcode retrieves product information by barcode code
func (r *Repository) GetProductByBarcode(ctx context.Context, code string) (*ProductInfo, error) {
	if dbutil.IsSQLite(r.db) {
		var row struct {
			ID           string  `db:"id"`
			Name         string  `db:"name"`
			SKU          string  `db:"sku"`
			Barcode      string  `db:"barcode"`
			SellingPrice float64 `db:"selling_price"`
			CostPrice    float64 `db:"cost_price"`
			Stock        int     `db:"stock"`
			Condition    string  `db:"condition"`
			Category     string  `db:"category"`
		}
		if err := r.db.GetContext(ctx, &row, `SELECT p.id,p.name,p.sku,COALESCE(p.barcode,'') AS barcode,p.selling_price,p.cost_price,COALESCE((SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id=p.id AND UPPER(ii.status)='AVAILABLE'),0) AS stock,'' AS condition,COALESCE(c.name,'') AS category FROM products p LEFT JOIN categories c ON c.id=p.category_id WHERE p.barcode = ? OR p.sku = ? LIMIT 1`, code, code); err != nil {
			return nil, fmt.Errorf("failed to get product by barcode: %w", err)
		}
		id, _ := uuid.Parse(row.ID)
		return &ProductInfo{ID: id, Name: row.Name, SKU: row.SKU, Barcode: row.Barcode, SellingPrice: row.SellingPrice, CostPrice: row.CostPrice, Stock: row.Stock, Condition: row.Condition, Category: row.Category}, nil
	}
	query := `
		SELECT id, name, sku, barcode, selling_price, cost_price, stock, condition, category
		FROM products
		WHERE barcode = $1 OR sku = $1
		LIMIT 1
	`

	var product ProductInfo
	err := r.db.GetContext(ctx, &product, query, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by barcode: %w", err)
	}

	return &product, nil
}

// GetProductBySKU retrieves product information by SKU
func (r *Repository) GetProductBySKU(ctx context.Context, sku string) (*ProductInfo, error) {
	if dbutil.IsSQLite(r.db) {
		var row struct {
			ID           string  `db:"id"`
			Name         string  `db:"name"`
			SKU          string  `db:"sku"`
			Barcode      string  `db:"barcode"`
			SellingPrice float64 `db:"selling_price"`
			CostPrice    float64 `db:"cost_price"`
			Stock        int     `db:"stock"`
			Condition    string  `db:"condition"`
			Category     string  `db:"category"`
		}
		if err := r.db.GetContext(ctx, &row, `SELECT p.id,p.name,p.sku,COALESCE(p.barcode,'') AS barcode,p.selling_price,p.cost_price,COALESCE((SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id=p.id AND UPPER(ii.status)='AVAILABLE'),0) AS stock,'' AS condition,COALESCE(c.name,'') AS category FROM products p LEFT JOIN categories c ON c.id=p.category_id WHERE p.sku = ? LIMIT 1`, sku); err != nil {
			return nil, fmt.Errorf("failed to get product by SKU: %w", err)
		}
		id, _ := uuid.Parse(row.ID)
		return &ProductInfo{ID: id, Name: row.Name, SKU: row.SKU, Barcode: row.Barcode, SellingPrice: row.SellingPrice, CostPrice: row.CostPrice, Stock: row.Stock, Condition: row.Condition, Category: row.Category}, nil
	}
	query := `
		SELECT id, name, sku, barcode, selling_price, cost_price, stock, condition, category
		FROM products
		WHERE sku = $1
		LIMIT 1
	`

	var product ProductInfo
	err := r.db.GetContext(ctx, &product, query, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}

	return &product, nil
}

// GetBarcodeByID retrieves a barcode by its ID
func (r *Repository) GetBarcodeByID(ctx context.Context, id uuid.UUID) (*Barcode, error) {
	if dbutil.IsSQLite(r.db) {
		var row localBarcodeRow
		if err := r.db.GetContext(ctx, &row, `SELECT id,code,COALESCE(product_id,'') AS product_id,COALESCE(inventory_item_id,'') AS inventory_item_id,type,is_active,COALESCE(generated_at,'') AS generated_at,created_at,updated_at FROM barcodes WHERE id = ?`, id.String()); err != nil {
			return nil, fmt.Errorf("failed to get barcode: %w", err)
		}
		return row.model()
	}
	query := `
		SELECT id, code, product_id, type, generated_at, created_at, updated_at
		FROM barcodes
		WHERE id = $1
	`

	var barcode Barcode
	err := r.db.GetContext(ctx, &barcode, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode: %w", err)
	}

	return &barcode, nil
}

// CreateBarcode creates a new barcode
func (r *Repository) CreateBarcode(ctx context.Context, barcode *Barcode) error {
	if dbutil.IsSQLite(r.db) {
		if barcode.ID == uuid.Nil {
			barcode.ID = uuid.New()
		}
		now := time.Now().UTC()
		if barcode.CreatedAt.IsZero() {
			barcode.CreatedAt = now
		}
		if barcode.UpdatedAt.IsZero() {
			barcode.UpdatedAt = now
		}
		_, err := r.db.ExecContext(ctx, `INSERT INTO barcodes (id,code,product_id,inventory_item_id,type,is_active,generated_at,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?)`, barcode.ID.String(), barcode.Code, idArg(barcode.ProductID), idArg(barcode.InventoryItemID), string(barcode.Type), barcode.IsActive, barcode.GeneratedAt, barcode.CreatedAt.Format(time.RFC3339Nano), barcode.UpdatedAt.Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("failed to create barcode: %w", err)
		}
		return nil
	}
	query := `
		INSERT INTO barcodes (id, code, product_id, inventory_item_id, type, is_active, generated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		barcode.ID, barcode.Code, barcode.ProductID, barcode.InventoryItemID, barcode.Type, barcode.IsActive,
		barcode.GeneratedAt, barcode.CreatedAt, barcode.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create barcode: %w", err)
	}

	return nil
}

// ListBarcodes lists barcodes
func (r *Repository) ListBarcodes(ctx context.Context, limit, offset int) ([]*Barcode, int64, error) {
	if dbutil.IsSQLite(r.db) {
		var rows []localBarcodeRow
		if err := r.db.SelectContext(ctx, &rows, `SELECT id,code,COALESCE(product_id,'') AS product_id,COALESCE(inventory_item_id,'') AS inventory_item_id,type,is_active,COALESCE(generated_at,'') AS generated_at,created_at,updated_at FROM barcodes WHERE is_active = 1 ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset); err != nil {
			return nil, 0, fmt.Errorf("failed to list barcodes: %w", err)
		}
		result := make([]*Barcode, 0, len(rows))
		for _, row := range rows {
			b, e := row.model()
			if e != nil {
				return nil, 0, e
			}
			result = append(result, b)
		}
		var total int64
		if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM barcodes WHERE is_active = 1`); err != nil {
			return nil, 0, fmt.Errorf("failed to count barcodes: %w", err)
		}
		return result, total, nil
	}
	query := `
		SELECT id, code, product_id, inventory_item_id, type, is_active, generated_at, created_at, updated_at
		FROM barcodes
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var barcodes []*Barcode
	err := r.db.SelectContext(ctx, &barcodes, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list barcodes: %w", err)
	}

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM barcodes`
	err = r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count barcodes: %w", err)
	}

	return barcodes, total, nil
}

// DeleteBarcode soft deletes a barcode
func (r *Repository) DeleteBarcode(ctx context.Context, id uuid.UUID) error {
	if dbutil.IsSQLite(r.db) {
		if _, err := r.db.ExecContext(ctx, `UPDATE barcodes SET is_active = 0, updated_at = ? WHERE id = ?`, time.Now().UTC().Format(time.RFC3339Nano), id.String()); err != nil {
			return fmt.Errorf("failed to delete barcode: %w", err)
		}
		return nil
	}
	query := fmt.Sprintf(`
		UPDATE barcodes
		SET is_active = false, updated_at = %s
		WHERE id = $1
	`, dbutil.NowSQL(r.db))

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete barcode: %w", err)
	}

	return nil
}
