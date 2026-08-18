package organizations

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/partflow/smart-store/pkg/response"
)

// Handler handles HTTP requests for organizations
type Handler struct {
	service *Service
}

// NewHandler creates a new organizations handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create handles the creation of a new organization
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	org := NewOrganization(req.Name, req.Slug)
	if req.Email != "" {
		org.Email = &req.Email
	}
	if req.Phone != "" {
		org.Phone = &req.Phone
	}
	if req.Address != "" {
		org.Address = &req.Address
	}
	if req.City != "" {
		org.City = &req.City
	}
	if req.Country != "" {
		org.Country = &req.Country
	}
	if req.LogoURL != "" {
		org.LogoURL = &req.LogoURL
	}
	if req.Settings != nil {
		org.Settings = req.Settings
	}

	if err := h.service.Create(c.Request.Context(), org); err != nil {
		if err.Error() == "organization with slug '"+req.Slug+"' already exists" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, org.ToResponse(), "Organization created successfully")
}

// GetByID handles getting an organization by ID
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid organization ID")
		return
	}

	org, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Organization not found")
		return
	}

	response.OK(c, org.ToResponse(), "Organization retrieved successfully")
}

// GetBySlug handles getting an organization by slug
func (h *Handler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "Slug is required")
		return
	}

	org, err := h.service.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		response.NotFound(c, "Organization not found")
		return
	}

	response.OK(c, org.ToResponse(), "Organization retrieved successfully")
}

// Update handles updating an organization
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid organization ID")
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	org, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Organization not found")
		return
	}

	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.Slug != nil {
		org.Slug = *req.Slug
	}
	if req.Email != nil {
		org.Email = req.Email
	}
	if req.Phone != nil {
		org.Phone = req.Phone
	}
	if req.Address != nil {
		org.Address = req.Address
	}
	if req.City != nil {
		org.City = req.City
	}
	if req.Country != nil {
		org.Country = req.Country
	}
	if req.LogoURL != nil {
		org.LogoURL = req.LogoURL
	}
	if req.Settings != nil {
		org.Settings = req.Settings
	}
	if req.SubscriptionPlan != nil {
		org.SubscriptionPlan = *req.SubscriptionPlan
	}
	if req.SubscriptionStatus != nil {
		org.SubscriptionStatus = *req.SubscriptionStatus
	}

	if err := h.service.Update(c.Request.Context(), org); err != nil {
		if err.Error() == "organization with slug '"+org.Slug+"' already exists" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, org.ToResponse(), "Organization updated successfully")
}

// Delete handles deleting an organization
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid organization ID")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.NotFound(c, "Organization not found")
		return
	}

	response.NoContent(c)
}

// List handles listing all organizations with pagination
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	orgs, total, err := h.service.List(c.Request.Context(), page, perPage)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	responses := make([]*Response, len(orgs))
	for i, org := range orgs {
		responses[i] = org.ToResponse()
	}

	totalPages := (total + perPage - 1) / perPage
	meta := &response.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}

	response.Paginated(c, responses, meta, "Organizations retrieved successfully")
}
