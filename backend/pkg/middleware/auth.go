package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/partflow/smart-store/internal/auth"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(jwtService *auth.JWTService, db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token using the actual JWT service
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Parse user ID from claims
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in token"})
			c.Abort()
			return
		}

		// Verify the account still exists and is allowed to use the service. This
		// runs on every protected cloud request, so an administrator's manual
		// disable/delete takes effect without waiting for a token to expire.
		var userActive bool
		var subscriptionStatus string
		err = db.QueryRowContext(c.Request.Context(),
			"SELECT is_active, COALESCE(subscription_status, 'active') FROM users WHERE id = $1", userID).Scan(&userActive, &subscriptionStatus)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}
		status := strings.ToLower(strings.TrimSpace(subscriptionStatus))
		if !userActive || status == "expired" || status == "canceled" || status == "cancelled" || status == "deleted" {
			c.JSON(http.StatusForbidden, gin.H{"error": "subscription expired or account disabled", "code": "SUBSCRIPTION_EXPIRED"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", userID)
		if deviceID := strings.TrimSpace(c.GetHeader("X-PartFlow-Device-ID")); deviceID != "" {
			c.Set("device_id", deviceID)
		}

		c.Next()
	}
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}
	return userID.(uuid.UUID)
}
