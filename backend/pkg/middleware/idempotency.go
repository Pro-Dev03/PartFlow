package middleware

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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
	db *sqlx.DB
}

var idempotencyMu sync.Mutex
var idempotencyCache = map[string]struct {
	requestHash  string
	responseCode int
	responseBody []byte
	expiresAt    time.Time
}{}

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
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS idempotency_keys (
		id TEXT PRIMARY KEY,
		idempotency_key TEXT NOT NULL UNIQUE,
		resource_type TEXT NOT NULL,
		request_hash TEXT NOT NULL,
		response_code INTEGER NOT NULL,
		response_body TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	return &IdempotencyMiddleware{db: db}
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
			c.Next()
			return
		}
		idempotencyMu.Lock()
		defer idempotencyMu.Unlock()

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
		if cached, ok := idempotencyCache[idempotencyKey]; ok && time.Now().Before(cached.expiresAt) {
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

func (im *IdempotencyMiddleware) cacheResponse(idempotencyKey, requestHash string, statusCode int, responseBody []byte) {
	expiresAt := time.Now().Add(24 * time.Hour) // Cache for 24 hours
	idempotencyCache[idempotencyKey] = struct {
		requestHash  string
		responseCode int
		responseBody []byte
		expiresAt    time.Time
	}{requestHash, statusCode, append([]byte(nil), responseBody...), expiresAt}

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

func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
