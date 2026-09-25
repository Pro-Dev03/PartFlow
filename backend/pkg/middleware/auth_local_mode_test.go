package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestAuthRequiresCloudWhenLocalSQLiteHasNoUserRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use a SQLite DB without the users table to simulate the local sync-only database.
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("connect sqlite: %v", err)
	}
	defer db.Close()

	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "false")
	SetJWTSecret("test-secret")
	SetDatabase(db)

	userID := uuid.NewString()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID})
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	Auth()(c)

	if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "CLOUD_CONNECTION_REQUIRED") {
		t.Fatalf("local JWT without live cloud validation was accepted: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuthCannotDisableCloudRequirementInLocalMode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "false")
	SetDisableAuth(false)

	localToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": uuid.NewString()})
	localTokenString, err := localToken.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+localTokenString)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	Auth()(c)

	if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "CLOUD_CONNECTION_REQUIRED") {
		t.Fatalf("local auth opt-out bypassed cloud check: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuthUsesCloudTokenHeaderWhenLocalJWTIsPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	var receivedAuth string
	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"valid":true,"user":{"id":"%s","email":"owner@example.test"}}}`, userID)
	}))
	defer cloud.Close()

	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	t.Setenv("PARTFLOW_CLOUD_API_URL", cloud.URL)
	SetDisableAuth(false)

	localToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID.String()})
	localTokenString, err := localToken.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(Auth())
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+localTokenString)
	req.Header.Set("X-PartFlow-Cloud-Token", "cloud-access-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected cloud header to authorize the local request, status=%d body=%s", w.Code, w.Body.String())
	}
	if receivedAuth != "Bearer cloud-access-token" {
		t.Fatalf("cloud validate received %q, want the cloud token header", receivedAuth)
	}
}

func TestAuthRejectsLocalJWTWhenCloudHeaderMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("cloud validate should not be called with the local JWT")
	}))
	defer cloud.Close()

	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	t.Setenv("PARTFLOW_CLOUD_API_URL", cloud.URL)
	SetDisableAuth(false)

	localToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": uuid.NewString()})
	localTokenString, err := localToken.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(Auth())
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+localTokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected missing cloud header to be rejected outside a configured grace grant, status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuthDisableFlagCannotBypassRequiredCloudSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("SERVER_MODE", "debug")
	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	SetDisableAuth(true)
	t.Cleanup(func() { SetDisableAuth(false) })

	router := gin.New()
	router.Use(Auth())
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "AUTH_DISABLED") {
		t.Fatalf("disabled auth bypassed required cloud check: status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestLocalJWTReadsLegacySubjectClaim(t *testing.T) {
	SetJWTSecret("test-secret")
	userID := uuid.New()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": userID.String()})
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	parsedID, valid := localJWTUserID(tokenString)
	if !valid || parsedID != userID {
		t.Fatalf("legacy subject id=%s valid=%t", parsedID, valid)
	}
}
