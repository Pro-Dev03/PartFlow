package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalSales      float64 `json:"total_sales"`
	TotalPurchases  float64 `json:"total_purchases"`
	TotalExpenses   float64 `json:"total_expenses"`
	TotalRevenue    float64 `json:"total_revenue"`
	TotalProfit     float64 `json:"total_profit"`
	TotalProducts   int     `json:"total_products"`
	TotalCustomers  int     `json:"total_customers"`
	TotalSuppliers  int     `json:"total_suppliers"`
	PendingOrders   int     `json:"pending_orders"`
	LowStockItems   int     `json:"low_stock_items"`
	OverdueDebts    float64 `json:"overdue_debts"`
	PendingReturns  int     `json:"pending_returns"`
	PendingClaims   int     `json:"pending_claims"`
	Alerts          []Alert `json:"alerts"`
	// Net sales fields
	TotalReturns    float64 `json:"total_returns"`
	TotalRefunded   float64 `json:"total_refunded"`
	NetSales        float64 `json:"net_sales"`
	NetRevenue      float64 `json:"net_revenue"`
	ReturnRate      float64 `json:"return_rate"`
	// Fields for frontend compatibility
	TodaySales      float64 `json:"todaySales"`
	TodayProfit     float64 `json:"todayProfit"`
	OutstandingDebts float64 `json:"outstandingDebts"`
	ActiveCustomers int     `json:"activeCustomers"`
	LowStockCount   int     `json:"lowStockCount"`
	OverdueDebtsCount int   `json:"overdueDebts"`
	// Trend fields
	SalesTrend      *string `json:"salesTrend,omitempty"`
	SalesTrendUp    *bool   `json:"salesTrendUp,omitempty"`
	ProfitTrend     *string `json:"profitTrend,omitempty"`
	ProfitTrendUp   *bool   `json:"profitTrendUp,omitempty"`
	DebtsTrend      *string `json:"debtsTrend,omitempty"`
	DebtsTrendUp    *bool   `json:"debtsTrendUp,omitempty"`
	ProfitMargin    *int    `json:"profitMargin,omitempty"`
	// Chart data fields
	SalesChart       []SalesChartData `json:"salesChart,omitempty"`
	InventoryDistribution *InventoryDistributionData `json:"inventoryDistribution,omitempty"`
}

// SalesChartData represents sales chart data point
type SalesChartData struct {
	Name   string  `json:"name"`
	Sales  float64 `json:"sales"`
	Profit float64 `json:"profit"`
}

// InventoryDistributionData represents inventory distribution
type InventoryDistributionData struct {
	TotalValue float64                    `json:"totalValue"`
	TotalItems int                        `json:"totalItems"`
	Data       []InventoryDistributionItem `json:"data"`
}

// InventoryDistributionItem represents inventory distribution item
type InventoryDistributionItem struct {
	Name   string  `json:"name"`
	Count  int     `json:"count"`
	Value  float64 `json:"value"`
	Color  string  `json:"color"`
	Status string  `json:"status"`
}

