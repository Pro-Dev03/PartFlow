package middleware

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/repositories"
)

// RepositoryFactoryMiddleware injects the appropriate repository factory based on operating mode
func RepositoryFactoryMiddleware(postgresDB *sqlx.DB, sqliteDB *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get operating mode from SQLite metadata
		mode, err := getOperatingMode(sqliteDB)
		if err != nil {
			log.Printf("Error reading operating mode: %v", err)
			mode = "online" // Default to online if error
		}

		// Create factory
		factory := repositories.NewRepositoryFactory(postgresDB, sqliteDB, mode)

		// Store in context
		c.Set("repositoryFactory", factory)
		c.Set("operatingMode", mode)

		c.Next()
	}
}

// getOperatingMode reads the operating mode from SQLite metadata
func getOperatingMode(db *sql.DB) (string, error) {
	var mode string
	query := "SELECT value FROM local_metadata WHERE key = 'operating_mode' LIMIT 1"
	err := db.QueryRow(query).Scan(&mode)
	if err != nil {
		// If not found, assume online mode
		if err == sql.ErrNoRows {
			return "online", nil
		}
		return "", err
	}
	return mode, nil
}

// GetRepositoryFactory retrieves the factory from context
func GetRepositoryFactory(c *gin.Context) repositories.RepositoryFactory {
	factory, exists := c.Get("repositoryFactory")
	if !exists {
		log.Fatal("Repository factory not found in context")
	}
	return factory.(repositories.RepositoryFactory)
}

// GetOperatingMode retrieves the operating mode from context
func GetOperatingMode(c *gin.Context) string {
	mode, exists := c.Get("operatingMode")
	if !exists {
		return "online" // Default
	}
	return mode.(string)
}
