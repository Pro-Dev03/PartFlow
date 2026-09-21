package archive

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Handler struct{ repo *Repository }

func NewHandler(db *sqlx.DB) *Handler { return &Handler{repo: NewRepository(db)} }

func (h *Handler) ListEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "25"))
	events, total, err := h.repo.List(c.Request.Context(), ListRequest{
		Page: page, PerPage: perPage,
		Search: c.Query("search"), Section: c.Query("section"), EventType: c.Query("event_type"),
		UserID: c.Query("user_id"), Status: c.Query("status"), EntityType: c.Query("entity_type"),
		EntityID: c.Query("entity_id"), Reference: c.Query("reference"), StartDate: c.Query("start_date"), EndDate: c.Query("end_date"),
		SortBy: c.Query("sort_by"), SortOrder: c.Query("sort_order"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events, "meta": gin.H{"page": page, "per_page": perPage, "total": total}})
}
