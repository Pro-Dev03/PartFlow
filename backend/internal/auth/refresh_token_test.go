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
	_, err = db.Exec(`CREATE TABLE password_reset_tokens (
		id TEXT PRIMARY KEY, user_id TEXT NOT NULL, token TEXT NOT NULL UNIQUE,
		expires_at TEXT NOT NULL, used INTEGER NOT NULL DEFAULT 0, used_at TEXT, created_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatal(err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("TestOwnerPassword123!"), bcrypt.DefaultCost)
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

	service, err := NewService(sqlx.NewDb(db, "sqlite"), "test-secret", false, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	return service, db, userID
}

func TestRefreshTokenIsRotatedAndOldTokenRejected(t *testing.T) {
	service, db, userID := newRefreshTokenTestService(t)
	ctx := context.Background()

	login, err := service.Login(ctx, &LoginRequest{Email: "refresh@example.test", Password: "TestOwnerPassword123!"})
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

func TestDatabaseOutageIsNotReportedAsMissingUser(t *testing.T) {
	service, db, userID := newRefreshTokenTestService(t)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Login(context.Background(), &LoginRequest{Email: "refresh@example.test", Password: "TestOwnerPassword123!"}); err == nil || errors.Is(err, ErrUserNotFound) {
		t.Fatalf("login error=%v, want a transient database error", err)
	}
	if _, err := service.GetUserByID(context.Background(), userID); err == nil || errors.Is(err, ErrUserNotFound) {
		t.Fatalf("GetUserByID error=%v, want a transient database error", err)
	}

	refreshToken, err := service.jwtService.GenerateRefreshToken(userID.String())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.RefreshToken(context.Background(), refreshToken); err == nil || errors.Is(err, ErrUserNotFound) {
		t.Fatalf("RefreshToken error=%v, want a transient database error", err)
	}
}

func TestLoginAndRefreshRejectSuspendedSubscription(t *testing.T) {
	service, db, userID := newRefreshTokenTestService(t)
	ctx := context.Background()
	login, err := service.Login(ctx, &LoginRequest{Email: "refresh@example.test", Password: "TestOwnerPassword123!"})
	if err != nil {
		t.Fatalf("login before suspension: %v", err)
	}
	if _, err := db.Exec(`UPDATE users SET subscription_status = 'suspended' WHERE id = ?`, userID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RefreshToken(ctx, login.RefreshToken); !errors.Is(err, ErrSubscriptionSuspended) {
		t.Fatalf("refresh after suspension error=%v, want ErrSubscriptionSuspended", err)
	}
	if _, err := service.Login(ctx, &LoginRequest{Email: "refresh@example.test", Password: "TestOwnerPassword123!"}); !errors.Is(err, ErrSubscriptionSuspended) {
		t.Fatalf("login after suspension error=%v, want ErrSubscriptionSuspended", err)
	}
}

func TestMissingRefreshTokenTableFailsClosed(t *testing.T) {
	service, db, userID := newRefreshTokenTestService(t)
	if _, err := db.Exec(`DROP TABLE refresh_tokens`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Login(context.Background(), &LoginRequest{Email: "refresh@example.test", Password: "TestOwnerPassword123!"}); err == nil {
		t.Fatal("login must not issue an unrevocable refresh token")
	}
	token, err := service.jwtService.GenerateRefreshToken(userID.String())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.RefreshToken(context.Background(), token); err == nil {
		t.Fatal("legacy refresh token must not work without revocation storage")
	}
}

func TestChangePasswordRevokesAllRefreshTokens(t *testing.T) {
	service, db, userID := newRefreshTokenTestService(t)
	ctx := context.Background()
	login, err := service.Login(ctx, &LoginRequest{Email: "refresh@example.test", Password: "TestOwnerPassword123!"})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ChangePassword(ctx, userID, &ChangePasswordRequest{CurrentPassword: "TestOwnerPassword123!", NewPassword: "NewPassword123!"}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens WHERE user_id = ?`, userID.String()).Scan(&count); err != nil || count != 0 {
		t.Fatalf("remaining refresh tokens=%d err=%v", count, err)
	}
	if _, err := service.RefreshToken(ctx, login.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("old token error=%v, want ErrInvalidToken", err)
	}
}

func TestPasswordResetConsumesTokenAndRevokesEveryRefreshToken(t *testing.T) {
	service, db, userID := newRefreshTokenTestService(t)
	login, err := service.Login(context.Background(), &LoginRequest{Email: "refresh@example.test", Password: "TestOwnerPassword123!"})
	if err != nil {
		t.Fatalf("login before password reset: %v", err)
	}
	resetToken := uuid.NewString()
	expires := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	created := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO password_reset_tokens (id, user_id, token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`, uuid.NewString(), userID.String(), resetToken, expires, created); err != nil {
		t.Fatal(err)
	}

	if err := service.ResetPassword(context.Background(), resetToken, "ResetPassword123!"); err != nil {
		t.Fatalf("reset password: %v", err)
	}
	var storedHash string
	if err := db.QueryRow(`SELECT password_hash FROM users WHERE id = ?`, userID.String()).Scan(&storedHash); err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("ResetPassword123!")); err != nil {
		t.Fatalf("reset password was not stored: %v", err)
	}
	var refreshCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens WHERE user_id = ?`, userID.String()).Scan(&refreshCount); err != nil || refreshCount != 0 {
		t.Fatalf("refresh token count=%d err=%v, want 0 after reset", refreshCount, err)
	}
	var used bool
	if err := db.QueryRow(`SELECT used FROM password_reset_tokens WHERE token = ?`, resetToken).Scan(&used); err != nil || !used {
		t.Fatalf("reset token used=%t err=%v, want consumed", used, err)
	}
	if _, err := service.RefreshToken(context.Background(), login.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("refresh after password reset error=%v, want ErrInvalidToken", err)
	}
}

func TestPasswordResetFailsClosedWhenRefreshRevocationStorageIsMissing(t *testing.T) {
	service, db, userID := newRefreshTokenTestService(t)
	resetToken := uuid.NewString()
	expires := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	created := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO password_reset_tokens (id, user_id, token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`, uuid.NewString(), userID.String(), resetToken, expires, created); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DROP TABLE refresh_tokens`); err != nil {
		t.Fatal(err)
	}

	if err := service.ResetPassword(context.Background(), resetToken, "ResetPassword123!"); err == nil {
		t.Fatal("password reset succeeded without refresh-token revocation storage")
	}
	var passwordHash string
	if err := db.QueryRow(`SELECT password_hash FROM users WHERE id = ?`, userID.String()).Scan(&passwordHash); err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("TestOwnerPassword123!")); err != nil {
		t.Fatalf("password changed despite missing refresh-token storage: %v", err)
	}
	var used bool
	if err := db.QueryRow(`SELECT used FROM password_reset_tokens WHERE token = ?`, resetToken).Scan(&used); err != nil || used {
		t.Fatalf("reset token used=%t err=%v, want unused after rollback", used, err)
	}
}
