package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/internal/permissions"
)

var jwtSecret = []byte("your-secret-key-change-in-production")

var permissionService *permissions.Service

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

// Auth middleware for JWT authentication (based on worktrack)
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
			// Extract user_id and role from claims (based on worktrack)
			userID, ok := claims["user_id"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
				c.Abort()
				return
			}

			role, ok := claims["role"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role in token"})
				c.Abort()
				return
			}

			c.Set("user_id", userID)
			c.Set("role", role)
		}

		c.Next()
	}
}

// RequirePermission middleware to check if user has specific permission (updated for simplified claims)
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if permission service is initialized
		if permissionService == nil {
			// If permission service is not initialized, allow the request
			// This is a safety fallback for development
			c.Next()
			return
		}

		// Get user ID from context (set by Auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check if user is admin (bypass permission check)
		role, exists := c.Get("role")
		if exists && role == "admin" {
			c.Next()
			return
		}

		// Parse user ID as UUID for permission check
		userUUID, err := uuid.Parse(userID.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// Construct permission name
		permissionName := resource + "." + action

		// Check if user has the permission
		hasPermission, err := permissionService.HasPermission(context.Background(), userUUID, permissionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Permission denied",
				"message": "You don't have permission to perform this action",
				"required": permissionName,
			})
			c.Abort()
			return
		}

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

// SetPermissionService sets the permission service for middleware
func SetPermissionService(service *permissions.Service) {
	permissionService = service
}

