package settings

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	Value *string `json:"value"`
}

var settingMetadata = map[string]struct {
	defaultValue string
	valueType    string
	category     string
	description  string
	isPublic     bool
}{
	"currency":                     {"ILS", "string", "general", "العملة الافتراضية", true},
	"store_name":                   {"PartFlow Store", "string", "general", "اسم المتجر", true},
	"discounts_enabled":            {"true", "boolean", "financial", "السماح بالخصومات", false},
	"tax_rate":                     {"0", "number", "financial", "نسبة الضريبة المئوية", true},
	"max_discount_rate":            {"15", "number", "financial", "الحد الأقصى للخصم المئوي", false},
	"default_profit_margin":        {"30", "number", "financial", "نسبة الربح المقترحة عند إضافة منتج", false},
	"country_code":                 {"IL", "string", "regional", "الدولة الافتراضية للمتجر", true},
	"store_timezone":               {accounting.DefaultStoreTimezone, "string", "regional", "المنطقة الزمنية للمتجر", true},
	"pos_products_per_page":        {"12", "number", "appearance", "عدد منتجات نقطة البيع في الصفحة", false},
	"pos_product_view_mode":        {"cards", "string", "appearance", "طريقة عرض منتجات نقطة البيع", false},
	"electronic_payments_enabled":  {"false", "boolean", "payments", "تفعيل الدفع الإلكتروني", false},
	"payment_provider":             {"manual", "string", "payments", "مزود الدفع الإلكتروني", false},
	"payment_environment":          {"test", "string", "payments", "بيئة الدفع الإلكتروني", false},
	"payment_public_key":           {"", "string", "payments", "المفتاح العام لمزود الدفع", false},
	"payment_secret_key":           {"", "string", "payments", "المفتاح السري لمزود الدفع", false},
	"payment_merchant_id":          {"", "string", "payments", "معرف التاجر لدى مزود الدفع", false},
	"payment_terminal_id":          {"", "string", "payments", "معرف جهاز الدفع", false},
	"payment_webhook_url":          {"", "string", "payments", "عنوان Webhook للدفع", false},
	"payment_webhook_secret":       {"", "string", "payments", "سر توقيع Webhook", false},
	"payment_methods":              {"[\"card\"]", "json", "payments", "طرق الدفع الإلكتروني المفعلة", false},
	"installment_whatsapp_number":  {"", "string", "payments", "رقم واتساب وكيل التقسيط", false},
	"installment_whatsapp_message": {"*طلب تقسيط جديد - {store_name}*\n\nالسلام عليكم،\nنرجو متابعة طلب التقسيط التالي:\n\n*اسم العميل:* {customer_name}\n*إجمالي الفاتورة:* ₪{total}\n*مدة التقسيط:* {months} أشهر\n*قيمة القسط التقريبية:* ₪{installment}\n\nيرجى تأكيد تسجيل الطلب ومتابعته.\n\nمع التحية،\n{store_name}", "string", "payments", "قالب رسالة واتساب للتقسيط", false},
}

func isSensitiveSetting(key string) bool {
	return key == "payment_secret_key" || key == "payment_webhook_secret"
}

func redactSetting(setting *Setting) {
	if isSensitiveSetting(setting.Key) && setting.Value != "" {
		setting.Value = "********"
	}
}

func validateSettingValue(key, value string, metadata struct {
	defaultValue string
	valueType    string
	category     string
	description  string
	isPublic     bool
}) error {
	trimmed := strings.TrimSpace(value)
	switch metadata.valueType {
	case "boolean":
		if _, err := strconv.ParseBool(trimmed); err != nil {
			return fmt.Errorf("setting %s must be true or false", key)
		}
	case "number":
		number, err := strconv.ParseFloat(trimmed, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return fmt.Errorf("setting %s must be a valid number", key)
		}
		switch key {
		case "tax_rate", "max_discount_rate":
			if number < 0 || number > 100 {
				return fmt.Errorf("setting %s must be between 0 and 100", key)
			}
		case "default_profit_margin":
			if number < 0 || number > 1000 {
				return fmt.Errorf("setting %s must be between 0 and 1000", key)
			}
		case "pos_products_per_page":
			if number < 1 || number > 100 || number != math.Trunc(number) {
				return fmt.Errorf("setting %s must be a whole number between 1 and 100", key)
			}
		}
	case "json":
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("setting %s must contain valid JSON", key)
		}
	}
	if key == "pos_product_view_mode" && trimmed != "cards" && trimmed != "list" {
		return fmt.Errorf("unsupported POS product view mode")
	}
	if key == "payment_environment" && trimmed != "test" && trimmed != "live" {
		return fmt.Errorf("payment environment must be test or live")
	}
	return nil
}

