package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte("your-secret-key-change-in-production")

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Auth middleware for JWT authentication
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Parse UUID from claims
			userID, err := uuid.Parse(claims["user_id"].(string))
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
				c.Abort()
				return
			}

			orgID, err := uuid.Parse(claims["organization_id"].(string))
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid organization ID in token"})
				c.Abort()
				return
			}

			roleID, err := uuid.Parse(claims["role_id"].(string))
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role ID in token"})
				c.Abort()
				return
			}

			c.Set("user_id", userID)
			c.Set("organization_id", orgID)
			c.Set("role_id", roleID)
			c.Set("email", claims["email"])
			c.Set("is_admin", claims["is_admin"])
		}

		c.Next()
	}
}

// RequirePermission middleware to check if user has specific permission
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement permission check
		// This would require injecting the auth service into middleware
		// For now, we'll just pass through
		c.Next()
	}
}

// Organization middleware to ensure organization context
func Organization() gin.HandlerFunc {
	return func(c *gin.Context) {
		organizationID := c.GetHeader("X-Organization-ID")
		if organizationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID required"})
			c.Abort()
			return
		}

		c.Set("organization_id", organizationID)
		c.Next()
	}
}

// Logger middleware
func Logger() gin.HandlerFunc {
	return gin.Logger()
}

// Recovery middleware
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// SetJWTSecret sets the JWT secret key
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// GetUserID gets the user ID from context
func GetUserID(c *gin.Context) uuid.UUID {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(uuid.UUID)
	}
	return uuid.Nil
}

// GetOrganizationID gets the organization ID from context
func GetOrganizationID(c *gin.Context) uuid.UUID {
	if orgID, exists := c.Get("organization_id"); exists {
		return orgID.(uuid.UUID)
	}
	return uuid.Nil
}