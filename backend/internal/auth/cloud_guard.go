package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CloudGuard rejects direct local-API access that does not carry a currently
// valid cloud session. This protects reads, writes, and manual sync equally.
func CloudGuard(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
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