func NewHandler(db *sql.DB) *Handler {
	ensureDefaultSettings(db)
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

func ensureDefaultSettings(db *sql.DB) {
	for key, metadata := range settingMetadata {
		_, _ = insertSetting(db, key, metadata.defaultValue, metadata.valueType, metadata.category, metadata.description, metadata.isPublic, false)
	}
}

func insertSetting(db *sql.DB, key, value, valueType, category, description string, isPublic, updateExisting bool) (sql.Result, error) {
	conflictAction := "DO NOTHING"
	if updateExisting {
		conflictAction = "DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP"
	}
	return db.Exec(`
		INSERT INTO settings (id, key, value, value_type, category, description, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (key) `+conflictAction,
		uuid.NewString(), key, value, valueType, category, description, isPublic)
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
	var countryCode string
	if h.db.QueryRow(`SELECT value FROM settings WHERE key = 'country_code'`).Scan(&countryCode) == nil {
		if selected, profileErr := RegionalProfileForCountry(countryCode); profileErr == nil {
			profile = selected
		}
	}
	if err := h.db.QueryRow(`SELECT value FROM settings WHERE key = 'store_timezone'`).Scan(&storeTimezone); err == nil {
		if validated, validateErr := accounting.ValidateStoreTimezone(storeTimezone); validateErr == nil {
			profile.Timezone = validated
			timezonePersisted = true
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
	if strings.TrimSpace(request.CountryCode) != "" {
		var err error
		profile, err = RegionalProfileForCountry(strings.TrimSpace(request.CountryCode))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		var currentCountry string
		if err := h.db.QueryRow(`SELECT value FROM settings WHERE key = 'country_code'`).Scan(&currentCountry); err == nil {
			if current, profileErr := RegionalProfileForCountry(currentCountry); profileErr == nil {
				profile = current
			}
		}
	}
	timezone := strings.TrimSpace(request.Timezone)
	if timezone == "" {
		timezone = profile.Timezone
	}
	validatedTimezone, err := accounting.ValidateStoreTimezone(timezone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile.Timezone = validatedTimezone
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update regional settings"})
		return
	}
	if profile.CountryCode != "" {
		if _, err := tx.Exec(`INSERT INTO settings (key, value, value_type, category, description, is_public) VALUES ($1, $2, 'string', 'regional', $3, true) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP`, "country_code", profile.CountryCode, "الدولة الافتراضية للمتجر"); err != nil {
			_ = tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update regional settings"})
			return
		}
	}
	if _, err := tx.Exec(`INSERT INTO settings (key, value, value_type, category, description, is_public) VALUES ($1, $2, 'string', 'regional', $3, true) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP`, "store_timezone", profile.Timezone, "المنطقة الزمنية للمتجر"); err != nil {
		_ = tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update regional settings"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update regional settings"})
		return
	}
	if err := accounting.ConfigureStoreTimezone(validatedTimezone); err != nil {
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
	timezone, err := accounting.ValidateStoreTimezone(request.Timezone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing string
	if err := h.db.QueryRow(`SELECT value FROM settings WHERE key = 'store_timezone'`).Scan(&existing); err == nil && strings.TrimSpace(existing) != "" {
		if configureErr := accounting.ConfigureStoreTimezone(existing); configureErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Stored store timezone is invalid"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"store_timezone": existing, "timezone_persisted": true}})
		return
	}
	if _, err := h.db.Exec(`INSERT INTO settings (key, value, value_type, category, description, is_public, created_at, updated_at) VALUES ($1, $2, 'string', 'regional', $3, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON CONFLICT (key) DO NOTHING`, "store_timezone", timezone, "المنطقة الزمنية للمتجر"); err != nil {
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
		if metadata, exists := settingMetadata[key]; exists {
			if _, insertErr := insertSetting(h.db, key, metadata.defaultValue, metadata.valueType, metadata.category, metadata.description, metadata.isPublic, false); insertErr == nil {
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
	metadata, knownSetting := settingMetadata[key]
	if !knownSetting {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Value == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	value := *req.Value
	preserveSensitive := isSensitiveSetting(key) && strings.TrimSpace(value) == "********"
	if !preserveSensitive {
		if err := validateSettingValue(key, value, metadata); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	var timezoneToConfigure string
	if key == "country_code" {
		profile, err := RegionalProfileForCountry(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if _, err := accounting.ValidateStoreTimezone(profile.Timezone); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		timezoneToConfigure = profile.Timezone
	}
	if key == "store_timezone" {
		validated, err := accounting.ValidateStoreTimezone(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		timezoneToConfigure = validated
	}
	storedValue := value
	if preserveSensitive {
		if err := h.db.QueryRow(`SELECT value FROM settings WHERE key = $1`, key).Scan(&storedValue); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
			return
		}
	} else if isSensitiveSetting(key) {
		var encryptErr error
		storedValue, encryptErr = secrets.Encrypt(value)
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
		_, err = insertSetting(h.db, key, storedValue, metadata.valueType, metadata.category, metadata.description, metadata.isPublic, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create setting"})
			return
		}
	}
	if key == "country_code" {
		if timezoneMetadata, exists := settingMetadata["store_timezone"]; exists {
			_, err = insertSetting(h.db, "store_timezone", timezoneToConfigure, timezoneMetadata.valueType, timezoneMetadata.category, timezoneMetadata.description, timezoneMetadata.isPublic, true)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update store timezone"})
				return
			}
		}
	}
	if timezoneToConfigure != "" {
		if err := accounting.ConfigureStoreTimezone(timezoneToConfigure); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure store timezone"})
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
		TaxRate *float64 `json:"tax_rate" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.TaxRate == nil || math.IsNaN(*req.TaxRate) || math.IsInf(*req.TaxRate, 0) || *req.TaxRate < 0 || *req.TaxRate > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tax rate value"})
		return
	}

	result, err := h.db.Exec(`UPDATE settings SET value = $1, updated_at = CURRENT_TIMESTAMP WHERE key = 'tax_rate'`, strconv.FormatFloat(*req.TaxRate, 'f', 2, 64))
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
			"tax_rate": *req.TaxRate,
		},
		"message": "Tax rate updated successfully",
	})
}
