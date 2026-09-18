package auth

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func requiresCloudAuthForLocalMode() bool {
	// SQLite is the store's local operational mode. A live cloud check is
	// opt-in so a temporary cloud outage cannot block the cashier locally.
	return strings.EqualFold(strings.TrimSpace(os.Getenv("PARTFLOW_REQUIRE_CLOUD_AUTH")), "true")
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

// CloudGuard optionally enforces the cloud authority for local API requests.
// Local operation remains available unless the deployment explicitly opts in.
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
