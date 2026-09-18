package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPublicRegistrationRemainsDisabledWhenEnvironmentEnablesIt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PARTFLOW_ALLOW_PUBLIC_REGISTRATION", "true")

	router := gin.New()
	handler := NewHandler(nil, nil)
	router.POST("/auth/register", handler.Register)

	request := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("registration status = %d, want %d", response.Code, http.StatusForbidden)
	}
}
