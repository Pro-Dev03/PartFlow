package inspections

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterRoutes registers inspections routes
func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Inspection routes
	inspections := router.Group("/inspections")
	{
		inspections.POST("", handler.CreateInspection)
		inspections.GET("/:id", handler.GetInspection)
		inspections.GET("", handler.ListInspections)
		inspections.PUT("/:id", handler.UpdateInspection)
		inspections.DELETE("/:id", handler.DeleteInspection)
		inspections.POST("/:id/pass", handler.PassInspection)
		inspections.POST("/:id/fail", handler.FailInspection)
		inspections.GET("/summary", handler.GetInspectionSummary)
	}
}
