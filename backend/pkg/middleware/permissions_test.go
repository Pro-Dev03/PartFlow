package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequirePermissionAllowsGrantedPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/archive/events", func(c *gin.Context) {
		c.Set("permissions", []string{"archive.read"})
		c.Next()
	}, RequirePermission("archive.read"), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/archive/events", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != 200 {
		t.Fatalf("permission holder status = %d, want 200", resp.Code)
	}
}

func TestRequirePermissionRejectsMissingPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/archive/events", RequirePermission("archive.read"), func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/archive/events", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != 403 {
		t.Fatalf("permission-less user status = %d, want 403", resp.Code)
	}
}
