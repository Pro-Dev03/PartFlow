package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
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

// Cloud validation is intentionally cached only for a very short period. The
// cloud remains the source of truth (a revoked account is rejected on the
// next validation window), while requests made by the dashboard do not cause
// a validation request for every single local endpoint.
const cloudValidationCacheTTL = 15 * time.Second

// Keep validation traffic bounded when several dashboard requests arrive at
// once. Render/Supabase can briefly return 502/503 under a burst; one small
// queue is enough for local requests and avoids turning that burst into a
// cascade of authentication failures.
var cloudValidationSlots = make(chan struct{}, 1)

type cloudAuthError struct {
	status int
	err    error
}

type cloudValidationCacheEntry struct {
	userID    uuid.UUID
	email     string
	expiresAt time.Time
}

type cloudValidationCall struct {
	done    chan struct{}
	userID  uuid.UUID
	email   string
	authErr *cloudAuthError
}

var cloudValidationMu sync.Mutex
var cloudValidationCache = make(map[string]cloudValidationCacheEntry)
var cloudValidationInFlight = make(map[string]*cloudValidationCall)

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
func validateWithCloud(ctx context.Context, tokenString string) (uuid.UUID, string, *cloudAuthError) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PARTFLOW_CLOUD_API_URL")), "/")
	if baseURL == "" {
		baseURL = defaultCloudAPIURL
	}
	keyBytes := sha256.Sum256([]byte(baseURL + "\x00" + tokenString))
	cacheKey := fmt.Sprintf("%x", keyBytes[:])

	now := time.Now()
	cloudValidationMu.Lock()
	if cached, ok := cloudValidationCache[cacheKey]; ok {
		if now.Before(cached.expiresAt) {
			cloudValidationMu.Unlock()
			return cached.userID, cached.email, nil
		}
		delete(cloudValidationCache, cacheKey)
	}
	if call, ok := cloudValidationInFlight[cacheKey]; ok {
		cloudValidationMu.Unlock()
		select {
		case <-call.done:
			return call.userID, call.email, call.authErr
		case <-ctx.Done():
			return uuid.Nil, "", &cloudAuthError{status: http.StatusServiceUnavailable, err: ctx.Err()}
		}
	}
	call := &cloudValidationCall{done: make(chan struct{})}
	cloudValidationInFlight[cacheKey] = call
	cloudValidationMu.Unlock()

	userID, email, authErr := validateWithCloudRemote(ctx, baseURL, tokenString)
	cloudValidationMu.Lock()
	delete(cloudValidationInFlight, cacheKey)
	call.userID = userID
	call.email = email
	call.authErr = authErr
	if authErr == nil {
		cloudValidationCache[cacheKey] = cloudValidationCacheEntry{
			userID:    userID,
			email:     email,
			expiresAt: time.Now().Add(cloudValidationCacheTTL),
		}
	}
	close(call.done)
	cloudValidationMu.Unlock()
	return userID, email, authErr
}

func validateWithCloudRemote(ctx context.Context, baseURL, tokenString string) (uuid.UUID, string, *cloudAuthError) {
	select {
	case cloudValidationSlots <- struct{}{}:
		defer func() { <-cloudValidationSlots }()
	case <-ctx.Done():
		return uuid.Nil, "", &cloudAuthError{status: http.StatusServiceUnavailable, err: ctx.Err()}
	}
	return validateWithCloudRemoteOnce(ctx, baseURL, tokenString)
}

