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

	// Remove the foreign key constraint completely for testing
	fmt.Println("Removing foreign key constraint on sales.user_id...")
	_, err = db.Exec(`ALTER TABLE sales DROP CONSTRAINT IF EXISTS sales_user_id_fkey`)
	if err != nil {
		log.Printf("Error dropping constraint: %v", err)
	} else {
		fmt.Println("Successfully removed constraint")
	}

	fmt.Println("Fix completed - foreign key constraint removed for testing")
}