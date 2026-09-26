package reports

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/business"
	dbutil "github.com/partflow/smart-store/internal/database"
)

// Repository handles report data operations
type Repository struct {
	db *sqlx.DB
}

func postgresStoreTimestampDateExpression(timestampExpression string) string {
	return accounting.PostgresStoreDateExpression(timestampExpression)
}

func (r *Repository) returnDateExpression(alias string) string {
	columns := []string{"refund_date", "return_date", "created_at"}
	if dbutil.IsSQLite(r.db) {
		available := make([]string, 0, len(columns))
		for _, column := range columns {
			if reportsSQLiteHasColumns(r.db, "accounting_returns", column) {
				available = append(available, alias+"."+column)
			}
		}
		if len(available) == 0 {
			return fmt.Sprintf("store_date(%s.return_date)", alias)
		}
		if len(available) == 1 {
			return "store_date(" + available[0] + ")"
		}
		return "store_date(COALESCE(" + strings.Join(available, ", ") + "))"
	}
	return fmt.Sprintf("date(COALESCE(%s.refund_date::date, %s.return_date::date, %s))", alias, alias, postgresStoreTimestampDateExpression(alias+".created_at"))
}

func (r *Repository) salesDateExpression(alias string) string {
	if dbutil.IsSQLite(r.db) {
		if !reportsSQLiteHasColumns(r.db, "sales", "created_at") {
			if alias == "" {
				return "store_date(sale_date)"
			}
			return fmt.Sprintf("store_date(%s.sale_date)", alias)
		}
		if alias == "" {
			return "store_date(COALESCE(sale_date, created_at))"
		}
		return fmt.Sprintf("store_date(COALESCE(%s.sale_date, %s.created_at))", alias, alias)
	}
	if alias == "" {
		return fmt.Sprintf("date(COALESCE(sale_date::date, %s))", postgresStoreTimestampDateExpression("created_at"))
	}
	return fmt.Sprintf("date(COALESCE(%s.sale_date::date, %s))", alias, postgresStoreTimestampDateExpression(alias+".created_at"))
}

func reportsSQLiteHasColumns(db *sqlx.DB, table string, required ...string) bool {
	if !dbutil.IsSQLite(db) {
		return true
	}
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false
	}
	defer rows.Close()
	found := map[string]bool{}
	for rows.Next() {
		var cid, notNull, pk int
		var name, dataType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false
		}
		found[strings.ToLower(name)] = true
	}
	if err := rows.Err(); err != nil {
		return false
	}
	for _, column := range required {
		if !found[strings.ToLower(column)] {
			return false
		}
	}
	return true
}

func reportsSQLiteHasRelation(ctx context.Context, db *sqlx.DB, name string) (bool, error) {
	var count int
	if err := db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM (
			SELECT name FROM sqlite_master WHERE type IN ('table', 'view') AND name = ?
			UNION ALL
			SELECT name FROM sqlite_temp_master WHERE type IN ('table', 'view') AND name = ?
		)`, name, name); err != nil {
		return false, err
	}
	return count > 0, nil
}

func profitTrendWithoutStoredSaleCost(query, lineCostExpr string) string {
	start := strings.Index(query, "CASE WHEN NULLIF(s.cost_amount, 0) IS NOT NULL THEN s.cost_amount")
	if start < 0 {
		return query
	}
	endOffset := strings.Index(query[start:], "END AS sale_cost")
	if endOffset < 0 {
		return query
	}
	end := start + endOffset + len("END AS sale_cost")
	return query[:start] + fmt.Sprintf("COALESCE(SUM(si.quantity * %s), 0) AS sale_cost", lineCostExpr) + query[end:]
}

// reportTimestamp accepts both PostgreSQL timestamps and the TEXT timestamps
// used by the local SQLite store.  SQLite returns TEXT values for date
// expressions (DATE/strftime), so scanning directly into time.Time silently
// dropped rows from several reports.
type reportTimestamp struct{ time.Time }

func (t *reportTimestamp) Scan(value any) error {
	parsed, err := dbutil.ParseTimestamp(value)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// reportDate scans SQL DATE/grouping keys as store calendar dates rather than
// treating midnight UTC as an instant that may belong to the previous store day.
type reportDate struct{ time.Time }

func (d *reportDate) Scan(value any) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	var dateKey string
	switch v := value.(type) {
	case time.Time:
		dateKey = v.Format("2006-01-02")
	case string:
		dateKey = strings.TrimSpace(v)
	case []byte:
		dateKey = strings.TrimSpace(string(v))
	default:
		parsed, err := dbutil.ParseTimestamp(value)
		if err != nil {
			return err
		}
		dateKey = parsed.Format("2006-01-02")
	}
	if len(dateKey) > len("2006-01-02") {
		dateKey = dateKey[:len("2006-01-02")]
	}
	if _, err := time.Parse("2006-01-02", dateKey); err != nil {
		return fmt.Errorf("parse store calendar date %q: %w", dateKey, err)
	}
	start, _, err := accounting.StoreDateBounds(dateKey)
	if err != nil {
		return err
	}
	location, err := accounting.StoreLocation()
	if err != nil {
		return err
	}
	d.Time = start.In(location)
	return nil
}

// NewRepository creates a new report repository
func NewRepository(db *sqlx.DB) *Repository {
	// Keep lightweight SQLite report fixtures and pre-ledger test databases
	// compatible. Initialized PartFlow databases create permanent union views
	// through the schema migration, so this fallback is only a direct alias.
	if db != nil && dbutil.IsSQLite(db) {
		ensureSQLiteReportAliasView(db, "accounting_returns", "returns")
		ensureSQLiteReportAliasView(db, "accounting_return_items", "return_items")
	}
	return &Repository{db: db}
}

func ensureSQLiteReportAliasView(db *sqlx.DB, viewName, sourceName string) {
	var viewCount int
	if err := db.Get(&viewCount, `SELECT COUNT(*) FROM sqlite_master WHERE type='view' AND name=?`, viewName); err != nil || viewCount > 0 {
		return
	}
	var tableCount int
	if err := db.Get(&tableCount, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, sourceName); err != nil || tableCount == 0 {
		return
	}
	_, _ = db.Exec(`CREATE TEMP VIEW ` + viewName + ` AS SELECT * FROM ` + sourceName)
}

// historicalSaleItemCostParts returns a schema-aware item-cost expression and
// joins. Zero line costs are treated as unknown and use the recorded inventory
// or product cost, which is needed for aggregate-stock sales without unit rows.
func (r *Repository) historicalSaleItemCostParts() (string, string) {
	unitCostExpr := "si.unit_cost"
	inventoryCostExpr := "ii.purchase_cost"
	productCostExpr := "p.cost_price"
	joins := " LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id LEFT JOIN products p ON p.id = si.product_id"
	if dbutil.IsSQLite(r.db) {
		unitCostExpr = "0"
		if reportsSQLiteHasColumns(r.db, "sale_items", "unit_cost") {
			unitCostExpr = "si.unit_cost"
		}
		inventoryCostExpr = "0"
		inventoryJoin := ""
		if reportsSQLiteHasColumns(r.db, "sale_items", "inventory_item_id") && reportsSQLiteHasColumns(r.db, "inventory_items", "id", "purchase_cost") {
			inventoryCostExpr = "ii.purchase_cost"
			inventoryJoin = " LEFT JOIN inventory_items ii ON ii.id = si.inventory_item_id"
		}
		productCostExpr = "0"
		productJoin := ""
		if reportsSQLiteHasColumns(r.db, "sale_items", "product_id") && reportsSQLiteHasColumns(r.db, "products", "id") {
			if reportsSQLiteHasColumns(r.db, "products", "cost_price") {
				productCostExpr = "p.cost_price"
			} else if reportsSQLiteHasColumns(r.db, "products", "purchase_price") {
				productCostExpr = "p.purchase_price"
			}
			productJoin = " LEFT JOIN products p ON p.id = si.product_id"
		}
		joins = inventoryJoin + productJoin
	}
	return fmt.Sprintf("COALESCE(NULLIF(%s, 0), NULLIF(%s, 0), %s, 0)", unitCostExpr, inventoryCostExpr, productCostExpr), joins
}

// historicalCOGSTotalSQL returns one COGS value per sale. A nonzero sale cost
// is authoritative; zero is treated as absent so captured item costs can fill
// sales created by the old aggregate-inventory path.
func (r *Repository) historicalCOGSTotalSQL(startPlaceholder, endPlaceholder string) string {
	saleDateExpr := r.salesDateExpression("s")
	lineCostExpr, costJoins := r.historicalSaleItemCostParts()
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "sales", "cost_amount") {
		return fmt.Sprintf(`
			SELECT COALESCE(SUM(sale_cost), 0)
			FROM (
				SELECT s.id, CASE WHEN NULLIF(s.cost_amount, 0) IS NOT NULL THEN s.cost_amount
					ELSE COALESCE(SUM(si.quantity * %s), 0) END AS sale_cost
				FROM sales s
				LEFT JOIN sale_items si ON si.sale_id = s.id
				%s
				WHERE %s >= date(%s)
				  AND %s < date(%s)
				  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
				GROUP BY s.id, s.cost_amount
			) costs`, lineCostExpr, costJoins, saleDateExpr, startPlaceholder, saleDateExpr, endPlaceholder)
	}
	return fmt.Sprintf(`
		SELECT COALESCE(SUM(si.quantity * %s), 0)
		FROM sale_items si JOIN sales s ON s.id = si.sale_id
		%s
		WHERE %s >= date(%s)
		  AND %s < date(%s)
		  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, lineCostExpr, costJoins, saleDateExpr, startPlaceholder, saleDateExpr, endPlaceholder)
}

type returnFinancialAdjustment struct {
	GrossRefund      float64
	RevenueReduction float64
	TaxReduction     float64
	ReturnedCost     float64
	Quantity         int
	Count            int
}

// DailySalesProfit is the tax-exclusive sales revenue and historical COGS for
// one store-calendar day, after completed customer returns are applied.
type DailySalesProfit struct {
	Date    string
	Revenue float64
	COGS    float64
}

// GetDailySalesProfit returns the same daily revenue and COGS basis used by
// profit reports. Keeping the dashboard chart on this calculation prevents it
// from silently falling back to incomplete sale-level costs or ignoring
// completed returns.
func (r *Repository) GetDailySalesProfit(ctx context.Context, startDate, endDate time.Time) ([]DailySalesProfit, error) {
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "sales", "id") {
		return []DailySalesProfit{}, nil
	}
	startDateKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize daily profit start date: %w", err)
	}
	endDateKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize daily profit end date: %w", err)
	}

	salesDate := r.salesDateExpression("s")
	taxExpr := "COALESCE(s.tax_amount, 0)"
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "sales", "tax_amount") {
		taxExpr = "0"
	}
	lineCostExpr, costJoins := r.historicalSaleItemCostParts()
	saleItemsJoin := " LEFT JOIN sale_items si ON si.sale_id = s.id"
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "sale_items", "sale_id", "quantity") {
		saleItemsJoin, costJoins, lineCostExpr = "", "", "0"
	}
	saleCostExpr := "0"
	if saleItemsJoin != "" {
		saleCostExpr = fmt.Sprintf("COALESCE(SUM(COALESCE(si.quantity, 0) * %s), 0)", lineCostExpr)
	}
	groupCost := ""
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "sales", "cost_amount") {
		saleCostExpr = fmt.Sprintf("CASE WHEN NULLIF(s.cost_amount, 0) IS NOT NULL THEN s.cost_amount ELSE %s END", saleCostExpr)
		groupCost = ", s.cost_amount"
	}
	groupTax := ", s.tax_amount"
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "sales", "tax_amount") {
		groupTax = ""
	}
	query := fmt.Sprintf(`WITH per_sale AS (
		SELECT %s AS day, s.id,
		       COALESCE(s.total_amount, 0) - %s AS revenue,
		       %s AS cogs
		FROM sales s
		%s%s
		WHERE %s >= date(?) AND %s < date(?)
		  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		GROUP BY %s, s.id, s.total_amount%s%s
	)
	SELECT day, COALESCE(SUM(revenue), 0), COALESCE(SUM(cogs), 0)
	FROM per_sale GROUP BY day ORDER BY day`, salesDate, taxExpr, saleCostExpr, saleItemsJoin, costJoins, salesDate, salesDate, salesDate, groupTax, groupCost)
	rows, err := r.db.QueryContext(ctx, r.db.Rebind(query), startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("retrieve daily sales profit: %w", err)
	}
	defer rows.Close()

	byDay := make(map[string]*DailySalesProfit)
	for rows.Next() {
		var date reportDate
		var point DailySalesProfit
		if err := rows.Scan(&date, &point.Revenue, &point.COGS); err != nil {
			return nil, fmt.Errorf("scan daily sales profit: %w", err)
		}
		point.Date = date.Format("2006-01-02")
		byDay[point.Date] = &point
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read daily sales profit: %w", err)
	}

	returnsByDay, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "day")
	if err != nil {
		return nil, err
	}
	for day, adjustment := range returnsByDay {
		point := byDay[day]
		if point == nil {
			point = &DailySalesProfit{Date: day}
			byDay[day] = point
		}
		point.Revenue -= adjustment.RevenueReduction
		point.COGS -= adjustment.ReturnedCost
	}

	result := make([]DailySalesProfit, 0, len(byDay))
	for _, point := range byDay {
		result = append(result, *point)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Date < result[j].Date })
	return result, nil
}

