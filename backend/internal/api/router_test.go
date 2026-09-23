package api

import (
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/auth"
	"github.com/partflow/smart-store/internal/localdb"
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
