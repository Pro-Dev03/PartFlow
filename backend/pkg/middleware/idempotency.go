package middleware

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IdempotencyMiddleware struct {
	db          *sqlx.DB
	initErr     error
	lastCleanup time.Time
}

var idempotencyMu sync.Mutex
var idempotencyLocks = map[string]*idempotencyKeyLock{}
var idempotencyCache = map[string]struct {
	requestHash  string
	responseCode int
	responseBody []byte
	expiresAt    time.Time
}{}

type idempotencyKeyLock struct {
	mu   sync.Mutex
	refs int
}

func lockIdempotencyKey(key string) func() {
	idempotencyMu.Lock()
	keyLock := idempotencyLocks[key]
	if keyLock == nil {
		keyLock = &idempotencyKeyLock{}
		idempotencyLocks[key] = keyLock
	}
	keyLock.refs++
	idempotencyMu.Unlock()

	keyLock.mu.Lock()
	return func() {
		keyLock.mu.Unlock()
		idempotencyMu.Lock()
		keyLock.refs--
		if keyLock.refs == 0 {
			delete(idempotencyLocks, key)
		}
		idempotencyMu.Unlock()
	}
}

func NewIdempotencyMiddleware(database interface{}) *IdempotencyMiddleware {
	var db *sqlx.DB
	switch value := database.(type) {
	case *sqlx.DB:
		db = value
	case *sql.DB:
		db = sqlx.NewDb(value, "sqlite")
	default:
		panic("unsupported database type for idempotency middleware")
	}
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS idempotency_keys (
		id TEXT PRIMARY KEY,
		idempotency_key TEXT NOT NULL UNIQUE,
		resource_type TEXT NOT NULL,
		request_hash TEXT NOT NULL,
		response_code INTEGER NOT NULL,
		response_body TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	return &IdempotencyMiddleware{db: db, initErr: err}
}

// Idempotency handles idempotent requests
func (im *IdempotencyMiddleware) Idempotency() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to POST, PUT, PATCH, DELETE requests
		if c.Request.Method != "POST" && c.Request.Method != "PUT" &&
			c.Request.Method != "PATCH" && c.Request.Method != "DELETE" {
			c.Next()
			return
		}

		// Get idempotency key from header
		idempotencyKey := c.GetHeader("Idempotency-Key")
		if idempotencyKey == "" {
			const code = "IDEMPOTENCY_KEY_REQUIRED"
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"code":    code,
				"error": gin.H{
					"code":    code,
					"message": "Idempotency-Key is required for sale creation.",
				},
			})
			return
		}
		if im.initErr != nil {
			abortIdempotencyStoreUnavailable(c)
			return
		}
		// Read request body for hashing
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			c.Abort()
			return
		}

		// Restore request body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		// Calculate request hash
		requestHash := calculateHash(body)
		unlockKey := lockIdempotencyKey(idempotencyKey)
		defer unlockKey()
		im.cleanupExpiredEntries()

		idempotencyMu.Lock()
		cached, ok := idempotencyCache[idempotencyKey]
		if ok && !time.Now().Before(cached.expiresAt) {
			delete(idempotencyCache, idempotencyKey)
			ok = false
		}
		idempotencyMu.Unlock()
		if ok {
			if cached.requestHash != requestHash {
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "idempotency key was already used with a different request"})
				return
			}
			c.Data(cached.responseCode, "application/json", cached.responseBody)
			c.Abort()
			return
		}

		// Check if idempotency key exists
		var existingRecord struct {
			RequestHash  string          `db:"request_hash"`
			ResponseCode int             `db:"response_code"`
			ResponseBody json.RawMessage `db:"response_body"`
		}

		expiresPredicate := "expires_at > CURRENT_TIMESTAMP"
		query := `
			SELECT request_hash, response_code, response_body 
			FROM idempotency_keys 
			WHERE idempotency_key = $1 AND ` + expiresPredicate
		err = im.db.Get(&existingRecord, query, idempotencyKey)
		if err == nil {
			if existingRecord.RequestHash != requestHash {
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "idempotency key was already used with a different request"})
				return
			}
			// Key exists, return cached response
			c.Data(existingRecord.ResponseCode, "application/json", existingRecord.ResponseBody)
			c.Abort()
			return
		}
		if !errors.Is(err, sql.ErrNoRows) {
			abortIdempotencyStoreUnavailable(c)
			return
		}

		// Use a custom writer to capture response
		w := &responseWriter{ResponseWriter: c.Writer, body: bytes.NewBufferString("")}
		c.Writer = w

		// Process request
		c.Next()

		// If request was successful, cache the response
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			im.cacheResponse(idempotencyKey, requestHash, c.Writer.Status(), w.body.Bytes())
		}
	}
}

func abortIdempotencyStoreUnavailable(c *gin.Context) {
	const code = "IDEMPOTENCY_STORE_UNAVAILABLE"
	c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
		"success": false,
		"code":    code,
		"error": gin.H{
			"code":    code,
			"message": "Sale idempotency could not be verified; retry after the database is available.",
		},
	})
}

func (im *IdempotencyMiddleware) cacheResponse(idempotencyKey, requestHash string, statusCode int, responseBody []byte) {
	expiresAt := time.Now().Add(24 * time.Hour) // Cache for 24 hours
	idempotencyMu.Lock()
	idempotencyCache[idempotencyKey] = struct {
		requestHash  string
		responseCode int
		responseBody []byte
		expiresAt    time.Time
	}{requestHash, statusCode, append([]byte(nil), responseBody...), expiresAt}
	idempotencyMu.Unlock()

	query := `
		INSERT INTO idempotency_keys (id, idempotency_key, resource_type,
			request_hash, response_code, response_body, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		ON CONFLICT (idempotency_key) DO NOTHING
	`

	_, err := im.db.Exec(query,
		uuid.New(), idempotencyKey, "api_request",
		requestHash, statusCode, strings.TrimSpace(string(responseBody)), expiresAt)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache idempotency response: %v\n", err)
	}
}

func (im *IdempotencyMiddleware) cleanupExpiredEntries() {
	now := time.Now()
	idempotencyMu.Lock()
	if now.Sub(im.lastCleanup) < time.Hour {
		idempotencyMu.Unlock()
		return
	}
	im.lastCleanup = now
	for key, cached := range idempotencyCache {
		if !now.Before(cached.expiresAt) {
			delete(idempotencyCache, key)
		}
	}
	idempotencyMu.Unlock()

	if _, err := im.db.Exec(`DELETE FROM idempotency_keys WHERE expires_at <= CURRENT_TIMESTAMP`); err != nil {
		fmt.Printf("Warning: failed to clean expired idempotency keys: %v\n", err)
	}
}

func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