// returnFinancialAdjustments calculates completed customer returns once per
// return, while deriving the tax-exclusive revenue and returned cost from its
// individual lines. Joining a return header directly to its lines duplicates
// the header refund for multi-item returns.
func (r *Repository) returnFinancialAdjustments(ctx context.Context, startDate, endDate time.Time, bucket string) (map[string]returnFinancialAdjustment, error) {
	result := make(map[string]returnFinancialAdjustment)
	if dbutil.IsSQLite(r.db) {
		hasReturns, relationErr := reportsSQLiteHasRelation(ctx, r.db, "accounting_returns")
		if relationErr != nil || !hasReturns {
			return result, nil
		}
		if !reportsSQLiteHasColumns(r.db, "accounting_returns", "status", "total_refund_amount") {
			return result, nil
		}
		if !reportsSQLiteHasColumns(r.db, "accounting_returns", "refund_date") &&
			!reportsSQLiteHasColumns(r.db, "accounting_returns", "return_date") &&
			!reportsSQLiteHasColumns(r.db, "accounting_returns", "created_at") {
			return result, nil
		}
	}

	returnDate := r.returnDateExpression("r")

	lineJoin := ""
	lineGross, lineNet, lineTax, lineCost, lineQuantity := "0", "0", "0", "0", "0"
	hasReturnItems := !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_return_items", "return_id", "total_refund_amount")
	if hasReturnItems {
		quantityColumn := "ri.quantity_returned"
		costExpr := "0"
		if dbutil.IsSQLite(r.db) {
			if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity_returned") {
				quantityColumn = "ri.quantity_returned"
			} else if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity") {
				quantityColumn = "ri.quantity"
			} else {
				quantityColumn = "0"
			}
		}
		saleItemJoin := ""
		productJoin := ""
		taxRatio := "1"
		unitCost := "0"
		if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "sale_items", "id", "total_amount", "tax_amount", "unit_cost") {
			saleItemJoin = " LEFT JOIN sale_items si ON si.id = ri.sale_item_id"
			taxRatio = `CASE WHEN COALESCE(si.total_amount, 0) > 0 THEN
				CASE WHEN COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0) > 0
					THEN (COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) / si.total_amount ELSE 0 END
				ELSE 1 END`
			unitCost = "si.unit_cost"
		}
		if dbutil.IsSQLite(r.db) && reportsSQLiteHasColumns(r.db, "sale_items", "id", "unit_cost") {
			saleItemJoin = " LEFT JOIN sale_items si ON si.id = ri.sale_item_id"
			unitCost = "si.unit_cost"
			if reportsSQLiteHasColumns(r.db, "sale_items", "total_amount", "tax_amount") {
				taxRatio = `CASE WHEN COALESCE(si.total_amount, 0) > 0 THEN
					CASE WHEN COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0) > 0
						THEN (COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) / si.total_amount ELSE 0 END
					ELSE 1 END`
			}
		}
		if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "products", "id", "cost_price") {
			productJoin = " LEFT JOIN products p ON p.id = ri.product_id"
		}
		if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_return_items", "original_cost") {
			costExpr = fmt.Sprintf("COALESCE(NULLIF(ri.original_cost, 0), NULLIF(%s, 0)%s, 0)", unitCost, func() string {
				if productJoin != "" {
					return ", p.cost_price"
				}
				return ""
			}())
		} else {
			costExpr = fmt.Sprintf("NULLIF(%s, 0)", unitCost)
			if productJoin != "" {
				costExpr = "COALESCE(" + costExpr + ", p.cost_price, 0)"
			}
		}
		lineJoin = fmt.Sprintf(` LEFT JOIN (
			SELECT ri.return_id,
				COALESCE(SUM(COALESCE(ri.total_refund_amount, 0)), 0) AS line_gross,
				COALESCE(SUM(COALESCE(ri.total_refund_amount, 0) * (%s)), 0) AS line_net,
				COALESCE(SUM(COALESCE(ri.total_refund_amount, 0) * (1 - (%s))), 0) AS line_tax,
				COALESCE(SUM(COALESCE(%s, 0) * COALESCE(%s, 0)), 0) AS line_cost,
				COALESCE(SUM(COALESCE(%s, 0)), 0) AS line_quantity
			FROM accounting_return_items ri%s%s
			GROUP BY ri.return_id
		) lines ON lines.return_id = r.id`, taxRatio, taxRatio, quantityColumn, costExpr, quantityColumn, saleItemJoin, productJoin)
		lineGross, lineNet, lineTax, lineCost, lineQuantity = "COALESCE(lines.line_gross, 0)", "COALESCE(lines.line_net, 0)", "COALESCE(lines.line_tax, 0)", "COALESCE(lines.line_cost, 0)", "COALESCE(lines.line_quantity, 0)"
	}

	bucketExpr := "'total'"
	extraJoin := ""
	switch bucket {
	case "day":
		bucketExpr = "date(" + returnDate + ")"
	case "month":
		if dbutil.IsSQLite(r.db) {
			bucketExpr = "strftime('%Y-%m', date(" + returnDate + "))"
		} else {
			bucketExpr = "TO_CHAR(DATE_TRUNC('month', " + returnDate + "), 'YYYY-MM')"
		}
	case "payment":
		bucketExpr = `COALESCE(NULLIF(TRIM(s.payment_method), ''), 'غير محدد')`
		extraJoin = " LEFT JOIN sales s ON s.id = r.sale_id"
	}

	referenceFilter := ""
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_returns", "reference_number") {
		referenceFilter = ` AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'`
	}
	grouping := " GROUP BY " + bucketExpr + " ORDER BY " + bucketExpr
	if bucket == "total" {
		grouping = ""
	}
	query := fmt.Sprintf(`SELECT %s,
		COALESCE(SUM(COALESCE(r.total_refund_amount, 0)), 0),
		COALESCE(SUM(CASE WHEN %s > 0 THEN COALESCE(r.total_refund_amount, 0) * %s / %s ELSE COALESCE(r.total_refund_amount, 0) END), 0),
		COALESCE(SUM(CASE WHEN %s > 0 THEN COALESCE(r.total_refund_amount, 0) * %s / %s ELSE 0 END), 0),
		COALESCE(SUM(%s), 0), COALESCE(SUM(%s), 0), COUNT(*)
		FROM accounting_returns r%s%s
		WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'%s
		  AND date(%s) >= date(?) AND date(%s) < date(?)
		%s`, bucketExpr, lineGross, lineNet, lineGross, lineGross, lineTax, lineGross, lineCost, lineQuantity, lineJoin, extraJoin, referenceFilter, returnDate, returnDate, grouping)
	startKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize return report start date: %w", err)
	}
	endKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize return report end date: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, r.db.Rebind(query), startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("calculate completed return adjustments: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		var bucketDate reportDate
		var adjustment returnFinancialAdjustment
		keyTarget := any(&key)
		if bucket == "day" {
			// PostgreSQL returns DATE as time.Time through database/sql, while
			// SQLite returns a string. Normalize both to the store date key so
			// daily return credits line up with the daily report buckets.
			keyTarget = &bucketDate
		}
		if err := rows.Scan(keyTarget, &adjustment.GrossRefund, &adjustment.RevenueReduction, &adjustment.TaxReduction, &adjustment.ReturnedCost, &adjustment.Quantity, &adjustment.Count); err != nil {
			return nil, fmt.Errorf("scan completed return adjustments: %w", err)
		}
		if bucket == "day" {
			key = bucketDate.Format("2006-01-02")
		}
		result[key] = adjustment
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read completed return adjustments: %w", err)
	}
	return result, nil
}

type returnCategoryAdjustment struct {
	RevenueReduction float64
	ReturnedCost     float64
}

func (r *Repository) returnCategoryAdjustments(ctx context.Context, startDate, endDate time.Time) (map[string]returnCategoryAdjustment, error) {
	result := make(map[string]returnCategoryAdjustment)
	if dbutil.IsSQLite(r.db) && (!reportsSQLiteHasColumns(r.db, "accounting_returns", "id", "status", "total_refund_amount") ||
		!reportsSQLiteHasColumns(r.db, "accounting_return_items", "return_id", "total_refund_amount")) {
		return result, nil
	}
	quantityColumn := "ri.quantity_returned"
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity_returned") {
		if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity") {
			quantityColumn = "ri.quantity"
		} else {
			quantityColumn = "0"
		}
	}
	returnProduct := "ri.product_id"
	saleItemJoin, inventoryItemJoin := "", ""
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "sale_items", "id", "product_id") {
		saleItemJoin = " LEFT JOIN sale_items si ON si.id = ri.sale_item_id"
	}
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "inventory_items", "id", "product_id") {
		inventoryItemJoin = " LEFT JOIN inventory_items ii ON ii.id = ri.inventory_item_id"
	}
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "accounting_return_items", "product_id") {
		productCandidates := []string{}
		if saleItemJoin != "" {
			productCandidates = append(productCandidates, "si.product_id")
		}
		if inventoryItemJoin != "" {
			productCandidates = append(productCandidates, "ii.product_id")
		}
		if len(productCandidates) == 0 {
			return result, nil
		}
		returnProduct = "COALESCE(" + strings.Join(productCandidates, ", ") + ")"
	} else if saleItemJoin != "" || inventoryItemJoin != "" {
		productCandidates := []string{"ri.product_id"}
		if saleItemJoin != "" {
			productCandidates = append(productCandidates, "si.product_id")
		}
		if inventoryItemJoin != "" {
			productCandidates = append(productCandidates, "ii.product_id")
		}
		returnProduct = "COALESCE(" + strings.Join(productCandidates, ", ") + ")"
	}
	productJoin, categoryJoin := "", ""
	categoryName := "'غير مصنف'"
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "products", "id", "name") {
		productJoin = " LEFT JOIN products p ON p.id = " + returnProduct
		categoryName = "COALESCE(p.name, 'غير مصنف')"
		if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "products", "category_id") {
			if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "categories", "id", "name") {
				categoryJoin = " LEFT JOIN categories c ON c.id = p.category_id"
				categoryName = "COALESCE(NULLIF(TRIM(c.name), ''), p.name, 'غير مصنف')"
			}
		}
	}
	taxRatio := "1"
	unitCost := "0"
	if saleItemJoin != "" {
		if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "sale_items", "total_amount", "tax_amount") {
			taxRatio = `CASE WHEN COALESCE(si.total_amount, 0) > 0
				THEN CASE WHEN COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0) > 0
					THEN (COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) / si.total_amount ELSE 0 END
				ELSE 1 END`
		}
		if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "sale_items", "unit_cost") {
			unitCost = "si.unit_cost"
		}
	}
	costParts := []string{}
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_return_items", "original_cost") {
		costParts = append(costParts, "ri.original_cost")
	}
	if unitCost != "0" {
		costParts = append(costParts, unitCost)
	}
	if productJoin != "" && (!dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "products", "cost_price")) {
		costParts = append(costParts, "p.cost_price")
	}
	costExpr := "0"
	if len(costParts) > 0 {
		costExpr = "COALESCE(" + strings.Join(costParts, ", ") + ", 0)"
	}
	returnDate := r.returnDateExpression("r")
	referenceFilter := ""
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_returns", "reference_number") {
		referenceFilter = `AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'`
	}
	query := fmt.Sprintf(`SELECT %s,
		COALESCE(SUM(COALESCE(ri.total_refund_amount, 0) * (%s)), 0),
		COALESCE(SUM(COALESCE(%s, 0) * %s), 0)
		FROM accounting_returns r
		JOIN accounting_return_items ri ON ri.return_id = r.id%s%s%s%s
		WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED' %s
			AND %s >= date(?) AND %s < date(?)
		GROUP BY %s`, categoryName, taxRatio, quantityColumn, costExpr, saleItemJoin, inventoryItemJoin, productJoin, categoryJoin, referenceFilter, returnDate, returnDate, categoryName)
	startKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize category return start date: %w", err)
	}
	endKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize category return end date: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, r.db.Rebind(query), startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("retrieve returned amounts by category: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var category string
		var adjustment returnCategoryAdjustment
		if err := rows.Scan(&category, &adjustment.RevenueReduction, &adjustment.ReturnedCost); err != nil {
			return nil, fmt.Errorf("scan returned amounts by category: %w", err)
		}
		result[category] = adjustment
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read returned amounts by category: %w", err)
	}
	return result, nil
}

// CreateReport creates a new report
func (r *Repository) CreateReport(ctx context.Context, report *Report) error {
	if strings.EqualFold(r.db.DriverName(), "sqlite") {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO reports (id, type, title, description, parameters,
				data, status, generated_by, generated_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			report.ID, report.Type, report.Title, report.Description, report.Parameters,
			report.Data, report.Status, report.GeneratedBy, report.GeneratedAt, report.CreatedAt, report.UpdatedAt)
		return err
	}
	query := `
		INSERT INTO reports (type, title, description, parameters,
			data, status, generated_by, generated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		report.Type, report.Title, report.Description, report.Parameters,
		report.Data, report.Status, report.GeneratedBy, report.GeneratedAt,
		report.CreatedAt, report.UpdatedAt,
	).Scan(&report.ID, &report.CreatedAt, &report.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}
	return nil
}

// GetReportByID retrieves a report by ID
func (r *Repository) GetReportByID(ctx context.Context, id uuid.UUID) (*Report, error) {
	var report Report
	query := `
		SELECT id, type, title, description, parameters,
			data, status, generated_by, generated_at, created_at, updated_at
		FROM reports
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &report, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReportNotFound
		}
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	return &report, nil
}

// ListReports retrieves reports with pagination and filters
func (r *Repository) ListReports(ctx context.Context, req ReportListRequest) ([]Report, int, error) {
	var reports []Report
	var count int

	// Build base query
	baseQuery := `
		SELECT id, type, title, description, parameters,
			data, status, generated_by, generated_at, created_at, updated_at
		FROM reports
		WHERE 1=1
	`

	countQuery := `
		SELECT COUNT(*)
		FROM reports
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	// Add filters
	if req.Type != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, req.Type)
	}

	if req.Status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}

	if req.GeneratedBy != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND generated_by = $%d", argCount)
		countQuery += fmt.Sprintf(" AND generated_by = $%d", argCount)
		args = append(args, *req.GeneratedBy)
	}

	if req.StartDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND generated_at >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND generated_at >= $%d", argCount)
		args = append(args, *req.StartDate)
	}

	if req.EndDate != nil {
		argCount++
		baseQuery += fmt.Sprintf(" AND generated_at < $%d", argCount)
		countQuery += fmt.Sprintf(" AND generated_at < $%d", argCount)
		args = append(args, *req.EndDate)
	}

	if req.Search != "" {
		argCount++
		like := "ILIKE"
		if dbutil.IsSQLite(r.db) {
			like = "LIKE"
		}
		baseQuery += fmt.Sprintf(" AND (title %s $%d OR description %s $%d)", like, argCount, like, argCount)
		countQuery += fmt.Sprintf(" AND (title %s $%d OR description %s $%d)", like, argCount, like, argCount)
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern)
	}

	// Get total count
	err := r.db.GetContext(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	// Add sorting
	sortColumns := map[string]string{
		"generated_at": "generated_at",
		"created_at":   "created_at",
		"updated_at":   "updated_at",
		"title":        "title",
		"type":         "type",
		"status":       "status",
	}
	sortBy := sortColumns[req.SortBy]
	if sortBy == "" {
		sortBy = "generated_at"
	}
	sortOrder := "DESC"
	if strings.EqualFold(req.SortOrder, "ASC") {
		sortOrder = "ASC"
	}
	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	offset := (req.Page - 1) * req.PerPage
	argCount++
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.PerPage, offset)

	err = r.db.SelectContext(ctx, &reports, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}

	return reports, count, nil
}

// UpdateReport updates a report
func (r *Repository) UpdateReport(ctx context.Context, report *Report) error {
	if strings.EqualFold(r.db.DriverName(), "sqlite") {
		result, err := r.db.ExecContext(ctx, `UPDATE reports SET data = $2, status = $3, updated_at = $4 WHERE id = $1`, report.ID, report.Data, report.Status, report.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to update report: %w", err)
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			return ErrReportNotFound
		}
		return nil
	}
	query := `
		UPDATE reports
		SET data = $2, status = $3, updated_at = $4
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		report.ID, report.Data, report.Status, report.UpdatedAt,
	).Scan(&report.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrReportNotFound
		}
		return fmt.Errorf("failed to update report: %w", err)
	}
	return nil
}

// DeleteReport deletes a report
func (r *Repository) DeleteReport(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reports WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrReportNotFound
	}

	return nil
}

