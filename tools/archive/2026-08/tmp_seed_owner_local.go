package main

import (
    "database/sql"
    "fmt"
    "time"

    "github.com/google/uuid"
    _ "modernc.org/sqlite"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    dbPath := `C:\Users\Administrator\Desktop\PartFlow\backend\partflow-local.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)`
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

    hash, err := bcrypt.GenerateFromPassword([]byte("Owner123456"), bcrypt.DefaultCost)
    if err != nil { panic(err) }

    now := time.Now().UTC().Format(time.RFC3339)
    expires := time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339)
    uid := uuid.NewString()

    _, err = db.Exec(`
        INSERT INTO users (
            id, email, password_hash, first_name, last_name, phone, role,
            is_active, is_verified, subscription_status, subscription_expires_at,
            created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, 1, 1, 'active', ?, ?, ?)
        ON CONFLICT(email) DO UPDATE SET
            password_hash = excluded.password_hash,
            first_name = excluded.first_name,
            last_name = excluded.last_name,
            phone = excluded.phone,
            role = excluded.role,
            is_active = 1,
            is_verified = 1,
            subscription_status = excluded.subscription_status,
            subscription_expires_at = excluded.subscription_expires_at,
            updated_at = excluded.updated_at
    `, uid, "owner@partflow.com", string(hash), "Owner", "Admin", "+970599000000", "owner", expires, now, now)
    if err != nil { panic(err) }

    fmt.Println("seed ok")
}