func validateWithCloudRemoteOnce(ctx context.Context, baseURL, tokenString string) (uuid.UUID, string, *cloudAuthError) {

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/auth/validate", strings.NewReader("{}"))
	if err != nil {
		return uuid.Nil, "", &cloudAuthError{status: http.StatusServiceUnavailable, err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return uuid.Nil, "", &cloudAuthError{status: http.StatusServiceUnavailable, err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status := http.StatusServiceUnavailable
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			status = resp.StatusCode
		}
		return uuid.Nil, "", &cloudAuthError{status: status, err: fmt.Errorf("cloud validation returned HTTP %d", resp.StatusCode)}
	}

	var envelope struct {
		Data struct {
			Valid bool `json:"valid"`
			User  struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return uuid.Nil, "", &cloudAuthError{status: http.StatusServiceUnavailable, err: err}
	}
	if !envelope.Data.Valid {
		return uuid.Nil, "", &cloudAuthError{status: http.StatusForbidden, err: fmt.Errorf("cloud account is not valid")}
	}
	userID, err := uuid.Parse(envelope.Data.User.ID)
	if err != nil {
		return uuid.Nil, "", &cloudAuthError{status: http.StatusUnauthorized, err: fmt.Errorf("cloud response did not contain a valid user id")}
	}
	return userID, strings.TrimSpace(envelope.Data.User.Email), nil
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
	allowedOrigins := configuredCORSOrigins()
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.Request.Header.Get("Origin"))
		allowed := origin == "" || corsOriginAllowed(origin, allowedOrigins)
		if origin != "" && corsOriginAllowed(origin, allowedOrigins) {
			// Reflect only an explicitly allowed origin. This is important when
			// Authorization headers are used by the browser client.
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Add("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-PartFlow-Cloud-Token, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			if !allowed {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func configuredCORSOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" {
		// Development and Electron defaults. Production should set an explicit
		// comma-separated allowlist in Render/environment configuration.
		raw = "http://localhost:5173,http://127.0.0.1:5173,http://localhost:3000,http://127.0.0.1:3000,null"
	}
	origins := make([]string, 0)
	for _, value := range strings.Split(raw, ",") {
		if value = strings.TrimSpace(value); value != "" {
			origins = append(origins, value)
		}
	}
	return origins
}

func corsOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || strings.EqualFold(allowed, origin) {
			return true
		}
	}
	return false
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
			cloudToken := strings.TrimSpace(c.GetHeader("X-PartFlow-Cloud-Token"))
			if cloudToken == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "يلزم توكن الجلسة السحابية للتحقق من الاشتراك",
					"code":  "CLOUD_AUTH_REQUIRED",
				})
				c.Abort()
				return
			}
			userUUID, cloudEmail, cloudErr := validateWithCloud(c.Request.Context(), cloudToken)
			if cloudErr != nil {
				message := "تعذر التحقق من الحساب عبر الخادم السحابي"
				switch cloudErr.status {
				case http.StatusUnauthorized:
					message = "جلسة الدخول غير صالحة"
				case http.StatusForbidden:
					message = "الحساب غير نشط أو أن الاشتراك منتهٍ"
				}
				c.JSON(cloudErr.status, gin.H{"error": message, "code": "CLOUD_AUTH_REQUIRED"})
				c.Abort()
				return
			}
			c.Set("user_id", userUUID)
			c.Set("user_id_string", userUUID.String())
			if cloudEmail != "" {
				c.Set("cloud_user_email", cloudEmail)
			}
			c.Next()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// The local fallback accepts only the configured HMAC algorithm.  In
			// particular, never let a token choose an RSA/none algorithm while a
			// shared HMAC secret is being used.
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}
			return jwtSecret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil || token == nil || !token.Valid {
			message := "Invalid token"
			if err != nil {
				message += ": " + err.Error()
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": message})
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

type rateLimitBucket struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
}

var rateLimitBuckets sync.Map

func rateLimitSetting(name string, fallback float64) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv(name)), 64)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

// RateLimiter applies a small in-process token bucket keyed by client IP.
// It deliberately has no Redis dependency so the embedded/local service is
// protected as well. Deployments with multiple API replicas should still put
// a shared limiter at the edge (or use one process per tenant).
func RateLimiter() gin.HandlerFunc {
	rps := rateLimitSetting("RATE_LIMIT_RPS", 100)
	burst := rateLimitSetting("RATE_LIMIT_BURST", 10)
	if burst < 1 {
		burst = 1
	}

	return func(c *gin.Context) {
		key := c.ClientIP()
		value, _ := rateLimitBuckets.LoadOrStore(key, &rateLimitBucket{tokens: burst, last: time.Now()})
		bucket := value.(*rateLimitBucket)
		allowed, remaining, retryAfter := func() (bool, int, int) {
			bucket.mu.Lock()
			defer bucket.mu.Unlock()
			now := time.Now()
			elapsed := now.Sub(bucket.last).Seconds()
			if elapsed > 0 {
				bucket.tokens = minFloat(burst, bucket.tokens+elapsed*rps)
				bucket.last = now
			}
			if bucket.tokens < 1 {
				wait := int((1 - bucket.tokens) / rps)
				if wait < 1 {
					wait = 1
				}
				return false, 0, wait
			}
			bucket.tokens--
			return true, int(bucket.tokens), 0
		}()
		c.Header("X-RateLimit-Limit", strconv.Itoa(int(rps)))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter,
			})
			return
		}
		c.Next()
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// SetJWTSecret sets the JWT secret key
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}
