package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func newRefreshTokenTestService(t *testing.T) (*Service, *sql.DB, uuid.UUID) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL,
		first_name TEXT NOT NULL, last_name TEXT NOT NULL, phone TEXT,
		is_active INTEGER NOT NULL DEFAULT 1, last_login_at TEXT,
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
		subscription_status TEXT DEFAULT 'active', subscription_expires_at TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE refresh_tokens (
		id TEXT PRIMARY KEY, user_id TEXT NOT NULL, token TEXT NOT NULL UNIQUE,
		expires_at TEXT NOT NULL, created_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatal(err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("Owner123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339)
	expires := time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339)
	_, err = db.Exec(`INSERT INTO users
		(id, email, password_hash, first_name, last_name, phone, is_active, created_at, updated_at, subscription_status, subscription_expires_at)
		VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, 'active', ?)`,
		userID.String(), "refresh@example.test", string(passwordHash), "Refresh", "Tester", "",
		now, now, expires)
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(sqlx.NewDb(db, "sqlite"), "test-secret", false, "", "")
	if err != nil {
		t.Fatal(err)
	}
	return service, db, userID
}

func TestRefreshTokenIsRotatedAndOldTokenRejected(t *testing.T) {
	service, db, userID := newRefreshTokenTestService(t)
	ctx := context.Background()

	login, err := service.Login(ctx, &LoginRequest{Email: "refresh@example.test", Password: "Owner123456"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("stored refresh token count=%d err=%v, want 1", count, err)
	}

	refreshed, err := service.RefreshToken(ctx, login.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.RefreshToken == login.RefreshToken {
		t.Fatal("expected refresh token rotation")
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("rotated refresh token count=%d err=%v, want 1", count, err)
	}

	if _, err := service.RefreshToken(ctx, login.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("old refresh token error=%v, want ErrInvalidToken", err)
	}
	if err := service.Logout(ctx, userID); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("after logout token count=%d err=%v, want 0", count, err)
	}
}
