package barcodes

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

func (r *Repository) ResolveBarcode(ctx context.Context, code string) (*BarcodeResolution, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("barcode is required")
	}

	resolution := &BarcodeResolution{Code: code, SaleIDs: []uuid.UUID{}, PurchaseIDs: []uuid.UUID{}, ReturnIDs: []uuid.UUID{}}
	var item struct {
		ID           string  `db:"id"`
		ProductID    string  `db:"product_id"`
		Barcode      string  `db:"barcode"`
		SerialNumber *string `db:"serial_number"`
		Condition    string  `db:"condition"`
		SupplierID   *string `db:"supplier_id"`
		PurchaseDate *string `db:"purchase_date"`
		PurchaseCost float64 `db:"purchase_cost"`
		SellingPrice float64 `db:"selling_price"`
		Status       string  `db:"status"`
	}
	itemQuery := r.db.Rebind(`SELECT id, product_id, barcode, serial_number, condition, status FROM inventory_items WHERE barcode = ? LIMIT 1`)
	if err := r.db.GetContext(ctx, &item, itemQuery, code); err == nil {
		itemID, parseItemErr := uuid.Parse(item.ID)
		productID, parseProductErr := uuid.Parse(item.ProductID)
		if parseItemErr != nil || parseProductErr != nil {
			return nil, fmt.Errorf("invalid barcode identity: item_id=%q product_id=%q item_error=%v product_error=%v", item.ID, item.ProductID, parseItemErr, parseProductErr)
		}
		resolved := &InventoryItemInfo{ID: itemID, ProductID: productID, Barcode: item.Barcode, SerialNumber: item.SerialNumber, Condition: item.Condition, PurchaseCost: item.PurchaseCost, SellingPrice: item.SellingPrice, Status: item.Status}
		if item.SupplierID != nil && strings.TrimSpace(*item.SupplierID) != "" {
			if supplierID, parseErr := uuid.Parse(*item.SupplierID); parseErr == nil {
				resolved.SupplierID = &supplierID
			}
		}
		if item.PurchaseDate != nil && strings.TrimSpace(*item.PurchaseDate) != "" {
			if purchaseDate, parseErr := dbutil.ParseTimestamp(*item.PurchaseDate); parseErr == nil {
				resolved.PurchaseDate = &purchaseDate
			}
		}
		resolution.InventoryItem = resolved
		product, productErr := r.GetProductByBarcode(ctx, code)
		if productErr == nil {
			resolution.Product = product
		}
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("resolve inventory item by barcode: %w", err)
	} else {
		product, productErr := r.GetProductByBarcode(ctx, code)
		if productErr == nil {
			resolution.Product = product
		}
	}

	if resolution.Product == nil && resolution.InventoryItem == nil {
		var productID string
		if err := r.db.GetContext(ctx, &productID, r.db.Rebind(`SELECT product_id FROM purchase_items WHERE barcode = ? LIMIT 1`), code); err == nil {
			if parsed, parseErr := uuid.Parse(productID); parseErr == nil {
				if product, productErr := r.getProductByID(ctx, parsed); productErr == nil {
					resolution.Product = product
				}
			}
		}
		if resolution.Product == nil {
			if err := r.db.GetContext(ctx, &productID, r.db.Rebind(`SELECT product_id FROM return_items WHERE barcode = ? LIMIT 1`), code); err == nil {
				if parsed, parseErr := uuid.Parse(productID); parseErr == nil {
					if product, productErr := r.getProductByID(ctx, parsed); productErr == nil {
						resolution.Product = product
					}
				}
			}
		}
	}
	if resolution.InventoryItem == nil {
		if item, err := r.resolveInventoryItemFromLifecycle(ctx, code); err == nil && item != nil {
			resolution.InventoryItem = item
		}
	}
	if resolution.Product == nil && resolution.InventoryItem != nil {
		if product, err := r.getProductByID(ctx, resolution.InventoryItem.ProductID); err == nil {
			resolution.Product = product
		}
	}

	productID := uuid.Nil
	if resolution.InventoryItem != nil {
		productID = resolution.InventoryItem.ProductID
	} else if resolution.Product != nil {
		productID = resolution.Product.ID
	}
	if resolution.Product == nil && resolution.InventoryItem == nil {
		return nil, fmt.Errorf("barcode not found")
	}

	resolution.SaleIDs = r.resolveIDs(ctx, `SELECT DISTINCT s.id FROM sales s JOIN sale_items si ON si.sale_id = s.id WHERE si.inventory_item_id = ? OR si.product_id = ? OR si.id IN (SELECT id FROM sale_items WHERE inventory_item_id = ? AND EXISTS (SELECT 1 FROM inventory_items ii WHERE ii.id = sale_items.inventory_item_id AND ii.barcode = ?))`, resolution.InventoryItemID(), productID, resolution.InventoryItemID(), code)
	resolution.PurchaseIDs = r.resolveIDs(ctx, `SELECT DISTINCT p.id FROM purchases p JOIN purchase_items pi ON pi.purchase_id = p.id WHERE pi.product_id = ? OR pi.barcode = ?`, productID, code)
	resolution.ReturnIDs = r.resolveIDs(ctx, `SELECT DISTINCT r.id FROM returns r JOIN return_items ri ON ri.return_id = r.id WHERE ri.inventory_item_id = ? OR ri.product_id = ? OR ri.barcode = ?`, resolution.InventoryItemID(), productID, code)
	if len(resolution.ReturnIDs) == 0 {
		resolution.ReturnIDs = r.resolveIDs(ctx, `SELECT DISTINCT r.id FROM returns r JOIN return_items ri ON ri.return_id = r.id WHERE ri.product_id = ? OR ri.barcode = ?`, productID, code)
	}
	return resolution, nil
}

