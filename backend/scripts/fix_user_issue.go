package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
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

	// Drop the foreign key constraint temporarily
	fmt.Println("Dropping foreign key constraint on sales.user_id...")
	_, err = db.Exec(`ALTER TABLE sales DROP CONSTRAINT IF EXISTS sales_user_id_fkey`)
	if err != nil {
		log.Printf("Error dropping constraint: %v", err)
	} else {
		fmt.Println("Successfully dropped constraint")
	}

	// Recreate constraint without NOT NULL requirement
	fmt.Println("Recreating constraint...")
	_, err = db.Exec(`
		ALTER TABLE sales 
		ADD CONSTRAINT sales_user_id_fkey 
		FOREIGN KEY (user_id) REFERENCES users(id) 
		ON DELETE SET NULL
	`)
	if err != nil {
		log.Printf("Error recreating constraint: %v", err)
	} else {
		fmt.Println("Successfully recreated constraint with ON DELETE SET NULL")
	}

	// Also make sure the column is nullable
	fmt.Println("Ensuring column is nullable...")
	_, err = db.Exec(`ALTER TABLE sales ALTER COLUMN user_id DROP NOT NULL`)
	if err != nil {
		log.Printf("Error making nullable: %v", err)
	} else {
		fmt.Println("Successfully made column nullable")
	}

	// Test with a valid user ID
	testUserID := uuid.MustParse("6deb0af1-136e-4aab-a46d-bb4799627321")
	fmt.Printf("Testing with user ID: %s\n", testUserID)
	
	var userExists bool
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, testUserID).Scan(&userExists)
	if err != nil {
		log.Printf("Error checking user: %v", err)
	} else {
		fmt.Printf("User exists: %v\n", userExists)
	}

	fmt.Println("Fix completed!")
}