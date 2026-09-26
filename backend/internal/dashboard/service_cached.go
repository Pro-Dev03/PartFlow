package dashboard

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/business"
	dbutil "github.com/partflow/smart-store/internal/database"
)

type CachedService struct {
	db    *sqlx.DB
	cache *Cache
}

var globalCacheService *CachedService

func NewCachedService(db *sqlx.DB) *CachedService {
	if db != nil && db.DriverName() == "sqlite" {
		ensureSQLiteAccountingReturnViews(db.DB)
	}
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

// fetchFromDatabase retrieves stats from database using the current store time.
func (s *CachedService) fetchFromDatabase(ctx context.Context) (*DashboardStats, error) {
	return s.fetchFromDatabaseAt(ctx, time.Now().UTC())
}

// fetchFromDatabaseAt retrieves stats for a specific business instant so date-scoped
// regressions remain deterministic and match the store's official business date.
func (s *CachedService) fetchFromDatabaseAt(ctx context.Context, now time.Time) (*DashboardStats, error) {
	stats := &DashboardStats{}
	storeDate, err := accounting.StoreDate(now)
	if err != nil {
		return nil, fmt.Errorf("calculate dashboard store date: %w", err)
	}

	// Query to get real statistics. Keep outstanding and overdue balances
	// separate: a future-dated debt must not appear in the overdue card.
	overdueDebtExpr := fmt.Sprintf(`(SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE remaining_amount > 0 AND due_date::date < DATE '%s')`, storeDate)
	if strings.EqualFold(s.db.DriverName(), "sqlite") {
		overdueDebtExpr = fmt.Sprintf(`(SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE remaining_amount > 0 AND date(due_date) < date('%s'))`, storeDate)
	}
	debtorCountExpr := `(SELECT COUNT(DISTINCT customer_id) FROM debts WHERE remaining_amount > 0 AND ` + business.OpenDebtStatusSQL("status") + `)`
	activeCustomerCountExpr := `0`
	if (s.db.DriverName() != "sqlite" || sqliteHasColumns(s.db, "sales", "customer_id")) && sqliteHasColumns(s.db, "customers", "id") {
		activeCustomerCountExpr = `(SELECT COUNT(*) FROM customers c WHERE EXISTS (
			SELECT 1 FROM sales s
			WHERE s.customer_id = c.id
			  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		))`
	}
	lowStockExpr := `(SELECT COUNT(*) FROM products p
			 LEFT JOIN inventory inv ON inv.product_id = p.id
			 WHERE p.is_active = true
			 AND p.deleted_at IS NULL
			 AND p.min_stock_level > 0
			 AND NOT EXISTS (SELECT 1 FROM inventory_items used_i WHERE used_i.product_id = p.id AND UPPER(COALESCE(used_i.condition, '')) = 'USED')
			 AND (NOT EXISTS (SELECT 1 FROM inventory_items ii WHERE ii.product_id = p.id)
			      OR EXISTS (SELECT 1 FROM inventory_items ii WHERE ii.product_id = p.id AND COALESCE(ii.condition, '') <> 'USED'))
			 AND COALESCE(inv.quantity, (SELECT COUNT(*) FROM inventory_items ii
			      WHERE ii.product_id = p.id AND COALESCE(ii.condition, '') <> 'USED' AND ii.status = 'AVAILABLE'), 0) <= p.min_stock_level)`
	query := fmt.Sprintf(`
		SELECT
			(SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) as total_sales,
			(SELECT COUNT(*) FROM sales WHERE status = 'pending') as pending_orders,
			(SELECT COALESCE(SUM(total_amount), 0) FROM purchases WHERE LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) as total_purchases,
			(SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')) as total_expenses,
			(SELECT COUNT(*) FROM products WHERE is_active = true AND deleted_at IS NULL) as total_products,
			(SELECT COUNT(*) FROM customers) as total_customers,
			(SELECT COUNT(*) FROM suppliers) as total_suppliers,
			%s as low_stock_items,
			%s as overdue_debts,
			(SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE remaining_amount > 0) as outstanding_debts,
			%s as debtors_count,
			%s as active_customers_count,
			(SELECT COUNT(*) FROM returns WHERE LOWER(COALESCE(status, 'pending')) IN ('pending', 'approved', 'processing')) as pending_returns,
			0 as pending_claims
	`, lowStockExpr, overdueDebtExpr, debtorCountExpr, activeCustomerCountExpr)

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
		DebtorsCount     int     `db:"debtors_count"`
		ActiveCustomers  int     `db:"active_customers_count"`
		PendingReturns   int     `db:"pending_returns"`
		PendingClaims    int     `db:"pending_claims"`
	}

	err = s.db.GetContext(ctx, &result, query)
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
	returnReferenceFilter := ""
	if isSQLiteDriver(s.db.DriverName()) {
		if sqliteHasColumns(s.db, "returns", "reference_number") {
			returnReferenceFilter = " AND COALESCE(reference_number, '') NOT LIKE 'REV-%'"
		}
	} else {
		returnReferenceFilter = " AND COALESCE(reference_number, '') NOT LIKE 'REV-%'"
	}
	var returnSummary struct {
		TotalReturns int     `db:"total_returns"`
		Refunded     float64 `db:"refunded"`
		GrossSales   int     `db:"gross_sales"`
	}
	if err := s.db.GetContext(ctx, &returnSummary, `
		SELECT COUNT(*) AS total_returns,
		       COALESCE(SUM(CASE WHEN UPPER(COALESCE(status, '')) = 'COMPLETED'`+returnReferenceFilter+` THEN total_refund_amount ELSE 0 END), 0) AS refunded
		       ,(SELECT COUNT(*) FROM sales WHERE LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) AS gross_sales
		FROM accounting_returns`); err == nil {
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
	grossProfitQuery := `
		SELECT COALESCE(SUM(si.total_amount - COALESCE(si.tax_amount, 0) - si.quantity * COALESCE(NULLIF(si.unit_cost, 0), p.cost_price, 0)), 0)
		FROM sale_items si JOIN sales sl ON sl.id = si.sale_id JOIN products p ON p.id = si.product_id
		WHERE LOWER(COALESCE(sl.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`
	if !isSQLiteDriver(s.db.DriverName()) || sqliteHasColumns(s.db, "sales", "cost_amount") {
		grossProfitQuery = `
			SELECT COALESCE(SUM((sale_revenue - sale_tax) - COALESCE(NULLIF(sale_cost, 0), line_cost)), 0)
			FROM (
				SELECT sl.id, sl.total_amount AS sale_revenue, COALESCE(sl.tax_amount, 0) AS sale_tax,
					sl.cost_amount AS sale_cost,
					COALESCE(SUM(si.total_amount - COALESCE(si.tax_amount, 0) - si.quantity * COALESCE(NULLIF(si.unit_cost, 0), p.cost_price, 0)), 0) AS line_cost
				FROM sales sl
				LEFT JOIN sale_items si ON si.sale_id = sl.id
				LEFT JOIN products p ON p.id = si.product_id
				WHERE LOWER(COALESCE(sl.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
				GROUP BY sl.id, sl.total_amount, sl.tax_amount, sl.cost_amount
			) sale_costs`
	}
	if s.db.DriverName() == "sqlite" && !sqliteHasColumns(s.db, "sale_items", "tax_amount") {
		grossProfitQuery = strings.ReplaceAll(grossProfitQuery, "COALESCE(si.tax_amount, 0)", "0")
	}
	if err := s.db.GetContext(ctx, &grossProfit, grossProfitQuery); err == nil {
		_ = s.db.GetContext(ctx, &refunded, `SELECT COALESCE(SUM(total_refund_amount), 0) FROM accounting_returns WHERE UPPER(COALESCE(status, '')) = 'COMPLETED'`+returnReferenceFilter)
		_ = s.db.GetContext(ctx, &returnedCost, `
			SELECT COALESCE(SUM(ri.quantity_returned * COALESCE(NULLIF(ri.original_cost, 0), NULLIF(si.unit_cost, 0), p.cost_price, 0)), 0)
			FROM accounting_return_items ri JOIN accounting_returns r ON r.id = ri.return_id
			LEFT JOIN sale_items si ON si.id = ri.sale_item_id
			LEFT JOIN products p ON p.id = ri.product_id
			WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'`+strings.ReplaceAll(returnReferenceFilter, "reference_number", "r.reference_number"))
		stats.TotalProfit = grossProfit - stats.TotalExpenses - refunded + returnedCost
	} else {
		stats.TotalProfit = stats.TotalSales - stats.TotalExpenses
	}

	// Populate frontend-compatible fields
	// These fields are date-scoped. Do not reuse the lifetime totals above:
	// purchases are cash outflows, not today's cost of goods sold.
	if today, todayErr := fetchTodayMetrics(ctx, s.db, now); todayErr == nil {
		stats.TodaySales = today.Sales
		stats.TodayProfit = today.Profit
		stats.TodaySupplierReturns = today.SupplierReturns
		stats.TodayCollected = today.Collected
		stats.TodayDebtCollected = today.DebtCollected
		stats.TodaySupplierPaid = today.SupplierPaid
		stats.TodayExpenses = today.Expenses
		// Approved expenses affect accrual profit, but there is no payment_status
		// field to prove that they were paid. Keep cash flow separate until that
		// state exists instead of treating approval as a cash movement.
		stats.TodayCashDifference = today.Collected - today.SupplierPaid + today.SupplierReturns
	}
	stats.OutstandingDebts = result.OutstandingDebts
	stats.OutstandingDebtorCount = result.DebtorsCount
	stats.ActiveCustomers = result.ActiveCustomers
	stats.LowStockCount = result.LowStockItems

	// Calculate overdue debts count properly
	var overdueDebtsCount int
	countQuery := `
		SELECT COUNT(DISTINCT customer_id)
		FROM debts
		WHERE remaining_amount > 0
		AND due_date::date < DATE '%s'
		  AND ` + business.OpenDebtStatusSQL("status") + `
	`
	if s.db.DriverName() == "sqlite" {
		countQuery = `
			SELECT COUNT(DISTINCT customer_id)
			FROM debts
			WHERE remaining_amount > 0
			  AND date(due_date) < date('%s')
			  AND ` + business.OpenDebtStatusSQL("status") + `
		`
	}
	countQuery = fmt.Sprintf(countQuery, storeDate)
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
	chartStart, _, err := accounting.StoreDateRange(time.Now(), 90)
	if err != nil {
		return []SalesChartData{}
	}
	if isSQLiteDriver(s.db.DriverName()) {
		return s.fetchSQLiteSalesChart(ctx, chartStart)
	}
	// Zero-valued persisted costs are legacy "unknown" values in sales created
	// while aggregate inventory had no individual purchase rows. Fall back to
	// the captured inventory/product cost so gross profit is not overstated.
	productCostExpr := "COALESCE(NULLIF(si.unit_cost, 0), NULLIF(ii.purchase_cost, 0), p.cost_price, 0)"
	taxExpr := "COALESCE(s.tax_amount, 0)"
	if s.db.DriverName() == "sqlite" {
		unitCostExpr := "0"
		if sqliteHasColumns(s.db, "sale_items", "unit_cost") {
			unitCostExpr = "si.unit_cost"
		}
		inventoryCostExpr := "0"
		if sqliteHasColumns(s.db, "inventory_items", "purchase_cost") {
			inventoryCostExpr = "ii.purchase_cost"
		}
		productCostColumn := "purchase_price"
		if sqliteHasColumns(s.db, "products", "cost_price") {
			productCostColumn = "cost_price"
		}
		productCostExpr = fmt.Sprintf("COALESCE(NULLIF(%s, 0), NULLIF(%s, 0), p.%s, 0)", unitCostExpr, inventoryCostExpr, productCostColumn)
		if !sqliteHasColumns(s.db, "sales", "tax_amount") {
			taxExpr = "0"
		}
	}
	saleCostExpression := fmt.Sprintf("COALESCE(SUM(%s * COALESCE(si.quantity, 0)), 0)", productCostExpr)
	saleCostGroup := ""
	if s.db.DriverName() != "sqlite" || sqliteHasColumns(s.db, "sales", "cost_amount") {
		saleCostExpression = fmt.Sprintf("CASE WHEN NULLIF(s.cost_amount, 0) IS NOT NULL THEN s.cost_amount ELSE %s END", saleCostExpression)
		saleCostGroup = ", s.cost_amount"
	}
	query := fmt.Sprintf(`
		WITH sale_costs AS (
			SELECT s.id, s.sale_date, s.total_amount, %s AS tax_amount,
				%s AS cost
			FROM sales s
			LEFT JOIN sale_items si ON si.sale_id = s.id
			LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
			LEFT JOIN products p ON p.id = si.product_id
			WHERE LOWER(COALESCE(s.status, '')) = 'completed'
			  AND s.sale_date IS NOT NULL
			  AND date(s.sale_date) >= date('%s')
			GROUP BY s.id, s.sale_date, s.total_amount, s.tax_amount%s
		)
		SELECT strftime('%%Y-%%m-%%d', sale_date) AS name,
		       COALESCE(SUM(total_amount), 0) AS sales,
		       COALESCE(SUM(total_amount - tax_amount - cost), 0) AS profit
		FROM sale_costs
		GROUP BY strftime('%%Y-%%m-%%d', sale_date)
		ORDER BY name
	`, taxExpr, saleCostExpression, chartStart, saleCostGroup)
	if !isSQLiteDriver(s.db.DriverName()) {
		query = fmt.Sprintf(`
			WITH sale_costs AS (
				SELECT s.id, s.sale_date, s.total_amount, COALESCE(s.tax_amount, 0) AS tax_amount,
				       %s AS cost
				FROM sales s
				LEFT JOIN sale_items si ON si.sale_id = s.id
				LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id
				LEFT JOIN products p ON p.id = si.product_id
				WHERE LOWER(COALESCE(s.status, '')) = 'completed'
				  AND s.sale_date IS NOT NULL
				  AND s.sale_date::date >= DATE '%s'
				GROUP BY s.id, s.sale_date, s.total_amount, s.tax_amount%s
			)
			SELECT TO_CHAR(DATE(sale_date), 'YYYY-MM-DD') AS name,
			       COALESCE(SUM(total_amount - tax_amount), 0) AS sales,
			       COALESCE(SUM(total_amount - tax_amount - cost), 0) AS profit
			FROM sale_costs
			GROUP BY DATE(sale_date)
			ORDER BY DATE(sale_date)
		`, saleCostExpression, saleCostGroup, chartStart)
	}
	if isSQLiteDriver(s.db.DriverName()) && !sqliteHasColumns(s.db, "sales", "tax_amount") {
		query = strings.ReplaceAll(query, "COALESCE(s.tax_amount, 0)", "0")
		query = strings.ReplaceAll(query, "s.tax_amount", "0")
		query = strings.ReplaceAll(query, "GROUP BY s.id, s.created_at, s.total_amount, 0", "GROUP BY s.id, s.created_at, s.total_amount")
	}

	var rows []SalesChartData
	queryErr := s.db.SelectContext(ctx, &rows, query)
	if queryErr != nil || len(rows) == 0 {
		// Keep the chart useful on legacy local schemas where optional item-cost
		// columns or joins are incomplete. Sales totals remain authoritative.
		dateExpr := "COALESCE(s.sale_date, s.created_at)"
		if isSQLiteDriver(s.db.DriverName()) && !sqliteHasColumns(s.db, "sales", "sale_date") {
			dateExpr = "s.created_at"
		}
		taxColumn := "0"
		costColumn := "0"
		if !isSQLiteDriver(s.db.DriverName()) || sqliteHasColumns(s.db, "sales", "tax_amount") {
			taxColumn = "COALESCE(s.tax_amount, 0)"
		}
		if !isSQLiteDriver(s.db.DriverName()) || sqliteHasColumns(s.db, "sales", "cost_amount") {
			costColumn = "COALESCE(s.cost_amount, 0)"
		}
		fallbackQuery := fmt.Sprintf(`
			SELECT substr(CAST(%s AS TEXT), 1, 10) AS name,
			       COALESCE(SUM(s.total_amount - %s), 0) AS sales,
			       COALESCE(SUM(s.total_amount - %s - %s), 0) AS profit
			FROM sales s
			WHERE LOWER(COALESCE(s.status, '')) = 'completed'
			  AND %s IS NOT NULL
			GROUP BY substr(CAST(%s AS TEXT), 1, 10)
			ORDER BY name`, dateExpr, taxColumn, taxColumn, costColumn, dateExpr, dateExpr)
		rows = nil
		if err := s.db.SelectContext(ctx, &rows, fallbackQuery); err != nil {
			return []SalesChartData{}
		}
	}
	return rows
}

func (s *CachedService) fetchSQLiteSalesChart(ctx context.Context, chartStart string) []SalesChartData {
	// sale_date is the official commercial day. created_at is only a fallback
	// for legacy local rows that predate the normalized sale date.
	saleDateColumn := "NULL"
	if sqliteHasColumns(s.db, "sales", "sale_date") {
		saleDateColumn = "s.sale_date"
	}
	taxExpr := "0"
	if sqliteHasColumns(s.db, "sales", "tax_amount") {
		taxExpr = "COALESCE(s.tax_amount, 0)"
	}
	itemJoin := ""
	lineCostExpr := "0"
	if sqliteHasColumns(s.db, "sale_items", "sale_id", "quantity") {
		itemJoin = " LEFT JOIN sale_items si ON si.sale_id = s.id"
		unitCostExpr := "0"
		if sqliteHasColumns(s.db, "sale_items", "unit_cost") {
			unitCostExpr = "si.unit_cost"
		}
		productCostExpr := "0"
		productJoin := ""
		if sqliteHasColumns(s.db, "sale_items", "product_id") && sqliteHasColumns(s.db, "products", "id") {
			productCostColumn := "0"
			if sqliteHasColumns(s.db, "products", "cost_price") {
				productCostColumn = "p.cost_price"
			} else if sqliteHasColumns(s.db, "products", "purchase_price") {
				productCostColumn = "p.purchase_price"
			}
			productCostExpr = productCostColumn
			productJoin = " LEFT JOIN products p ON p.id = si.product_id"
		}
		inventoryCostExpr := "0"
		inventoryJoin := ""
		if sqliteHasColumns(s.db, "sale_items", "inventory_item_id") && sqliteHasColumns(s.db, "inventory_items", "id", "purchase_cost") {
			inventoryCostExpr = "ii.purchase_cost"
			inventoryJoin = " LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id"
		}
		lineCostExpr = fmt.Sprintf("COALESCE(SUM(si.quantity * COALESCE(NULLIF(%s, 0), NULLIF(%s, 0), %s, 0)), 0)", unitCostExpr, inventoryCostExpr, productCostExpr)
		itemJoin += productJoin + inventoryJoin
	}
	costExpr := lineCostExpr
	costGroup := ""
	if sqliteHasColumns(s.db, "sales", "cost_amount") {
		costExpr = fmt.Sprintf("CASE WHEN NULLIF(s.cost_amount, 0) IS NOT NULL THEN s.cost_amount ELSE %s END", lineCostExpr)
		costGroup = ", s.cost_amount"
	}

	dateGroupExpr := "NULL"
	if saleDateColumn != "NULL" {
		dateGroupExpr = "s.sale_date"
	}
	groupBy := fmt.Sprintf("s.id, %s, s.created_at, s.total_amount", dateGroupExpr)
	if sqliteHasColumns(s.db, "sales", "tax_amount") {
		groupBy += ", s.tax_amount"
	}
	groupBy += costGroup
	query := fmt.Sprintf(`
		SELECT COALESCE(%s, '') AS sale_date, s.created_at, s.total_amount, %s AS tax_amount, %s AS cost
		FROM sales s%s
		WHERE LOWER(COALESCE(s.status, '')) = 'completed'
		  AND (%s IS NOT NULL OR s.created_at IS NOT NULL)
		GROUP BY %s`, saleDateColumn, taxExpr, costExpr, itemJoin, saleDateColumn, groupBy)
	var rows []struct {
		SaleDate  string  `db:"sale_date"`
		CreatedAt string  `db:"created_at"`
		Sales     float64 `db:"total_amount"`
		Tax       float64 `db:"tax_amount"`
		Cost      float64 `db:"cost"`
	}
	if err := s.db.SelectContext(ctx, &rows, query); err != nil {
		return []SalesChartData{}
	}

	byDate := make(map[string]*SalesChartData)
	for _, row := range rows {
		date := strings.TrimSpace(row.SaleDate)
		if len(date) > len("2006-01-02") {
			date = date[:len("2006-01-02")]
		}
		if date == "" {
			value := strings.TrimSpace(row.CreatedAt)
			parsed, parseErr := time.Parse(time.RFC3339Nano, strings.Replace(value, " ", "T", 1))
			if parseErr != nil {
				parsed, parseErr = time.Parse("2006-01-02 15:04:05", value)
			}
			if parseErr != nil {
				continue
			}
			date, parseErr = accounting.StoreDate(parsed)
			if parseErr != nil {
				continue
			}
		}
		if date < chartStart {
			continue
		}
		point := byDate[date]
		if point == nil {
			point = &SalesChartData{Name: date}
			byDate[date] = point
		}
		point.Sales += row.Sales - row.Tax
		point.Profit += row.Sales - row.Tax - row.Cost
	}

	result := make([]SalesChartData, 0, len(byDate))
	for _, point := range byDate {
		result = append(result, *point)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func (s *CachedService) fetchInventoryDistribution(ctx context.Context) *InventoryDistributionData {
	query := `
		SELECT UPPER(COALESCE(inventory_items.status, 'UNKNOWN')) AS status,
		       UPPER(COALESCE(inventory_items.condition, '')) AS condition,
		       COUNT(*) AS count,
		       COALESCE(SUM(CASE
				WHEN UPPER(COALESCE(inventory_items.status, '')) = 'REVERSED' THEN purchase_cost
				ELSE selling_price
			END), 0) AS value,
		       CASE
				WHEN UPPER(COALESCE(inventory_items.status, 'UNKNOWN')) = 'SOLD' THEN
					COALESCE((SELECT SUM(s.total_amount) FROM sales s
						WHERE LOWER(COALESCE(s.status, 'completed')) = 'completed'), 0)
				ELSE 0
			END AS tax_inclusive_value
		FROM inventory_items
		WHERE UPPER(COALESCE(inventory_items.status, 'UNKNOWN')) <> 'ARCHIVED'
		GROUP BY UPPER(COALESCE(inventory_items.status, 'UNKNOWN')), UPPER(COALESCE(inventory_items.condition, ''))
		ORDER BY status
	`

	var rows []struct {
		Status            string  `db:"status"`
		Condition         string  `db:"condition"`
		Count             int     `db:"count"`
		Value             float64 `db:"value"`
		TaxInclusiveValue float64 `db:"tax_inclusive_value"`
	}
	if err := s.db.SelectContext(ctx, &rows, query); err != nil {
		rows = nil
		fallbackQuery := `
			SELECT UPPER(COALESCE(status, 'UNKNOWN')) AS status,
			       UPPER(COALESCE(condition, '')) AS condition,
			       COUNT(*) AS count,
			       COALESCE(SUM(selling_price), 0) AS value,
			       0 AS tax_inclusive_value
			FROM inventory_items
			WHERE UPPER(COALESCE(status, 'UNKNOWN')) <> 'ARCHIVED'
			GROUP BY UPPER(COALESCE(status, 'UNKNOWN')), UPPER(COALESCE(condition, ''))`
		if fallbackErr := s.db.SelectContext(ctx, &rows, fallbackQuery); fallbackErr != nil {
			return nil
		}
	}

	data := make([]InventoryDistributionItem, 0, len(rows))
	var totalValue float64
	var totalItems int
	for _, row := range rows {
		name, color, health := inventoryStatusPresentation(row.Status, row.Condition)
		data = append(data, InventoryDistributionItem{
			Name:              name,
			Count:             row.Count,
			Value:             row.Value,
			TaxInclusiveValue: row.TaxInclusiveValue,
			Color:             color,
			Status:            health,
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

func inventoryStatusPresentation(status, condition string) (name, color, health string) {
	if strings.EqualFold(strings.TrimSpace(condition), "USED") {
		if strings.EqualFold(strings.TrimSpace(status), "SOLD") {
			return "قطع مستعملة مباعة", "#8b5cf6", "neutral"
		}
		if strings.EqualFold(strings.TrimSpace(status), "AVAILABLE") {
			return "قطع مستعملة متاحة", "#06b6d4", "good"
		}
	}
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "AVAILABLE", "IN_STOCK", "IN STOCK":
		return "متاح", "#10b981", "good"
	case "RESERVED":
		return "محجوز", "#f59e0b", "attention"
	case "RETURNED":
		return "مرتجع", "#f59e0b", "attention"
	case "REVERSED":
		return "شراء ملغى", "#94a3b8", "neutral"
	case "SOLD":
		return "قطع مباعة", "#64748b", "neutral"
	case "DAMAGED", "IN_REPAIR", "IN REPAIR":
		return "تالف/قيد الإصلاح", "#ef4444", "critical"
	case "ARCHIVED":
		return "مؤرشف", "#94a3b8", ""
	default:
		if strings.TrimSpace(status) == "" {
			return "غير محدد", "#94a3b8", "attention"
		}
		return status, "#94a3b8", "attention"
	}
}

func hasTable(ctx context.Context, db *sqlx.DB, tableName string) bool {
	if isSQLiteDriver(db.DriverName()) {
		var count int
		err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, tableName)
		return err == nil && count > 0
	}

	var count int
	err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1`, tableName)
	return err == nil && count > 0
}

func hasColumn(ctx context.Context, db *sqlx.DB, tableName, columnName string) bool {
	if isSQLiteDriver(db.DriverName()) {
		var count int
		err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, tableName, columnName)
		return err == nil && count > 0
	}

	var count int
	err := db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2
	`, tableName, columnName)
	return err == nil && count > 0
}

func (s *CachedService) fetchRecentActivity(ctx context.Context) []RecentActivityItem {
	hasPurchasesTable := hasTable(ctx, s.db, "purchases")
	hasReturnsTable := hasTable(ctx, s.db, "returns")
	hasUsersTable := hasTable(ctx, s.db, "users")
	saleSellerExpr := "'' AS seller_name"
	if hasUsersTable && (s.db.DriverName() != "sqlite" || sqliteHasColumns(s.db, "sales", "user_id")) {
		saleSellerExpr = "COALESCE((SELECT first_name || ' ' || last_name FROM users WHERE users.id = s.user_id), '') AS seller_name"
	}
	query := fmt.Sprintf(`
		SELECT id, type, title, description, amount, activity_time AS time, sale_date, status, seller_name
		FROM (
			SELECT s.id, 'sale' AS type, 'بيع' AS title, 'عملية بيع' AS description,
			       s.total_amount AS amount, datetime(s.created_at) AS activity_time, s.sale_date,
			       %s,
			       s.status
			FROM sales s
			%s
		) AS activity
		ORDER BY activity_time DESC
		LIMIT 5
	`, saleSellerExpr, func() string {
		parts := make([]string, 0, 2)
		if hasPurchasesTable {
			parts = append(parts, `UNION ALL
			SELECT p.id, 'purchase' AS type, 'شراء' AS title, 'عملية شراء' AS description,
			       p.total_amount AS amount, datetime(p.created_at) AS activity_time, '' AS sale_date,
			       '' AS seller_name, p.status
			FROM purchases p`)
		}
		if hasReturnsTable {
			parts = append(parts, `UNION ALL
			SELECT r.id, 'return' AS type, 'مرتجع' AS title, 'استرداد مرتجع' AS description,
			       r.total_refund_amount AS amount, datetime(COALESCE(r.return_date, r.created_at)) AS activity_time, '' AS sale_date,
			       '' AS seller_name, r.status
			FROM returns r
			WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED' AND COALESCE(r.total_refund_amount, 0) >= 0`)
		}
		return strings.Join(parts, "\n")
	}())
	if s.db.DriverName() != "sqlite" {
		query = fmt.Sprintf(`
			SELECT id, type, title, description, amount, activity_time AS time, sale_date, status, seller_name
			FROM (
				SELECT s.id, 'sale' AS type, 'بيع' AS title, 'عملية بيع' AS description,
				       s.total_amount AS amount, TO_CHAR(s.created_at, 'YYYY-MM-DD HH24:MI:SS') AS activity_time, TO_CHAR(s.sale_date, 'YYYY-MM-DD') AS sale_date,
				       %s,
				       s.status
				FROM sales s
				%s
			) AS activity
			ORDER BY activity_time DESC
			LIMIT 5
		`, saleSellerExpr, func() string {
			parts := make([]string, 0, 2)
			if hasPurchasesTable {
				parts = append(parts, `UNION ALL
				SELECT p.id, 'purchase' AS type, 'شراء' AS title, 'عملية شراء' AS description,
				       p.total_amount AS amount, TO_CHAR(p.created_at, 'YYYY-MM-DD HH24:MI:SS') AS activity_time, '' AS sale_date,
				       '' AS seller_name, p.status
				FROM purchases p`)
			}
			if hasReturnsTable {
				parts = append(parts, `UNION ALL
				SELECT r.id, 'return' AS type, 'مرتجع' AS title, 'استرداد مرتجع' AS description,
				       r.total_refund_amount AS amount, TO_CHAR(COALESCE(r.return_date, r.created_at), 'YYYY-MM-DD HH24:MI:SS') AS activity_time, '' AS sale_date,
				       '' AS seller_name, r.status
				FROM returns r
				WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED' AND COALESCE(r.total_refund_amount, 0) >= 0`)
			}
			return strings.Join(parts, "\n")
		}())
	}

	var activities []RecentActivityItem
	if err := s.db.SelectContext(ctx, &activities, query); err != nil {
		return []RecentActivityItem{}
	}
	return activities
}

func (s *CachedService) GetActivity(ctx context.Context, page, perPage int, activityType string) (*ActivityPage, error) {
	return s.GetActivityWithFilters(ctx, page, perPage, activityType, "", "", "")
}

func (s *CachedService) GetActivityWithFilters(ctx context.Context, page, perPage int, activityType, search, startDate, endDate string) (*ActivityPage, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	hasPurchasesTable := hasTable(ctx, s.db, "purchases")
	hasSalesTable := hasTable(ctx, s.db, "sales")
	hasReturnsTable := hasTable(ctx, s.db, "returns")
	hasUsersTable := hasTable(ctx, s.db, "users")
	hasCustomersTable := hasTable(ctx, s.db, "customers")
	saleSellerExpr := "'' AS seller_name"
	if hasSalesTable && hasUsersTable && hasColumn(ctx, s.db, "sales", "user_id") {
		saleSellerExpr = "COALESCE((SELECT first_name || ' ' || last_name FROM users WHERE users.id = s.user_id), '') AS seller_name"
	}
	saleDateExpr := "'' AS sale_date"
	if hasSalesTable && hasColumn(ctx, s.db, "sales", "sale_date") {
		if dbutil.IsSQLite(s.db) {
			saleDateExpr = "COALESCE(date(s.sale_date), '') AS sale_date"
		} else {
			saleDateExpr = "COALESCE(TO_CHAR(s.sale_date, 'YYYY-MM-DD'), '') AS sale_date"
		}
	}
	saleInvoiceExpr := "'' AS invoice_number"
	if hasColumn(ctx, s.db, "sales", "invoice_number") {
		saleInvoiceExpr = "COALESCE(s.invoice_number, '') AS invoice_number"
	}
	saleCustomerIDExpr := "'' AS customer_id"
	saleCustomerNameExpr := "'' AS customer_name"
	if hasColumn(ctx, s.db, "sales", "customer_id") {
		saleCustomerIDExpr = "COALESCE(CAST(s.customer_id AS TEXT), '') AS customer_id"
		if hasCustomersTable {
			saleCustomerNameExpr = "COALESCE((SELECT c.name FROM customers c WHERE c.id = s.customer_id), '') AS customer_name"
		}
	}
	salePaymentExpr := "'' AS payment_method"
	if hasColumn(ctx, s.db, "sales", "payment_method") {
		salePaymentExpr = "COALESCE(s.payment_method, '') AS payment_method"
	}
	salePaidExpr := "0 AS paid_amount"
	if hasColumn(ctx, s.db, "sales", "paid_amount") {
		salePaidExpr = "COALESCE(s.paid_amount, 0) AS paid_amount"
	}
	saleRemainingExpr := "0 AS remaining_amount"
	if hasColumn(ctx, s.db, "sales", "paid_amount") {
		saleRemainingExpr = "CASE WHEN COALESCE(s.total_amount, 0) - COALESCE(s.paid_amount, 0) > 0 THEN COALESCE(s.total_amount, 0) - COALESCE(s.paid_amount, 0) ELSE 0 END AS remaining_amount"
	}
	saleCashExpr := "0 AS cash_received"
	if hasColumn(ctx, s.db, "sales", "cash_received") {
		saleCashExpr = "COALESCE(s.cash_received, 0) AS cash_received"
	}
	saleChangeExpr := "0 AS change_amount"
	if hasColumn(ctx, s.db, "sales", "change_amount") {
		saleChangeExpr = "COALESCE(s.change_amount, 0) AS change_amount"
	}
	purchaseStatusExpr := "'' AS status"
	if hasPurchasesTable && hasColumn(ctx, s.db, "purchases", "status") {
		purchaseStatusExpr = "p.status"
	}

	activityParts := make([]string, 0, 3)
	if hasSalesTable {
		activityParts = append(activityParts, fmt.Sprintf(`
			SELECT s.id, 'sale' AS type, 'بيع' AS title, 'عملية بيع' AS description,
			       s.total_amount AS amount, s.created_at AS activity_time, %s, s.status,
			       %s, %s, %s, %s, %s, %s, %s, %s, %s
			FROM sales s`, saleDateExpr, saleSellerExpr, saleInvoiceExpr, saleCustomerIDExpr, saleCustomerNameExpr, salePaymentExpr, salePaidExpr, saleRemainingExpr, saleCashExpr, saleChangeExpr))
	}
	if hasPurchasesTable {
		activityParts = append(activityParts, fmt.Sprintf(`
		SELECT p.id, 'purchase' AS type, 'شراء' AS title, 'عملية شراء' AS description,
		       p.total_amount AS amount, p.created_at AS activity_time, '' AS sale_date, %s,
		       '' AS seller_name, '' AS invoice_number, '' AS customer_id, '' AS customer_name,
		       '' AS payment_method, 0 AS paid_amount, 0 AS remaining_amount, 0 AS cash_received, 0 AS change_amount
		FROM purchases p`, purchaseStatusExpr))
	}
	if hasReturnsTable {
		returnStatusExpr := "r.status"
		returnTimeExpr := "COALESCE(r.return_date, r.created_at)"
		activityParts = append(activityParts, fmt.Sprintf(`
		SELECT r.id, 'return' AS type, 'مرتجع' AS title, 'استرداد مرتجع' AS description,
		       r.total_refund_amount AS amount, %s AS activity_time, '' AS sale_date, %s,
		       '' AS seller_name, '' AS invoice_number, '' AS customer_id, '' AS customer_name,
		       '' AS payment_method, 0 AS paid_amount, 0 AS remaining_amount, 0 AS cash_received, 0 AS change_amount
		FROM returns r
		WHERE COALESCE(r.total_refund_amount, 0) >= 0`, returnTimeExpr, returnStatusExpr))
	}
	if len(activityParts) == 0 {
		return &ActivityPage{
			Items: []RecentActivityItem{}, Page: page, PerPage: perPage,
		}, nil
	}
	activityQuery := fmt.Sprintf("(%s) AS activity", strings.Join(activityParts, " UNION ALL "))
	where := ""
	args := []interface{}{}
	if activityType == "sale" || activityType == "purchase" || activityType == "return" {
		where = " WHERE type = ?"
		args = append(args, activityType)
	}
	if strings.TrimSpace(search) != "" {
		where += func() string {
			if where == "" {
				return " WHERE"
			}
			return " AND"
		}() + " (LOWER(title) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?) OR CAST(id AS TEXT) LIKE ? OR LOWER(COALESCE(seller_name, '')) LIKE LOWER(?))"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}
	if strings.TrimSpace(startDate) != "" {
		where += func() string {
			if where == "" {
				return " WHERE"
			}
			return " AND"
		}() + " date(activity_time) >= date(?)"
		args = append(args, strings.TrimSpace(startDate))
	}
	if strings.TrimSpace(endDate) != "" {
		where += func() string {
			if where == "" {
				return " WHERE"
			}
			return " AND"
		}() + " date(activity_time) <= date(?)"
		args = append(args, strings.TrimSpace(endDate))
	}

	var total int
	countQuery := s.db.Rebind("SELECT COUNT(*) FROM " + activityQuery + where)
	if err := s.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, fmt.Errorf("failed to count activity: %w", err)
	}

	offset := (page - 1) * perPage
	query := "SELECT id, type, title, description, amount, activity_time AS time, sale_date, status, seller_name, invoice_number, customer_id, customer_name, payment_method, paid_amount, remaining_amount, cash_received, change_amount FROM " + activityQuery + where + " ORDER BY activity_time DESC LIMIT ? OFFSET ?"
	query = s.db.Rebind(query)
	queryArgs := append(args, perPage, offset)
	var items []RecentActivityItem
	if err := s.db.SelectContext(ctx, &items, query, queryArgs...); err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	totalPages := (total + perPage - 1) / perPage
	return &ActivityPage{Items: items, Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, nil
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
			COALESCE(inv.quantity, COUNT(CASE WHEN i.status = 'AVAILABLE' AND i.condition <> 'USED' THEN i.id END)) as quantity,
			p.min_stock_level,
			p.cost_price,
			p.selling_price,
			p.preferred_supplier_id
		FROM products p
		LEFT JOIN inventory inv ON inv.product_id = p.id
		LEFT JOIN inventory_items i ON p.id = i.product_id
		WHERE p.is_active = true
			AND p.deleted_at IS NULL
			AND p.min_stock_level > 0
			AND NOT EXISTS (SELECT 1 FROM inventory_items used_i WHERE used_i.product_id = p.id AND UPPER(COALESCE(used_i.condition, '')) = 'USED')
		GROUP BY p.id, p.name, p.min_stock_level, p.cost_price, p.selling_price, p.preferred_supplier_id, inv.quantity
		HAVING COALESCE(inv.quantity, COUNT(CASE WHEN i.status = 'AVAILABLE' AND COALESCE(i.condition, '') <> 'USED' THEN i.id END)) <= p.min_stock_level
		ORDER BY (p.min_stock_level - COALESCE(inv.quantity, COUNT(CASE WHEN i.status = 'AVAILABLE' AND i.condition <> 'USED' THEN i.id END))) DESC
			LIMIT 5
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
	storeDate, err := accounting.StoreDate(time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to calculate store date: %w", err)
	}
	query := `
		SELECT 
			d.id,
			d.customer_id,
			c.name as customer_name,
			d.remaining_amount,
			d.due_date,
			($1::date - d.due_date::date)::int as days_overdue,
			COALESCE(c.phone, '') as phone
		FROM debts d
		JOIN customers c ON d.customer_id = c.id
		WHERE d.remaining_amount > 0
		AND d.due_date::date < $1::date
		AND d.status IN ('pending', 'partial', 'overdue')
		ORDER BY d.due_date ASC
		LIMIT 10
	`
	if s.db.DriverName() == "sqlite" {
		query = `
			SELECT d.id, d.customer_id, c.name AS customer_name, d.remaining_amount, d.due_date,
			CAST(julianday(?) - julianday(substr(d.due_date, 1, 10)) AS INTEGER) AS days_overdue,
			COALESCE(c.phone, '') AS phone
			FROM debts d JOIN customers c ON d.customer_id = c.id
			WHERE d.remaining_amount > 0
			  AND date(substr(d.due_date, 1, 10)) < date(?)
			  AND ` + business.OpenDebtStatusSQL("d.status") + `
			ORDER BY d.due_date ASC LIMIT 10
		`
	}

	var debts []OverdueDebtItem
	args := []any{storeDate}
	if s.db.DriverName() == "sqlite" {
		args = []any{storeDate, storeDate}
	}
	err = s.db.SelectContext(ctx, &debts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue debts: %w", err)
	}

	return debts, nil
}
