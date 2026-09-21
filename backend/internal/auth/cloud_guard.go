package auth

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func requiresCloudAuthForLocalMode() bool {
	// SQLite is the store's local operational mode, but it is not an
	// authorization boundary. Subscription verification is therefore enabled
	// by default and can only be disabled explicitly for isolated tests.
	value := strings.TrimSpace(strings.ToLower(os.Getenv("PARTFLOW_REQUIRE_CLOUD_AUTH")))
	return value != "false" && value != "0" && value != "no"
}

func isLocalDatabaseMode() bool {
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("DB_CONNECTION_MODE")))
	if mode == "" {
		mode = strings.TrimSpace(strings.ToLower(os.Getenv("DATABASE_MODE")))
	}
	if mode == "local" || mode == "sqlite" {
		return true
	}
	if mode == "cloud" {
		return false
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if strings.HasPrefix(databaseURL, "sqlite://") {
		return true
	}
	if databaseURL != "" {
		return false
	}
	if strings.TrimSpace(os.Getenv("DATABASE_URL_CLOUD")) != "" || strings.TrimSpace(os.Getenv("DB_CLOUD_URL")) != "" || strings.TrimSpace(os.Getenv("CLOUD_DATABASE_URL")) != "" {
		return false
	}
	return true
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