// GetUserName retrieves user name by ID
func (r *Repository) GetUserName(ctx context.Context, userID uuid.UUID) (string, error) {
	var name string
	query := `SELECT first_name || ' ' || last_name as name FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &name, query, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user name: %w", err)
	}
	return name, nil
}

// GetSalesData retrieves sales data for report
func (r *Repository) GetSalesData(ctx context.Context, startDate, endDate time.Time) (*SalesReport, error) {
	var report SalesReport
	report.StartDate = startDate
	report.EndDate = endDate

	startDateKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize sales report start date: %w", err)
	}
	endDateKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize sales report end date: %w", err)
	}

	var totals struct {
		TotalSales     int     `db:"total_sales"`
		TotalRevenue   float64 `db:"total_revenue"`
		TotalTax       float64 `db:"total_tax"`
		TotalCOGS      float64 `db:"total_cogs"`
		TotalItemsSold int     `db:"total_items_sold"`
		CashRevenue    float64 `db:"cash_revenue"`
		CreditRevenue  float64 `db:"credit_revenue"`
		TotalPaid      float64 `db:"total_paid"`
		CashReceived   float64 `db:"cash_received"`
		ChangeAmount   float64 `db:"change_amount"`
	}

	if dbutil.IsSQLite(r.db) {
		saleDateExpr := r.salesDateExpression("")
		paidAmountExpr := "0"
		cashReceivedExpr := "0"
		changeAmountExpr := "0"
		if reportsSQLiteHasColumns(r.db, "sales", "paid_amount") {
			paidAmountExpr = "COALESCE(paid_amount, 0)"
		}
		if reportsSQLiteHasColumns(r.db, "sales", "cash_received") {
			cashReceivedExpr = "COALESCE(cash_received, 0)"
		}
		if reportsSQLiteHasColumns(r.db, "sales", "change_amount") {
			changeAmountExpr = "COALESCE(change_amount, 0)"
		}
		err = r.db.GetContext(ctx, &totals, fmt.Sprintf(`
			SELECT COUNT(*) AS total_sales,
				COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) AS total_revenue,
				COALESCE(SUM(COALESCE(tax_amount, 0)), 0) AS total_tax,
				COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('cash', 'card') THEN COALESCE(total_amount, 0) - COALESCE(tax_amount, 0) ELSE 0 END), 0) AS cash_revenue,
				COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('credit', 'debt') THEN COALESCE(total_amount, 0) - COALESCE(tax_amount, 0) ELSE 0 END), 0) AS credit_revenue,
				COALESCE(SUM(%s), 0) AS total_paid,
				COALESCE(SUM(%s), 0) AS cash_received,
				COALESCE(SUM(%s), 0) AS change_amount
			FROM sales
			WHERE %s >= date(?) AND %s < date(?)
				AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, paidAmountExpr, cashReceivedExpr, changeAmountExpr, saleDateExpr, saleDateExpr), startDateKey, endDateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve sales totals: %w", err)
		}

		err = r.db.GetContext(ctx, &totals.TotalCOGS, r.historicalCOGSTotalSQL("?", "?"), startDateKey, endDateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate sales COGS: %w", err)
		}

		err = r.db.GetContext(ctx, &totals.TotalItemsSold, fmt.Sprintf(`
			SELECT COALESCE(SUM(si.quantity), 0)
			FROM sale_items si
			JOIN sales s ON s.id = si.sale_id
			WHERE %s >= date(?) AND %s < date(?)
			  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, r.salesDateExpression("s"), r.salesDateExpression("s")), startDateKey, endDateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to count sold items: %w", err)
		}
	} else {
		itemCOGSQuery := r.historicalCOGSTotalSQL("$1", "$2")
		salesTotalsQuery := fmt.Sprintf(`
			SELECT st.total_sales, st.total_revenue, st.total_tax, st.total_paid, st.cash_received, st.change_amount, it.total_cogs, it.total_items_sold,
				st.cash_revenue, st.credit_revenue
			FROM (
				SELECT COUNT(*) AS total_sales,
					COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) AS total_revenue,
					COALESCE(SUM(COALESCE(tax_amount, 0)), 0) AS total_tax,
					COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('cash', 'card') THEN COALESCE(total_amount, 0) - COALESCE(tax_amount, 0) ELSE 0 END), 0) AS cash_revenue,
					COALESCE(SUM(CASE WHEN LOWER(COALESCE(payment_method, '')) IN ('credit', 'debt') THEN COALESCE(total_amount, 0) - COALESCE(tax_amount, 0) ELSE 0 END), 0) AS credit_revenue,
					COALESCE(SUM(COALESCE(paid_amount, 0)), 0) AS total_paid,
					COALESCE(SUM(COALESCE(cash_received, 0)), 0) AS cash_received,
					COALESCE(SUM(COALESCE(change_amount, 0)), 0) AS change_amount
				FROM sales
				WHERE date(COALESCE(sale_date, created_at)) >= date($1) AND date(COALESCE(sale_date, created_at)) < date($2)
					AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			) st
			CROSS JOIN (
				SELECT (%s) AS total_cogs,
					(SELECT COALESCE(SUM(si.quantity), 0)
						 FROM sale_items si JOIN sales s ON s.id = si.sale_id
						 WHERE date(COALESCE(s.sale_date, s.created_at)) >= date($3) AND date(COALESCE(s.sale_date, s.created_at)) < date($4)
						   AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')) AS total_items_sold
			) it`,
			itemCOGSQuery)
		salesTotalsQuery = strings.ReplaceAll(salesTotalsQuery, "date(COALESCE(sale_date, created_at))", r.salesDateExpression(""))
		salesTotalsQuery = strings.ReplaceAll(salesTotalsQuery, "date(COALESCE(s.sale_date, s.created_at))", r.salesDateExpression("s"))
		err = r.db.GetContext(ctx, &totals, salesTotalsQuery, startDateKey, endDateKey, startDateKey, endDateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve sales totals: %w", err)
		}
	}

	// Completed returns are counted once per return and split into tax and
	// revenue portions using the original sale lines.
	returnAdjustments, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "total")
	if err != nil {
		return nil, err
	}
	periodReturns := returnAdjustments["total"]
	report.TotalSales = totals.TotalSales
	report.TotalRevenue = totals.TotalRevenue - periodReturns.RevenueReduction
	report.TotalTax = totals.TotalTax - periodReturns.TaxReduction
	report.TotalCOGS = totals.TotalCOGS - periodReturns.ReturnedCost
	report.TotalItemsSold = totals.TotalItemsSold - periodReturns.Quantity
	report.CashRevenue = totals.CashRevenue
	report.CreditRevenue = totals.CreditRevenue
	report.TotalPaid = totals.TotalPaid
	report.CashReceived = totals.CashReceived
	report.ChangeAmount = totals.ChangeAmount
	report.GrossProfit = report.TotalRevenue - report.TotalCOGS
	if report.TotalRevenue > 0 {
		report.ProfitMargin = (report.GrossProfit / report.TotalRevenue) * 100
	} else {
		report.ProfitMargin = 0
	}

	dailySalesDate := r.salesDateExpression("")
	dailySalesQuery := fmt.Sprintf(`SELECT %s as date, COUNT(*) as sales, COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) as revenue
		 FROM sales 
		 WHERE %s >= date(?) AND %s < date(?)
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		GROUP BY %s
		ORDER BY date`, dailySalesDate, dailySalesDate, dailySalesDate, dailySalesDate)
	rows, err := r.db.QueryContext(ctx, r.db.Rebind(dailySalesQuery), startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve daily sales trend: %w", err)
	}
	for rows.Next() {
		var daily DailySales
		var date reportDate
		if err := rows.Scan(&date, &daily.Sales, &daily.Revenue); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan daily sales trend: %w", err)
		}
		daily.Date = date.Time
		report.ByDay = append(report.ByDay, daily)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read daily sales trend: %w", err)
	}
	rows.Close()
	dailyReturns, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "day")
	if err != nil {
		return nil, err
	}
	dailySalesIndex := make(map[string]int, len(report.ByDay))
	for index := range report.ByDay {
		dailySalesIndex[report.ByDay[index].Date.Format("2006-01-02")] = index
	}
	for dayKey, adjustment := range dailyReturns {
		if index, ok := dailySalesIndex[dayKey]; ok {
			report.ByDay[index].Revenue -= adjustment.RevenueReduction
			continue
		}
		day, parseErr := time.Parse("2006-01-02", dayKey)
		if parseErr == nil {
			report.ByDay = append(report.ByDay, DailySales{Date: day, Revenue: -adjustment.RevenueReduction})
		}
	}
	sort.Slice(report.ByDay, func(i, j int) bool { return report.ByDay[i].Date.Before(report.ByDay[j].Date) })

	report.TopProducts = []ProductSales{}
	returnedJoin := ""
	returnedQuantityExpr := "0"
	returnedRefundExpr := "0"
	returnedCostExpr := "0"
	if !dbutil.IsSQLite(r.db) {
		returnedJoin = fmt.Sprintf(`LEFT JOIN (
			SELECT COALESCE(ri.product_id, si2.product_id, ii2.product_id) AS product_id,
				SUM(COALESCE(ri.quantity_returned, 0)) AS quantity,
				SUM(COALESCE(ri.total_refund_amount, 0) * CASE WHEN COALESCE(si2.total_amount, 0) > 0
					THEN GREATEST((COALESCE(si2.total_amount, 0) - COALESCE(si2.tax_amount, 0)) / si2.total_amount, 0)
					ELSE 1 END) AS refund_amount,
				SUM(COALESCE(ri.quantity_returned, 0) * COALESCE(ri.original_cost, si2.unit_cost, 0)) AS returned_cost
			FROM accounting_return_items ri
			JOIN accounting_returns r ON r.id = ri.return_id
			LEFT JOIN sale_items si2 ON si2.id = ri.sale_item_id
			LEFT JOIN inventory_items ii2 ON ii2.id = ri.inventory_item_id
			WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
				AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%%'
				AND COALESCE(ri.total_refund_amount, 0) >= 0
				AND %s >= date(?)
				AND %s < date(?)
			GROUP BY COALESCE(ri.product_id, si2.product_id, ii2.product_id)
		) returned ON returned.product_id = si.product_id`, r.returnDateExpression("r"), r.returnDateExpression("r"))
		returnedQuantityExpr = "COALESCE(MAX(returned.quantity), 0)"
		returnedRefundExpr = "COALESCE(MAX(returned.refund_amount), 0)"
		returnedCostExpr = "COALESCE(MAX(returned.returned_cost), 0)"
	} else {
		var returnItemsTableCount int
		_ = r.db.Get(&returnItemsTableCount, `SELECT COUNT(*) FROM sqlite_master WHERE type IN ('table','view') AND name = 'accounting_return_items'`)
		returnQuantityColumn := ""
		if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity_returned") {
			returnQuantityColumn = "ri.quantity_returned"
		} else if reportsSQLiteHasColumns(r.db, "accounting_return_items", "quantity") {
			returnQuantityColumn = "ri.quantity"
		}
		returnCostColumn := "0"
		if reportsSQLiteHasColumns(r.db, "accounting_return_items", "original_cost") {
			returnCostColumn = "ri.original_cost"
		}
		if returnItemsTableCount > 0 && returnQuantityColumn != "" && reportsSQLiteHasColumns(r.db, "accounting_return_items", "product_id", "total_refund_amount") && reportsSQLiteHasColumns(r.db, "accounting_returns", "status", "total_refund_amount") {
			returnDate := r.returnDateExpression("r")
			taxRatio := "1"
			if reportsSQLiteHasColumns(r.db, "sale_items", "id", "total_amount", "tax_amount") {
				taxRatio = `CASE WHEN COALESCE(si2.total_amount, 0) > 0
					THEN MAX((COALESCE(si2.total_amount, 0) - COALESCE(si2.tax_amount, 0)) / si2.total_amount, 0)
					ELSE 1 END`
			}
			referenceFilter := ""
			if reportsSQLiteHasColumns(r.db, "accounting_returns", "reference_number") {
				referenceFilter = `AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'`
			}
			returnedJoin = fmt.Sprintf(`LEFT JOIN (
				SELECT COALESCE(ri.product_id, si2.product_id, ii2.product_id) AS product_id,
					SUM(COALESCE(%s, 0)) AS quantity,
					SUM(COALESCE(ri.total_refund_amount, 0) * (%s)) AS refund_amount,
					SUM(COALESCE(%s, 0) * COALESCE(%s, si2.unit_cost, 0)) AS returned_cost
				FROM accounting_return_items ri
				JOIN accounting_returns r ON r.id = ri.return_id
				LEFT JOIN sale_items si2 ON si2.id = ri.sale_item_id
				LEFT JOIN inventory_items ii2 ON ii2.id = ri.inventory_item_id
				WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
					%s
					AND COALESCE(ri.total_refund_amount, 0) >= 0
					AND %s >= date(?)
					AND %s < date(?)
				GROUP BY COALESCE(ri.product_id, si2.product_id, ii2.product_id)
			) returned ON returned.product_id = si.product_id`, returnQuantityColumn, taxRatio, returnQuantityColumn, returnCostColumn, referenceFilter, returnDate, returnDate)
			returnedQuantityExpr = "COALESCE(MAX(returned.quantity), 0)"
			returnedRefundExpr = "COALESCE(MAX(returned.refund_amount), 0)"
			returnedCostExpr = "COALESCE(MAX(returned.returned_cost), 0)"
		}
	}
	topProductsQuery := fmt.Sprintf(`
		SELECT p.id, p.name,
			COALESCE(SUM(si.quantity), 0) - %s,
			COALESCE(SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)), 0) - %s,
			COALESCE(SUM((COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) - (si.quantity * COALESCE(si.unit_cost, 0))), 0)
				- %s + %s
		FROM sale_items si
		JOIN sales s ON s.id = si.sale_id
		JOIN products p ON p.id = si.product_id
		%s
		WHERE %s >= date(?) AND %s < date(?)
			AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		GROUP BY p.id, p.name
		ORDER BY (COALESCE(SUM((COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) - (si.quantity * COALESCE(si.unit_cost, 0))), 0)
			- %s + %s) DESC
		LIMIT 10`, returnedQuantityExpr, returnedRefundExpr, returnedRefundExpr, returnedCostExpr, returnedJoin, r.salesDateExpression("s"), r.salesDateExpression("s"), returnedRefundExpr, returnedCostExpr)
	returnArgs := []interface{}{startDateKey, endDateKey}
	if returnedJoin == "" {
		returnArgs = nil
	}
	returnArgs = append(returnArgs, startDateKey, endDateKey)
	rows, err = r.db.QueryContext(ctx, r.db.Rebind(topProductsQuery), returnArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve top products: %w", err)
	}
	for rows.Next() {
		var product ProductSales
		if err := rows.Scan(&product.ProductID, &product.ProductName, &product.Quantity, &product.Revenue, &product.Profit); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan top product: %w", err)
		}
		report.TopProducts = append(report.TopProducts, product)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read top products: %w", err)
	}
	rows.Close()

	report.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx, r.db.Rebind(
		`SELECT COALESCE(NULLIF(TRIM(payment_method), ''), 'غير محدد'),
			COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0)
		 FROM sales
		 WHERE `+r.salesDateExpression("")+` >= date(?) AND `+r.salesDateExpression("")+` < date(?)
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY COALESCE(NULLIF(TRIM(payment_method), ''), 'غير محدد')`),
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sales by payment method: %w", err)
	}
	for rows.Next() {
		var method string
		var total float64
		if err := rows.Scan(&method, &total); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan sales by payment method: %w", err)
		}
		report.ByPaymentMethod[method] = total
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read sales by payment method: %w", err)
	}
	rows.Close()

	refundByPaymentMethod, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "payment")
	if err != nil {
		return nil, err
	}
	for method, adjustment := range refundByPaymentMethod {
		report.ByPaymentMethod[method] -= adjustment.RevenueReduction
	}

	return &report, nil
}

