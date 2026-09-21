package auth

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestRefreshTokenCookieUsesCrossSitePolicyForTLS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/cookie", func(c *gin.Context) {
		setRefreshTokenCookie(c, "refresh", refreshTokenCookieAge)
	})

	request := httptest.NewRequest(http.MethodGet, "https://api.example.test/cookie", nil)
	request.TLS = &tls.ConnectionState{}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	cookie := response.Header().Get("Set-Cookie")
	if !strings.Contains(cookie, "HttpOnly") || !strings.Contains(cookie, "Secure") || !strings.Contains(cookie, "SameSite=None") {
		t.Fatalf("secure refresh cookie = %q, want HttpOnly, Secure, SameSite=None", cookie)
	}
}

func TestRefreshTokenUsesCookieFallbackWhenBodyIsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service, _, _ := newRefreshTokenTestService(t)
	router := gin.New()
	handler := NewHandler(service, nil)
	router.POST("/auth/refresh", handler.RefreshToken)

	login, err := service.Login(t.Context(), &LoginRequest{Email: "refresh@example.test", Password: "TestOwnerPassword123!"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: login.RefreshToken, Path: "/api/v1/auth"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d, body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "access_token") {
		t.Fatalf("refresh response = %q, want access_token", response.Body.String())
	}
}
