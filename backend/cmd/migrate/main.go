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

	argumentIndex := 1
	if len(os.Args) > argumentIndex && os.Args[argumentIndex] == "--" {
		argumentIndex++
	}
	if len(os.Args) > argumentIndex {
		targetVersion := strings.TrimSuffix(filepath.Base(os.Args[argumentIndex]), ".sql")
		filtered := make([]string, 0, 1)
		for _, file := range migrationFiles {
			if strings.TrimSuffix(filepath.Base(file), ".sql") == targetVersion {
				filtered = append(filtered, file)
				break
			}
		}
		if len(filtered) == 0 {
			log.Fatalf("Migration not found: %s", os.Args[argumentIndex])
		}
		migrationFiles = filtered
		fmt.Printf("Running selected migration: %s\n", targetVersion)
	}

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

		// Check both the current canonical name (without .sql) and the legacy
		// name used by older scripts (with .sql). This prevents a deployment
		// with an existing schema from trying to replay the initial migration.
		appliedVersion, appliedAt, err := findAppliedMigration(db, version, filename)
		if err == sql.ErrNoRows {
			// Migration not applied, run it
			fmt.Printf("Applying migration: %s\n", filename)

			content, err := ioutil.ReadFile(file)
			if err != nil {
				log.Fatalf("Failed to read migration file %s: %v", filename, err)
			}

			// Execute the migration and record it atomically. PostgreSQL DDL is
			// transactional, so a failed statement cannot leave a migration
			// half-applied while its version is marked as complete.
			tx, err := db.Begin()
			if err != nil {
				log.Fatalf("Failed to begin migration %s: %v", filename, err)
			}
			if _, err = tx.Exec(stripTransactionControlStatements(string(content))); err != nil {
				_ = tx.Rollback()
				log.Fatalf("Failed to execute migration %s: %v", filename, err)
			}
			if _, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
				_ = tx.Rollback()
				log.Fatalf("Failed to record migration %s: %v", filename, err)
			}
			if err = tx.Commit(); err != nil {
				log.Fatalf("Failed to commit migration %s: %v", filename, err)
			}

			fmt.Printf("Successfully applied migration: %s\n", filename)
		} else if err != nil {
			log.Fatalf("Failed to check migration status: %v", err)
		} else {
			fmt.Printf("Migration already applied: %s (recorded as %s at %s)\n", filename, appliedVersion, appliedAt)
		}
	}

	fmt.Println("All migrations completed successfully")
}

func findAppliedMigration(db *sql.DB, version, filename string) (appliedVersion, appliedAt string, err error) {
	err = db.QueryRow(`
		SELECT version, applied_at
		FROM schema_migrations
		WHERE version = $1 OR version = $2
		ORDER BY CASE WHEN version = $1 THEN 0 ELSE 1 END
		LIMIT 1
	`, version, filename).Scan(&appliedVersion, &appliedAt)
	return
}

// stripTransactionControlStatements keeps migration files that historically
// wrapped themselves in BEGIN/COMMIT compatible with the migrator's outer
// transaction. PostgreSQL treats a nested COMMIT as a commit of the outer
// transaction, which would make the schema_migrations record non-atomic.
// Only standalone transaction-control lines are removed; procedural BEGIN
// blocks and statements containing these words remain untouched.
func stripTransactionControlStatements(sqlText string) string {
	lines := strings.Split(sqlText, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		statement := strings.ToUpper(strings.TrimSpace(strings.TrimSuffix(line, "\r")))
		if statement == "BEGIN;" || statement == "COMMIT;" || statement == "ROLLBACK;" {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}
