package warranty

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers warranty routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Warranty routes
	warranties := router.Group("/warranties")
	{
		warranties.POST("", handler.CreateWarranty)
		warranties.GET("/:id", handler.GetWarranty)
		warranties.GET("/claims/summary", handler.GetWarrantyClaimsSummary)
	}

	// Warranty claim routes
	claims := router.Group("/warranty-claims")
	{
		claims.POST("", handler.CreateWarrantyClaim)
		claims.GET("/:id", handler.GetWarrantyClaim)
		claims.GET("", handler.ListWarrantyClaims)
		claims.PUT("/:id", handler.UpdateWarrantyClaim)
		claims.POST("/:id/approve", handler.ApproveWarrantyClaim)
		claims.POST("/:id/reject", handler.RejectWarrantyClaim)
	}
}