func (r *Repository) resolveInventoryItemFromLifecycle(ctx context.Context, code string) (*InventoryItemInfo, error) {
	var itemID string
	if err := r.db.GetContext(ctx, &itemID, r.db.Rebind(`SELECT inventory_item_id FROM return_items WHERE barcode = ? AND inventory_item_id IS NOT NULL LIMIT 1`), code); err == nil && itemID != "" {
		return r.loadInventoryItemInfoByID(ctx, itemID)
	}
	var productID string
	if err := r.db.GetContext(ctx, &productID, r.db.Rebind(`SELECT product_id FROM purchase_items WHERE barcode = ? LIMIT 1`), code); err != nil {
		_ = r.db.GetContext(ctx, &productID, r.db.Rebind(`SELECT product_id FROM return_items WHERE barcode = ? LIMIT 1`), code)
	}
	if productID == "" {
		return nil, sql.ErrNoRows
	}
	if _, err := r.getProductByID(ctx, mustUUID(productID)); err != nil {
		return nil, err
	}
	var itemRow struct {
		ID string `db:"id"`
	}
	if err := r.db.GetContext(ctx, &itemRow, r.db.Rebind(`SELECT id FROM inventory_items WHERE product_id = ? ORDER BY created_at LIMIT 1`), productID); err == nil && itemRow.ID != "" {
		return r.loadInventoryItemInfoByID(ctx, itemRow.ID)
	}
	return nil, sql.ErrNoRows
}

func (r *Repository) loadInventoryItemInfoByID(ctx context.Context, itemID string) (*InventoryItemInfo, error) {
	var item struct {
		ID           string  `db:"id"`
		ProductID    string  `db:"product_id"`
		Barcode      string  `db:"barcode"`
		SerialNumber *string `db:"serial_number"`
		Condition    string  `db:"condition"`
		SupplierID   *string `db:"supplier_id"`
		PurchaseDate *string `db:"purchase_date"`
		PurchaseCost float64 `db:"purchase_cost"`
		SellingPrice float64 `db:"selling_price"`
		Status       string  `db:"status"`
	}
	if err := r.db.GetContext(ctx, &item, r.db.Rebind(`SELECT id, product_id, barcode, serial_number, condition, supplier_id, purchase_date, purchase_cost, selling_price, status FROM inventory_items WHERE id = ? LIMIT 1`), itemID); err != nil {
		return nil, err
	}
	parsedItemID, err := uuid.Parse(item.ID)
	if err != nil {
		return nil, err
	}
	parsedProductID, err := uuid.Parse(item.ProductID)
	if err != nil {
		return nil, err
	}
	resolved := &InventoryItemInfo{ID: parsedItemID, ProductID: parsedProductID, Barcode: item.Barcode, SerialNumber: item.SerialNumber, Condition: item.Condition, PurchaseCost: item.PurchaseCost, SellingPrice: item.SellingPrice, Status: item.Status}
	if item.SupplierID != nil && strings.TrimSpace(*item.SupplierID) != "" {
		if supplierID, parseErr := uuid.Parse(*item.SupplierID); parseErr == nil {
			resolved.SupplierID = &supplierID
		}
	}
	if item.PurchaseDate != nil && strings.TrimSpace(*item.PurchaseDate) != "" {
		if purchaseDate, parseErr := dbutil.ParseTimestamp(*item.PurchaseDate); parseErr == nil {
			resolved.PurchaseDate = &purchaseDate
		}
	}
	return resolved, nil
}

func (r *Repository) getProductByID(ctx context.Context, productID uuid.UUID) (*ProductInfo, error) {
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
	if err := r.db.GetContext(ctx, &row, r.db.Rebind(`SELECT p.id,p.name,p.sku,COALESCE(p.barcode,'') AS barcode,p.selling_price,p.cost_price,COALESCE((SELECT COUNT(*) FROM inventory_items ii WHERE ii.product_id=p.id AND UPPER(ii.status)='AVAILABLE'),0) AS stock,'' AS condition,COALESCE(c.name,'') AS category FROM products p LEFT JOIN categories c ON c.id=p.category_id WHERE p.id = ? LIMIT 1`), productID.String()); err != nil {
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}
	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("parse product id: %w", err)
	}
	return &ProductInfo{ID: id, Name: row.Name, SKU: row.SKU, Barcode: row.Barcode, SellingPrice: row.SellingPrice, CostPrice: row.CostPrice, Stock: row.Stock, Condition: row.Condition, Category: row.Category}, nil
}

func mustUUID(value string) uuid.UUID {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func (r *Repository) resolveIDs(ctx context.Context, query string, args ...interface{}) []uuid.UUID {
	query = r.db.Rebind(query)
	var rawIDs []string
	if err := r.db.SelectContext(ctx, &rawIDs, query, args...); err != nil {
		return []uuid.UUID{}
	}
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		if id, err := uuid.Parse(rawID); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func (r *BarcodeResolution) InventoryItemID() uuid.UUID {
	if r.InventoryItem == nil {
		return uuid.Nil
	}
	return r.InventoryItem.ID
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
