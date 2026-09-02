package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/business"
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

	// Query to get real statistics. Keep outstanding and overdue balances
	// separate: a future-dated debt must not appear in the overdue card.
	overdueDebtExpr := `(SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE remaining_amount > 0 AND due_date < NOW())`
	if strings.EqualFold(s.db.DriverName(), "sqlite") {
		overdueDebtExpr = `(SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE remaining_amount > 0 AND date(due_date) < date('now'))`
	}
	query := fmt.Sprintf(`
		SELECT
			(SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) as total_sales,
			(SELECT COUNT(*) FROM sales WHERE status = 'pending') as pending_orders,
			(SELECT COALESCE(SUM(total_amount), 0) FROM purchases WHERE LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) as total_purchases,
			(SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed')) as total_expenses,
			(SELECT COUNT(*) FROM products WHERE is_active = true) as total_products,
			(SELECT COUNT(*) FROM customers) as total_customers,
			(SELECT COUNT(*) FROM suppliers) as total_suppliers,
			(SELECT COUNT(*) FROM products p
			 WHERE p.is_active = true
			 AND p.deleted_at IS NULL
			 AND p.min_stock_level > 0
			 AND (SELECT COUNT(*) FROM inventory_items ii
			      WHERE ii.product_id = p.id
			      AND ii.condition <> 'USED'
			      AND ii.status = 'AVAILABLE') < p.min_stock_level) as low_stock_items,
			%s as overdue_debts,
			(SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE remaining_amount > 0) as outstanding_debts,
			(SELECT COUNT(*) FROM returns WHERE LOWER(COALESCE(status, 'pending')) IN ('pending', 'approved', 'processing')) as pending_returns,
			0 as pending_claims
	`, overdueDebtExpr)

	var result struct {
		TotalSales       float64 `db:"total_sales"`
		PendingOrders    int     `db:"pending_orders"`
		TotalPurchases   float64 `db:"total_purchases"`
		TotalExpenses    float64 `db:"total_expenses"`
		TotalProducts    int     `db:"total_products"`
		TotalCustomers   int     `db:"total_customers"`
		TotalSuppliers   int     `db:"total_suppliers"`
		LowStockItems    int     `db:"low_stock_items"`
		OverdueDebts     float64 `db:"overdue_debts"`
		OutstandingDebts float64 `db:"outstanding_debts"`
		PendingReturns   int     `db:"pending_returns"`
		PendingClaims    int     `db:"pending_claims"`
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
	// Keep return counters in sync with the cards when the enhanced returns
	// schema is present. Legacy local databases simply report zero here.
	var returnSummary struct {
		TotalReturns int     `db:"total_returns"`
		Refunded     float64 `db:"refunded"`
		GrossSales   int     `db:"gross_sales"`
	}
	if err := s.db.GetContext(ctx, &returnSummary, `
		SELECT COUNT(*) AS total_returns,
		       COALESCE(SUM(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED' THEN total_refund_amount ELSE 0 END), 0) AS refunded
		       ,(SELECT COUNT(*) FROM sales WHERE LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) AS gross_sales
		FROM returns`); err == nil {
		stats.TotalReturns = float64(returnSummary.TotalReturns)
		stats.TotalRefunded = returnSummary.Refunded
		// Both fields are monetary amounts; the return count is exposed by
		// total_returns and must not be subtracted from currency.
		stats.NetSales = stats.TotalSales - stats.TotalRefunded
		stats.NetRevenue = stats.TotalSales - stats.TotalRefunded
		if returnSummary.GrossSales > 0 {
			stats.ReturnRate = (stats.TotalReturns / float64(returnSummary.GrossSales)) * 100
		}
	}

	// Calculate lifetime profit from sold-item cost, not purchase cash outflow.
	// Subtract completed refunds and restore the cost of returned items so this
	// field remains a real profit metric even when a return is processed.
	stats.TotalRevenue = stats.TotalSales
	var grossProfit, refunded, returnedCost float64
	if err := s.db.GetContext(ctx, &grossProfit, `
		SELECT COALESCE(SUM(si.total_amount - si.quantity * COALESCE(si.unit_cost, p.cost_price, 0)), 0)
		FROM sale_items si JOIN sales sl ON sl.id = si.sale_id JOIN products p ON p.id = si.product_id
		WHERE LOWER(COALESCE(sl.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`); err == nil {
		_ = s.db.GetContext(ctx, &refunded, `SELECT COALESCE(SUM(total_refund_amount), 0) FROM returns WHERE UPPER(COALESCE(status, '')) = 'COMPLETED'`)
		_ = s.db.GetContext(ctx, &returnedCost, `
			SELECT COALESCE(SUM(ri.quantity_returned * COALESCE(ri.original_cost, si.unit_cost, p.cost_price, 0)), 0)
			FROM return_items ri JOIN returns r ON r.id = ri.return_id
			LEFT JOIN sale_items si ON si.id = ri.sale_item_id
			LEFT JOIN products p ON p.id = ri.product_id
			WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'`)
		stats.TotalProfit = grossProfit - stats.TotalExpenses - refunded + returnedCost
	} else {
		stats.TotalProfit = stats.TotalSales - stats.TotalExpenses
	}

	// Populate frontend-compatible fields
	// These fields are date-scoped. Do not reuse the lifetime totals above:
	// purchases are cash outflows, not today's cost of goods sold.
	if today, todayErr := fetchTodayMetrics(ctx, s.db, time.Now().UTC()); todayErr == nil {
		stats.TodaySales = today.Sales
		stats.TodayProfit = today.Profit
	}
	stats.OutstandingDebts = result.OutstandingDebts
	stats.ActiveCustomers = result.TotalCustomers
	stats.LowStockCount = result.LowStockItems

	// Calculate overdue debts count properly
	var overdueDebtsCount int
	countQuery := `
		SELECT COUNT(DISTINCT customer_id)
		FROM debts
		WHERE remaining_amount > 0
		  AND due_date < NOW()
		  AND ` + business.OpenDebtStatusSQL("status") + `
	`
	if s.db.DriverName() == "sqlite" {
		countQuery = `
			SELECT COUNT(DISTINCT customer_id)
			FROM debts
			WHERE remaining_amount > 0
			  AND julianday(due_date) < julianday('now')
			  AND ` + business.OpenDebtStatusSQL("status") + `
		`
	}
	err = s.db.GetContext(ctx, &overdueDebtsCount, countQuery)
	if err != nil {
		// Fallback to 0 if query fails
		overdueDebtsCount = 0
	}
	stats.OverdueDebtsCount = overdueDebtsCount

	// Alerts disabled for performance
	stats.Alerts = []Alert{}

	// Populate the data used by the dashboard charts and activity panel. These
	// reads are intentionally best-effort: a missing optional table must not
	// make the main dashboard request fail in an existing local database.
	stats.SalesChart = s.fetchSalesChart(ctx)
	stats.InventoryDistribution = s.fetchInventoryDistribution(ctx)
	stats.RecentActivity = s.fetchRecentActivity(ctx)

	return stats, nil
}

