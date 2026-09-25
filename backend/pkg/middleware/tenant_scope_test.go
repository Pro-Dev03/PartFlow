package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestTenantScopeFailsClosedForLocalSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PARTFLOW_TENANT_RLS_ENABLED", "true")
	t.Setenv("DB_CONNECTION_MODE", "local")
	previousDB := db
	t.Cleanup(func() { db = previousDB })
	SetDatabase(nil)

	called := false
	router := gin.New()
	router.Use(TenantScope())
	router.GET("/business", func(c *gin.Context) {
		called = true
		c.Status(http.StatusNoContent)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/business", nil))
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "TENANT_DATABASE_REQUIRED") {
		t.Fatalf("tenant RLS accepted a local SQLite request: status=%d body=%s", response.Code, response.Body.String())
	}
	if called {
		t.Fatal("business handler ran without a tenant-capable database")
	}
}

func TestTenantScopeIsNoopUntilCloudMigrationIsEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PARTFLOW_TENANT_RLS_ENABLED", "false")
	t.Setenv("DB_CONNECTION_MODE", "local")

	router := gin.New()
	router.Use(TenantScope())
	router.GET("/business", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/business", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("disabled tenant RLS changed local request behavior: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTenantScopeAllowsSingleStoreCloudRequestsWhenRLSDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PARTFLOW_TENANT_RLS_ENABLED", "false")
	t.Setenv("DB_CONNECTION_MODE", "cloud")
	previousDB := db
	t.Cleanup(func() { db = previousDB })
	SetDatabase(nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uuid.New())
		c.Next()
	})
	router.Use(TenantScope())
	router.GET("/business", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/business", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("single-store cloud request was blocked while tenant RLS is disabled: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTenantIDFromContextOnlyReturnsNonZeroUUID(t *testing.T) {
	tenantID := uuid.New()
	ctx := context.WithValue(context.Background(), tenantContextKey{}, tenantID)
	got, ok := TenantIDFromContext(ctx)
	if !ok || got != tenantID {
		t.Fatalf("tenant from context=%s ok=%t; want %s", got, ok, tenantID)
	}
	if _, ok := TenantIDFromContext(context.Background()); ok {
		t.Fatal("missing tenant context was accepted")
	}
}
