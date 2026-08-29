package supplierreturns

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func RegisterRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	h := NewHandler(NewService(db))
	r := router.Group("/supplier-returns")
	r.GET("", h.List)
	r.POST("", h.Create)
	r.POST("/:id/items", h.AddItem)
	r.DELETE("/:id", h.Delete)
	r.POST("/:id/reject", h.Reject)
	r.POST("/:id/complete", h.Complete)
}