func (s *CachedService) fetchSalesChart(ctx context.Context) []SalesChartData {
	query := `
		WITH sale_costs AS (
			SELECT s.id, s.created_at, s.total_amount,
				COALESCE(SUM(COALESCE(ii.purchase_cost, p.purchase_price, p.cost_price, 0) * COALESCE(si.quantity, 0)), 0) AS cost
			FROM sales s
			LEFT JOIN sale_items si ON si.sale_id = s.id
			LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
			LEFT JOIN products p ON p.id = si.product_id
			WHERE LOWER(COALESCE(s.status, '')) = 'completed'
			  AND datetime(s.created_at) >= datetime('now', '-30 days')
			GROUP BY s.id, s.created_at, s.total_amount
		)
		SELECT strftime('%Y-%m-%d', created_at) AS name,
		       COALESCE(SUM(total_amount), 0) AS sales,
		       COALESCE(SUM(total_amount - cost), 0) AS profit
		FROM sale_costs
		GROUP BY strftime('%Y-%m-%d', created_at)
		ORDER BY name
	`
	if s.db.DriverName() != "sqlite" {
		query = `
			WITH sale_costs AS (
				SELECT s.id, s.created_at, s.total_amount,
				       COALESCE(SUM(COALESCE(ii.purchase_cost, p.purchase_price, p.cost_price, 0) * COALESCE(si.quantity, 0)), 0) AS cost
				FROM sales s
				LEFT JOIN sale_items si ON si.sale_id = s.id
				LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
				LEFT JOIN products p ON p.id = si.product_id
				WHERE LOWER(COALESCE(s.status, '')) = 'completed'
				  AND s.created_at >= NOW() - INTERVAL '30 days'
				GROUP BY s.id, s.created_at, s.total_amount
			)
			SELECT TO_CHAR(DATE(created_at), 'YYYY-MM-DD') AS name,
			       COALESCE(SUM(total_amount), 0) AS sales,
			       COALESCE(SUM(total_amount - cost), 0) AS profit
			FROM sale_costs
			GROUP BY DATE(created_at)
			ORDER BY DATE(created_at)
		`
	}

	var rows []SalesChartData
	if err := s.db.SelectContext(ctx, &rows, query); err != nil {
		return []SalesChartData{}
	}
	return rows
}

