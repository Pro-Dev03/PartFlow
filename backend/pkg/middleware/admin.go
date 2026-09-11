package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Admin protects account-management and destructive settings endpoints.
//
// The cloud schema intentionally does not require a role column, so production
// deployments should set PARTFLOW_ADMIN_EMAILS to a comma-separated allowlist.
// The owner address is kept as a backwards-compatible bootstrap administrator;
// it cannot be used without first authenticating as that account.
func Admin() gin.HandlerFunc {
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

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "administrator privileges required",
			"code":  "ADMIN_REQUIRED",
		})
	}
}

func IsConfiguredAdmin(c *gin.Context, userID uuid.UUID) bool {
	// Keep the type assertion in one place while avoiding a second user lookup
	// for development requests where authentication is intentionally disabled.
	if disableAuth && isLocalDatabaseMode() {
		return true
	}
	// In a local-first desktop process the authenticated user exists in the
	// cloud database, not necessarily in the local SQLite users table. Auth()
	// records the email returned by Render so the configured administrator can
	// still access protected settings without copying cloud user rows locally.
	if cloudEmail := strings.TrimSpace(c.GetString("cloud_user_email")); cloudEmail != "" {
		for _, configured := range strings.Split(os.Getenv("PARTFLOW_ADMIN_EMAILS"), ",") {
			if strings.EqualFold(strings.TrimSpace(configured), cloudEmail) && strings.TrimSpace(configured) != "" {
				return true
			}
		}
		if strings.EqualFold(cloudEmail, "owner@partflow.com") {
			return true
		}
	}
	if db == nil {
		return false
	}

	var email string
	if err := db.QueryRowContext(c.Request.Context(), "SELECT email FROM users WHERE id = $1", userID.String()).Scan(&email); err != nil {
		return false
	}

	for _, configured := range strings.Split(os.Getenv("PARTFLOW_ADMIN_EMAILS"), ",") {
		if strings.EqualFold(strings.TrimSpace(configured), strings.TrimSpace(email)) && strings.TrimSpace(configured) != "" {
			return true
		}
	}
	if strings.EqualFold(strings.TrimSpace(email), "owner@partflow.com") {
		return true
	}

	// Legacy role columns are intentionally ignored. Older migrations assigned
	// owner-like values to regular accounts, which must never grant admin access.
	return false
}
