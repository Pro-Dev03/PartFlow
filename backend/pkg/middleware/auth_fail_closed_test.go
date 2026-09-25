package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestAuthRejectsValidlySignedTokenWhenDatabaseIsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db
	t.Cleanup(func() { SetDatabase(previousDB) })
	SetDatabase(nil)
	SetDisableAuth(false)
	SetJWTSecret("database-unavailable-auth-test-secret")
	t.Setenv("DB_CONNECTION_MODE", "cloud")
	t.Setenv("DATABASE_URL", "postgres://example.test/partflow")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": uuid.NewString(),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte("database-unavailable-auth-test-secret"))
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(Auth())
	router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "AUTH_SERVICE_UNAVAILABLE") {
		t.Fatalf("database outage must deny the request: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAuthRejectsAccessTokenIssuedBeforePasswordChange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db
	t.Cleanup(func() { SetDatabase(previousDB) })
	SetDisableAuth(false)
	SetJWTSecret("password-change-session-test-secret")
	t.Setenv("DB_CONNECTION_MODE", "cloud")
	t.Setenv("DATABASE_URL", "postgres://example.test/partflow")
	t.Setenv("SERVER_MODE", "production")
	t.Setenv("APP_ENV", "production")

	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if _, err := sqlDB.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY, is_active INTEGER NOT NULL,
		subscription_status TEXT NOT NULL, subscription_expires_at TEXT,
		updated_at TEXT NOT NULL
	)`); err != nil {
		t.Fatal(err)
	}

	userID := uuid.New()
	now := time.Now().UTC()
	if _, err := sqlDB.Exec(`INSERT INTO users (id, is_active, subscription_status, subscription_expires_at, updated_at)
		VALUES (?, 1, 'active', ?, ?)`, userID.String(), now.Add(time.Hour).Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	SetDatabase(sqlx.NewDb(sqlDB, "sqlite"))

	issuedAt := now.Add(-time.Minute).Truncate(time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"iat":     issuedAt.Unix(),
		"nbf":     issuedAt.Unix(),
		"exp":     now.Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte("password-change-session-test-secret"))
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(Auth())
	router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), "SESSION_REVOKED") {
		t.Fatalf("old access token after password change: status=%d body=%s", response.Code, response.Body.String())
	}
}
