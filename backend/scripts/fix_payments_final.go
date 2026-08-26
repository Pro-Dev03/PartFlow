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

	// Add reference_number column
	fmt.Println("Adding reference_number column to payments table...")
	_, err = db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'payments' AND column_name = 'reference_number'
			) THEN
				ALTER TABLE payments ADD COLUMN reference_number VARCHAR(100);
				RAISE NOTICE 'Added reference_number column';
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Error adding reference_number: %v", err)
	} else {
		fmt.Println("Successfully added reference_number column")
	}

	// Make reference_number nullable
	fmt.Println("Making reference_number nullable...")
	_, err = db.Exec(`ALTER TABLE payments ALTER COLUMN reference_number DROP NOT NULL`)
	if err != nil {
		log.Printf("Error making reference_number nullable: %v", err)
	} else {
		fmt.Println("Successfully made reference_number nullable")
	}

	fmt.Println("Payments table final fix completed!")
}