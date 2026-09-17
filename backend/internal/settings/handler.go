package settings

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/partflow/smart-store/internal/accounting"
	"github.com/partflow/smart-store/internal/secrets"
)

type Handler struct {
	db *sql.DB
}

type Setting struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	ValueType   string `json:"value_type"`
	Category    string `json:"category"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

type UpdateSettingRequest struct {
	Value string `json:"value" binding:"required"`
}

var financialSettingMetadata = map[string]struct {
	defaultValue string
	valueType    string
	category     string
	description  string
	isPublic     bool
}{
	"currency":                    {"ILS", "string", "general", "العملة الافتراضية", true},
	"discounts_enabled":           {"true", "boolean", "financial", "السماح بالخصومات", false},
	"tax_rate":                    {"0", "number", "financial", "نسبة الضريبة المئوية", true},
	"max_discount_rate":           {"15", "number", "financial", "الحد الأقصى للخصم المئوي", false},
	"default_profit_margin":       {"30", "number", "financial", "نسبة الربح المقترحة عند إضافة منتج", false},
	"country_code":                {"IL", "string", "regional", "الدولة الافتراضية للمتجر", true},
	"store_timezone":              {accounting.DefaultStoreTimezone, "string", "regional", "المنطقة الزمنية الثابتة للمتجر", true},
	"pos_products_per_page":       {"12", "number", "appearance", "عدد منتجات نقطة البيع في الصفحة", false},
	"pos_product_view_mode":       {"cards", "string", "appearance", "طريقة عرض منتجات نقطة البيع", false},
	"electronic_payments_enabled": {"false", "boolean", "payments", "تفعيل الدفع الإلكتروني", false},
	"payment_provider":            {"manual", "string", "payments", "مزود الدفع الإلكتروني", false},
	"payment_environment":         {"test", "string", "payments", "بيئة الدفع الإلكتروني", false},
	"payment_public_key":          {"", "string", "payments", "المفتاح العام لمزود الدفع", false},
	"payment_secret_key":          {"", "secret", "payments", "المفتاح السري لمزود الدفع", false},
	"payment_merchant_id":         {"", "string", "payments", "معرف التاجر لدى مزود الدفع", false},
	"payment_terminal_id":         {"", "string", "payments", "معرف جهاز الدفع", false},
	"payment_webhook_url":         {"", "string", "payments", "عنوان Webhook للدفع", false},
	"payment_webhook_secret":      {"", "secret", "payments", "سر توقيع Webhook", false},
	"payment_methods":             {"[\"card\"]", "json", "payments", "طرق الدفع الإلكتروني المفعلة", false},
}

func isSensitiveSetting(key string) bool {
	return key == "payment_secret_key" || key == "payment_webhook_secret"
}

func redactSetting(setting *Setting) {
	if isSensitiveSetting(setting.Key) && setting.Value != "" {
		setting.Value = "********"
	}
}

func NewHandler(db *sql.DB) *Handler {
	var storeTimezone string
	if err := db.QueryRow(`SELECT value FROM settings WHERE key = 'store_timezone'`).Scan(&storeTimezone); err == nil {
		_ = accounting.ConfigureStoreTimezone(storeTimezone)
		return &Handler{db: db}
	}
	var countryCode string
	if err := db.QueryRow(`SELECT value FROM settings WHERE key = 'country_code'`).Scan(&countryCode); err == nil {
		if profile, profileErr := RegionalProfileForCountry(countryCode); profileErr == nil {
			_ = accounting.ConfigureStoreTimezone(profile.Timezone)
		}
	}
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
		redactSetting(&setting)
		settings = append(settings, setting)
	}
	_ = rows.Err()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// GetRegionalSettings returns the country catalog and the selected store profile.
func (h *Handler) GetRegionalSettings(c *gin.Context) {
	profile := DefaultRegionalProfile()
	storeTimezone := ""
	timezonePersisted := false
	if err := h.db.QueryRow(`SELECT value FROM settings WHERE key = 'store_timezone'`).Scan(&storeTimezone); err == nil {
		if _, loadErr := time.LoadLocation(storeTimezone); loadErr == nil {
			profile.Timezone = storeTimezone
			timezonePersisted = true
		}
	}
	var countryCode string
	if !timezonePersisted && h.db.QueryRow(`SELECT value FROM settings WHERE key = 'country_code'`).Scan(&countryCode) == nil {
		if selected, profileErr := RegionalProfileForCountry(countryCode); profileErr == nil {
			profile = selected
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"profile":            profile,
		"countries":          RegionalProfiles(),
		"store_timezone":     profile.Timezone,
		"timezone_persisted": timezonePersisted,
	}})
}

// UpdateRegionalSettings changes the country and its derived timezone policy.
func (h *Handler) UpdateRegionalSettings(c *gin.Context) {
	var request struct {
		CountryCode string `json:"country_code"`
		Timezone    string `json:"timezone"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "country_code is required"})
		return
	}
	profile := DefaultRegionalProfile()
	if strings.TrimSpace(request.Timezone) != "" {
		profile.Timezone = strings.TrimSpace(request.Timezone)
		if _, err := time.LoadLocation(profile.Timezone); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid IANA timezone %q", profile.Timezone)})
			return
		}
		for _, candidate := range RegionalProfiles() {
			if candidate.Timezone == profile.Timezone {
				profile = candidate
				break
			}
		}
	} else {
		var err error
		profile, err = RegionalProfileForCountry(request.CountryCode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if profile.CountryCode != "" {
		if _, err := h.db.Exec(`INSERT INTO settings (key, value, value_type, category, description, is_public) VALUES ($1, $2, 'string', 'regional', $3, true) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP`, "country_code", profile.CountryCode, "الدولة الافتراضية للمتجر"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update regional settings"})
			return
		}
	}
	if _, err := h.db.Exec(`INSERT INTO settings (key, value, value_type, category, description, is_public) VALUES ($1, $2, 'string', 'regional', $3, true) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP`, "store_timezone", profile.Timezone, "المنطقة الزمنية الثابتة للمتجر"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update regional settings"})
		return
	}
	if err := accounting.ConfigureStoreTimezone(profile.Timezone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure store timezone"})
		return
	}
	encoded, _ := json.Marshal(profile)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"profile": profile, "serialized": string(encoded)}})
}

