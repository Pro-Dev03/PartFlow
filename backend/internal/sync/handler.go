package sync

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Handler struct {
	db *sqlx.DB
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{db: db}
}

// GetInitialData returns all user data for initial sync
func (h *Handler) GetInitialData(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User ID not found in context",
		})
		return
	}

	data := gin.H{
		"customers":    h.getCustomersData(userID),
		"products":     h.getProductsData(userID),
		"categories":   h.getCategoriesData(userID),
		"suppliers":    h.getSuppliersData(userID),
		"sales":        h.getSalesData(userID),
		"purchases":    h.getPurchasesData(userID),
		"debts":        h.getDebtsData(userID),
		"payments":     h.getPaymentsData(userID),
		"expenses":     h.getExpensesData(userID),
		"inspections":  h.getInspectionsData(userID),
		"part_types":   h.getPartTypesData(userID),
		"acquisitions": h.getAcquisitionsData(userID),
		"returns":      h.getReturnsData(userID),
		"used_parts":   h.getUsedPartsData(userID),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// Helper functions to fetch data from database
// Note: These are simplified - adjust based on actual database schema

func (h *Handler) getCustomersData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, name, phone, email, address, balance, created_at, updated_at
		FROM customers
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getProductsData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, name, barcode, category_id, quantity, min_quantity, price, cost, 
		       description, active, created_at, updated_at
		FROM products
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getCategoriesData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getSuppliersData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, name, phone, email, address, city, balance, created_at, updated_at
		FROM suppliers
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getSalesData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, customer_id, total_amount, discount, tax, notes, payment_status, 
		       created_at, updated_at
		FROM sales
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getPurchasesData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, supplier_id, total_amount, tax, status, received_at, created_at, updated_at
		FROM purchases
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getDebtsData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, customer_id, amount, remaining_amount, due_date, notes, status, 
		       created_at, updated_at
		FROM debts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getPaymentsData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, debt_id, customer_id, amount, payment_method, notes, created_at, updated_at
		FROM payments
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getExpensesData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, category, amount, description, created_at, updated_at
		FROM expenses
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getInspectionsData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, product_id, quantity, status, notes, created_at, updated_at
		FROM inspections
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getPartTypesData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM part_types
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getAcquisitionsData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, product_id, quantity, cost, source, created_at, updated_at
		FROM acquisitions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getReturnsData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, sale_id, product_id, quantity, reason, notes, created_at, updated_at
		FROM returns
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}

func (h *Handler) getUsedPartsData(userID string) []map[string]interface{} {
	var results []map[string]interface{}
	query := `
		SELECT id, name, description, quantity, price, created_at, updated_at
		FROM used_parts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := h.db.Queryx(query, userID)
	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		record := make(map[string]interface{})
		if err := rows.MapScan(record); err != nil {
			continue
		}
		results = append(results, record)
	}
	return results
}
