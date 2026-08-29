//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to database successfully")

	// Clear inventory data
	if err := clearInventoryData(db); err != nil {
		log.Fatalf("Failed to clear inventory data: %v", err)
	}

	fmt.Println("🎉 Inventory data cleared successfully!")
}

func clearInventoryData(db *sqlx.DB) error {
	fmt.Println("🧹 Clearing inventory data...")

	// Helper function to safely delete from a table with error handling
	safeDelete := func(tableName string) (int64, error) {
		result, err := db.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
		if err != nil {
			return 0, err
		}
		count, _ := result.RowsAffected()
		return count, nil
	}

	// 1. Delete inventory movements first (they reference items)
	fmt.Println("   - Deleting inventory movements...")
	movementsDeleted, err := safeDelete("inventory_movements")
	if err != nil {
		fmt.Printf("     ⚠️  Warning: %v\n", err)
	} else {
		fmt.Printf("     ✅ Deleted %d inventory movements\n", movementsDeleted)
	}

	// 2. Delete reservations (they reference items)
	fmt.Println("   - Deleting reservations...")
	reservationsDeleted, err := safeDelete("reservations")
	if err != nil {
		fmt.Printf("     ⚠️  Warning: %v\n", err)
	} else {
		fmt.Printf("     ✅ Deleted %d reservations\n", reservationsDeleted)
	}

	// 3. Delete item specification values
	fmt.Println("   - Deleting item specification values...")
	specsDeleted, err := safeDelete("item_specification_values")
	if err != nil {
		fmt.Printf("     ⚠️  Warning: %v\n", err)
	} else {
		fmt.Printf("     ✅ Deleted %d item specification values\n", specsDeleted)
	}

	// 4. Delete trade-ins records
	fmt.Println("   - Deleting trade-ins records...")
	tradeInsDeleted, err := safeDelete("trade_ins")
	if err != nil {
		fmt.Printf("     ⚠️  Warning: %v\n", err)
	} else {
		fmt.Printf("     ✅ Deleted %d trade-ins records\n", tradeInsDeleted)
	}

	// 5. Delete all inventory items
	fmt.Println("   - Deleting inventory items...")
	itemsDeleted, err := safeDelete("inventory_items")
	if err != nil {
		return fmt.Errorf("failed to delete inventory items: %w", err)
	}
	fmt.Printf("     ✅ Deleted %d inventory items\n", itemsDeleted)

	// 6. Delete the old inventory table records (if exists)
	fmt.Println("   - Deleting old inventory records...")
	inventoryDeleted, err := safeDelete("inventory")
	if err != nil {
		fmt.Printf("     ⚠️  Warning: %v\n", err)
	} else {
		fmt.Printf("     ✅ Deleted %d old inventory records\n", inventoryDeleted)
	}

	// Verify the cleanup
	fmt.Println("\n📊 Verification - Remaining records:")
	
	tables := []string{"inventory_movements", "reservations", "item_specification_values", "trade_ins", "inventory_items", "inventory"}
	for _, table := range tables {
		var count int
		err := db.Get(&count, fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
		if err != nil {
			fmt.Printf("   - %s: Table does not exist\n", table)
		} else {
			fmt.Printf("   - %s: %d\n", table, count)
		}
	}

	return nil
}