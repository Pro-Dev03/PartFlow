package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
	partflowdb "github.com/partflow/smart-store/pkg/database"
	"github.com/partflow/smart-store/pkg/logger"
)

var jwtSecret = []byte("your-secret-key-change-in-production")

var db *sqlx.DB
var disableAuth = false

// The current repositories share one *sqlx.DB and many do not bind a
// transaction to the request. Until that is refactored, a single PostgreSQL
// session plus this gate keeps a tenant setting from leaking between requests.
var tenantRequestMu sync.Mutex

type tenantContextKey struct{}

func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(tenantContextKey{}).(uuid.UUID)
	return tenantID, ok && tenantID != uuid.Nil
}

// TenantScope resolves tenant ownership only from the authenticated account,
// applies the RLS session setting for the full request, and always clears it.
// The database package limits the pool to one connection while this mode is on.
func TenantScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !partflowdb.TenantIsolationEnabled() {
			// The deployed schema has global business tables until migration 080
			// is applied. Do not let ordinary subscribers reach that shared data
			// during the rollout window.
			if !isLocalDatabaseMode() && !IsConfiguredAdmin(c, GetUserID(c)) {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tenant isolation must be installed before subscriber business access", "code": "TENANT_ISOLATION_REQUIRED"})
				c.Abort()
				return
			}
			c.Next()
			return
		}
		if isLocalDatabaseMode() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tenant isolation requires the cloud PostgreSQL database", "code": "TENANT_DATABASE_REQUIRED"})
			c.Abort()
			return
		}
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tenant database is unavailable", "code": "TENANT_DATABASE_UNAVAILABLE"})
			c.Abort()
			return
		}

		userID := GetUserID(c)
		if userID == uuid.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authenticated account is required", "code": "AUTHENTICATION_REQUIRED"})
			c.Abort()
			return
		}

		tenantRequestMu.Lock()
		defer tenantRequestMu.Unlock()
		defer func() {
			resetContext, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			var cleared string
			if err := db.GetContext(resetContext, &cleared,
				`SELECT set_config('partflow.tenant_id', '', false) || set_config('partflow.user_id', '', false)`); err != nil {
				// A session with an unknown tenant must never return to the pool.
				log.Printf("failed to clear tenant RLS session context; closing database pool: %v", err)
				_ = db.Close()
			}
		}()

		var appliedUser string
		if err := db.GetContext(c.Request.Context(), &appliedUser,
			`SELECT set_config('partflow.user_id', $1, false)`, userID.String()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unable to establish account access", "code": "TENANT_SCOPE_UNAVAILABLE"})
			c.Abort()
			return
		}

		var tenantID uuid.UUID
		if err := db.GetContext(c.Request.Context(), &tenantID,
			`SELECT tenant_id FROM tenant_memberships WHERE user_id = $1`, userID); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusForbidden, gin.H{"error": "Account is not assigned to a store", "code": "TENANT_NOT_CONFIGURED"})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unable to resolve store access", "code": "TENANT_LOOKUP_UNAVAILABLE"})
			}
			c.Abort()
			return
		}
		if tenantID == uuid.Nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is not assigned to a store", "code": "TENANT_NOT_CONFIGURED"})
			c.Abort()
			return
		}

		var applied string
		if err := db.GetContext(c.Request.Context(), &applied,
			`SELECT set_config('partflow.tenant_id', $1, false)`, tenantID.String()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unable to establish store access", "code": "TENANT_SCOPE_UNAVAILABLE"})
			c.Abort()
			return
		}

		c.Set("tenant_id", tenantID)
		requestContext := context.WithValue(c.Request.Context(), tenantContextKey{}, tenantID)
		c.Request = c.Request.WithContext(requestContext)
		c.Next()
	}
}

func allowLocalAuthBypass() bool {
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("SERVER_MODE")))
	if mode == "" {
		mode = strings.TrimSpace(strings.ToLower(os.Getenv("APP_ENV")))
	}
	if mode == "release" || mode == "production" {
		return false
	}

	flag := strings.TrimSpace(strings.ToLower(os.Getenv("PARTFLOW_ALLOW_LOCAL_AUTH_BYPASS")))
	if flag == "1" || flag == "true" || flag == "yes" {
		return true
	}

	return mode == "debug" || mode == "development" || mode == "test"
}

const defaultCloudAPIURL = "https://partflow-api.onrender.com/api/v1"

// Keep validation traffic bounded when several dashboard requests arrive at
// once. Render/Supabase can briefly return 502/503 under a burst; one small
// queue is enough for local requests and avoids turning that burst into a
// cascade of authentication failures.
var cloudValidationSlots = make(chan struct{}, 1)

type cloudAuthError struct {
	status int
	code   string
	err    error
}

type cloudValidationCall struct {
	done    chan struct{}
	userID  uuid.UUID
	email   string
	authErr *cloudAuthError
}

var cloudValidationMu sync.Mutex
var cloudValidationInFlight = make(map[string]*cloudValidationCall)

