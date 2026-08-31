package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestAuthAllowsValidJWTWhenLocalSQLiteHasNoUserRecords(t *testing.T) {
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

	if w.Code == 401 {
		t.Fatalf("expected JWT to be accepted in local mode even when users table is empty")
	}

	if got := GetUserID(c); got.String() != userID {
		t.Fatalf("expected user_id=%s in context, got %s", userID, got.String())
	}
}
