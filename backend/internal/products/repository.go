package products

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

// Repository handles products data operations
type Repository struct {
	db *sqlx.DB
}

// scanBrandRow keeps SQLite TEXT timestamps compatible with the time.Time
// fields exposed by the API. PostgreSQL and SQLite both safely scan their
// UUID/timestamp values into strings here, after which they are parsed once.
func scanBrandRow(scanner interface{ Scan(...any) error }) (Brand, error) {
	var brand Brand
	var idRaw, createdRaw, updatedRaw string
	if err := scanner.Scan(&idRaw, &brand.Name, &brand.Description, &brand.LogoURL, &createdRaw, &updatedRaw); err != nil {
		return Brand{}, err
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return Brand{}, fmt.Errorf("parse brand id: %w", err)
	}
	brand.ID = id
	brand.CreatedAt, err = parseSQLiteTimestamp(createdRaw)
	if err != nil {
		return Brand{}, fmt.Errorf("parse brand created_at: %w", err)
	}
	brand.UpdatedAt, err = parseSQLiteTimestamp(updatedRaw)
	if err != nil {
		return Brand{}, fmt.Errorf("parse brand updated_at: %w", err)
	}
	return brand, nil
}

// NewRepository creates a new products repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func parseSQLiteTimestamp(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, nil
	}
	if index := strings.Index(trimmed, " m="); index >= 0 {
		trimmed = trimmed[:index]
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05.999999999 -0700",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported SQLite timestamp format: %q", raw)
}

func scanReturnedTimestamps(scanner interface{ Scan(...any) error }, valueID *uuid.UUID, createdAt *time.Time, updatedAt *time.Time) error {
	var idRaw string
	var createdRaw, updatedRaw any
	if err := scanner.Scan(&idRaw, &createdRaw, &updatedRaw); err != nil {
		return err
	}

	parsedID, err := uuid.Parse(idRaw)
	if err != nil {
		return fmt.Errorf("parse returned id: %w", err)
	}
	*valueID = parsedID

	parsedCreatedAt, err := dbutil.ParseTimestamp(createdRaw)
	if err != nil {
		return fmt.Errorf("parse returned created_at: %w", err)
	}
	*createdAt = parsedCreatedAt

	parsedUpdatedAt, err := dbutil.ParseTimestamp(updatedRaw)
	if err != nil {
		return fmt.Errorf("parse returned updated_at: %w", err)
	}
	*updatedAt = parsedUpdatedAt
	return nil
}

func scanProductRow(rows *sql.Rows) (Product, error) {
	var product Product
	var description, model, barcode, deletedAt, createdAt, updatedAt sql.NullString
	var categoryID, brandID, preferredSupplierID sql.NullString

	if err := rows.Scan(
		&product.ID,
		&categoryID,
		&brandID,
		&preferredSupplierID,
		&product.Name,
		&description,
		&model,
		&product.SKU,
		&barcode,
		&product.CostPrice,
		&product.SellingPrice,
		&product.TrackSerial,
		&product.TrackIndividual,
		&product.MinStockLevel,
		&product.WarrantyDays,
		&product.IsActive,
		&deletedAt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Product{}, err
	}

	if categoryID.Valid {
		id, err := uuid.Parse(categoryID.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse category id: %w", err)
		}
		product.CategoryID = &id
	}
	if brandID.Valid {
		id, err := uuid.Parse(brandID.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse brand id: %w", err)
		}
		product.BrandID = &id
	}
	if preferredSupplierID.Valid {
		id, err := uuid.Parse(preferredSupplierID.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse preferred supplier id: %w", err)
		}
		product.PreferredSupplierID = &id
	}
	if description.Valid && strings.TrimSpace(description.String) != "" {
		v := description.String
		product.Description = &v
	}
	if model.Valid && strings.TrimSpace(model.String) != "" {
		v := model.String
		product.Model = &v
	}
	if barcode.Valid && strings.TrimSpace(barcode.String) != "" {
		product.Barcode = barcode.String
	}
	if deletedAt.Valid && strings.TrimSpace(deletedAt.String) != "" {
		parsed, err := parseSQLiteTimestamp(deletedAt.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse deleted_at: %w", err)
		}
		product.DeletedAt = &parsed
	}
	if createdAt.Valid && strings.TrimSpace(createdAt.String) != "" {
		parsed, err := parseSQLiteTimestamp(createdAt.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse created_at: %w", err)
		}
		product.CreatedAt = parsed
	}
	if updatedAt.Valid && strings.TrimSpace(updatedAt.String) != "" {
		parsed, err := parseSQLiteTimestamp(updatedAt.String)
		if err != nil {
			return Product{}, fmt.Errorf("parse updated_at: %w", err)
		}
		product.UpdatedAt = parsed
	}
	return product, nil
}

// Category CRUD operations

