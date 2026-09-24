package middleware

import (
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
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
	"github.com/partflow/smart-store/pkg/offlinegrant"
	_ "modernc.org/sqlite"
)

func TestLocalAuthUsesSignedGraceDuringCloudOutageAndRejectsSuspension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db
	t.Cleanup(func() { SetDatabase(previousDB) })

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PARTFLOW_OFFLINE_GRANT_PRIVATE_KEY", base64.StdEncoding.EncodeToString(privateKey))
	t.Setenv("PARTFLOW_OFFLINE_GRANT_PUBLIC_KEY", base64.StdEncoding.EncodeToString(publicKey))
	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	SetJWTSecret("offline-grant-test-secret")

	localDB, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer localDB.Close()
	SetDatabase(localDB)

	userID := uuid.New()
	grant, err := offlinegrant.Issue(userID.String(), nil, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	var suspended atomic.Bool
	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if suspended.Load() {
			w.WriteHeader(http.StatusForbidden)
			_, _ = fmt.Fprint(w, `{"code":"SUBSCRIPTION_SUSPENDED","error":"suspended"}`)
			return
		}
		if r.URL.Path == "/auth/validate" && r.Method == http.MethodPost {
			if r.Header.Get("Authorization") != "Bearer cloud-access-token" {
				t.Errorf("Authorization=%q", r.Header.Get("Authorization"))
			}
			_, _ = fmt.Fprintf(w, `{"data":{"valid":true,"offline_grant":%q,"user":{"id":%q,"email":"owner@example.test"}}}`, grant, userID.String())
			return
		}
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer cloud.Close()
	t.Setenv("PARTFLOW_CLOUD_API_URL", cloud.URL)

	localToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	localTokenString, err := localToken.SignedString([]byte("offline-grant-test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	request := func() *httptest.ResponseRecorder {
		router := gin.New()
		router.Use(Auth())
		router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+localTokenString)
		req.Header.Set("X-PartFlow-Cloud-Token", "cloud-access-token")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}

	if response := request(); response.Code != http.StatusOK {
		t.Fatalf("online validation status=%d body=%s", response.Code, response.Body.String())
	}
	var stored string
	if err := localDB.Get(&stored, `SELECT grant_token FROM cloud_auth_grants WHERE user_id = $1`, userID.String()); err != nil || stored != grant {
		t.Fatalf("stored grant=%q err=%v", stored, err)
	}

	cloud.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, `{"error":"temporary cloud outage"}`)
	})
	if response := request(); response.Code != http.StatusOK {
		t.Fatalf("valid grace grant should permit local request during outage, status=%d body=%s", response.Code, response.Body.String())
	}

	cloud.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = fmt.Fprint(w, `{"code":"SUBSCRIPTION_SUSPENDED","error":"suspended"}`)
	})
	if response := request(); response.Code != http.StatusForbidden {
		t.Fatalf("explicit suspension must override a cached grant, status=%d body=%s", response.Code, response.Body.String())
	}
	if err := localDB.Get(&stored, `SELECT grant_token FROM cloud_auth_grants WHERE user_id = $1`, userID.String()); err == nil {
		t.Fatal("explicit suspension should clear the cached offline grant")
	}
}

func TestLocalAuthRejectsExpiredOfflineGrant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db
	t.Cleanup(func() { SetDatabase(previousDB) })
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PARTFLOW_OFFLINE_GRANT_PRIVATE_KEY", base64.StdEncoding.EncodeToString(privateKey))
	t.Setenv("PARTFLOW_OFFLINE_GRANT_PUBLIC_KEY", base64.StdEncoding.EncodeToString(publicKey))
	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	t.Setenv("PARTFLOW_CLOUD_API_URL", "http://127.0.0.1:1/api/v1")
	SetJWTSecret("expired-grant-test-secret")

	localDB, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer localDB.Close()
	SetDatabase(localDB)
	userID := uuid.New()
	grant, err := offlinegrant.Issue(userID.String(), nil, time.Now().UTC().Add(-offlinegrant.GracePeriod-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := localDB.Exec(`CREATE TABLE cloud_auth_grants (user_id TEXT PRIMARY KEY, grant_token TEXT NOT NULL, updated_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := localDB.Exec(`INSERT INTO cloud_auth_grants (user_id, grant_token, updated_at) VALUES ($1, $2, $3)`, userID.String(), grant, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	localToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID.String(), "exp": time.Now().Add(time.Hour).Unix()})
	localTokenString, err := localToken.SignedString([]byte("expired-grant-test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(Auth())
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+localTokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expired grace grant status=%d body=%s", w.Code, w.Body.String())
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
	if _, err := rawDB.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, is_active INTEGER NOT NULL, subscription_status TEXT, subscription_expires_at TEXT)`); err != nil {
		t.Fatal(err)
	}
	userID := uuid.New()
	if _, err := rawDB.Exec(`INSERT INTO users (id, is_active, subscription_status) VALUES (?, 1, 'suspended')`, userID.String()); err != nil {
		t.Fatal(err)
	}
	SetDatabase(sqlx.NewDb(rawDB, "sqlite"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID.String(), "exp": time.Now().Add(time.Hour).Unix()})
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
	if w.Code != http.StatusForbidden {
		t.Fatalf("suspended cloud request status=%d body=%s, want 403", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "SUBSCRIPTION_SUSPENDED") {
		t.Fatalf("suspended response lacks stable code: %s", w.Body.String())
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
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("database failure status=%d body=%s, want temporary 503", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "AUTH_SERVICE_UNAVAILABLE") {
		t.Fatalf("database failure lacks stable temporary code: %s", w.Body.String())
	}
}
