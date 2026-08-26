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

	// Check current state
	var isNullable string
	err = db.QueryRow(`
		SELECT is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'sales' AND column_name = 'user_id'
	`).Scan(&isNullable)
	
	if err != nil {
		log.Printf("Error checking column: %v", err)
	} else {
		fmt.Printf("Current sales.user_id nullable status: %s\n", isNullable)
	}

	// Force fix
	fmt.Println("Forcing fix for sales.user_id...")
	_, err = db.Exec(`ALTER TABLE sales ALTER COLUMN user_id DROP NOT NULL`)
	if err != nil {
		log.Printf("Error dropping NOT NULL: %v", err)
	} else {
		fmt.Println("Successfully made sales.user_id nullable")
	}

	// Verify
	err = db.QueryRow(`
		SELECT is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'sales' AND column_name = 'user_id'
	`).Scan(&isNullable)
	
	if err != nil {
		log.Printf("Error verifying: %v", err)
	} else {
		fmt.Printf("Verified sales.user_id nullable status: %s\n", isNullable)
	}

	fmt.Println("Fix completed!")
}