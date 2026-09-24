package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestIdempotencyStoreReadFailureDoesNotRunHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := sqlx.Open("sqlite", filepath.Join(t.TempDir(), "idempotency.db"))
	if err != nil {
		t.Fatalf("open SQLite test database: %v", err)
	}
	defer db.Close()

	middleware := NewIdempotencyMiddleware(db)
	if middleware.initErr != nil {
		t.Fatalf("initialize idempotency store: %v", middleware.initErr)
	}
	if _, err := db.Exec(`DROP TABLE idempotency_keys`); err != nil {
		t.Fatalf("drop idempotency table: %v", err)
	}

	handlerCalls := 0
	router := gin.New()
	router.POST("/sales", middleware.Idempotency(), func(c *gin.Context) {
		handlerCalls++
		c.Status(http.StatusCreated)
	})

	request := httptest.NewRequest(http.MethodPost, "/sales", strings.NewReader(`{"total": 10}`))
	request.Header.Set("Idempotency-Key", "sale-retry-key")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when the idempotency store cannot be read, got %d", response.Code)
	}
	if handlerCalls != 0 {
		t.Fatalf("sale handler ran %d times while idempotency could not be verified", handlerCalls)
	}
}

func TestIdempotencyKeyIsRequiredBeforeCreatingSale(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := sqlx.Open("sqlite", filepath.Join(t.TempDir(), "idempotency.db"))
	if err != nil {
		t.Fatalf("open SQLite test database: %v", err)
	}
	defer db.Close()

	middleware := NewIdempotencyMiddleware(db)
	if middleware.initErr != nil {
		t.Fatalf("initialize idempotency store: %v", middleware.initErr)
	}
	handlerCalls := 0
	router := gin.New()
	router.POST("/sales", middleware.Idempotency(), func(c *gin.Context) {
		handlerCalls++
		c.Status(http.StatusCreated)
	})

	request := httptest.NewRequest(http.MethodPost, "/sales", strings.NewReader(`{"total": 10}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without an idempotency key, got %d", response.Code)
	}
	if handlerCalls != 0 {
		t.Fatalf("sale handler ran %d times without an idempotency key", handlerCalls)
	}
}
