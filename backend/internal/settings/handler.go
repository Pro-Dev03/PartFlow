package settings

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *sql.DB
}

type Setting struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	ValueType   string  `json:"value_type"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	IsPublic    bool    `json:"is_public"`
}

type UpdateSettingRequest struct {
	Value string `json:"value" binding:"required"`
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// GetPublicSettings returns all public settings
func (h *Handler) GetPublicSettings(c *gin.Context) {
	query := `
		SELECT key, value, value_type, category, description, is_public
		FROM settings
		WHERE is_public = true
		ORDER BY category, key
	`

	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}
	defer rows.Close()

	var settings []Setting
	for rows.Next() {
		var setting Setting
		if err := rows.Scan(&setting.Key, &setting.Value, &setting.ValueType, &setting.Category, &setting.Description, &setting.IsPublic); err != nil {
			continue
		}
		settings = append(settings, setting)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// GetSetting returns a specific setting by key
func (h *Handler) GetSetting(c *gin.Context) {
	key := c.Param("key")

	var setting Setting
	query := `
		SELECT key, value, value_type, category, description, is_public
		FROM settings
		WHERE key = $1
	`

	err := h.db.QueryRow(query, key).Scan(
		&setting.Key, &setting.Value, &setting.ValueType, &setting.Category, &setting.Description, &setting.IsPublic,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    setting,
	})
}

// UpdateSetting updates a setting value
func (h *Handler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")

	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	query := `
		UPDATE settings
		SET value = $1, updated_at = NOW()
		WHERE key = $2
		RETURNING key, value, value_type, category, description, is_public
	`

	var setting Setting
	err := h.db.QueryRow(query, req.Value, key).Scan(
		&setting.Key, &setting.Value, &setting.ValueType, &setting.Category, &setting.Description, &setting.IsPublic,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    setting,
		"message": "Setting updated successfully",
	})
}

// GetTaxRate returns the current tax rate
func (h *Handler) GetTaxRate(c *gin.Context) {
	var taxRate float64
	query := `SELECT CAST(value AS FLOAT) FROM settings WHERE key = 'tax_rate'`

	err := h.db.QueryRow(query).Scan(&taxRate)
	if err != nil {
		taxRate = 0 // Default to 0 if not found
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tax_rate": taxRate,
		},
	})
}

// UpdateTaxRate updates the tax rate
func (h *Handler) UpdateTaxRate(c *gin.Context) {
	var req struct {
		TaxRate float64 `json:"tax_rate" binding:"required,min=0,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tax rate value"})
		return
	}

	query := `
		UPDATE settings
		SET value = $1, updated_at = NOW()
		WHERE key = 'tax_rate'
		RETURNING value
	`

	var updatedValue string
	err := h.db.QueryRow(query, strconv.FormatFloat(req.TaxRate, 'f', 2, 64)).Scan(&updatedValue)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tax rate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tax_rate": req.TaxRate,
		},
		"message": "Tax rate updated successfully",
	})
}
