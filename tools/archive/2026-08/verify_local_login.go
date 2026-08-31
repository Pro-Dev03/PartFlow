package main

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/partflow/smart-store/internal/auth"
    "golang.org/x/crypto/bcrypt"
    _ "modernc.org/sqlite"
)

func main() {
    dbPath := `C:\Users\Administrator\Desktop\PartFlow\local-partflow.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)`
    db, err := sql.Open("sqlite", dbPath)
    if err != nil { panic(err) }
    defer db.Close()

    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        email TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        phone TEXT,
        avatar_url TEXT,
        role TEXT DEFAULT 'owner',
        is_active INTEGER NOT NULL DEFAULT 1,
        is_verified INTEGER NOT NULL DEFAULT 0,
        last_login_at TEXT,
        subscription_status TEXT DEFAULT 'active',
        subscription_expires_at TEXT,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL
    )`)
    if err != nil { panic(err) }

    if _, err = db.Exec(`DELETE FROM users WHERE email = ?`, "owner@partflow.com"); err != nil {
        panic(err)
    }

    hash, err := bcrypt.GenerateFromPassword([]byte("Owner123456"), bcrypt.DefaultCost)
    if err != nil { panic(err) }

    userID := uuid.NewString()
    now := time.Now().UTC().Format(time.RFC3339)
    expires := time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339)

    _, err = db.Exec(`
        INSERT INTO users (
            id, email, password_hash, first_name, last_name, phone, role,
            is_active, is_verified, subscription_status, subscription_expires_at,
            created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, 1, 1, 'active', ?, ?, ?)
    `, userID, "owner@partflow.com", string(hash), "Owner", "Admin", "+970599000000", "owner", expires, now, now)
    if err != nil { panic(err) }

    dbx, err := sqlx.Connect("sqlite", dbPath)
    if err != nil { panic(err) }
    defer dbx.Close()

    svc, err := auth.NewService(dbx, "test-secret", false, "", "")
    if err != nil { panic(err) }

    result, err := svc.Login(context.Background(), &auth.LoginRequest{Email: "owner@partflow.com", Password: "Owner123456"})
    if err != nil { panic(err) }

    fmt.Printf("user_id=%s\n", userID)
    fmt.Printf("email=%s\n", result.User.Email)
    fmt.Printf("token_length=%d\n", len(result.AccessToken))
    fmt.Printf("subscription_status=%s\n", result.User.SubscriptionStatus)
}
