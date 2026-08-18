package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== Security Review for PartFlow Backend ===\n")

	// Check DATABASE_URL security
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	fmt.Println("🔒 Database Connection Security:")
	
	// Check SSL mode
	if strings.Contains(dbURL, "sslmode=require") {
		fmt.Println("   ✅ SSL mode: require (secure)")
	} else if strings.Contains(dbURL, "sslmode=disable") {
		fmt.Println("   ❌ SSL mode: disable (INSECURE!)")
	} else {
		fmt.Println("   ⚠️  SSL mode: not explicitly set")
	}

	// Check if using pooler
	if strings.Contains(dbURL, "pooler.supabase.com") {
		if strings.Contains(dbURL, ":6543") {
			fmt.Println("   ✅ Using Session Pooler (port 6543) - secure for long connections")
		} else if strings.Contains(dbURL, ":5432") {
			fmt.Println("   ⚠️  Using Transaction Pooler (port 5432) - may have limitations")
		}
	}

	// Check JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	fmt.Println("\n🔒 JWT Security:")
	if jwtSecret == "" {
		fmt.Println("   ❌ JWT_SECRET not set")
	} else if jwtSecret == "your-jwt-secret-change-in-production" {
		fmt.Println("   ❌ JWT_SECRET is default value (INSECURE!)")
	} else if len(jwtSecret) < 16 {
		fmt.Println("   ⚠️  JWT_SECRET is too short (should be at least 16 characters)")
	} else {
		fmt.Println("   ✅ JWT_SECRET is set")
	}

	// Check CORS settings
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	fmt.Println("\n🔒 CORS Security:")
	if corsOrigins == "*" {
		fmt.Println("   ⚠️  CORS allows all origins (*) - consider restricting in production")
	} else {
		fmt.Println("   ✅ CORS origins are restricted")
	}

	// Check server mode
	serverMode := os.Getenv("SERVER_MODE")
	fmt.Println("\n🔒 Server Mode:")
	if serverMode == "debug" {
		fmt.Println("   ⚠️  Server running in debug mode - not recommended for production")
	} else if serverMode == "release" {
		fmt.Println("   ✅ Server running in release mode")
	}

	// Check Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	fmt.Println("\n🔒 Supabase Configuration:")
	if supabaseURL != "" && supabaseKey != "" {
		fmt.Println("   ✅ Supabase credentials configured")
		if strings.HasPrefix(supabaseKey, "sbp_") {
			fmt.Println("   ⚠️  Using anon key - consider using service role key for backend operations")
		}
	}

	// Check rate limiting
	rateLimitEnabled := os.Getenv("RATE_LIMIT_ENABLED")
	fmt.Println("\n🔒 Rate Limiting:")
	if rateLimitEnabled == "true" {
		fmt.Println("   ✅ Rate limiting enabled")
	} else {
		fmt.Println("   ⚠️  Rate limiting disabled - consider enabling for production")
	}

	fmt.Println("\n=== Security Review Complete ===")
	fmt.Println("\n📋 Recommendations:")
	fmt.Println("1. Change JWT_SECRET to a strong, random value")
	fmt.Println("2. Restrict CORS_ALLOWED_ORIGINS to specific domains in production")
	fmt.Println("3. Change SERVER_MODE to 'release' in production")
	fmt.Println("4. Consider enabling rate limiting (RATE_LIMIT_ENABLED=true)")
	fmt.Println("5. Review Supabase key permissions")
	fmt.Println("6. Ensure SSL mode is always 'require'")
}
