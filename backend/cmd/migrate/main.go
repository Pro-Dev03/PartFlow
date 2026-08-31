package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Run migrations
	migrationsDir := "./migrations"
	migrationFiles, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Sort migration files
	sort.Strings(migrationFiles)
	// The initial schema must precede the legacy users fix migration. Both
	// share the 001 prefix, so lexical ordering alone is unsafe on a new DB.
	sort.SliceStable(migrationFiles, func(i, j int) bool {
		left, right := filepath.Base(migrationFiles[i]), filepath.Base(migrationFiles[j])
		if strings.HasPrefix(left, "001_") && strings.HasPrefix(right, "001_") {
			if strings.Contains(left, "_initial_") {
				return true
			}
			if strings.Contains(right, "_initial_") {
				return false
			}
		}
		return left < right
	})

	// Create migrations table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Run each migration
	for _, file := range migrationFiles {
		filename := filepath.Base(file)
		version := strings.TrimSuffix(filename, ".sql")

		// Check if migration already applied
		var appliedAt string
		err := db.QueryRow("SELECT applied_at FROM schema_migrations WHERE version = $1", version).Scan(&appliedAt)
		if err == sql.ErrNoRows {
			// Migration not applied, run it
			fmt.Printf("Applying migration: %s\n", filename)

			content, err := ioutil.ReadFile(file)
			if err != nil {
				log.Fatalf("Failed to read migration file %s: %v", filename, err)
			}

			// Execute migration
			_, err = db.Exec(string(content))
			if err != nil {
				log.Fatalf("Failed to execute migration %s: %v", filename, err)
			}

			// Record migration
			_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
			if err != nil {
				log.Fatalf("Failed to record migration %s: %v", filename, err)
			}

			fmt.Printf("Successfully applied migration: %s\n", filename)
		} else if err != nil {
			log.Fatalf("Failed to check migration status: %v", err)
		} else {
			fmt.Printf("Migration already applied: %s (at %s)\n", filename, appliedAt)
		}
	}

	fmt.Println("All migrations completed successfully")
}