func (s *CachedService) fetchInventoryDistribution(ctx context.Context) *InventoryDistributionData {
	query := `
		SELECT UPPER(COALESCE(status, 'UNKNOWN')) AS status,
		       COUNT(*) AS count,
		       COALESCE(SUM(selling_price), 0) AS value
		FROM inventory_items
		GROUP BY UPPER(COALESCE(status, 'UNKNOWN'))
		ORDER BY status
	`

	var rows []struct {
		Status string  `db:"status"`
		Count  int     `db:"count"`
		Value  float64 `db:"value"`
	}
	if err := s.db.SelectContext(ctx, &rows, query); err != nil {
		return nil
	}

	data := make([]InventoryDistributionItem, 0, len(rows))
	var totalValue float64
	var totalItems int
	for _, row := range rows {
		name, color, health := inventoryStatusPresentation(row.Status)
		data = append(data, InventoryDistributionItem{
			Name:   name,
			Count:  row.Count,
			Value:  row.Value,
			Color:  color,
			Status: health,
		})
		totalItems += row.Count
		totalValue += row.Value
	}

	return &InventoryDistributionData{
		TotalValue: totalValue,
		TotalItems: totalItems,
		Data:       data,
	}
}

func inventoryStatusPresentation(status string) (name, color, health string) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "AVAILABLE", "IN_STOCK", "IN STOCK":
		return "متاح", "#10b981", "good"
	case "RESERVED", "RETURNED":
		return "محجوز/مرتجع", "#f59e0b", "low"
	case "SOLD":
		return "مباع", "#64748b", "low"
	case "DAMAGED", "IN_REPAIR", "IN REPAIR":
		return "تالف/قيد الإصلاح", "#ef4444", "critical"
	default:
		if strings.TrimSpace(status) == "" {
			return "غير محدد", "#94a3b8", "low"
		}
		return status, "#94a3b8", "low"
	}
}

