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

func TestAdminAllowsConfiguredCloudEmailWithoutLocalUserRow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/validate" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"valid":true,"user":{"id":"%s","email":"admin@example.test"}}}`, userID)
	}))
	defer cloud.Close()

	t.Setenv("DB_CONNECTION_MODE", "local")
	t.Setenv("PARTFLOW_REQUIRE_CLOUD_AUTH", "true")
	t.Setenv("PARTFLOW_CLOUD_API_URL", cloud.URL)
	t.Setenv("PARTFLOW_ADMIN_EMAILS", "admin@example.test")
	SetJWTSecret("test-secret")
	SetDisableAuth(false)
	previousDB := db
	SetDatabase(nil)
	t.Cleanup(func() { SetDatabase(previousDB) })

	router := gin.New()
	router.Use(Auth(), Admin())
	router.GET("/admin", func(c *gin.Context) {
		if GetUserID(c) != userID {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/admin", nil)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID.String()})
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("X-PartFlow-Cloud-Token", "cloud-access-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected configured cloud admin to be allowed, status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAdminRejectsRegularUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("connect sqlite: %v", err)
	}
	defer testDB.Close()
	if _, err := testDB.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT NOT NULL)`); err != nil {
		t.Fatalf("create users table: %v", err)
	}
	userID := uuid.New()
	if _, err := testDB.Exec(`INSERT INTO users (id, email) VALUES (?, ?)`, userID.String(), "subscriber@example.test"); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	previousDB := db
	SetDatabase(testDB)
	t.Cleanup(func() { SetDatabase(previousDB) })
	t.Setenv("PARTFLOW_ADMIN_EMAILS", "")
	SetDisableAuth(false)

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	Admin()(c)
	if w.Code != 403 {
		t.Fatalf("expected regular user to be rejected, got %d", w.Code)
	}
}

func TestAdminAllowsConfiguredEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("connect sqlite: %v", err)
	}
	defer testDB.Close()
	if _, err := testDB.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT NOT NULL)`); err != nil {
		t.Fatalf("create users table: %v", err)
	}
	userID := uuid.New()
	if _, err := testDB.Exec(`INSERT INTO users (id, email) VALUES (?, ?)`, userID.String(), "admin@example.test"); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	previousDB := db
	SetDatabase(testDB)
	t.Cleanup(func() { SetDatabase(previousDB) })
	t.Setenv("PARTFLOW_ADMIN_EMAILS", "admin@example.test")
	SetJWTSecret("test-secret")
	SetDisableAuth(false)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}, Admin())
	router.GET("/admin", func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected configured admin to be allowed, status=%d", w.Code)
	}
}

func TestOwnerOnlyRequiresExactOwnerEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PARTFLOW_ADMIN_EMAILS", "admin@example.test")
	testDB, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	if _, err := testDB.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	ownerID, adminID := uuid.New(), uuid.New()
	if _, err := testDB.Exec(`INSERT INTO users (id,email) VALUES (?,?),(?,?)`, ownerID.String(), "owner@partflow.com", adminID.String(), "admin@example.test"); err != nil {
		t.Fatal(err)
	}
	previousDB := db
	SetDatabase(testDB)
	t.Cleanup(func() { SetDatabase(previousDB) })

	for _, tc := range []struct {
		name   string
		id     uuid.UUID
		status int
	}{
		{name: "owner", id: ownerID, status: http.StatusOK},
		{name: "configured admin is not owner", id: adminID, status: http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/owner", nil)
			ctx.Set("user_id", tc.id)
			ctx.Set("cloud_user_email", "")
			OwnerOnly()(ctx)
			if response.Code != tc.status {
				t.Fatalf("OwnerOnly status=%d body=%s; want %d", response.Code, response.Body.String(), tc.status)
			}
			if tc.status == http.StatusForbidden && !strings.Contains(response.Body.String(), "OWNER_REQUIRED") {
				t.Fatalf("configured admin denial lacks OWNER_REQUIRED: %s", response.Body.String())
			}
		})
	}
}

func TestLocalDatabaseOnlyRejectsCloudDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousDB := db
	SetDatabase(sqlx.NewDb(nil, "pgx"))
	t.Cleanup(func() { SetDatabase(previousDB) })
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/settings/database/sync", nil)
	LocalDatabaseOnly()(ctx)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "LOCAL_DATABASE_REQUIRED") {
		t.Fatalf("cloud database sync guard status=%d body=%s; want LOCAL_DATABASE_REQUIRED", response.Code, response.Body.String())
	}
}
