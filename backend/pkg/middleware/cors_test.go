package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestCORSAllowsCloudSessionHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5174")
	router := gin.New()
	router.Use(CORS())
	router.OPTIONS("/auth/admin-check", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest("OPTIONS", "/auth/admin-check", nil)
	req.Header.Set("Origin", "http://localhost:5174")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "X-PartFlow-Cloud-Token, Authorization")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	got := w.Header().Get("Access-Control-Allow-Headers")
	if !strings.Contains(got, "X-PartFlow-Cloud-Token") {
		t.Fatalf("expected X-PartFlow-Cloud-Token header to be allowed, got %q", got)
	}
	if !strings.Contains(got, "Idempotency-Key") {
		t.Fatalf("expected idempotency header to be allowed, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5174" {
		t.Fatalf("expected allowed origin, got %q", got)
	}
}

func TestCORSAllowsCloudAPIURLHeader(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5174")
	router := gin.New()
	router.Use(CORS())
	router.POST("/sync", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodOptions, "/sync", nil)
	req.Header.Set("Origin", "http://localhost:5174")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "x-partflow-cloud-api-url")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected successful preflight, got %d", response.Code)
	}
	if !strings.Contains(strings.ToLower(response.Header().Get("Access-Control-Allow-Headers")), "x-partflow-cloud-api-url") {
		t.Fatalf("expected cloud API URL header to be allowed, got %q", response.Header().Get("Access-Control-Allow-Headers"))
	}
}
