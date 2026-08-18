package dashboard

import (
	"time"

	"github.com/google/uuid"
)

// DashboardData represents aggregated dashboard data
type DashboardData struct {
	OrganizationID uuid.UUID        `json:"organization_id"`
	Date           time.Time        `json:"date"`
	
	// Sales metrics
	TodaySales     MoneySummary     `json:"today_sales"`
	WeekSales      MoneySummary     `json:"week_sales"`
	MonthSales     MoneySummary     `json:"month_sales"`
	
	// Profit metrics
	TodayProfit    MoneySummary     `json:"today_profit"`
	WeekProfit     MoneySummary     `json:"week_profit"`
	MonthProfit    MoneySummary     `json:"month_profit"`
	
	// Inventory metrics
	InventoryValue MoneySummary     `json:"inventory_value"`
	LowStockCount  int              `json:"low_stock_count"`
	TotalProducts  int              `json:"total_products"`
	
	// Debt metrics
	OutstandingDebt MoneySummary    `json:"outstanding_debt"`
	OverdueDebt    MoneySummary     `json:"overdue_debt"`
	OverdueCount   int              `json:"overdue_count"`
	
	// Alerts
	LowStockAlerts    []LowStockAlert    `json:"low_stock_alerts"`
	OverdueAlerts     []OverdueAlert     `json:"overdue_alerts"`
	WarrantyAlerts    []WarrantyAlert    `json:"warranty_alerts"`
	
	// Top performers
	TopProducts       []TopProduct       `json:"top_products"`
	TopCustomers      []TopCustomer      `json:"top_customers"`
	
	// Insights
	Insights          []Insight          `json:"insights"`
}

// MoneySummary represents a monetary summary
type MoneySummary struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// LowStockAlert represents a low stock alert
type LowStockAlert struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	CurrentStock int      `json:"current_stock"`
	MinStockLevel int     `json:"min_stock_level"`
}

// OverdueAlert represents an overdue debt alert
type OverdueAlert struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Amount       int64     `json:"amount"`
	DaysOverdue  int       `json:"days_overdue"`
	DueDate      time.Time `json:"due_date"`
}

// WarrantyAlert represents a warranty expiring alert
type WarrantyAlert struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	CustomerID   uuid.UUID `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	ExpiresAt   time.Time `json:"expires_at"`
	DaysRemaining int    `json:"days_remaining"`
}

// TopProduct represents a top performing product
type TopProduct struct {
	ProductID    uuid.UUID `json:"product_id"`
	ProductName  string    `json:"product_name"`
	SalesCount   int       `json:"sales_count"`
	Revenue      int64     `json:"revenue"`
	Profit       int64     `json:"profit"`
}

// TopCustomer represents a top customer
type TopCustomer struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	PurchaseCount int      `json:"purchase_count"`
	TotalSpent   int64     `json:"total_spent"`
}

// Insight represents a business insight
type Insight struct {
	Type        string      `json:"type"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Priority    string      `json:"priority"` // high, medium, low
	CreatedAt   time.Time   `json:"created_at"`
}
