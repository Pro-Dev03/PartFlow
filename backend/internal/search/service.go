package search

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Service handles search business logic
type Service struct {
	db *sqlx.DB
}

// searchTimestamp accepts SQLite's TEXT timestamps as well as PostgreSQL
// timestamp values. Without this scanner SQLite search rows were skipped
// while the count still reported matches, resulting in "results: null".
type searchTimestamp struct{ time.Time }

func (t *searchTimestamp) Scan(value any) error {
	parsed, err := dbutil.ParseTimestamp(value)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// NewService creates a new search service
func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

func (s *Service) likeOperator() string {
	if dbutil.IsSQLite(s.db) {
		return "LIKE"
	}
	return "ILIKE"
}

// Search performs a global search across all entities
func (s *Service) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	var results []SearchResult
	total := 0

	// If no types specified, search all
	types := req.Types
	if len(types) == 0 {
		types = []string{"products", "customers", "suppliers", "sales", "purchases"}
	}

	// Search each type
	for _, searchType := range types {
		var typeResults []SearchResult
		var typeTotal int

		switch searchType {
		case "products":
			typeResults, typeTotal = s.searchProducts(ctx, req.Query, req.Limit, req.Offset)
		case "customers":
			typeResults, typeTotal = s.searchCustomers(ctx, req.Query, req.Limit, req.Offset)
		case "suppliers":
			typeResults, typeTotal = s.searchSuppliers(ctx, req.Query, req.Limit, req.Offset)
		case "sales":
			typeResults, typeTotal = s.searchSales(ctx, req.Query, req.Limit, req.Offset)
		case "purchases":
			typeResults, typeTotal = s.searchPurchases(ctx, req.Query, req.Limit, req.Offset)
		}

		results = append(results, typeResults...)
		total += typeTotal
	}

	return &SearchResponse{
		Query:   req.Query,
		Results: results,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: req.Offset+req.Limit < total,
	}, nil
}

// searchProducts searches for products
func (s *Service) searchProducts(ctx context.Context, query string, limit, offset int) ([]SearchResult, int) {
	var results []SearchResult
	searchPattern := "%" + query + "%"

	like := s.likeOperator()
	queryStr := `
		SELECT id, name, sku, model, barcode, created_at
		FROM products
		WHERE is_active = true
		AND (name ` + like + ` $1 OR sku ` + like + ` $1 OR model ` + like + ` $1 OR barcode ` + like + ` $1)
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, queryStr, searchPattern, limit, offset)
	if err != nil {
		return results, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name, sku, model, barcode string
		var createdAt searchTimestamp

		if err := rows.Scan(&id, &name, &sku, &model, &barcode, &createdAt); err != nil {
			continue
		}

		metadata := map[string]interface{}{
			"sku":     sku,
			"model":   model,
			"barcode": barcode,
		}

		results = append(results, SearchResult{
			Type:      "product",
			ID:        id,
			Title:     name,
			Subtitle:  fmt.Sprintf("SKU: %s", sku),
			Metadata:  metadata,
			Score:     1.0,
			CreatedAt: createdAt.Time,
		})
	}
	_ = rows.Err()

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM products
		WHERE is_active = true
		AND (name ` + like + ` $1 OR sku ` + like + ` $1 OR model ` + like + ` $1 OR barcode ` + like + ` $1)
	`
	s.db.GetContext(ctx, &total, countQuery, searchPattern)

	return results, total
}

// searchCustomers searches for customers
func (s *Service) searchCustomers(ctx context.Context, query string, limit, offset int) ([]SearchResult, int) {
	var results []SearchResult
	searchPattern := "%" + query + "%"

	like := s.likeOperator()
	queryStr := `
		SELECT id, name, email, phone, created_at
		FROM customers
		WHERE is_active = true
		AND (name ` + like + ` $1 OR email ` + like + ` $1 OR phone ` + like + ` $1)
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, queryStr, searchPattern, limit, offset)
	if err != nil {
		return results, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name, email, phone string
		var createdAt searchTimestamp

		if err := rows.Scan(&id, &name, &email, &phone, &createdAt); err != nil {
			continue
		}

		metadata := map[string]interface{}{
			"email": email,
			"phone": phone,
		}

		results = append(results, SearchResult{
			Type:      "customer",
			ID:        id,
			Title:     name,
			Subtitle:  fmt.Sprintf("Email: %s", email),
			Metadata:  metadata,
			Score:     1.0,
			CreatedAt: createdAt.Time,
		})

	}
	_ = rows.Err()
	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM customers
		WHERE is_active = true
		AND (name ` + like + ` $1 OR email ` + like + ` $1 OR phone ` + like + ` $1)
	`
	s.db.GetContext(ctx, &total, countQuery, searchPattern)

	return results, total
}

