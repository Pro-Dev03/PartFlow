package reports

import (
	"time"

	"github.com/google/uuid"
)

// Report represents a generated report
type Report struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Type           string     `json:"type" db:"type"` // sales, inventory, expenses, purchases, returns, warranties, customers, suppliers
	Title          string     `json:"title" db:"title"`
	Description    string     `json:"description" db:"description"`
	Parameters     string     `json:"parameters" db:"parameters"` // JSON string of report parameters
	Data           string     `json:"data" db:"data"` // JSON string of report data
	Status         string     `json:"status" db:"status"` // pending, completed, failed
	GeneratedBy    uuid.UUID  `json:"generated_by" db:"generated_by"`
	GeneratedAt    time.Time  `json:"generated_at" db:"generated_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// SalesReport represents sales report data
type SalesReport struct {
	Period         string    `json:"period"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	TotalSales     int       `json:"total_sales"`
	TotalRevenue   float64   `json:"total_revenue"`
	TotalCOGS      float64   `json:"total_cogs"`
	GrossProfit    float64   `json:"gross_profit"`
	ProfitMargin   float64   `json:"profit_margin"`
	ByDay          []DailySales `json:"by_day"`
	TopProducts    []ProductSales `json:"top_products"`
	ByPaymentMethod map[string]float64 `json:"by_payment_method"`
}

// DailySales represents daily sales data
type DailySales struct {
	Date    time.Time `json:"date"`
	Sales   int       `json:"sales"`
	Revenue float64   `json:"revenue"`
}

// ProductSales represents product sales data
type ProductSales struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	Revenue     float64   `json:"revenue"`
	Profit      float64   `json:"profit"`
}

// InventoryReport represents inventory report data
type InventoryReport struct {
	TotalItems       int                `json:"total_items"`
	TotalValue       float64            `json:"total_value"`
	ByCondition      map[string]int     `json:"by_condition"`
	ByCategory       map[string]int     `json:"by_category"`
	LowStockItems    []LowStockItem     `json:"low_stock_items"`
	OverstockItems   []OverstockItem    `json:"overstock_items"`
	StagnantItems    []StagnantItem     `json:"stagnant_items"`
	Valuation        InventoryValuation `json:"valuation"`
}

// LowStockItem represents low stock item
type LowStockItem struct {
	ProductID     uuid.UUID `json:"product_id"`
	ProductName   string    `json:"product_name"`
	CurrentStock  int       `json:"current_stock"`
	MinStock      int       `json:"min_stock"`
	ReorderLevel  int       `json:"reorder_level"`
}

// OverstockItem represents overstock item
type OverstockItem struct {
	ProductID    uuid.UUID `json:"product_id"`
	ProductName  string    `json:"product_name"`
	CurrentStock int       `json:"current_stock"`
	AvgMonthlySales int     `json:"avg_monthly_sales"`
	MonthsOfSupply int      `json:"months_of_supply"`
}

// StagnantItem represents stagnant item
type StagnantItem struct {
	ProductID     uuid.UUID `json:"product_id"`
	ProductName   string    `json:"product_name"`
	CurrentStock  int       `json:"current_stock"`
	LastSaleDate  time.Time `json:"last_sale_date"`
	DaysSinceSale int       `json:"days_since_sale"`
	Value         float64   `json:"value"`
}

// InventoryValuation represents inventory valuation
type InventoryValuation struct {
	TotalCost      float64 `json:"total_cost"`
	TotalRetail    float64 `json:"total_retail"`
	PotentialProfit float64 `json:"potential_profit"`
	ByCondition    map[string]float64 `json:"by_condition"`
}

// ExpensesReport represents expenses report data
type ExpensesReport struct {
	Period          string             `json:"period"`
	StartDate       time.Time          `json:"start_date"`
	EndDate         time.Time          `json:"end_date"`
	TotalExpenses   float64            `json:"total_expenses"`
	ByCategory      map[string]float64 `json:"by_category"`
	ByPaymentMethod map[string]float64 `json:"by_payment_method"`
	ByMonth         []MonthlyExpenses  `json:"by_month"`
	TopExpenses     []ExpenseItem      `json:"top_expenses"`
}

// MonthlyExpenses represents monthly expenses data
type MonthlyExpenses struct {
	Month   time.Time `json:"month"`
	Amount  float64   `json:"amount"`
}

// ExpenseItem represents expense item data
type ExpenseItem struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Amount       float64   `json:"amount"`
	Count        int       `json:"count"`
}

// ProfitsReport represents profits report data
type ProfitsReport struct {
	Period            string    `json:"period"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	TotalRevenue      float64   `json:"total_revenue"`
	TotalCOGS         float64   `json:"total_cogs"`
	GrossProfit       float64   `json:"gross_profit"`
	TotalExpenses     float64   `json:"total_expenses"`
	NetProfit         float64   `json:"net_profit"`
	ProfitMargin      float64   `json:"profit_margin"`
	ByMonth           []MonthlyProfit `json:"by_month"`
	ByCategory        map[string]float64 `json:"by_category"`
}

// MonthlyProfit represents monthly profit data
type MonthlyProfit struct {
	Month       time.Time `json:"month"`
	Revenue     float64   `json:"revenue"`
	COGS        float64   `json:"cogs"`
	GrossProfit float64   `json:"gross_profit"`
	Expenses    float64   `json:"expenses"`
	NetProfit   float64   `json:"net_profit"`
}

// DebtsReport represents debts report data
type DebtsReport struct {
	TotalDebt           float64         `json:"total_debt"`
	TotalPaid           float64         `json:"total_paid"`
	Outstanding         float64         `json:"outstanding"`
	OverdueDebt         float64         `json:"overdue_debt"`
	ByCustomer          []CustomerDebt  `json:"by_customer"`
	ByAge               map[string]int  `json:"by_age"`
	PaymentHistory      []PaymentRecord `json:"payment_history"`
}

// CustomerDebt represents customer debt data
type CustomerDebt struct {
	CustomerID    uuid.UUID `json:"customer_id"`
	CustomerName  string    `json:"customer_name"`
	TotalDebt     float64   `json:"total_debt"`
	PaidAmount    float64   `json:"paid_amount"`
	Outstanding   float64   `json:"outstanding"`
	OverdueAmount float64   `json:"overdue_amount"`
	LastPayment   time.Time `json:"last_payment"`
}

// PaymentRecord represents payment record
type PaymentRecord struct {
	Date        time.Time `json:"date"`
	CustomerID  uuid.UUID `json:"customer_id"`
	CustomerName string   `json:"customer_name"`
	Amount      float64   `json:"amount"`
}

// ReportRequest represents report generation request
type ReportRequest struct {
	Type        string                 `json:"type" binding:"required,oneof=sales inventory expenses profits debts purchases returns warranties"`
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ReportListRequest represents report list query parameters
type ReportListRequest struct {
	Page         int        `form:"page" binding:"min=1"`
	PerPage      int        `form:"per_page" binding:"min=1,max=100"`
	Type         string     `form:"type" binding:"omitempty,oneof=sales inventory expenses profits debts purchases returns warranties"`
	Status       string     `form:"status" binding:"omitempty,oneof=pending completed failed"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
	GeneratedBy  *uuid.UUID `form:"generated_by"`
	Search       string     `form:"search"`
	SortBy       string     `form:"sort_by"`
	SortOrder    string     `form:"sort_order"`
}
