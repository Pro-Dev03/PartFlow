//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres.vjaokclajtvpdjmdrybu:9IyxKnou9CUgrj5k@aws-0-eu-central-1.pooler.supabase.com:5432/postgres"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Add missing columns to payments table
	columns := []struct {
		name     string
		colType  string
		defaultVal string
	}{
		{"created_by", "UUID", "NULL"},
		{"reference_type", "VARCHAR(50)", "NULL"},
		{"reference_id", "UUID", "NULL"},
	}

	for _, col := range columns {
		fmt.Printf("Adding column %s to payments table...\n", col.name)
		query := fmt.Sprintf(`
			DO $$
			BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'payments' AND column_name = '%s'
				) THEN
					ALTER TABLE payments ADD COLUMN %s %s %s;
					RAISE NOTICE 'Added %s column';
				END IF;
			END $$;
		`, col.name, col.name, col.colType, col.defaultVal, col.name)
		
		_, err = db.Exec(query)
		if err != nil {
			log.Printf("Error adding %s: %v", col.name, err)
		} else {
			fmt.Printf("Successfully added %s\n", col.name)
		}
	}

	fmt.Println("Payments table completely fixed!")
}