// GetInventoryData retrieves inventory data for report
func (r *Repository) GetInventoryData(ctx context.Context) (*InventoryReport, error) {
	var report InventoryReport
	storeNow := accounting.StoreNow()
	storeDate, err := accounting.StoreDate(storeNow)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate inventory report date: %w", err)
	}
	overstockStart, _, err := accounting.StoreDateRange(storeNow, 90)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate overstock report period: %w", err)
	}

	report.ByCondition = make(map[string]int)
	report.ByCategory = make(map[string]int)
	report.Items = []InventoryItem{}
	report.LowStockItems = []LowStockItem{}
	report.OverstockItems = []OverstockItem{}
	report.StagnantItems = []StagnantItem{}
	report.Valuation.ByCondition = make(map[string]float64)

	err = r.db.GetContext(ctx, &report.TotalItems,
		`SELECT COALESCE(SUM(stock), 0) FROM (
			SELECT p.id, COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0) AS stock
			FROM products p
			LEFT JOIN (SELECT product_id, SUM(quantity) AS quantity FROM inventory GROUP BY product_id) inv ON inv.product_id = p.id
			LEFT JOIN inventory_items ii ON ii.product_id = p.id
			WHERE p.deleted_at IS NULL
			GROUP BY p.id, inv.quantity
		) stock_totals WHERE stock > 0`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory totals: %w", err)
	}

	err = r.db.GetContext(ctx, &report.TotalValue,
		`SELECT COALESCE(SUM(stock * p.cost_price), 0) FROM (
			SELECT p.id, COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0) AS stock
			FROM products p
			LEFT JOIN (SELECT product_id, SUM(quantity) AS quantity FROM inventory GROUP BY product_id) inv ON inv.product_id = p.id
			LEFT JOIN inventory_items ii ON ii.product_id = p.id
			WHERE p.deleted_at IS NULL
			GROUP BY p.id, inv.quantity
		) stock_totals JOIN products p ON p.id = stock_totals.id WHERE stock > 0`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory value: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`WITH available_items AS (
		     SELECT product_id, COUNT(*) AS quantity
		     FROM inventory_items
		     WHERE status = 'AVAILABLE' AND UPPER(COALESCE(condition, '')) <> 'USED'
		     GROUP BY product_id
		 )
		 SELECT stock.condition, COALESCE(SUM(stock.quantity), 0),
		        COALESCE(SUM(stock.quantity * stock.unit_cost), 0)
		 FROM (
		     SELECT COALESCE(NULLIF(UPPER(ii.condition), ''), 'NEW') AS condition,
		            1 AS quantity, COALESCE(p.cost_price, 0) AS unit_cost
		     FROM inventory_items ii
		     JOIN products p ON p.id = ii.product_id
		     LEFT JOIN (SELECT product_id, SUM(quantity) AS quantity FROM inventory GROUP BY product_id) inv ON inv.product_id = p.id
		     LEFT JOIN available_items ai ON ai.product_id = p.id
		     WHERE ii.status = 'AVAILABLE'
		       AND UPPER(COALESCE(ii.condition, '')) <> 'USED'
		       AND p.deleted_at IS NULL
		       AND (inv.quantity IS NULL OR ai.quantity <= inv.quantity)
		     UNION ALL
		     SELECT 'NEW' AS condition,
		            CASE WHEN COALESCE(ai.quantity, 0) <= inv.quantity
		                 THEN inv.quantity - COALESCE(ai.quantity, 0)
		                 ELSE inv.quantity END AS quantity,
		            COALESCE(p.cost_price, 0) AS unit_cost
			 FROM (SELECT product_id, SUM(quantity) AS quantity FROM inventory GROUP BY product_id) inv
		     JOIN products p ON p.id = inv.product_id
		     LEFT JOIN available_items ai ON ai.product_id = p.id
		     WHERE inv.quantity > 0 AND p.deleted_at IS NULL
		 ) stock
		 WHERE stock.quantity > 0
		 GROUP BY stock.condition`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory conditions: %w", err)
	}
	for rows.Next() {
		var condition string
		var count int
		var value float64
		if err := rows.Scan(&condition, &count, &value); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan inventory condition: %w", err)
		}
		report.ByCondition[condition] = count
		report.Valuation.ByCondition[condition] = value
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read inventory conditions: %w", err)
	}
	rows.Close()
	report.Valuation.TotalCost = report.TotalValue
	err = r.db.GetContext(ctx, &report.Valuation.TotalRetail,
		`SELECT COALESCE(SUM(stock * p.selling_price), 0) FROM (
			SELECT p.id, COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0) AS stock
			FROM products p
			LEFT JOIN (SELECT product_id, SUM(quantity) AS quantity FROM inventory GROUP BY product_id) inv ON inv.product_id = p.id
			LEFT JOIN inventory_items ii ON ii.product_id = p.id
			WHERE p.deleted_at IS NULL
			GROUP BY p.id, inv.quantity
		) stock_totals JOIN products p ON p.id = stock_totals.id WHERE stock > 0`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve retail inventory value: %w", err)
	}
	var taxRate float64
	if err := r.db.GetContext(ctx, &taxRate, `SELECT COALESCE((SELECT CAST(value AS FLOAT) FROM settings WHERE key = 'tax_rate' LIMIT 1), 0)`); err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory tax rate: %w", err)
	}
	report.Valuation.TotalRetailWithTax = report.Valuation.TotalRetail * (1 + taxRate/100)
	report.Valuation.PotentialProfit = report.Valuation.TotalRetail - report.Valuation.TotalCost

	rows, err = r.db.QueryContext(ctx,
		`SELECT p.id, p.name,
				COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0),
				p.min_stock_level
		 FROM products p
		 LEFT JOIN (SELECT product_id, SUM(quantity) AS quantity FROM inventory GROUP BY product_id) inv ON inv.product_id = p.id
		 LEFT JOIN inventory_items ii ON ii.product_id = p.id
		 WHERE p.is_active = true AND p.deleted_at IS NULL
		 GROUP BY p.id, p.name, p.min_stock_level, inv.quantity
		 HAVING COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0) > 0
		 ORDER BY p.name`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory items: %w", err)
	}
	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.CurrentStock, &item.MinStock); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		report.Items = append(report.Items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read inventory items: %w", err)
	}
	rows.Close()

	rows, err = r.db.QueryContext(ctx,
		`SELECT p.id, p.name,
					 COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0),
		        p.min_stock_level, p.min_stock_level
		 FROM products p
		 LEFT JOIN (SELECT product_id, SUM(quantity) AS quantity FROM inventory GROUP BY product_id) inv ON inv.product_id = p.id
		 LEFT JOIN inventory_items ii ON ii.product_id = p.id
		 WHERE p.is_active = true AND p.deleted_at IS NULL AND p.min_stock_level > 0
		 GROUP BY p.id, p.name, p.min_stock_level, inv.quantity
			 HAVING COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0) <= p.min_stock_level
			 ORDER BY (p.min_stock_level - COALESCE(inv.quantity, SUM(CASE WHEN ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' THEN 1 ELSE 0 END), 0)) DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve low stock items: %w", err)
	}
	for rows.Next() {
		var item LowStockItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.CurrentStock, &item.MinStock, &item.ReorderLevel); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan low stock item: %w", err)
		}
		report.LowStockItems = append(report.LowStockItems, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read low stock items: %w", err)
	}
	rows.Close()
	rows, err = r.db.QueryContext(ctx,
		`SELECT COALESCE(c.name, 'غير مصنف'), COUNT(*)
		 FROM inventory_items ii
		 JOIN products p ON p.id = ii.product_id
		 LEFT JOIN categories c ON c.id = p.category_id
			 WHERE ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED' AND p.deleted_at IS NULL
		 GROUP BY c.name ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory categories: %w", err)
	}
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan inventory category: %w", err)
		}
		report.ByCategory[category] = count
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read inventory categories: %w", err)
	}
	rows.Close()

	// Category totals must use the same product stock source as TotalItems:
	// inventory.quantity is authoritative when present, with serialized stock
	// as the fallback for older products that have no aggregate row.
	report.ByCategory = make(map[string]int)
	rows, err = r.db.QueryContext(ctx, `WITH serialized_stock AS (
			SELECT product_id, COUNT(*) AS quantity
			FROM inventory_items
			WHERE status = 'AVAILABLE' AND UPPER(COALESCE(condition, '')) <> 'USED'
			GROUP BY product_id
		), product_stock AS (
			SELECT p.id, p.category_id, COALESCE(inv.quantity, serialized_stock.quantity, 0) AS quantity
			FROM products p
			LEFT JOIN (SELECT product_id, SUM(quantity) AS quantity FROM inventory GROUP BY product_id) inv ON inv.product_id = p.id
			LEFT JOIN serialized_stock ON serialized_stock.product_id = p.id
			WHERE p.deleted_at IS NULL
		)
		SELECT COALESCE(NULLIF(TRIM(c.name), ''), 'غير مصنف'), COALESCE(SUM(stock.quantity), 0)
		FROM product_stock stock
		LEFT JOIN categories c ON c.id = stock.category_id
		WHERE stock.quantity > 0
		GROUP BY COALESCE(NULLIF(TRIM(c.name), ''), 'غير مصنف')
		ORDER BY SUM(stock.quantity) DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve inventory category totals: %w", err)
	}
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan inventory category total: %w", err)
		}
		report.ByCategory[category] = count
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read inventory category totals: %w", err)
	}
	rows.Close()

	overstockQuery := `WITH serialized_stock AS (
			SELECT product_id, COUNT(*) AS quantity
			FROM inventory_items
			WHERE status = 'AVAILABLE' AND UPPER(COALESCE(condition, '')) <> 'USED'
			GROUP BY product_id
		), inventory_stock AS (
			SELECT product_id, SUM(quantity) AS quantity
			FROM inventory
			GROUP BY product_id
		), product_stock AS (
			SELECT p.id, p.name, COALESCE(inventory_stock.quantity, serialized_stock.quantity, 0) AS quantity
			FROM products p
			LEFT JOIN inventory_stock ON inventory_stock.product_id = p.id
			LEFT JOIN serialized_stock ON serialized_stock.product_id = p.id
			WHERE p.is_active = true AND p.deleted_at IS NULL
		), recent_sales AS (
			SELECT si.product_id, COALESCE(SUM(si.quantity), 0) AS quantity
			FROM sale_items si
			JOIN sales s ON s.id = si.sale_id
			WHERE ` + r.salesDateExpression("s") + ` >= ?
			  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			GROUP BY si.product_id
		)
		SELECT p.id, p.name, p.quantity,
		       COALESCE(recent_sales.quantity, 0) / 3.0,
		       p.quantity / NULLIF(COALESCE(recent_sales.quantity, 0) / 3.0, 0)
		FROM product_stock p
		LEFT JOIN recent_sales ON recent_sales.product_id = p.id
		WHERE p.quantity > 0
		  AND (COALESCE(recent_sales.quantity, 0) = 0 OR p.quantity / NULLIF(COALESCE(recent_sales.quantity, 0) / 3.0, 0) >= 6)
		ORDER BY p.quantity DESC`
	if !dbutil.IsSQLite(r.db) {
		overstockQuery = strings.ReplaceAll(overstockQuery, "?", "$1::date")
	}
	overstockArgs := []any{overstockStart}
	rows, err = r.db.QueryContext(ctx, overstockQuery, overstockArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve overstock items: %w", err)
	}
	for rows.Next() {
		var item OverstockItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.CurrentStock, &item.AvgMonthlySales, &item.MonthsOfSupply); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan overstock item: %w", err)
		}
		report.OverstockItems = append(report.OverstockItems, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read overstock items: %w", err)
	}
	rows.Close()

	latestSaleDate := "MAX(" + r.salesDateExpression("s") + ")"
	latestItemDate := "MAX(store_date(ii.created_at))"
	latestMovementDate := "MAX(store_date(m.created_at))"
	inventoryRestockDate := "store_date(inventory_stock.last_restocked_at)"
	inventoryCreatedDate := "store_date(inventory_stock.created_at)"
	productCreatedDate := "store_date(p.created_at)"
	activityDate := "MAX(COALESCE(sales_activity.last_sale_date, '0001-01-01'), COALESCE(item_activity.last_item_date, '0001-01-01'), COALESCE(movement_activity.last_movement_date, '0001-01-01'), COALESCE(" + inventoryRestockDate + ", " + inventoryCreatedDate + ", " + productCreatedDate + ", '0001-01-01'))"
	daysSinceActivity := "CAST(julianday(?) - julianday(stock_activity.last_activity_date) AS INTEGER)"
	if !dbutil.IsSQLite(r.db) {
		latestItemDate = postgresStoreTimestampDateExpression("MAX(ii.created_at)")
		latestMovementDate = postgresStoreTimestampDateExpression("MAX(m.created_at)")
		inventoryRestockDate = postgresStoreTimestampDateExpression("inventory_stock.last_restocked_at")
		inventoryCreatedDate = postgresStoreTimestampDateExpression("inventory_stock.created_at")
		productCreatedDate = postgresStoreTimestampDateExpression("p.created_at")
		activityDate = "GREATEST(COALESCE(sales_activity.last_sale_date, DATE '0001-01-01'), COALESCE(item_activity.last_item_date, DATE '0001-01-01'), COALESCE(movement_activity.last_movement_date, DATE '0001-01-01'), COALESCE(" + inventoryRestockDate + ", " + inventoryCreatedDate + ", " + productCreatedDate + ", DATE '0001-01-01'))"
		daysSinceActivity = "$1::date - stock_activity.last_activity_date"
	}
	stagnantQuery := `WITH serialized_stock AS (
			SELECT product_id, COUNT(*) AS quantity, MAX(created_at) AS latest_item_created_at
			FROM inventory_items
			WHERE status = 'AVAILABLE' AND UPPER(COALESCE(condition, '')) <> 'USED'
			GROUP BY product_id
		), inventory_stock AS (
			SELECT product_id, SUM(quantity) AS quantity,
			       MAX(last_restocked_at) AS last_restocked_at, MAX(created_at) AS created_at
			FROM inventory
			GROUP BY product_id
		), sales_activity AS (
			SELECT si.product_id, ` + latestSaleDate + ` AS last_sale_date
			FROM sale_items si
			JOIN sales s ON s.id = si.sale_id
			WHERE LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			GROUP BY si.product_id
		), item_activity AS (
			SELECT ii.product_id, ` + latestItemDate + ` AS last_item_date
			FROM inventory_items ii
			WHERE ii.status = 'AVAILABLE' AND UPPER(COALESCE(ii.condition, '')) <> 'USED'
			GROUP BY ii.product_id
		), movement_activity AS (
			SELECT m.product_id, ` + latestMovementDate + ` AS last_movement_date
			FROM inventory_movements m
			WHERE m.product_id IS NOT NULL AND m.quantity > 0
			GROUP BY m.product_id
		), stock_activity AS (
			SELECT p.id, p.name,
			       COALESCE(inventory_stock.quantity, serialized_stock.quantity, 0) AS quantity,
			       ` + activityDate + ` AS last_activity_date,
			       COALESCE(p.cost_price, 0) AS unit_cost
			FROM products p
			LEFT JOIN inventory_stock ON inventory_stock.product_id = p.id
			LEFT JOIN serialized_stock ON serialized_stock.product_id = p.id
			LEFT JOIN sales_activity ON sales_activity.product_id = p.id
			LEFT JOIN item_activity ON item_activity.product_id = p.id
			LEFT JOIN movement_activity ON movement_activity.product_id = p.id
			WHERE p.is_active = true AND p.deleted_at IS NULL
		)
		SELECT stock_activity.id, stock_activity.name, stock_activity.quantity,
		       stock_activity.last_activity_date, ` + daysSinceActivity + `,
		       stock_activity.quantity * stock_activity.unit_cost
		FROM stock_activity
		WHERE stock_activity.quantity > 0 AND ` + daysSinceActivity + ` >= 30
		ORDER BY ` + daysSinceActivity + ` DESC`
	stagnantArgs := []any{storeDate}
	if dbutil.IsSQLite(r.db) {
		stagnantArgs = []any{storeDate, storeDate, storeDate}
	}
	rows, err = r.db.QueryContext(ctx, stagnantQuery, stagnantArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve stagnant inventory items: %w", err)
	}
	for rows.Next() {
		var item StagnantItem
		var lastSale reportDate
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.CurrentStock, &lastSale, &item.DaysSinceSale, &item.Value); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan stagnant inventory item: %w", err)
		}
		item.LastSaleDate = lastSale.Time
		report.StagnantItems = append(report.StagnantItems, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read stagnant inventory items: %w", err)
	}
	rows.Close()

	return &report, nil
}

// GetExpensesData retrieves expenses data for report
func (r *Repository) GetExpensesData(ctx context.Context, startDate, endDate time.Time) (*ExpensesReport, error) {
	var report ExpensesReport
	report.StartDate = startDate
	report.EndDate = endDate
	startDateKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize expenses report start date: %w", err)
	}
	endDateKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize expenses report end date: %w", err)
	}
	report.ByCategory = make(map[string]float64)
	report.ByPaymentMethod = make(map[string]float64)
	report.ByMonth = []MonthlyExpenses{}
	report.TopExpenses = []ExpenseItem{}
	expensePeriod := "date(expense_date) >= date(substr($1, 1, 10)) AND date(expense_date) < date(substr($2, 1, 10))"
	if dbutil.IsSQLite(r.db) {
		expensePeriod = "date(substr(expense_date, 1, 10)) >= date(substr($1, 1, 10)) AND date(substr(expense_date, 1, 10)) < date(substr($2, 1, 10))"
	}
	expenseStatus := accounting.AccountingExpenseStatusSQL()

	report.TotalExpenses, err = accounting.AccountingExpensesForPeriod(ctx, r.db, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve expenses total: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT category, COALESCE(SUM(amount), 0) as total
		 FROM expenses
		 WHERE %s AND %s
		 GROUP BY category`, expensePeriod, expenseStatus),
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve expenses by category: %w", err)
	}
	for rows.Next() {
		var category string
		var total float64
		if err := rows.Scan(&category, &total); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan expenses by category: %w", err)
		}
		report.ByCategory[category] = total
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read expenses by category: %w", err)
	}
	rows.Close()

	rows, err = r.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT payment_method, COALESCE(SUM(amount), 0) as total
		 FROM expenses
		 WHERE %s AND %s
		 GROUP BY payment_method`, expensePeriod, expenseStatus),
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve expenses by payment method: %w", err)
	}
	for rows.Next() {
		var method string
		var total float64
		if err := rows.Scan(&method, &total); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan expenses by payment method: %w", err)
		}
		report.ByPaymentMethod[method] = total
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read expenses by payment method: %w", err)
	}
	rows.Close()

	monthlyExpensesQuery := fmt.Sprintf(`SELECT DATE_TRUNC('month', expense_date), COALESCE(SUM(amount), 0)
		 FROM expenses
		 WHERE %s AND %s
		 GROUP BY DATE_TRUNC('month', expense_date)
		 ORDER BY DATE_TRUNC('month', expense_date)`, expensePeriod, expenseStatus)
	if dbutil.IsSQLite(r.db) {
		monthlyExpensesQuery = fmt.Sprintf(`SELECT strftime('%%Y-%%m-01', substr(expense_date, 1, 10)), COALESCE(SUM(amount), 0)
		 FROM expenses
		 WHERE %s AND %s
		 GROUP BY strftime('%%Y-%%m', substr(expense_date, 1, 10))
		 ORDER BY strftime('%%Y-%%m', substr(expense_date, 1, 10))`, expensePeriod, expenseStatus)
	}
	rows, err = r.db.QueryContext(ctx, monthlyExpensesQuery, startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve monthly expenses: %w", err)
	}
	for rows.Next() {
		var monthly MonthlyExpenses
		var month reportDate
		if err := rows.Scan(&month, &monthly.Amount); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan monthly expenses: %w", err)
		}
		monthly.Month = month.Time
		report.ByMonth = append(report.ByMonth, monthly)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read monthly expenses: %w", err)
	}
	rows.Close()

	topExpensesQuery := fmt.Sprintf(`SELECT expense_date, amount,
		COALESCE(NULLIF(description, ''), NULLIF(title, ''), 'مصروف') AS description,
		COALESCE(status, 'approved')
		FROM expenses
		WHERE %s AND %s
		ORDER BY amount DESC, date(expense_date) DESC
		LIMIT 20`, expensePeriod, expenseStatus)
	rows, err = r.db.QueryContext(ctx, topExpensesQuery, startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve top expenses: %w", err)
	}
	for rows.Next() {
		var expense ExpenseItem
		var expenseDate reportDate
		if err := rows.Scan(&expenseDate, &expense.Amount, &expense.Description, &expense.Status); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan top expense: %w", err)
		}
		expense.Date = expenseDate.Time
		report.TopExpenses = append(report.TopExpenses, expense)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read top expenses: %w", err)
	}
	rows.Close()
	return &report, nil
}

// GetProfitsData retrieves profits data for report
func (r *Repository) GetProfitsData(ctx context.Context, startDate, endDate time.Time) (*ProfitsReport, error) {
	var report ProfitsReport
	report.StartDate = startDate
	report.EndDate = endDate

	startDateKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize profit report start date: %w", err)
	}
	endDateKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize profit report end date: %w", err)
	}

	salesDateExpr := r.salesDateExpression("")
	if dbutil.IsSQLite(r.db) {
		err = r.db.GetContext(ctx, &report.TotalRevenue,
			fmt.Sprintf(`SELECT COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) FROM sales
			 WHERE %s >= date(?) AND %s < date(?)
			   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, salesDateExpr, salesDateExpr),
			startDateKey, endDateKey)
	} else {
		err = r.db.GetContext(ctx, &report.TotalRevenue,
			fmt.Sprintf(`SELECT COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) FROM sales
			 WHERE %s >= date($1) AND %s < date($2)
			   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, salesDateExpr, salesDateExpr),
			startDateKey, endDateKey)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve profit revenue: %w", err)
	}

	if dbutil.IsSQLite(r.db) {
		err = r.db.GetContext(ctx, &report.TotalCOGS, r.historicalCOGSTotalSQL("?", "?"), startDateKey, endDateKey)
	} else {
		err = r.db.GetContext(ctx, &report.TotalCOGS, r.historicalCOGSTotalSQL("$1", "$2"), startDateKey, endDateKey)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve profit cost: %w", err)
	}
	returnAdjustments, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "total")
	if err != nil {
		return nil, err
	}
	periodReturns := returnAdjustments["total"]
	report.TotalRevenue -= periodReturns.RevenueReduction
	report.TotalCOGS -= periodReturns.ReturnedCost

	report.GrossProfit = report.TotalRevenue - report.TotalCOGS

	report.TotalExpenses, err = accounting.AccountingExpensesForPeriod(ctx, r.db, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve profit expenses: %w", err)
	}

	report.NetProfit = report.GrossProfit - report.TotalExpenses
	if report.TotalRevenue > 0 {
		report.ProfitMargin = (report.NetProfit / report.TotalRevenue) * 100
	} else {
		report.ProfitMargin = 0
	}
	report.ByMonth = []MonthlyProfit{}
	monthlyProfitsQuery := `WITH sales_by_month AS (
			        SELECT DATE_TRUNC('month', COALESCE(s.sale_date, s.created_at)) AS month,
		               COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0) AS revenue
		        FROM sales s
		        WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		          AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		        GROUP BY DATE_TRUNC('month', COALESCE(s.sale_date, s.created_at))
		),
		cogs_by_month AS (
		        SELECT month, COALESCE(SUM(sale_cost), 0) AS cogs
		        FROM (
		                SELECT DATE_TRUNC('month', COALESCE(s.sale_date, s.created_at)) AS month, s.id,
		                       CASE WHEN s.cost_amount IS NOT NULL THEN s.cost_amount
		                            ELSE COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0)
		                       END AS sale_cost
		                FROM sales s
		                LEFT JOIN sale_items si ON si.sale_id = s.id
		                WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		                  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		                GROUP BY DATE_TRUNC('month', COALESCE(s.sale_date, s.created_at)), s.id, s.cost_amount
		        ) costs
		        GROUP BY month
		),
		expenses_by_month AS (
		        SELECT DATE_TRUNC('month', expense_date) AS month,
		               COALESCE(SUM(amount), 0) AS expenses
		        FROM expenses
			        WHERE date(expense_date) >= date(?) AND date(expense_date) < date(?)
		          AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')
		        GROUP BY DATE_TRUNC('month', expense_date)
		)
		SELECT s.month, s.revenue, COALESCE(c.cogs, 0), COALESCE(e.expenses, 0)
		FROM sales_by_month s
		LEFT JOIN cogs_by_month c ON c.month = s.month
		LEFT JOIN expenses_by_month e ON e.month = s.month
		ORDER BY s.month`
	if dbutil.IsSQLite(r.db) {
		monthlyProfitsQuery = `WITH sales_by_month AS (
		        SELECT strftime('%Y-%m', substr(COALESCE(s.sale_date, s.created_at), 1, 10)) AS month,
		               COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0) AS revenue
		        FROM sales s
		        WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		          AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		        GROUP BY strftime('%Y-%m', substr(COALESCE(s.sale_date, s.created_at), 1, 10))
		),
		cogs_by_month AS (
		        SELECT month, COALESCE(SUM(sale_cost), 0) AS cogs
		        FROM (
		                SELECT strftime('%Y-%m', substr(COALESCE(s.sale_date, s.created_at), 1, 10)) AS month, s.id,
		                       CASE WHEN s.cost_amount IS NOT NULL THEN s.cost_amount
		                            ELSE COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0)
		                       END AS sale_cost
		                FROM sales s
		                LEFT JOIN sale_items si ON si.sale_id = s.id
		                WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		                  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		                GROUP BY strftime('%Y-%m', substr(COALESCE(s.sale_date, s.created_at), 1, 10)), s.id, s.cost_amount
		        ) costs
		        GROUP BY month
		),
		expenses_by_month AS (
		        SELECT strftime('%Y-%m', expense_date) AS month,
		               COALESCE(SUM(amount), 0) AS expenses
		        FROM expenses
			        WHERE date(expense_date) >= date(?) AND date(expense_date) < date(?)
		          AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')
		        GROUP BY strftime('%Y-%m', expense_date)
		)
		SELECT s.month || '-01', s.revenue, COALESCE(c.cogs, 0), COALESCE(e.expenses, 0)
		FROM sales_by_month s
		LEFT JOIN cogs_by_month c ON c.month = s.month
		LEFT JOIN expenses_by_month e ON e.month = s.month
		ORDER BY s.month`
		monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery,
			"date(expense_date) >= date(?) AND date(expense_date) < date(?)",
			"date(substr(expense_date, 1, 10)) >= date(?) AND date(substr(expense_date, 1, 10)) < date(?)")
		monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "strftime('%Y-%m', expense_date)", "strftime('%Y-%m', substr(expense_date, 1, 10))")
		if !reportsSQLiteHasColumns(r.db, "sales", "created_at") {
			monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "COALESCE(s.sale_date, s.created_at)", "s.sale_date")
		}
		if !reportsSQLiteHasColumns(r.db, "sales", "cost_amount") {
			monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, ", s.cost_amount", "")
		}
		monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "COALESCE(s.sale_date, s.created_at)", r.salesDateExpression("s"))
	} else {
		monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "COALESCE(s.sale_date, s.created_at)", r.salesDateExpression("s"))
	}
	monthlyReturnPeriods := ""
	monthlyProfitArgs := []any{startDateKey, endDateKey, startDateKey, endDateKey, startDateKey, endDateKey}
	returnReferenceFilter := `AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'`
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "accounting_returns", "reference_number") {
		returnReferenceFilter = ""
	}
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_returns", "status", "total_refund_amount", "refund_date", "return_date", "created_at") {
		if dbutil.IsSQLite(r.db) {
			monthlyReturnPeriods = ` UNION SELECT strftime('%Y-%m', substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10))
				FROM accounting_returns r WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
				` + returnReferenceFilter + `
				AND date(substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10)) >= date(?)
				AND date(substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10)) < date(?)`
		} else {
			monthlyReturnPeriods = ` UNION SELECT DATE_TRUNC('month', COALESCE(r.refund_date, r.return_date, r.created_at))
				FROM accounting_returns r WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
				` + returnReferenceFilter + `
				AND date(COALESCE(r.refund_date, r.return_date, r.created_at)) >= date(?)
				AND date(COALESCE(r.refund_date, r.return_date, r.created_at)) < date(?)`
		}
		if !dbutil.IsSQLite(r.db) {
			monthlyReturnPeriods = strings.ReplaceAll(monthlyReturnPeriods, "COALESCE(r.refund_date, r.return_date, r.created_at)", r.returnDateExpression("r"))
		} else {
			monthlyReturnPeriods = strings.ReplaceAll(monthlyReturnPeriods, "COALESCE(r.refund_date, r.return_date, r.created_at)", r.returnDateExpression("r"))
		}
		monthlyProfitArgs = append(monthlyProfitArgs, startDateKey, endDateKey)
	}
	monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "SELECT s.month, s.revenue", "SELECT months.month, COALESCE(s.revenue, 0)")
	monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "SELECT s.month || '-01', s.revenue", "SELECT months.month || '-01', COALESCE(s.revenue, 0)")
	monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "FROM sales_by_month s\n", "FROM (SELECT month FROM sales_by_month UNION SELECT month FROM expenses_by_month"+monthlyReturnPeriods+") months\n\t\tLEFT JOIN sales_by_month s ON s.month = months.month\n")
	monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "LEFT JOIN cogs_by_month c ON c.month = s.month", "LEFT JOIN cogs_by_month c ON c.month = months.month")
	monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "LEFT JOIN expenses_by_month e ON e.month = s.month", "LEFT JOIN expenses_by_month e ON e.month = months.month")
	monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "ORDER BY s.month", "ORDER BY months.month")
	lineCostExpr, lineCostJoins := r.historicalSaleItemCostParts()
	monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "CASE WHEN s.cost_amount IS NOT NULL THEN s.cost_amount", "CASE WHEN NULLIF(s.cost_amount, 0) IS NOT NULL THEN s.cost_amount")
	monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0)", fmt.Sprintf("COALESCE(SUM(si.quantity * %s), 0)", lineCostExpr))
	monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, "LEFT JOIN sale_items si ON si.sale_id = s.id", "LEFT JOIN sale_items si ON si.sale_id = s.id"+lineCostJoins)
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "sales", "cost_amount") {
		monthlyProfitsQuery = profitTrendWithoutStoredSaleCost(monthlyProfitsQuery, lineCostExpr)
		monthlyProfitsQuery = strings.ReplaceAll(monthlyProfitsQuery, ", s.cost_amount", "")
	}
	rows, err := r.db.QueryContext(ctx, r.db.Rebind(monthlyProfitsQuery), monthlyProfitArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve monthly profit trend: %w", err)
	}
	for rows.Next() {
		var month MonthlyProfit
		var cogs, expenses float64
		var monthDate reportDate
		if err := rows.Scan(&monthDate, &month.Revenue, &cogs, &expenses); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan monthly profit trend: %w", err)
		}
		month.Month = monthDate.Time
		month.COGS = cogs
		month.GrossProfit = month.Revenue - cogs
		month.Expenses = expenses
		month.NetProfit = month.GrossProfit - expenses
		report.ByMonth = append(report.ByMonth, month)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read monthly profit trend: %w", err)
	}
	rows.Close()
	monthlyAdjustments, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "month")
	if err != nil {
		return nil, err
	}
	for index := range report.ByMonth {
		adjustment := monthlyAdjustments[report.ByMonth[index].Month.Format("2006-01")]
		report.ByMonth[index].Revenue -= adjustment.RevenueReduction
		report.ByMonth[index].COGS -= adjustment.ReturnedCost
		report.ByMonth[index].GrossProfit = report.ByMonth[index].Revenue - report.ByMonth[index].COGS
		report.ByMonth[index].NetProfit = report.ByMonth[index].GrossProfit - report.ByMonth[index].Expenses
	}
	monthlyNetProfit := 0.0
	for _, month := range report.ByMonth {
		monthlyNetProfit += month.NetProfit
	}
	if len(report.ByMonth) > 0 && (monthlyNetProfit-report.NetProfit > 0.01 || report.NetProfit-monthlyNetProfit > 0.01) {
		return nil, fmt.Errorf("monthly profit trend does not reconcile with period total: months %.2f, total %.2f", monthlyNetProfit, report.NetProfit)
	}

	// Build the daily series from the same revenue, historical COGS, and
	// approved-expense sources used by the period totals.
	report.ByDay = []DailyProfit{}
	dailyProfitsQuery := `WITH sales_by_day AS (
		SELECT DATE(COALESCE(s.sale_date, s.created_at)) AS day,
		       COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0) AS revenue
		FROM sales s
		WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
		  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		GROUP BY DATE(COALESCE(s.sale_date, s.created_at))
	), cogs_by_day AS (
		SELECT day, COALESCE(SUM(sale_cost), 0) AS cogs
		FROM (
			SELECT DATE(COALESCE(s.sale_date, s.created_at)) AS day, s.id,
			       CASE WHEN s.cost_amount IS NOT NULL THEN s.cost_amount ELSE COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0) END AS sale_cost
			FROM sales s LEFT JOIN sale_items si ON si.sale_id = s.id
			WHERE date(COALESCE(s.sale_date, s.created_at)) >= date(?) AND date(COALESCE(s.sale_date, s.created_at)) < date(?)
			  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			GROUP BY DATE(COALESCE(s.sale_date, s.created_at)), s.id, s.cost_amount
		) costs
		GROUP BY day
	), expenses_by_day AS (
		SELECT DATE(expense_date) AS day, COALESCE(SUM(amount), 0) AS expenses
		FROM expenses
		WHERE date(expense_date) >= date(?) AND date(expense_date) < date(?)
		  AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')
		GROUP BY DATE(expense_date)
	)
	SELECT s.day, s.revenue, COALESCE(c.cogs, 0), COALESCE(e.expenses, 0)
	FROM sales_by_day s
	LEFT JOIN cogs_by_day c ON c.day = s.day
	LEFT JOIN expenses_by_day e ON e.day = s.day
	ORDER BY s.day`
	if dbutil.IsSQLite(r.db) {
		dailyProfitsQuery = `WITH sales_by_day AS (
			SELECT strftime('%Y-%m-%d', substr(COALESCE(s.sale_date, s.created_at), 1, 10)) AS day,
			       COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0) AS revenue
			FROM sales s
			WHERE date(substr(COALESCE(s.sale_date, s.created_at), 1, 10)) >= date(?) AND date(substr(COALESCE(s.sale_date, s.created_at), 1, 10)) < date(?)
			  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			GROUP BY strftime('%Y-%m-%d', substr(COALESCE(s.sale_date, s.created_at), 1, 10))
		), cogs_by_day AS (
			SELECT day, COALESCE(SUM(sale_cost), 0) AS cogs
			FROM (
				SELECT strftime('%Y-%m-%d', substr(COALESCE(s.sale_date, s.created_at), 1, 10)) AS day, s.id,
				       CASE WHEN s.cost_amount IS NOT NULL THEN s.cost_amount ELSE COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0) END AS sale_cost
				FROM sales s LEFT JOIN sale_items si ON si.sale_id = s.id
				WHERE date(substr(COALESCE(s.sale_date, s.created_at), 1, 10)) >= date(?) AND date(substr(COALESCE(s.sale_date, s.created_at), 1, 10)) < date(?)
				  AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
				GROUP BY strftime('%Y-%m-%d', substr(COALESCE(s.sale_date, s.created_at), 1, 10)), s.id, s.cost_amount
			) costs
			GROUP BY day
		), expenses_by_day AS (
			SELECT strftime('%Y-%m-%d', substr(expense_date, 1, 10)) AS day, COALESCE(SUM(amount), 0) AS expenses
			FROM expenses
			WHERE date(substr(expense_date, 1, 10)) >= date(?) AND date(substr(expense_date, 1, 10)) < date(?)
			  AND LOWER(COALESCE(status, 'approved')) IN ('approved', 'paid', 'completed', 'archived')
			GROUP BY strftime('%Y-%m-%d', substr(expense_date, 1, 10))
		)
		SELECT s.day, s.revenue, COALESCE(c.cogs, 0), COALESCE(e.expenses, 0)
		FROM sales_by_day s
		LEFT JOIN cogs_by_day c ON c.day = s.day
		LEFT JOIN expenses_by_day e ON e.day = s.day
		ORDER BY s.day`
		if !reportsSQLiteHasColumns(r.db, "sales", "created_at") {
			dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "COALESCE(s.sale_date, s.created_at)", "s.sale_date")
		}
		if !reportsSQLiteHasColumns(r.db, "sales", "cost_amount") {
			dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, ", s.cost_amount", "")
		}
		dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "COALESCE(s.sale_date, s.created_at)", r.salesDateExpression("s"))
	} else {
		dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "COALESCE(s.sale_date, s.created_at)", r.salesDateExpression("s"))
	}
	dailyReturnPeriods := ""
	dailyProfitArgs := []any{}
	dailyProfitArgs = []any{startDateKey, endDateKey, startDateKey, endDateKey, startDateKey, endDateKey}
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_returns", "status", "total_refund_amount", "refund_date", "return_date", "created_at") {
		if dbutil.IsSQLite(r.db) {
			dailyReturnPeriods = ` UNION SELECT date(substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10))
				FROM accounting_returns r WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
				` + returnReferenceFilter + `
				AND date(substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10)) >= date(?)
				AND date(substr(COALESCE(r.refund_date, r.return_date, r.created_at), 1, 10)) < date(?)`
		} else {
			dailyReturnPeriods = ` UNION SELECT date(COALESCE(r.refund_date, r.return_date, r.created_at))
				FROM accounting_returns r WHERE UPPER(COALESCE(r.status, '')) = 'COMPLETED'
				` + returnReferenceFilter + `
				AND date(COALESCE(r.refund_date, r.return_date, r.created_at)) >= date(?)
				AND date(COALESCE(r.refund_date, r.return_date, r.created_at)) < date(?)`
		}
		if !dbutil.IsSQLite(r.db) {
			dailyReturnPeriods = strings.ReplaceAll(dailyReturnPeriods, "COALESCE(r.refund_date, r.return_date, r.created_at)", r.returnDateExpression("r"))
		} else {
			dailyReturnPeriods = strings.ReplaceAll(dailyReturnPeriods, "COALESCE(r.refund_date, r.return_date, r.created_at)", r.returnDateExpression("r"))
		}
		dailyProfitArgs = append(dailyProfitArgs, startDateKey, endDateKey)
	}
	dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "SELECT s.day, s.revenue", "SELECT days.day, COALESCE(s.revenue, 0)")
	dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "FROM sales_by_day s\n", "FROM (SELECT day FROM sales_by_day UNION SELECT day FROM expenses_by_day"+dailyReturnPeriods+") days\n\tLEFT JOIN sales_by_day s ON s.day = days.day\n")
	dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "LEFT JOIN cogs_by_day c ON c.day = s.day", "LEFT JOIN cogs_by_day c ON c.day = days.day")
	dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "LEFT JOIN expenses_by_day e ON e.day = s.day", "LEFT JOIN expenses_by_day e ON e.day = days.day")
	dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "ORDER BY s.day", "ORDER BY days.day")
	dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "CASE WHEN s.cost_amount IS NOT NULL THEN s.cost_amount", "CASE WHEN NULLIF(s.cost_amount, 0) IS NOT NULL THEN s.cost_amount")
	dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "COALESCE(SUM(si.quantity * COALESCE(si.unit_cost, 0)), 0)", fmt.Sprintf("COALESCE(SUM(si.quantity * %s), 0)", lineCostExpr))
	dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, "LEFT JOIN sale_items si ON si.sale_id = s.id", "LEFT JOIN sale_items si ON si.sale_id = s.id"+lineCostJoins)
	if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "sales", "cost_amount") {
		dailyProfitsQuery = profitTrendWithoutStoredSaleCost(dailyProfitsQuery, lineCostExpr)
		dailyProfitsQuery = strings.ReplaceAll(dailyProfitsQuery, ", s.cost_amount", "")
	}
	dailyRows, dailyErr := r.db.QueryContext(ctx, r.db.Rebind(dailyProfitsQuery), dailyProfitArgs...)
	if dailyErr != nil {
		return nil, fmt.Errorf("failed to retrieve daily profit trend: %w", dailyErr)
	}
	for dailyRows.Next() {
		var day DailyProfit
		var cogs, expenses float64
		var dayDate reportDate
		if scanErr := dailyRows.Scan(&dayDate, &day.Revenue, &cogs, &expenses); scanErr != nil {
			dailyRows.Close()
			return nil, fmt.Errorf("failed to scan daily profit trend: %w", scanErr)
		}
		day.Date = dayDate.Time
		day.COGS = cogs
		day.Expenses = expenses
		day.NetProfit = day.Revenue - cogs - expenses
		report.ByDay = append(report.ByDay, day)
	}
	if rowsErr := dailyRows.Err(); rowsErr != nil {
		dailyRows.Close()
		return nil, fmt.Errorf("failed to read daily profit trend: %w", rowsErr)
	}
	dailyRows.Close()
	dailyAdjustments, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "day")
	if err != nil {
		return nil, err
	}
	for index := range report.ByDay {
		adjustment := dailyAdjustments[report.ByDay[index].Date.Format("2006-01-02")]
		report.ByDay[index].Revenue -= adjustment.RevenueReduction
		report.ByDay[index].COGS -= adjustment.ReturnedCost
		report.ByDay[index].NetProfit = report.ByDay[index].Revenue - report.ByDay[index].COGS - report.ByDay[index].Expenses
	}
	dailyNetProfit := 0.0
	for _, day := range report.ByDay {
		dailyNetProfit += day.NetProfit
	}
	if len(report.ByDay) > 0 && (dailyNetProfit-report.NetProfit > 0.01 || report.NetProfit-dailyNetProfit > 0.01) {
		return nil, fmt.Errorf("daily profit trend does not reconcile with period total: days %.2f, total %.2f", dailyNetProfit, report.NetProfit)
	}
	report.ByCategory = make(map[string]float64)
	categoryRevenue := make(map[string]float64)
	categoryStartDate, categoryEndDate := any(startDateKey), any(endDateKey)
	categorySchemaAvailable := true
	if dbutil.IsSQLite(r.db) {
		productsAvailable, relationErr := reportsSQLiteHasRelation(ctx, r.db, "products")
		if relationErr != nil {
			return nil, fmt.Errorf("check profit category schema: %w", relationErr)
		}
		categoriesAvailable, relationErr := reportsSQLiteHasRelation(ctx, r.db, "categories")
		if relationErr != nil {
			return nil, fmt.Errorf("check profit category schema: %w", relationErr)
		}
		categorySchemaAvailable = productsAvailable && categoriesAvailable &&
			reportsSQLiteHasColumns(r.db, "products", "id", "name", "category_id") &&
			reportsSQLiteHasColumns(r.db, "categories", "id", "name")
	}
	if categorySchemaAvailable {
		rows, err = r.db.QueryContext(ctx,
			r.db.Rebind(`SELECT COALESCE(NULLIF(TRIM(c.name), ''), p.name, 'غير مصنف'),
		        COALESCE(SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)), 0),
		        COALESCE(SUM((COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) - (si.quantity * COALESCE(si.unit_cost, 0))), 0)
		 FROM sale_items si
		 JOIN sales s ON s.id = si.sale_id
		 JOIN products p ON p.id = si.product_id
		 LEFT JOIN categories c ON c.id = p.category_id
		 WHERE `+r.salesDateExpression("s")+` >= date(?) AND `+r.salesDateExpression("s")+` < date(?)
		   AND LOWER(COALESCE(s.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY COALESCE(NULLIF(TRIM(c.name), ''), p.name, 'غير مصنف')
		 ORDER BY SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)) DESC`,
			), categoryStartDate, categoryEndDate)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve profit by category: %w", err)
		}
		for rows.Next() {
			var category string
			var revenue, profit float64
			if err := rows.Scan(&category, &revenue, &profit); err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to scan profit category: %w", err)
			}
			categoryRevenue[category] += revenue
			report.ByCategory[category] = profit
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to read profit categories: %w", err)
		}
		rows.Close()
	}
	categoryReturns, err := r.returnCategoryAdjustments(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	for category, adjustment := range categoryReturns {
		report.ByCategory[category] -= adjustment.RevenueReduction
		report.ByCategory[category] += adjustment.ReturnedCost
	}
	// The report's total COGS uses each sale's stored historical cost, and its
	// margin includes completed returns. Allocate the difference from the
	// line-level category estimate by category revenue so category profits add
	// back to the reported gross profit.
	categoryProfitTotal, categoryRevenueTotal := 0.0, 0.0
	for category, profit := range report.ByCategory {
		categoryProfitTotal += profit
		categoryRevenueTotal += categoryRevenue[category]
	}
	if len(report.ByCategory) == 0 && report.GrossProfit != 0 {
		report.ByCategory["غير مصنف"] = report.GrossProfit
	} else if len(report.ByCategory) > 0 {
		difference := report.GrossProfit - categoryProfitTotal
		if categoryRevenueTotal > 0 {
			for category := range report.ByCategory {
				report.ByCategory[category] += difference * categoryRevenue[category] / categoryRevenueTotal
			}
		} else {
			for category := range report.ByCategory {
				report.ByCategory[category] += difference / float64(len(report.ByCategory))
			}
		}
	}

	return &report, nil
}

