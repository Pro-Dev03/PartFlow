package middleware

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterRejectsRequestsOverBurst(t *testing.T) {
	oldRPS, oldBurst := os.Getenv("RATE_LIMIT_RPS"), os.Getenv("RATE_LIMIT_BURST")
	t.Cleanup(func() {
		_ = os.Setenv("RATE_LIMIT_RPS", oldRPS)
		_ = os.Setenv("RATE_LIMIT_BURST", oldBurst)
	})
	_ = os.Setenv("RATE_LIMIT_RPS", "1")
	_ = os.Setenv("RATE_LIMIT_BURST", "1")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimiter())
	router.GET("/limited", func(c *gin.Context) { c.Status(204) })

	first := httptest.NewRequest("GET", "/limited", nil)
	first.RemoteAddr = "198.51.100.20:1234"
	firstResponse := httptest.NewRecorder()
	router.ServeHTTP(firstResponse, first)
	if firstResponse.Code != 204 {
		t.Fatalf("first request should be allowed, got %d", firstResponse.Code)
	}

	second := httptest.NewRequest("GET", "/limited", nil)
	second.RemoteAddr = "198.51.100.20:5678"
	secondResponse := httptest.NewRecorder()
	router.ServeHTTP(secondResponse, second)
	if secondResponse.Code != 429 {
		t.Fatalf("second request should be rate limited, got %d", secondResponse.Code)
	}
	if secondResponse.Header().Get("Retry-After") == "" {
		t.Fatal("rate-limited response should include Retry-After")
	}
}
