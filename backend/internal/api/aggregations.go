package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/aggregations"
)

// AggregationHandler handles aggregation endpoints (ARCHITECTURE-PRINCIPLES.md)
type AggregationHandler struct {
	db *sqlx.DB
}

// NewAggregationHandler creates a new aggregation handler
func NewAggregationHandler(db *sqlx.DB) *AggregationHandler {
	return &AggregationHandler{db: db}
}

func buildDailySalesSummaryQuery() string {
	return `
		SELECT
			? AS date,
			COALESCE(s.total_sales, 0) AS total_sales,
			COALESCE(s.total_revenue, 0) AS total_revenue,
			COALESCE(s.total_profit, 0) AS total_profit,
			COALESCE(s.total_customers, 0) AS total_customers,
			COALESCE(s.average_order_value, 0) AS average_order_value,
			COALESCE(s.total_items_sold, 0) AS total_items_sold,
			COALESCE(s.cash_sales, 0) AS cash_sales,
			COALESCE(s.card_sales, 0) AS card_sales,
			COALESCE(s.debt_sales, 0) AS debt_sales,
			COALESCE(s.updated_at, datetime('now')) AS updated_at
		FROM (SELECT ? AS summary_date) dates
		LEFT JOIN daily_sales_summary s ON s.date = dates.summary_date
	`
}

func buildMonthlySalesSummaryQuery() string {
	return `
		SELECT
			? AS year,
			? AS month,
			COALESCE(s.total_sales, 0) AS total_sales,
			COALESCE(s.total_revenue, 0) AS total_revenue,
			COALESCE(s.total_profit, 0) AS total_profit,
			COALESCE(s.total_customers, 0) AS total_customers,
			COALESCE(s.average_order_value, 0) AS average_order_value,
			COALESCE(s.total_items_sold, 0) AS total_items_sold,
			COALESCE(s.cash_sales, 0) AS cash_sales,
			COALESCE(s.card_sales, 0) AS card_sales,
			COALESCE(s.debt_sales, 0) AS debt_sales,
			COALESCE(s.updated_at, datetime('now')) AS updated_at
		FROM (SELECT ? AS summary_year, ? AS summary_month) dates
		LEFT JOIN monthly_sales_summary s ON s.year = dates.summary_year AND s.month = dates.summary_month
	`
}