// GetDebtsData retrieves debts data for report
func (r *Repository) GetDebtsData(ctx context.Context) (*DebtsReport, error) {
	var report DebtsReport
	report.ByCustomer = []CustomerDebt{}
	report.ByAge = make(map[string]int)
	report.PaymentHistory = []PaymentRecord{}
	storeDate, err := accounting.StoreDate(time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to calculate debt report date: %w", err)
	}
	validDebt := `LOWER(COALESCE(status, 'pending')) NOT IN ('cancelled', 'canceled', 'reversed', 'deleted')`
	openDebt := business.OpenDebtStatusSQL("LOWER(COALESCE(status, 'pending'))")
	customerOpenDebt := business.OpenDebtStatusSQL("LOWER(COALESCE(d.status, 'pending'))")
	if err := r.db.GetContext(ctx, &report.TotalDebt,
		`SELECT COALESCE(SUM(amount), 0) FROM debts WHERE `+validDebt); err != nil {
		if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "debts", "amount", "remaining_amount") {
			return &report, nil
		}
		return nil, fmt.Errorf("failed to retrieve total debt: %w", err)
	}
	if err := r.db.GetContext(ctx, &report.TotalPaid,
		`SELECT COALESCE(SUM(CASE WHEN amount > remaining_amount THEN amount - remaining_amount ELSE 0 END), 0) FROM debts WHERE `+validDebt); err != nil {
		return nil, fmt.Errorf("failed to retrieve paid debt: %w", err)
	}
	if err := r.db.GetContext(ctx, &report.Outstanding,
		`SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE `+openDebt+` AND remaining_amount > 0`); err != nil {
		return nil, fmt.Errorf("failed to retrieve outstanding debt: %w", err)
	}
	overdueQuery := `SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE ` + openDebt + ` AND date(due_date) < date(?) AND remaining_amount > 0`
	overdueArgs := []any{storeDate}
	if !dbutil.IsSQLite(r.db) {
		overdueQuery = `SELECT COALESCE(SUM(remaining_amount), 0) FROM debts WHERE ` + openDebt + ` AND due_date::date < $1::date AND remaining_amount > 0`
	}
	if err := r.db.GetContext(ctx, &report.OverdueDebt, overdueQuery, overdueArgs...); err != nil {
		return nil, fmt.Errorf("failed to retrieve overdue debt: %w", err)
	}

	paymentTableExists := !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "payments", "customer_id")
	paymentStatusColumn := "status"
	if dbutil.IsSQLite(r.db) {
		if reportsSQLiteHasColumns(r.db, "payments", "payment_status") {
			paymentStatusColumn = "payment_status"
		} else if !reportsSQLiteHasColumns(r.db, "payments", "status") {
			paymentStatusColumn = ""
		}
	}
	paymentValidityFilter := ""
	if paymentStatusColumn != "" {
		paymentValidityFilter += fmt.Sprintf(" AND LOWER(COALESCE(p.%s, 'completed')) IN ('completed', 'paid')", paymentStatusColumn)
	}
	if dbutil.IsSQLite(r.db) {
		if reportsSQLiteHasColumns(r.db, "payments", "is_reversed") {
			paymentValidityFilter += " AND COALESCE(p.is_reversed, 0) = 0"
		}
	} else {
		paymentValidityFilter += " AND COALESCE(p.is_reversed, FALSE) = FALSE"
	}
	paymentValidityFilterWithoutAlias := strings.ReplaceAll(paymentValidityFilter, "p.", "")
	lastPaymentExpr := "'0001-01-01'"
	lastPaymentJoin := ""
	if paymentTableExists {
		paymentDateColumn := "payment_date"
		if dbutil.IsSQLite(r.db) {
			if reportsSQLiteHasColumns(r.db, "payments", "payment_date") {
				paymentDateColumn = "payment_date"
			} else if reportsSQLiteHasColumns(r.db, "payments", "created_at") {
				paymentDateColumn = "created_at"
			} else {
				paymentDateColumn = ""
			}
		}
		if paymentDateColumn != "" {
			lastPaymentExpr = "COALESCE(MAX(last_payment), '0001-01-01')"
			lastPaymentJoin = fmt.Sprintf(`LEFT JOIN (
				SELECT customer_id, MAX(%s) AS last_payment FROM payments
				WHERE customer_id IS NOT NULL%s GROUP BY customer_id
			) p ON p.customer_id = d.customer_id`, paymentDateColumn, paymentValidityFilterWithoutAlias)
		}
	}
	customerQuery := `SELECT d.customer_id, c.name, COALESCE(SUM(d.amount), 0),
		COALESCE(SUM(CASE WHEN d.amount > d.remaining_amount THEN d.amount - d.remaining_amount ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN ` + customerOpenDebt + ` AND d.remaining_amount > 0 THEN d.remaining_amount ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN ` + customerOpenDebt + ` AND d.remaining_amount > 0 AND date(d.due_date) < date(?) THEN d.remaining_amount ELSE 0 END), 0),
		` + lastPaymentExpr + `
		FROM debts d JOIN customers c ON c.id = d.customer_id
		` + lastPaymentJoin + `
		WHERE ` + strings.ReplaceAll(validDebt, "status", "d.status") + `
		GROUP BY d.customer_id, c.name ORDER BY SUM(CASE WHEN ` + customerOpenDebt + ` THEN d.remaining_amount ELSE 0 END) DESC`
	if !dbutil.IsSQLite(r.db) {
		customerQuery = strings.Replace(customerQuery, "date(d.due_date) < date(?)", "d.due_date::date < $1::date", 1)
	}
	rows, err := r.db.QueryContext(ctx, customerQuery, storeDate)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve debt customers: %w", err)
	}
	for rows.Next() {
		var item CustomerDebt
		var lastPayment reportTimestamp
		if err := rows.Scan(&item.CustomerID, &item.CustomerName, &item.TotalDebt, &item.PaidAmount, &item.Outstanding, &item.OverdueAmount, &lastPayment); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan debt customer: %w", err)
		}
		item.LastPayment = lastPayment.Time
		report.ByCustomer = append(report.ByCustomer, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read debt customers: %w", err)
	}
	rows.Close()
	ageQuery := `SELECT CASE
			WHEN due_date IS NULL OR date(due_date) >= date(?) THEN 'current'
			WHEN julianday(?) - julianday(due_date) <= 7 THEN 'overdue_1_7'
			WHEN julianday(?) - julianday(due_date) <= 30 THEN 'overdue_8_30'
			ELSE 'overdue_30_plus' END, COUNT(*)
		 FROM debts WHERE ` + openDebt + ` AND remaining_amount > 0 GROUP BY 1`
	ageArgs := []any{storeDate, storeDate, storeDate}
	if dbutil.IsSQLite(r.db) {
		ageQuery = `SELECT CASE
			WHEN due_date IS NULL OR date(due_date) >= date(?) THEN 'current'
			WHEN julianday(?) - julianday(due_date) <= 7 THEN 'overdue_1_7'
			WHEN julianday(?) - julianday(due_date) <= 30 THEN 'overdue_8_30'
			ELSE 'overdue_30_plus' END, COUNT(*)
		 FROM debts WHERE ` + openDebt + ` AND remaining_amount > 0 GROUP BY 1`
	} else {
		ageQuery = `SELECT CASE
			WHEN due_date IS NULL OR due_date::date >= $1::date THEN 'current'
			WHEN $1::date - due_date::date <= 7 THEN 'overdue_1_7'
			WHEN $1::date - due_date::date <= 30 THEN 'overdue_8_30'
			ELSE 'overdue_30_plus' END, COUNT(*)
		 FROM debts WHERE ` + openDebt + ` AND remaining_amount > 0 GROUP BY 1`
		ageArgs = []any{storeDate}
	}
	rows, err = r.db.QueryContext(ctx, ageQuery, ageArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve debt aging: %w", err)
	}
	for rows.Next() {
		var age string
		var count int
		if err := rows.Scan(&age, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan debt aging: %w", err)
		}
		report.ByAge[age] = count
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read debt aging: %w", err)
	}
	rows.Close()
	if paymentTableExists {
		paymentDateColumn := "payment_date"
		paymentMethodColumn, paymentReferenceColumn, paymentNotesColumn := "payment_method", "reference_number", "notes"
		if dbutil.IsSQLite(r.db) {
			if !reportsSQLiteHasColumns(r.db, "payments", "payment_date") {
				paymentDateColumn = "created_at"
			}
			if !reportsSQLiteHasColumns(r.db, "payments", "payment_method") {
				paymentMethodColumn = "method"
			}
			if !reportsSQLiteHasColumns(r.db, "payments", "reference_number") {
				if reportsSQLiteHasColumns(r.db, "payments", "reference") {
					paymentReferenceColumn = "reference"
				} else if reportsSQLiteHasColumns(r.db, "payments", "transaction_number") {
					paymentReferenceColumn = "transaction_number"
				} else {
					paymentReferenceColumn = "''"
				}
			}
			if !reportsSQLiteHasColumns(r.db, "payments", "notes") {
				paymentNotesColumn = "''"
			}
		}
		if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "payments", paymentDateColumn) {
			return &report, nil
		}
		if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "payments", paymentMethodColumn) {
			paymentMethodColumn = "''"
		}
		paymentQuery := fmt.Sprintf(`SELECT p.%s, p.customer_id, COALESCE(c.name, 'عميل غير معروف'), p.amount,
			COALESCE(p.%s, ''), COALESCE(p.%s, ''), COALESCE(p.%s, '')
			FROM payments p LEFT JOIN customers c ON c.id = p.customer_id
			WHERE p.customer_id IS NOT NULL%s ORDER BY p.%s DESC LIMIT 100`, paymentDateColumn, paymentMethodColumn, paymentReferenceColumn, paymentNotesColumn, paymentValidityFilter, paymentDateColumn)
		rows, err = r.db.QueryContext(ctx, paymentQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve debt payment history: %w", err)
		}
		for rows.Next() {
			var payment PaymentRecord
			var paymentDate reportTimestamp
			if err := rows.Scan(&paymentDate, &payment.CustomerID, &payment.CustomerName, &payment.Amount, &payment.PaymentMethod, &payment.ReferenceNumber, &payment.Notes); err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to scan debt payment history: %w", err)
			}
			payment.Date = paymentDate.Time
			report.PaymentHistory = append(report.PaymentHistory, payment)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to read debt payment history: %w", err)
		}
		rows.Close()
	}
	return &report, nil
}

