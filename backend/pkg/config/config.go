package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config represents application configuration
type Config struct {
	// Server
	ServerPort         string
	ServerMode         string // debug, release, test
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration

	// Database
	DatabaseMode      string
	DatabaseSource    string
	DatabaseURL       string
	DatabaseURLLocal  string
	DatabaseURLCloud  string
	DatabaseMaxOpenConns int
	DatabaseMaxIdleConns int
	DatabaseConnMaxLifetime time.Duration

	// Supabase
	SupabaseURL       string
	SupabaseKey       string
	UseSupabaseAuth   bool

	// JWT
	JWTSecret         string
	JWTAccessTokenTTL time.Duration
	JWTRefreshTokenTTL time.Duration

	// Redis
	RedisURL          string
	RedisPassword     string
	RedisDB           int

	// Logging
	LogLevel          string
	LogFormat         string // json, text

	// CORS
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string

	// Rate Limiting
	RateLimitEnabled   bool
	RateLimitRPS       int // requests per second
	RateLimitBurst     int

	// Auth
	DisableAuth       bool // Disable authentication for development

	// File Upload
	MaxUploadSize      int64
	AllowedFileTypes   []string

	// Timezone
	DefaultTimezone    string

	// Currency
	DefaultCurrency    string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	godotenv.Load()

	selectedURL, selectedSource, err := ResolveDatabaseURL()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		// Server
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		ServerMode:         getEnv("SERVER_MODE", "debug"),
		ReadTimeout:       getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:      getDurationEnv("WRITE_TIMEOUT", 15*time.Second),

		// Database
		DatabaseMode:      getEnv("DB_CONNECTION_MODE", getEnv("DATABASE_MODE", "default")),
		DatabaseSource:    selectedSource,
		DatabaseURL:       selectedURL,
		DatabaseURLLocal:  getEnv("DATABASE_URL_LOCAL", getEnv("DB_LOCAL_URL", getEnv("LOCAL_DATABASE_URL", ""))),
		DatabaseURLCloud:  getEnv("DATABASE_URL_CLOUD", getEnv("DB_CLOUD_URL", getEnv("CLOUD_DATABASE_URL", ""))),
		DatabaseMaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 25),
		DatabaseMaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 5),
		DatabaseConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),

		// Supabase
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseKey:       getEnv("SUPABASE_KEY", ""),
		UseSupabaseAuth:   getBoolEnv("USE_SUPABASE_AUTH", false),

		// JWT
		JWTSecret:         getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTAccessTokenTTL: getDurationEnv("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
		JWTRefreshTokenTTL: getDurationEnv("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour),

		// Redis
		RedisURL:          getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           getIntEnv("REDIS_DB", 0),

		// Logging
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogFormat:         getEnv("LOG_FORMAT", "json"),

		// CORS
		CORSAllowedOrigins: []string{getEnv("CORS_ALLOWED_ORIGINS", "*")},
		CORSAllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		CORSAllowedHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},

		// Rate Limiting
		RateLimitEnabled:   getBoolEnv("RATE_LIMIT_ENABLED", false),
		RateLimitRPS:       getIntEnv("RATE_LIMIT_RPS", 100),
		RateLimitBurst:     getIntEnv("RATE_LIMIT_BURST", 10),

		// Auth
		DisableAuth:       getBoolEnv("DISABLE_AUTH", false),

		// File Upload
		MaxUploadSize:      getInt64Env("MAX_UPLOAD_SIZE", 10*1024*1024), // 10MB
		AllowedFileTypes:   []string{"image/jpeg", "image/png", "image/gif", "application/pdf"},

		// Timezone
		DefaultTimezone:    getEnv("DEFAULT_TIMEZONE", "Asia/Jerusalem"),

		// Currency
		DefaultCurrency:    getEnv("DEFAULT_CURRENCY", "ILS"),
	}

	// Persist the resolved active connection so the rest of the codebase and scripts continue to use it.
	if cfg.DatabaseURL != "" {
		_ = os.Setenv("DATABASE_URL", cfg.DatabaseURL)
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("No database connection configured. Set DATABASE_URL or configure DATABASE_URL_LOCAL / DATABASE_URL_CLOUD together with DB_CONNECTION_MODE")
	}

	if cfg.JWTSecret == "change-this-secret-in-production" && cfg.ServerMode == "release" {
		return nil, fmt.Errorf("JWT_SECRET must be changed in production")
	}

	return cfg, nil
}

// Helper functions

func ResolveDatabaseURL() (string, string, error) {
	mode := strings.TrimSpace(strings.ToLower(getEnv("DB_CONNECTION_MODE", getEnv("DATABASE_MODE", "default"))))
	localCandidates := []string{
		getEnv("DATABASE_URL_LOCAL", ""),
		getEnv("DB_LOCAL_URL", ""),
		getEnv("LOCAL_DATABASE_URL", ""),
	}
	cloudCandidates := []string{
		getEnv("DATABASE_URL_CLOUD", ""),
		getEnv("DB_CLOUD_URL", ""),
		getEnv("CLOUD_DATABASE_URL", ""),
	}
	primaryURL := getEnv("DATABASE_URL", "")

	choose := func(candidates ...string) string {
		for _, candidate := range candidates {
			if strings.TrimSpace(candidate) != "" {
				return strings.TrimSpace(candidate)
			}
		}
		return ""
	}

	switch mode {
	case "local":
		if url := choose(localCandidates...); url != "" {
			return url, "local", nil
		}
		if primaryURL != "" {
			return primaryURL, "database_url", nil
		}
		return "", "", fmt.Errorf("No local database URL configured. Set DATABASE_URL_LOCAL or DB_LOCAL_URL")
	case "cloud":
		if url := choose(cloudCandidates...); url != "" {
			return url, "cloud", nil
		}
		if primaryURL != "" {
			return primaryURL, "database_url", nil
		}
		return "", "", fmt.Errorf("No cloud database URL configured. Set DATABASE_URL_CLOUD or DB_CLOUD_URL")
	case "auto":
		if primaryURL != "" {
			return primaryURL, "database_url", nil
		}
		if url := choose(localCandidates...); url != "" {
			return url, "local", nil
		}
		if url := choose(cloudCandidates...); url != "" {
			return url, "cloud", nil
		}
		return "", "", fmt.Errorf("No database connection configured. Set DATABASE_URL, or configure local/cloud alternatives")
	default:
		if primaryURL != "" {
			return primaryURL, "database_url", nil
		}
		if url := choose(localCandidates...); url != "" {
			return url, "local", nil
		}
		if url := choose(cloudCandidates...); url != "" {
			return url, "cloud", nil
		}
		return "", "", fmt.Errorf("No database connection configured. Set DATABASE_URL, or configure local/cloud alternatives")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
