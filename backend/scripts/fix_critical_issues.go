package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to database
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

	// Apply fixes
	if err := applyFixes(db); err != nil {
		log.Fatalf("Failed to apply fixes: %v", err)
	}

	fmt.Println("All critical fixes applied successfully!")
}

func applyFixes(db *sql.DB) error {
	// 1. Fix trade_ins table column name
	fmt.Println("Fixing trade_ins table column name...")
	_, err := db.Exec(`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'trade_ins' AND column_name = 'inventory_item_id'
			) THEN
				ALTER TABLE trade_ins RENAME COLUMN inventory_item_id TO item_id;
				RAISE NOTICE 'Renamed inventory_item_id to item_id in trade_ins table';
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to fix trade_ins column: %w", err)
	}

	// 2. Make sales user_id nullable
	fmt.Println("Making sales.user_id nullable...")
	_, err = db.Exec(`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'sales' AND column_name = 'user_id' AND is_nullable = 'NO'
			) THEN
				ALTER TABLE sales ALTER COLUMN user_id DROP NOT NULL;
				RAISE NOTICE 'Made sales.user_id nullable';
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to make sales.user_id nullable: %w", err)
	}

	// 3. Ensure main warehouse location exists
	fmt.Println("Ensuring main warehouse location exists...")
	_, err = db.Exec(`
		INSERT INTO locations (id, name, type, is_active, created_at, updated_at)
		VALUES ('9b561ebf-0382-42c5-82e6-61d9627e918d', 'المخزن الرئيسي', 'warehouse', true, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to ensure main warehouse exists: %w", err)
	}

	// 4. Add proper barcode uniqueness constraint
	fmt.Println("Adding barcode uniqueness constraint...")
	_, err = db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.table_constraints 
				WHERE table_name = 'inventory_items' AND constraint_name = 'inventory_items_barcode_key'
			) THEN
				ALTER TABLE inventory_items DROP CONSTRAINT IF EXISTS inventory_items_barcode_key;
				ALTER TABLE inventory_items ADD CONSTRAINT inventory_items_barcode_key UNIQUE (barcode);
				RAISE NOTICE 'Added proper barcode uniqueness constraint';
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add barcode constraint: %w", err)
	}

	// 5. Add purchase_cost to trade_ins if missing
	fmt.Println("Adding purchase_cost to trade_ins...")
	_, err = db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'trade_ins' AND column_name = 'purchase_cost'
			) THEN
				ALTER TABLE trade_ins ADD COLUMN purchase_cost INT;
				RAISE NOTICE 'Added purchase_cost column to trade_ins';
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add purchase_cost to trade_ins: %w", err)
	}

	return nil
}