package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestLocalAuthRequiresLiveCloudValidationForEveryRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db
	t.Cleanup(func() { SetDatabase(previousDB) })
	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "false") // Local mode must ignore this opt-out.
	SetDisableAuth(false)
	SetJWTSecret("strict-cloud-auth-test-secret")

	localDB, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer localDB.Close()
	SetDatabase(localDB)
	userID := uuid.New()
	var outage atomic.Bool
	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if outage.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprint(w, `{"error":"cloud unavailable"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"data":{"valid":true,"user":{"id":%q,"email":"owner@example.test"}}}`, userID.String())
	}))
	defer cloud.Close()
	t.Setenv("PARTFLOW_CLOUD_API_URL", cloud.URL)
	localToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID.String(), "exp": time.Now().Add(time.Hour).Unix()})
	localTokenString, err := localToken.SignedString([]byte("strict-cloud-auth-test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	request := func(cloudToken string) *httptest.ResponseRecorder {
		router := gin.New()
		router.Use(Auth())
		router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+localTokenString)
		if cloudToken != "" {
			req.Header.Set("X-PartFlow-Cloud-Token", cloudToken)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}

	if response := request("live-cloud-token"); response.Code != http.StatusOK {
		t.Fatalf("live validation status=%d body=%s", response.Code, response.Body.String())
	}
	outage.Store(true)
	if response := request("live-cloud-token"); response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "CLOUD_CONNECTION_REQUIRED") {
		t.Fatalf("cloud outage must block protected request: status=%d body=%s", response.Code, response.Body.String())
	}
	if response := request(""); response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "CLOUD_CONNECTION_REQUIRED") {
		t.Fatalf("missing cloud token must block protected request: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestLegacyCachedOfflineGrantCannotAuthorizeLocalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db
	t.Cleanup(func() { SetDatabase(previousDB) })
	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	t.Setenv("PARTFLOW_CLOUD_API_URL", "http://127.0.0.1:1/api/v1")
	SetDisableAuth(false)
	SetJWTSecret("legacy-offline-grant-test-secret")
	localDB, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer localDB.Close()
	SetDatabase(localDB)
	userID := uuid.New()
	if _, err := localDB.Exec(`CREATE TABLE cloud_auth_grants (user_id TEXT PRIMARY KEY, grant_token TEXT NOT NULL, updated_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := localDB.Exec(`INSERT INTO cloud_auth_grants (user_id, grant_token, updated_at) VALUES (?, ?, ?)`, userID.String(), "legacy-grant", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	localToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID.String(), "exp": time.Now().Add(time.Hour).Unix()})
	token, err := localToken.SignedString([]byte("legacy-offline-grant-test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(Auth())
	router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-PartFlow-Cloud-Token", "cloud-token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "CLOUD_CONNECTION_REQUIRED") {
		t.Fatalf("legacy offline grant authorized request: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestCloudAuthRejectsSuspendedAccountOnEveryRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db
	t.Cleanup(func() { SetDatabase(previousDB) })
	t.Setenv("DB_CONNECTION_MODE", "cloud")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	SetDisableAuth(false)
	SetJWTSecret("cloud-suspension-test-secret")
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer rawDB.Close()
	if _, err := rawDB.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, is_active INTEGER NOT NULL, subscription_status TEXT, subscription_expires_at TEXT, updated_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	userID := uuid.New()
	updatedAt := time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)
	if _, err := rawDB.Exec(`INSERT INTO users (id, is_active, subscription_status, updated_at) VALUES (?, 1, 'suspended', ?)`, userID.String(), updatedAt); err != nil {
		t.Fatal(err)
	}
	SetDatabase(sqlx.NewDb(rawDB, "sqlite"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID.String(), "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix()})
	tokenString, err := token.SignedString([]byte("cloud-suspension-test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(Auth())
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "SUBSCRIPTION_SUSPENDED") {
		t.Fatalf("suspended cloud request status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestCloudAuthTreatsDatabaseFailureAsTemporary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db
	t.Cleanup(func() { SetDatabase(previousDB) })
	t.Setenv("DB_CONNECTION_MODE", "cloud")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	SetDisableAuth(false)
	SetJWTSecret("cloud-database-failure-test-secret")
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := rawDB.Close(); err != nil {
		t.Fatal(err)
	}
	SetDatabase(sqlx.NewDb(rawDB, "sqlite"))
	userID := uuid.New()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID.String(), "exp": time.Now().Add(time.Hour).Unix()})
	tokenString, err := token.SignedString([]byte("cloud-database-failure-test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(Auth())
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "AUTH_SERVICE_UNAVAILABLE") {
		t.Fatalf("database failure status=%d body=%s", w.Code, w.Body.String())
	}
}
