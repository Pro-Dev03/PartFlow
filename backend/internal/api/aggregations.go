package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/aggregations"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// AggregationHandler handles aggregation endpoints (ARCHITECTURE-PRINCIPLES.md)
type AggregationHandler struct {
	db *sqlx.DB
}

// NewAggregationHandler creates a new aggregation handler
func NewAggregationHandler(db *sqlx.DB) *AggregationHandler {
	return &AggregationHandler{db: db}
}

// sqliteSaleDateExpression keeps summary refresh compatible with older local
// databases that do not yet have the optional sale_date column. Newer local
// schemas prefer the explicit sale date, while legacy rows fall back to
// created_at.
func (h *AggregationHandler) sqliteSaleDateExpression(ctx context.Context) string {
	rows, err := h.db.QueryxContext(ctx, "PRAGMA table_info(sales)")
	if err != nil {
		return "date(s.created_at)"
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notNull, pk int
		var name, dataType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return "date(s.created_at)"
		}
		if strings.EqualFold(name, "sale_date") {
			return "date(COALESCE(s.sale_date, s.created_at))"
		}
	}
	return "date(s.created_at)"
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

func buildDailySalesSummaryQueryForDB(db *sqlx.DB) string {
	if dbutil.IsSQLite(db) {
		return buildDailySalesSummaryQuery()
	}
	return `
		SELECT
			$1::date AS date,
			COALESCE(s.total_sales, 0) AS total_sales,
			COALESCE(s.total_revenue, 0) AS total_revenue,
			COALESCE(s.total_profit, 0) AS total_profit,
			COALESCE(s.total_customers, 0) AS total_customers,
			COALESCE(s.average_order_value, 0) AS average_order_value,
			COALESCE(s.total_items_sold, 0) AS total_items_sold,
			COALESCE(s.cash_sales, 0) AS cash_sales,
			COALESCE(s.card_sales, 0) AS card_sales,
			COALESCE(s.debt_sales, 0) AS debt_sales,
			COALESCE(s.updated_at, NOW()) AS updated_at
		FROM (SELECT $1::date AS summary_date) dates
		LEFT JOIN daily_sales_summary s ON s.date = dates.summary_date
	`
}

