package aggregations

import "time"

// DailySalesSummary represents daily sales aggregation (ARCHITECTURE-PRINCIPLES.md)
type DailySalesSummary struct {
	Date              string  `json:"date" db:"date"`
	TotalSales        int     `json:"total_sales" db:"total_sales"`
	TotalRevenue      float64 `json:"total_revenue" db:"total_revenue"`
	TotalProfit       float64 `json:"total_profit" db:"total_profit"`
	TotalCustomers    int     `json:"total_customers" db:"total_customers"`
	AverageOrderValue float64 `json:"average_order_value" db:"average_order_value"`
	TotalItemsSold    int     `json:"total_items_sold" db:"total_items_sold"`
	CashSales         float64 `json:"cash_sales" db:"cash_sales"`
	CardSales         float64 `json:"card_sales" db:"card_sales"`
	DebtSales         float64 `json:"debt_sales" db:"debt_sales"`
	UpdatedAt         string  `json:"updated_at" db:"updated_at"`
}

// MonthlySalesSummary represents monthly sales aggregation (ARCHITECTURE-PRINCIPLES.md)
type MonthlySalesSummary struct {
	Year              int     `json:"year" db:"year"`
	Month             int     `json:"month" db:"month"`
	TotalSales        int     `json:"total_sales" db:"total_sales"`
	TotalRevenue      float64 `json:"total_revenue" db:"total_revenue"`
	TotalProfit       float64 `json:"total_profit" db:"total_profit"`
	TotalCustomers    int     `json:"total_customers" db:"total_customers"`
	AverageOrderValue float64 `json:"average_order_value" db:"average_order_value"`
	TotalItemsSold    int     `json:"total_items_sold" db:"total_items_sold"`
	CashSales         float64 `json:"cash_sales" db:"cash_sales"`
	CardSales         float64 `json:"card_sales" db:"card_sales"`
	DebtSales         float64 `json:"debt_sales" db:"debt_sales"`
	UpdatedAt         string  `json:"updated_at" db:"updated_at"`
}

// DailyInventorySummary represents daily inventory aggregation (ARCHITECTURE-PRINCIPLES.md)
type DailyInventorySummary struct {
	Date            time.Time `json:"date"`
	TotalItems      int       `json:"total_items"`
	TotalValue      float64   `json:"total_value"`
	LowStockCount   int       `json:"low_stock_count"`
	OutOfStockCount int       `json:"out_of_stock_count"`
	NewItemsAdded   int       `json:"new_items_added"`
	ItemsSold       int       `json:"items_sold"`
	ItemsReturned   int       `json:"items_returned"`
	ItemsDamaged    int       `json:"items_damaged"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// MonthlyInventorySummary represents monthly inventory aggregation (ARCHITECTURE-PRINCIPLES.md)
type MonthlyInventorySummary struct {
	Year            int       `json:"year"`
	Month           int       `json:"month"`
	TotalItems      int       `json:"total_items"`
	TotalValue      float64   `json:"total_value"`
	LowStockCount   int       `json:"low_stock_count"`
	OutOfStockCount int       `json:"out_of_stock_count"`
	NewItemsAdded   int       `json:"new_items_added"`
	ItemsSold       int       `json:"items_sold"`
	ItemsReturned   int       `json:"items_returned"`
	ItemsDamaged    int       `json:"items_damaged"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DailyDebtSummary represents daily debt aggregation (ARCHITECTURE-PRINCIPLES.md)
type DailyDebtSummary struct {
	Date             time.Time `json:"date"`
	TotalDebt        float64   `json:"total_debt"`
	NewDebt          float64   `json:"new_debt"`
	PaymentsReceived float64   `json:"payments_received"`
	OverdueDebt      float64   `json:"overdue_debt"`
	OverdueCount     int       `json:"overdue_count"`
	PaidDebt         float64   `json:"paid_debt"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// MonthlyDebtSummary represents monthly debt aggregation (ARCHITECTURE-PRINCIPLES.md)
type MonthlyDebtSummary struct {
	Year             int       `json:"year"`
	Month            int       `json:"month"`
	TotalDebt        float64   `json:"total_debt"`
	NewDebt          float64   `json:"new_debt"`
	PaymentsReceived float64   `json:"payments_received"`
	OverdueDebt      float64   `json:"overdue_debt"`
	OverdueCount     int       `json:"overdue_count"`
	PaidDebt         float64   `json:"paid_debt"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// DailyProfitSummary represents daily profit aggregation (ARCHITECTURE-PRINCIPLES.md)
type DailyProfitSummary struct {
	Date         time.Time `json:"date"`
	GrossProfit  float64   `json:"gross_profit"`
	NetProfit    float64   `json:"net_profit"`
	TotalRevenue float64   `json:"total_revenue"`
	TotalCost    float64   `json:"total_cost"`
	ProfitMargin float64   `json:"profit_margin"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MonthlyProfitSummary represents monthly profit aggregation (ARCHITECTURE-PRINCIPLES.md)
type MonthlyProfitSummary struct {
	Year         int       `json:"year"`
	Month        int       `json:"month"`
	GrossProfit  float64   `json:"gross_profit"`
	NetProfit    float64   `json:"net_profit"`
	TotalRevenue float64   `json:"total_revenue"`
	TotalCost    float64   `json:"total_cost"`
	ProfitMargin float64   `json:"profit_margin"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AggregationStatus represents the status of aggregation tables (ARCHITECTURE-PRINCIPLES.md)
type AggregationStatus struct {
	LastDailySalesUpdate       time.Time `json:"last_daily_sales_update"`
	LastMonthlySalesUpdate     time.Time `json:"last_monthly_sales_update"`
	LastDailyInventoryUpdate   time.Time `json:"last_daily_inventory_update"`
	LastMonthlyInventoryUpdate time.Time `json:"last_monthly_inventory_update"`
	LastDailyDebtUpdate        time.Time `json:"last_daily_debt_update"`
	LastMonthlyDebtUpdate      time.Time `json:"last_monthly_debt_update"`
	LastDailyProfitUpdate      time.Time `json:"last_daily_profit_update"`
	LastMonthlyProfitUpdate    time.Time `json:"last_monthly_profit_update"`
	IsProcessing               bool      `json:"is_processing"`
	ProcessingSince            time.Time `json:"processing_since"`
}
