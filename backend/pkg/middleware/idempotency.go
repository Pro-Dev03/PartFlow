package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IdempotencyMiddleware struct {
	db *sqlx.DB
}

func NewIdempotencyMiddleware(db *sqlx.DB) *IdempotencyMiddleware {
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

		// Get organization ID from context
		organizationID, exists := c.Get("organization_id")
		if !exists {
			c.Next()
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

		// Check if idempotency key exists
		var existingRecord struct {
			ResponseCode int       `db:"response_code"`
			ResponseBody json.RawMessage `db:"response_body"`
		}

		query := `
			SELECT response_code, response_body 
			FROM idempotency_keys 
			WHERE organization_id = $1 AND idempotency_key = $2 
			AND expires_at > NOW()
		`
		err = im.db.Get(&existingRecord, query, organizationID, idempotencyKey)
		if err == nil {
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
			im.cacheResponse(organizationID, idempotencyKey, requestHash, c.Writer.Status(), w.body.Bytes())
		}
	}
}

func (im *IdempotencyMiddleware) cacheResponse(organizationID interface{}, idempotencyKey, requestHash string, statusCode int, responseBody []byte) {
	expiresAt := time.Now().Add(24 * time.Hour) // Cache for 24 hours

	query := `
		INSERT INTO idempotency_keys (id, organization_id, idempotency_key, resource_type,
			request_hash, response_code, response_body, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (organization_id, idempotency_key) DO NOTHING
	`

	_, err := im.db.Exec(query,
		uuid.New(), organizationID, idempotencyKey, "api_request",
		requestHash, statusCode, responseBody, expiresAt)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache idempotency response: %v\n", err)
	}
}

func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}