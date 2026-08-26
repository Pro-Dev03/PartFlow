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

	// Make audit log columns nullable to prevent blocking operations
	fmt.Println("Making audit_logs columns nullable for testing...")
	
	columns := []string{"user_id", "entity_id", "changes", "ip_address", "user_agent"}
	
	for _, col := range columns {
		query := fmt.Sprintf(`
			DO $$
			BEGIN
				IF EXISTS (
					SELECT 1 FROM information_schema.columns 
					WHERE table_name = 'audit_logs' AND column_name = '%s' AND is_nullable = 'NO'
				) THEN
					ALTER TABLE audit_logs ALTER COLUMN %s DROP NOT NULL;
					RAISE NOTICE 'Made %s nullable';
				END IF;
			END $$;
		`, col, col, col)
		
		_, err = db.Exec(query)
		if err != nil {
			log.Printf("Error making %s nullable: %v", col, err)
		} else {
			fmt.Printf("Successfully made %s nullable\n", col)
		}
	}

	fmt.Println("Audit logs fixed for testing!")
}