func buildMonthlySalesSummaryQueryForDB(db *sqlx.DB) string {
	if dbutil.IsSQLite(db) {
		return buildMonthlySalesSummaryQuery()
	}
	return `
		SELECT
			$1::int AS year,
			$2::int AS month,
			COALESCE(s.total_sales, 0) AS total_sales,
			COALESCE(s.total_revenue, 0) AS total_revenue,
			COALESCE(s.total_profit, 0) AS total_profit,
			COALESCE(s.total_customers, 0) AS total_customers,
			COALESCE(s.average_order_value, 0) AS average_order_value,
			COALESCE(s.total_items_sold, 0) AS total_items_sold,
			COALESCE(s.cash_sales, 0) AS cash_sales,
			COALESCE(s.card_sales, 0) AS card_sales,
			COALESCE(s.debt_sales, 0) AS debt_sales,
			COALESCE(s.updated_at, NOW()) AS updated_at
		FROM (SELECT $1::int AS summary_year, $2::int AS summary_month) dates
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
	if err := h.refreshSQLiteSummaries(c.Request.Context(), date, date); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sales summary"})
		return
	}
	query := buildDailySalesSummaryQueryForDB(h.db)
	args := []interface{}{dateValue, dateValue}
	if !dbutil.IsSQLite(h.db) {
		args = []interface{}{dateValue}
	}
	err = h.db.GetContext(c.Request.Context(), &summary, query, args...)
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
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	if err := h.refreshSQLiteSummaries(c.Request.Context(), start, start.AddDate(0, 1, -1)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sales summary"})
		return
	}
	query := buildMonthlySalesSummaryQueryForDB(h.db)
	args := []interface{}{year, month, year, month}
	if !dbutil.IsSQLite(h.db) {
		args = []interface{}{year, month}
	}
	err := h.db.GetContext(c.Request.Context(), &summary, query, args...)
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

	if err := h.refreshSQLiteSummaries(c.Request.Context(), date, date); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory summary"})
		return
	}
	row, err := h.getDailyInventorySummary(c.Request.Context(), dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve daily inventory summary"})
		return
	}
	summary := row.toModel(date)

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetMonthlyInventorySummary returns monthly inventory summary
func (h *AggregationHandler) GetMonthlyInventorySummary(c *gin.Context) {
	year, month, err := parseAggregationMonth(c)
	if err != nil {
		return
	}
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	if err := h.refreshSQLiteSummaries(c.Request.Context(), start, end); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory summary"})
		return
	}
	row, err := h.getMonthlyInventorySummary(c.Request.Context(), year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve monthly inventory summary"})
		return
	}
	summary := row.toModel(year, month)

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

	if err := h.refreshSQLiteSummaries(c.Request.Context(), date, date); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update debt summary"})
		return
	}
	row, err := h.getDailyDebtSummary(c.Request.Context(), dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve daily debt summary"})
		return
	}
	summary := row.toModel(date)

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetMonthlyDebtSummary returns monthly debt summary
func (h *AggregationHandler) GetMonthlyDebtSummary(c *gin.Context) {
	year, month, err := parseAggregationMonth(c)
	if err != nil {
		return
	}
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	if err := h.refreshSQLiteSummaries(c.Request.Context(), start, end); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update debt summary"})
		return
	}
	row, err := h.getMonthlyDebtSummary(c.Request.Context(), year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve monthly debt summary"})
		return
	}
	summary := row.toModel(year, month)

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

	if err := h.refreshSQLiteSummaries(c.Request.Context(), date, date); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profit summary"})
		return
	}
	row, err := h.getDailyProfitSummary(c.Request.Context(), dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve daily profit summary"})
		return
	}
	summary := row.toModel(date)

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetMonthlyProfitSummary returns monthly profit summary
func (h *AggregationHandler) GetMonthlyProfitSummary(c *gin.Context) {
	year, month, err := parseAggregationMonth(c)
	if err != nil {
		return
	}
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	if err := h.refreshSQLiteSummaries(c.Request.Context(), start, end); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profit summary"})
		return
	}
	row, err := h.getMonthlyProfitSummary(c.Request.Context(), year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve monthly profit summary"})
		return
	}
	summary := row.toModel(year, month)

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetAggregationStatus returns the status of aggregation tables
func (h *AggregationHandler) GetAggregationStatus(c *gin.Context) {
	status, err := h.getAggregationStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve aggregation status"})
		return
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

	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()
	var err error
	if req.StartDate != "" {
		startDate, err = time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return
		}
	}
	if req.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}
	}
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be on or after start_date"})
		return
	}
	if err := h.refreshSQLiteSummaries(c.Request.Context(), startDate, endDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update aggregations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Aggregations updated",
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
		"force":      req.Force,
	})
}

// parseAggregationMonth parses optional year/month parameters consistently.
func parseAggregationMonth(c *gin.Context) (int, int, error) {
	now := time.Now()
	year, month := now.Year(), int(now.Month())
	if value := c.Query("year"); value != "" {
		if _, err := fmt.Sscanf(value, "%d", &year); err != nil || year < 2000 || year > 9999 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
			if err == nil {
				err = fmt.Errorf("year out of range")
			}
			return 0, 0, err
		}
	}
	if value := c.Query("month"); value != "" {
		if _, err := fmt.Sscanf(value, "%d", &month); err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month"})
			if err == nil {
				err = fmt.Errorf("month out of range")
			}
			return 0, 0, err
		}
	}
	return year, month, nil
}

type inventorySummaryRow struct {
	Date            string  `db:"date"`
	TotalItems      int     `db:"total_items"`
	TotalValue      float64 `db:"total_value"`
	LowStockCount   int     `db:"low_stock_count"`
	OutOfStockCount int     `db:"out_of_stock_count"`
	NewItemsAdded   int     `db:"new_items_added"`
	ItemsSold       int     `db:"items_sold"`
	ItemsReturned   int     `db:"items_returned"`
	ItemsDamaged    int     `db:"items_damaged"`
	UpdatedAt       string  `db:"updated_at"`
}

func (r inventorySummaryRow) toModel(date time.Time) aggregations.DailyInventorySummary {
	updatedAt, _ := dbutil.ParseTimestamp(r.UpdatedAt)
	return aggregations.DailyInventorySummary{Date: date, TotalItems: r.TotalItems, TotalValue: r.TotalValue, LowStockCount: r.LowStockCount, OutOfStockCount: r.OutOfStockCount, NewItemsAdded: r.NewItemsAdded, ItemsSold: r.ItemsSold, ItemsReturned: r.ItemsReturned, ItemsDamaged: r.ItemsDamaged, UpdatedAt: updatedAt}
}

type monthlyInventorySummaryRow struct {
	TotalItems      int     `db:"total_items"`
	TotalValue      float64 `db:"total_value"`
	LowStockCount   int     `db:"low_stock_count"`
	OutOfStockCount int     `db:"out_of_stock_count"`
	NewItemsAdded   int     `db:"new_items_added"`
	ItemsSold       int     `db:"items_sold"`
	ItemsReturned   int     `db:"items_returned"`
	ItemsDamaged    int     `db:"items_damaged"`
	UpdatedAt       string  `db:"updated_at"`
}

func (r monthlyInventorySummaryRow) toModel(year, month int) aggregations.MonthlyInventorySummary {
	updatedAt, _ := dbutil.ParseTimestamp(r.UpdatedAt)
	return aggregations.MonthlyInventorySummary{Year: year, Month: month, TotalItems: r.TotalItems, TotalValue: r.TotalValue, LowStockCount: r.LowStockCount, OutOfStockCount: r.OutOfStockCount, NewItemsAdded: r.NewItemsAdded, ItemsSold: r.ItemsSold, ItemsReturned: r.ItemsReturned, ItemsDamaged: r.ItemsDamaged, UpdatedAt: updatedAt}
}

type debtSummaryRow struct {
	Date             string  `db:"date"`
	TotalDebt        float64 `db:"total_debt"`
	NewDebt          float64 `db:"new_debt"`
	PaymentsReceived float64 `db:"payments_received"`
	OverdueDebt      float64 `db:"overdue_debt"`
	OverdueCount     int     `db:"overdue_count"`
	PaidDebt         float64 `db:"paid_debt"`
	UpdatedAt        string  `db:"updated_at"`
}

func (r debtSummaryRow) toModel(date time.Time) aggregations.DailyDebtSummary {
	updatedAt, _ := dbutil.ParseTimestamp(r.UpdatedAt)
	return aggregations.DailyDebtSummary{Date: date, TotalDebt: r.TotalDebt, NewDebt: r.NewDebt, PaymentsReceived: r.PaymentsReceived, OverdueDebt: r.OverdueDebt, OverdueCount: r.OverdueCount, PaidDebt: r.PaidDebt, UpdatedAt: updatedAt}
}

type monthlyDebtSummaryRow struct {
	TotalDebt        float64 `db:"total_debt"`
	NewDebt          float64 `db:"new_debt"`
	PaymentsReceived float64 `db:"payments_received"`
	OverdueDebt      float64 `db:"overdue_debt"`
	OverdueCount     int     `db:"overdue_count"`
	PaidDebt         float64 `db:"paid_debt"`
	UpdatedAt        string  `db:"updated_at"`
}

func (r monthlyDebtSummaryRow) toModel(year, month int) aggregations.MonthlyDebtSummary {
	updatedAt, _ := dbutil.ParseTimestamp(r.UpdatedAt)
	return aggregations.MonthlyDebtSummary{Year: year, Month: month, TotalDebt: r.TotalDebt, NewDebt: r.NewDebt, PaymentsReceived: r.PaymentsReceived, OverdueDebt: r.OverdueDebt, OverdueCount: r.OverdueCount, PaidDebt: r.PaidDebt, UpdatedAt: updatedAt}
}

type profitSummaryRow struct {
	Date         string  `db:"date"`
	GrossProfit  float64 `db:"gross_profit"`
	NetProfit    float64 `db:"net_profit"`
	TotalRevenue float64 `db:"total_revenue"`
	TotalCost    float64 `db:"total_cost"`
	ProfitMargin float64 `db:"profit_margin"`
	UpdatedAt    string  `db:"updated_at"`
}

func (r profitSummaryRow) toModel(date time.Time) aggregations.DailyProfitSummary {
	updatedAt, _ := dbutil.ParseTimestamp(r.UpdatedAt)
	return aggregations.DailyProfitSummary{Date: date, GrossProfit: r.GrossProfit, NetProfit: r.NetProfit, TotalRevenue: r.TotalRevenue, TotalCost: r.TotalCost, ProfitMargin: r.ProfitMargin, UpdatedAt: updatedAt}
}

type monthlyProfitSummaryRow struct {
	GrossProfit  float64 `db:"gross_profit"`
	NetProfit    float64 `db:"net_profit"`
	TotalRevenue float64 `db:"total_revenue"`
	TotalCost    float64 `db:"total_cost"`
	ProfitMargin float64 `db:"profit_margin"`
	UpdatedAt    string  `db:"updated_at"`
}

func (r monthlyProfitSummaryRow) toModel(year, month int) aggregations.MonthlyProfitSummary {
	updatedAt, _ := dbutil.ParseTimestamp(r.UpdatedAt)
	return aggregations.MonthlyProfitSummary{Year: year, Month: month, GrossProfit: r.GrossProfit, NetProfit: r.NetProfit, TotalRevenue: r.TotalRevenue, TotalCost: r.TotalCost, ProfitMargin: r.ProfitMargin, UpdatedAt: updatedAt}
}

func (h *AggregationHandler) getDailyInventorySummary(ctx context.Context, date string) (inventorySummaryRow, error) {
	var row inventorySummaryRow
	err := h.db.GetContext(ctx, &row, fmt.Sprintf(`SELECT date, total_items, total_value, low_stock_count, out_of_stock_count, new_items_added, items_sold, items_returned, items_damaged, updated_at FROM daily_inventory_summary WHERE date = %s`, dbutil.Placeholder(h.db, 1)), date)
	return row, err
}

func (h *AggregationHandler) getMonthlyInventorySummary(ctx context.Context, year, month int) (monthlyInventorySummaryRow, error) {
	var row monthlyInventorySummaryRow
	err := h.db.GetContext(ctx, &row, fmt.Sprintf(`SELECT total_items, total_value, low_stock_count, out_of_stock_count, new_items_added, items_sold, items_returned, items_damaged, updated_at FROM monthly_inventory_summary WHERE year = %s AND month = %s`, dbutil.Placeholder(h.db, 1), dbutil.Placeholder(h.db, 2)), year, month)
	return row, err
}

func (h *AggregationHandler) getDailyDebtSummary(ctx context.Context, date string) (debtSummaryRow, error) {
	var row debtSummaryRow
	err := h.db.GetContext(ctx, &row, fmt.Sprintf(`SELECT date, total_debt, new_debt, payments_received, overdue_debt, overdue_count, paid_debt, updated_at FROM daily_debt_summary WHERE date = %s`, dbutil.Placeholder(h.db, 1)), date)
	return row, err
}

func (h *AggregationHandler) getMonthlyDebtSummary(ctx context.Context, year, month int) (monthlyDebtSummaryRow, error) {
	var row monthlyDebtSummaryRow
	err := h.db.GetContext(ctx, &row, fmt.Sprintf(`SELECT total_debt, new_debt, payments_received, overdue_debt, overdue_count, paid_debt, updated_at FROM monthly_debt_summary WHERE year = %s AND month = %s`, dbutil.Placeholder(h.db, 1), dbutil.Placeholder(h.db, 2)), year, month)
	return row, err
}

func (h *AggregationHandler) getDailyProfitSummary(ctx context.Context, date string) (profitSummaryRow, error) {
	var row profitSummaryRow
	err := h.db.GetContext(ctx, &row, fmt.Sprintf(`SELECT date, gross_profit, net_profit, total_revenue, total_cost, profit_margin, updated_at FROM daily_profit_summary WHERE date = %s`, dbutil.Placeholder(h.db, 1)), date)
	return row, err
}

func (h *AggregationHandler) getMonthlyProfitSummary(ctx context.Context, year, month int) (monthlyProfitSummaryRow, error) {
	var row monthlyProfitSummaryRow
	err := h.db.GetContext(ctx, &row, fmt.Sprintf(`SELECT gross_profit, net_profit, total_revenue, total_cost, profit_margin, updated_at FROM monthly_profit_summary WHERE year = %s AND month = %s`, dbutil.Placeholder(h.db, 1), dbutil.Placeholder(h.db, 2)), year, month)
	return row, err
}

func (h *AggregationHandler) refreshSQLiteSummaries(ctx context.Context, startDate, endDate time.Time) error {
	if !dbutil.IsSQLite(h.db) {
		return h.refreshPostgresSummaries(ctx, startDate, endDate)
	}
	for day := startDate.Truncate(24 * time.Hour); !day.After(endDate); day = day.AddDate(0, 0, 1) {
		if err := h.refreshSQLiteDay(ctx, day); err != nil {
			return err
		}
	}
	for month := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC); !month.After(endDate); month = month.AddDate(0, 1, 0) {
		if err := h.refreshSQLiteMonth(ctx, month); err != nil {
			return err
		}
	}
	return nil
}

// refreshPostgresSummaries recomputes the same summaries used by the local
// SQLite store. The previous trigger-based implementation was removed because
// legacy and current schemas used different column names; explicit upserts are
// deterministic and keep dashboard values tied to the source tables.
func (h *AggregationHandler) refreshPostgresSummaries(ctx context.Context, startDate, endDate time.Time) error {
	for day := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC); !day.After(endDate); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		queries := []string{
			`INSERT INTO daily_sales_summary (date,total_sales,total_revenue,total_profit,total_customers,average_order_value,total_items_sold,cash_sales,card_sales,debt_sales,updated_at)
			 SELECT $1::date,COUNT(s.id),COALESCE(SUM(s.total_amount),0),COALESCE(SUM(s.total_amount-COALESCE(cost.total_cost,0)),0),COUNT(DISTINCT s.customer_id),COALESCE(SUM(s.total_amount)/NULLIF(COUNT(s.id),0),0),COALESCE(SUM(cost.total_items),0),COALESCE(SUM(CASE WHEN LOWER(COALESCE(s.payment_method,'')) IN ('cash','cash_payment') THEN s.total_amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN LOWER(COALESCE(s.payment_method,'')) IN ('card','credit_card') THEN s.total_amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN LOWER(COALESCE(s.payment_method,'')) IN ('debt','credit','on_account') THEN s.total_amount ELSE 0 END),0),NOW()
			 FROM sales s LEFT JOIN (SELECT si.sale_id,SUM(si.quantity) AS total_items,SUM(si.quantity*COALESCE(p.purchase_price,0)) AS total_cost FROM sale_items si LEFT JOIN products p ON p.id=si.product_id GROUP BY si.sale_id) cost ON cost.sale_id=s.id
				 WHERE s.sale_date::date=$1::date AND LOWER(COALESCE(s.status,'completed'))='completed'
			 ON CONFLICT (date) DO UPDATE SET total_sales=EXCLUDED.total_sales,total_revenue=EXCLUDED.total_revenue,total_profit=EXCLUDED.total_profit,total_customers=EXCLUDED.total_customers,average_order_value=EXCLUDED.average_order_value,total_items_sold=EXCLUDED.total_items_sold,cash_sales=EXCLUDED.cash_sales,card_sales=EXCLUDED.card_sales,debt_sales=EXCLUDED.debt_sales,updated_at=NOW()`,
			`INSERT INTO daily_inventory_summary (date,total_items,total_value,low_stock_count,out_of_stock_count,new_items_added,items_sold,items_returned,items_damaged,updated_at)
				 SELECT $1::date,(SELECT COUNT(*) FROM inventory_items WHERE UPPER(COALESCE(status,'')) NOT IN ('SOLD','ARCHIVED')),(SELECT COALESCE(SUM(purchase_cost),0) FROM inventory_items WHERE UPPER(COALESCE(status,'')) NOT IN ('SOLD','ARCHIVED')),(SELECT COUNT(*) FROM products p WHERE p.is_active AND p.min_stock_level>0 AND (SELECT COUNT(*) FROM inventory_items i WHERE i.product_id=p.id AND UPPER(i.status)='AVAILABLE') BETWEEN 1 AND p.min_stock_level),(SELECT COUNT(*) FROM products p WHERE p.is_active AND (SELECT COUNT(*) FROM inventory_items i WHERE i.product_id=p.id AND UPPER(i.status)='AVAILABLE')=0),(SELECT COUNT(*) FROM inventory_items WHERE created_at::date=$1::date),(SELECT COALESCE(SUM(si.quantity),0) FROM sale_items si JOIN sales s ON s.id=si.sale_id WHERE s.sale_date::date=$1::date AND LOWER(COALESCE(s.status,'completed'))='completed'),(SELECT COALESCE(SUM(ri.quantity_returned),0) FROM return_items ri JOIN returns r ON r.id=ri.return_id WHERE COALESCE(r.return_date,r.created_at)::date=$1::date AND LOWER(COALESCE(r.status,'completed'))='completed'),(SELECT COUNT(*) FROM inventory_movements WHERE created_at::date=$1::date AND UPPER(movement_type)='DAMAGE'),NOW()
			 ON CONFLICT (date) DO UPDATE SET total_items=EXCLUDED.total_items,total_value=EXCLUDED.total_value,low_stock_count=EXCLUDED.low_stock_count,out_of_stock_count=EXCLUDED.out_of_stock_count,new_items_added=EXCLUDED.new_items_added,items_sold=EXCLUDED.items_sold,items_returned=EXCLUDED.items_returned,items_damaged=EXCLUDED.items_damaged,updated_at=NOW()`,
			`INSERT INTO daily_debt_summary (date,total_debt,new_debt,payments_received,overdue_debt,overdue_count,paid_debt,updated_at)
			 SELECT $1::date,COALESCE(SUM(CASE WHEN created_at::date<=$1::date AND remaining_amount>0 THEN remaining_amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN created_at::date=$1::date THEN amount ELSE 0 END),0),(SELECT COALESCE(SUM(amount),0) FROM payments WHERE customer_id IS NOT NULL AND payment_date::date=$1::date),COALESCE(SUM(CASE WHEN due_date<$1::date AND remaining_amount>0 THEN remaining_amount ELSE 0 END),0),COUNT(*) FILTER (WHERE due_date<$1::date AND remaining_amount>0),COALESCE(SUM(CASE WHEN remaining_amount<=0 AND updated_at::date=$1::date THEN amount ELSE 0 END),0),NOW() FROM debts
			 ON CONFLICT (date) DO UPDATE SET total_debt=EXCLUDED.total_debt,new_debt=EXCLUDED.new_debt,payments_received=EXCLUDED.payments_received,overdue_debt=EXCLUDED.overdue_debt,overdue_count=EXCLUDED.overdue_count,paid_debt=EXCLUDED.paid_debt,updated_at=NOW()`,
			`INSERT INTO daily_profit_summary (date,gross_profit,net_profit,total_revenue,total_cost,profit_margin,updated_at)
				 SELECT $1::date,COALESCE(SUM(s.total_amount-COALESCE(cost.total_cost,0)),0),COALESCE(SUM(s.total_amount-COALESCE(cost.total_cost,0)),0)-COALESCE((SELECT SUM(amount) FROM expenses WHERE expense_date::date=$1::date AND LOWER(COALESCE(status,'approved')) IN ('approved','paid','completed')),0),COALESCE(SUM(s.total_amount),0),COALESCE(SUM(cost.total_cost),0),CASE WHEN COALESCE(SUM(s.total_amount),0)=0 THEN 0 ELSE ((COALESCE(SUM(s.total_amount-COALESCE(cost.total_cost,0)),0)-COALESCE((SELECT SUM(amount) FROM expenses WHERE expense_date::date=$1::date AND LOWER(COALESCE(status,'approved')) IN ('approved','paid','completed')),0))/SUM(s.total_amount))*100 END,NOW() FROM sales s LEFT JOIN (SELECT si.sale_id,SUM(si.quantity*COALESCE(p.purchase_price,0)) AS total_cost FROM sale_items si LEFT JOIN products p ON p.id=si.product_id GROUP BY si.sale_id) cost ON cost.sale_id=s.id WHERE s.sale_date::date=$1::date AND LOWER(COALESCE(s.status,'completed'))='completed'
			 ON CONFLICT (date) DO UPDATE SET gross_profit=EXCLUDED.gross_profit,net_profit=EXCLUDED.net_profit,total_revenue=EXCLUDED.total_revenue,total_cost=EXCLUDED.total_cost,profit_margin=EXCLUDED.profit_margin,updated_at=NOW()`,
		}
		for _, query := range queries {
			if _, err := h.db.ExecContext(ctx, query, date); err != nil {
				return fmt.Errorf("refresh PostgreSQL summaries for %s: %w", date, err)
			}
		}
	}
	for month := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC); !month.After(endDate); month = month.AddDate(0, 1, 0) {
		year, monthNumber := month.Year(), int(month.Month())
		next := month.AddDate(0, 1, 0).Format("2006-01-02")
		start := month.Format("2006-01-02")
		queries := []string{
			`INSERT INTO monthly_sales_summary (year,month,total_sales,total_revenue,total_profit,total_customers,average_order_value,total_items_sold,cash_sales,card_sales,debt_sales,updated_at) SELECT $1,$2,COUNT(s.id),COALESCE(SUM(s.total_amount),0),COALESCE(SUM(s.total_amount-COALESCE(cost.total_cost,0)),0),COUNT(DISTINCT s.customer_id),COALESCE(SUM(s.total_amount)/NULLIF(COUNT(s.id),0),0),COALESCE(SUM(cost.total_items),0),COALESCE(SUM(CASE WHEN LOWER(COALESCE(s.payment_method,'')) IN ('cash','cash_payment') THEN s.total_amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN LOWER(COALESCE(s.payment_method,'')) IN ('card','credit_card') THEN s.total_amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN LOWER(COALESCE(s.payment_method,'')) IN ('debt','credit','on_account') THEN s.total_amount ELSE 0 END),0),NOW() FROM sales s LEFT JOIN (SELECT si.sale_id,SUM(si.quantity) total_items,SUM(si.quantity*COALESCE(p.purchase_price,0)) total_cost FROM sale_items si LEFT JOIN products p ON p.id=si.product_id GROUP BY si.sale_id) cost ON cost.sale_id=s.id WHERE s.sale_date >= $3::date AND s.sale_date < $4::date AND LOWER(COALESCE(s.status,'completed'))='completed' ON CONFLICT(year,month) DO UPDATE SET total_sales=EXCLUDED.total_sales,total_revenue=EXCLUDED.total_revenue,total_profit=EXCLUDED.total_profit,total_customers=EXCLUDED.total_customers,average_order_value=EXCLUDED.average_order_value,total_items_sold=EXCLUDED.total_items_sold,cash_sales=EXCLUDED.cash_sales,card_sales=EXCLUDED.card_sales,debt_sales=EXCLUDED.debt_sales,updated_at=NOW()`,
			`INSERT INTO monthly_inventory_summary (year,month,total_items,total_value,low_stock_count,out_of_stock_count,new_items_added,items_sold,items_returned,items_damaged,updated_at) SELECT $1,$2,(SELECT COUNT(*) FROM inventory_items WHERE UPPER(COALESCE(status,'')) NOT IN ('SOLD','ARCHIVED')),(SELECT COALESCE(SUM(purchase_cost),0) FROM inventory_items WHERE UPPER(COALESCE(status,'')) NOT IN ('SOLD','ARCHIVED')),(SELECT COUNT(*) FROM products p WHERE p.is_active AND p.min_stock_level>0 AND (SELECT COUNT(*) FROM inventory_items i WHERE i.product_id=p.id AND UPPER(i.status)='AVAILABLE') BETWEEN 1 AND p.min_stock_level),(SELECT COUNT(*) FROM products p WHERE p.is_active AND (SELECT COUNT(*) FROM inventory_items i WHERE i.product_id=p.id AND UPPER(i.status)='AVAILABLE')=0),(SELECT COUNT(*) FROM inventory_items WHERE created_at >= $3::date AND created_at < $4::date),(SELECT COALESCE(SUM(si.quantity),0) FROM sale_items si JOIN sales s ON s.id=si.sale_id WHERE s.sale_date >= $3::date AND s.sale_date < $4::date AND LOWER(COALESCE(s.status,'completed'))='completed'),(SELECT COALESCE(SUM(ri.quantity_returned),0) FROM return_items ri JOIN returns r ON r.id=ri.return_id WHERE COALESCE(r.return_date,r.created_at) >= $3::date AND COALESCE(r.return_date,r.created_at) < $4::date AND LOWER(COALESCE(r.status,'completed'))='completed'),(SELECT COUNT(*) FROM inventory_movements WHERE created_at >= $3::date AND created_at < $4::date AND UPPER(movement_type)='DAMAGE'),NOW() ON CONFLICT(year,month) DO UPDATE SET total_items=EXCLUDED.total_items,total_value=EXCLUDED.total_value,low_stock_count=EXCLUDED.low_stock_count,out_of_stock_count=EXCLUDED.out_of_stock_count,new_items_added=EXCLUDED.new_items_added,items_sold=EXCLUDED.items_sold,items_returned=EXCLUDED.items_returned,items_damaged=EXCLUDED.items_damaged,updated_at=NOW()`,
			`INSERT INTO monthly_debt_summary (year,month,total_debt,new_debt,payments_received,overdue_debt,overdue_count,paid_debt,updated_at) SELECT $1,$2,COALESCE(SUM(CASE WHEN created_at<$4::date AND remaining_amount>0 THEN remaining_amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN created_at >= $3::date AND created_at < $4::date THEN amount ELSE 0 END),0),(SELECT COALESCE(SUM(amount),0) FROM payments WHERE customer_id IS NOT NULL AND payment_date >= $3::date AND payment_date < $4::date),COALESCE(SUM(CASE WHEN due_date<$4::date AND remaining_amount>0 THEN remaining_amount ELSE 0 END),0),COUNT(*) FILTER (WHERE due_date<$4::date AND remaining_amount>0),COALESCE(SUM(CASE WHEN remaining_amount<=0 AND updated_at >= $3::date AND updated_at < $4::date THEN amount ELSE 0 END),0),NOW() FROM debts ON CONFLICT(year,month) DO UPDATE SET total_debt=EXCLUDED.total_debt,new_debt=EXCLUDED.new_debt,payments_received=EXCLUDED.payments_received,overdue_debt=EXCLUDED.overdue_debt,overdue_count=EXCLUDED.overdue_count,paid_debt=EXCLUDED.paid_debt,updated_at=NOW()`,
			`INSERT INTO monthly_profit_summary (year,month,gross_profit,net_profit,total_revenue,total_cost,profit_margin,updated_at) SELECT $1,$2,COALESCE(SUM(s.total_amount-COALESCE(cost.total_cost,0)),0),COALESCE(SUM(s.total_amount-COALESCE(cost.total_cost,0)),0)-COALESCE((SELECT SUM(amount) FROM expenses WHERE expense_date >= $3::date AND expense_date < $4::date AND LOWER(COALESCE(status,'approved')) IN ('approved','paid','completed')),0),COALESCE(SUM(s.total_amount),0),COALESCE(SUM(cost.total_cost),0),CASE WHEN COALESCE(SUM(s.total_amount),0)=0 THEN 0 ELSE ((COALESCE(SUM(s.total_amount-COALESCE(cost.total_cost,0)),0)-COALESCE((SELECT SUM(amount) FROM expenses WHERE expense_date >= $3::date AND expense_date < $4::date AND LOWER(COALESCE(status,'approved')) IN ('approved','paid','completed')),0))/SUM(s.total_amount))*100 END,NOW() FROM sales s LEFT JOIN (SELECT si.sale_id,SUM(si.quantity*COALESCE(p.purchase_price,0)) total_cost FROM sale_items si LEFT JOIN products p ON p.id=si.product_id GROUP BY si.sale_id) cost ON cost.sale_id=s.id WHERE s.sale_date >= $3::date AND s.sale_date < $4::date AND LOWER(COALESCE(s.status,'completed'))='completed' ON CONFLICT(year,month) DO UPDATE SET gross_profit=EXCLUDED.gross_profit,net_profit=EXCLUDED.net_profit,total_revenue=EXCLUDED.total_revenue,total_cost=EXCLUDED.total_cost,profit_margin=EXCLUDED.profit_margin,updated_at=NOW()`,
		}
		for _, query := range queries {
			if _, err := h.db.ExecContext(ctx, query, year, monthNumber, start, next); err != nil {
				return fmt.Errorf("refresh PostgreSQL monthly summaries for %04d-%02d: %w", year, monthNumber, err)
			}
		}
	}
	return nil
}

func (h *AggregationHandler) refreshSQLiteDay(ctx context.Context, day time.Time) error {
	date := day.Format("2006-01-02")
	queries := []struct {
		query string
		args  []any
	}{
		{`INSERT OR REPLACE INTO daily_sales_summary (date, total_sales, total_revenue, total_profit, total_customers, average_order_value, total_items_sold, cash_sales, card_sales, debt_sales, updated_at)
			SELECT ?, COUNT(s.id), COALESCE(SUM(s.total_amount), 0), COALESCE(SUM(s.total_amount - COALESCE(cost.total_cost, 0)), 0),
			COUNT(DISTINCT s.customer_id), COALESCE(SUM(s.total_amount) / NULLIF(COUNT(s.id), 0), 0),
			COALESCE(SUM(cost.total_items), 0),
			COALESCE(SUM(CASE WHEN lower(COALESCE(s.payment_method, '')) IN ('cash', 'cash_payment') THEN s.total_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN lower(COALESCE(s.payment_method, '')) IN ('card', 'credit_card') THEN s.total_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN lower(COALESCE(s.payment_method, '')) IN ('debt', 'credit', 'on_account') THEN s.total_amount ELSE 0 END), 0), CURRENT_TIMESTAMP
			FROM sales s LEFT JOIN (SELECT si.sale_id, SUM(si.quantity) AS total_items, SUM(si.quantity * COALESCE(ii.purchase_cost, p.purchase_price, 0)) AS total_cost FROM sale_items si LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id LEFT JOIN products p ON p.id = si.product_id GROUP BY si.sale_id) cost ON cost.sale_id = s.id
			WHERE date(COALESCE(s.sale_date, s.created_at)) = ? AND lower(COALESCE(s.status, 'completed')) = 'completed'
			`, []any{date, date}},
		{`INSERT OR REPLACE INTO daily_inventory_summary (date, total_items, total_value, low_stock_count, out_of_stock_count, new_items_added, items_sold, items_returned, items_damaged, updated_at)
			SELECT ?, (SELECT COUNT(*) FROM inventory_items WHERE status NOT IN ('SOLD', 'ARCHIVED')), (SELECT COALESCE(SUM(purchase_cost), 0) FROM inventory_items WHERE status NOT IN ('SOLD', 'ARCHIVED')),
			(SELECT COUNT(*) FROM products p WHERE p.is_active = 1 AND p.min_stock_level > 0 AND (SELECT COUNT(*) FROM inventory_items i WHERE i.product_id = p.id AND i.status = 'AVAILABLE') BETWEEN 1 AND p.min_stock_level),
			(SELECT COUNT(*) FROM products p WHERE p.is_active = 1 AND (SELECT COUNT(*) FROM inventory_items i WHERE i.product_id = p.id AND i.status = 'AVAILABLE') = 0),
			(SELECT COUNT(*) FROM inventory_items WHERE date(created_at) = ?),
			(SELECT COALESCE(SUM(si.quantity), 0) FROM sale_items si JOIN sales s ON s.id = si.sale_id WHERE date(COALESCE(s.sale_date, s.created_at)) = ? AND lower(COALESCE(s.status, 'completed')) = 'completed'),
			(SELECT COALESCE(SUM(ri.quantity), 0) FROM return_items ri JOIN returns r ON r.id = ri.return_id WHERE date(COALESCE(r.return_date, r.created_at)) = ? AND lower(COALESCE(r.status, 'completed')) = 'completed'),
			(SELECT COUNT(*) FROM inventory_movements WHERE date(created_at) = ? AND upper(movement_type) = 'DAMAGE'), CURRENT_TIMESTAMP
			`, []any{date, date, date, date, date, date}},
		{`INSERT OR REPLACE INTO daily_debt_summary (date, total_debt, new_debt, payments_received, overdue_debt, overdue_count, paid_debt, updated_at)
			SELECT ?, COALESCE(SUM(CASE WHEN date(created_at) <= ? AND remaining_amount > 0 THEN remaining_amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN date(created_at) = ? THEN amount ELSE 0 END), 0),
			COALESCE((SELECT SUM(amount) FROM payments WHERE customer_id IS NOT NULL AND date(created_at) = ?), 0), COALESCE(SUM(CASE WHEN due_date < ? AND remaining_amount > 0 THEN remaining_amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN due_date < ? AND remaining_amount > 0 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN remaining_amount <= 0 AND date(updated_at) = ? THEN amount ELSE 0 END), 0), CURRENT_TIMESTAMP FROM debts`, []any{date, date, date, date, date, date, date}},
		{`INSERT OR REPLACE INTO daily_profit_summary (date, gross_profit, net_profit, total_revenue, total_cost, profit_margin, updated_at)
			WITH sale_costs AS (SELECT s.id, s.total_amount, COALESCE(SUM(si.quantity * COALESCE(ii.purchase_cost, p.purchase_price, 0)), 0) AS total_cost FROM sales s LEFT JOIN sale_items si ON si.sale_id = s.id LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id LEFT JOIN products p ON p.id = si.product_id WHERE date(COALESCE(s.sale_date, s.created_at)) = ? AND lower(COALESCE(s.status, 'completed')) = 'completed' GROUP BY s.id), totals AS (SELECT COALESCE(SUM(total_amount), 0) AS revenue, COALESCE(SUM(total_cost), 0) AS cost FROM sale_costs), expenses_total AS (SELECT COALESCE(SUM(amount), 0) AS amount FROM expenses WHERE date(expense_date) = ? AND lower(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed'))
			SELECT ?, totals.revenue - totals.cost, totals.revenue - totals.cost - expenses_total.amount, totals.revenue, totals.cost, CASE WHEN totals.revenue = 0 THEN 0 ELSE ((totals.revenue - totals.cost - expenses_total.amount) / totals.revenue) * 100 END, CURRENT_TIMESTAMP FROM totals, expenses_total
			`, []any{date, date, date}},
	}
	saleDateExpr := h.sqliteSaleDateExpression(ctx)
	for _, item := range queries {
		query := strings.ReplaceAll(item.query, "date(COALESCE(s.sale_date, s.created_at))", saleDateExpr)
		if _, err := h.db.ExecContext(ctx, query, item.args...); err != nil {
			return fmt.Errorf("refresh daily summaries for %s: %w", date, err)
		}
	}
	return nil
}

func (h *AggregationHandler) refreshSQLiteMonth(ctx context.Context, month time.Time) error {
	year, monthNumber := month.Year(), int(month.Month())
	start := month.Format("2006-01-02")
	end := month.AddDate(0, 1, 0).Format("2006-01-02")
	queries := []struct {
		query string
		args  []any
	}{
		{`INSERT OR REPLACE INTO monthly_sales_summary (year, month, total_sales, total_revenue, total_profit, total_customers, average_order_value, total_items_sold, cash_sales, card_sales, debt_sales, updated_at)
			SELECT ?, ?, COUNT(s.id), COALESCE(SUM(s.total_amount), 0), COALESCE(SUM(s.total_amount - COALESCE(cost.total_cost, 0)), 0), COUNT(DISTINCT s.customer_id), COALESCE(SUM(s.total_amount) / NULLIF(COUNT(s.id), 0), 0), COALESCE(SUM(cost.total_items), 0),
			COALESCE(SUM(CASE WHEN lower(COALESCE(s.payment_method, '')) IN ('cash', 'cash_payment') THEN s.total_amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN lower(COALESCE(s.payment_method, '')) IN ('card', 'credit_card') THEN s.total_amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN lower(COALESCE(s.payment_method, '')) IN ('debt', 'credit', 'on_account') THEN s.total_amount ELSE 0 END), 0), CURRENT_TIMESTAMP
			FROM sales s LEFT JOIN (SELECT si.sale_id, SUM(si.quantity) AS total_items, SUM(si.quantity * COALESCE(ii.purchase_cost, p.purchase_price, 0)) AS total_cost FROM sale_items si LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id LEFT JOIN products p ON p.id = si.product_id GROUP BY si.sale_id) cost ON cost.sale_id = s.id
			WHERE date(COALESCE(s.sale_date, s.created_at)) >= ? AND date(COALESCE(s.sale_date, s.created_at)) < ? AND lower(COALESCE(s.status, 'completed')) = 'completed'
			`, []any{year, monthNumber, start, end}},
		{`INSERT OR REPLACE INTO monthly_inventory_summary (year, month, total_items, total_value, low_stock_count, out_of_stock_count, new_items_added, items_sold, items_returned, items_damaged, updated_at)
			SELECT ?, ?, (SELECT COUNT(*) FROM inventory_items WHERE status NOT IN ('SOLD', 'ARCHIVED')), (SELECT COALESCE(SUM(purchase_cost), 0) FROM inventory_items WHERE status NOT IN ('SOLD', 'ARCHIVED')), (SELECT COUNT(*) FROM products p WHERE p.is_active = 1 AND p.min_stock_level > 0 AND (SELECT COUNT(*) FROM inventory_items i WHERE i.product_id = p.id AND i.status = 'AVAILABLE') BETWEEN 1 AND p.min_stock_level), (SELECT COUNT(*) FROM products p WHERE p.is_active = 1 AND (SELECT COUNT(*) FROM inventory_items i WHERE i.product_id = p.id AND i.status = 'AVAILABLE') = 0), (SELECT COUNT(*) FROM inventory_items WHERE date(created_at) >= ? AND date(created_at) < ?), (SELECT COALESCE(SUM(si.quantity), 0) FROM sale_items si JOIN sales s ON s.id = si.sale_id WHERE date(COALESCE(s.sale_date, s.created_at)) >= ? AND date(COALESCE(s.sale_date, s.created_at)) < ? AND lower(COALESCE(s.status, 'completed')) = 'completed'), (SELECT COALESCE(SUM(ri.quantity), 0) FROM return_items ri JOIN returns r ON r.id = ri.return_id WHERE date(COALESCE(r.return_date, r.created_at)) >= ? AND date(COALESCE(r.return_date, r.created_at)) < ? AND lower(COALESCE(r.status, 'completed')) = 'completed'), (SELECT COUNT(*) FROM inventory_movements WHERE date(created_at) >= ? AND date(created_at) < ? AND upper(movement_type) = 'DAMAGE'), CURRENT_TIMESTAMP
			`, []any{year, monthNumber, start, end, start, end, start, end, start, end}},
		{`INSERT OR REPLACE INTO monthly_debt_summary (year, month, total_debt, new_debt, payments_received, overdue_debt, overdue_count, paid_debt, updated_at)
			SELECT ?, ?, COALESCE(SUM(CASE WHEN date(created_at) < ? AND remaining_amount > 0 THEN remaining_amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN date(created_at) >= ? AND date(created_at) < ? THEN amount ELSE 0 END), 0), COALESCE((SELECT SUM(amount) FROM payments WHERE customer_id IS NOT NULL AND date(created_at) >= ? AND date(created_at) < ?), 0), COALESCE(SUM(CASE WHEN due_date < ? AND remaining_amount > 0 THEN remaining_amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN due_date < ? AND remaining_amount > 0 THEN 1 ELSE 0 END), 0), COALESCE(SUM(CASE WHEN remaining_amount <= 0 AND date(updated_at) >= ? AND date(updated_at) < ? THEN amount ELSE 0 END), 0), CURRENT_TIMESTAMP FROM debts
			`, []any{year, monthNumber, end, start, end, start, end, end, end, start, end}},
		{`INSERT OR REPLACE INTO monthly_profit_summary (year, month, gross_profit, net_profit, total_revenue, total_cost, profit_margin, updated_at)
			WITH sale_costs AS (SELECT s.id, s.total_amount, COALESCE(SUM(si.quantity * COALESCE(ii.purchase_cost, p.purchase_price, 0)), 0) AS total_cost FROM sales s LEFT JOIN sale_items si ON si.sale_id = s.id LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id LEFT JOIN products p ON p.id = si.product_id WHERE date(COALESCE(s.sale_date, s.created_at)) >= ? AND date(COALESCE(s.sale_date, s.created_at)) < ? AND lower(COALESCE(s.status, 'completed')) = 'completed' GROUP BY s.id), totals AS (SELECT COALESCE(SUM(total_amount), 0) AS revenue, COALESCE(SUM(total_cost), 0) AS cost FROM sale_costs), expenses_total AS (SELECT COALESCE(SUM(amount), 0) AS amount FROM expenses WHERE date(expense_date) >= ? AND date(expense_date) < ? AND lower(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed'))
			SELECT ?, ?, totals.revenue - totals.cost, totals.revenue - totals.cost - expenses_total.amount, totals.revenue, totals.cost, CASE WHEN totals.revenue = 0 THEN 0 ELSE ((totals.revenue - totals.cost - expenses_total.amount) / totals.revenue) * 100 END, CURRENT_TIMESTAMP FROM totals, expenses_total
			`, []any{start, end, start, end, year, monthNumber}},
	}
	saleDateExpr := h.sqliteSaleDateExpression(ctx)
	for _, item := range queries {
		query := strings.ReplaceAll(item.query, "date(COALESCE(s.sale_date, s.created_at))", saleDateExpr)
		if _, err := h.db.ExecContext(ctx, query, item.args...); err != nil {
			return fmt.Errorf("refresh monthly summaries for %04d-%02d: %w", year, monthNumber, err)
		}
	}
	return nil
}

func (h *AggregationHandler) getAggregationStatus(ctx context.Context) (aggregations.AggregationStatus, error) {
	status := aggregations.AggregationStatus{}
	fields := []struct {
		table string
		dest  *time.Time
	}{
		{"daily_sales_summary", &status.LastDailySalesUpdate}, {"monthly_sales_summary", &status.LastMonthlySalesUpdate},
		{"daily_inventory_summary", &status.LastDailyInventoryUpdate}, {"monthly_inventory_summary", &status.LastMonthlyInventoryUpdate},
		{"daily_debt_summary", &status.LastDailyDebtUpdate}, {"monthly_debt_summary", &status.LastMonthlyDebtUpdate},
		{"daily_profit_summary", &status.LastDailyProfitUpdate}, {"monthly_profit_summary", &status.LastMonthlyProfitUpdate},
	}
	for _, field := range fields {
		var value sql.NullString
		if err := h.db.GetContext(ctx, &value, fmt.Sprintf("SELECT MAX(updated_at) FROM %s", field.table)); err != nil {
			return status, err
		}
		if value.Valid {
			parsed, err := dbutil.ParseTimestamp(value.String)
			if err != nil {
				return status, err
			}
			*field.dest = parsed
		}
	}
	return status, nil
}
