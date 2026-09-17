package settings

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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

func TestPersistedStoreTimezoneTakesPriorityOverCountry(t *testing.T) {
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
		t.Fatalf("response did not preserve store timezone: %s", request.Body.String())
	}
}
