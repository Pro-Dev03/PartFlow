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

	// Remove the foreign key constraint on inventory_movements.item_id
	fmt.Println("Removing foreign key constraint on inventory_movements.item_id...")
	_, err = db.Exec(`ALTER TABLE inventory_movements DROP CONSTRAINT IF EXISTS inventory_movements_item_id_fkey`)
	if err != nil {
		log.Printf("Error dropping constraint: %v", err)
	} else {
		fmt.Println("Successfully removed constraint")
	}

	// Make item_id nullable
	fmt.Println("Making inventory_movements.item_id nullable...")
	_, err = db.Exec(`ALTER TABLE inventory_movements ALTER COLUMN item_id DROP NOT NULL`)
	if err != nil {
		log.Printf("Error making nullable: %v", err)
	} else {
		fmt.Println("Successfully made column nullable")
	}

	// Make product_id nullable
	fmt.Println("Making inventory_movements.product_id nullable...")
	_, err = db.Exec(`ALTER TABLE inventory_movements ALTER COLUMN product_id DROP NOT NULL`)
	if err != nil {
		log.Printf("Error making product_id nullable: %v", err)
	} else {
		fmt.Println("Successfully made product_id nullable")
	}

	fmt.Println("Fix completed!")
}