package settings

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/partflow/smart-store/internal/accounting"
	_ "modernc.org/sqlite"
)

func TestRegionalProfileCatalogUsesIANATimezones(t *testing.T) {
	profiles := RegionalProfiles()
	if len(profiles) < 6 {
		t.Fatalf("catalog contains %d profiles, want at least 6", len(profiles))
	}
	for _, profile := range profiles {
		if _, err := time.LoadLocation(profile.Timezone); err != nil {
			t.Fatalf("%s timezone %q: %v", profile.CountryCode, profile.Timezone, err)
		}
		if profile.CountryCode == "" || profile.CountryName == "" || profile.Currency == "" || profile.Locale == "" {
			t.Fatalf("incomplete regional profile: %+v", profile)
		}
	}
}

func TestRegionalProfileForCountryRejectsUnknownCode(t *testing.T) {
	if _, err := RegionalProfileForCountry("XX"); err == nil {
		t.Fatal("expected unsupported country code error")
	}
}

func TestRegionalSettingsReturnsPersistedStoreTimezone(t *testing.T) {
	originalTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(originalTimezone) })
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT, value_type TEXT, category TEXT, description TEXT, is_public INTEGER, updated_at TEXT); INSERT INTO settings (key, value) VALUES ('country_code', 'IL'), ('store_timezone', 'America/New_York');`); err != nil {
		t.Fatal(err)
	}

	handler := NewHandler(db)
	request := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(request)
	handler.GetRegionalSettings(context)

	if request.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", request.Code)
	}
	if !strings.Contains(request.Body.String(), `"store_timezone":"America/New_York"`) {
		t.Fatalf("response did not use the persisted store timezone: %s", request.Body.String())
	}
}

func TestUpdateRegionalSettingsPersistsAndAppliesCustomTimezone(t *testing.T) {
	originalTimezone := accounting.CurrentStoreTimezone()
	t.Cleanup(func() { _ = accounting.ConfigureStoreTimezone(originalTimezone) })
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE settings (
		key TEXT PRIMARY KEY, value TEXT NOT NULL, value_type TEXT, category TEXT,
		description TEXT, is_public INTEGER, created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP
	); INSERT INTO settings (key, value) VALUES ('country_code', 'IL'), ('store_timezone', 'Asia/Jerusalem');`); err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(db)
	request := httptest.NewRequest(http.MethodPut, "/settings/regional", strings.NewReader(`{"country_code":"SA","timezone":"America/New_York"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = request
	handler.UpdateRegionalSettings(context)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var countryCode, timezone string
	if err := db.QueryRow(`SELECT value FROM settings WHERE key='country_code'`).Scan(&countryCode); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT value FROM settings WHERE key='store_timezone'`).Scan(&timezone); err != nil {
		t.Fatal(err)
	}
	if countryCode != "SA" || timezone != "America/New_York" {
		t.Fatalf("persisted country/timezone = %s/%s, want SA/America/New_York", countryCode, timezone)
	}
	if got := accounting.CurrentStoreTimezone(); got != timezone {
		t.Fatalf("active timezone = %s, want %s", got, timezone)
	}
}

func TestRegionalCountryProfilesUseTheirOwnTimezones(t *testing.T) {
	want := map[string]string{
		"PS": "Asia/Hebron",
		"IL": "Asia/Jerusalem",
		"SA": "Asia/Riyadh",
		"AE": "Asia/Dubai",
		"JO": "Asia/Amman",
		"EG": "Africa/Cairo",
	}
	for code, timezone := range want {
		profile, err := RegionalProfileForCountry(code)
		if err != nil {
			t.Fatalf("country %s: %v", code, err)
		}
		if profile.Timezone != timezone {
			t.Errorf("country %s timezone = %s, want %s", code, profile.Timezone, timezone)
		}
	}
}
