package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

const testOwnerPassword = "TestOwnerPassword123!"

func TestNewServiceRejectsUnsupportedSupabaseAuth(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = NewService(sqlx.NewDb(db, "sqlite"), "test-secret", true, "https://example.supabase.co", "test-key", "")
	if err == nil || !strings.Contains(err.Error(), "USE_SUPABASE_AUTH=false") {
		t.Fatalf("expected explicit unsupported Supabase error, got %v", err)
	}
}

func TestServiceRegistrationCannotCreateFreeActiveAccount(t *testing.T) {
	service, db, _ := newRefreshTokenTestService(t)

	response, err := service.Register(context.Background(), &RegisterRequest{
		Email:     "attacker@example.test",
		Password:  "AttackerPassword123!",
		FirstName: "Free",
		LastName:  "Account",
	})
	if !errors.Is(err, ErrRegistrationDisabled) {
		t.Fatalf("Register error = %v, want ErrRegistrationDisabled", err)
	}
	if response != nil {
		t.Fatalf("Register response = %#v, want nil", response)
	}

	var users, tokens int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, "attacker@example.test").Scan(&users); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens`).Scan(&tokens); err != nil {
		t.Fatal(err)
	}
	if users != 0 || tokens != 0 {
		t.Fatalf("public registration side effects: users=%d refresh_tokens=%d, want both zero", users, tokens)
	}
}

func TestLoginWorksWithSQLiteUserTimestamps(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE users (
        id TEXT PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        phone TEXT,
        is_active INTEGER NOT NULL DEFAULT 1,
        last_login_at TEXT,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL,
        subscription_status TEXT DEFAULT 'active',
        subscription_expires_at TEXT,
        session_version INTEGER NOT NULL DEFAULT 0
    )`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE refresh_tokens (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, token TEXT NOT NULL UNIQUE, expires_at TEXT NOT NULL, created_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(testOwnerPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	exp := time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339)
	uid := uuid.NewString()

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash, first_name, last_name, phone, is_active, created_at, updated_at, subscription_status, subscription_expires_at)
        VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, 'active', ?)`, uid, "owner@partflow.com", string(hash), "Owner", "Admin", "+970500000000", now, now, exp)
	if err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlite")
	svc, err := NewService(sqlxDB, "test-secret", false, "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := svc.Login(context.Background(), &LoginRequest{Email: "owner@partflow.com", Password: testOwnerPassword})
	if err != nil {
		t.Fatalf("Login returned error for valid SQLite user: %v", err)
	}
	if res == nil || res.AccessToken == "" {
		t.Fatal("expected access token for valid login")
	}
	if res.User.Email != "owner@partflow.com" {
		t.Fatalf("wrong user email: got %q", res.User.Email)
	}
}

func TestLoginHandlesSQLiteGoTimeStringFormat(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE users (
        id TEXT PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        phone TEXT,
        is_active INTEGER NOT NULL DEFAULT 1,
        last_login_at TEXT,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL,
        subscription_status TEXT DEFAULT 'active',
        subscription_expires_at TEXT,
        session_version INTEGER NOT NULL DEFAULT 0
    )`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE refresh_tokens (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, token TEXT NOT NULL UNIQUE, expires_at TEXT NOT NULL, created_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(testOwnerPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	uid := uuid.NewString()
	createdAt := "2026-08-30T04:47:42Z"
	updatedAt := "2026-08-29 21:51:00.5550644 -0700 PDT m=+0.066138301"
	expiresAt := time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339)

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash, first_name, last_name, phone, is_active, created_at, updated_at, subscription_status, subscription_expires_at)
        VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, 'active', ?)`, uid, "owner@partflow.com", string(hash), "Owner", "Admin", "+970500000000", createdAt, updatedAt, expiresAt)
	if err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlite")
	svc, err := NewService(sqlxDB, "test-secret", false, "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := svc.Login(context.Background(), &LoginRequest{Email: "owner@partflow.com", Password: testOwnerPassword})
	if err != nil {
		t.Fatalf("Login failed with SQLite Go time string: %v", err)
	}
	if res == nil || res.AccessToken == "" {
		t.Fatal("expected access token for valid SQLite Go time string row")
	}
	if res.User.Email != "owner@partflow.com" {
		t.Fatalf("wrong user email: got %q", res.User.Email)
	}
}
