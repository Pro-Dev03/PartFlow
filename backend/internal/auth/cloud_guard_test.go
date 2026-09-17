package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCloudGuardRequiresCloudTokenInLocalModeByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("SERVER_MODE", "debug")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "")

	router := gin.New()
	router.Use(CloudGuard(&Service{}))
	router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	if response.Code != http.StatusForbidden {
		t.Fatalf("missing cloud token status = %d, want %d", response.Code, http.StatusForbidden)
	}
}
