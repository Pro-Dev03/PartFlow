package config

import (
	"fmt"
	"os"
	"strconv"
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
	DatabaseURL       string
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

	cfg := &Config{
		// Server
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		ServerMode:         getEnv("SERVER_MODE", "debug"),
		ReadTimeout:       getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:      getDurationEnv("WRITE_TIMEOUT", 15*time.Second),

		// Database
		DatabaseURL:       getEnv("DATABASE_URL", ""),
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
		CORSAllowedHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-Organization-ID"},

		// Rate Limiting
		RateLimitEnabled:   getBoolEnv("RATE_LIMIT_ENABLED", false),
		RateLimitRPS:       getIntEnv("RATE_LIMIT_RPS", 100),
		RateLimitBurst:     getIntEnv("RATE_LIMIT_BURST", 10),

		// File Upload
		MaxUploadSize:      getInt64Env("MAX_UPLOAD_SIZE", 10*1024*1024), // 10MB
		AllowedFileTypes:   []string{"image/jpeg", "image/png", "image/gif", "application/pdf"},

		// Timezone
		DefaultTimezone:    getEnv("DEFAULT_TIMEZONE", "Asia/Jerusalem"),

		// Currency
		DefaultCurrency:    getEnv("DEFAULT_CURRENCY", "ILS"),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "change-this-secret-in-production" && cfg.ServerMode == "release" {
		return nil, fmt.Errorf("JWT_SECRET must be changed in production")
	}

	return cfg, nil
}

// Helper functions

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
