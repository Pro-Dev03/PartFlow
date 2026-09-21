package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequirePermission uses the existing role model and accepts permissions
// injected by an upstream cloud authorizer when one is available.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if hasPermission(c, permission) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":      "permission required",
			"code":       "PERMISSION_REQUIRED",
			"permission": permission,
		})
	}
}

func hasPermission(c *gin.Context, permission string) bool {
	if values, exists := c.Get("permissions"); exists {
		switch permissions := values.(type) {
		case []string:
			for _, value := range permissions {
				if strings.EqualFold(strings.TrimSpace(value), permission) {
					return true
				}
			}
		case map[string]bool:
			return permissions[permission]
		}
	}

	userID := GetUserID(c)
	if userID == uuid.Nil {
		return false
	}
	if IsConfiguredAdmin(c, userID) {
		return true
	}
	var role string
	if db == nil || db.QueryRowContext(c.Request.Context(), `SELECT COALESCE(role, '') FROM users WHERE id = $1`, userID).Scan(&role) != nil {
		return false
	}
	role = strings.ToLower(strings.TrimSpace(role))
	switch permission {
	case "archive.read", "financial-history.read", "audit.read":
		return role == "owner" || role == "admin" || role == "manager"
	case "archive.export":
		return role == "owner" || role == "admin"
	default:
		return false
	}
}