// GetPurchasesData retrieves purchases data for report
func (r *Repository) GetPurchasesData(ctx context.Context, startDate, endDate time.Time) (*PurchasesReport, error) {
	var report PurchasesReport
	report.StartDate = startDate
	report.EndDate = endDate
	startDateKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize purchases report start date: %w", err)
	}
	endDateKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize purchases report end date: %w", err)
	}
	usedAcquisitionSchemaAvailable := !dbutil.IsSQLite(r.db) ||
		reportsSQLiteHasColumns(r.db, "acquisitions", "id", "type", "status", "acquisition_date", "total_cost", "paid_amount") &&
			reportsSQLiteHasColumns(r.db, "acquisition_items", "acquisition_id", "inventory_item_id") &&
			reportsSQLiteHasColumns(r.db, "inventory_items", "id", "status")
	if usedAcquisitionSchemaAvailable {
		if err := r.db.GetContext(ctx, &report.UsedPartPurchases, `
		SELECT
		  (SELECT COUNT(*) FROM acquisitions a
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
			 AND date(a.acquisition_date) < date(substr($2, 1, 10))) AS acquisition_count,
		  (SELECT COUNT(*) FROM acquisition_items ai JOIN acquisitions a ON a.id = ai.acquisition_id
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
			 AND date(a.acquisition_date) < date(substr($2, 1, 10))) AS item_count,
		  (SELECT COALESCE(SUM(a.total_cost), 0) FROM acquisitions a
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
			 AND date(a.acquisition_date) < date(substr($2, 1, 10))) AS total_cost,
		  (SELECT COALESCE(SUM(a.paid_amount), 0) FROM acquisitions a
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
			 AND date(a.acquisition_date) < date(substr($2, 1, 10))) AS total_paid,
		  (SELECT COALESCE(SUM(a.total_cost - a.paid_amount), 0) FROM acquisitions a
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed')
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
			 AND date(a.acquisition_date) < date(substr($2, 1, 10))) AS outstanding,
		  (SELECT COUNT(*) FROM inventory_items ii JOIN acquisition_items ai ON ai.inventory_item_id = ii.id JOIN acquisitions a ON a.id = ai.acquisition_id
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed') AND ii.status = 'AVAILABLE'
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
			 AND date(a.acquisition_date) < date(substr($2, 1, 10))) AS available_items,
		  (SELECT COUNT(*) FROM inventory_items ii JOIN acquisition_items ai ON ai.inventory_item_id = ii.id JOIN acquisitions a ON a.id = ai.acquisition_id
		   WHERE a.type = 'CUSTOMER' AND a.status NOT IN ('cancelled', 'reversed') AND ii.status = 'SOLD'
		     AND date(a.acquisition_date) >= date(substr($1, 1, 10))
			 AND date(a.acquisition_date) < date(substr($2, 1, 10))) AS sold_items`, startDateKey, endDateKey); err != nil {
			return nil, fmt.Errorf("failed to retrieve used-part acquisitions: %w", err)
		}
	}
	itemTotalColumn := "total_amount"
	if dbutil.IsSQLite(r.db) {
		itemTotalColumn = "item_total"
		exists, err := reportsSQLiteHasRelation(ctx, r.db, "purchases")
		if err != nil {
			return nil, fmt.Errorf("failed to inspect purchases schema: %w", err)
		}
		if !exists {
			return &report, nil
		}
	}

	err = r.db.GetContext(ctx, &report.TotalPurchases,
		`SELECT COUNT(*) FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`,
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve purchase count: %w", err)
	}

	err = r.db.GetContext(ctx, &report.TotalCost,
		`SELECT COALESCE(SUM(total_amount), 0) FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`,
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve purchase total: %w", err)
	}
	returnSchemaAvailable := !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "supplier_returns", "refund_amount", "created_at", "status")
	if returnSchemaAvailable {
		supplierReturnDate := "store_date(created_at)"
		if !dbutil.IsSQLite(r.db) {
			supplierReturnDate = "date(" + postgresStoreTimestampDateExpression("created_at") + ")"
		}
		query := fmt.Sprintf(`SELECT COALESCE(SUM(refund_amount), 0) FROM supplier_returns
			 WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
			   AND UPPER(COALESCE(status, '')) = 'COMPLETED'`, supplierReturnDate, supplierReturnDate)
		if err = r.db.GetContext(ctx, &report.SupplierReturnCredits, query, startDateKey, endDateKey); err != nil {
			return nil, fmt.Errorf("failed to retrieve supplier return credits: %w", err)
		}
	}
	rawNetPurchases := report.TotalCost - report.SupplierReturnCredits
	report.NetPurchases = rawNetPurchases
	report.SupplierCreditBalance = maxFloat(-rawNetPurchases, 0)
	report.TaxAmount = 0
	if err := r.db.GetContext(ctx, &report.TaxAmount,
		fmt.Sprintf(`SELECT COALESCE(SUM(CASE
			WHEN COALESCE(tax_amount, 0) > 0 THEN tax_amount
			WHEN total_amount > COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount) + 0.01
				THEN total_amount - COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount)
			ELSE 0 END), 0) FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, itemTotalColumn, itemTotalColumn),
		startDateKey, endDateKey); err != nil {
		return nil, fmt.Errorf("failed to retrieve purchase tax amount: %w", err)
	}
	var taxedPurchases int
	if err := r.db.GetContext(ctx, &taxedPurchases,
		fmt.Sprintf(`SELECT COUNT(*) FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND CASE
				WHEN COALESCE(tax_amount, 0) > 0 THEN tax_amount
				WHEN total_amount > COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount) + 0.01
					THEN total_amount - COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount)
				ELSE 0 END > 0
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, itemTotalColumn, itemTotalColumn),
		startDateKey, endDateKey); err != nil {
		return nil, fmt.Errorf("failed to calculate taxed purchase count: %w", err)
	}
	var taxedPurchaseCost float64
	if err := r.db.GetContext(ctx, &taxedPurchaseCost,
		fmt.Sprintf(`SELECT COALESCE(SUM(total_amount), 0) FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND CASE
				WHEN COALESCE(tax_amount, 0) > 0 THEN tax_amount
				WHEN total_amount > COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount) + 0.01
					THEN total_amount - COALESCE((SELECT SUM(%s) FROM purchase_items WHERE purchase_id = purchases.id), total_amount)
				ELSE 0 END > 0
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')`, itemTotalColumn, itemTotalColumn),
		startDateKey, endDateKey); err != nil {
		return nil, fmt.Errorf("failed to calculate taxed purchase cost: %w", err)
	}
	report.UntaxedPurchases = report.TotalPurchases - taxedPurchases
	report.UntaxedPurchaseCost = report.TotalCost - taxedPurchaseCost
	report.Subtotal = report.TotalCost - report.TaxAmount

	report.BySupplier = []SupplierPurchases{}
	rows, err := r.db.QueryContext(ctx,
		`SELECT CAST(p.supplier_id AS TEXT), COALESCE(s.name, 'مورد غير معروف'),
		        COALESCE(SUM(p.total_amount), 0), COALESCE(SUM(pi.item_count), 0)
		 FROM purchases p
		 LEFT JOIN suppliers s ON s.id = p.supplier_id
		 LEFT JOIN (
		   SELECT purchase_id, SUM(quantity) AS item_count FROM purchase_items GROUP BY purchase_id
		 ) pi ON pi.purchase_id = p.id
		 WHERE date(p.purchase_date) >= date(substr($1, 1, 10)) AND date(p.purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(p.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY p.supplier_id, s.name ORDER BY SUM(p.total_amount) DESC`,
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve purchases by supplier: %w", err)
	}
	for rows.Next() {
		var item SupplierPurchases
		var supplierID string
		if err := rows.Scan(&supplierID, &item.SupplierName, &item.TotalCost, &item.ItemCount); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan supplier purchases: %w", err)
		}
		if parsedID, parseErr := uuid.Parse(strings.TrimSpace(supplierID)); parseErr == nil {
			item.SupplierID = parsedID
		}
		report.BySupplier = append(report.BySupplier, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read supplier purchases: %w", err)
	}
	rows.Close()

	report.ByCategory = make(map[string]int)
	rows, err = r.db.QueryContext(ctx,
		`SELECT COALESCE(c.name, 'غير مصنف'), COALESCE(SUM(pi.quantity), 0)
		 FROM purchase_items pi
		 JOIN purchases p ON p.id = pi.purchase_id
		 JOIN products pr ON pr.id = pi.product_id
		 LEFT JOIN categories c ON c.id = pr.category_id
		 WHERE date(p.purchase_date) >= date(substr($1, 1, 10)) AND date(p.purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(p.status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY c.name ORDER BY SUM(pi.quantity) DESC`,
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve purchases by category: %w", err)
	}
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan purchases by category: %w", err)
		}
		report.ByCategory[category] = count
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read purchases by category: %w", err)
	}
	rows.Close()
	report.ByMonth = []MonthlyPurchases{}
	monthlyPurchasesQuery := `SELECT DATE_TRUNC('month', purchase_date), COALESCE(SUM(total_amount), 0), COUNT(*)
			 FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
			 GROUP BY DATE_TRUNC('month', purchase_date)
			 ORDER BY DATE_TRUNC('month', purchase_date)`
	if dbutil.IsSQLite(r.db) {
		monthlyPurchasesQuery = `SELECT strftime('%Y-%m-01', purchase_date), COALESCE(SUM(total_amount), 0), COUNT(*)
			 FROM purchases
		 WHERE date(purchase_date) >= date(substr($1, 1, 10)) AND date(purchase_date) < date(substr($2, 1, 10))
		   AND LOWER(COALESCE(status, 'completed')) NOT IN ('cancelled', 'canceled', 'reversed')
		 GROUP BY strftime('%Y-%m', purchase_date)
		 ORDER BY strftime('%Y-%m', purchase_date)`
	}
	rows, err = r.db.QueryContext(ctx, monthlyPurchasesQuery, startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve monthly purchases: %w", err)
	}
	for rows.Next() {
		var month MonthlyPurchases
		var monthDate reportDate
		if err := rows.Scan(&monthDate, &month.Cost, &month.Count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan monthly purchases: %w", err)
		}
		month.Month = monthDate.Time
		report.ByMonth = append(report.ByMonth, month)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read monthly purchases: %w", err)
	}
	rows.Close()

	return &report, nil
}

