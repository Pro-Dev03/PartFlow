package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
}

// Alert represents a dashboard alert
type Alert struct {
	Type        string      `json:"type"` // LOW_STOCK, OVERDUE_DEBT, WARRANTY_EXPIRING
	Title       string      `json:"title"`
	Message     string      `json:"message"`
	Severity    string      `json:"severity"` // high, medium, low
	ActionURL   string      `json:"action_url,omitempty"`
	Data        interface{} `json:"data,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

// GetDashboardStats retrieves dashboard statistics for an organization
func (s *Service) GetDashboardStats(ctx context.Context, organizationID uuid.UUID) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Get total sales
	query := `SELECT COALESCE(SUM(total_amount), 0) FROM sales WHERE organization_id = $1 AND status = 'completed'`
	err := s.db.GetContext(ctx, &stats.TotalSales, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total sales: %w", err)
	}

	// Get total purchases
	query = `SELECT COALESCE(SUM(total_amount), 0) FROM purchases WHERE organization_id = $1 AND status = 'received'`
	err = s.db.GetContext(ctx, &stats.TotalPurchases, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total purchases: %w", err)
	}

	// Get total expenses
	query = `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE organization_id = $1 AND status = 'approved'`
	err = s.db.GetContext(ctx, &stats.TotalExpenses, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total expenses: %w", err)
	}

	// Get total products
	query = `SELECT COUNT(*) FROM products WHERE organization_id = $1 AND is_active = true`
	err = s.db.GetContext(ctx, &stats.TotalProducts, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total products: %w", err)
	}

	// Get total customers
	query = `SELECT COUNT(*) FROM customers WHERE organization_id = $1 AND is_active = true`
	err = s.db.GetContext(ctx, &stats.TotalCustomers, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total customers: %w", err)
	}

	// Get total suppliers
	query = `SELECT COUNT(*) FROM suppliers WHERE organization_id = $1 AND is_active = true`
	err = s.db.GetContext(ctx, &stats.TotalSuppliers, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total suppliers: %w", err)
	}

	// Get pending orders (sales)
	query = `SELECT COUNT(*) FROM sales WHERE organization_id = $1 AND status = 'pending'`
	err = s.db.GetContext(ctx, &stats.PendingOrders, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}

	// Get low stock items
	query = `
		SELECT COUNT(*) FROM products p
		WHERE p.organization_id = $1
		AND p.is_active = true
		AND (SELECT COUNT(*) FROM inventory_items WHERE product_id = p.id AND status = 'AVAILABLE') < p.min_stock_level
	`
	err = s.db.GetContext(ctx, &stats.LowStockItems, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock items: %w", err)
	}

	// Get overdue debts
	query = `
		SELECT COALESCE(SUM(current_balance), 0) FROM customers
		WHERE organization_id = $1
		AND current_balance > 0
		AND id IN (
			SELECT DISTINCT customer_id FROM customer_ledger
			WHERE type = 'debit'
			AND created_at < NOW() - INTERVAL '30 days'
		)
	`
	err = s.db.GetContext(ctx, &stats.OverdueDebts, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue debts: %w", err)
	}

	// Get pending returns
	query = `SELECT COUNT(*) FROM returns WHERE organization_id = $1 AND status = 'pending'`
	err = s.db.GetContext(ctx, &stats.PendingReturns, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending returns: %w", err)
	}

	// Get pending warranty claims
	query = `SELECT COUNT(*) FROM warranty_claims WHERE organization_id = $1 AND status = 'pending'`
	err = s.db.GetContext(ctx, &stats.PendingClaims, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending claims: %w", err)
	}

	// Calculate totals
	stats.TotalRevenue = stats.TotalSales
	stats.TotalProfit = stats.TotalSales - stats.TotalPurchases - stats.TotalExpenses

	// Fetch alerts
	alerts, err := s.getAlerts(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	stats.Alerts = alerts

	return stats, nil
}

// getAlerts retrieves all active alerts for an organization
func (s *Service) getAlerts(ctx context.Context, organizationID uuid.UUID) ([]Alert, error) {
	var alerts []Alert

	// Get overdue debt alerts
	overdueQuery := `
		SELECT 
			d.id,
			d.customer_id,
			c.name as customer_name,
			d.remaining_amount,
			d.due_date,
			EXTRACT(DAY FROM (NOW() - d.due_date)) as days_overdue
		FROM debts d
		JOIN customers c ON d.customer_id = c.id
		WHERE d.organization_id = $1
		AND d.status = 'overdue'
		ORDER BY d.due_date ASC
		LIMIT 10
	`

	var overdueDebts []struct {
		ID           uuid.UUID `db:"id"`
		CustomerID   uuid.UUID `db:"customer_id"`
		CustomerName string    `db:"customer_name"`
		Amount       float64   `db:"remaining_amount"`
		DueDate      time.Time `db:"due_date"`
		DaysOverdue  int       `db:"days_overdue"`
	}

	err := s.db.SelectContext(ctx, &overdueDebts, overdueQuery, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch overdue debts: %w", err)
	}

	for _, debt := range overdueDebts {
		alert := Alert{
			Type:      "OVERDUE_DEBT",
			Title:     "دين متأخر",
			Message:   fmt.Sprintf("العميل %s لديه دين متأخر بقيمة %.2f ₪ منذ %d أيام", debt.CustomerName, debt.Amount, debt.DaysOverdue),
			Severity:  "high",
			ActionURL: fmt.Sprintf("/debts/%s", debt.ID),
			Data: map[string]interface{}{
				"debt_id":       debt.ID,
				"customer_id":   debt.CustomerID,
				"customer_name": debt.CustomerName,
				"amount":        debt.Amount,
				"days_overdue":  debt.DaysOverdue,
				"due_date":      debt.DueDate,
			},
			CreatedAt: debt.DueDate,
		}
		alerts = append(alerts, alert)
	}

	// Get low stock alerts
	lowStockQuery := `
		SELECT 
			p.id,
			p.name,
			p.min_stock_level,
			COALESCE(SUM(ii.quantity), 0) as current_stock
		FROM products p
		LEFT JOIN inventory_items ii ON p.id = ii.product_id AND ii.status = 'AVAILABLE'
		WHERE p.organization_id = $1
		AND p.is_active = true
		GROUP BY p.id, p.name, p.min_stock_level
		HAVING COALESCE(SUM(ii.quantity), 0) <= p.min_stock_level
		ORDER BY COALESCE(SUM(ii.quantity), 0) ASC
		LIMIT 10
	`

	var lowStockItems []struct {
		ID            uuid.UUID `db:"id"`
		Name          string    `db:"name"`
		MinStockLevel int       `db:"min_stock_level"`
		CurrentStock  int       `db:"current_stock"`
	}

	err = s.db.SelectContext(ctx, &lowStockItems, lowStockQuery, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch low stock items: %w", err)
	}

	for _, item := range lowStockItems {
		alert := Alert{
			Type:      "LOW_STOCK",
			Title:     "مخزون منخفض",
			Message:   fmt.Sprintf("المنتج %s لديه مخزون منخفض (%d من %d)", item.Name, item.CurrentStock, item.MinStockLevel),
			Severity:  "medium",
			ActionURL: fmt.Sprintf("/products/%s", item.ID),
			Data: map[string]interface{}{
				"product_id":     item.ID,
				"product_name":   item.Name,
				"current_stock":  item.CurrentStock,
				"min_stock_level": item.MinStockLevel,
			},
			CreatedAt: time.Now(),
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}