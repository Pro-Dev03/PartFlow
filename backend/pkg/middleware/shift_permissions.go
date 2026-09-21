package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ShiftManager permits only managerial roles to open or close cash shifts.
// Cashiers may still use POS sales while a manager-owned shift is open.
func ShiftManager(database *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		if IsConfiguredAdmin(c, userID) {
			c.Next()
			return
		}

		var role string
		if err := database.QueryRowContext(c.Request.Context(), `SELECT COALESCE(role, '') FROM users WHERE id = $1`, userID.String()).Scan(&role); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "shift management permission required", "code": "SHIFT_PERMISSION_REQUIRED"})
			return
		}
		switch strings.ToLower(strings.TrimSpace(role)) {
		case "owner", "admin", "manager":
			c.Next()
		default:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "shift management permission required", "code": "SHIFT_PERMISSION_REQUIRED"})
		}
	}
}