// GetReturnsData retrieves returns data for report
func (r *Repository) GetReturnsData(ctx context.Context, startDate, endDate time.Time) (*ReturnsReport, error) {
	var report ReturnsReport
	report.StartDate = startDate
	report.EndDate = endDate
	startDateKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize returns report start date: %w", err)
	}
	endDateKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize returns report end date: %w", err)
	}
	returnDate := r.returnDateExpression("r")

	// Only completed returns affect financial reports.
	err = r.db.GetContext(ctx, &report.TotalReturns,
		fmt.Sprintf(`SELECT COUNT(*) FROM accounting_returns r
		 WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
			   AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'
			   AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%%'`, returnDate, returnDate),
		startDateKey, endDateKey)
	if err != nil {
		if dbutil.IsSQLite(r.db) && !reportsSQLiteHasColumns(r.db, "accounting_returns", "id") {
			report.ByReason = make(map[string]int)
			return &report, nil
		}
		return nil, fmt.Errorf("failed to retrieve completed returns: %w", err)
	}

	err = r.db.GetContext(ctx, &report.TotalRefunded,
		fmt.Sprintf(`SELECT COALESCE(SUM(r.total_refund_amount), 0) FROM accounting_returns r
		 WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
			   AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'
			   AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%%'`, returnDate, returnDate),
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve completed return amounts: %w", err)
	}

	report.ByReason = make(map[string]int)
	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT COALESCE(r.reason, ''), COUNT(*) FROM accounting_returns r
		 WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
			   AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'
			   AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%%'
		 GROUP BY r.reason`, returnDate, returnDate),
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve return reasons: %w", err)
	}
	for rows.Next() {
		var reason string
		var count int
		if err := rows.Scan(&reason, &count); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan return reason: %w", err)
		}
		report.ByReason[reason] = count
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read return reasons: %w", err)
	}
	rows.Close()

	report.ByProduct = []ProductReturns{}
	rows, err = r.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT ri.product_id, COALESCE(p.name, 'منتج محذوف'),
			        COUNT(DISTINCT r.id), COALESCE(SUM(ri.total_refund_amount), 0)
			 FROM accounting_returns r
			 JOIN accounting_return_items ri ON ri.return_id = r.id
			 LEFT JOIN products p ON p.id = ri.product_id
			 WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
			   AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'
			   AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%%'
			 GROUP BY ri.product_id, p.name
			 ORDER BY SUM(ri.total_refund_amount) DESC`, returnDate, returnDate),
		startDateKey, endDateKey)
	if err != nil {
		if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_return_items", "return_id") {
			return nil, fmt.Errorf("failed to retrieve returned products: %w", err)
		}
	} else {
		for rows.Next() {
			var product ProductReturns
			if err := rows.Scan(&product.ProductID, &product.ProductName, &product.ReturnCount, &product.RefundAmount); err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to scan returned product: %w", err)
			}
			report.ByProduct = append(report.ByProduct, product)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to read returned products: %w", err)
		}
		rows.Close()
	}

	report.ByMonth = []MonthlyReturns{}
	monthBucket := fmt.Sprintf("DATE_TRUNC('month', %s)", returnDate)
	if dbutil.IsSQLite(r.db) {
		monthBucket = fmt.Sprintf("strftime('%%Y-%%m-01', %s)", returnDate)
	}
	monthlyReturnsQuery := fmt.Sprintf(`SELECT %s, COUNT(*), COALESCE(SUM(r.total_refund_amount), 0)
		 FROM accounting_returns r
		 WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
		   AND UPPER(COALESCE(r.status, '')) = 'COMPLETED'
		   AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%%'
		 GROUP BY %s ORDER BY %s`, monthBucket, returnDate, returnDate, monthBucket, monthBucket)
	rows, err = r.db.QueryContext(ctx, monthlyReturnsQuery, startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve returns by month: %w", err)
	}
	for rows.Next() {
		var monthly MonthlyReturns
		var month reportDate
		if err := rows.Scan(&month, &monthly.Count, &monthly.Amount); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan returns by month: %w", err)
		}
		monthly.Month = month.Time
		report.ByMonth = append(report.ByMonth, monthly)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read returns by month: %w", err)
	}
	rows.Close()

	return &report, nil
}