// CreateCategory creates a new category
func (r *Repository) CreateCategory(ctx context.Context, category *Category) error {
	query := `
		INSERT INTO categories (id, name, description, parent_id, icon, color, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	category.ID = uuid.New()
	category.CreatedAt = now
	category.UpdatedAt = now

	if dbutil.IsSQLite(r.db) {
		_, err := r.db.ExecContext(ctx, strings.Replace(query, "\n\t\tRETURNING id, created_at, updated_at", "", 1),
			category.ID,
			category.Name,
			category.Description,
			category.ParentID,
			category.Icon,
			category.Color,
			category.IsActive,
			category.CreatedAt,
			category.UpdatedAt,
		)
		return err
	}

	return scanReturnedTimestamps(r.db.QueryRowContext(ctx, query,
		category.ID,
		category.Name,
		category.Description,
		category.ParentID,
		category.Icon,
		category.Color,
		category.IsActive,
		category.CreatedAt,
		category.UpdatedAt,
	), &category.ID, &category.CreatedAt, &category.UpdatedAt)
}

// GetCategoryByID retrieves a category by ID
func (r *Repository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	query := `
		SELECT id, name, description, parent_id, icon, color, is_active, created_at, updated_at
		FROM categories
		WHERE id = $1
	`
	var category Category
	var description, parentID, icon, color, createdAt, updatedAt sql.NullString
	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &description, &parentID, &icon, &color,
		&category.IsActive, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	if description.Valid {
		category.Description = description.String
	}
	if parentID.Valid && strings.TrimSpace(parentID.String) != "" {
		parsed, parseErr := uuid.Parse(parentID.String)
		if parseErr != nil {
			return nil, fmt.Errorf("parse parent id: %w", parseErr)
		}
		category.ParentID = &parsed
	}
	if icon.Valid {
		category.Icon = &icon.String
	}
	if color.Valid {
		category.Color = &color.String
	}
	category.CreatedAt, err = parseSQLiteTimestamp(createdAt.String)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	category.UpdatedAt, err = parseSQLiteTimestamp(updatedAt.String)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	return &category, err
}

// ListCategories retrieves all categories
func (r *Repository) ListCategories(ctx context.Context) ([]Category, error) {
	query := `
		SELECT id, name, description, parent_id, icon, color, is_active, created_at, updated_at
		FROM categories
		ORDER BY name
	`
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var category Category
		var description, parentID, icon, color, createdAt, updatedAt sql.NullString
		if err := rows.Scan(
			&category.ID, &category.Name, &description, &parentID, &icon, &color,
			&category.IsActive, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		if description.Valid {
			category.Description = description.String
		}
		if parentID.Valid && strings.TrimSpace(parentID.String) != "" {
			parsed, parseErr := uuid.Parse(parentID.String)
			if parseErr != nil {
				return nil, fmt.Errorf("parse parent id: %w", parseErr)
			}
			category.ParentID = &parsed
		}
		if icon.Valid {
			category.Icon = &icon.String
		}
		if color.Valid {
			category.Color = &color.String
		}
		category.CreatedAt, err = parseSQLiteTimestamp(createdAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		category.UpdatedAt, err = parseSQLiteTimestamp(updatedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse updated_at: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

// UpdateCategory updates a category
func (r *Repository) UpdateCategory(ctx context.Context, category *Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2, parent_id = $3, icon = $4, color = $5, is_active = $6, updated_at = $7
		WHERE id = $8
	`
	category.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		category.Name,
		category.Description,
		category.ParentID,
		category.Icon,
		category.Color,
		category.IsActive,
		category.UpdatedAt,
		category.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// DeleteCategory deletes a category
func (r *Repository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	// First, set category_id to NULL for all products in this category
	updateQuery := `UPDATE products SET category_id = NULL WHERE category_id = $1`
	_, err := r.db.ExecContext(ctx, updateQuery, id)
	if err != nil {
		return err
	}

	// Now delete the category
	query := `DELETE FROM categories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// Brand CRUD operations

// CreateBrand creates a new brand
func (r *Repository) CreateBrand(ctx context.Context, brand *Brand) error {
	query := `
		INSERT INTO brands (id, name, description, logo_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	brand.ID = uuid.New()
	brand.CreatedAt = now
	brand.UpdatedAt = now

	if dbutil.IsSQLite(r.db) {
		_, err := r.db.ExecContext(ctx, strings.Replace(query, "\n\t\tRETURNING id, created_at, updated_at", "", 1),
			brand.ID,
			brand.Name,
			brand.Description,
			brand.LogoURL,
			brand.CreatedAt,
			brand.UpdatedAt,
		)
		return err
	}

	return scanReturnedTimestamps(r.db.QueryRowContext(ctx, query,
		brand.ID,
		brand.Name,
		brand.Description,
		brand.LogoURL,
		brand.CreatedAt,
		brand.UpdatedAt,
	), &brand.ID, &brand.CreatedAt, &brand.UpdatedAt)
}

// GetBrandByID retrieves a brand by ID
func (r *Repository) GetBrandByID(ctx context.Context, id uuid.UUID) (*Brand, error) {
	query := `
		SELECT id, name, COALESCE(description, '') AS description, COALESCE(logo_url, '') AS logo_url, created_at, updated_at
		FROM brands
		WHERE id = $1
	`
	row := r.db.QueryRowxContext(ctx, query, id)
	brand, err := scanBrandRow(row)
	if err == sql.ErrNoRows {
		return nil, ErrBrandNotFound
	}
	if err != nil {
		return nil, err
	}
	return &brand, nil
}

// ListBrands retrieves all brands
func (r *Repository) ListBrands(ctx context.Context) ([]Brand, error) {
	query := `
		SELECT id, name, COALESCE(description, '') AS description, COALESCE(logo_url, '') AS logo_url, created_at, updated_at
		FROM brands
		ORDER BY name
	`
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	brands := make([]Brand, 0)
	for rows.Next() {
		brand, scanErr := scanBrandRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		brands = append(brands, brand)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return brands, nil
}

// UpdateBrand updates a brand
func (r *Repository) UpdateBrand(ctx context.Context, brand *Brand) error {
	query := `
		UPDATE brands
		SET name = $1, description = $2, logo_url = $3, updated_at = $4
	`
	brand.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		brand.Name,
		brand.Description,
		brand.LogoURL,
		brand.UpdatedAt,
		brand.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrBrandNotFound
	}

	return nil
}

// DeleteBrand deletes a brand
func (r *Repository) DeleteBrand(ctx context.Context, id uuid.UUID) error {
	// Check if brand has products
	var count int
	checkQuery := `SELECT COUNT(*) FROM products WHERE brand_id = $1`
	err := r.db.GetContext(ctx, &count, checkQuery, id)
	if err != nil {
		return err
	}

	if count > 0 {
		return ErrBrandHasProducts
	}

	query := `DELETE FROM brands WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrBrandNotFound
	}

	return nil
}

// Product CRUD operations

// CreateProduct creates a new product
func (r *Repository) CreateProduct(ctx context.Context, product *Product) error {
	query := `
		INSERT INTO products (id, category_id, brand_id, preferred_supplier_id, name, description, model, sku, barcode, cost_price, selling_price, track_serial, track_individual, min_stock_level, warranty_days, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	product.ID = uuid.New()
	product.CreatedAt = now
	product.UpdatedAt = now
	product.IsActive = true
	if dbutil.IsSQLite(r.db) {
		_, err := r.db.ExecContext(ctx, strings.Replace(query, "\n\t\tRETURNING id, created_at, updated_at", "", 1),
			product.ID, product.CategoryID, product.BrandID, product.PreferredSupplierID,
			product.Name, product.Description, product.Model, product.SKU, product.Barcode,
			product.CostPrice, product.SellingPrice, product.TrackSerial, product.TrackIndividual,
			product.MinStockLevel, product.WarrantyDays, product.IsActive, product.CreatedAt, product.UpdatedAt)
		return err
	}

	return scanReturnedTimestamps(r.db.QueryRowContext(ctx, query,
		product.ID,
		product.CategoryID,
		product.BrandID,
		product.PreferredSupplierID,
		product.Name,
		product.Description,
		product.Model,
		product.SKU,
		product.Barcode,
		product.CostPrice,
		product.SellingPrice,
		product.TrackSerial,
		product.TrackIndividual,
		product.MinStockLevel,
		product.WarrantyDays,
		product.IsActive,
		product.CreatedAt,
		product.UpdatedAt,
	), &product.ID, &product.CreatedAt, &product.UpdatedAt)
}

// GetProductByID retrieves a product by ID
func (r *Repository) GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	query := `
		SELECT id, category_id, brand_id, preferred_supplier_id, name, description, model, sku, COALESCE(barcode, '') AS barcode, cost_price, selling_price, track_serial, track_individual, min_stock_level, warranty_days, is_active, deleted_at, created_at, updated_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`
	if dbutil.IsSQLite(r.db) {
		rows, err := r.db.QueryContext(ctx, query, id)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return nil, err
			}
			return nil, ErrProductNotFound
		}
		product, err := scanProductRow(rows)
		if err != nil {
			return nil, err
		}
		return &product, nil
	}
	var product Product
	err := r.db.GetContext(ctx, &product, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	return &product, err
}

// GetProductByBarcode retrieves a product by barcode
func (r *Repository) GetProductByBarcode(ctx context.Context, barcode string) (*Product, error) {
	if dbutil.IsSQLite(r.db) {
		var productID string
		err := r.db.GetContext(ctx, &productID,
			`SELECT id FROM products WHERE barcode = $1 AND deleted_at IS NULL`, barcode)
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		if err != nil {
			return nil, err
		}

		id, err := uuid.Parse(productID)
		if err != nil {
			return nil, fmt.Errorf("parse product id: %w", err)
		}
		return r.GetProductByID(ctx, id)
	}

	query := `
		SELECT id, category_id, brand_id, preferred_supplier_id, name, description, model, sku, barcode, cost_price, selling_price, track_serial, track_individual, min_stock_level, warranty_days, is_active, deleted_at, created_at, updated_at
		FROM products
		WHERE barcode = $1 AND deleted_at IS NULL
	`
	var product Product
	err := r.db.GetContext(ctx, &product, query, barcode)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	return &product, err
}

// ListProducts retrieves products with pagination and filters
func (r *Repository) ListProducts(ctx context.Context, req *ProductListRequest) ([]Product, int, error) {
	// Build base query with current quantity from inventory
	baseQuery := `
		SELECT p.id, p.category_id, p.brand_id, p.preferred_supplier_id, p.name, p.description, p.model, p.sku, p.barcode, p.cost_price, p.selling_price, p.track_serial, p.track_individual, p.min_stock_level, p.warranty_days, p.is_active, p.deleted_at, p.created_at, p.updated_at
		FROM products p
		WHERE p.deleted_at IS NULL
	`
	countQuery := `SELECT COUNT(DISTINCT p.id) FROM products p WHERE p.deleted_at IS NULL`

	args := []interface{}{}
	argCount := 0
	likeOperator := "ILIKE"
	if r.db.DriverName() == "sqlite" {
		likeOperator = "LIKE"
	}

	// Add filters
	if req.CategoryID != nil {
		argCount++
		paramNum := argCount
		baseQuery += ` AND category_id = $` + fmt.Sprint(paramNum)
		countQuery += ` AND category_id = $` + fmt.Sprint(paramNum)
		args = append(args, req.CategoryID)
	}

	if req.BrandID != nil {
		argCount++
		paramNum := argCount
		baseQuery += ` AND brand_id = $` + fmt.Sprint(paramNum)
		countQuery += ` AND brand_id = $` + fmt.Sprint(paramNum)
		args = append(args, req.BrandID)
	}

	if req.Search != "" {
		argCount++
		searchPattern := "%" + req.Search + "%"
		baseQuery += ` AND (name ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR model ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR sku ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR description ` + likeOperator + ` $` + fmt.Sprint(argCount) + `)`
		countQuery += ` AND (name ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR model ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR sku ` + likeOperator + ` $` + fmt.Sprint(argCount) + ` OR description ` + likeOperator + ` $` + fmt.Sprint(argCount) + `)`
		args = append(args, searchPattern)
	}

	if req.TrackSerial != nil {
		argCount++
		paramNum := argCount
		baseQuery += ` AND track_serial = $` + fmt.Sprint(paramNum)
		countQuery += ` AND track_serial = $` + fmt.Sprint(paramNum)
		args = append(args, req.TrackSerial)
	}

	if req.TrackIndividual != nil {
		argCount++
		paramNum := argCount
		baseQuery += ` AND track_individual = $` + fmt.Sprint(paramNum)
		countQuery += ` AND track_individual = $` + fmt.Sprint(paramNum)
		args = append(args, req.TrackIndividual)
	}

	// Advanced filtering: low stock only
	if req.LowStockOnly != nil && *req.LowStockOnly {
		baseQuery += ` AND (
			SELECT COALESCE(COUNT(*), 0)
			FROM inventory_items ii
			WHERE ii.product_id = p.id AND UPPER(COALESCE(ii.status, '')) = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED'
		) < p.min_stock_level
		AND p.min_stock_level > 0`
		countQuery += ` AND (
			SELECT COALESCE(COUNT(*), 0)
			FROM inventory_items ii
			WHERE ii.product_id = p.id AND UPPER(COALESCE(ii.status, '')) = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED'
		) < p.min_stock_level
		AND p.min_stock_level > 0`
	}

	// Advanced filtering: in stock only
	if req.InStockOnly != nil && *req.InStockOnly {
		baseQuery += ` AND (
			SELECT COALESCE(COUNT(*), 0)
			FROM inventory_items ii
			WHERE ii.product_id = p.id AND UPPER(COALESCE(ii.status, '')) = 'AVAILABLE'
		) > 0`
		countQuery += ` AND (
			SELECT COALESCE(COUNT(*), 0)
			FROM inventory_items ii
			WHERE ii.product_id = p.id AND UPPER(COALESCE(ii.status, '')) = 'AVAILABLE'
		) > 0`
	}

	if req.ManualOnly != nil && *req.ManualOnly {
		manualStock := `EXISTS (SELECT 1 FROM inventory_movements im WHERE im.product_id = p.id AND UPPER(COALESCE(im.source_type, '')) = 'OPENING_STOCK')`
		baseQuery += ` AND ` + manualStock
		countQuery += ` AND ` + manualStock
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Add sorting
	sortBy := "name"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}

	sortOrder := "ASC"
	if req.SortOrder == "DESC" {
		sortOrder = "DESC"
	}

	// Add pagination
	offset := (req.Page - 1) * req.PerPage
	paramNum := argCount + 1
	baseQuery += ` ORDER BY ` + sortBy + ` ` + sortOrder + ` LIMIT $` + fmt.Sprint(paramNum) + ` OFFSET $` + fmt.Sprint(paramNum+1)
	args = append(args, req.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]Product, 0, req.PerPage)
	for rows.Next() {
		product, err := scanProductRow(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// UpdateProduct updates a product
func (r *Repository) UpdateProduct(ctx context.Context, product *Product) error {
	query := `
		UPDATE products
		SET category_id = $1, brand_id = $2, preferred_supplier_id = $3, name = $4, description = $5, model = $6, sku = $7, barcode = $8, cost_price = $9, selling_price = $10, track_serial = $11, track_individual = $12, min_stock_level = $13, warranty_days = $14, updated_at = $15
		WHERE id = $16
	`
	product.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		product.CategoryID,
		product.BrandID,
		product.PreferredSupplierID,
		product.Name,
		product.Description,
		product.Model,
		product.SKU,
		product.Barcode,
		product.CostPrice,
		product.SellingPrice,
		product.TrackSerial,
		product.TrackIndividual,
		product.MinStockLevel,
		product.WarrantyDays,
		product.UpdatedAt,
		product.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (r *Repository) UpdateProductName(ctx context.Context, id uuid.UUID, name string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE products SET name = $1, updated_at = $2 WHERE id = $3`, name, time.Now(), id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *Repository) UpdateMinimumStock(ctx context.Context, id uuid.UUID, minStockLevel int) error {
	result, err := r.db.ExecContext(ctx, `UPDATE products SET min_stock_level = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`, minStockLevel, time.Now(), id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrProductNotFound
	}
	return nil
}

// DeleteProduct permanently deletes a product and all transaction history that
// is linked to it. The cleanup is intentionally done in one
// transaction: dashboard/report totals read the parent transaction tables, so
// deleting only sale_items or purchase_items would leave stale totals behind.
func (r *Repository) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var productExists bool
	if err := tx.GetContext(ctx, &productExists, tx.Rebind(`SELECT EXISTS (SELECT 1 FROM products WHERE id = ?)`), id); err != nil {
		return err
	}
	if !productExists {
		return ErrProductNotFound
	}

	// Capture parent and child IDs before deleting any rows. This lets us remove
	// dependent records in the correct order on schemas with foreign keys.
	saleIDs, err := collectProductDeleteIDs(ctx, tx, `SELECT DISTINCT sale_id FROM sale_items WHERE product_id = ? AND sale_id IS NOT NULL`, id)
	if err != nil {
		return fmt.Errorf("find product sales: %w", err)
	}
	purchaseIDs, err := collectProductDeleteIDs(ctx, tx, `SELECT DISTINCT purchase_id FROM purchase_items WHERE product_id = ? AND purchase_id IS NOT NULL`, id)
	if err != nil {
		return fmt.Errorf("find product purchases: %w", err)
	}
	saleItemIDs, err := collectProductDeleteIDs(ctx, tx, `SELECT id FROM sale_items WHERE product_id = ?`, id)
	if err != nil {
		return fmt.Errorf("find product sale items: %w", err)
	}
	purchaseItemIDs, err := collectProductDeleteIDs(ctx, tx, `SELECT id FROM purchase_items WHERE product_id = ?`, id)
	if err != nil {
		return fmt.Errorf("find product purchase items: %w", err)
	}
	inventoryItemIDs, err := collectProductDeleteIDs(ctx, tx, `SELECT id FROM inventory_items WHERE product_id = ?`, id)
	if err != nil {
		return fmt.Errorf("find product inventory items: %w", err)
	}
	returnIDs, err := collectProductDeleteIDs(ctx, tx, `SELECT DISTINCT return_id FROM return_items WHERE product_id = ? AND return_id IS NOT NULL`, id)
	if err != nil {
		return fmt.Errorf("find product returns: %w", err)
	}
	acquisitionIDs, err := collectProductDeleteIDs(ctx, tx, `SELECT DISTINCT acquisition_id FROM acquisition_items WHERE product_id = ? AND acquisition_id IS NOT NULL`, id)
	if err != nil {
		return fmt.Errorf("find product acquisitions: %w", err)
	}
	supplierReturnIDs, err := collectProductDeleteIDs(ctx, tx, `SELECT DISTINCT supplier_return_id FROM supplier_return_items WHERE product_id = ? AND supplier_return_id IS NOT NULL`, id)
	if err != nil {
		return fmt.Errorf("find product supplier returns: %w", err)
	}

	// A return can be linked through a sale/purchase or only through its item.
	returnIDs = appendUniqueProductDeleteIDs(returnIDs, collectProductDeleteIDsForIDs(ctx, tx, `SELECT id FROM returns WHERE sale_id IN (?)`, saleIDs)...)
	returnIDs = appendUniqueProductDeleteIDs(returnIDs, collectProductDeleteIDsForIDs(ctx, tx, `SELECT id FROM returns WHERE purchase_id IN (?)`, purchaseIDs)...)
	returnIDs = appendUniqueProductDeleteIDs(returnIDs, collectProductDeleteIDsForIDs(ctx, tx, `SELECT DISTINCT return_id FROM return_items WHERE sale_item_id IN (?)`, saleItemIDs)...)
	supplierReturnIDs = appendUniqueProductDeleteIDs(supplierReturnIDs, collectProductDeleteIDsForIDs(ctx, tx, `SELECT id FROM supplier_returns WHERE purchase_id IN (?) OR sale_id IN (?)`, purchaseIDs, saleIDs)...)
	supplierReturnIDs = appendUniqueProductDeleteIDs(supplierReturnIDs, collectProductDeleteIDsForIDs(ctx, tx, `SELECT DISTINCT supplier_return_id FROM supplier_return_items WHERE customer_return_id IN (?)`, returnIDs)...)

	paymentIDs, err := collectProductDeleteIDsForIDsChecked(ctx, tx, `SELECT id FROM payments WHERE sale_id IN (?) OR purchase_id IN (?)`, saleIDs, purchaseIDs)
	if err != nil {
		return fmt.Errorf("find product payments: %w", err)
	}
	paymentTransactionIDs, err := collectProductDeleteIDsForIDsChecked(ctx, tx, `SELECT id FROM payment_transactions WHERE sale_id IN (?) OR payment_id IN (?)`, saleIDs, paymentIDs)
	if err != nil {
		return fmt.Errorf("find product payment transactions: %w", err)
	}
	paymentTransactionIDs = appendUniqueProductDeleteIDs(paymentTransactionIDs, collectProductDeleteIDsForIDs(ctx, tx, `SELECT payment_transaction_id FROM return_payment_refunds WHERE return_id IN (?)`, returnIDs)...)

	// Remove return and payment dependents before their parent records. These
	// tables were introduced over time, so absent legacy tables are ignored.
	for _, cleanup := range []struct {
		query string
		args  []any
	}{
		{`DELETE FROM return_payment_refunds WHERE return_id IN (?) OR payment_transaction_id IN (?)`, []any{returnIDs, paymentTransactionIDs}},
		{`DELETE FROM return_refunds WHERE return_id IN (?)`, []any{returnIDs}},
		{`DELETE FROM return_inspection WHERE return_item_id IN (SELECT id FROM return_items WHERE return_id IN (?))`, []any{returnIDs}},
		{`DELETE FROM return_audit_log WHERE return_id IN (?) OR return_item_id IN (SELECT id FROM return_items WHERE return_id IN (?))`, []any{returnIDs, returnIDs}},
		{`DELETE FROM payment_refunds WHERE payment_transaction_id IN (?)`, []any{paymentTransactionIDs}},
		{`DELETE FROM payment_webhook_events WHERE payment_transaction_id IN (?)`, []any{paymentTransactionIDs}},
		{`DELETE FROM supplier_return_items WHERE supplier_return_id IN (?)`, []any{supplierReturnIDs}},
		{`DELETE FROM supplier_return_items WHERE product_id = ?`, []any{id}},
		{`DELETE FROM supplier_return_items WHERE purchase_item_id IN (?)`, []any{purchaseItemIDs}},
		{`DELETE FROM supplier_return_items WHERE sale_item_id IN (?)`, []any{saleItemIDs}},
		{`DELETE FROM return_items WHERE return_id IN (?)`, []any{returnIDs}},
		{`DELETE FROM return_items WHERE product_id = ?`, []any{id}},
		{`DELETE FROM return_items WHERE sale_item_id IN (?)`, []any{saleItemIDs}},
		{`DELETE FROM return_items WHERE inventory_item_id IN (?)`, []any{inventoryItemIDs}},
	} {
		if err := execProductCleanup(ctx, tx, cleanup.query, cleanup.args...); err != nil {
			return fmt.Errorf("clean product transaction details: %w", err)
		}
	}

	for _, cleanup := range []struct {
		table string
		col   string
		ids   []string
	}{
		{"supplier_returns", "id", supplierReturnIDs},
		{"returns", "id", returnIDs},
		{"payment_transactions", "id", paymentTransactionIDs},
		{"debts", "sale_id", saleIDs},
		{"financial_transactions", "sale_id", saleIDs},
		{"payments", "id", paymentIDs},
		{"sale_payment_allocations", "sale_id", saleIDs},
	} {
		if err := deleteProductRowsByIDs(ctx, tx, cleanup.table, cleanup.col, cleanup.ids); err != nil {
			return fmt.Errorf("clean product %s: %w", cleanup.table, err)
		}
	}

	// Delete the transaction line items before deleting their parent records.
	for _, cleanup := range []struct {
		query string
		args  []any
	}{
		{`DELETE FROM sale_items WHERE product_id = ? OR sale_id IN (?)`, []any{id, saleIDs}},
		{`DELETE FROM purchase_items WHERE product_id = ? OR purchase_id IN (?)`, []any{id, purchaseIDs}},
		{`DELETE FROM seller_payments WHERE acquisition_id IN (?)`, []any{acquisitionIDs}},
		{`DELETE FROM item_repair_costs WHERE inventory_item_id IN (?)`, []any{inventoryItemIDs}},
		{`DELETE FROM item_repair_costs WHERE acquisition_item_id IN (SELECT id FROM acquisition_items WHERE product_id = ?)`, []any{id}},
		{`DELETE FROM item_history WHERE inventory_item_id IN (?)`, []any{inventoryItemIDs}},
		{`DELETE FROM inspection_items WHERE inspection_id IN (SELECT id FROM inspections WHERE product_id = ?)`, []any{id}},
		{`DELETE FROM inspection_items WHERE inspection_id IN (SELECT id FROM inspections WHERE inventory_item_id IN (?))`, []any{inventoryItemIDs}},
		{`DELETE FROM inspection_items WHERE inspection_id IN (SELECT id FROM inspections WHERE acquisition_item_id IN (SELECT id FROM acquisition_items WHERE product_id = ?))`, []any{id}},
		{`DELETE FROM inspections WHERE product_id = ?`, []any{id}},
		{`DELETE FROM inspections WHERE inventory_item_id IN (?)`, []any{inventoryItemIDs}},
		{`DELETE FROM inspections WHERE acquisition_item_id IN (SELECT id FROM acquisition_items WHERE product_id = ?)`, []any{id}},
		{`DELETE FROM acquisition_items WHERE product_id = ? OR acquisition_id IN (?)`, []any{id, acquisitionIDs}},
	} {
		if err := execProductCleanup(ctx, tx, cleanup.query, cleanup.args...); err != nil {
			return fmt.Errorf("clean product transaction lines: %w", err)
		}
	}
	for _, cleanup := range []struct {
		table string
		col   string
		ids   []string
	}{
		{"sales", "id", saleIDs},
		{"purchases", "id", purchaseIDs},
		{"acquisitions", "id", acquisitionIDs},
	} {
		if err := deleteProductRowsByIDs(ctx, tx, cleanup.table, cleanup.col, cleanup.ids); err != nil {
			return fmt.Errorf("clean product %s: %w", cleanup.table, err)
		}
	}

	// Finally remove product-level operational/history rows and the product.
	for _, cleanup := range []struct {
		query string
		args  []any
	}{
		{`DELETE FROM inventory_movements WHERE product_id = ?`, []any{id}},
		{`DELETE FROM inventory_movements WHERE item_id IN (?)`, []any{inventoryItemIDs}},
		{`DELETE FROM reservations WHERE item_id IN (?)`, []any{inventoryItemIDs}},
		{`DELETE FROM item_specification_values WHERE inventory_item_id IN (?)`, []any{inventoryItemIDs}},
		{`DELETE FROM barcodes WHERE product_id = ?`, []any{id}},
		{`DELETE FROM barcodes WHERE inventory_item_id IN (?)`, []any{inventoryItemIDs}},
		{`DELETE FROM trade_ins WHERE inventory_item_id IN (?)`, []any{inventoryItemIDs}},
		{`DELETE FROM ledger_entries WHERE product_id = ? OR reference_id IN (?)`, []any{id, appendUniqueProductDeleteIDs(append(append([]string{}, saleIDs...), purchaseIDs...), acquisitionIDs...)}},
		{`DELETE FROM warranty_claims WHERE product_id = ?`, []any{id}},
		{`DELETE FROM inventory WHERE product_id = ?`, []any{id}},
		{`DELETE FROM inventory_items WHERE product_id = ?`, []any{id}},
	} {
		if err := execProductCleanup(ctx, tx, cleanup.query, cleanup.args...); err != nil {
			return fmt.Errorf("clean product history: %w", err)
		}
	}

	result, err := tx.ExecContext(ctx, tx.Rebind(`DELETE FROM products WHERE id = ?`), id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProductNotFound
	}

	return tx.Commit()
}

func collectProductDeleteIDs(ctx context.Context, tx *sqlx.Tx, query string, args ...any) ([]string, error) {
	rows, err := tx.QueryxContext(ctx, tx.Rebind(query), args...)
	if err != nil {
		if isOptionalProductCleanupError(err) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if strings.TrimSpace(id) != "" {
			ids = appendUniqueProductDeleteIDs(ids, id)
		}
	}
	return ids, rows.Err()
}

// collectProductDeleteIDsForIDs is best-effort for optional legacy tables and
// is used only after the primary product-linked IDs have been collected.
func collectProductDeleteIDsForIDs(ctx context.Context, tx *sqlx.Tx, query string, idLists ...[]string) []string {
	ids, _ := collectProductDeleteIDsForIDsChecked(ctx, tx, query, idLists...)
	return ids
}

func collectProductDeleteIDsForIDsChecked(ctx context.Context, tx *sqlx.Tx, query string, idLists ...[]string) ([]string, error) {
	hasValues := false
	args := make([]any, len(idLists))
	for _, ids := range idLists {
		if len(ids) > 0 {
			hasValues = true
		}
	}
	if !hasValues {
		return nil, nil
	}
	for index, ids := range idLists {
		if len(ids) == 0 {
			args[index] = []string{uuid.Nil.String()}
		} else {
			args[index] = ids
		}
	}
	expanded, values, err := sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}
	return collectProductDeleteIDs(ctx, tx, expanded, values...)
}

func execProductCleanup(ctx context.Context, tx *sqlx.Tx, query string, args ...any) error {
	if len(args) == 0 {
		return nil
	}
	args = normalizeProductDeleteArgs(args)
	expanded, values, err := sqlx.In(query, args...)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, tx.Rebind(expanded), values...)
	if isOptionalProductCleanupError(err) {
		return nil
	}
	return err
}

func normalizeProductDeleteArgs(args []any) []any {
	normalized := make([]any, len(args))
	for index, arg := range args {
		switch values := arg.(type) {
		case []string:
			if len(values) == 0 {
				// sqlx.In rejects empty slices. A valid zero UUID keeps the
				// optional cleanup query valid for PostgreSQL UUID columns
				// without matching any real row.
				normalized[index] = []string{uuid.Nil.String()}
			} else {
				normalized[index] = values
			}
		default:
			normalized[index] = arg
		}
	}
	return normalized
}

func deleteProductRowsByIDs(ctx context.Context, tx *sqlx.Tx, table, column string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	query, args, err := sqlx.In(fmt.Sprintf("DELETE FROM %s WHERE %s IN (?)", table, column), ids)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, tx.Rebind(query), args...)
	if isOptionalProductCleanupError(err) {
		return nil
	}
	return err
}

func appendUniqueProductDeleteIDs(ids []string, candidates ...string) []string {
	seen := make(map[string]struct{}, len(ids)+len(candidates))
	for _, id := range ids {
		seen[id] = struct{}{}
	}
	for _, id := range candidates {
		if strings.TrimSpace(id) == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func isOptionalProductCleanupError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	for _, fragment := range []string{
		"no such table",
		"no such column",
		"relation \"",
		"does not exist",
		"undefined table",
		"undefined column",
	} {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

// RestoreProduct restores a soft-deleted product
func (r *Repository) RestoreProduct(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET deleted_at = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NOT NULL`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrProductNotFound
	}

	return nil
}

// GetProductStockCount returns the stock count for a product
func (r *Repository) GetProductStockCount(ctx context.Context, productID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE((SELECT quantity FROM inventory WHERE product_id = $1), (
			SELECT COUNT(*) FROM inventory_items WHERE product_id = $1 AND UPPER(COALESCE(status, '')) = 'AVAILABLE'
		))
	`
	var count int
	err := r.db.GetContext(ctx, &count, query, productID)
	return count, err
}

func (r *Repository) GetProductStockSettings(ctx context.Context, productID uuid.UUID) (bool, int, error) {
	var trackIndividual bool
	var minStockLevel int
	err := r.db.QueryRowContext(ctx, `SELECT track_individual, min_stock_level FROM products WHERE id = $1 AND deleted_at IS NULL`, productID).Scan(&trackIndividual, &minStockLevel)
	return trackIndividual, minStockLevel, err
}

// ArchiveProduct archives a product (sets is_active to false)
func (r *Repository) ArchiveProduct(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET is_active = false, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrProductNotFound
	}

	return nil
}

// GetAvailableItemCount returns the count of available items for a product
func (r *Repository) GetAvailableItemCount(ctx context.Context, productID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM inventory_items
		WHERE product_id = $1 AND UPPER(COALESCE(status, '')) = 'AVAILABLE'
	`
	var count int
	err := r.db.GetContext(ctx, &count, query, productID)
	return count, err
}

// GetReservedItemCount returns the count of reserved items for a product
func (r *Repository) GetReservedItemCount(ctx context.Context, productID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM inventory_items
		WHERE product_id = $1 AND status = 'RESERVED'
	`
	var count int
	err := r.db.GetContext(ctx, &count, query, productID)
	return count, err
}

// SearchProducts searches products by name, SKU, or barcode
func (r *Repository) SearchProducts(ctx context.Context, query string, limit int) ([]Product, error) {
	likeOperator := "ILIKE"
	if r.db.DriverName() == "sqlite" {
		likeOperator = "LIKE"
	}
	searchQuery := `
		SELECT id, category_id, brand_id, preferred_supplier_id, name, description, model, sku, barcode, cost_price, selling_price, track_serial, track_individual, min_stock_level, warranty_days, is_active, deleted_at, created_at, updated_at
		FROM products
		WHERE (name ` + likeOperator + ` $1 OR model ` + likeOperator + ` $1 OR sku ` + likeOperator + ` $1 OR barcode ` + likeOperator + ` $1)
		AND is_active = true
		AND deleted_at IS NULL
		ORDER BY name
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	products := make([]Product, 0)
	for rows.Next() {
		product, scanErr := scanProductRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}
