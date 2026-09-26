package debts

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/dashboard"
	dbutil "github.com/partflow/smart-store/internal/database"
)

type Handler struct {
	db    *sqlx.DB
	cache *debtsCache
}

type DebtSummary struct {
	TotalDebt       float64 `json:"total_debt" db:"total_debt"`
	PaidAmount      float64 `json:"paid_amount" db:"paid_amount"`
	RemainingAmount float64 `json:"remaining_amount" db:"remaining_amount"`
	CustomerCount   int     `json:"customer_count" db:"customer_count"`
}

type debtsCache struct {
	data       interface{}
	expiration time.Time
	mu         sync.RWMutex
}

func newDebtsCache() *debtsCache {
	return &debtsCache{}
}

func (c *debtsCache) get() (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if time.Now().Before(c.expiration) {
		return c.data, true
	}
	return nil, false
}

func (c *debtsCache) set(data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = data
	c.expiration = time.Now().Add(ttl)
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{
		db:    db,
		cache: newDebtsCache(),
	}
}

func (h *Handler) loadDebtSummary() (DebtSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(amount), 0) AS total_debt,
			COALESCE(SUM(CASE WHEN amount > COALESCE(remaining_amount, 0) THEN amount - COALESCE(remaining_amount, 0) ELSE 0 END), 0) AS paid_amount,
			COALESCE(SUM(COALESCE(remaining_amount, 0)), 0) AS remaining_amount,
			COUNT(DISTINCT CASE WHEN COALESCE(remaining_amount, 0) > 0 THEN customer_id END) AS customer_count
		FROM debts
	`
	var summary DebtSummary
	if err := h.db.Get(&summary, query); err != nil {
		return DebtSummary{}, fmt.Errorf("failed to load debt summary: %w", err)
	}
	return summary, nil
}

// RegisterRoutes registers debt routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	debts := router.Group("/debts")
	{
		debts.GET("", h.ListDebts)
		debts.GET("/:id", h.GetDebt)
		debts.POST("", h.CreateDebt)
		debts.PUT("/:id", h.UpdateDebt)
		debts.DELETE("/:id", h.DeleteDebt)
		debts.GET("/customer/:customer_id", h.GetCustomerDebts)
		debts.GET("/overdue", h.GetOverdueDebts)
		debts.GET("/summary", h.GetDebtSummary)
		debts.POST("/:id/payment", h.AddPayment)
	}
}

// ListDebt lists all debts with pagination
func (h *Handler) ListDebts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 || perPage < 1 || perPage > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
		return
	}
	summary, err := h.loadDebtSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	query, countQuery, filterArgs, err := buildDebtListQueries(h.db, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	placeholder := func(index int) string {
		if dbutil.IsSQLite(h.db) {
			return "?"
		}
		return fmt.Sprintf("$%d", index)
	}

	listArgs := append([]interface{}{}, filterArgs...)
	listArgs = append(listArgs, perPage, (page-1)*perPage)
	query += fmt.Sprintf(" ORDER BY due_date DESC, created_at DESC LIMIT %s OFFSET %s", placeholder(len(filterArgs)+1), placeholder(len(filterArgs)+2))

	var rows []struct {
		ID              string  `db:"id"`
		CustomerID      string  `db:"customer_id"`
		CustomerName    string  `db:"customer_name"`
		CustomerCode    string  `db:"customer_code"`
		CustomerPhone   string  `db:"customer_phone"`
		InvoiceNumber   string  `db:"invoice_number"`
		Amount          float64 `db:"amount"`
		RemainingAmount float64 `db:"remaining_amount"`
		DueDate         string  `db:"due_date"`
		Status          string  `db:"status"`
		CreatedAt       string  `db:"created_at"`
	}
	if err := h.db.Select(&rows, query, listArgs...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var total int
	if err := h.db.Get(&total, countQuery, filterArgs...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		data = append(data, gin.H{
			"id": row.ID, "customer_id": row.CustomerID, "customer_name": row.CustomerName,
			"customer_code": row.CustomerCode, "customer_phone": row.CustomerPhone, "invoice_number": row.InvoiceNumber,
			"amount": row.Amount, "remaining_amount": row.RemainingAmount, "due_date": row.DueDate,
			"status": row.Status, "created_at": row.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data, "meta": gin.H{"page": page, "per_page": perPage, "total": total, "summary": summary}})
}

func buildDebtListQueries(db *sqlx.DB, c *gin.Context) (string, string, []interface{}, error) {
	placeholder := func(index int) string {
		if dbutil.IsSQLite(db) {
			return "?"
		}
		return fmt.Sprintf("$%d", index)
	}
	idExpression := "(array_agg(d.id ORDER BY d.due_date DESC))[1]"
	if dbutil.IsSQLite(db) {
		idExpression = "MIN(d.id)"
	}
	groupedQuery := fmt.Sprintf(`
		SELECT %s AS id, d.customer_id AS customer_id, c.name AS customer_name,
			c.code AS customer_code, COALESCE(c.phone, '') AS customer_phone,
			CASE WHEN COUNT(*) = 1 THEN COALESCE(MAX(s.invoice_number), '') ELSE 'عدة ديون' END AS invoice_number,
			SUM(d.amount) AS amount, SUM(d.remaining_amount) AS remaining_amount,
			COALESCE(MIN(CASE WHEN d.remaining_amount > 0 THEN CAST(d.due_date AS TEXT) END), MAX(CAST(d.due_date AS TEXT))) AS due_date,
			CASE WHEN SUM(d.remaining_amount) <= 0 THEN 'paid'
				WHEN SUM(CASE WHEN d.status = 'overdue' THEN 1 ELSE 0 END) > 0 THEN 'overdue'
				WHEN SUM(d.remaining_amount) < SUM(d.amount) THEN 'partial' ELSE 'pending' END AS status,
			MAX(CAST(d.created_at AS TEXT)) AS created_at
		FROM debts d
		JOIN customers c ON d.customer_id = c.id
		LEFT JOIN sales s ON d.sale_id = s.id
		GROUP BY d.customer_id, c.name, c.code, c.phone
	`, idExpression)

	conditions := make([]string, 0, 6)
	args := make([]interface{}, 0, 8)
	addArg := func(value interface{}) string {
		args = append(args, value)
		return placeholder(len(args))
	}
	search := strings.TrimSpace(c.Query("search"))
	searchType := strings.ToLower(strings.TrimSpace(c.DefaultQuery("search_type", "all")))
	if search != "" {
		escaped := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(search)
		pattern := "%" + escaped + "%"
		var searchFields []string
		switch searchType {
		case "all":
			searchFields = []string{"LOWER(customer_name) LIKE LOWER(%s) ESCAPE '!'", "LOWER(customer_code) LIKE LOWER(%s) ESCAPE '!'", "LOWER(customer_phone) LIKE LOWER(%s) ESCAPE '!'"}
		case "name":
			searchFields = []string{"LOWER(customer_name) LIKE LOWER(%s) ESCAPE '!'"}
		case "code":
			searchFields = []string{"LOWER(customer_code) LIKE LOWER(%s) ESCAPE '!'"}
		case "phone":
			searchFields = []string{"LOWER(customer_phone) LIKE LOWER(%s) ESCAPE '!'"}
		case "amount":
			searchFields = []string{"CAST(amount AS TEXT) LIKE %s ESCAPE '!'"}
		default:
			return "", "", nil, fmt.Errorf("invalid search_type")
		}
		parts := make([]string, 0, len(searchFields))
		for _, field := range searchFields {
			parts = append(parts, fmt.Sprintf(field, addArg(pattern)))
		}
		conditions = append(conditions, "("+strings.Join(parts, " OR ")+")")
	}
	if customerIDValue := strings.TrimSpace(c.Query("customer_id")); customerIDValue != "" {
		customerID, err := uuid.Parse(customerIDValue)
		if err != nil {
			return "", "", nil, fmt.Errorf("invalid customer_id")
		}
		conditions = append(conditions, "customer_id = "+addArg(customerID.String()))
	}

	status := strings.ToLower(strings.TrimSpace(c.Query("status")))
	if status != "" && status != "all" {
		if status != "paid" && status != "overdue" && status != "partial" {
			return "", "", nil, fmt.Errorf("invalid debt status")
		}
		conditions = append(conditions, "status = "+addArg(status))
	}
	switch strings.ToLower(strings.TrimSpace(c.Query("tab"))) {
	case "", "all":
	case "open":
		conditions = append(conditions, "status <> 'paid'")
	case "paid":
		conditions = append(conditions, "status = 'paid'")
	default:
		return "", "", nil, fmt.Errorf("invalid debt tab")
	}

	for _, field := range []struct {
		name string
		op   string
	}{
		{name: "min_amount", op: ">="},
		{name: "max_amount", op: "<="},
	} {
		value := strings.TrimSpace(c.Query(field.name))
		if value == "" {
			continue
		}
		amount, err := strconv.ParseFloat(value, 64)
		if err != nil || amount < 0 {
			return "", "", nil, fmt.Errorf("invalid %s", field.name)
		}
		conditions = append(conditions, "amount "+field.op+" "+addArg(amount))
	}
	if minValue, maxValue := c.Query("min_amount"), c.Query("max_amount"); minValue != "" && maxValue != "" {
		minAmount, _ := strconv.ParseFloat(minValue, 64)
		maxAmount, _ := strconv.ParseFloat(maxValue, 64)
		if minAmount > maxAmount {
			return "", "", nil, fmt.Errorf("min_amount cannot exceed max_amount")
		}
	}

	for _, field := range []struct {
		name string
		op   string
	}{
		{name: "due_date_from", op: ">="},
		{name: "due_date_to", op: "<="},
	} {
		value := strings.TrimSpace(c.Query(field.name))
		if value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return "", "", nil, fmt.Errorf("invalid %s", field.name)
		}
		conditions = append(conditions, "SUBSTR(CAST(due_date AS TEXT), 1, 10) "+field.op+" "+addArg(value))
	}
	if from, to := c.Query("due_date_from"), c.Query("due_date_to"); from != "" && to != "" && from > to {
		return "", "", nil, fmt.Errorf("due_date_from cannot exceed due_date_to")
	}

	listQuery := "SELECT * FROM (" + groupedQuery + ") AS grouped_debts"
	countQuery := "SELECT COUNT(*) FROM (" + groupedQuery + ") AS grouped_debts"
	if len(conditions) > 0 {
		where := " WHERE " + strings.Join(conditions, " AND ")
		listQuery += where
		countQuery += where
	}
	return listQuery, countQuery, args, nil
}

// GetDebt retrieves a debt by ID
func (h *Handler) GetDebt(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt ID"})
		return
	}
	if dbutil.IsSQLite(h.db) {
		var debt struct {
			ID, CustomerID, CustomerName, DueDate, Status, Notes, CreatedAt, UpdatedAt string
			Amount, RemainingAmount                                                    float64
		}
		err = h.db.Get(&debt, `SELECT d.id, d.customer_id, c.name AS customer_name, d.amount, d.remaining_amount,
			d.due_date, d.status, COALESCE(d.notes,''), d.created_at, d.updated_at
			FROM debts d JOIN customers c ON d.customer_id = c.id WHERE d.id = ?`, id.String())
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "debt not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": debt})
		return
	}

	var debt struct {
		ID              uuid.UUID `json:"id"`
		CustomerID      uuid.UUID `json:"customer_id"`
		CustomerName    string    `json:"customer_name"`
		Amount          float64   `json:"amount"`
		RemainingAmount float64   `json:"remaining_amount"`
		DueDate         string    `json:"due_date"`
		Status          string    `json:"status"`
		Notes           string    `json:"notes"`
		CreatedAt       string    `json:"created_at"`
		UpdatedAt       string    `json:"updated_at"`
	}

	query := `
		SELECT d.id, d.customer_id, c.name as customer_name, d.amount, d.remaining_amount,
		       d.due_date, d.status, d.notes, d.created_at, d.updated_at
		FROM debts d
		JOIN customers c ON d.customer_id = c.id
		WHERE d.id = $1
	`

	row := h.db.QueryRow(query, id)
	err = row.Scan(&debt.ID, &debt.CustomerID, &debt.CustomerName, &debt.Amount, &debt.RemainingAmount, &debt.DueDate, &debt.Status, &debt.Notes, &debt.CreatedAt, &debt.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "debt not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    debt,
	})
}

// CreateDebt creates a new debt
func (h *Handler) CreateDebt(c *gin.Context) {
	var req struct {
		CustomerID uuid.UUID `json:"customer_id" binding:"required"`
		Amount     float64   `json:"amount" binding:"required"`
		DueDate    string    `json:"due_date" binding:"required"`
		Notes      string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than zero"})
		return
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(req.DueDate)); err != nil {
		if _, err = time.Parse(time.RFC3339, strings.TrimSpace(req.DueDate)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "due_date must be a valid date"})
			return
		}
	}

	id := uuid.New()
	query := fmt.Sprintf(`
		INSERT INTO debts (id, customer_id, amount, remaining_amount, due_date, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $3, $4, 'pending', $5, %s, %s)
	`, dbutil.NowSQL(h.db), dbutil.NowSQL(h.db))

	_, err := h.db.Exec(query, id, req.CustomerID, req.Amount, req.DueDate, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.cache.set(nil, 0)
	dashboard.InvalidateDashboardCache()

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id": id,
		},
		"message": "debt created successfully",
	})
}

// UpdateDebt updates a debt
func (h *Handler) UpdateDebt(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt ID"})
		return
	}

	// Debt amounts and payment state must change through financial transactions,
	// not by overwriting the debt row. This endpoint only edits descriptive data.
	var req struct {
		DueDate *string `json:"due_date"`
		Notes   *string `json:"notes"`
	}
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only due_date and notes can be updated"})
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body must contain one JSON object"})
		return
	}
	if req.DueDate == nil && req.Notes == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide due_date or notes"})
		return
	}
	if req.DueDate != nil {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*req.DueDate)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "due_date must use YYYY-MM-DD"})
			return
		}
	}
	if req.Notes != nil && len([]rune(*req.Notes)) > 5000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notes must not exceed 5000 characters"})
		return
	}

	updates := make([]string, 0, 2)
	args := make([]any, 0, 3)
	if req.DueDate != nil {
		updates = append(updates, "due_date = ?")
		args = append(args, strings.TrimSpace(*req.DueDate))
	}
	if req.Notes != nil {
		updates = append(updates, "notes = ?")
		args = append(args, *req.Notes)
	}
	query := fmt.Sprintf("UPDATE debts SET %s, updated_at = %s WHERE id = ?", strings.Join(updates, ", "), dbutil.NowSQL(h.db))
	args = append(args, id.String())
	result, err := h.db.ExecContext(c.Request.Context(), h.db.Rebind(query), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update debt details"})
		return
	}
	affected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to confirm debt update"})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "debt not found"})
		return
	}
	h.cache.set(nil, 0)
	dashboard.InvalidateDashboardCacheWithReason("debt_details_updated")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "debt updated successfully",
	})
}

// DeleteDebt deletes a debt
func (h *Handler) DeleteDebt(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt ID"})
		return
	}

	_, err = h.db.Exec("DELETE FROM debts WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.cache.set(nil, 0)
	dashboard.InvalidateDashboardCache()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "debt deleted successfully",
	})
}

// GetCustomerDebts retrieves debts for a specific customer
func (h *Handler) GetCustomerDebts(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("customer_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer ID"})
		return
	}
	type debtLineItem struct {
		ProductID   string  `json:"product_id" db:"product_id"`
		ProductName string  `json:"product_name" db:"product_name"`
		Quantity    int     `json:"quantity" db:"quantity"`
		UnitPrice   float64 `json:"unit_price" db:"unit_price"`
		TotalAmount float64 `json:"total_amount" db:"total_amount"`
	}
	type customerDebt struct {
		ID              string         `json:"id" db:"id"`
		CustomerID      string         `json:"customer_id" db:"customer_id"`
		SaleID          sql.NullString `json:"-" db:"sale_id"`
		ReferenceType   string         `json:"reference_type" db:"reference_type"`
		InvoiceNumber   string         `json:"invoice_number" db:"invoice_number"`
		Amount          float64        `json:"amount" db:"amount"`
		PaidAmount      float64        `json:"paid_amount" db:"paid_amount"`
		RemainingAmount float64        `json:"remaining_amount" db:"remaining_amount"`
		DueDate         string         `json:"due_date" db:"due_date"`
		Status          string         `json:"status" db:"status"`
		Notes           string         `json:"notes" db:"notes"`
		CreatedAt       string         `json:"created_at" db:"created_at"`
		Items           []debtLineItem `json:"items" db:"-"`
	}

	query := `
		SELECT * FROM (
			SELECT CAST(d.id AS TEXT) AS id,
				CAST(d.customer_id AS TEXT) AS customer_id,
				CAST(d.sale_id AS TEXT) AS sale_id,
				CASE WHEN d.sale_id IS NOT NULL THEN 'sale'
					WHEN UPPER(TRIM(COALESCE(d.notes, ''))) = 'OPENING_DEBT' THEN 'opening_debt'
					WHEN UPPER(TRIM(COALESCE(d.notes, ''))) = 'MANUAL_ADJUSTMENT' THEN 'manual_adjustment'
					ELSE 'manual' END AS reference_type,
				COALESCE(s.invoice_number, '') AS invoice_number,
				d.amount,
				COALESCE(d.paid_amount, 0) AS paid_amount,
				COALESCE(d.remaining_amount, d.amount - COALESCE(d.paid_amount, 0)) AS remaining_amount,
				CAST(d.due_date AS TEXT) AS due_date,
				COALESCE(d.status, 'pending') AS status,
				COALESCE(d.notes, '') AS notes,
				CAST(d.created_at AS TEXT) AS created_at
			FROM debts d
			LEFT JOIN sales s ON s.id = d.sale_id
			WHERE d.customer_id = ?

			UNION ALL

			SELECT CAST(cd.id AS TEXT) AS id,
				CAST(cd.customer_id AS TEXT) AS customer_id,
				CASE WHEN LOWER(TRIM(COALESCE(cd.reference_type, ''))) = 'sale'
					THEN CAST(cd.reference_id AS TEXT) ELSE '' END AS sale_id,
				COALESCE(NULLIF(LOWER(TRIM(cd.reference_type)), ''), 'manual') AS reference_type,
				COALESCE(s.invoice_number, '') AS invoice_number,
				cd.amount,
				COALESCE(cd.paid_amount, 0) AS paid_amount,
				CASE WHEN cd.amount - COALESCE(cd.paid_amount, 0) < 0 THEN 0
					ELSE cd.amount - COALESCE(cd.paid_amount, 0) END AS remaining_amount,
				CAST(cd.due_date AS TEXT) AS due_date,
				CASE WHEN cd.is_paid THEN 'paid'
					WHEN COALESCE(cd.paid_amount, 0) > 0 THEN 'partial'
					ELSE 'pending' END AS status,
				COALESCE(cd.reference_type, '') AS notes,
				CAST(cd.created_at AS TEXT) AS created_at
			FROM customer_debts cd
			LEFT JOIN sales s ON s.id = cd.reference_id
				AND LOWER(TRIM(COALESCE(cd.reference_type, ''))) = 'sale'
			WHERE cd.customer_id = ?
				AND NOT EXISTS (SELECT 1 FROM debts d WHERE d.customer_id = cd.customer_id AND d.id = cd.id)
				AND (LOWER(TRIM(COALESCE(cd.reference_type, ''))) <> 'sale'
					OR NOT EXISTS (SELECT 1 FROM debts d WHERE d.customer_id = cd.customer_id AND d.sale_id = cd.reference_id))
		) AS customer_debt_history
		ORDER BY created_at DESC, id DESC
	`
	var debts []customerDebt
	if err := h.db.SelectContext(c.Request.Context(), &debts, h.db.Rebind(query), customerID.String(), customerID.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if debts == nil {
		debts = []customerDebt{}
	}

	saleIDs := make([]uuid.UUID, 0, len(debts))
	for i := range debts {
		debts[i].Items = []debtLineItem{}
		if !debts[i].SaleID.Valid || strings.TrimSpace(debts[i].SaleID.String) == "" {
			continue
		}
		saleID, parseErr := uuid.Parse(debts[i].SaleID.String)
		if parseErr != nil {
			continue
		}
		saleIDs = append(saleIDs, saleID)
	}
	if len(saleIDs) > 0 {
		itemsQuery, args, inErr := sqlx.In(`
			SELECT CAST(si.sale_id AS TEXT) AS sale_id,
				CAST(si.product_id AS TEXT) AS product_id,
				COALESCE(p.name, 'منتج محذوف') AS product_name,
				si.quantity,
				si.unit_price,
				COALESCE(si.total_amount, si.unit_price * si.quantity) AS total_amount
			FROM sale_items si
		LEFT JOIN products p ON p.id = si.product_id
		WHERE si.sale_id IN (?)
		ORDER BY si.created_at, si.id`, saleIDs)
		if inErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": inErr.Error()})
			return
		}
		var itemQueryRows []struct {
			SaleID      string  `db:"sale_id"`
			ProductID   string  `db:"product_id"`
			ProductName string  `db:"product_name"`
			Quantity    int     `db:"quantity"`
			UnitPrice   float64 `db:"unit_price"`
			TotalAmount float64 `db:"total_amount"`
		}
		if err := h.db.SelectContext(c.Request.Context(), &itemQueryRows, h.db.Rebind(itemsQuery), args...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		itemsBySale := make(map[string][]debtLineItem, len(saleIDs))
		for _, row := range itemQueryRows {
			itemsBySale[row.SaleID] = append(itemsBySale[row.SaleID], debtLineItem{
				ProductID: row.ProductID, ProductName: row.ProductName,
				Quantity: row.Quantity, UnitPrice: row.UnitPrice, TotalAmount: row.TotalAmount,
			})
		}
		for i := range debts {
			if debts[i].SaleID.Valid {
				debts[i].Items = itemsBySale[debts[i].SaleID.String]
				if debts[i].Items == nil {
					debts[i].Items = []debtLineItem{}
				}
			}
		}
	}

	data := make([]gin.H, 0, len(debts))
	for _, debt := range debts {
		data = append(data, gin.H{
			"id": debt.ID, "customer_id": debt.CustomerID,
			"reference_type": debt.ReferenceType,
			"sale_id":        strings.TrimSpace(debt.SaleID.String), "invoice_number": debt.InvoiceNumber,
			"amount": debt.Amount, "paid_amount": debt.PaidAmount,
			"remaining_amount": debt.RemainingAmount, "due_date": debt.DueDate,
			"status": debt.Status, "notes": debt.Notes, "created_at": debt.CreatedAt,
			"items": debt.Items,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// GetOverdueDebts retrieves all overdue debts
func (h *Handler) GetOverdueDebts(c *gin.Context) {
	// Try cache first
	if cached, found := h.cache.get(); found {
		c.JSON(http.StatusOK, cached)
		return
	}
	if dbutil.IsSQLite(h.db) {
		storeDate, err := accounting.StoreDate(time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var rows []struct {
			ID, CustomerID, CustomerName, DueDate, CreatedAt string
			Amount, RemainingAmount                          float64
			DaysOverdue                                      int
		}
		query := `SELECT d.id, d.customer_id, c.name AS customer_name, d.amount, d.remaining_amount, d.due_date,
			CAST(julianday(?) - julianday(substr(d.due_date, 1, 10)) AS INTEGER) AS days_overdue, d.created_at
			FROM debts d JOIN customers c ON d.customer_id = c.id WHERE date(substr(d.due_date, 1, 10)) < date(?) AND d.remaining_amount > 0 ORDER BY d.due_date ASC`
		if err := h.db.Select(&rows, query, storeDate, storeDate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		data := make([]gin.H, 0, len(rows))
		for _, row := range rows {
			data = append(data, gin.H{"id": row.ID, "customer_id": row.CustomerID, "customer_name": row.CustomerName, "amount": row.Amount, "remaining_amount": row.RemainingAmount, "due_date": row.DueDate, "days_overdue": row.DaysOverdue, "created_at": row.CreatedAt})
		}
		response := gin.H{"success": true, "data": data}
		h.cache.set(response, 2*time.Minute)
		c.JSON(http.StatusOK, response)
		return
	}

	var debts []struct {
		ID              uuid.UUID `json:"id"`
		CustomerID      uuid.UUID `json:"customer_id"`
		CustomerName    string    `json:"customer_name"`
		Amount          float64   `json:"amount"`
		RemainingAmount float64   `json:"remaining_amount"`
		DueDate         string    `json:"due_date"`
		DaysOverdue     string    `json:"days_overdue"`
		CreatedAt       string    `json:"created_at"`
	}

	query := `
		SELECT d.id, d.customer_id, c.name as customer_name, d.amount, d.remaining_amount,
		       d.due_date, ($1::date - d.due_date::date) as days_overdue, d.created_at
		FROM debts d
		JOIN customers c ON d.customer_id = c.id
		WHERE d.due_date < $1::date AND d.remaining_amount > 0
		ORDER BY d.due_date ASC
	`

	storeDate, err := accounting.StoreDate(time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := h.db.Query(query, storeDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var debt struct {
			ID              uuid.UUID `json:"id"`
			CustomerID      uuid.UUID `json:"customer_id"`
			CustomerName    string    `json:"customer_name"`
			Amount          float64   `json:"amount"`
			RemainingAmount float64   `json:"remaining_amount"`
			DueDate         string    `json:"due_date"`
			DaysOverdue     string    `json:"days_overdue"`
			CreatedAt       string    `json:"created_at"`
		}
		if err := rows.Scan(&debt.ID, &debt.CustomerID, &debt.CustomerName, &debt.Amount, &debt.RemainingAmount, &debt.DueDate, &debt.DaysOverdue, &debt.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		debts = append(debts, debt)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"success": true,
		"data":    debts,
	}

	// Cache the response
	h.cache.set(response, 2*time.Minute)

	c.JSON(http.StatusOK, response)
}

// GetDebtSummary retrieves debt summary statistics
func (h *Handler) GetDebtSummary(c *gin.Context) {
	var summary struct {
		TotalDebts     float64 `json:"total_debts"`
		TotalRemaining float64 `json:"total_remaining"`
		OverdueDebts   float64 `json:"overdue_debts"`
		OverdueCount   int     `json:"overdue_count"`
		PendingDebts   float64 `json:"pending_debts"`
		PaidDebts      float64 `json:"paid_debts"`
	}

	query := `
		SELECT 
			COALESCE(SUM(amount), 0) as total_debts,
			COALESCE(SUM(remaining_amount), 0) as total_remaining,
			COALESCE(SUM(CASE WHEN due_date < $1::date AND remaining_amount > 0 THEN remaining_amount ELSE 0 END), 0) as overdue_debts,
			COUNT(CASE WHEN due_date < $1::date AND remaining_amount > 0 THEN 1 END) as overdue_count,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN remaining_amount ELSE 0 END), 0) as pending_debts,
			COALESCE(SUM(CASE WHEN remaining_amount = 0 THEN amount ELSE 0 END), 0) as paid_debts
		FROM debts
	`

	storeDate, err := accounting.StoreDate(time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dbutil.IsSQLite(h.db) {
		query = strings.ReplaceAll(query, "$1::date", "date(?)")
	}
	args := []any{storeDate}
	if dbutil.IsSQLite(h.db) {
		args = []any{storeDate, storeDate}
	}
	row := h.db.QueryRow(query, args...)
	err = row.Scan(&summary.TotalDebts, &summary.TotalRemaining, &summary.OverdueDebts, &summary.OverdueCount, &summary.PendingDebts, &summary.PaidDebts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// AddPayment adds a payment to a debt
func (h *Handler) AddPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt ID"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
		Notes  string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Amount <= 0 || math.IsNaN(req.Amount) || math.IsInf(req.Amount, 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than zero"})
		return
	}

	// Start transaction
	tx, err := h.db.BeginTxx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()
	var customerID string
	var debt struct {
		Amount          float64 `db:"amount"`
		PaidAmount      float64 `db:"paid_amount"`
		RemainingAmount float64 `db:"remaining_amount"`
	}
	if dbutil.IsSQLite(h.db) {
		// Acquire SQLite's write lock before reading the balance so two concurrent
		// collections cannot both pass the same remaining-amount check.
		result, err := tx.ExecContext(c.Request.Context(), `UPDATE debts SET updated_at = updated_at WHERE id = ?`, id.String())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			c.JSON(http.StatusNotFound, gin.H{"error": "debt not found"})
			return
		}
		if err := tx.GetContext(c.Request.Context(), &customerID, `SELECT customer_id FROM debts WHERE id = ?`, id.String()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "debt not found"})
			return
		}
		if err := tx.GetContext(c.Request.Context(), &debt, `SELECT amount, COALESCE(paid_amount, 0) AS paid_amount, COALESCE(remaining_amount, MAX(amount - COALESCE(paid_amount, 0), 0)) AS remaining_amount FROM debts WHERE id = ?`, id.String()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "debt not found"})
			return
		}
	} else {
		if err := tx.GetContext(c.Request.Context(), &customerID, `SELECT customer_id FROM debts WHERE id = $1 FOR UPDATE`, id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "debt not found"})
			return
		}
		if err := tx.GetContext(c.Request.Context(), &debt, `SELECT amount, COALESCE(paid_amount, 0) AS paid_amount, COALESCE(remaining_amount, GREATEST(amount - COALESCE(paid_amount, 0), 0)) AS remaining_amount FROM debts WHERE id = $1`, id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "debt not found"})
			return
		}
	}
	if req.Amount > debt.RemainingAmount+0.000001 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment amount exceeds remaining debt balance"})
		return
	}
	newRemainingAmount := debt.RemainingAmount - req.Amount
	if newRemainingAmount <= 0.000001 {
		newRemainingAmount = 0
	}
	newPaidAmount := debt.Amount - newRemainingAmount
	if newPaidAmount > debt.Amount {
		newPaidAmount = debt.Amount
	}
	updatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	updateQuery := `UPDATE debts SET paid_amount = ?, remaining_amount = ?, status = CASE WHEN ? THEN 'paid' ELSE status END, updated_at = ? WHERE id = ? AND COALESCE(remaining_amount, 0) >= ?`
	result, err := tx.ExecContext(c.Request.Context(), tx.Rebind(updateQuery), newPaidAmount, newRemainingAmount, newRemainingAmount == 0, updatedAt, id.String(), req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "debt balance changed; refresh and try again"})
		return
	}
	if dbutil.IsSQLite(h.db) {
		paymentID := uuid.New().String()
		if _, err := tx.ExecContext(c.Request.Context(), `INSERT INTO payments (id, transaction_number, customer_id, amount, payment_method, reference, notes, created_at) VALUES (?, ?, ?, ?, 'cash', ?, ?, ?)`, paymentID, "PAY-"+paymentID[:8], customerID, req.Amount, id.String(), req.Notes, updatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		storeDate, err := accounting.StoreDate(accounting.StoreNow())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		paymentID := uuid.New()
		paymentQuery := `
			INSERT INTO payments (id, reference_number, customer_id, amount, payment_method, payment_date, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'cash', $5, $6, NOW(), NOW())
		`
		if _, err = tx.ExecContext(c.Request.Context(), paymentQuery, paymentID, "PAY-"+paymentID.String()[:8], customerID, req.Amount, storeDate, req.Notes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.cache.set(nil, 0)
	dashboard.InvalidateDashboardCache()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "payment added successfully",
	})
}
