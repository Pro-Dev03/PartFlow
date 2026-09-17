package auth

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func requiresCloudAuthForLocalMode() bool {
	// Local SQLite is the operational store, but every protected request still
	// requires a live cloud subscription decision in every environment.
	return true
}

func isLocalDatabaseMode() bool {
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("DB_CONNECTION_MODE")))
	if mode == "" {
		mode = strings.TrimSpace(strings.ToLower(os.Getenv("DATABASE_MODE")))
	}
	if mode == "local" || mode == "sqlite" {
		return true
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	return strings.HasPrefix(databaseURL, "sqlite://")
}

// CloudGuard rejects direct local-API access that does not carry a currently
// valid cloud session. This protects reads, writes, and manual sync equally.
func CloudGuard(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requiresCloudAuthForLocalMode() || !isLocalDatabaseMode() {
			c.Next()
			return
		}
		cloudToken := strings.TrimSpace(c.GetHeader("X-PartFlow-Cloud-Token"))
		if cloudToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{
				"code":    "CLOUD_AUTH_REQUIRED",
				"message": "cloud verification is required",
			}})
			return
		}
		if err := service.ValidateCloudAccess(c.Request.Context(), cloudToken); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{
				"code":    "CLOUD_AUTH_REQUIRED",
				"message": err.Error(),
			}})
			return
		}
		c.Next()
	}
}