// GetNetSalesData retrieves net sales data for report (gross sales minus returns)
func (r *Repository) GetNetSalesData(ctx context.Context, startDate, endDate time.Time) (*NetSalesReport, error) {
	var report NetSalesReport
	report.StartDate = startDate
	report.EndDate = endDate
	startDateKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize net sales report start date: %w", err)
	}
	endDateKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize net sales report end date: %w", err)
	}

	// Get gross sales data
	var grossSales, grossRevenue float64
	salesDate := r.salesDateExpression("")
	err = r.db.GetContext(ctx, &grossSales,
		fmt.Sprintf(`SELECT COUNT(*) FROM sales WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
			AND LOWER(COALESCE(status, 'completed')) = 'completed'`,
			salesDate, salesDate),
		startDateKey, endDateKey)
	if err != nil {
		if dbutil.IsSQLite(r.db) {
			exists, schemaErr := reportsSQLiteHasRelation(ctx, r.db, "sales")
			if schemaErr != nil {
				return nil, fmt.Errorf("failed to inspect sales schema for net sales report: %w", schemaErr)
			}
			if !exists {
				return &report, nil
			}
		}
		return nil, fmt.Errorf("failed to retrieve gross sale count: %w", err)
	}
	report.GrossSales = int(grossSales)

	err = r.db.GetContext(ctx, &grossRevenue,
		fmt.Sprintf(`SELECT COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) FROM sales WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
			AND LOWER(COALESCE(status, 'completed')) = 'completed'`, salesDate, salesDate),
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve gross sales revenue: %w", err)
	}
	report.GrossRevenue = grossRevenue

	returnAdjustments, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "total")
	if err != nil {
		return nil, err
	}
	periodReturns := returnAdjustments["total"]
	report.TotalReturns = periodReturns.Count
	report.TotalRefunded = periodReturns.GrossRefund

	// Calculate net sales
	report.NetSales = report.GrossSales - report.TotalReturns
	report.NetRevenue = report.GrossRevenue - periodReturns.RevenueReduction

	// Calculate return rate
	if report.GrossSales > 0 {
		report.ReturnRate = (float64(report.TotalReturns) / float64(report.GrossSales)) * 100
	} else {
		report.ReturnRate = 0
	}

	// Daily series is an event ledger: sales use sale date, refunds use refund
	// date. A refund on a day without sales must still appear as a negative day.
	dailySalesDate := r.salesDateExpression("s")
	dailySalesQuery := fmt.Sprintf(`SELECT %s, COUNT(*), COALESCE(SUM(COALESCE(s.total_amount, 0) - COALESCE(s.tax_amount, 0)), 0)
		FROM sales s WHERE %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))
		AND LOWER(COALESCE(s.status, 'completed')) = 'completed' GROUP BY %s`, dailySalesDate, dailySalesDate, dailySalesDate, dailySalesDate)
	rows, err := r.db.QueryContext(ctx, dailySalesQuery, startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("retrieve daily net sales: %w", err)
	}
	dailyByDate := make(map[string]DailyNetSales)
	for rows.Next() {
		var daily DailyNetSales
		var date reportDate
		if err := rows.Scan(&date, &daily.GrossSales, &daily.GrossRevenue); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan daily net sales: %w", err)
		}
		daily.Date = date.Time
		dailyByDate[daily.Date.Format("2006-01-02")] = daily
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("read daily net sales: %w", err)
	}
	rows.Close()
	dailyReturns, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "day")
	if err != nil {
		return nil, err
	}
	for date, adjustment := range dailyReturns {
		daily := dailyByDate[date]
		if daily.Date.IsZero() {
			parsed, parseErr := time.Parse("2006-01-02", date)
			if parseErr != nil {
				continue
			}
			daily.Date = parsed
		}
		daily.Returns = adjustment.Count
		daily.Refunded = adjustment.GrossRefund
		dailyByDate[date] = daily
	}
	dailyKeys := make([]string, 0, len(dailyByDate))
	for date := range dailyByDate {
		dailyKeys = append(dailyKeys, date)
	}
	sort.Strings(dailyKeys)
	for _, date := range dailyKeys {
		daily := dailyByDate[date]
		daily.NetSales = daily.GrossSales - daily.Returns
		daily.NetRevenue = daily.GrossRevenue - dailyReturns[date].RevenueReduction
		report.ByDay = append(report.ByDay, daily)
	}

	// Get payment method breakdown for gross sales
	report.ByCategory = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx,
		`SELECT COALESCE(c.name, p.name, 'غير مصنف'), COALESCE(SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)), 0) AS total
		 FROM sale_items si
		 JOIN sales s ON s.id = si.sale_id
		 JOIN products p ON p.id = si.product_id
			 LEFT JOIN categories c ON c.id = p.category_id
				 WHERE `+r.salesDateExpression("s")+` >= date(substr($1, 1, 10)) AND `+r.salesDateExpression("s")+` < date(substr($2, 1, 10))
			   AND LOWER(COALESCE(s.status, 'completed')) = 'completed'
				 GROUP BY c.name, p.name
		 ORDER BY total DESC`,
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve net sales by category: %w", err)
	}
	for rows.Next() {
		var category string
		var total float64
		if err := rows.Scan(&category, &total); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan net sales category: %w", err)
		}
		report.ByCategory[category] += total
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read net sales categories: %w", err)
	}
	rows.Close()
	categoryReturns, err := r.returnCategoryAdjustments(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	for category, adjustment := range categoryReturns {
		report.ByCategory[category] -= adjustment.RevenueReduction
	}

	report.ByPaymentMethod = make(map[string]float64)
	rows, err = r.db.QueryContext(ctx,
		`SELECT COALESCE(NULLIF(TRIM(payment_method), ''), 'غير محدد'), COALESCE(SUM(COALESCE(total_amount, 0) - COALESCE(tax_amount, 0)), 0) as total
				 FROM sales
				 WHERE `+r.salesDateExpression("")+` >= date(substr($1, 1, 10)) AND `+r.salesDateExpression("")+` < date(substr($2, 1, 10))
			   AND LOWER(COALESCE(status, 'completed')) = 'completed'
			 GROUP BY COALESCE(NULLIF(TRIM(payment_method), ''), 'غير محدد')`,
		startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve net sales by payment method: %w", err)
	}
	for rows.Next() {
		var method string
		var total float64
		if err := rows.Scan(&method, &total); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan net sales payment method: %w", err)
		}
		report.ByPaymentMethod[method] = total
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read net sales by payment method: %w", err)
	}
	rows.Close()
	refundsByPaymentMethod, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "payment")
	if err != nil {
		return nil, err
	}
	for method, adjustment := range refundsByPaymentMethod {
		report.ByPaymentMethod[method] -= adjustment.RevenueReduction
	}

	// Get top returned products with net sales analysis
	report.TopReturnedProducts = []ProductNetSales{}
	returnDate := r.returnDateExpression("r")
	returnReferenceFilter := ""
	if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_returns", "reference_number") {
		returnReferenceFilter = `AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'`
	}
	returnTaxRatio := `CASE WHEN COALESCE(si2.total_amount, 0) > 0
		THEN MAX((COALESCE(si2.total_amount, 0) - COALESCE(si2.tax_amount, 0)) / si2.total_amount, 0)
		ELSE 1 END`
	if !dbutil.IsSQLite(r.db) {
		returnTaxRatio = `CASE WHEN COALESCE(si2.total_amount, 0) > 0
			THEN GREATEST((COALESCE(si2.total_amount, 0) - COALESCE(si2.tax_amount, 0)) / si2.total_amount, 0)
			ELSE 1 END`
	}
	returnFilter := fmt.Sprintf(`UPPER(COALESCE(r.status, '')) = 'COMPLETED' %s
		AND %s >= date(substr($1, 1, 10)) AND %s < date(substr($2, 1, 10))`, returnReferenceFilter, returnDate, returnDate)
	returnedQuantityByProduct := fmt.Sprintf(`SELECT SUM(COALESCE(ri.quantity_returned, 0))
		FROM accounting_return_items ri
		LEFT JOIN sale_items si2 ON si2.id = ri.sale_item_id
		LEFT JOIN inventory_items ii2 ON ii2.id = ri.inventory_item_id
		JOIN accounting_returns r ON r.id = ri.return_id
		WHERE COALESCE(ri.product_id, si2.product_id, ii2.product_id) = p.id AND %s`, returnFilter)
	returnedRevenueByProduct := fmt.Sprintf(`SELECT SUM(COALESCE(ri.total_refund_amount, 0) * (%s))
		FROM accounting_return_items ri
		LEFT JOIN sale_items si2 ON si2.id = ri.sale_item_id
		LEFT JOIN inventory_items ii2 ON ii2.id = ri.inventory_item_id
		JOIN accounting_returns r ON r.id = ri.return_id
		WHERE COALESCE(ri.product_id, si2.product_id, ii2.product_id) = p.id AND %s`, returnTaxRatio, returnFilter)
	topReturnedProductsQuery := fmt.Sprintf(`SELECT
			p.id AS product_id, p.name AS product_name,
			COALESCE(SUM(si.quantity), 0) AS gross_quantity,
			COALESCE(SUM(COALESCE(si.total_amount, 0) - COALESCE(si.tax_amount, 0)), 0) AS gross_revenue,
			COALESCE((%s), 0) AS returned_quantity,
			COALESCE((%s), 0) AS refunded_amount
		 FROM products p
		 LEFT JOIN sale_items si ON p.id = si.product_id AND si.sale_id IN (
			SELECT id FROM sales WHERE %s >= date(substr($1, 1, 10))
			AND %s < date(substr($2, 1, 10))
			AND LOWER(COALESCE(status, 'completed')) = 'completed')
		 WHERE p.is_active = true
		 GROUP BY p.id, p.name
		 HAVING COALESCE(SUM(si.quantity), 0) > 0 OR COALESCE((%s), 0) > 0
		 ORDER BY returned_quantity DESC
		 LIMIT 10`, returnedQuantityByProduct, returnedRevenueByProduct, r.salesDateExpression(""), r.salesDateExpression(""), returnedQuantityByProduct)
	rows, err = r.db.QueryContext(ctx, topReturnedProductsQuery, startDateKey, endDateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve top returned products: %w", err)
	}
	for rows.Next() {
		var product ProductNetSales
		if err := rows.Scan(&product.ProductID, &product.ProductName, &product.GrossQuantity, &product.GrossRevenue, &product.ReturnedQuantity, &product.RefundedAmount); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan top returned product: %w", err)
		}
		product.NetQuantity = product.GrossQuantity - product.ReturnedQuantity
		product.NetRevenue = product.GrossRevenue - product.RefundedAmount
		if product.GrossQuantity > 0 {
			product.ReturnRate = (float64(product.ReturnedQuantity) / float64(product.GrossQuantity)) * 100
		} else {
			product.ReturnRate = 0
		}
		report.TopReturnedProducts = append(report.TopReturnedProducts, product)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to read top returned products: %w", err)
	}
	rows.Close()

	return &report, nil
}

func (r *Repository) GetTaxData(ctx context.Context, startDate, endDate time.Time) (*TaxReport, error) {
	report := &TaxReport{StartDate: startDate, EndDate: endDate, Period: startDate.Format("2006-01-02") + " to " + endDate.Format("2006-01-02")}
	startDateKey, err := accounting.StoreDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("normalize tax report start date: %w", err)
	}
	endDateKey, err := accounting.StoreDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("normalize tax report end date: %w", err)
	}
	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(CASE WHEN tax_amount > 0 THEN subtotal ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tax_amount > 0 THEN discount_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tax_amount > 0 THEN subtotal - discount_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tax_amount > 0 THEN tax_amount ELSE 0 END), 0),
			COALESCE(SUM(total_amount), 0)
		FROM sales
		WHERE %s >= date(substr($1, 1, 10))
		  AND %s < date(substr($2, 1, 10))
		  AND LOWER(COALESCE(status, 'completed')) = 'completed'`, r.salesDateExpression(""), r.salesDateExpression(""))
	if err := r.db.QueryRowxContext(ctx, query, startDateKey, endDateKey).Scan(&report.GrossSales, &report.Discounts, &report.TaxableSales, &report.TaxCollected, &report.SalesTotal); err != nil {
		return nil, err
	}
	report.ExemptSales = report.SalesTotal - report.TaxableSales - report.TaxCollected
	if report.ExemptSales < 0 && report.ExemptSales > -0.01 {
		report.ExemptSales = 0
	}
	periodReturns, err := r.returnFinancialAdjustments(ctx, startDate, endDate, "total")
	if err != nil {
		return nil, err
	}
	returnTotals := periodReturns["total"]
	report.ReturnsTotal = returnTotals.GrossRefund
	report.ReturnedTax = returnTotals.TaxReduction
	canCalculateReturnedTax := !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_return_items", "return_id", "sale_item_id", "total_refund_amount") && reportsSQLiteHasColumns(r.db, "sale_items", "id", "tax_amount", "total_amount")
	returnedTaxable := 0.0
	if canCalculateReturnedTax {
		returnDate := r.returnDateExpression("r")
		returnReferenceFilter := ""
		if !dbutil.IsSQLite(r.db) || reportsSQLiteHasColumns(r.db, "accounting_returns", "reference_number") {
			returnReferenceFilter = `AND COALESCE(r.reference_number, '') NOT LIKE 'REV-%'`
		}
		var returnedAmounts struct {
			Taxable float64 `db:"taxable"`
		}
		if err := r.db.GetContext(ctx, &returnedAmounts, fmt.Sprintf(`
		SELECT
			COALESCE(SUM(CASE WHEN COALESCE(si.tax_amount, 0) > 0 THEN COALESCE(ri.total_refund_amount, 0) * COALESCE((si.total_amount - si.tax_amount) / NULLIF(si.total_amount, 0), 0) ELSE 0 END), 0) AS taxable
		FROM accounting_returns r
		JOIN accounting_return_items ri ON ri.return_id = r.id
		LEFT JOIN sale_items si ON si.id = ri.sale_item_id
		WHERE %s >= date(substr($1, 1, 10))
		  AND %s < date(substr($2, 1, 10))
			AND UPPER(COALESCE(r.status, '')) = 'COMPLETED' %s`, returnDate, returnDate, returnReferenceFilter), startDateKey, endDateKey); err != nil {
			return nil, fmt.Errorf("calculate returned taxable sales: %w", err)
		}
		returnedTaxable = returnedAmounts.Taxable
	}
	report.NetTaxableSales = report.TaxableSales - returnedTaxable
	report.NetTaxCollected = report.TaxCollected - report.ReturnedTax
	report.NetSalesTotal = report.SalesTotal - report.ReturnsTotal
	return report, nil
}
