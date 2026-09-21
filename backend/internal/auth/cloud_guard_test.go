package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCloudGuardAllowsLocalOperationByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("SERVER_MODE", "debug")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "false")

	router := gin.New()
	router.Use(CloudGuard(&Service{}))
	router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("local operation status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestCloudGuardRequiresCloudTokenWhenExplicitlyEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")

	router := gin.New()
	router.Use(CloudGuard(&Service{}))
	router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/protected", nil))
	if response.Code != http.StatusForbidden {
		t.Fatalf("missing cloud token status = %d, want %d", response.Code, http.StatusForbidden)
	}
}
