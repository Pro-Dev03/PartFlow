package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Println("=== Applying Database Migrations ===\n")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create schema_migrations table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}
	fmt.Println("✅ schema_migrations table ready")

	// Get migration files
	migrationsDir := "/home/dev-bit/project/PartFlow/backend/migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	fmt.Printf("📁 Found %d migration files\n\n", len(files))

	// Apply each migration
	for _, file := range files {
		filename := filepath.Base(file)
		version := strings.TrimSuffix(filename, ".sql")

		// Check if already applied
		var applied bool
		err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied)
		if err != nil {
			log.Printf("⚠️  Failed to check migration status for %s: %v", version, err)
			continue
		}

		if applied {
			fmt.Printf("⏭️  Skipping %s (already applied)\n", version)
			continue
		}

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("❌ Failed to read migration file %s: %v", filename, err)
			continue
		}

		fmt.Printf("🔄 Applying %s...\n", version)

		// Execute migration
		tx, err := db.Begin()
		if err != nil {
			log.Printf("❌ Failed to begin transaction for %s: %v", version, err)
			continue
		}

		// Split by semicolon and execute each statement
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" || strings.HasPrefix(stmt, "--") {
				continue
			}

			_, err = tx.Exec(stmt)
			if err != nil {
				log.Printf("❌ Failed to execute statement in %s: %v\nStatement: %s", version, err, stmt[:min(100, len(stmt))])
				tx.Rollback()
				continue
			}
		}

		// Record migration
		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			log.Printf("❌ Failed to record migration %s: %v", version, err)
			tx.Rollback()
			continue
		}

		if err := tx.Commit(); err != nil {
			log.Printf("❌ Failed to commit migration %s: %v", version, err)
			continue
		}

		fmt.Printf("✅ Applied %s successfully\n", version)
	}

	fmt.Println("\n=== Migration Complete ===")
	
	// Show current status
	rows, err := db.Query("SELECT version, applied_at FROM schema_migrations ORDER BY applied_at")
	if err == nil {
		fmt.Println("\n📋 Applied migrations:")
		defer rows.Close()
		for rows.Next() {
			var version, appliedAt string
			if err := rows.Scan(&version, &appliedAt); err == nil {
				fmt.Printf("   - %s (applied at %s)\n", version, appliedAt)
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
