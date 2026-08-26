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

	// Add payment_status column if it doesn't exist
	fmt.Println("Adding payment_status column to payments table...")
	_, err = db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'payments' AND column_name = 'payment_status'
			) THEN
				ALTER TABLE payments ADD COLUMN payment_status VARCHAR(50) DEFAULT 'completed';
				RAISE NOTICE 'Added payment_status column';
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Error adding payment_status: %v", err)
	} else {
		fmt.Println("Successfully added payment_status column")
	}

	fmt.Println("Payments table fixed!")
}