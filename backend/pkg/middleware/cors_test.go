package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.test")
	router := gin.New()
	router.Use(CORS())
	router.GET("/health", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "https://app.example.test")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.test" {
		t.Fatalf("expected allowed origin header, got %q", got)
	}
}

func TestCORSRejectsUnconfiguredPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.test")
	router := gin.New()
	router.Use(CORS())
	router.OPTIONS("/health", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "https://evil.example.test")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("expected disallowed preflight to return 403, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS allow header for disallowed origin, got %q", got)
	}
}