// InitializeRegionalSettings persists the device-detected timezone once.
// Later changes must go through the administrator-only update endpoint.
func (h *Handler) InitializeRegionalSettings(c *gin.Context) {
	var request struct {
		Timezone string `json:"timezone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timezone is required"})
		return
	}
	timezone := strings.TrimSpace(request.Timezone)
	if _, err := time.LoadLocation(timezone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid IANA timezone %q", timezone)})
		return
	}

	var existing string
	if err := h.db.QueryRow(`SELECT value FROM settings WHERE key = 'store_timezone'`).Scan(&existing); err == nil && strings.TrimSpace(existing) != "" {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"store_timezone": existing, "timezone_persisted": true}})
		return
	}
	if _, err := h.db.Exec(`INSERT INTO settings (key, value, value_type, category, description, is_public, created_at, updated_at) VALUES ($1, $2, 'string', 'regional', $3, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON CONFLICT (key) DO NOTHING`, "store_timezone", timezone, "المنطقة الزمنية الثابتة للمتجر"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize store timezone"})
		return
	}
	if err := accounting.ConfigureStoreTimezone(timezone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure store timezone"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"store_timezone": timezone, "timezone_persisted": true}})
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
		if metadata, exists := financialSettingMetadata[key]; exists {
			if _, insertErr := h.db.Exec(`INSERT INTO settings (key, value, value_type, category, description, is_public) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (key) DO NOTHING`, key, metadata.defaultValue, metadata.valueType, metadata.category, metadata.description, metadata.isPublic); insertErr == nil {
				err = h.db.QueryRow(query, key).Scan(
					&setting.Key, &setting.Value, &setting.ValueType, &setting.Category, &setting.Description, &setting.IsPublic,
				)
			}
		}
	}
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch setting"})
		return
	}
	redactSetting(&setting)

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
	if key == "country_code" {
		profile, err := RegionalProfileForCountry(req.Value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := accounting.ConfigureStoreTimezone(profile.Timezone); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if key == "store_timezone" {
		if _, err := time.LoadLocation(strings.TrimSpace(req.Value)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid IANA timezone %q", req.Value)})
			return
		}
		if err := accounting.ConfigureStoreTimezone(strings.TrimSpace(req.Value)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if isSensitiveSetting(key) && strings.TrimSpace(req.Value) == "********" {
		_ = h.db.QueryRow(`SELECT value FROM settings WHERE key = $1`, key).Scan(&req.Value)
	}
	storedValue := req.Value
	if isSensitiveSetting(key) {
		var encryptErr error
		storedValue, encryptErr = secrets.Encrypt(req.Value)
		if encryptErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": encryptErr.Error()})
			return
		}
	}

	result, err := h.db.Exec(`UPDATE settings SET value = $1, updated_at = CURRENT_TIMESTAMP WHERE key = $2`, storedValue, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		metadata, exists := financialSettingMetadata[key]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
			return
		}
		_, err = h.db.Exec(`INSERT INTO settings (key, value, value_type, category, description, is_public) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP`, key, storedValue, metadata.valueType, metadata.category, metadata.description, metadata.isPublic)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create setting"})
			return
		}
	}

	var setting Setting
	err = h.db.QueryRow(`SELECT key, value, value_type, category, description, is_public FROM settings WHERE key = $1`, key).Scan(
		&setting.Key, &setting.Value, &setting.ValueType, &setting.Category, &setting.Description, &setting.IsPublic,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
		return
	}
	redactSetting(&setting)

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

	result, err := h.db.Exec(`UPDATE settings SET value = $1, updated_at = CURRENT_TIMESTAMP WHERE key = 'tax_rate'`, strconv.FormatFloat(req.TaxRate, 'f', 2, 64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tax rate"})
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "tax_rate setting not found"})
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