func requiresCloudAuth() bool {
	// A local database is an operational cache, not an authentication bypass.
	if isLocalDatabaseMode() {
		return true
	}

	mode := strings.TrimSpace(strings.ToLower(os.Getenv("SERVER_MODE")))
	if mode == "" {
		mode = strings.TrimSpace(strings.ToLower(os.Getenv("APP_ENV")))
	}
	if mode == "release" || mode == "production" {
		return true
	}

	value := strings.TrimSpace(strings.ToLower(os.Getenv("PARTFLOW_REQUIRE_CLOUD_AUTH")))
	if value == "" {
		return false
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
	cacheKey := baseURL + "\x00" + tokenString

	cloudValidationMu.Lock()
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
	req.Header.Set("X-PartFlow-Cloud-Token", tokenString)

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
		var rejection struct {
			Code  string `json:"code"`
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&rejection)
		code := rejection.Code
		if code == "" {
			code = rejection.Error.Code
		}
		return uuid.Nil, "", &cloudAuthError{status: status, code: code, err: fmt.Errorf("cloud validation returned HTTP %d", resp.StatusCode)}
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

const accessTokenTimestampTolerance = 5 * time.Second

// accessTokenIsCurrent rejects access JWTs minted before the account's latest
// update. Password changes update users.updated_at in the same transaction as
// refresh-token revocation, so already-issued access tokens stop working too.
// The small tolerance accounts for JWT NumericDate's second precision and
// minor clock skew between the API process and PostgreSQL.
func accessTokenIsCurrent(claims jwt.MapClaims, updatedAtRaw interface{}) (bool, error) {
	issuedAtRaw, exists := claims["iat"]
	if !exists {
		return false, nil
	}

	var issuedAt time.Time
	switch value := issuedAtRaw.(type) {
	case float64:
		issuedAt = time.Unix(int64(value), 0)
	case int64:
		issuedAt = time.Unix(value, 0)
	case json.Number:
		seconds, err := value.Int64()
		if err != nil {
			return false, nil
		}
		issuedAt = time.Unix(seconds, 0)
	default:
		return false, nil
	}

	updatedAt, err := parseSubscriptionExpiry(updatedAtRaw)
	if err != nil {
		return false, err
	}
	if updatedAt == nil || issuedAt.After(time.Now().Add(accessTokenTimestampTolerance)) {
		return false, nil
	}

	return !issuedAt.Add(accessTokenTimestampTolerance).Before(*updatedAt), nil
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
	if mode == "cloud" {
		return false
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if strings.HasPrefix(databaseURL, "sqlite://") {
		return true
	}
	if databaseURL != "" {
		return false
	}

	// ResolveDatabaseURL defaults to SQLite when no primary or cloud database
	// URL is configured. Mirror that decision here so auth policy matches the
	// actual database selected by the application.
	if strings.TrimSpace(os.Getenv("DATABASE_URL_CLOUD")) != "" || strings.TrimSpace(os.Getenv("DB_CLOUD_URL")) != "" || strings.TrimSpace(os.Getenv("CLOUD_DATABASE_URL")) != "" {
		return false
	}
	return true
}

func isLoopbackRequest(c *gin.Context) bool {
	host := strings.ToLower(strings.TrimSpace(c.Request.Host))
	return strings.HasPrefix(host, "localhost:") || strings.HasPrefix(host, "127.0.0.1:") || host == "localhost" || host == "127.0.0.1"
}

func ensureUserAuthorized(ctx context.Context, userUUID uuid.UUID) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("authentication database is unavailable")
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
	if isReleaseOrProductionMode() {
		productionOrigins := make([]string, 0, len(allowedOrigins))
		for _, origin := range allowedOrigins {
			if origin == "*" || strings.EqualFold(origin, "null") {
				continue
			}
			productionOrigins = append(productionOrigins, origin)
		}
		allowedOrigins = productionOrigins
	}
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.Request.Header.Get("Origin"))
		allowed := origin == "" || corsOriginAllowed(origin, allowedOrigins)
		if origin != "" && corsOriginAllowed(origin, allowedOrigins) {
			// Reflect only an explicitly allowed origin. This is important when
			// Authorization headers are used by the browser client.
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Add("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-PartFlow-Cloud-Token, Idempotency-Key, accept, origin, Cache-Control, X-Requested-With")
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

func isReleaseOrProductionMode() bool {
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("SERVER_MODE")))
	if mode == "" {
		mode = strings.TrimSpace(strings.ToLower(os.Getenv("APP_ENV")))
	}
	return mode == "release" || mode == "production"
}

func configuredCORSOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" {
		// Development and Electron defaults. Production should set an explicit
		// comma-separated allowlist in Render/environment configuration.
		raw = "https://partflow-hpv7.onrender.com,partflow://app,http://localhost:5173,http://127.0.0.1:5173,http://localhost:5174,http://127.0.0.1:5174,http://localhost:5175,http://127.0.0.1:5175,http://localhost:3000,http://127.0.0.1:3000"
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
			if (requiresCloudAuth() && isLocalDatabaseMode()) || !allowLocalAuthBypass() {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "authentication is disabled only for local development; set PARTFLOW_ALLOW_LOCAL_AUTH_BYPASS=true in non-production",
					"code":  "AUTH_DISABLED",
				})
				c.Abort()
				return
			}
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

		// Local sessions must match an active cloud account before protected
		// operations are authorized.
		cloudToken := strings.TrimSpace(c.GetHeader("X-PartFlow-Cloud-Token"))

		if requiresCloudAuth() && isLocalDatabaseMode() {
			localUserID, localTokenValid := localJWTUserID(tokenString)
			if !localTokenValid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid local session", "code": "INVALID_TOKEN"})
				c.Abort()
				return
			}
			if cloudToken == "" {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Cloud connection is required", "code": "CLOUD_CONNECTION_REQUIRED"})
				c.Abort()
				return
			}
			userUUID, cloudEmail, cloudErr := validateWithCloud(c.Request.Context(), cloudToken)
			if cloudErr != nil {
				message := "Cloud account validation failed"
				code := cloudErr.code
				if code == "" || cloudErr.status == http.StatusServiceUnavailable {
					code = "CLOUD_CONNECTION_REQUIRED"
				}
				c.JSON(cloudErr.status, gin.H{"error": message, "code": code})
				c.Abort()
				return
			}
			if localUserID != userUUID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Local and cloud sessions do not match", "code": "SESSION_IDENTITY_MISMATCH"})
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
			// Extract user_id from claims, while also accepting the standard JWT
			// subject claim (sub) for older or cross-service tokens. This keeps the
			// auth contract backwards compatible without allowing arbitrary claim
			// shapes to pass through.
			userIDValue := claims["user_id"]
			if userIDValue == nil {
				userIDValue = claims["sub"]
			}

			userID, ok := userIDValue.(string)
			if !ok {
				logger.Warn("User ID not found in token claims", map[string]interface{}{
					"claim_keys": getClaimsKeys(claims),
				})
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
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unable to verify account", "code": "AUTH_SERVICE_UNAVAILABLE"})
				c.Abort()
				return
			}
			if !allowsUser && !isLocalDatabaseMode() {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found", "code": "ACCOUNT_DELETED"})
				c.Abort()
				return
			}

			if !isLocalDatabaseMode() && db != nil {
				var isActive bool
				var subscriptionStatus string
				var subscriptionExpiresAtRaw interface{}
				var userUpdatedAtRaw interface{}
				err = db.QueryRowContext(c.Request.Context(),
					"SELECT is_active, subscription_status, subscription_expires_at, updated_at FROM users WHERE id = $1", userUUID).
					Scan(&isActive, &subscriptionStatus, &subscriptionExpiresAtRaw, &userUpdatedAtRaw)
				if err != nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unable to verify subscription status", "code": "AUTH_SERVICE_UNAVAILABLE"})
					c.Abort()
					return
				}
				currentSession, sessionCheckErr := accessTokenIsCurrent(claims, userUpdatedAtRaw)
				if sessionCheckErr != nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unable to verify session state", "code": "AUTH_SERVICE_UNAVAILABLE"})
					c.Abort()
					return
				}
				if !currentSession {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Session has been revoked", "code": "SESSION_REVOKED"})
					c.Abort()
					return
				}
				subscriptionExpiresAt, err := parseSubscriptionExpiry(subscriptionExpiresAtRaw)
				if err != nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unable to verify subscription status", "code": "AUTH_SERVICE_UNAVAILABLE"})
					c.Abort()
					return
				}

				normalizedStatus := strings.ToLower(strings.TrimSpace(subscriptionStatus))
				if !isActive || (normalizedStatus != "active" && normalizedStatus != "trial") || (subscriptionExpiresAt != nil && !time.Now().Before(*subscriptionExpiresAt)) {
					code := "SUBSCRIPTION_EXPIRED"
					if normalizedStatus == "suspended" || !isActive {
						code = "SUBSCRIPTION_SUSPENDED"
					} else if normalizedStatus == "deleted" {
						code = "ACCOUNT_DELETED"
					}
					c.JSON(http.StatusForbidden, gin.H{
						"error": "اشتراكك منتهي، يرجى التواصل مع الإدارة لتجديد الخدمة.",
						"code":  code,
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

func localJWTUserID(tokenString string) (uuid.UUID, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || token == nil || !token.Valid {
		return uuid.Nil, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, false
	}
	userID, _ := claims["user_id"].(string)
	if userID == "" {
		userID, _ = claims["sub"].(string)
	}
	parsed, err := uuid.Parse(userID)
	return parsed, err == nil
}

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
	burst := rateLimitSetting("RATE_LIMIT_BURST", 50)
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

// getClaimsKeys returns the keys in the JWT claims map
func getClaimsKeys(claims jwt.MapClaims) []string {
	keys := make([]string, 0, len(claims))
	for k := range claims {
		keys = append(keys, k)
	}
	return keys
}

// SetJWTSecret sets the JWT secret key
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}
