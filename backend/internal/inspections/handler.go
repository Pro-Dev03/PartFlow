package inspections

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/middleware"
)

// Handler handles HTTP requests for inspections
type Handler struct {
	service *Service
}

// NewHandler creates a new inspection handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateInspection handles inspection creation
// @Summary Create a new inspection
// @Description Create a new inspection for a used item
// @Tags inspections
// @Accept json
// @Produce json
// @Param request body InspectionRequest true "Inspection request"
// @Success 201 {object} InspectionResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/inspections [post]
func (h *Handler) CreateInspection(c *gin.Context) {
	var req InspectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)
	userID := middleware.GetUserID(c)

	response, err := h.service.CreateInspection(c.Request.Context(), organizationID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetInspection handles getting an inspection by ID
// @Summary Get an inspection
// @Description Get an inspection by ID
// @Tags inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 200 {object} InspectionResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/inspections/{id} [get]
func (h *Handler) GetInspection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inspection ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.GetInspection(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListInspections handles listing inspections
// @Summary List inspections
// @Description List inspections with pagination and filters
// @Tags inspections
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param product_id query string false "Product ID filter"
// @Param status query string false "Status filter" Enums(pending, passed, failed, needs_repair)
// @Param condition query string false "Condition filter" Enums(excellent, very_good, good, fair, poor)
// @Param grade query string false "Grade filter" Enums(A, B, C, D, F)
// @Param start_date query string false "Start date filter"
// @Param end_date query string false "End date filter"
// @Param inspected_by query string false "Inspector ID filter"
// @Param search query string false "Search in serial number and notes"
// @Param sort_by query string false "Sort by field" default(inspection_date)
// @Param sort_order query string false "Sort order" default(DESC)
// @Success 200 {object} middleware.PaginatedResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/inspections [get]
func (h *Handler) ListInspections(c *gin.Context) {
	var req InspectionListRequest
	
	// Parse query parameters
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		req.Page = page
	}
	if perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil {
		req.PerPage = perPage
	}
	
	if productID := c.Query("product_id"); productID != "" {
		if id, err := uuid.Parse(productID); err == nil {
			req.ProductID = &id
		}
	}
	
	if inspectedBy := c.Query("inspected_by"); inspectedBy != "" {
		if id, err := uuid.Parse(inspectedBy); err == nil {
			req.InspectedBy = &id
		}
	}
	
	req.Status = c.Query("status")
	req.Condition = c.Query("condition")
	req.Grade = c.Query("grade")
	req.Search = c.Query("search")
	req.SortBy = c.DefaultQuery("sort_by", "inspection_date")
	req.SortOrder = c.DefaultQuery("sort_order", "DESC")
	
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			req.StartDate = &t
		}
	}
	
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			req.EndDate = &t
		}
	}

	organizationID := middleware.GetOrganizationID(c)

	inspections, total, err := h.service.ListInspections(c.Request.Context(), organizationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": inspections,
		"meta": gin.H{
			"page":        req.Page,
			"per_page":    req.PerPage,
			"total":       total,
			"total_pages": (total + req.PerPage - 1) / req.PerPage,
		},
	})
}

// UpdateInspection handles updating an inspection
// @Summary Update an inspection
// @Description Update an inspection by ID
// @Tags inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Param request body InspectionUpdateRequest true "Inspection update request"
// @Success 200 {object} InspectionResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/inspections/{id} [put]
func (h *Handler) UpdateInspection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inspection ID"})
		return
	}

	var req InspectionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.UpdateInspection(c.Request.Context(), id, organizationID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteInspection handles deleting an inspection
// @Summary Delete an inspection
// @Description Delete an inspection by ID
// @Tags inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 204
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/inspections/{id} [delete]
func (h *Handler) DeleteInspection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inspection ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	if err := h.service.DeleteInspection(c.Request.Context(), id, organizationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// PassInspection handles marking an inspection as passed
// @Summary Pass an inspection
// @Description Mark an inspection as passed
// @Tags inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 200 {object} InspectionResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/inspections/{id}/pass [post]
func (h *Handler) PassInspection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inspection ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.PassInspection(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// FailInspection handles marking an inspection as failed
// @Summary Fail an inspection
// @Description Mark an inspection as failed
// @Tags inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 200 {object} InspectionResponse
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 404 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/inspections/{id}/fail [post]
func (h *Handler) FailInspection(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inspection ID"})
		return
	}

	organizationID := middleware.GetOrganizationID(c)

	response, err := h.service.FailInspection(c.Request.Context(), id, organizationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetInspectionSummary handles getting inspection summary
// @Summary Get inspection summary
// @Description Get inspection summary statistics
// @Tags inspections
// @Accept json
// @Produce json
// @Success 200 {object} InspectionSummary
// @Failure 400 {object} middleware.ErrorResponse
// @Failure 401 {object} middleware.ErrorResponse
// @Failure 500 {object} middleware.ErrorResponse
// @Router /api/v1/inspections/summary [get]
func (h *Handler) GetInspectionSummary(c *gin.Context) {
	organizationID := middleware.GetOrganizationID(c)

	summary, err := h.service.GetInspectionSummary(c.Request.Context(), organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