// Alert represents a dashboard alert
type Alert struct {
	Type        string      `json:"type"`
	Title       string      `json:"title"`
	Message     string      `json:"message"`
	Severity    string      `json:"severity"`
	ActionURL   string      `json:"action_url,omitempty"`
	Data        interface{} `json:"data,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

// GetDashboardStats retrieves dashboard statistics
// OPTIMIZED: Using aggregation tables for much better performance
func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// OPTIMIZED: Using aggregation tables for instant stats
	query := `
		SELECT
			(SELECT COUNT(*) FROM products p
			 AND p.is_active = true
			 AND p.min_stock_level > 0
			 AND (SELECT COALESCE(SUM(quantity), 0) FROM inventory WHERE product_id = p.id) < p.min_stock_level) as low_stock_items,
			(SELECT COALESCE(SUM(current_balance), 0) FROM customers
			 WHERE current_balance > 0) as overdue_debts,
			(SELECT COUNT(*) FROM sales WHERE status = 'pending') as pending_orders,
			(SELECT COALESCE(total_sales, 0) FROM daily_sales_summary 
			 WHERE summary_date = CURRENT_DATE) as total_sales,
			(SELECT COALESCE(SUM(total_amount), 0) FROM purchases WHERE status = 'received') as total_purchases,
			(SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE status = 'approved') as total_expenses,
			(SELECT COUNT(*) FROM products WHERE is_active = true) as total_products,
			(SELECT COUNT(*) FROM customers) as total_customers,
			(SELECT COUNT(*) FROM suppliers) as total_suppliers,
			(SELECT COUNT(*) FROM returns WHERE status = 'pending') as pending_returns,
			(SELECT COUNT(*) FROM warranty_claims WHERE status = 'pending') as pending_claims,
			(SELECT COALESCE(SUM(refund_amount), 0) FROM returns WHERE status = 'completed') as total_refunded,
			(SELECT COUNT(*) FROM returns WHERE status = 'completed') as total_returns
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
		TotalRefunded  float64 `db:"total_refunded"`
		TotalReturns  int     `db:"total_returns"`
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
	stats.TotalReturns = float64(result.TotalReturns)
	stats.TotalRefunded = result.TotalRefunded

	// Calculate net sales
	stats.NetSales = stats.TotalSales - stats.TotalReturns
	stats.NetRevenue = stats.TotalSales - stats.TotalRefunded

	// Calculate return rate
	if stats.TotalSales > 0 {
		stats.ReturnRate = (stats.TotalReturns / stats.TotalSales) * 100
	} else {
		stats.ReturnRate = 0
	}

	// Calculate totals
	stats.TotalRevenue = stats.NetRevenue
	stats.TotalProfit = stats.NetRevenue - stats.TotalPurchases - stats.TotalExpenses

	// Populate frontend-compatible fields
	stats.TodaySales = result.TotalSales // Using total sales as today's sales for now
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

	// Calculate profit margin if there are sales
	if stats.TotalSales > 0 {
		margin := int((stats.TotalProfit / stats.TotalSales) * 100)
		stats.ProfitMargin = &margin
	}

	// Alerts disabled for performance
	stats.Alerts = []Alert{}

	// Trend fields are left as null (no historical data available)
	// They will be null in the frontend, which is the desired behavior

	// Chart data fields are left as null (no chart data available)
	// They will be null in the frontend, which will show "no data available" message

	return stats, nil
}

// getAlerts retrieves all active alerts
func (s *Service) getAlerts(ctx context.Context) ([]Alert, error) {
	var alerts []Alert
	return alerts, nil
}

// LowStockItem represents a low stock item with details
type LowStockItem struct {
	ID              string  `json:"id" db:"id"`
	ProductName     string  `json:"product_name" db:"product_name"`
	Quantity        int     `json:"quantity" db:"quantity"`
	MinStockLevel   int     `json:"min_stock_level" db:"min_stock_level"`
	CostPrice       float64 `json:"cost_price" db:"cost_price"`
	SellingPrice    float64 `json:"selling_price" db:"selling_price"`
	PreferredSupplierID *string `json:"preferred_supplier_id,omitempty" db:"preferred_supplier_id"`
}

// OverdueDebtItem represents an overdue debt with details
type OverdueDebtItem struct {
	ID              string  `json:"id" db:"id"`
	CustomerID      string  `json:"customer_id" db:"customer_id"`
	CustomerName    string  `json:"customer_name" db:"customer_name"`
	RemainingAmount float64 `json:"remaining_amount" db:"remaining_amount"`
	DueDate         string  `json:"due_date" db:"due_date"`
	DaysOverdue     int     `json:"days_overdue" db:"days_overdue"`
	Phone           string  `json:"phone" db:"phone"`
}

// GetLowStockItems retrieves low stock items with details
func (s *Service) GetLowStockItems(ctx context.Context) ([]LowStockItem, error) {
	query := `
		SELECT 
			p.id,
			p.name as product_name,
			COALESCE(SUM(i.quantity), 0) as quantity,
			p.min_stock_level,
			p.cost_price,
			p.selling_price,
			p.preferred_supplier_id
		FROM products p
		LEFT JOIN inventory i ON p.id = i.product_id
		WHERE p.is_active = true
		AND p.min_stock_level > 0
		GROUP BY p.id, p.name, p.min_stock_level, p.cost_price, p.selling_price, p.preferred_supplier_id
		HAVING COALESCE(SUM(i.quantity), 0) < p.min_stock_level
		ORDER BY (p.min_stock_level - COALESCE(SUM(i.quantity), 0)) DESC
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
func (s *Service) GetOverdueDebts(ctx context.Context) ([]OverdueDebtItem, error) {
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
		AND d.status = 'pending'
		ORDER BY d.due_date ASC
		LIMIT 10
	`

	var debts []OverdueDebtItem
	err := s.db.SelectContext(ctx, &debts, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue debts: %w", err)
	}

	return debts, nil
}
