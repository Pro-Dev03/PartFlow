package sales

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegisterShiftRoutes exposes shift route registration to the central API router.
func RegisterShiftRoutes(router *gin.RouterGroup, db *sqlx.DB) error {
	return registerShiftRoutes(router, db)
}
