package users

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func newSubscriptionSessionDB(t *testing.T) (*Service, *sql.DB, uuid.UUID) {
	t.Helper()
	rawDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rawDB.Close() })
	_, err = rawDB.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY, email TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL,
		first_name TEXT NOT NULL, last_name TEXT NOT NULL, phone TEXT, avatar_url TEXT,
		is_active INTEGER NOT NULL DEFAULT 1, is_verified INTEGER NOT NULL DEFAULT 0,
		last_login_at TEXT, subscription_status TEXT DEFAULT 'active', subscription_expires_at TEXT,
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rawDB.Exec(`CREATE TABLE refresh_tokens (
		id TEXT PRIMARY KEY, user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token TEXT NOT NULL UNIQUE, expires_at TEXT NOT NULL, created_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.New()
	now := time.Now().UTC().Format(time.RFC3339)
	expires := time.Now().AddDate(0, 1, 0).UTC().Format(time.RFC3339)
	_, err = rawDB.Exec(`INSERT INTO users (id, email, password_hash, first_name, last_name, is_active, subscription_status, subscription_expires_at, created_at, updated_at)
		VALUES (?, ?, '', 'Test', 'Subscriber', 1, 'active', ?, ?, ?)`, userID.String(), "subscriber@example.test", expires, now, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rawDB.Exec(`INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`, uuid.NewString(), userID.String(), "refresh-secret", expires, now)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(sqlx.NewDb(rawDB, "sqlite"))
	return NewService(repo), rawDB, userID
}

func TestSuspensionRevokesAllRefreshTokensAndReactivationKeepsThemRevoked(t *testing.T) {
	service, db, userID := newSubscriptionSessionDB(t)
	user, err := service.UpdateSubscription(context.Background(), userID, "suspended", nil)
	if err != nil {
		t.Fatalf("suspend subscriber: %v", err)
	}
	if user.SubscriptionStatus != "suspended" {
		t.Fatalf("status=%q, want suspended", user.SubscriptionStatus)
	}
	var tokenCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens WHERE user_id = ?`, userID.String()).Scan(&tokenCount); err != nil || tokenCount != 0 {
		t.Fatalf("refresh token count=%d err=%v, want 0 after suspension", tokenCount, err)
	}
	if _, err := service.RenewSubscription(context.Background(), userID, 30); err != nil {
		t.Fatalf("reactivate subscriber: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens WHERE user_id = ?`, userID.String()).Scan(&tokenCount); err != nil || tokenCount != 0 {
		t.Fatalf("refresh token count=%d err=%v, want 0 after reactivation", tokenCount, err)
	}
}

func TestDeletingSubscriberRevokesRefreshTokens(t *testing.T) {
	service, db, userID := newSubscriptionSessionDB(t)
	if err := service.DeleteUser(context.Background(), userID); err != nil {
		t.Fatalf("delete subscriber: %v", err)
	}
	var tokenCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens`).Scan(&tokenCount); err != nil || tokenCount != 0 {
		t.Fatalf("refresh token count=%d err=%v, want 0 after deletion", tokenCount, err)
	}
}
