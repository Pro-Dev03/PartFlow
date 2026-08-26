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

	// Remove inventory_movements.created_by constraint
	fmt.Println("Removing inventory_movements.created_by constraint...")
	_, err = db.Exec(`ALTER TABLE inventory_movements DROP CONSTRAINT IF EXISTS inventory_movements_created_by_fkey`)
	if err != nil {
		log.Printf("Error dropping constraint: %v", err)
	} else {
		fmt.Println("Successfully removed constraint")
	}

	// Make created_by nullable
	fmt.Println("Making inventory_movements.created_by nullable...")
	_, err = db.Exec(`ALTER TABLE inventory_movements ALTER COLUMN created_by DROP NOT NULL`)
	if err != nil {
		log.Printf("Error making created_by nullable: %v", err)
	} else {
		fmt.Println("Successfully made created_by nullable")
	}

	fmt.Println("Constraints fixed for testing!")
}