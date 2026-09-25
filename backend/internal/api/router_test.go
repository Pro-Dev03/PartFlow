package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/auth"
	"github.com/partflow/smart-store/internal/localdb"
	"github.com/partflow/smart-store/pkg/middleware"
)

func TestSetupRoutesRegistersCustomerDebtRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "router-test.db"))

	database, err := localdb.Open()
	if err != nil {
		t.Fatalf("open local database: %v", err)
	}
	defer database.DB.Close()

	db := sqlx.NewDb(database.DB, "sqlite")
	authService, err := auth.NewService(db, "test-secret", false, "", "", "")
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}

	router := gin.New()
	SetupRoutes(router, db, authService)

	paths := make(map[string]bool)
	for _, route := range router.Routes() {
		if route.Method == "GET" && route.Path == "/api/v1/customers/:id/debts" {
			paths[route.Path] = true
		}
		if route.Method == "POST" && route.Path == "/api/v1/customers/:id/debts" {
			paths[route.Path] = true
		}
		if route.Method == "POST" && route.Path == "/api/v1/customers/:id/debt-payments" {
			paths[route.Path] = true
		}
	}

	if !paths["/api/v1/customers/:id/debts"] {
		t.Fatalf("missing GET /api/v1/customers/:id/debts in registered routes")
	}
	if !paths["/api/v1/customers/:id/debt-payments"] {
		t.Fatalf("missing POST /api/v1/customers/:id/debt-payments in registered routes")
	}
}

func TestCloudBusinessAccessUsesAuthenticatedSingleStoreModeWithoutTenantIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PARTFLOW_LOCAL_DB_PATH", filepath.Join(t.TempDir(), "sync-admin-test.db"))
	t.Setenv("DB_CONNECTION_MODE", "cloud")
	t.Setenv("PARTFLOW_ADMIN_EMAILS", "admin@example.test")
	t.Setenv("PARTFLOW_TENANT_RLS_ENABLED", "false")
	middleware.SetDisableAuth(false)
	middleware.SetJWTSecret("sync-admin-test-secret")
	t.Cleanup(func() {
		middleware.SetDisableAuth(false)
		middleware.SetJWTSecret("your-secret-key-change-in-production")
		middleware.SetDatabase(nil)
	})

	database, err := localdb.Open()
	if err != nil {
		t.Fatalf("open local database: %v", err)
	}
	defer database.DB.Close()
	db := sqlx.NewDb(database.DB, "sqlite")
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO users (id, email, password_hash, first_name, last_name, created_at, updated_at, is_active, subscription_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, 'active')`, userID.String(), "subscriber@example.test", "unused", "Regular", "Subscriber", time.Now().UTC(), time.Now().UTC()); err != nil {
		t.Fatalf("insert subscriber: %v", err)
	}
	service, err := auth.NewService(db, "sync-admin-test-secret", false, "", "", "")
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	middleware.SetDatabase(db)
	router := gin.New()
	SetupRoutes(router, db, service)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte("sync-admin-test-secret"))
	if err != nil {
		t.Fatalf("sign subscriber token: %v", err)
	}
	unauthenticated := httptest.NewRecorder()
	router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/api/v1/auth/admin-check", nil))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated cloud request status=%d body=%s; want authentication to remain required", unauthenticated.Code, unauthenticated.Body.String())
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/admin-check", nil)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("authenticated active subscriber status=%d body=%s; want single-store access without tenant-isolation rollout gate", response.Code, response.Body.String())
	}

	// Exercise the actual route tree: subscribers can download and synchronize
	// store data, while the destructive database settings remain administrator-only.
	initialDataRequest := httptest.NewRequest(http.MethodGet, "/api/v1/sync/initial-data", nil)
	initialDataRequest.Header.Set("Authorization", "Bearer "+tokenString)
	initialDataResponse := httptest.NewRecorder()
	router.ServeHTTP(initialDataResponse, initialDataRequest)
	if initialDataResponse.Code != http.StatusOK {
		t.Fatalf("subscriber GET /sync/initial-data status=%d body=%s; want authenticated data download", initialDataResponse.Code, initialDataResponse.Body.String())
	}

	for _, path := range []string{"/api/v1/sync/push", "/api/v1/settings/sync/push"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"operations":[{"id":"legacy-1","entity_type":"customers","entity_id":"customer-1","operation":"upsert","payload":"{}"}]}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenString)
		pushResponse := httptest.NewRecorder()
		router.ServeHTTP(pushResponse, req)
		if pushResponse.Code == http.StatusForbidden && strings.Contains(pushResponse.Body.String(), "ADMIN_REQUIRED") {
			t.Fatalf("subscriber POST %s was blocked by administrator role: %s", path, pushResponse.Body.String())
		}
	}

	deleteDatabaseRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/settings/database?confirmation_token=DELETE%20ALL%20DATA", nil)
	deleteDatabaseRequest.Header.Set("Authorization", "Bearer "+tokenString)
	deleteDatabaseResponse := httptest.NewRecorder()
	router.ServeHTTP(deleteDatabaseResponse, deleteDatabaseRequest)
	if deleteDatabaseResponse.Code != http.StatusForbidden || !strings.Contains(deleteDatabaseResponse.Body.String(), "ADMIN_REQUIRED") {
		t.Fatalf("subscriber database deletion status=%d body=%s; want destructive database settings to remain administrator-only", deleteDatabaseResponse.Code, deleteDatabaseResponse.Body.String())
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutRequest.Header.Set("Authorization", "Bearer "+tokenString)
	logoutResponse := httptest.NewRecorder()
	router.ServeHTTP(logoutResponse, logoutRequest)
	if logoutResponse.Code != http.StatusOK {
		t.Fatalf("authenticated logout status=%d body=%s; want session revocation to remain available before tenant migration", logoutResponse.Code, logoutResponse.Body.String())
	}

	reusedTokenRequest := httptest.NewRequest(http.MethodGet, "/api/v1/auth/admin-check", nil)
	reusedTokenRequest.Header.Set("Authorization", "Bearer "+tokenString)
	reusedTokenResponse := httptest.NewRecorder()
	router.ServeHTTP(reusedTokenResponse, reusedTokenRequest)
	if reusedTokenResponse.Code != http.StatusUnauthorized || !strings.Contains(reusedTokenResponse.Body.String(), "SESSION_REVOKED") {
		t.Fatalf("access token reused after logout status=%d body=%s; want immediate session revocation", reusedTokenResponse.Code, reusedTokenResponse.Body.String())
	}
}
