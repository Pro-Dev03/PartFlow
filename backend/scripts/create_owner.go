package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type OwnerConfig struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
}

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

	// Default owner config
	config := OwnerConfig{
		Email:     getEnv("OWNER_EMAIL", "owner@partflow.com"),
		Password:  getEnv("OWNER_PASSWORD", "Owner123456"),
		FirstName: getEnv("OWNER_FIRST_NAME", "Admin"),
		LastName:  getEnv("OWNER_LAST_NAME", "Owner"),
		Phone:     getEnv("OWNER_PHONE", "+970599000000"),
	}

	// Create or update owner user
	userID, err := createOrUpdateOwner(db, config)
	if err != nil {
		log.Fatalf("Failed to create/update owner: %v", err)
	}
	fmt.Printf("✅ Owner account: %s (ID: %s)\n", config.Email, userID)

	fmt.Println("\n🎉 Owner account created successfully!")
	fmt.Println("📧 Email:", config.Email)
	fmt.Println("🔑 Password:", config.Password)
	fmt.Println("🌐 Login at: http://localhost:5173/")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createOrUpdateOwner(db *sqlx.DB, config OwnerConfig) (string, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(config.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if user exists
	var userID string
	err = db.Get(&userID, "SELECT id FROM users WHERE email = $1 LIMIT 1", config.Email)
	
	if err == nil {
		// Update existing user
		query := `
			UPDATE users 
			SET password_hash = $1, first_name = $2, last_name = $3, phone = $4,
			    role = 'owner', is_active = true,
			    subscription_status = 'active', subscription_expires_at = NOW() + INTERVAL '1 year',
			    updated_at = NOW()
			WHERE email = $5
			RETURNING id
		`
		err = db.QueryRow(query, string(hashedPassword), config.FirstName, config.LastName, 
			config.Phone, config.Email).Scan(&userID)
		if err != nil {
			return "", fmt.Errorf("failed to update user: %w", err)
		}
		return userID, nil
	}

	// Create new user
	userID = uuid.New().String()
	
	// Make sure role column exists
	db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'owner'")
	
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, 
		                  phone, role, is_active, subscription_status, subscription_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'owner', true, 'active', NOW() + INTERVAL '1 year', NOW(), NOW())
		RETURNING id
	`

	err = db.QueryRow(query,
		userID, config.Email, string(hashedPassword),
		config.FirstName, config.LastName, config.Phone,
	).Scan(&userID)

	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}