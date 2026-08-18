package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
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

		// Validate token (this would use the actual JWT service)
		// For now, this is a placeholder
		claims, err := validateToken(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("organization_id", claims.OrganizationID)
		c.Set("role_id", claims.RoleID)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// Claims represents JWT claims (simplified version)
type Claims struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	RoleID         uuid.UUID
	Email          string
}

// validateToken validates a JWT token (placeholder)
func validateToken(token string, secret string) (*Claims, error) {
	// This would use the actual JWT service from auth package
	// For now, return a placeholder
	return &Claims{
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		RoleID:         uuid.New(),
		Email:          "user@example.com",
	}, nil
}

// GetUserID retrieves user ID from context
func GetUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}
	return userID.(uuid.UUID)
}

// GetOrganizationID retrieves organization ID from context
func GetOrganizationID(c *gin.Context) uuid.UUID {
	orgID, exists := c.Get("organization_id")
	if !exists {
		return uuid.Nil
	}
	return orgID.(uuid.UUID)
}

// GetRoleID retrieves role ID from context
func GetRoleID(c *gin.Context) uuid.UUID {
	roleID, exists := c.Get("role_id")
	if !exists {
		return uuid.Nil
	}
	return roleID.(uuid.UUID)
}
