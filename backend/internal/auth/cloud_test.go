package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func TestCreateLocalSessionUsesCloudAuthorityNotLocalSubscription(t *testing.T) {
	userID := uuid.New()
	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/validate" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer cloud-access-token" {
			http.Error(w, "unexpected token", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"success": true,
			"data": {
				"valid": true,
				"user": {
					"id": "%s",
					"email": "owner@example.test",
					"is_active": true,
					"subscription_status": "active"
				},
				"subscription_status": "active"
			}
		}`, userID)
	}))
	defer cloud.Close()

	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		password_hash TEXT NOT NULL DEFAULT '',
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		phone TEXT,
		is_active INTEGER NOT NULL DEFAULT 0,
		last_login_at TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		subscription_status TEXT,
		subscription_expires_at TEXT
	)`); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	expiredAt := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO users (
		id, email, first_name, last_name, is_active, created_at, updated_at, subscription_status, subscription_expires_at
	) VALUES (?, ?, ?, ?, 0, ?, ?, 'expired', ?)`,
		userID.String(), "owner@example.test", "Owner", "Shop", now, now, expiredAt,
	); err != nil {
		t.Fatal(err)
	}

	jwtService := NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	service := NewCloudAuthService(cloud.URL)
	session, err := service.CreateLocalSession(context.Background(), jwtService, db, "cloud-access-token")
	if err != nil {
		t.Fatalf("cloud-approved account should receive a local session even if SQLite is expired: %v", err)
	}
	if session.AccessToken == "" {
		t.Fatal("expected a local access token")
	}
	if session.User.SubscriptionStatus != "active" {
		t.Fatalf("expected returned user to reflect cloud status, got %q", session.User.SubscriptionStatus)
	}
}

func TestCreateLocalSessionRejectsWhenCloudValidateFails(t *testing.T) {
	cloud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer cloud.Close()

	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	jwtService := NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)
	service := NewCloudAuthService(cloud.URL)
	if _, err := service.CreateLocalSession(context.Background(), jwtService, db, "cloud-access-token"); err == nil {
		t.Fatal("expected cloud rejection to block local session creation")
	}
}
