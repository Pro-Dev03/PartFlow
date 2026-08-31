package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var jwtSecret = []byte("your-secret-key-change-in-production")

var db *sqlx.DB
var disableAuth = false

const defaultCloudAPIURL = "https://partflow-api.onrender.com/api/v1"

type cloudAuthError struct {
	status int
	err    error
}

func requiresCloudAuth() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("PARTFLOW_REQUIRE_CLOUD_AUTH")))
	if value == "" {
		return true
	}
	return value == "1" || value == "true" || value == "yes"
}

// validateWithCloud keeps the local SQLite API subject to the same cloud
// account decision as the renderer. The local service never receives cloud
// business data; it only forwards the bearer token for account validation.
func validateWithCloud(ctx context.Context, tokenString string) (uuid.UUID, *cloudAuthError) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PARTFLOW_CLOUD_API_URL")), "/")
	if baseURL == "" {
		baseURL = defaultCloudAPIURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/auth/validate", strings.NewReader("{}"))
	if err != nil {
		return uuid.Nil, &cloudAuthError{status: http.StatusServiceUnavailable, err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return uuid.Nil, &cloudAuthError{status: http.StatusServiceUnavailable, err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status := http.StatusServiceUnavailable
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			status = resp.StatusCode
		}
		return uuid.Nil, &cloudAuthError{status: status, err: fmt.Errorf("cloud validation returned HTTP %d", resp.StatusCode)}
	}

	var envelope struct {
		Data struct {
			Valid bool `json:"valid"`
			User  struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return uuid.Nil, &cloudAuthError{status: http.StatusServiceUnavailable, err: err}
	}
	if !envelope.Data.Valid {
		return uuid.Nil, &cloudAuthError{status: http.StatusForbidden, err: fmt.Errorf("cloud account is not valid")}
	}
	userID, err := uuid.Parse(envelope.Data.User.ID)
	if err != nil {
		return uuid.Nil, &cloudAuthError{status: http.StatusUnauthorized, err: fmt.Errorf("cloud response did not contain a valid user id")}
	}
	return userID, nil
}

func parseSubscriptionExpiry(value interface{}) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}

	if parsed, ok := value.(time.Time); ok {
		return &parsed, nil
	}

	raw, ok := value.(string)
	if !ok {
		if bytes, bytesOK := value.([]byte); bytesOK {
			raw = string(bytes)
		} else {
			return nil, fmt.Errorf("unsupported subscription expiry type %T", value)
		}
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if idx := strings.Index(raw, " m="); idx > 0 {
		raw = strings.TrimSpace(raw[:idx])
	}

	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return &parsed, nil
		}
	}

	return nil, fmt.Errorf("unsupported subscription expiry format %q", raw)
}

// SetDatabase sets the database connection for middleware
func SetDatabase(database *sqlx.DB) {
	db = database
}

// SetDisableAuth sets the disable auth flag for development
func SetDisableAuth(disable bool) {
	disableAuth = disable
}

func isLocalDatabaseMode() bool {
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("DB_CONNECTION_MODE")))
	if mode == "" {
		mode = strings.TrimSpace(strings.ToLower(os.Getenv("DATABASE_MODE")))
	}
	if mode == "local" || mode == "sqlite" {
		return true
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	return strings.HasPrefix(databaseURL, "sqlite://")
}

func ensureUserAuthorized(ctx context.Context, userUUID uuid.UUID) (bool, error) {
	if db == nil {
		return true, nil
	}

	var userExists bool
	err := db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userUUID).Scan(&userExists)
	if err == nil {
		return userExists, nil
	}

	msg := err.Error()
	if isLocalDatabaseMode() || strings.Contains(msg, "no such table: users") || strings.Contains(msg, "does not exist") {
		return true, nil
	}

	return false, err
}

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Auth middleware for JWT authentication (based on worktrack)
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication if disabled (development mode)
		if disableAuth {
			// Use a real user when available so endpoints with non-null user
			// references (inspections and seller payments) remain usable.
			// Fall back to the zero UUID for installations without users yet.
			userID := uuid.Nil
			if db != nil {
				_ = db.QueryRowContext(c.Request.Context(),
					"SELECT id FROM users ORDER BY created_at LIMIT 1").Scan(&userID)
			}
			c.Set("user_id", userID)
			c.Set("user_id_string", userID.String())
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		// The embedded/local API uses SQLite for business data but delegates account
		// authorization to Render. This prevents a direct local API call from
		// bypassing the cloud subscription decision.
		if isLocalDatabaseMode() && requiresCloudAuth() {
			userUUID, cloudErr := validateWithCloud(c.Request.Context(), tokenString)
			if cloudErr != nil {
				message := "تعذر التحقق من الحساب عبر الخادم السحابي"
				if cloudErr.status == http.StatusUnauthorized {
					message = "جلسة الدخول غير صالحة"
				} else if cloudErr.status == http.StatusForbidden {
					message = "الحساب غير نشط أو أن الاشتراك منتهٍ"
				}
				c.JSON(cloudErr.status, gin.H{"error": message, "code": "CLOUD_AUTH_REQUIRED"})
				c.Abort()
				return
			}
			c.Set("user_id", userUUID)
			c.Set("user_id_string", userUUID.String())
			c.Next()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Extract user_id from claims
			userID, ok := claims["user_id"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
				c.Abort()
				return
			}

			// Parse user ID as UUID
			userUUID, err := uuid.Parse(userID)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
				c.Abort()
				return
			}

			allowsUser, err := ensureUserAuthorized(c.Request.Context(), userUUID)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				c.Abort()
				return
			}
			if !allowsUser && !isLocalDatabaseMode() {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				c.Abort()
				return
			}

			if !isLocalDatabaseMode() && db != nil {
				var subscriptionStatus string
				var subscriptionExpiresAtRaw interface{}
				err = db.QueryRowContext(c.Request.Context(),
					"SELECT subscription_status, subscription_expires_at FROM users WHERE id = $1", userUUID).
					Scan(&subscriptionStatus, &subscriptionExpiresAtRaw)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to verify subscription status"})
					c.Abort()
					return
				}
				subscriptionExpiresAt, err := parseSubscriptionExpiry(subscriptionExpiresAtRaw)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to verify subscription status"})
					c.Abort()
					return
				}

				if subscriptionStatus == "canceled" || subscriptionStatus == "cancelled" || subscriptionStatus == "expired" || (subscriptionExpiresAt != nil && time.Now().After(*subscriptionExpiresAt)) {
					c.JSON(http.StatusForbidden, gin.H{
						"error": "اشتراكك منتهي، يرجى التواصل مع الإدارة لتجديد الخدمة.",
					})
					c.Abort()
					return
				}
			}

			// Set user context
			c.Set("user_id", userUUID)
			c.Set("user_id_string", userID) // Keep string version for compatibility
		}

		c.Next()
	}
}

// Organization middleware removed - single-tenant system

// Logger middleware with structured logging
func Logger() gin.HandlerFunc {
	return gin.Logger()
}

// Recovery middleware
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// GetRequestID gets the request ID from context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// RateLimiter middleware (basic implementation)
func RateLimiter() gin.HandlerFunc {
	// TODO: Implement proper rate limiting with Redis
	// For now, this is a placeholder
	return func(c *gin.Context) {
		c.Next()
	}
}

// SetJWTSecret sets the JWT secret key
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}
