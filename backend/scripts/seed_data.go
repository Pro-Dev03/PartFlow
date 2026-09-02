//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
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

	// Run seed data
	if err := seedBasicData(db); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	fmt.Println("🎉 Seed data completed successfully!")
}

func seedBasicData(db *sqlx.DB) error {
	fmt.Println("🌱 Seeding basic data...")

	// 1. Ensure roles exist
	fmt.Println("   - Creating default roles...")
	if err := ensureDefaultRoles(db); err != nil {
		return fmt.Errorf("failed to create roles: %w", err)
	}

	// 2. Ensure categories exist
	fmt.Println("   - Creating default categories...")
	if err := ensureDefaultCategories(db); err != nil {
		return fmt.Errorf("failed to create categories: %w", err)
	}

	// 3. Ensure warehouses exist
	fmt.Println("   - Creating default warehouse...")
	if err := ensureDefaultWarehouse(db); err != nil {
		return fmt.Errorf("failed to create warehouse: %w", err)
	}

	return nil
}

func ensureDefaultRoles(db *sqlx.DB) error {
	roles := []struct {
		ID   string
		Name string
	}{
		{"00000000-0000-0000-0000-000000000001", "Admin"},
		{"bf6917b3-e2a7-476c-a604-10d5683cc511", "owner"},
		{"00000000-0000-0000-0000-000000000003", "Staff"},
	}

	for _, role := range roles {
		var count int
		query := "SELECT COUNT(*) FROM roles WHERE id = $1"
		err := db.Get(&count, query, role.ID)
		if err != nil {
			// If error, try to create anyway
			count = 0
		}

		if count == 0 {
			query := `
				INSERT INTO roles (id, name, description, is_system, created_at, updated_at)
				VALUES ($1, $2, $3, true, NOW(), NOW())
				ON CONFLICT (id) DO NOTHING
			`
			_, err = db.Exec(query, role.ID, role.Name, "Default system role")
			if err != nil {
				fmt.Printf("     ⚠️  Warning: Could not create role '%s': %v\n", role.Name, err)
			} else {
				fmt.Printf("     ✅ Role '%s' created\n", role.Name)
			}
		} else {
			fmt.Printf("     ✅ Role '%s' already exists\n", role.Name)
		}
	}

	return nil
}

func ensureDefaultCategories(db *sqlx.DB) error {
	categories := []struct {
		Name        string
		Description string
		Icon        string
		Color       string
	}{
		{"هواتف", "هواتف ذكية وأجهزة لوحية", "smartphone", "#3b82f6"},
		{"لابتوب", "أجهزة الكمبيوتر المحمولة", "laptop", "#8b5cf6"},
		{"كمبيوتر مكتبي", "أجهزة الكمبيوتر المكتبي", "monitor", "#10b981"},
		{"قطع الكمبيوتر", "معالجات، رام، كروت شاشة، لوحات أم", "cpu", "#f59e0b"},
		{"تخزين", "هارد ديسك، SSD، فلاشات", "hard-drive", "#ef4444"},
		{"شاشات", "شاشات الكمبيوتر والتلفزيون", "monitor", "#06b6d4"},
		{"كاميرات", "كاميرات رقمية وكاميرات أمنية", "camera", "#ec4899"},
		{"طابعات", "طابعات وماسحات ضوئية", "printer", "#6366f1"},
		{"شبكات", "راوترات، مودمات، كابلات", "wifi", "#14b8a6"},
		{"إكسسوارات", "سماعات، كيبورد، ماوس، شواحن", "headphones", "#f97316"},
		{"صوتيات", "مكبرات صوت وأنظمة صوتية", "speaker", "#a855f7"},
		{"كيبلات", "كابلات ووصلات متنوعة", "cable", "#64748b"},
	}

	for _, cat := range categories {
		var count int
		query := "SELECT COUNT(*) FROM categories WHERE name = $1"
		err := db.Get(&count, query, cat.Name)
		if err != nil {
			count = 0
		}

		if count == 0 {
			query := `
				INSERT INTO categories (id, name, description, icon, color, created_at, updated_at)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
				ON CONFLICT (name) DO UPDATE SET
					description = EXCLUDED.description,
					icon = EXCLUDED.icon,
					color = EXCLUDED.color,
					updated_at = NOW()
			`
			_, err = db.Exec(query, cat.Name, cat.Description, cat.Icon, cat.Color)
			if err != nil {
				fmt.Printf("     ⚠️  Warning: Could not create category '%s': %v\n", cat.Name, err)
			} else {
				fmt.Printf("     ✅ Category '%s' created\n", cat.Name)
			}
		} else {
			fmt.Printf("     ✅ Category '%s' already exists\n", cat.Name)
		}
	}

	return nil
}

func ensureDefaultWarehouse(db *sqlx.DB) error {
	var count int
	query := "SELECT COUNT(*) FROM warehouses WHERE name = 'المخزن الرئيسي'"
	err := db.Get(&count, query)
	if err != nil {
		count = 0
	}

	if count == 0 {
		query := `
			INSERT INTO warehouses (id, name, address, city, country, is_active, created_at, updated_at)
			VALUES (gen_random_uuid(), 'المخزن الرئيسي', 'Main Storage Location', 'Gaza', 'Palestine', true, NOW(), NOW())
		`
		_, err = db.Exec(query)
		if err != nil {
			fmt.Printf("     ⚠️  Warning: Could not create warehouse: %v\n", err)
		} else {
			fmt.Println("     ✅ Default warehouse created")
		}
	} else {
		fmt.Println("     ✅ Default warehouse already exists")
	}

	return nil
}
