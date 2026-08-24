package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type CachedService struct {
	db    *sqlx.DB
	cache *Cache
}

var globalCacheService *CachedService

func NewCachedService(db *sqlx.DB) *CachedService {
	service := &CachedService{
		db:    db,
		cache: NewCache(5 * time.Minute), // 5 minute cache
	}
	globalCacheService = service
	return service
}

// GetGlobalCacheService returns the global cache service instance for invalidation
func GetGlobalCacheService() *CachedService {
	return globalCacheService
}

// GetDashboardStats retrieves dashboard statistics with caching
func (s *CachedService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	// Try to get from cache first
	if cached, found := s.cache.Get(); found {
		return cached, nil
	}

	// Cache miss - fetch from database
	stats, err := s.fetchFromDatabase(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.Set(stats)

	return stats, nil
}

// fetchFromDatabase retrieves stats from database
func (s *CachedService) fetchFromDatabase(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Query to get real statistics
	query := `
		SELECT
			(SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE status = 'completed') as total_sales,
			(SELECT COUNT(*) FROM sales WHERE status = 'pending') as pending_orders,
			(SELECT COALESCE(SUM(total_amount), 0) FROM purchases WHERE status = 'received') as total_purchases,
			(SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE status = 'approved') as total_expenses,
			(SELECT COUNT(*) FROM products WHERE is_active = true) as total_products,
			(SELECT COUNT(*) FROM customers) as total_customers,
			(SELECT COUNT(*) FROM suppliers) as total_suppliers,
			(SELECT COUNT(*) FROM products p
			 WHERE p.is_active = true
			 AND p.min_stock_level > 0
			 AND (SELECT COALESCE(SUM(quantity), 0) FROM inventory WHERE product_id = p.id) < p.min_stock_level) as low_stock_items,
			(SELECT COALESCE(SUM(current_balance), 0) FROM customers WHERE current_balance > 0) as overdue_debts,
			(SELECT COUNT(*) FROM returns WHERE status = 'pending') as pending_returns,
			0 as pending_claims
	`

	var result struct {
		TotalSales     float64 `db:"total_sales"`
		PendingOrders  int     `db:"pending_orders"`
		TotalPurchases float64 `db:"total_purchases"`
		TotalExpenses  float64 `db:"total_expenses"`
		TotalProducts  int     `db:"total_products"`
		TotalCustomers int     `db:"total_customers"`
		TotalSuppliers int     `db:"total_suppliers"`
		LowStockItems  int     `db:"low_stock_items"`
		OverdueDebts   float64 `db:"overdue_debts"`
		PendingReturns int     `db:"pending_returns"`
		PendingClaims  int     `db:"pending_claims"`
	}

	err := s.db.GetContext(ctx, &result, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard stats: %w", err)
	}

	// Map result to stats
	stats.TotalSales = result.TotalSales
	stats.PendingOrders = result.PendingOrders
	stats.TotalPurchases = result.TotalPurchases
	stats.TotalExpenses = result.TotalExpenses
	stats.TotalProducts = result.TotalProducts
	stats.TotalCustomers = result.TotalCustomers
	stats.TotalSuppliers = result.TotalSuppliers
	stats.LowStockItems = result.LowStockItems
	stats.OverdueDebts = result.OverdueDebts
	stats.PendingReturns = result.PendingReturns
	stats.PendingClaims = result.PendingClaims

	// Calculate totals
	stats.TotalRevenue = stats.TotalSales
	stats.TotalProfit = stats.TotalSales - stats.TotalPurchases - stats.TotalExpenses

	// Populate frontend-compatible fields
	stats.TodaySales = result.TotalSales
	stats.TodayProfit = stats.TotalProfit
	stats.OutstandingDebts = result.OverdueDebts
	stats.ActiveCustomers = result.TotalCustomers
	stats.LowStockCount = result.LowStockItems
	
	// Calculate overdue debts count properly
	var overdueDebtsCount int
	countQuery := `
		SELECT COUNT(DISTINCT c.id) 
		FROM customers c
		WHERE c.current_balance > 0
		AND c.id IN (
			SELECT DISTINCT customer_id FROM debts
			WHERE remaining_amount > 0
			AND due_date < NOW()
		)
	`
	err = s.db.GetContext(ctx, &overdueDebtsCount, countQuery)
	if err != nil {
		// Fallback to 0 if query fails
		overdueDebtsCount = 0
	}
	stats.OverdueDebtsCount = overdueDebtsCount

	// Alerts disabled for performance
	stats.Alerts = []Alert{}

	return stats, nil
}

// InvalidateCache clears the cache (call after data changes)
func (s *CachedService) InvalidateCache() {
	s.cache.Clear()
}