// searchSuppliers searches for suppliers
func (s *Service) searchSuppliers(ctx context.Context, query string, limit, offset int) ([]SearchResult, int) {
	var results []SearchResult
	searchPattern := "%" + query + "%"

	like := s.likeOperator()
	queryStr := `
		SELECT id, name, email, phone, created_at
		FROM suppliers
		WHERE is_active = true
		AND (name ` + like + ` $1 OR email ` + like + ` $1 OR phone ` + like + ` $1)
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, queryStr, searchPattern, limit, offset)
	if err != nil {
		return results, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name, email, phone string
		var createdAt searchTimestamp

		if err := rows.Scan(&id, &name, &email, &phone, &createdAt); err != nil {
			continue
		}

		metadata := map[string]interface{}{
			"email": email,
			"phone": phone,
		}

		results = append(results, SearchResult{
			Type:      "supplier",
			ID:        id,
			Title:     name,
			Subtitle:  fmt.Sprintf("Email: %s", email),
			Metadata:  metadata,
			Score:     1.0,
			CreatedAt: createdAt.Time,
		})
	}
	_ = rows.Err()

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM suppliers
		WHERE is_active = true
		AND (name ` + like + ` $1 OR email ` + like + ` $1 OR phone ` + like + ` $1)
	`
	s.db.GetContext(ctx, &total, countQuery, searchPattern)

	return results, total
}

// searchSales searches for sales
func (s *Service) searchSales(ctx context.Context, query string, limit, offset int) ([]SearchResult, int) {
	var results []SearchResult
	searchPattern := "%" + query + "%"

	like := s.likeOperator()
	queryStr := `
		SELECT id, invoice_number, total_amount, sale_date, created_at
		FROM sales
		WHERE invoice_number ` + like + ` $1
		ORDER BY sale_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, queryStr, searchPattern, limit, offset)
	if err != nil {
		return results, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var invoiceNumber string
		var totalAmount float64
		var saleDate, createdAt searchTimestamp

		if err := rows.Scan(&id, &invoiceNumber, &totalAmount, &saleDate, &createdAt); err != nil {
			continue
		}

		metadata := map[string]interface{}{
			"total_amount": totalAmount,
			"sale_date":    saleDate.Time,
		}

		results = append(results, SearchResult{
			Type:      "sale",
			ID:        id,
			Title:     invoiceNumber,
			Subtitle:  fmt.Sprintf("Amount: %.2f", totalAmount),
			Metadata:  metadata,
			Score:     1.0,
			CreatedAt: createdAt.Time,
		})
	}
	_ = rows.Err()

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM sales
		WHERE invoice_number ` + like + ` $1
	`
	s.db.GetContext(ctx, &total, countQuery, searchPattern)

	return results, total
}

// searchPurchases searches for purchases
func (s *Service) searchPurchases(ctx context.Context, query string, limit, offset int) ([]SearchResult, int) {
	var results []SearchResult
	searchPattern := "%" + query + "%"

	like := s.likeOperator()
	queryStr := `
		SELECT id, invoice_number, total_amount, purchase_date, created_at
		FROM purchases
		WHERE invoice_number ` + like + ` $1
		ORDER BY purchase_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, queryStr, searchPattern, limit, offset)
	if err != nil {
		return results, 0
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var invoiceNumber string
		var totalAmount float64
		var purchaseDate, createdAt searchTimestamp

		if err := rows.Scan(&id, &invoiceNumber, &totalAmount, &purchaseDate, &createdAt); err != nil {
			continue
		}

		metadata := map[string]interface{}{
			"total_amount":  totalAmount,
			"purchase_date": purchaseDate.Time,
		}

		results = append(results, SearchResult{
			Type:      "purchase",
			ID:        id,
			Title:     invoiceNumber,
			Subtitle:  fmt.Sprintf("Amount: %.2f", totalAmount),
			Metadata:  metadata,
			Score:     1.0,
			CreatedAt: createdAt.Time,
		})

	}
	_ = rows.Err()
	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM purchases
		WHERE invoice_number ` + like + ` $1
	`
	s.db.GetContext(ctx, &total, countQuery, searchPattern)

	return results, total
}

// GetSearchStats retrieves search statistics
func (s *Service) GetSearchStats(ctx context.Context) (*SearchStats, error) {
	stats := &SearchStats{}

	// Get total products
	err := s.db.GetContext(ctx, &stats.TotalProducts, `SELECT COUNT(*) FROM products WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total products: %w", err)
	}

	// Get total customers
	err = s.db.GetContext(ctx, &stats.TotalCustomers, `SELECT COUNT(*) FROM customers WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total customers: %w", err)
	}

	// Get total suppliers
	err = s.db.GetContext(ctx, &stats.TotalSuppliers, `SELECT COUNT(*) FROM suppliers WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total suppliers: %w", err)
	}

	// Get total sales
	err = s.db.GetContext(ctx, &stats.TotalSales, `SELECT COUNT(*) FROM sales`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total sales: %w", err)
	}

	// Get total purchases
	err = s.db.GetContext(ctx, &stats.TotalPurchases, `SELECT COUNT(*) FROM purchases`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total purchases: %w", err)
	}

	return stats, nil
}