func (s *CachedService) fetchRecentActivity(ctx context.Context) []RecentActivityItem {
	query := `
		SELECT id, type, title, description, amount, activity_time AS time, status
		FROM (
			SELECT id, 'sale' AS type, 'بيع' AS title, 'عملية بيع' AS description,
			       total_amount AS amount, datetime(created_at) AS activity_time, status
			FROM sales
			UNION ALL
			SELECT id, 'purchase' AS type, 'شراء' AS title, 'عملية شراء' AS description,
			       total_amount AS amount, datetime(created_at) AS activity_time, status
			FROM purchases
		) AS activity
		ORDER BY activity_time DESC
		LIMIT 10
	`
	if s.db.DriverName() != "sqlite" {
		query = `
			SELECT id, type, title, description, amount, activity_time AS time, status
			FROM (
				SELECT id, 'sale' AS type, 'بيع' AS title, 'عملية بيع' AS description,
				       total_amount AS amount, TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS') AS activity_time, status
				FROM sales
				UNION ALL
				SELECT id, 'purchase' AS type, 'شراء' AS title, 'عملية شراء' AS description,
				       total_amount AS amount, TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS') AS activity_time, status
				FROM purchases
			) AS activity
			ORDER BY activity_time DESC
			LIMIT 10
		`
	}

	var activities []RecentActivityItem
	if err := s.db.SelectContext(ctx, &activities, query); err != nil {
		return []RecentActivityItem{}
	}
	return activities
}

// InvalidateCache clears the cache (call after data changes)
func (s *CachedService) InvalidateCache() {
	s.cache.Clear()
}

// GetLowStockItems retrieves low stock items with details
func (s *CachedService) GetLowStockItems(ctx context.Context) ([]LowStockItem, error) {
	query := `
		SELECT 
			p.id,
			p.name as product_name,
			COUNT(CASE WHEN i.status = 'AVAILABLE' THEN i.id END) as quantity,
			p.min_stock_level,
			p.cost_price,
			p.selling_price,
			p.preferred_supplier_id
		FROM products p
		LEFT JOIN inventory_items i ON p.id = i.product_id
		WHERE p.is_active = true
		AND p.deleted_at IS NULL
		AND p.min_stock_level > 0
		GROUP BY p.id, p.name, p.min_stock_level, p.cost_price, p.selling_price, p.preferred_supplier_id
		HAVING COUNT(CASE WHEN i.status = 'AVAILABLE' THEN i.id END) < p.min_stock_level
		ORDER BY (p.min_stock_level - COUNT(CASE WHEN i.status = 'AVAILABLE' THEN i.id END)) DESC
		LIMIT 10
	`

	var items []LowStockItem
	err := s.db.SelectContext(ctx, &items, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock items: %w", err)
	}

	return items, nil
}

// GetOverdueDebts retrieves overdue debts with details
func (s *CachedService) GetOverdueDebts(ctx context.Context) ([]OverdueDebtItem, error) {
	query := `
		SELECT 
			d.id,
			d.customer_id,
			c.name as customer_name,
			d.remaining_amount,
			d.due_date,
			EXTRACT(DAY FROM NOW() - d.due_date)::int as days_overdue,
			COALESCE(c.phone, '') as phone
		FROM debts d
		JOIN customers c ON d.customer_id = c.id
		WHERE d.remaining_amount > 0
		AND d.due_date < NOW()
		AND d.status IN ('pending', 'partial', 'overdue')
		ORDER BY d.due_date ASC
		LIMIT 10
	`
	if s.db.DriverName() == "sqlite" {
		query = `
			SELECT d.id, d.customer_id, c.name AS customer_name, d.remaining_amount, d.due_date,
			CAST((julianday('now') - julianday(d.due_date)) AS INTEGER) AS days_overdue,
			COALESCE(c.phone, '') AS phone
			FROM debts d JOIN customers c ON d.customer_id = c.id
			WHERE d.remaining_amount > 0
			  AND julianday(d.due_date) < julianday('now')
			  AND ` + business.OpenDebtStatusSQL("d.status") + `
			ORDER BY d.due_date ASC LIMIT 10
		`
	}

	var debts []OverdueDebtItem
	err := s.db.SelectContext(ctx, &debts, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue debts: %w", err)
	}

	return debts, nil
}
