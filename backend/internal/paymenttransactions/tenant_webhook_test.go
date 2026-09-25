package paymenttransactions

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func TestWebhookFailsClosedUntilTenantRoutingIsConfigured(t *testing.T) {
	t.Setenv("PARTFLOW_TENANT_RLS_ENABLED", "true")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/payment-webhooks/:provider", NewHandler(nil).Webhook)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/payment-webhooks/cardcom", strings.NewReader(`{}`)))

	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "PAYMENT_WEBHOOK_TENANT_ROUTING_REQUIRED") {
		t.Fatalf("webhook did not fail closed without tenant scope: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCloudWebhookFailsClosedBeforeTenantMigration(t *testing.T) {
	t.Setenv("PARTFLOW_TENANT_RLS_ENABLED", "false")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/payment-webhooks/:provider", NewHandler(sqlx.NewDb(nil, "pgx")).Webhook)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/payment-webhooks/cardcom", strings.NewReader(`{}`)))

	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "PAYMENT_WEBHOOK_TENANT_ROUTING_REQUIRED") {
		t.Fatalf("unscoped cloud webhook did not fail closed before migration: status=%d body=%s", response.Code, response.Body.String())
	}
}
