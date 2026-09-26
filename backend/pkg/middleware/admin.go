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

// OwnerOnly protects operations reserved for the single store owner. Unlike
// Admin, configured administrator addresses do not grant access here.
func OwnerOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		if email := strings.TrimSpace(c.GetString("cloud_user_email")); email != "" {
			if strings.EqualFold(email, "owner@partflow.com") {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "store owner privileges required", "code": "OWNER_REQUIRED"})
			return
		}

		if db != nil {
			var email string
			if err := db.QueryRowContext(c.Request.Context(), "SELECT email FROM users WHERE id = $1", userID.String()).Scan(&email); err == nil && strings.EqualFold(strings.TrimSpace(email), "owner@partflow.com") {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "store owner privileges required", "code": "OWNER_REQUIRED"})
	}
}

// LocalDatabaseOnly prevents device SQLite synchronization routes from being
// used on the cloud PostgreSQL service, where they would address the server's
// own ephemeral filesystem instead of the caller's device.
func LocalDatabaseOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil || (!strings.EqualFold(db.DriverName(), "sqlite") && !strings.EqualFold(db.DriverName(), "sqlite3")) {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "this operation is available only on a device with local SQLite", "code": "LOCAL_DATABASE_REQUIRED"})
			return
		}
		c.Next()
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
