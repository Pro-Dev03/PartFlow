package reports

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/auth"
	"github.com/partflow/smart-store/pkg/middleware"
	_ "modernc.org/sqlite"
)

func TestReportEndpointRejectsUnauthenticatedRequestWith401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(auth.NewJWTService("report-status-test-secret", time.Hour, time.Hour), nil))
	RegisterRoutes(protected, nil)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/reports/sales", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated report request status = %d, want 401; body=%s", response.Code, response.Body.String())
	}
}

func TestReportEndpointReturns500ForBackendQueryFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	router := gin.New()
	db := sqlx.NewDb(database, "sqlite")
	RegisterRoutes(router.Group("/api/v1"), db)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/reports/sales?start_date=2026-09-16&end_date=2026-09-16", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("backend query failure status = %d, want 500; body=%s", response.Code, response.Body.String())
	}
}