// GetDailySalesSummary returns daily sales summary
func (h *AggregationHandler) GetDailySalesSummary(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	summary := aggregations.DailySalesSummary{
		Date: date.Format("2006-01-02"),
	}
	dateValue := date.Format("2006-01-02")
	err = h.db.GetContext(c.Request.Context(), &summary, buildDailySalesSummaryQuery(), dateValue, dateValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve daily sales summary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetMonthlySalesSummary returns monthly sales summary
func (h *AggregationHandler) GetMonthlySalesSummary(c *gin.Context) {
	now := time.Now()
	year, month := now.Year(), int(now.Month())
	if value := c.Query("year"); value != "" {
		if _, err := fmt.Sscanf(value, "%d", &year); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
			return
		}
	}
	if value := c.Query("month"); value != "" {
		if _, err := fmt.Sscanf(value, "%d", &month); err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month"})
			return
		}
	}
	summary := aggregations.MonthlySalesSummary{Year: year, Month: month}
	err := h.db.GetContext(c.Request.Context(), &summary, buildMonthlySalesSummaryQuery(), year, month, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve monthly sales summary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetDailyInventorySummary returns daily inventory summary
func (h *AggregationHandler) GetDailyInventorySummary(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// TODO: Implement actual database query
	summary := aggregations.DailyInventorySummary{
		Date:            date,
		TotalItems:      500,
		TotalValue:      250000.00,
		LowStockCount:   12,
		OutOfStockCount: 3,
		NewItemsAdded:   20,
		ItemsSold:       15,
		ItemsReturned:   2,
		ItemsDamaged:    1,
		UpdatedAt:       time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetMonthlyInventorySummary returns monthly inventory summary
func (h *AggregationHandler) GetMonthlyInventorySummary(c *gin.Context) {
	// TODO: Implement actual database query
	summary := aggregations.MonthlyInventorySummary{
		Year:            2026,
		Month:           8,
		TotalItems:      500,
		TotalValue:      250000.00,
		LowStockCount:   12,
		OutOfStockCount: 3,
		NewItemsAdded:   100,
		ItemsSold:       80,
		ItemsReturned:   10,
		ItemsDamaged:    5,
		UpdatedAt:       time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetDailyDebtSummary returns daily debt summary
func (h *AggregationHandler) GetDailyDebtSummary(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// TODO: Implement actual database query
	summary := aggregations.DailyDebtSummary{
		Date:             date,
		TotalDebt:        15000.00,
		NewDebt:          2000.00,
		PaymentsReceived: 1500.00,
		OverdueDebt:      3000.00,
		OverdueCount:     5,
		PaidDebt:         1500.00,
		UpdatedAt:        time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetMonthlyDebtSummary returns monthly debt summary
func (h *AggregationHandler) GetMonthlyDebtSummary(c *gin.Context) {
	// TODO: Implement actual database query
	summary := aggregations.MonthlyDebtSummary{
		Year:             2026,
		Month:            8,
		TotalDebt:        15000.00,
		NewDebt:          6000.00,
		PaymentsReceived: 4500.00,
		OverdueDebt:      3000.00,
		OverdueCount:     5,
		PaidDebt:         4500.00,
		UpdatedAt:        time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetDailyProfitSummary returns daily profit summary
func (h *AggregationHandler) GetDailyProfitSummary(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	// TODO: Implement actual database query
	summary := aggregations.DailyProfitSummary{
		Date:         date,
		GrossProfit:  1500.00,
		NetProfit:    1200.00,
		TotalRevenue: 5000.00,
		TotalCost:    3500.00,
		ProfitMargin: 24.00,
		UpdatedAt:    time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetMonthlyProfitSummary returns monthly profit summary
func (h *AggregationHandler) GetMonthlyProfitSummary(c *gin.Context) {
	// TODO: Implement actual database query
	summary := aggregations.MonthlyProfitSummary{
		Year:         2026,
		Month:        8,
		GrossProfit:  45000.00,
		NetProfit:    36000.00,
		TotalRevenue: 150000.00,
		TotalCost:    105000.00,
		ProfitMargin: 24.00,
		UpdatedAt:    time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetAggregationStatus returns the status of aggregation tables
func (h *AggregationHandler) GetAggregationStatus(c *gin.Context) {
	// TODO: Implement actual database query
	status := aggregations.AggregationStatus{
		LastDailySalesUpdate:       time.Now().Add(-2 * time.Hour),
		LastMonthlySalesUpdate:     time.Now().Add(-24 * time.Hour),
		LastDailyInventoryUpdate:   time.Now().Add(-2 * time.Hour),
		LastMonthlyInventoryUpdate: time.Now().Add(-24 * time.Hour),
		LastDailyDebtUpdate:        time.Now().Add(-2 * time.Hour),
		LastMonthlyDebtUpdate:      time.Now().Add(-24 * time.Hour),
		LastDailyProfitUpdate:      time.Now().Add(-2 * time.Hour),
		LastMonthlyProfitUpdate:    time.Now().Add(-24 * time.Hour),
		IsProcessing:               false,
		ProcessingSince:            time.Time{},
	}

	c.JSON(http.StatusOK, gin.H{"data": status})
}

// UpdateAggregations triggers aggregation updates
func (h *AggregationHandler) UpdateAggregations(c *gin.Context) {
	var req struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Force     bool   `json:"force"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual aggregation update logic
	// For now return success
	c.JSON(http.StatusOK, gin.H{
		"message":    "Aggregation update triggered",
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
		"force":      req.Force,
	})
}
